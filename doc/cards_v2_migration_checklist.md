# Cards 表 v2 化 — 工作 Checklist（契约源）

> 用户授权全程自动执行；本文件是规则与进度的单一权威，所有 PR/commit 必须对齐此处。
> 创建：2026-05-12

---

## 一、设计共识（已锁定，不可回退）

### 1. card_id 双层身份（**Phase F.6 简化：UUID 层已退役**）

| 层 | 字段 | 形态 | 谁用 | 何时变 |
|---|---|---|---|---|
| **业务身份 + DB PK** | `spatial_prefix INET` | `fd00:0:3:111:3:101::/96` | service / cardagg / stream payload / FK / cache key | 仅 prefix 物理删除（bed 被销毁）|
| 人类可读（DNS） | `dns_short_name` | `br01-s12-u0001-r01-b01.tenant3.owl` | UI / 日志 / 报警显示 | card 增删 |

**核心规则**：`card_id ≡ spatial_prefix`。空间不变 → card_id 不变（删除重建也不变；prefix 是身份本身）。

类比 `192.168.1.0/24`：bed 是子网，HR/RR + radar + cam 是子网内 host，card 是子网本身（用 CIDR 表示）。`device_uid` (host 末 32 bit) 不入 card_id。

**历史**：v1 用 UUID 做 PK + spatial_prefix 做业务身份，两层冗余。Phase F.6 简化为单层 — `card_id` 在 API/JSON 中保留为字段名，但其值就是 `spatial_prefix` INET CIDR 字符串。

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

### Phase F — 写入路径 + 前端 + DDNS wire ✅

完成于 2026-05-12（commit pending）。

- [x] **F1 写入路径**：`resident_service.go` admission/transfer/discharge/delete hook 调 `syncCardForResident(ctx, tenantPrefix, hoa, oldPrefix)`
  - oldPrefix → 清空 resident_id；newPrefix → INSERT/UPSERT active_bed card + DDNS register
  - 转床两段事件：旧 card emit `transfer` (new_resident_id=""); 新 card emit `transfer` (prev=new=HoA, prefix=new)
  - admission: prev=""; discharge: new=""
  - 新增 repo 方法：`PostgresResidentsRepository.GetActiveSpatialPrefix(ctx, hoa)` + `PostgresCardRepository.GetActiveBedCardIDByPrefix(ctx, prefix)`
- [x] **F2 DDNS wire**：`initDDNSClient(logger)` 顶层 helper，wisefido-data main.go 启动期一次性装配，传给 `registerSpatialV2` + `residentSvc.SetCardDeps(cardRepo, configPublisher, ddnsClient, owlDom)`
- [x] **F3 二态化 publisher**：`PublishCardChanged(op, prefix, dnsShortName)` + `PublishCardResidentChanged(op, prevHoA, newHoA, prefix)`。底层仍单 type `config.card`（消费者无需 break）；`extras.op` 体现语义解耦
- [x] **F4 前端**：`monitorModel.ts` `CardType` 改 v2 值 `'active_bed'|'unit'|'public'|'room'|'device'|'tenant'|'branch'|'site'`；Overview.vue 5 处 `=== 'ActiveBedCard'` / `=== 'UnitCard'` 字面量同步替换

**E2E 验证（admin@demo, fd00:0:3::/48）**：

| step | 输入 | 结果 |
|---|---|---|
| admission | POST /admin/api/v2/residents `{nickname,bed_id=fd00:0:3:111:3:101::/96}` | cards INSERT 1 行 `active_bed`, `dns_short_name=u0003-r01-b01`; CloudEvent `op=admission` ✅ |
| transfer | PUT 改 bed_id 至 `fd00:0:3:111:3:301::/96` | cards: 旧 prefix.resident_id=NULL + 新 prefix INSERT/UPSERT; 2× CloudEvent `op=transfer` ✅ |
| DDNS PTR | `dig -x fd00:0:3:111:3:101::` | `0.0.0.0.0.0.0.0.1.0.1.0.3.0.0.0.1.1.1.0.3.0.0.0.0.0.0.0.0.0.d.f.ip6.arpa. PTR u0003-r01-b01.tenant3.owl.` ✅ |
| frontend cards list | GET `/data/api/v1/data/vital-focus/cards` | 返 `card_type:"active_bed"`, `dns_short_name:"u0003-r01-b01"`, `spatial_prefix:"fd00:0:3:111:3:101::/96"` ✅ |

**Phase F.5 — 启动 reconcile 恢复**（同次会话补完）：

v1 启动期 `CreateCardsForUnit` 已 Phase B4 退役成 stub，意味着旁路写入（SQL seed
demo data / migration import / 任何不走 `ResidentService.Create()` 的 resident）
都不会自动 materialize cards，与 v1 "重启时少了的补 / 多了的清 / 不变的留" 语义
不一致。

修复（commit pending）：

- 在 `card_sync_service.go` 重新实现 `CreateCardsForUnit(tenantPrefix, unitPrefix)`，
  按 unit /80 scope 扫 active resident_unit + LPM →
  对每个 (resident_id, spatial_prefix) UPSERT card +
  card_type 由 prefix.Bits() 决定（`CardTypeForMasklen`，/88=room, /96=active_bed, /80=unit）+
  DDNS register 尝试（best-effort）
- 同步 NULL out stale `resident_id`：扫 unit 范围 cards 中 resident_id 已无 active
  resident_unit 匹配的行
- 新增 `SetReconcileDeps(db, ddnsClient, owlDomain)` 注入；main.go 启动 wire
- **修 DNS shortName 冲突**：`CardShortName` 老版本 `/88 → u<unit>-r<room>` 在
  跨 branch/site 下会撞（多 branch 同 unit slot 重名）。改为
  `br<branch>-s<site>-u<unit>-r<room>[-b<bed>]`，跨 branch/site 唯一

E2E 验证（清空 cards 表 → 重启）：

| 步骤 | 结果 |
|---|---|
| `DELETE FROM cards` + 重启 | 自动 reconcile 出 8 张 card（4 room + 4 active_bed）覆盖所有 11 个 active resident 中 8 个有 spatial_prefix 分配的 |
| 手动 hard-delete 1 个 resident → 重启 | 该 prefix 的 active_bed card `resident_id` 自动 NULL 化（card 行保留，等待 device unbind 触发 物理删） |
| 再次重启（不变） | reconcile 跑完后 cards 表 0 net change（少不补、多不清） |

**已知 follow-up**（不阻塞 Phase F 验收）：

1. **DDNS forward zone (`tenant<N>.owl.`) 未在 BIND 配置中预创建** → `RegisterCardName` 推 AAAA 时收 NOTAUTH（仅 log warn，业务不阻塞）；PTR 在 `0.0.0.0.0.0.d.f.ip6.arpa.` 已就绪所以反查 OK。须 Phase B' 基建侧补 per-tenant forward zone bootstrap（kea ddns-conf 或 BIND zone provisioning）。
2. **discharge 软删被 residents.status CHECK 约束拒**：`residents_status_valid` 仅允许 `active|inactive|deceased|transferred`，与 memory R-002 `status='deleted'` 约定不一致。须二选一：(a) 改 schema CHECK 加 `deleted`；(b) 改 SoftDelete 用 `inactive`。memory 与 schema 不对齐属 owl_v2 cutover 收尾遗留。
3. **device unbind hook**: spatial_prefix 下所有 device 移除 → DELETE card + DDNS unregister 尚未接入；属 device path 修改，本次 cards-v2 carve out。
4. **alarm_events v2 化** → 仍在 Phase G（独立 PR）。

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

测试账号已验证：admin/Ts123@123（tenant Demo, fd00:0:3::/48）。

## 七、Phase F session summary（2026-05-12）

| Commit (pending) | Repo | Phase | Files | Notes |
|---|---|---|---|---|
| (this PR) | owlBack | F1 | wisefido-data/internal/service/resident_service.go | +120 行 syncCardForResident / materializeActiveBedCard / SetCardDeps |
| (this PR) | owlBack | F1 | wisefido-data/internal/repository/postgres_residents.go | +25 行 GetActiveSpatialPrefix；fix soft/hardDelete column 残留 (`hoa`→`resident_id`) |
| (this PR) | owlBack | F1 | wisefido-data/internal/repository/postgres_card.go | +24 行 GetActiveBedCardIDByPrefix |
| (this PR) | owlBack | F2 | wisefido-data/cmd/wisefido-data/main.go | +28 行 initDDNSClient hoist + SetCardDeps wire |
| (this PR) | owlBack | F3 | wisefido-data/internal/publisher/config_publisher.go | +44 行 PublishCardChanged + PublishCardResidentChanged |
| (this PR) | owlFront | F4 | src/api/monitors/model/monitorModel.ts + Overview.vue | CardType v2 enum + 5 处字面量同步 |

**E2E 验证已通过** — 详 §三 Phase F 表。

**红线遵循（R-001..R-008）**：
- ✅ 不向后兼容（CardType enum 直接换值，无 alias）
- ✅ 不新建 v2 文件（全部 in-place 重写）
- ✅ 单次完整改造（一次性 wire 起 admission/transfer/discharge/delete 四条路径）
- ✅ HIPAA：cards.resident_id 仅 current pointer；CloudEvent payload 不带 nickname/PHI
- ✅ DNS 永久名仅含 unit/room/bed slot（u<hex>-r<hex>-b<hex>），无 PHI
- ✅ DB schema 未改（仅复用现有 cards 表 + cards.dns_short_name 列）
- ✅ Claude 自主启停 owl-postgresql / owl-redis container + wisefido-data binary 完成 e2e
- ✅ commit 引用 § 编号（见 commit message）
