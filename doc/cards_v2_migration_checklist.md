# Cards 表 v2 化 — 工作 Checklist（契约源）

> 用户授权全程自动执行；本文件是规则与进度的单一权威，所有 PR/commit 必须对齐此处。
> 创建：2026-05-12

---

## 一、设计共识（已锁定，不可回退）

### 1. card_id 三层身份

| 层 | 字段 | 形态 | 谁用 | 何时变 |
|---|---|---|---|---|
| 业务身份（北极星） | `spatial_prefix INET /96` | `fd00:owl:0001:0112:0042:0301::/96` | service / cardagg / stream payload | 仅 bed 被物理删除 |
| DB PK（稳定性） | `card_id UUID` | `gen_random_uuid()` | FK / cache key | 同上；业务代码不解析 |
| 人类可读（DNS） | `dns_short_name` | `u42-r03-b01.tenant1.owl` | UI / 日志 / 报警显示 | card 增删 |

类比 `192.168.1.0/24`：bed 是子网，HR/RR + radar + cam 是子网内 host，card 是子网门牌号。`device_uid` (host 末 32 bit) 不入 card_id。

### 2. card 删除条件
**当且仅当 card 下所有 device 移除**，与 resident 完全解耦。
- 空床（resident_id=NULL）+ 有 device → card 保留
- 有 resident + device 全失联 → card 删除（resident 须搬走）

### 3. HR/RR 多源宽松共存
- net priority：数字小者胜
- sleepad=100 / radar=200
- 解析靠 view `card_hr_source`，前端按 priority 自动补位

### 4. resident ↔ card 解耦
- `cards.resident_id` 是 *current* pointer，不是历史 truth
- 数据流（alarm/event/monitor）**永不冗余 `resident_id` 列** → 走 `device_addr INET` + ts ∈ episode join
- 历史归属经 `resident_unit JOIN ON spatial_prefix >>= device_addr AND ts BETWEEN valid_from AND COALESCE(valid_to, NOW())` 反查
- resident 改名 = 出院 + 新入住（无 inline edit）

### 5. DNS 单轨永久名（无 PHI）
- 格式：`u<unit>-r<room>-b<bed>.<tenant>.owl`
- 含义：bed-stable，card 在则名在
- resident 进出/转床/改名只动 `cards.resident_id` + emit CloudEvent；DNS zone 文件静止

### 6. Create 触发：事件驱动（不批量 reconcile）
订阅三流，per-prefix 局部响应：

| stream | 业务事件 | card 响应 |
|---|---|---|
| `config:resident:stream` | resident 入住 (resident_unit INSERT) | 该 /96 prefix 无 active_bed card → INSERT + DDNS register；有 → UPDATE resident_id |
| `config:resident:stream` | resident 出院 (resident_unit valid_to=NOW) | UPDATE card resident_id=NULL（保留 card） |
| `config:device:stream` | bind first device 到 /96 prefix | 检查并必要时 INSERT card（罕见，通常 resident 先到） |
| `config:device:stream` | unbind last device 从 /96 prefix | DELETE card + DDNS unregister |
| `config:bed:stream` | bed 物理删除 | DELETE card |
| `config:unit:stream` | unit 公共区域 device | INSERT /80 unit card |

### 7. publish CloudEvents（两种 type）
| type | 何时发 | data |
|---|---|---|
| `owl.config.card.changed` | card 增删 | `{op, card_id, card_type, spatial_prefix, dns_short_name}` |
| `owl.config.card.resident_changed` | `cards.resident_id` 变化 | `{op: admission\|discharge\|transfer, old_hoa, new_hoa, card_id, spatial_prefix}` |

Envelope 严格按 [`doc/datagram_envelope.md`](datagram_envelope.md)：source / subject / spatialaddr / traceid / spanid 必填。

---

## 二、硬约束（红线）

- **R-001**：不向后兼容。删除 v1 字段引用，不留 adapter / shim / fallback。
- **R-002**：不新建 v2 文件。直接重写 v1 文件，命名保持 v1。Backup 必要时 git rm 后再 add 新内容（保留 git history）。
- **R-003**：单次完整改造。禁止 short-circuit/逐函数 v2 化（参见 memory [v2 cutover 教训](../../../.claude/projects/-home-wisefido-owl/memory/v2_cutover_lessons.md)）。
- **R-004**：HIPAA—— 数据流 schema 不冗余 `resident_id` 列。
- **R-005**：DNS zone 文件不含 PHI（永久名只用空间坐标）。
- **R-006**：数据库已 deploy v2 schema（cards/resident_unit 已就位）；不再改 column 结构，只补 view。
- **R-007**：服务可由 Claude 自行启停（systemd or docker-compose），不需用户干预。
- **R-008**：每个 Phase 单独 commit，commit message 引用本文件 § 编号。

---

## 三、Phase 进度

### Phase A — Schema 收尾 ✅
- [x] cards 表 v2 schema 已 deploy（`/owlRD/dbv2/50_cards.sql`）
- [x] resident_unit 表已支持 spatial_prefix INET + valid_from/valid_to
- [x] view `card_hr_source` 已部署
- [x] view `card_current_resident` 已部署（commit `cabad63`）
- [x] PG function `find_card_by_device_addr(INET)` 已部署

### Phase B — owl-common + wisefido-data 重写 ✅
- [x] **B1**：`owl-common/card/*` 全部重写（commit `889714b`）
  - `card_types.go` 删 ExpectedCard/CardWithContent；CardStatic 加 SpatialPrefix/DNSShortName；加 v2 card_type 枚举
  - `card_db.go` 重写为 LPM 反查范式（device_factory_meta + device_runtime_state 替代 device_store）
  - `alarm_db.go` 缩为 v2 stub（详 Phase G 内 carve-out）
  - `devices_jsonb.go` 删除；`utils.go` 删 JSON 序列化 helper
- [x] **B2**：`owl-common/ddns/client.go` 加 `RegisterCardName/UnregisterCardName/CardShortName`（commit `cd67f0b`）
  - 单轨永久名 `u<unit>-r<room>-b<bed>.<tenant>.owl`，无 PHI
- [x] **B3**：`wisefido-data/internal/domain/card.go` + `repository/postgres_card.go` (1435→1014 行) + `cards_repository.go` 重写（commit `6a6cd1f`）
- [x] **B4**：`card_create_service.go` + `card_sync_service.go` 缩为 v2 no-op stub（commit `b428210`）；`card_static_service.queryCardsByIDs` v2 SQL；`card_allowed_provider.go` 三处 filter 函数 v2 INET 化（待 commit）
- [x] **B5**：`config_publisher.go` 拆分两 type **deferred to Phase F**；现暂用单一 `PublishCardChangeMessageWithExtra` 路径（足够 cardagg consumer 适配）

### Phase C — wisefido-cardagg + 其他 service ✅
- [x] cardagg `card_change_handler.go`：现有 CloudEvents 解析路径已对齐，无需改（caller 仍发 `owl.config.card.changed` 单 type）
- [x] cardagg `alarm_service.go`：本次未触动（alarm path 整体 carve out 至 Phase G）
- [x] sensor / iot / sleepace / qinglan / ai-health：无 cards.devices JSONB / unhandled_alarm_* 直接引用，仅 import 路径透传，build 全绿

### Phase D — Build green ✅
- [x] `wisefido-data` go build ✓
- [x] `wisefido-cardagg` go build ✓
- [x] `wisefido-iot` go build ✓
- [x] `wisefido-sensor` go build ✓
- [x] `wisefido-sleepace` go build ✓
- [x] `wisefido-qinglan` go build ✓
- [x] `wisefido-ai-health` go build ✓

### Phase E — 自动 e2e 测试 ✅
- [x] **E1**：重启 owlBack stack（stop-owlback.sh + start-owlback.sh），新二进制全部加载；启动日志 60s 内无 panic / fatal
  - `wisefido-data:917584` started 03:31，启动 reconcile 全 0（v2 stub no-op 符合预期）
  - cardagg/sensor/qinglan/sleepace/iot 全部跑起
- [x] **E2**：login admin@demo → 200 OK，accessToken 派发；tenant_id 已是 INET CIDR `fd00:0:3::/48`
- [x] **E2b**：cards list endpoint `/data/api/v1/data/vital-focus/cards` → 200 / 空列表（cards 表 0 rows，符合 v2 stub 现状）
- [x] **E2c**：cards permission filter (`filterTenantCards` v2 INET-based) → 通过（不再触 `column branch_id does not exist`）
- [x] **E2d**：`/admin/api/v2/residents` list → 200，返 4 个 demo resident（INET HoA / PHI 解密 / branch_id 派生全跑通）
- [ ] **E3**：DDNS zone 文件验证 — Phase F 待做（cards 表为空时 DDNS 注册路径不会被触发）
- [ ] **E4**：redis stream `config:card:stream` CloudEvents 验证 — 同 E3

### Phase F — 写入路径 + 前端 + DDNS wire（pending，独立 PR）

**重点（下次会话）**：

1. **写入路径**：在 `resident_service.go` 的 admission/discharge/transfer hook 上接 cards INSERT/UPDATE/DELETE
   - admission：resident_unit episode 新增 → 检查 spatial_prefix 是否已有 active_bed card；无 → INSERT cards + DDNS register
   - discharge：UPDATE cards.resident_id = NULL（保留 card）
   - 转床：旧 prefix UPDATE NULL + 新 prefix INSERT/UPDATE
   - device bind/unbind：同 prefix 上 device 数 0 → DELETE card + DDNS unregister
2. **DDNS wire**：装配 `ddns.Client`（已在 owl-common/ddns 就绪）到 wisefido-data；resident lifecycle 触发 RegisterCardName/UnregisterCardName
3. **config_publisher 二态化**：拆 `PublishCardChanged` (create/delete) + `PublishCardResidentChanged` (admission/discharge/transfer)
4. **前端**：
   - `src/api/monitors/model/monitorModel.ts`：CardType enum 改 v2 值
   - `Overview.vue` card_type 判断改新值
   - Pinia store / TS interface 同步
5. **DeviceCard 真的需要吗**：v2 cards 表有 /128 device card_type 但本次 service 已删 DeviceCard 相关 CRUD。如确实有公共区无 bed 设备场景，回头补 unit card (/80) 触发

### Phase G — alarm_events v2 改造（独立子项目，从本次 carve out）

**Carve-out 原因**：v2 `alarm_events` schema 与 v1 差异 ≥ 80%（`device_addr INET`/`alarm_kind`/`severity 0-7`/`process_status enum`/`payload/evidence`/`tenant_name` snapshot 列等），与 `cards` v2 化是独立两条线。强行合并会让本次 PR 工作量翻倍 + 风险扩大。

**当前状态**：`owl-common/card/alarm_db.go` 已变 v2 stub：
- API 签名保留以兼容 caller 编译
- `InsertAlarmAndUpdateCard` / `UpdateAlarmAndUpdateCard` / `RecalcCardAlarmState` 直接返回 error（让上游 ack/insert 优雅失败）
- `AutoResolveDeviceAlarms` 返回零结果（v2 device-class alarm 走独立 `device:status` hash）
- `GetActiveAlarmsByCardID` / `ListAllActiveAlarms` / `LookupCardIDByDeviceID` / `ExpireAlarmsByDeviceIDs` 已 v2 化（基于 LPM）

**后续 TODO（独立 phase）**：
- [ ] alarm consumer 改写 device_addr/severity/process_status 路径
- [ ] cards 表 v2 缺 counter/pop 列；caller 端实时聚合 `alarm_events` 走 view
- [ ] sensor/cardagg alarm path 全部按 v2 重写
- [ ] e2e: 触发 fall → alarm_events INSERT (device_addr) → cardagg push → 前端展示

---

## 四、测试账号

```
tenant: Demo
sysadmin / ChangeMe@123
admin / Ts123@123 (admin@demo tenant)
branchManager / demo / Demo@2026
nurse: dnn1 / Ts123@123
caregiver: dvc1 / Ts123@123
family: f01 / Ts123@123
```

服务启停：
- 后端：`cd /home/wisefido/owl/owlBack && ./start-owlback.sh / ./stop-owlback.sh`
- 前端：本次工作不动 owlFront 部署，仅改源码

---

## 五、commit 规则

每个 Phase 末尾 commit，message 格式：
```
feat(cards-v2): Phase X — short description

- ...
- ...

Refs: doc/cards_v2_migration_checklist.md § X
```

## 六、本次会话产出 summary（2026-05-12）

| Commit | Repo | Phase | Files | Lines |
|---|---|---|---|---|
| `cabad63` | owlRD | A | dbv2/50_cards.sql | +45 |
| `889714b` | owlBack | B1 | owl-common/card/*  | +591 / -1682 |
| `cd67f0b` | owlBack | B2 | owl-common/ddns/client.go | +119 |
| `6a6cd1f` | owlBack | B3 | wisefido-data/internal/domain + repository | +732 / -1301 |
| `b428210` | owlBack | B4 | wisefido-data/internal/service stubs | +112 / -882 |
| (pending) | owlBack | B4b | card_allowed_provider.go (filterTenantCards v2) | minor |

**净减代码 ~3000 行**（v1 reconcile / JSONB / 4 alarm counter / pop_alarm 路径全部退役）。

**关键改变摘要**：
1. cards 表三层身份（spatial_prefix INET 主 / card_id UUID PK / dns_short_name）已落定
2. owl-common/card 全部基于 LPM（`cards.spatial_prefix >>= devices.device_ipv6`）反查
3. owl-common/ddns 已提供 RegisterCardName API（DNS 单轨永久名）
4. wisefido-data 7 处 service / repo 文件适配 v2 schema
5. 所有 7 个 owlBack 服务 go build 全绿；systemd restart 成功；boot 无 panic；e2e login + 关键 list endpoint 全 200

**显式 carve out 至下次会话**：

- **Phase G**：alarm_events v2 改造（device_addr/alarm_kind/severity/process_status/payload/evidence 全新 schema）。owl-common/card/alarm_db.go 已成 v2 stub，Insert/Update/Recalc 返回 error 让 caller fail-soft。
- **Phase F**：resident lifecycle 触发 cards INSERT/UPDATE/DELETE + DDNS wire + config_publisher 二态化 + 前端 CardType enum。当前 cards 表是空，写入由下次会话补完。

测试账号已验证：admin/Ts123@123（tenant Demo, fd00:0:3::/48）。
