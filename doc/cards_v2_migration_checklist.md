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

### Phase A — Schema 收尾
- [x] cards 表 v2 schema 已 deploy（`/owlRD/dbv2/50_cards.sql`）
- [x] resident_unit 表已支持 spatial_prefix INET + valid_from/valid_to
- [x] view `card_hr_source` 已部署
- [ ] **TODO**: 加 view `card_current_resident`
- [ ] **TODO**: 加 PG function `find_card_by_device_addr(INET)` LPM 反查（封装）

### Phase B — owl-common + wisefido-data 重写
- [ ] B1：`owl-common/card/*` 全部重写
  - 删除 `card_db.go` / `alarm_db.go` 引用的 v1 字段
  - 删除 `devices_jsonb.go` / `utils.go` (ConvertDevices/ResidentsToJSON)
  - 新 `card_types.go` 枚举：`tenant/branch/site/unit/public/room/active_bed/device`
- [ ] B2：`owl-common/ddns/client.go` 加 `RegisterCardName/UnregisterCardName`
- [ ] B3：`wisefido-data/internal/domain/card.go` 重写为 v2 model
- [ ] B3：`wisefido-data/internal/repository/postgres_card.go` SQL 全部重写
- [ ] B3：`wisefido-data/internal/repository/cards_repository.go` 权限查询用新 card_type 值
- [ ] B4：`wisefido-data/internal/service/card_create_service.go` 改为事件驱动
- [ ] B4：删除 `staff_pop_pending_counter.go` 中 unhandled_alarm_* 引用 → 走 alarm_events 实时聚合
- [ ] B4：`card_static_service.go` / `card_sync_service.go` / `card_realtime_service.go` / `card_allowed_provider.go` / `alarm_event_service.go` 适配新枚举值
- [ ] B5：`config_publisher.go` 拆成 `PublishCardChanged` + `PublishCardResidentChanged`

### Phase C — wisefido-cardagg
- [ ] `card_change_handler.go` 处理两种新 CloudEvents type
- [ ] `alarm_service.go` 用新 card_type 值
- [ ] 内存反查 map 改为 `device_addr INET → card_id`（已是这样的不动）

### Phase D — Build green
- [ ] `wisefido-data` go build
- [ ] `wisefido-cardagg` go build
- [ ] `wisefido-iot` go build（如有引用）
- [ ] `wisefido-sensor` go build
- [ ] `wisefido-sleepace` go build
- [ ] `wisefido-qinglan` go build
- [ ] `wisefido-ai-health` go build

### Phase E — 自动 e2e 测试
- [ ] E1：重启服务（owlBack stack）+ tail 启动日志 60s 无 panic
- [ ] E2：login sysadmin → admin → 创建 resident 关联 bed → 验证 cards 行自动创建
- [ ] E3：验证 DDNS zone 文件含新永久名
- [ ] E4：验证 redis stream `config:card:stream` 收到 CloudEvents
- [ ] E5：login family/nurse 视角查看 card list（覆盖权限）

### Phase F — 前端（按需）
- [ ] `src/api/monitors/model/monitorModel.ts` CardType enum
- [ ] `Overview.vue` card_type 判断改新值
- [ ] Pinia store / TS interface 同步

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

最后 push 前在本文件 § 三勾选并写 Phase G summary。
