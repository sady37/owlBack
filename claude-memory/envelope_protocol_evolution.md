---
name: Envelope 协议演进 — Datagram v1（CloudEvents + IPv6 ULA + Protobuf）
description: TDP Datagram v1 设计定稿 2026-05-07；不重造轮子，envelope=CloudEvents/寻址=IPv6 ULA/payload=Protobuf/trace=W3C/severity=syslog；device-as-host stateless 派生；详 doc/datagram_envelope.md
type: project
originSessionId: d0f8e7ff-d83c-4c38-a1b0-a05ce44699e9
---
**用户 2026-05-07 终稿校准**：把 envelope/alarm/event 协议作 IPv6 标准化，**不重造轮子，借标准**。

衔接：[agent_pipeline_north_star.md](agent_pipeline_north_star.md) 的 TDPv2 北极星——这条是它的具体投影。

详 [owlBack/doc/datagram_envelope.md](../../../owl/owlBack/doc/datagram_envelope.md)（设计定稿）。

## 核心决策（设计定稿）

**不自创，借这套标准**：

| 关注点 | 借的标准 |
|---|---|
| Envelope 格式 | **CloudEvents v1.0**（CNCF）|
| 空间寻址 | **IPv6 ULA `fd00:owl:`**（RFC 4193 + RFC 4291）|
| 寻址实现 | **`net/netip`** Go 标准库 + Postgres **INET** |
| Payload | **Protobuf**（`owl-common/proto/` git monorepo）|
| Trace / 因果 | **OpenTelemetry W3C TraceContext** |
| 严重度 | **syslog RFC 5424** (0-7) |
| 消息 ID | **ULID** |
| Tags | **`{Type, Value}` dot-namespaced**（不绑 FHIR；FHIR 是众多 namespace 之一）|
| Transport | **二期决定**（短期 Redis Streams 沿用，远期评估 NATS JetStream）|
| DB 主键 | UUID 不动；新加 `spatial_addr INET` 列 + GiST 索引 |

## IPv6 寻址布局（128 bit）

```
fd00 : owl  : TTTT : BB SS : UUUU : RR BB : DDDD DDDD
└──┘  └──┘  └──┘   └─┬─┘   └──┘   └─┬─┘  └──── 32 bit ────┘
ULA   owl   tenant  br/site unit   rm bd   = device_uid 末 32 bit
```

| 段 | 位宽 | 起始位 |
|---|---|---|
| ULA `fd00:owl:` | 32 | 0-31 |
| Tenant | 16 | 32-47 (= /48) |
| Branch | 8 | 48-55 (= /56) |
| Site (Bldg+Floor 4+4 split) | 8 | 56-63 (= /64 subnet 边界) |
| Unit | 16 | 64-79 (= /80) |
| Room | 8 | 80-87 (= /88) |
| Bed | 8 | 88-95 (= /96) |
| **Device-host** | **32** | **96-127 (stateless 派生 = device_uid 末 32 bit)** |

例：`fd00:owl:0001:0112:0042:0301:A2AC:D523`
- Tenant=0001, Branch=01, Site=12 (Bldg=1 Floor=2), Unit=0042, Room=03, Bed=01
- Device-host=A2ACD523（= device_uid `E598A2ACD523` 末 8 hex）

## Device-as-host 是关键设计

- Device 是 IPv6 host (lowest 32 bit)
- Host 部分 = device_uid 末 32 bit，**stateless 派生**，无须分配协议
- 设备搬房间 → prefix 变，host 不变（跟真 IPv6 mobility 同构）
- device_id UUID 仅作 DB PK 稳定性，业务层退役

## card_id 是 cache 不是 snapshot

- DB 表（alarm_events / iot_timeseries）**不存 card_id 列**，存 `device_addr INET`
- cards 表是单一权威源（spatial_prefix INET ↔ card_id UUID 映射）
- 各服务启动加载 spatial→card_id 内存 map，cardChange stream 增量 invalidation
- 路由 = Postgres INET longest-prefix-match (`>>=` 操作符 + GIST index)
- 类比 IPv6+DNS：IP packet 只带 IP，不带 hostname；resolver cache 做映射

## Trace-based replay：Datagram v1 的关键业务收益

producer 契约（每条 AI verdict 必发 monitor + 派生 event/alarm）+ traceid + parentspan 解决之前 D523 fall replay 痛点：

- AI verdict 在 monitor 流第一公民（pose=5 fall icon 自动渲染）
- replay 一句 `WHERE traceid = X` 拿完整因果树（firmware tracks → AI synth verdict → alarm → user ack）
- 不再需要 trigger_data.evidence.replay_anchor_ms 救场（AI synth datagram 自带 anchor 时刻）
- 不再需要"AI overlay 要不要也回放 card:alarmstatus:stream"特殊 infra
- 多次临时修补 → 统一架构吸收

详 [doc §8.5](../../../owl/owlBack/doc/datagram_envelope.md)。

## Producer identity 三层（service / instance / message）

每个 service 表达自己产生的 datagram 时，三层身份各占字段：

| 层 | 字段 | 例 |
|---|---|---|
| 软件类型 + 版本 | `source` URN | `sensor://owl.engine.lost_fall/v1.0.0` |
| 进程实例（可选） | `nodeid` extension | `uuid-of-this-process` |
| 这条消息 | `id` (= ULID) | `01HXY3K7G2...` |

**ULID 一物多用**：每条 datagram 自己 mint ULID 同时是 CloudEvents `id` + W3C `spanid` + idempotency key + parentspan 引用目标。16 字节天然兼容 W3C span_id 格式，**不需要单独再生成 span_id**。

ULID 80 bit 随机段保证：每秒 1 亿条消息撞 ID 期望 ≈ 3 亿年。无须中央协调，每个 service 自己 mint 即可。

## spatial_prefix 是通用 key（不只 cards）

**所有"按空间范围生效的资源"都走同一套 longest-prefix-match 模式**：

| 资源 | spatial_prefix 深度 |
|---|---|
| Layout | /88 (room) 或 /80 (unit 多 room 合一) |
| Wall polygon | /88 |
| Grid 学习状态 | /88 |
| Cell 语义标签 (AI_fall) | /88 |
| Card | /96 (ActiveBed) / /80 (Unit) |
| 通知策略 | /80 unit / /48 tenant |

- 自然支持"父级 default + 子级 override"
- 解决 WaveMonitor 老注释"未来 bedroom+bookroom 合一 layout"问题
- 跟 cards multi-card 共存（IPv6 /48 与 /64 共存）完全同构
- 详 [doc/datagram_envelope.md §9.4](../../../owl/owlBack/doc/datagram_envelope.md)

## Tags 用途

**放不进 envelope 结构性字段、又跨多 agent 通用的"分类元数据"**。`{Type, Value}` 两字段，dot-namespaced。

6 类典型用途：消息生命周期 / subject 分类 / ML 训练标签 / 领域功能 / 灰度 / 多 agent 协同。

不应放 Tag：自由文本、数值测量、标识符、严重度、位置、时间、因果。

namespace 约定：`owl.*` (内置) / `ml.*` (训练) / `ops.*` (运维) / `agent.<name>.*` (定制) / `external.<vocab>.*` (FHIR/SNOMED 等) / `tenant.<id>.*` (私有)

## 三个核心契约

### 1. Producer 契约（co-emit monitor + 派生）

每条 AI verdict 必发 monitor channel（track sample = AI 对当下空间的事实），event/alarm 是基于此判定的二级派生。代码层把 emit 收敛到 `emitVerdict(track, opts)`。

### 2. 空间作用域分层

| 层 | 作用域 | 边界 |
|---|---|---|
| firmware | per-device | 单设备 |
| sensor / engine | **per-unit**（含其所有 rooms）| unit_id |
| cardagg | per-card | card_id（1:1 device）|
| UI | per-card | card_id |

**关键不变量**：AI 关联**可以跨房间，不能跨 unit**（HIPAA + 业务无意义）。

### 3. SpatialPath 锁深度，纯空间

判断标准：**该维度是不是路由层必需？** 否 → 进 Tag/Subject/Producer，不进 path。

已拒绝的扩展：unit_attr / device_kind / cohort / device_role 都归 Tag 或 Subject。

## Migration phasing（A-G ~3-4 月，含 H 二期 ~6-9 月）

- Phase A：基础依赖 + 重构 spatial 包为 net/netip 薄包装（1-2 周）
- Phase B：DB schema 加 INET 列 + 回填（1-2 周）
- Phase C：Gateway 双发 CloudEvents（2 周）
- Phase D：分配服务（1.5 周）
- Phase E：Pilot consumer 切换（2 周）
- Phase F：其余 consumer 逐步切换（4-6 周）
- Phase G：老路径下线（2 周）
- **Phase H：Transport（二期，待评估）**——NATS JetStream vs Redis Streams ROI

## 已澄清不做

- ❌ Quality / Confidence 在 envelope（per-item，归 payload）
- ❌ Privacy / PHI class 在 envelope（内部 UUID 化已解决）
- ❌ 独立 Stance 字段（用 Tag `owl.stance:*`）
- ❌ SpatialPath 加 unit_attr / device_kind / cohort 段
- ❌ alarm_events 入库 card_id 列（用 device_addr + cards cache 反查）
- ❌ unbound device 用 device_id 占位 card_id（清理 hack）

## 当前实施状态（2026-05-08 update）

- ✅ Phase 1 落地：旧 `owl-common/spatial/spatial.go` 字符串 path 保留作过渡
- ✅ **Phase A 完成（2026-05-08 push 到 main）**：
  - `owl-common/spatial/address_v6.go` — IPv6 ULA + netip.Addr，全套 builder/parser/LPM/ReverseDNS
  - `owl-common/spatial/subject.go` — Mobile IPv6 风格 resident/caregiver HoA（详 [subject_addressing.md](subject_addressing.md)）
  - `owl-common/envelope/` — Datagram + ServiceIdentity + ULID + CloudEvents v1.0 binary mode
  - 38 unit tests + 2 e2e roundtrip 全过
  - 100% 标准 IPv6（无私有 sentinel），所有工业 IPv6 工具直接 work
- ⏳ Phase B 起：DB schema 加 INET 列 + creation flow 接 spatial v2

旧 `spatial.go` 字符串 path API 与新 IPv6 v2 共存；老调用方 `owl-common/redis/message_types.go` 的 SpatialPath 字段仍走老 API。

## 何时这条不再适用

- 当 TDPv2 真正落地，这条作为过渡指南可归档
- 当 cognitive layer (LLM agent) 接入，envelope 应已升级到 Datagram v1
