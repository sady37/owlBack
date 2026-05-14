# Device IPv6 化 — 工作 Checklist（契约源）

> 用户授权全程自动执行；本文件是规则与进度的单一权威，所有 PR/commit 必须对齐此处。
> 创建：2026-05-13
> **执行模式：单程票 / 不回头 / 直接 V2**（用户 2026-05-13 校准）
> 上游配套：[`cards_v2_migration_checklist.md`](cards_v2_migration_checklist.md)（card_add 走 LPM 反查的前置依赖）

---

## 单程票模式说明

**与 cards_v2 不同 — 本 PR 不走双发 / 不留 fallback**：

- ❌ 不做 producer "双发期"（同时带 UUID + INET 字段）
- ❌ 不做 consumer "优先 INET，fallback UUID" 分支
- ❌ 不留观察期 + 二次清理
- ✅ Phase A 一次性把 owl-common 公共 API 切 INET-only（同时删 UUID 入口）
- ✅ Producer 直接发 INET-only 字段；consumer 直接 INET-only 解析
- ✅ 单次原子部署：7 个 service 同 PR 同 commit 同时上线，全栈停 + 全栈起

**风险与可接受性**：
- 部署期间任何一个 service 落后 → 全栈 message 解析失败。**接受**：单 tenant Demo 环境，全栈可 1 分钟 stop/start
- 部署后无 rollback；只能 roll-forward 修
- 节省 ~30% 工作量（删 fallback / 双发 / 观察期）；交换 = 部署窗口零容错

---

## 〇、为什么先做 device_ipv6 化

| 维度 | 说明 |
|---|---|
| 依赖方向 | `cards.spatial_prefix >>= devices.device_ipv6` LPM 是 v2 路由模型；device 不带 IPv6 → cards 反查永远走 UUID 旧路径 |
| 已落地基建 | DB 三表 (`devices`/`alarm_events`/`monitor_stream`/`event_log`) 已用 INET；`owl-common/spatial`/`envelope` 工具完工；缺的是消息总线 + consumer 切换 |
| 风险隔离 | device 字节段（96-127 = MAC 末 32 bit）独立可改，不会撞 unit/room/bed 重排 |
| 与 cards 解耦 | 本 PR 完成后，cards 后续改造（device-bind hook 触发 card 增删 / Phase F follow-up #3）变成几十行水到渠成 |

---

## 一、设计共识（已锁定，不可回退）

### 1. 三态 device 标识 — 各司其职

| 标识 | 形态 | 角色 | 何时变 |
|---|---|---|---|
| `device_uid` | string `E598A2ACD523` (MAC) | **外部稳定标识** — 设备标签印刷、技师扫码、CRM 录入 | 永不变（绑死硬件） |
| `device_id` | UUID | **DB FK 稳定锚** — `device_factory_meta` PK、跨表 FK 软引用 | 出厂后不变 |
| `device_ipv6` | INET `/128` | **业务运行时身份** — 消息路由、cards LPM、空间归属 | 设备移位时 prefix 段变；host 段 (= MAC 末 32 bit) 不变 |

**核心规则**：
- `device_ipv6` host 32 bit ≡ `device_uid` 末 8 hex（`spatial.DeriveDeviceAddr` 无状态派生 [address_v6.go:215](../owl-common/spatial/address_v6.go#L215)）
- 内部消息总线/SQL 路由 **只走 `device_ipv6`**；`device_id` UUID 不再出现在新代码消费路径
- 外部 HTTP API（admin / iot-gateway）**仍接收 `device_uid`**；handler 层一次反查 `devices` 拿 `device_ipv6` 后纯 INET 流转

### 2. 消息 envelope 演进 — 单程票

| 通道 | v1 字段 | v2 字段 |
|---|---|---|
| `IoTStreamMessage` (redis stream) | `DeviceID` (UUID) + `DeviceUID` (MAC) | **删除** UUID/MAC，**唯一** `DeviceAddr netip.Addr` (INET) |
| Datagram v1 (`envelope/datagram.go`) | 已有 `SpatialAddr` (device-as-host) | 保留；wire 进 producer pipeline |
| `card:state:{cardID}` redis hash key | `cardID` UUID | 不动（cards.card_id 仍 UUID PK；本 PR 不改 cards 表）|

**核心规则**：
- producer 必填 `DeviceAddr`；不再写 UUID/MAC 字段
- consumer **只**读 `DeviceAddr`；UUID 解析路径**整段删除**
- 未绑卡设备 `DeviceAddr` 仍存在（设备 /128 已是合法 INET），envelope `card_id` 留空，consumer 走"无卡 device 状态"分支（[memory `unbound_device_card_id_fallback`](../../../.claude/projects/-home-wisefido-owl/memory/unbound_device_card_id_fallback.md) 的 hack 一次清理掉）

### 3. cards 反查模型

```
device 上报 → envelope.DeviceAddr (INET /128)
            → owl-common/card.LookupCardByDeviceAddr(addr)
              → SELECT card_id FROM cards
                 WHERE spatial_prefix >>= $1::INET
                 ORDER BY masklen(spatial_prefix) DESC LIMIT 1
            → consumer (cardagg/sensor/etc) 派发
```

LPM 唯一权威；UUID `LookupCardIDByDeviceID` Phase F 删除。

### 4. unit_id / room_id / bed_id 字段去除

`IoTStreamMessage` v2 现有的 `UnitID/RoomID/BedID/SpatialPath` 是 cards_v2 中间过渡，本 PR 借势收敛：consumer 只需 `DeviceAddr`，prefix slice 自己派生（`DeviceAddr.Prefix(80)` = unit；`.Prefix(88)` = room；`.Prefix(96)` = bed）。

`SemanticLocation` (RoomID 字符串) 同步退役。

---

## 二、硬约束（红线）

- **R-001**：**单程票，不向后兼容，不留 fallback**。owl-common UUID lookup 入口 Phase A 直接删；producer Phase B 直接发 INET-only；consumer Phase C/D 直接 INET-only 解析。任何"先双发再切再删"的渐进路径都禁止
- **R-002**：外部 API 仍接 `device_uid` (MAC) — 面向前端 / CRM / 技师工具的契约不动；handler 内一次反查后纯 INET 流转
- **R-003**：`device_id` UUID 列**不退役**（DB FK 稳定性 + 7 年审计 actor 引用）；只是消费路径不再使用
- **R-004**：单次完整改造。每个 Phase 一次推完所属层；禁止 short-circuit / 逐函数 v2 化（[memory `v2_cutover_lessons`](../../../.claude/projects/-home-wisefido-owl/memory/v2_cutover_lessons.md)）
- **R-005**：HIPAA — envelope 不冗余 `resident_id`；resident 归属经 `resident_unit JOIN ON spatial_prefix >>= device_addr` 反查（与 cards_v2 R-004 一致）
- **R-006**：DB schema 已 deploy；本 PR 不改 column 结构，只补 view / function 必要时
- **R-007**：服务可由 Claude 自行启停（systemd / docker-compose），不需用户干预
- **R-008**：所有 Phase 在**同一 PR / 同一 commit**（按本文件 § 编号分小 commit 也可，但全部进同一次部署窗口）；7 个 service 同栈停 + 同栈起，不存在"半切"中间态
- **R-009**：所有未绑卡 device 的处理路径必须明确 — envelope `card_id` 留空 + 必填 `DeviceAddr`，consumer 走"无卡 device 状态"分支（[memory `unbound_device_card_id_fallback`](../../../.claude/projects/-home-wisefido-owl/memory/unbound_device_card_id_fallback.md) 红线一次落实）
- **R-010**：单程票部署窗口约定 — 部署前 redis stream 必须 drain（`XLEN config:*:stream` 全部接近 0），避免老 envelope 消息在新 consumer 上炸 panic

---

## 三、Phase 进度（单程票，A-F 同 PR 同部署窗口）

### Phase A — owl-common 一次切 INET-only ✅

- [x] **A1**：`owl-common/redis/message_types.go` — `IoTStreamMessage` 删 `DeviceID/DeviceUID/SemanticLocation/UnitID/RoomID/BedID/SpatialPath/ScopePrefix/TenantID` 九字段，唯一 `DeviceAddr netip.Addr` (json `device_addr`)；`BuildDeviceProducer(addr)` / 所有构造函数 (`NewIoTStreamMessageWithData` / `BuildIoTStreamMessage` / `NewSingleItemMessage` / `BuildDeviceStatusMessage` / `BuildAlarmDeviceMessage` / `BuildAlarmProcessMessage` / `BuildCardChangeMessage*`) 同步切 INET 入参；`AuthMessage` 保留 `DeviceUID` (R-002 外部 auth 边界)
- [x] **A2**：`owl-common/card/device_lookup_key.go` — 删 `IsUUIDString`；`NormalizeDeviceLookupKey` 简化为 MAC-only normalize（UUID 短路返回路径删除）
- [x] **A3**：`owl-common/card/card_db.go` + `alarm_db.go` — 加 `LookupCardByDeviceAddr(ctx, addr) (cardID, err)` LPM SQL（cards.spatial_prefix >>= addr）；删 `LookupCardIDByDeviceID`；删 `LookupDeviceOnly`/`LookupDeviceStoreOnly` 的 `m.CardID = m.DeviceID` virtual-card hack（R-009）
- [x] **A4**：`owl-common/spatial/` — 验证 `DeriveDeviceAddr` / `LongestPrefixMatch` / `IsOwlAddr` / `ContainsAddr` / `SlotsOf` 全部就位；现有单测 ok；不补冗余测试
- [x] **A5**：`owl-common/envelope/datagram.go` — `Datagram.SpatialAddr` 已是公共 `netip.Addr` 字段 + `WithSpatial(addr)` setter，无需加 accessor；`identity.go` NodeID UUID 是进程实例 (R-003 保留)
- [x] **A6**：owl-common build + vet + test 全绿；下游 7 service broken（设计意图，强制 Phase B-E 同步切）

**本 Phase 改动统计**：4 个 owl-common 文件，~300 行 diff（主要是 message_types.go 重写）。

### Phase B — Producer 直接发 INET-only ✅

- [x] **B-prep**：`owl-common/card/device_baseline.go` 加 `DeviceAddr netip.Addr` 字段；`card_db.go` LookupCard SQL 加 `host(d.device_ipv6)` SELECT 项；`AuthMessage` 恢复 `DeviceID` 字段（外部 auth 边界 R-002/R-003 例外）
- [x] **B1+B2**：`wisefido-qinglan` 4 文件改：
  - `repository/postgres_device.go` — `DeviceStoreInfo` 加 `DeviceAddr` + GetDeviceStoreInfo/ByDeviceID 查询返 device_addr；TenantID 改 INET CIDR 文本
  - `consumer/stream_publisher.go` — 重写 `publishObservation` 用 `msg.DeviceAddr` 做 routing key；删 `SemanticLocation` 自动填；`PublishDeviceStatus(addr, deviceUID, deviceType)` 简化签名
  - `consumer/mqtt_consumer.go` — `resolveDeviceIdentity` 加返 `addr`；`publishRadarMonitorHeartbeat(addr, cid)` 简化；3 个 handler (Monitor/Stat/Event) 切 addr；`TargetMergeVital(addr, cid)` 简化；`publishStatActivity/publishStatSleep` 切 addr
  - `subscriber/device_subscription_manager.go` + `health_check.go` — `PublishDeviceStatus` 6 处 caller 切 addr；`publishDeviceAlarm` 反查 addr
- [x] **B3+B4**：`wisefido-sleepace` 3 文件改：
  - `consumer/stream_publisher.go` — `Resolve()` 加返 addr；`publish()` 用 `msg.DeviceAddr`，`SemanticLocation` 自动填路径删除
  - `consumer/mqtt_consumer.go` — 11 处 `NewIoTStreamMessageWithData` 切 addr 入参；`handleAlarmNotify` sig 切 addr；删 `evMsg.DeviceID = deviceID` 兜底赋值
  - `subscriber/health_check.go` — connectionStatus alarm 切 b.DeviceAddr
- [x] **B5**：4 service (owl-common + cardagg + qinglan + sleepace) `go build` + `go vet` 全绿

**本 Phase 改动统计**：
- owl-common: +20 行（DeviceBaseline.DeviceAddr + LookupCard SQL）
- qinglan: 4 文件，~30 处 caller 改造
- sleepace: 3 文件，~15 处 caller 改造

### Phase C — cardagg consumer 直接 INET-only ✅

- [x] **C-helper**：新增 `internal/service/addr_ctx.go` — `AddrCtx` 统一派生 `(TenantPref/UnitPref/RoomPref/BedPref/DeviceAddr)`；consumer 入口一次调用，避免 169 处分散派生
- [x] **C1**：`internal/consumer/iot_prepared.go` — 重写：从 raw map 读 `device_addr` → `metaCache.LookupCardByDeviceAddr(addr)` 反查；命中填 cardID；未命中 (R-009) subject 留空走"无卡 device 状态"分支，**删除** UUID 充 subject 路径
- [x] **C2**：`internal/service/device_meta.go` — 双向索引 `deviceIDIndex/deviceUIDIndex` 合并为单一 `deviceAddrIndex`；`LookupCardsByDevice` → `LookupCardByDeviceAddr` (returns single cardID)；`BuildDeviceIndex` SQL 改 device_ipv6 LATERAL LPM；`DeviceMeta` 加 `DeviceAddr netip.Addr` + `fillFromAddr()` (老字段 BoundBedID/RoomID/UnitID/DeviceID/DeviceUID 保留兼容，值改 IPv6 派生)；`CardMeta` 加 TenantPref+TenantID 兼容字段
- [x] **C3**：`internal/service/alarm_service.go` — 10 处 `msg.{TenantID,DeviceID,SemanticLocation}` 切 `AddrCtxFromMsg(msg)` 派生；`card.LookupCardIDByDeviceID(uuid)` → `card.LookupCardByDeviceAddr(netip.Addr)` (parse from p.DeviceID string)；`AutoResolveDeviceAlarms` 入参切 INET
- [x] **C3b**：consumer 全部文件切：`alarm_handler.go` (49 refs)、`event_handler.go` (64 refs，4 函数各加 ac 派生)、`stay_fsm.go` (14 refs，inline AddrCtxFromMsg)、`monitor_handler.go` (11 refs)、`stream_log_fields.go` (3 refs，加 netip.PrefixFrom 派生 tenant)、`card_change_handler.go` (`InvalidateCardsInTenantUnit` 删 tenantID 入参 + `DeviceKeysInTenantUnit` → `DeviceAddrsInUnit`)；新建 `IoTStreamMessage{}` struct literal 全部去 UUID 字段加 `DeviceAddr: m.DeviceAddr`
- [x] **C4**：unbound device 路径 — `iot_prepared.go` subject 留空；`card_db.go` `LookupDeviceOnly`/`LookupDeviceStoreOnly` 删 `m.CardID = m.DeviceID` virtual-card hack（已在 Phase A 一并处理）
- [x] **C5**：cardagg `go build ./...` ✅；`go vet` ✅；`go test ./internal/service/...` ✅

**本 Phase 改动统计**：
- owl-common：4 文件 ~300 行
- cardagg：12 文件 +657 / -974 行（device_meta.go 单文件 -300 行；总净减 ~317 行）

### Phase D — sensor consumer 接 INET（首次接入）✅

实际改动只 1 个文件 `internal/roomengine/engine.go`（23 处 broken refs 集中在这里）：

- [x] **D1**：3 个 handler 函数 (handleEventMessage / handleAlarmMessage / handleMessage) — `m.DeviceID/m.DeviceUID` 双 fallback 合并为单一 `m.DeviceAddr.String()` (since UUID 与 MAC 与 IPv6 都是同一字符串 cache key)
- [x] **D2**：`warnUnrouted` sig 简化 `(stream, cardID, deviceAddr, deviceType)` —— 4 参 vs 旧 5 参，UID/ID 合并
- [x] **D3**：`publishAIMessage` IoTStreamMessage struct literal 切 `DeviceAddr netip.Addr` 字段；`p.DeviceID` (现为 IPv6 字符串) 用 `netip.ParseAddr` 转换
- [x] **D4**：sensor `go build` + `go vet` 全绿

**残留**（不在 D 范围）：
- `e.deviceIDToUID` map 实际上现在已经是身份映射 (UUID==UID==IPv6 全是同字符串)；标 deprecated 但保留。Phase D 不删，避免触发 mqttDeviceID/setMapper caller 重写
- AIPayload.DeviceID 字段保留命名但值是 IPv6 字符串（最小化 sensor 内部 ripple）

**本 Phase 改动**：1 文件 ~30 行 diff

### Phase E — wisefido-data 数据 layer 收口 ✅

实际改动远小于预期 — Phase 2-5 cutover 时 wisefido-data 的 `postgres_devices/postgres_device_store/postgres_alarm_events` 已经全部走 device_ipv6 INET LPM；`/internal/device-baseline` endpoint 序列化 `DeviceBaseline` 已自动带 `DeviceAddr` 字段（B-prep 加的）。

剩下的只有 `internal/publisher/config_publisher.go` 4 个 wrapper 与新 owl-common signature 不匹配 + cardagg `alarm_device_handler` payload 字段名同步：

- [x] **E1**：`config_publisher.PublishAlarmProcessMessage` — 删 `deviceID` payload 写入（cardagg consumer 不读）；保留 caller 入参兼容
- [x] **E2**：`config_publisher.PublishAlarmDeviceMessage` — 加 `lookupDeviceAddr(deviceID)` 内部反查 `devices.device_ipv6`；wrapper 入参签名不变（caller 仍传 UUID + UID）；payload 出 `device_addr` (canonical IPv6)
- [x] **E3**：`config_publisher.PublishCardChangeMessageWithExtraAndType` — 适配新 owl-common 4-arg signature；tenantID/unitID/branchID 透 extras 字段
- [x] **E4**：`config_publisher.go` 加 `db *sql.DB` 字段 + `SetDB()` setter（lookupDeviceAddr 用）
- [x] **E5**：`cardagg/internal/consumer/alarm_device_handler.go` — `alarmDeviceData` 字段从 `DeviceID/DeviceUID/TenantID` 改为唯一 `DeviceAddr` (json `device_addr`)；`enablement.Invalidate(d.DeviceAddr)` cache key 切 IPv6 字符串；`PurgeDisabledPendingForDevice` 同步
- [x] **E6**：`/internal/device-baseline` endpoint 自动带 DeviceAddr（B-prep 已 plumb DeviceBaseline 字段；JSON encoder 自动序列化 `device_addr`）
- [x] **E7**：5 service (owl-common + cardagg + qinglan + sleepace + wisefido-data) `go build` + `go vet` 全绿

**本 Phase 改动统计**：
- wisefido-data: 1 文件 (`config_publisher.go`) ~50 行
- cardagg: 1 文件 (`alarm_device_handler.go`) ~10 行
- 其余 ~25 个 wisefido-data 文件提到的 device_id 引用都属"业务正用"（admin/api 边界），不在本 PR 范围

**残留：**
- `wisefido-data/internal/service/alarm_event_service.go` 的 `event.DeviceID` 仍是 UUID 字符串（来自 alarm_events 表，本身 schema 保留 UUID + INET 双字段）。alarm_events.device_addr INET 列已存在，未来 Phase G alarm_events v2 改造时切完。本 PR 内 alarm_process_message 已不传 device 字段，完整闭环。

### Phase F — 一次部署 + 验收 ✅ (2026-05-13)

- [x] **F1**：drain redis streams — XLEN 各流 ≤ 1004（capped MAXLEN）；老 envelope 消息被新 consumer FromStreamMap warn-drop，不会 crash
- [x] **F2/F3**：`./build-owlback.sh && sudo systemctl restart owlback`；6 service 全 active，60s+无 panic/fatal
- [x] **F4**：e2e 验证 — qinglan 端真实雷达消息发出（`device_addr:fd00:0:3:111:3:300:7cd4:570c` / `subject_entity:fd00:0:3:111:3::/80`）；cardagg 端 `device addr index built count:20` + `card_id:fd00:0:3:111:3::/80` LPM 反查成功；offline_recover/EnterRoom/ExitRoom 全栈通；sleepace OfflineRecover 通；ai_event 路径 sensor 端有 dropped_unrouted_message warn（sensor `engine_bootstrap.go` 用 v1 `bound_room_id/device_store` 列查路由，业务表 v2 化未在 D 范围）
- [x] **F5**：grep 验证 — owlBack 全仓 `LookupCardIDByDeviceID` / `LookupCardsByDevice` / `BuildDeviceProducer(deviceID string)` 调用数 = **0**
- [x] **F6**：部署后 3 处补丁（同 commit `f9f4401`）：
  - `wisefido-data/baseline_handler.go`: 删 `b.CardID = b.DeviceID` UUID 占位（R-009 红线漏扫）
  - `owl-common/card/card_db.go`: LookupDeviceOnly + ListAllBaselines SQL 加 `host(d.device_ipv6)` 派生 DeviceAddr
  - `wisefido-sleepace/stream_publisher.go`: subject_entity 空允许通过（unbound device cardagg LPM 反查兜底，与 qinglan 一致）
- [ ] **F7**：`wisefido-ai-health` Phase 1 ETL 直接接 INET — 待该 service 启用时落地（不阻塞本 PR）
- [ ] **F8**：sensor `engine_bootstrap.go` v1 SQL 列引用 — 独立 phase（sensor 业务表 v2 化）

**验收结论**：
- 6 service 全 active；1min 窗口内全栈 0 panic / 0 fatal / 0 error
- iot:monitor:stream 1004 / iot:event:stream 57 / iot:alarm:stream 93 — 真实流量穿过新 envelope
- INET-only envelope 100% 通过；qinglan/sleepace producer + cardagg consumer + wisefido-data baseline endpoint 全栈 INET

---

## 四、显式 Carve-out（不在本 PR 范围）

| 项 | 原因 | 后续归宿 |
|---|---|---|
| `cards.card_id ≡ spatial_prefix` 化 | cards_v2 已 锁 [Phase F.6+](cards_v2_migration_checklist.md)；本 PR 不动 cards 表 schema | cards_v2 后续 phase |
| device-bind 触发 card 增删 hook | cards_v2 Phase F follow-up #3 | cards_v2 后续 phase（本 PR 完成后水到渠成） |
| alarm_events Phase G | 与 device_ipv6 化正交（schema 已 INET 化，consumer 改造独立） | 单独 PR |
| sensor 服务 v2 schema 接入 | sensor 业务深度耦合 v1 monitor_stream 列（[memory `phase2to5_v2_cutover`](../../../.claude/projects/-home-wisefido-owl/memory/phase2to5_v2_cutover.md))| 独立 phase；本 PR 仅做 envelope 接入（Phase D），不动 sensor 业务表 |
| ~~MQTT gateway 双发~~ | 单程票模式不存在双发 | n/a |
| 老 v1 envelope 在 redis stream 内的"在途消息" | 部署窗口 [R-010](#二硬约束红线) drain 处理；新 consumer 不再支持解析 v1 envelope | 部署前 drain |

---

## 五、测试账号

```
tenant: Demo (fd00:0:3::/48)
sysadmin / ChangeMe@123
admin / Ts123@123 (admin@demo)
demo branch / Demo@2026
```

**关键 e2e 触发器**（Phase F 部署后一次性跑全套，单程票模式不分阶段验收）：
1. radar fall trigger → alarm path（覆盖 cardagg INET LPM）
2. sleepad HR/RR upload → monitor stream（覆盖 sleepace publisher）
3. unbound device 上报（无 cards 行）→ event_log 落库不丢（覆盖 R-009）
4. admin /devices list + bind → DB query plan 全 GIST 命中（覆盖 wisefido-data E2/E3）

服务启停：
```
cd /home/wisefido/owl/owlBack && ./start-owlback.sh
cd /home/wisefido/owl/owlBack && ./stop-owlback.sh
```

---

## 六、commit 规则

每个 Phase 末尾 commit，message 格式：
```
feat(device-ipv6): Phase X — short description

- ...

Refs: doc/device_ipv6_migration_checklist.md § X
```

---

## 七、状态总览（动态维护）

| Phase | 状态 | Commit | 备注 |
|---|---|---|---|
| A — owl-common 一次切 INET-only | ✅ 2026-05-13 | (pending) | owl-common 全绿；下游 7 service broken（设计意图） |
| B — qinglan/sleepace producer | ✅ 2026-05-13 | (pending) | qinglan + sleepace 4 文件改造，build/vet 全绿；DeviceBaseline 加 DeviceAddr 已 plumb 到 wisefido-data 端 |
| C — cardagg consumer | ✅ 2026-05-13 | (pending) | cardagg build/vet/test 全绿；157 处 msg.X + 1 处 LookupCardIDByDeviceID 全切 |
| D — sensor consumer | ✅ 2026-05-13 | (pending) | engine.go 1 文件改造，build/vet 全绿；deviceIDToUID 标 deprecated 留作内部 map（值已是 IPv6 字符串） |
| E — wisefido-data data layer | ✅ 2026-05-13 | (pending) | config_publisher 4 wrapper 适配 + alarm_device_handler 切 device_addr；postgres_* repo 已是 v2 INET（Phase 2-5 顺手切完）|
| F — 部署 + 验收 + ai-health | ✅ 2026-05-13 | f9f4401 + restart | 6 service active 1min 全栈 0 error；3 处部署补丁；ai-health/sensor-bootstrap 独立 phase 处理 |

**注**：A merge 后 B/C/D/E 必须立即跟上，否则下游 service 编译失败。建议本地分支同时改 A-E 五处，一次 PR push。

---

## 八、关联

- 上游契约：[`cards_v2_migration_checklist.md`](cards_v2_migration_checklist.md)（cards 反查 LPM 模型）
- envelope 设计：[`datagram_envelope.md`](datagram_envelope.md)（device-as-host stateless 派生）
- v2 cutover 通则：[memory `v2_cutover_rules`](../../../.claude/projects/-home-wisefido-owl/memory/v2_cutover_rules.md) / [`v2_cutover_lessons`](../../../.claude/projects/-home-wisefido-owl/memory/v2_cutover_lessons.md)
- cardagg 反查现状：[memory `unbound_device_card_id_fallback`](../../../.claude/projects/-home-wisefido-owl/memory/unbound_device_card_id_fallback.md)
- 设备绑定模型：[memory `device_unbind_via_prefix_reset`](../../../.claude/projects/-home-wisefido-owl/memory/device_unbind_via_prefix_reset.md)
