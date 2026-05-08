# TDP Datagram v1 — Envelope 协议设计

**状态**：设计定稿（2026-05-07）
**实施**：未开始 / 当前 owl-common/spatial 是 Phase 1 过渡产物
**关联 memory**：[envelope_protocol_evolution](../../.claude/projects/-home-wisefido-owl/memory/envelope_protocol_evolution.md) · [agent_pipeline_north_star](../../.claude/projects/-home-wisefido-owl/memory/agent_pipeline_north_star.md)

---

## 0. 设计哲学

> **不重造轮子。借标准。把工程力量留给业务（fall 检测 / AI / room engine）**。

我们当前的协议层 ~80% 在重造已有标准的子集。Datagram v1 的核心决策是**把这部分还给标准**：

| 关注点 | 借的标准 | 取代我们之前自创的 |
|---|---|---|
| Envelope 格式 | **CloudEvents v1.0**（CNCF） | 自创 IoTStreamMessage |
| 空间寻址 | **IPv6 ULA**（RFC 4193 + RFC 4291） | 自创 7-段字符串 path |
| 寻址实现 | **`net/netip`** Go 标准库 + Postgres **INET** | 自创 spatial 字符串 helpers |
| Payload 序列化 | **Protobuf** | 自创 JSON dataValue |
| Trace / 因果链 | **OpenTelemetry W3C TraceContext** | 自创 ProducerSeq / CausationID |
| 严重度 | **syslog RFC 5424** (0-7) | 自创 severity int |
| 消息 ID | **ULID** | 自创 sequence_number |
| Schema 治理 | git monorepo + `.proto` | 隐式约定 |

owl 自己保留专注的：fall 检测算法、room engine 空间认知、AI verdict、card 业务模型、HIPAA 合规。**协议层不在这个列表**。

---

## 1. Datagram 结构总览

```go
type Datagram struct {
    // ============ CloudEvents v1.0 标准字段 ============
    SpecVersion string  // "1.0"  (CloudEvents version)
    ID          string  // ULID per message（DatagramID）
    Source      string  // Producer URN: "device-gateway://qinglan/v2.0.0"
    Type        string  // Schema FQN: "owl.monitor.track.v1"
    Time        time.Time

    // ============ Subject (CloudEvents subject) ============
    Subject string  // 默认 "device:<uuid>"，少数场景 "card:<uuid>" / "resident:<uuid>"

    // ============ owl extensions（CloudEvents extension fields） ============
    SpatialAddr  netip.Addr  // IPv6 ULA address (设备 trigger 时的 5W where)
    Severity     int         // syslog 0-7（envelope 一等公民）
    TraceID      string      // W3C TraceContext (OpenTelemetry)
    SpanID       string      // 当前 span
    ParentSpan   string      // 因果上游（取代 CausationID）
    Supersedes   []string    // belief revision: 替换的先前 datagram_id
    Tags         []Tag       // 跨 agent 元数据

    // ============ Payload (CloudEvents data) ============
    DataContentType string  // "application/protobuf"
    Data            []byte  // protobuf marshaled bytes
}

type Tag struct {
    Type  string  // dot-namespaced（"owl.stance" / "ml.feedback" / "external.fhir.category"）
    Value string  // 枚举值（"asserted" / "positive"）
}
```

**线上形态（CloudEvents binary mode，HTTP/Kafka/NATS 都支持）**：

```
ce-id: 01HXY...
ce-source: device-gateway://qinglan/v2.0.0
ce-subject: device:fcaedf95-c8bf-4f81-8a2b-6bb4549fff27
ce-type: owl.monitor.track.v1
ce-time: 2026-05-07T08:35:27.640Z
ce-spatialaddr: fd00:owl:1:112:42:301:a2ac:d523
ce-severity: 6
ce-traceid: 4bf92f3577b34da6a3ce929d0e0e4736

content-type: application/protobuf
body: <protobuf bytes ~50 字节>
```

总长 ~200 字节 vs 当前 IoTStreamMessage JSON 约 1KB → **~5x 压缩**。

---

## 2. IPv6 寻址（核心设计）

### 2.1 128 位完整布局

```
fd00 : owl  : TTTT : BB SS : UUUU : RR BB : DDDD DDDD
└──┘  └──┘  └──┘   └─┬─┘   └──┘   └─┬─┘  └──── 32 bit ────┘
ULA   owl   tenant  br/site unit   rm bd   = device_uid 末 32 bit
```

| 段 | 位宽 | bit 起始位 | 容量 | 来源 |
|---|---|---|---|---|
| ULA `fd00:owl:` | 32 | 0-31 | 固定常量 | RFC 4193 ULA + owl 命名空间 |
| Tenant | 16 | 32-47 | 65k 租户 | 全局分配（= /48 prefix） |
| Branch | 8 | 48-55 | 256/tenant | tenant 内分配（= /56） |
| Site (Bldg+Floor) | 8 | 56-63 | 16 bldg × 16 floor | branch 内分配（= /64 subnet 边界） |
| Unit | 16 | 64-79 | 65k/site | site 内分配 |
| Room | 8 | 80-87 | 256/unit | unit 内分配 |
| Bed | 8 | 88-95 | 256/room | room 内分配（= /96 prefix） |
| **Device-host** | **32** | **96-127** | **device_uid 末 32 bit** | **stateless 派生** |

### 2.2 完整地址例

```
device_uid = E598A2ACD523              ← 现有 device_uid（MAC-style）
末 8 hex   = A2ACD523                  ← 直接当 host part

完整 IPv6  = fd00:owl:0001:0112:0042:0301:A2AC:D523
                       │ │  │    │ │   └────┬─────┘
                       │ │  │    │ │        └── Device-host (= MAC 末 32 bit)
                       │ │  │    │ └────── Bed=01
                       │ │  │    └──────── Room=03
                       │ │  └────────────── Unit=0042
                       │ └──────────────── Site=12 (Building=1, Floor=2，4+4 split)
                       └────────────────── Branch=01
                  Tenant=0001 在 /48 内
```

字符串紧凑：`fd00:owl:1:112:42:301:a2ac:d523`（28 字符，零位省略后）；二进制 16 字节。

### 2.3 Device-host 派生：stateless

```go
func DeriveDeviceAddr(spatialPrefix netip.Prefix, deviceUID string) netip.Addr {
    macSuffix := deviceUID[len(deviceUID)-8:]   // "A2ACD523"
    hostBits, _ := hex.DecodeString(macSuffix)
    addrBytes := spatialPrefix.Masked().Addr().As16()
    copy(addrBytes[12:], hostBits)
    return netip.AddrFrom16(addrBytes)
}
```

**完全无状态**：device 第一次接入 → gateway 拿 device_uid 拼 spatial prefix → 立即得到合法 IPv6。无须分配协议、无并发安全、无回收逻辑。

**冲突空间**：32 bit = 43 亿，床内 1-3 设备时冲突概率 ~10⁻⁹。装机时 SQL 一次性校验冲突，万一命中（百万分之一）人工指定 host part。

### 2.4 设备搬房间 = IPv6 mobility

```
搬迁前：fd00:owl:1:112:42:301:a2ac:d523  (room 03 bed 01)
搬迁后：fd00:owl:1:112:42:501:a2ac:d523  (room 05 bed 01)
        ↑ prefix 变了                    ↑ host 没变（同一台物理设备）
```

跟真 IPv6 mobility 完全同构——**网络位置变，硬件标识不变**。AI 可以用 host 部分（末 32 bit）追踪同一台物理设备的历史轨迹。

### 2.5 分层 prefix 与 card 类型

cards 表的 `spatial_prefix INET` 列就是 IPv6 prefix，cardagg 路由 = **真 IPv6 longest-prefix-match**：

| 卡型 | spatial_prefix | 含义 |
|---|---|---|
| ActiveBedCard | `/96` | bed 内所有设备的 alarm 都路由到此卡 |
| UnitCard | `/80` | unit 内（公共区域 device，无 bed） |
| DeviceCard | `/128` | 单设备完整地址（罕见，作 fallback） |

```sql
-- alarm 路由：标准 BGP-style longest-prefix-match
SELECT card_id FROM cards
WHERE spatial_prefix >>= $alarm_addr::inet  -- INET 包含操作符（内核级）
ORDER BY masklen(spatial_prefix) DESC        -- 最长 prefix 优先
LIMIT 1;
```

Postgres GIST 索引原生支持，O(log n) 命中，比 JOIN devices→beds→rooms→units 快百倍。

### 2.6 prefix-match 查询通用模式

```sql
-- 所有 unit U42 内的 alarm
SELECT * FROM alarm_events WHERE device_addr <<= 'fd00:owl:1:112:42::/80'::inet;

-- 所有 room R3 内的 alarm
SELECT * FROM alarm_events WHERE device_addr <<= 'fd00:owl:1:112:42:300::/88'::inet;

-- 跨 unit 范围 / branch 范围 / 整 tenant 范围 同样模式
```

---

## 3. CloudEvents Envelope 字段详解

### 3.1 必填（CloudEvents required）

| 字段 | 例 | 说明 |
|---|---|---|
| `specversion` | `"1.0"` | CloudEvents 版本，固定 |
| `id` | `"01HXY..."` | ULID，全局唯一消息 ID |
| `source` | `"device-gateway://qinglan/v2.0.0"` | producer URN（who emit） |
| `type` | `"owl.monitor.track.v1"` | schema FQN（payload 形态） |

### 3.2 推荐字段

| 字段 | 例 | 说明 |
|---|---|---|
| `subject` | `"device:fcaedf95-..."` | 这条消息说的是谁（about whom） |
| `time` | `"2026-05-07T08:35:27.640Z"` | 事发时刻 |
| `datacontenttype` | `"application/protobuf"` | 标识 data 序列化格式 |

### 3.3 owl 私有 extensions（CloudEvents extension 机制）

| extension | 例 | 说明 |
|---|---|---|
| `spatialaddr` | `"fd00:owl:1:112:42:301:a2ac:d523"` | 5W where 完整 IPv6 |
| `severity` | `2` | syslog 0-7（envelope 一等公民给路由层用） |
| `traceid` | `"01HXY3K7G2..."` | W3C TraceContext，整条业务流；常用 root datagram ID |
| `spanid` | `"01HXY3K7G2..."` | 当前 span = 当前 DatagramID（**一物两用**） |
| `parentspan` | `"01HXY-prev-id"` | 因果上游 = 上游 datagram 的 ID |
| `nodeid` | `"<process-instance-uuid>"` | 可选；进程实例 UUID（启动时分配，区分版本/重启） |
| `supersedes` | `["01HXY-prev-1", "01HXY-prev-2"]` | belief revision refs |
| `tags` | `[{"type":"owl.stance","value":"asserted"}]` | 跨 agent 元数据 |

### 3.4 字段映射关系表

| owl 概念 | Datagram 字段 |
|---|---|
| Producer（who emit） | `source` URN |
| Subject（about whom） | `subject` |
| When | `time` |
| Where (5W) | extension `spatialaddr`（IPv6） |
| Severity | extension `severity` |
| Schema | `type` |
| Causation chain | extension `traceid`/`spanid`/`parentspan` |
| Belief revision | extension `supersedes` |
| Cross-agent metadata | extension `tags` |
| Payload | `data` (protobuf) |

### 3.5 Producer identity 多层（service / instance / message）

每个 service 表达自己产生的 datagram 时，三层身份各占字段：

| 层 | 字段 | 含义 | 例 |
|---|---|---|---|
| **软件类型 + 版本** | `source` URN | "我这个 service 是什么 / 什么版本" | `sensor://owl.engine.lost_fall/v1.0.0` |
| **进程实例**（可选） | `nodeid` extension | "我这次启动的进程是哪个"（重启变） | `uuid-of-this-process` |
| **这条消息** | `id` (= ULID) | "我产生的具体某条消息" | `01HXY3K7G2...` |

每个 service 启动时确定自己的 `source` 和 `nodeid`，**每条消息自己 mint ULID 作 `id`**——无须中央协调。

**为什么需要三层**：

- 只有 `source`：分不清两次部署 / 两次重启
- 只有 `nodeid`：不知道是 firmware-gateway 还是 sensor 还是 cognitive
- 只有 `id`：无法快速归因到 service

debug / audit / dedup 都依赖这三层完整。

**ULID 的"一物多用"**：

每条 datagram 自己生成的 ULID 同时是：
1. **CloudEvents `id`**（消息唯一身份）
2. **W3C TraceContext `spanid`**（当前 trace 节点）
3. **idempotency key**（producer 内部去重）
4. **parentspan 引用目标**（被下游引用作"我从这条派生"）

ULID 16 字节 = W3C span_id 16 字节，格式天然兼容。**不需要单独再生成 span_id**。

**ULID 不撞 ID 的保证**：

- 头 48 bit = ms 时间戳，ULID 单调递增（lex-sortable）
- 后 80 bit = 随机
- 即使每秒 emit 1 亿条消息，撞 ID 期望 ≈ 3 亿年后

无须中央分配，每个 service 自己 mint 即可。

**生成示例**：

```go
// service 启动时一次性确定身份
identity := ServiceIdentity{
    Source:  "sensor://owl.engine.lost_fall/v1.0.0",
    NodeID:  uuid.New().String(),  // 进程级
    Version: "1.0.0",
}

// 每条消息
func emit(payload proto.Message) {
    evt := cloudevents.NewEvent("1.0")
    evt.SetID(ulid.Make().String())              // 自己 mint，无需协调
    evt.SetSource(identity.Source)
    evt.SetExtension("nodeid", identity.NodeID)
    evt.SetExtension("spanid", evt.ID())         // = id（一物两用）
    // parentspan / traceid 由 caller 传入因果上下文
    // ...
}
```

---

## 4. Tags（跨 agent 元数据）

### 4.1 用途定位

**放不进 envelope 结构性字段、又跨多 agent 通用的"分类元数据"**。不为某一种用途设计，而是给"需要分类标签 + 跨 producer/consumer 共识"的维度留接口。

### 4.2 6 类典型用途

| 类别 | 例子 | 谁打 | 谁用 |
|---|---|---|---|
| **消息生命周期** | stance（asserted/hypothesized/retracted） | producer 自己 | 下游决定通知/审计 |
| **subject 分类** | cohort（high-risk/post-discharge）、priority | 周期评估 agent | 通知策略 |
| **ML 训练标签** | feedback（positive/negative）、label-source | 用户 ack 反传 | 离线 ETL 训练集 |
| **领域 / 功能** | domain（fall/vital/sleep）、agent-name | producer 自己 | LLM 学习分桶 |
| **生命周期 / 灰度** | phase（canary/experimental/prod） | 灰度 PR 临时 | metrics 隔离 |
| **多 agent 协同** | vote（for/against）、confidence-tier | 各 agent 各发 | 仲裁 agent 聚合 |

### 4.3 不应放 Tag 的

| 内容 | 应该放哪 |
|---|---|
| 自由文本 | payload |
| 数值（confidence/score/heart_rate） | payload |
| 标识符（device_id/card_id） | Subject / Producer / spatial_addr |
| 严重度（0-7） | `severity` extension |
| 位置 | `spatialaddr` |
| 时间 | `time` |
| 因果 | `traceid`/`parentspan` |

### 4.4 格式：`{Type, Value}`

```go
type Tag struct {
    Type  string  // dot-namespaced 分类维度
    Value string  // 枚举值
}
```

**两字段，dot-namespaced**——比 FHIR `{System, Code, Display}` 简，但保留 namespacing。

### 4.5 命名空间约定（建议，不强制）

| Type prefix | 含义 | 例 |
|---|---|---|
| `owl.*` | owl 系统内置 | `owl.stance`, `owl.cohort`, `owl.domain` |
| `ml.*` | ML / 训练 | `ml.feedback`, `ml.label-source` |
| `ops.*` | 运维 / 生命周期 | `ops.phase`, `ops.shadow` |
| `agent.<name>.*` | 特定 agent 自定义 | `agent.health-llm.confidence-tier` |
| `external.<vocab>.*` | 外部词汇表 | `external.fhir.category`, `external.snomed.finding` |
| `tenant.<id>.*` | tenant 私有标签 | `tenant.lakewood.facility-tier` |

**FHIR 不是基础格式，是众多 namespace 之一**。需要时引：

```json
{ "type": "external.fhir.category", "value": "safety" }
{ "type": "external.snomed.finding", "value": "2761008" }
```

不需要时根本不引——大多数业务标签 `owl.*` 就够。

### 4.6 实例

一条 fall alarm：

```json
"tags": [
  { "type": "owl.stance", "value": "asserted" },
  { "type": "owl.domain", "value": "fall" },
  { "type": "owl.cohort", "value": "high-risk" }
]
```

---

## 5. Schema (Protobuf) 治理

### 5.1 命名

`owl.<channel>.<topic>.v<n>` —— 跟 CloudEvents `type` 字段对齐：

| Schema 名 | 用途 |
|---|---|
| `owl.monitor.track.v1` | radar track sample |
| `owl.monitor.vital.v1` | sleepad 生命体征 |
| `owl.event.enter_room.v1` | 进入房间 |
| `owl.event.left_bed.v1` | 离床事件 |
| `owl.alarm.fall.v1` | fall 报警 |
| `owl.alarm.night_absence.v1` | 整夜未归 |
| `owl.config.alarm_process.v1` | 用户处理 alarm |
| `owl.config.card_change.v1` | 卡片变更 |

### 5.2 仓库布局

```
owl-common/proto/
  monitor/
    v1/
      track.proto
      vital.proto
  event/
    v1/
      enter_room.proto
      left_bed.proto
  alarm/
    v1/
      fall.proto
      night_absence.proto
  config/
    v1/
      alarm_process.proto
```

git 单仓库管理，编译期 codegen 给 Go / Python / TypeScript。**暂不上 Buf Schema Registry**——业务规模到时再加。

### 5.3 演进规则（proto3 标准）

- 添加字段：tag 序号唯一，旧版本 reader 跳过未知字段（向后兼容）
- 删除字段：不要删 tag，标记 `reserved`，旧 reader 不会重用
- 改类型：禁（重大不兼容），需新版本 `.v2`
- enum 加值：兼容（reader 拿到未知值默认转 `UNKNOWN`）

### 5.4 实例：fall.proto

```proto
syntax = "proto3";
package owl.alarm.v1;

message FallAlarm {
    int32 fall_score = 1;
    string reason = 2;
    Position position = 3;
    Evidence evidence = 4;
}

message Position {
    int32 x = 1;
    int32 y = 2;
    int32 z = 3;
}

message Evidence {
    int64 wait_ms = 1;
    int64 frozen_start_ms = 2;
    int64 engine_fire_ms = 3;
    int64 replay_anchor_ms = 4;
}
```

---

## 6. 三个具体消息样例

### 6.1 Firmware radar track（连续 monitor 流）

```json
{
  "specversion": "1.0",
  "id": "01HXY-12345",
  "source": "device-gateway://qinglan/v2.0.0",
  "subject": "device:fcaedf95-c8bf-4f81-8a2b-6bb4549fff27",
  "type": "owl.monitor.track.v1",
  "time": "2026-05-07T08:35:27.640Z",

  "spatialaddr": "fd00:owl:1:112:42:301:a2ac:d523",
  "severity": 6,
  "traceid": "4bf92f3577b34da6a3ce929d0e0e4736",
  "tags": [
    { "type": "owl.domain", "value": "track" }
  ],

  "datacontenttype": "application/protobuf",
  "data": "<base64 protobuf>"
  // proto: MonitorPayload { tracks: [
  //   {track_id:1, position_x:250, position_y:210, pose:1, track_confidence:80},
  //   {track_id:2, position_x:-120, position_y:50, pose:0, track_confidence:35}
  // ]}
}
```

注意 payload 是数组——多 track 各自带 `track_confidence`，envelope 不染指 per-item quality。

### 6.2 AI engine_lost_fall（producer 契约 = monitor + alarm 共发）

**Datagram A — monitor verdict（AI 对当下空间的事实声明）**：

```json
{
  "specversion": "1.0",
  "id": "01HXY-MONITOR-VERDICT",
  "source": "sensor://owl.engine.lost_fall/v1.0.0",
  "subject": "device:fcaedf95-...",
  "type": "owl.monitor.track.v1",
  "time": "2026-05-07T08:29:57.640Z",

  "spatialaddr": "fd00:owl:1:112:42:301:a2ac:d523",
  "severity": 6,
  "traceid": "4bf92f3577b34da6a3ce929d0e0e4736",
  "tags": [
    { "type": "owl.domain", "value": "fall" },
    { "type": "owl.verdict", "value": "engine_inferred" },
    { "type": "owl.stance", "value": "asserted" }
  ],

  "datacontenttype": "application/protobuf",
  "data": "<base64>"
  // proto: TrackSample {
  //   track_id: 1, position_x: 250, position_y: 210, position_z: 0,
  //   pose: 5,  // FALL
  //   verdict: "engine_inferred",
  //   source: "engine_lost_fall",
  //   fall_score: 100
  // }
}
```

**Datagram B — alarm（基于 verdict 派生通知）**：

```json
{
  "specversion": "1.0",
  "id": "01HXY-ALARM-1",
  "source": "sensor://owl.engine.lost_fall/v1.0.0",
  "subject": "device:fcaedf95-...",
  "type": "owl.alarm.fall.v1",
  "time": "2026-05-07T08:29:57.640Z",

  "spatialaddr": "fd00:owl:1:112:42:301:a2ac:d523",
  "severity": 2,
  "traceid": "4bf92f3577b34da6a3ce929d0e0e4736",
  "parentspan": "01HXY-MONITOR-VERDICT",
  "tags": [
    { "type": "owl.domain", "value": "fall" },
    { "type": "owl.stance", "value": "asserted" }
  ],

  "datacontenttype": "application/protobuf",
  "data": "<base64>"
  // proto: FallAlarm {
  //   fall_score: 100,
  //   reason: "track_lost_no_exit_room",
  //   position: {x: 250, y: 210},
  //   evidence: {wait_ms: 300000, replay_anchor_ms: 1778164197640, engine_fire_ms: 1778164527640}
  // }
}
```

两条 datagram 共享 `traceid` —— 同一次 fall 业务流；alarm 用 `parentspan` 引用 verdict。

### 6.3 用户 ack alarm（config 流并轨）

```json
{
  "specversion": "1.0",
  "id": "01HXY-CONFIG-1",
  "source": "owlfront://user-action/v1.0.0",
  "subject": "card:12a63cf9-...",
  "type": "owl.config.alarm_process.v1",
  "time": "2026-05-07T08:34:05.000Z",

  "spatialaddr": "fd00:owl:1:112:42:301:a2ac:d523",
  "severity": 6,
  "traceid": "4bf92f3577b34da6a3ce929d0e0e4736",
  "parentspan": "01HXY-ALARM-1",
  "tags": [
    { "type": "ml.feedback", "value": "positive" },
    { "type": "ml.label-source", "value": "manual" },
    { "type": "owl.domain", "value": "fall" }
  ],

  "datacontenttype": "application/protobuf",
  "data": "<base64>"
  // proto: AlarmProcessAction {
  //   event_id: "c89c45c9-...",
  //   process_type: "ack",
  //   handle_type: "verified",
  //   handler_id: "user-uuid"
  // }
}
```

---

## 7. 与三大参考体系横向对比

| 维度 | IPv6 | Matter / Google Home | CloudEvents | **TDP Datagram** |
|---|---|---|---|---|
| 寻址 | 128-bit prefix+host | NodeID+Cluster+Endpoint | source URI | **真 IPv6 ULA**（直接用，不模拟） |
| isolation | AS-level | fabric-level | 无 | **TenantID = /48 prefix（HIPAA 硬边界）** |
| 消息 ID | 无 | NodeID | id | **ULID** |
| Producer | 无 | NodeID | source URI | source URN（CloudEvents 标准） |
| Subject | 无 | endpoint | subject | subject（CloudEvents 标准） |
| Schema | 无 | 强（Cluster+Attribute） | dataschema URI | **Protobuf**（强，跨语言） |
| 流分类 | next-header | ClusterID | type | type FQN + Stream 概念（pub/sub topic） |
| 因果链 | 无 | 无 | 无 | **W3C TraceContext**（OpenTelemetry 标准） |
| 严重度 | 无 | 无 | 无 | **syslog RFC 5424** |
| 跨 agent metadata | 无 | 无 | 无 | **Tags `{Type, Value}` dot-namespaced** |

**TDP Datagram = CloudEvents（envelope）+ IPv6 ULA（寻址）+ Protobuf（payload）+ OpenTelemetry（因果）+ syslog（严重度）+ FHIR-friendly Tags**。**几乎全部借标准，owl 只加少量 extension**。

---

## 8. Multi-agent AI 协同的 envelope 适配性

### 8.1 给 AI 的能力（envelope 直接提供）

| 能力 | 字段 |
|---|---|
| Provenance | `source` URN |
| Trust hierarchy | source URN scheme（`device-gateway://`/`sensor://`/`cognitive://`） |
| Typed payload | `type` + Protobuf schema |
| Spatial context | `spatialaddr` |
| Temporal reasoning | `time` |
| Multi-causation | `parentspan`（trace 树自然多源）+ `traceid`（同 trace 关联） |
| Belief revision | `supersedes` |
| Dispatchable identity | `id` (ULID) |
| 可学习元数据 | `tags` |

### 8.2 三层结构原则

| 关注点 | 放哪 |
|---|---|
| 单 datagram 的 confidence / quality | **payload**（per-item，多 track 各自 confidence） |
| stance/cohort/domain/training-label | **Tag** |
| supersedes / parentspan | **extension**（结构性 ref） |
| PHI / privacy | **暂不需要**（内部 UUID 化已解决） |
| 推理过程 / explanation | **payload** schema 的 optional 字段 |
| embedding / 向量 | **外部服务**（vector DB），datagram 引用 ID |

### 8.3 multi-agent 模式 → 字段映射

| 模式 | envelope 怎么表达 |
|---|---|
| **专精分工** | 各自 source URN |
| **多源融合** | trace 树自然支持多 parentspan（OpenTelemetry 标准 SpanLink） |
| **仲裁 / 投票** | 多 agent 同 traceid，`tags+=owl.vote:for/against`，仲裁 agent 聚合 |
| **belief revision** | 新 datagram `supersedes=[old_id]` + `tags+=owl.stance:retracted` |
| **假设态** | `tags+=owl.stance:hypothesized` |
| **训练数据 round-trip** | 用户 ack `parentspan=alarm_id` + `tags+=ml.feedback:positive`，离线 ETL 按 tag 分桶 |

---

## 8.5 业务收益示例：Trace-based Replay

Datagram v1 + producer 契约最直接的业务收益之一：**replay AI verdict 从"特殊 case"变成"trace 树自然查询"**。

### 痛点（当前架构的 D523 fall replay 问题）

[Memory: phi_encryption_restore](../../.claude/projects/-home-wisefido-owl/memory/phi_encryption_restore.md) 之前几次会话发现：

- engine_lost_fall 是引擎推断（track 丢失 wait_ms 后才 emit alarm）
- alarm trigger 时间已经在 track 实际活动结束之后 5min
- replay 默认窗口 `[trigger - 60s, trigger + 120s]` 完全错过 track 真实活动时段
- 我们曾手工塞 `trigger_data.evidence.replay_anchor_ms` 字段救场
- 但 pose=0 在画布上仍然显示不出 fall icon

**根因**：AI 的判定（"我认为这一刻 pose=fall"）只在 alarm payload 里，**没进 monitor 流**。replay 拉不到那一刻的 track 数据。

### Datagram v1 的解（producer 契约 + 同 trace）

每条 AI verdict **必发 monitor datagram**（producer 契约 §11.1），跟 alarm datagram 共享 traceid：

```
[firmware track #1]
  id: 01HXY-FW-1, traceid: 01HXY-TRACE
  source: device-gateway://qinglan/v2.0.0
  spatialaddr: fd00:owl:1:112:42:301:a2ac:d523
  time: 08:25:00, pose: 1 (stand)

[firmware track #2..N]
  id: 01HXY-FW-2..., traceid: 01HXY-TRACE   (同一 trace)
  pose: 1 (stand) 静止
  time: 08:27:00..08:29:00
  ── (track 突然断流) ──

[AI synth monitor verdict]
  id: 01HXY-VERDICT, traceid: 01HXY-TRACE
  source: sensor://owl.engine.lost_fall/v1.0.0
  spatialaddr: fd00:owl:1:112:42:301:a2ac:d523  (= 最后已知位置)
  parentspan: 01HXY-FW-N  (引用最后一条 firmware track)
  time: 08:29:57.640
  pose: 5 (FALL)  ← AI 推断的 pose，第一公民
  tags: [{type:owl.verdict, value:engine_inferred}]

[AI alarm Fall]
  id: 01HXY-ALARM, traceid: 01HXY-TRACE
  source: sensor://owl.engine.lost_fall/v1.0.0
  parentspan: 01HXY-VERDICT
  type: owl.alarm.fall.v1
  time: 08:29:57.640

[user ack]
  id: 01HXY-ACK, traceid: 01HXY-TRACE
  source: owlfront://user-action/v1.0.0
  parentspan: 01HXY-ALARM
  time: 08:34:05
```

### Replay 查询变成一句

```sql
-- 后端：拿 alarm 的 traceid → 一句拿全因果树
SELECT *
FROM datagrams_index
WHERE traceid = $1
ORDER BY time;
```

### Frontend 渲染逻辑

```ts
// 拿到 traceid 全部 datagrams
const datagrams = await fetchByTrace(alarm.traceid);

// 按 type 分类渲染时间线
for (const d of datagrams) {
  if (d.type === 'owl.monitor.track.v1') {
    // 走 RadarCanvas 标准渲染
    canvas.drawTrack(d.payload, {
      color: d.tags.includes('owl.verdict:engine_inferred')
        ? COLOR_AI_INFERRED   // AI synth track 高亮色
        : COLOR_FIRMWARE,     // 普通 firmware track
    });
  } else if (d.type === 'owl.alarm.fall.v1') {
    canvas.markFallEvent(d.spatialaddr, d.time);
  }
}
```

### 对比（前/后）

| 维度 | 当前架构 | Datagram v1 |
|---|---|---|
| 查询接口 | alarm_events + iot_timeseries 多表 JOIN | 单一 traceid 查 datagram store |
| AI verdict 可见性 | 藏在 alarm payload | monitor 流第一公民 |
| anchor 字段救场 | 必需（trigger_data.evidence.replay_anchor_ms） | 不需要——AI synth datagram 自带 anchor 时刻 |
| pose=fall 显示 | 失败（pose=0） | 成功（AI synth track pose=5） |
| 跨 producer 因果 | 不可见 | parentspan 链清晰 |
| AI 自纠错 | 无机制 | supersedes 替换历史 datagram，UI 自动刷新 |
| fusion verdict | 不可表达 | W3C SpanLinks 多 parent |

### 工程层撤掉的临时代码

落 Datagram v1 后，下面这些手工/特殊代码可以全部撤：

1. ❌ `alarm_events.trigger_data.evidence.replay_anchor_ms` 字段 — 不再需要 anchor 救场
2. ❌ frontend WaveMonitor.replayFromWaveAlarm 手工 1min 前/2min 后窗口逻辑 — 改成 traceid 查
3. ❌ 给历史 D523 fall patch trigger_data 的一次性 SQL — 不再需要历史救场
4. ❌ "AI overlay 要不要也回放 card:alarmstatus:stream" 的纠结 — 不需要单独流
5. ❌ pose=0 显示不了 fall 的兜底逻辑 — AI synth track 本来就 pose=5

**多次临时修补 → 一个统一架构吸收**。

---

## 9. DB schema 设计

### 9.1 主键策略：UUID 不动，新加 INET 列

```sql
-- 所有空间实体表加 spatial_addr INET（IPv6 地址或 prefix）
ALTER TABLE tenants    ADD COLUMN spatial_prefix INET;  -- /48
ALTER TABLE branches   ADD COLUMN spatial_prefix INET;  -- /56
ALTER TABLE buildings  ADD COLUMN spatial_prefix INET;  -- /60 (或 site /64)
ALTER TABLE units      ADD COLUMN spatial_prefix INET;  -- /80
ALTER TABLE rooms      ADD COLUMN spatial_prefix INET;  -- /88
ALTER TABLE beds       ADD COLUMN spatial_prefix INET;  -- /96
ALTER TABLE devices    ADD COLUMN current_addr   INET;  -- /128 完整地址
ALTER TABLE cards      ADD COLUMN spatial_prefix INET;  -- 卡型对应深度

-- alarm_events / iot_timeseries 加 device_addr 列（trigger 时 device 的完整地址）
ALTER TABLE alarm_events    ADD COLUMN device_addr INET;
ALTER TABLE iot_timeseries  ADD COLUMN device_addr INET;

-- GiST 索引支持 prefix 包含查询
CREATE INDEX idx_alarm_events_addr ON alarm_events USING GIST (device_addr inet_ops);
CREATE INDEX idx_iot_ts_addr       ON iot_timeseries USING GIST (device_addr inet_ops);
CREATE INDEX idx_cards_prefix      ON cards USING GIST (spatial_prefix inet_ops);
```

**UUID PK 不动**：外键 / API / 历史数据 / 第三方集成不破。

### 9.2 父级分配子级 local_id（应用层）

创建 unit U42 in branch B01 site S12：

```go
func CreateUnit(branchID uuid.UUID, ...) (Unit, error) {
    branch := db.Get("SELECT * FROM branches WHERE branch_id = $1", branchID)
    nextLocalID := db.Get("SELECT COALESCE(MAX(unit_local_id), 0) + 1 FROM units WHERE branch_id = $1", branchID)

    // 父级 prefix /64 + unit local 16 bit → /80 prefix
    parentPrefix := branch.SpatialPrefix  // e.g., "fd00:owl:1:112::/64"
    unitPrefix := netip.PrefixFrom(
        addBitsToAddr(parentPrefix.Addr(), uint16(nextLocalID), 64, 80),
        80,
    )

    return Unit{
        UnitID:         uuid.New(),
        UnitLocalID:    nextLocalID,
        SpatialPrefix:  unitPrefix,
        BranchID:       branchID,
    }, nil
}
```

**Device 例外**：device 不在父级"分配"，而是 stateless 派生（见 §2.3）。

### 9.3 alarm_events / iot_timeseries 不存 card_id

`card_id` 是 cache 不是 snapshot（[详见 §11.5](#115-card_id-是-cache-不是-snapshot)）：

```
真实源：cards 表 spatial_prefix ↔ card_id 映射
缓存：各服务启动加载到内存 map，cardChange stream 增量 invalidation
查询：spatial_addr <<= spatial_prefix（GIST 索引 longest-prefix-match）
```

### 9.4 通用模式：spatial-scoped resource（不只 cards / layouts）

**所有"按空间范围生效的资源"都走同一套 longest-prefix-match 模式**：

```sql
CREATE TABLE <resource> (
    <resource>_id  UUID PRIMARY KEY,
    spatial_prefix INET NOT NULL,
    config         JSONB,
    UNIQUE (spatial_prefix)
);
CREATE INDEX idx_<resource>_prefix ON <resource> USING GIST (spatial_prefix inet_ops);
```

应用到当前系统挂在 room_id 上的所有资源：

| 资源 | 现状 | 新模式 |
|---|---|---|
| Layout (`rooms.layout_config`) | room_id 主键 | 单独 `layouts` 表，spatial_prefix /88 (room) 或 /80 (unit 多 room 合一) |
| Wall polygon (`rooms.wall_polygon`) | room_id 主键 | `walls` 表，spatial_prefix /88 |
| Grid 学习状态 | room_id 主键 | `grids` 表，spatial_prefix /88 |
| Cell semantic labels (AI_fall) | room_id 主键 | `cell_labels` 表，spatial_prefix /88 |
| Card | card_id UUID | `cards` 表，spatial_prefix /96 (bed) / /80 (unit) |
| 通知策略 | 无（散落业务逻辑） | `notification_policies` 表，spatial_prefix /80 unit / /48 tenant |

**好处**：
- 统一一套查询机制（INET longest-prefix-match）
- 自然支持"父级配置 + 子级 override"（unit 级默认 + room 级覆盖）
- 解决 WaveMonitor 老注释里的"未来 bedroom+bookroom 合一 layout"问题（spatial_prefix 上推到 /80）
- 跟 cards 多 card 共存（ActiveBedCard /96 + UnitCard /80）完全同构

### 9.5 典型查询

```sql
-- 1. 按 card 查（前端传 card_id，service 层先 cache 反查 spatial_prefix）
SELECT * FROM alarm_events
WHERE tenant_id = $1
  AND device_addr <<= $card_spatial_prefix::inet  -- 从 cache 拿到的 prefix
ORDER BY triggered_at DESC;

-- 2. 按 unit 查
SELECT * FROM alarm_events
WHERE device_addr <<= 'fd00:owl:1:112:42::/80'::inet;

-- 3. 按 device 查（精确）
SELECT * FROM alarm_events
WHERE device_addr = 'fd00:owl:1:112:42:301:a2ac:d523'::inet;

-- 4. 按 device_uid 末段查（同一物理设备的所有历史，跨 rebind）
SELECT * FROM alarm_events
WHERE substring(device_addr::bytea FROM 13 FOR 4) = decode('a2acd523', 'hex');

-- 5. 反查这条 alarm 归哪张卡（cardagg longest-prefix-match）
SELECT card_id FROM cards
WHERE spatial_prefix >>= (SELECT device_addr FROM alarm_events WHERE event_id = $1)
ORDER BY masklen(spatial_prefix) DESC
LIMIT 1;
```

---

## 10. Migration Phases

不大重构，phased 推进。每 phase 独立交付价值，可独立回滚。

### Phase A：基础依赖 + 设计文档（1-2 周，低风险）

- 引入 CloudEvents Go SDK / ULID generator / `net/netip`
- 建 `owl-common/proto/` 仓库，定义 ~10 个 .proto schema
- 重构 `owl-common/spatial/` 为 `net/netip` 薄包装层（删自创字符串实现）
- doc 与 memory 落定（**当前阶段**）

### Phase B：DB schema 加 INET 列 + 回填（1-2 周，低风险）

- 7 张主表加 `spatial_prefix INET`
- alarm_events / iot_timeseries 加 `device_addr INET`
- migration script 递归回填（tenant→branch→site→unit→room→bed→device）
- GiST 索引

### Phase C：Gateway 双发（2 周，低风险）

- qinglan / sleepace 在原 publish 之外**额外**发 CloudEvents over Redis Stream
- 老消费者继续读老路径，新消费者读 CloudEvents 路径
- 双轨过渡，无破坏

### Phase D：分配服务（1.5 周，中风险）

- 写父级→子级 spatial 分配逻辑（应用层）
- 装机流程切到新路径
- Device-host = device_uid 末 32 bit，stateless 派生

### Phase E：Pilot consumer 切换（2 周，中风险）

- 选 cardagg event_handler 作 pilot：消费 CloudEvents + 用 spatial_addr 查询
- 验证 longest-prefix-match 路由正确性

### Phase F：其余 consumer 逐步切换（4-6 周，中风险）

- wisefido-data / sensor / iot-storage 一个一个迁
- 每个迁完观察 1-2 周生产稳定性

### Phase G：老路径下线（2 周，低风险）

- 删 IoTStreamMessage 旧字段
- 清理 SubjectEntity 占位逻辑（unbound device 用 device_id 充 card_id 的 hack）
- 清理 owl-common/redis 老 spatial helpers

### Phase H：Transport（**二期**，待评估）

- 评估 NATS JetStream 替换 Redis Streams 的 ROI
- 短期 Redis Streams 不动；只在 Redis 限制真出现时再做

**总工程量估算**：A-G 约 **3-4 月**（Redis 不换）；含 H 约 **6-9 月**（含运维迁移）。

---

## 11. 关键约束与原则

### 11.1 producer 契约：每条 verdict 必发 monitor + 派生 event/alarm

每条 AI 判定（verdict）必发 monitor channel（track sample = AI 对当下空间状态的事实），event/alarm 是基于此判定的二级派生。

代码层把 emit 收敛到统一入口 `emitVerdict(track, opts)`：必发 monitor，opts.event/alarm 非空时加发。

### 11.2 空间作用域分层

| 层 | 作用域 | 边界 |
|---|---|---|
| firmware | per-device | 单设备 |
| sensor / engine | **per-unit**（含其所有 rooms） | unit_id |
| cardagg | per-card | card_id（1:1 device） |
| UI | per-card | card_id |

**关键不变量**：AI 关联**可以跨房间，不能跨 unit**（HIPAA + 业务无意义）。

### 11.3 TenantID 是硬边界

- producer 端：tenant 段为空 → envelope 拒绝构造
- consumer 端：订阅 prefix 必须以 tenant 起头
- 跨 tenant 查询禁掉

### 11.4 Where 是事实，不是路由

producer 必须诚实填写发生地（数据质量约束）。路由由 transport 决定（pub/sub topic / stream name），不是 IPv6 routing。

### 11.5 SpatialPath 锁深度，纯空间

判断标准：**该维度是不是路由层必需？** 不是 → 不进 path，归 Tag/Subject/Producer。

**已拒绝的扩展**：
- `unit_attr` (vip/home/public) → `Tag: owl.unit-class:*`
- `device_kind` (radar/sleepad) → 已在 source URN 末段
- `cohort` → `Tag: owl.cohort:*`
- `device_role` → Subject 字段或 Producer

### 11.6 card_id 是 cache 不是 snapshot

**snapshot 入库的判断标准**：trigger 时刻 vs 当前是否会不同。

| 字段 | 会变吗 | snapshot 入库 |
|---|---|---|
| device_addr (spatial) | ✓（设备搬移） | ✓ 必须 |
| card_id (= bed/unit UUID) | ✗ 永不变 | ✗ 不入库（cache 反查永远等价） |

类比 IPv6+DNS：IP packet 只带 IP（spatial），不带 hostname（card_id）；resolver cache 做映射。

### 11.7 unbound device

之前 device_id 占位 card_id 是 hack，错。修正后 `alarm_events` 没有 card_id 列；按 device_addr 查询；envelope 里 source 仍正确表达 producer。

### 11.8 已澄清不做

- ❌ Quality / Confidence 在 envelope（per-item，归 payload）
- ❌ Privacy / PHI class 在 envelope（内部 UUID 化已解决）
- ❌ 独立 Stance 字段（用 Tag）
- ❌ SpatialPath 加 unit_attr / device_kind / cohort 段
- ❌ 任何"再加一段表达 X"的扩展（默认拒绝，问"是不是路由必需"）

---

## 12. 待定 / 二期

### 12.1 Transport（二期）

短期：Redis Streams 沿用。

二期评估：NATS JetStream 替换可行性。触发条件 = Redis Streams 出现以下一个：
- 跨数据中心复制需求
- 长期 replay（>30 天）性能问题
- 高频 fan-out 出现广播瓶颈
- 独立运维需求（消息存储与缓存分离）

### 12.2 Schema Registry（远期）

短期：git monorepo `owl-common/proto/`，编译期 codegen。

远期：业务规模到运维 schema 漂移成本时考虑 Buf Schema Registry 或自建。

### 12.3 Multi-agent registry（远期）

agent capability 协商（trust score / SLA / domain）的 control plane。当 cognitive layer 真正接入时再设计。

---

## 13. 参考

- [CloudEvents v1.0.2 spec](https://cloudevents.io/) — envelope 标准
- [RFC 4193 ULA](https://datatracker.ietf.org/doc/html/rfc4193) — IPv6 私有地址
- [RFC 4291 IPv6 addressing](https://datatracker.ietf.org/doc/html/rfc4291) — 寻址
- [W3C Trace Context](https://www.w3.org/TR/trace-context/) — OpenTelemetry trace headers
- [RFC 5424 syslog](https://datatracker.ietf.org/doc/html/rfc5424) — severity levels
- [ULID spec](https://github.com/ulid/spec) — DatagramID 选型
- [Protobuf v3](https://protobuf.dev/) — payload 格式
- [TDPv2-1122](../../owlRD/TDPv2-1122.md) — owl 长期协议愿景
- commit 4a854cb — alarm 5W where snapshot（已落 alarm_events.room_id/unit_id 列）
- commit de183a5 — card_id = 物理 UUID alias

---

## 14. 命名约定速查

| 字段 | 旧 | 新 | 来源 |
|---|---|---|---|
| 消息 ID | 无 | `id` (ULID) | CloudEvents id |
| Spec 版本 | 无 | `specversion: "1.0"` | CloudEvents specversion |
| Producer | 字符串 `"device:UUID"` | `source` URN | CloudEvents source |
| Subject | `SubjectEntity` 字符串 | `subject` | CloudEvents subject |
| 空间 | `SemanticLocation` 单字符串 | `spatialaddr` (IPv6 ULA) | owl extension |
| 时间 | `Timestamp` ms | `time` ISO 8601 | CloudEvents time |
| Schema | 隐式 (TopicType+Category) | `type` (FQN) + Protobuf | CloudEvents type + Protobuf |
| 严重度 | 隐式 | `severity` (0-7) | syslog RFC 5424 |
| 因果链 | 自创 ProducerSeq+CausationID | `traceid`/`spanid`/`parentspan` | W3C TraceContext |
| 撤回 | 无 | `supersedes` | owl extension |
| 跨 agent metadata | 无 | `tags []Tag{Type,Value}` | owl extension |
| Payload | `DataValue` JSON | `data` (Protobuf) | CloudEvents data |
| Card 路由 | SubjectEntity 占位 | spatialaddr longest-prefix-match | owl 派生 |
