# wisefido-data v2 ToDo — card / device 重写 backlog

起源：Mom 设备 unbind 失败（`devices_pkey` PK collision）→ 揭出 producer/consumer `has_bed` 反向 drift + card 寿命系在 device 上 + `device_ipv6` 当 PK 在 unbind 时被 `reset_device_prefix` 重写 → fixture MAC 冲突。决定 data 模块（特别 card / device 两大子模块）彻底重写。

**主线 = 必须按顺序做完的核心改造。**
**支线 = 顺手发现但独立可单做的修补；做支线时禁止改动主线条目，避免上次"从支线返回后主线全丢"的事故复发。**

---

## 主线 TODO

### ~~1. 重新规划 card 创建 / 维护规则~~（2026-05-22 Walk A → 2026-05-23 三轮迭代终版）

**最终规则（2026-05-23 拍板，权威 [rule.md C9](owlBack/doc/rule.md)）**：

```
N = bedCount_in_unit          (整 unit 床数 /96 候选)
M = noBedRoomCount_in_unit    (不含 bed 的 room 数 /88 候选)

Step 0: unit 无 room → 无卡
Step 1: Unit
  N ≤ 1 → /80 only (merge，吸所有)
  N > 1 → split：M > 0 时预创建 /80（装 noBed room 的 device）
Step 2: 每个含 bed 的 room
  Room.N ≤ 1 → /88 卡（吸 bed + room device）
  Room.N > 1 → N 张 /96 + room device → /80（M=0 时 lazy create /80）
```

三种实卡 **/80 + /88 + /96**。device/alarm 路由 = PG GiST `<<=` LPM 自动；alarm.producer 列保留发起设备身份。

**演进 3 版**：
- Walk A (2026-05-22) "bed 恒建 /96" — 在 merge 模式产生空 /96，违反原 spec
- 简化版 `37e413a` (2026-05-23 早) "N>1 split N /96 + /80 / N≤1 merge" — **砍 /88 太狠**，101 Bedroom radar E598A2ACD523 与 BedA device 分卡（落 /80），same-room 拓扑断裂
- 终版 (2026-05-23) — /88 回归，2-step + M + lazy /80 平衡 same-room 拓扑 vs 卡最少化

**关键修复**：
- [card_reconcile.go buildExpected](owlBack/wisefido-data/internal/service/card_reconcile.go) 重写 — 单 SQL 取每 unit 的 (room_count, bed_count, per-room beds[], per-room has-non-bed-device)，Go 端按 Step 0/1/2 决策含 lazy /80
- [card_static_service.go LATERAL JOIN](owlBack/wisefido-data/internal/service/card_static_service.go) 按 card_id mask 分级查 room/bed：/80 不查（room=NULL, bed=NULL），/88 取 exact room，/96 取 exact bed + parent room。修复 j0yvp3（/80）误显示"Denver/A/101/Bedroom/BedA"的 bug

**实测 tenant fd00:0:3 终态 12 unit**：MoM N=1 → 1 张 /80；101 N=2 M=1 → /80 + 2 /88（E598 与 BedA 同 /88 Bedroom ✓）；202 N=2 M=0 → 2 /96 无 /80；纯 bed unit → 多 /88 无 /80；空 unit → 无卡。所有 case 符合规则。

**has_bed 语义保留 ActiveBed**（C3，给 sensor 用），与卡结构 split 解耦 — upsertCard `EXISTS(devices JOIN beds <<= bed_id AND monitoring_enabled=TRUE)` 不变。

---

### ~~2. device_id 转 IPv6 化~~（2026-05-22 完成）

**完成方式**：Phase 2 一刀切 — `device_uid VARCHAR(50) logMAC` (identity) + `device_addr INET /128` (业务寻址) 两词收口。`device_id UUID` + `device_ipv6` 全栈退役。

详 [[phase2_device_uid_collapse_progress]] memory；rule.md §D1-D8；migration `owlRD/dbv2/migrations/2026-05-21_phase2_device_uid_collapse.sql`；8 hotfix commit (ffce64b / 05cdb00 / 2011a88 / 3e1e823 / f9e21c8 / f36832b / d6f25ea)。

---

### ~~3. 重写 wisefido-data card / device 模块~~（2026-05-22 重新定义 scope，写路径基本完成）

**Walk A 后的现实 scope**：

原 #3 写"重写 card / device 模块"指意是把"write path 整段重写、不逐函数 short-circuit"（按 [[v2_cutover_lessons]]）。Walk A 后 **写路径核心债已清完**：
- ✅ stub 全删（11 个 no-op）
- ✅ card_sync_service.go 瘦成 156 行（service 外壳 + globalCardSync hook）
- ✅ card_reconcile.go 算法层独立（buildExpected → diff → upsert → emit）
- ✅ has_bed producer/consumer 反向 drift 消除（核心 #1 案例）
- ✅ device_id UUID 收口（#2 完成）

**没动 / 真正剩的**（独立可单做，不再视作主线 #3 阻塞）：
- `device_service.go` 业务路径 v2 cutover 残留（独立 PR；现状已可工作）
- `postgres_devices.go` / `postgres_card.go` repository 层 audit（spatial_prefix→card_id 已扫，剩下做 schema diff 二次过滤）
- `device_store_service.go` / `cards_repository.go` 全删条件（看是否还有 caller）

详见下方 **§Card 模块 5 文件清单**。

---

## 支线 TODO（不掺主线 — 单独可做 / 单独可取舍）

各项独立。做支线时**不要**修改、拆并、删除主线条目；返回主线就按原 #1 / #2 / #3 顺序接续。

### ~~S1. `card_sync_service.upsertSpaceCard` stub 清理~~（2026-05-22 Walk A 完成）

整段删 + card_create_service.go 整文件删；详 Walk A 落地清单。

### ~~S2. cardagg 过时变量名 / 注释~~（2026-05-23 audit done）

audit 结论：
- `card_display_builder.go:21` 注释保留 OK（has_bed 改 ActiveBed 后语义对了）
- 其他 `active_bed` 字面量全部在 doc/comment 历史 reference，无运行时影响
- **唯一死代码** [cards_repository.go:136](owlBack/wisefido-data/internal/repository/cards_repository.go#L136) `case "active_bed"`：FE CardType 已改 `'bed'` 但仅本地过滤，**0 处发 `?card_type=` API 请求**，switch 永不命中 → **已清**

### ~~S3. devices_pkey 测试 fixture MAC 重复~~（2026-05-22 audit — Phase 2 没自动消失，code 修复）

**原 claim 错**："Phase 2 完成后此问题自然消失" — Phase 2 只重命名 device_ipv6→device_addr，没改 `reset_device_prefix` PG function（仍保 MAC32 + 清中间字节）。冲突根因还在。

**真根因 2 件**：
1. **`deriveMAC32Suffix` 偏离设计**：[doc/datagram_envelope.md §2.2](owlBack/doc/datagram_envelope.md) 规定 host 段 32 bit 派生自 device_uid 末 8 hex（剥非 hex + 不足左补 0）；code 实际是 MAC-first 偏离设计
2. **Seed fixture 完全无视派生**：[seed_demo_data.sql](owlBack/scripts/seed_demo_data.sql) 16 条设备硬编码 `::1`/`::2` 计数器，reset 后撞 PK

**修复落地（commit 待）**：
- `deriveMAC32Suffix` → `deriveUIDSuffix(deviceUID)` 单参数；按设计 strip 非 hex + 末 8 hex + 左补 0 + 小写归一
- 加单元测试 `postgres_device_store_test.go::TestDeriveUIDSuffix` 覆盖 doc 例子（E598A2ACD523 / BM87224700978 / 4D8710D5CABB / 不足左补 0 / 含分隔符）
- 现存 16 seed 数据保留（改动会波及 cards/alarms 历史指针）—— fixture exception，不影响真设备
- API 新建设备从此走正确 UID 派生

### ~~S4. Overview.vue 加 monitor=off 卡片过滤~~（2026-05-22 完成）

落地：[Overview.vue](owlFront/src/views/monitoring/overview/Overview.vue) 加 `isCardMonitored` helper，`effectiveCardIds` 末端 filter。API 已带 `devices[].monitoring_enabled`（card_types.go:139）无需补。

### ~~S5. qinglan `UpdateDeviceMonitoring` 并入 `devices.access`~~（2026-05-22 完成）

终态：**2 字段非 3 字段** — v1 `allow_access + business_access` 合并为 `devices.access` bool；`monitoring_enabled` 故意保留独立语义（access=平台审批 / monitoring_enabled=租户业务开关）。

落地清理（commit `86297f6`）：
- `domain.AllowAccessCache` → `AccessCache`（8 callsites + 1 var）
- 8 处 stale 注释字眼 "allow_access" / "business_access"
- `postgres_device.SearchDevices` 内 `business_access` criteria 死代码改 `access` bool
- owl-common `DeviceBaseline` 注释精简

详 [[qinglan_sleepace_access_gate_consolidation]]。

### ~~S6. SSE fan-out 维护点 audit~~（2026-05-23 audit done）

audit 结论：
- cardIndex/userIndex 按 cardID / userKey 索引，**本身无 device 字段引用**，post Phase 2 干净
- 唯一遗留 [card_realtime_service.go:1046](owlBack/wisefido-data/internal/service/card_realtime_service.go#L1046) 输出 `entry["device_id"] = d.deviceAddr` 是兼容历史命名；FE 全工程 grep 0 处读 `device_id` 字段（map key 已带 device_addr），注释陈旧 → **已清**

### S10. FE v1↔v2 字段别名 shim 拆除（C 类，跨视图主线）

owlFront 在 v2 cutover 期间留了 alias 回填层让未迁视图继续读 v1 字段：
- `api/units/unit.ts` `fillV1RoomAliases` / `fillV1UnitAliases`（v2 `prefix` → v1 `unit_id` / `room_id` / `bed_id`）
- `api/admin/branch/branch.ts` v1 `branch_id` UUID 别名 = v2 `prefix` /56
- `utils/http/axios/index.ts:277` X-Tenant-Id 兼容期注释

要拆需先把所有读 v1 别名的视图（UnitList / ResidentProfileContent / CareTeamList 等约 10 个）逐个迁到 v2 字段，再统一拆 shim。**主线 #3 v2 cutover 收尾环节**，跨 ~10 视图 + 大回归测试。

### S11. role 命名 v1/v2 双轨 + case 混乱（D 类，架构级）

**现状**：
- PG users 表 9 种 role 值：TitleCase v2 (`Caregiver/Nurse/Admin/Manager/Family`) + lowercase v1 (`platform_admin/tenant_admin/family/manager`) 混存
- backend 仍在用 v1 名：[auth_handler.go:443-444,544](owlBack/wisefido-data/internal/http/auth_handler.go) `u.role='platform_admin'`；[user_handler.go:542](owlBack/wisefido-data/internal/http/user_handler.go#L542) `tenant_admin`；[admin_tenants_handlers.go:65](owlBack/wisefido-data/internal/http/admin_tenants_handlers.go#L65) `role='tenant_admin'`
- FE [UserList.vue:1227-1229](owlFront/src/views/admin/users/UserList.vue#L1227-L1229) `.toLowerCase()` + 6 个名 (`systemadmin/platform_admin/admin/tenant_admin/manager/branchmanager`) 双套兜底 — **不能盲删**，盲删会让 platform_admin / tenant_admin / branchmanager 老账号丢管理权限

**需要**：
1. 拍板 role 标准化方案（统一 v2 TitleCase？保留 v1 兼容？）
2. PG 数据迁移（lowercase → TitleCase + v1 名 → v2 名）
3. backend SQL/代码统一
4. FE 清 v1 双名兜底

属于架构级决策 + 数据迁移，**不能在 v2 cutover 边缘子任务中顺手做**。

### ~~S7. Pending Alarm Records — Critical auto_resolved 漏入 list~~（2026-05-22 落地 + 修正）

**Why**：FE alarmBell 显示有未处理 Critical alarm，但点开 Pending list 为空。设计契约：
- `popAlarm` (浮动横幅) **不含** `operation='auto_resolved'`
- `Pending Records` (list) **应含** active + Critical(level<=2, EMERG/ALERT/CRITICAL) 已物理自动恢复但护理尚未 review 的 auto_resolved，"确保护理人员确认收到 alarm 消息"
- **acked 不入 Pending**（staff 已 ack=已确认收到，task done，进 Resolved tab）

**落地**：AlarmEventFilters 加 `Pending bool`；repo Pending=true 拼 `(alarm_status='active' OR (operation='auto_resolved' AND handler IS NULL AND alarm_level <= CRITICAL(2)))`；service `req.Status=="active"` 改设 `filters.Pending=true`。popAlarm 路径（GetRecentAlarmEvent）保留 `alarm_status='active'` 不动。

**初版误区**（2026-05-22 第一次落地用 acked + level<=3，被用户纠正）：把 acked 纳入 Pending 导致已 ack 的 alarm 反复回到 Pending；FE Handle modal 硬编码提交 `alarm_status='acked'`，对已 acked 是 no-op → 用户无法从 FE 移出 → 表现"卡住"。修正方案改回 active + Critical(level≤2)。

**Critical auto_resolved Handle 路径**（同次落地）：
- service.HandleAlarmEvent guard 加 `event.AlarmStatus == "auto_resolved"` 分支：强制 target=`resolved`（跳 acked 按 A3），operation 保留 `auto_resolved`（KPI `operation='auto_resolved' AND handler IS NOT NULL` 仍能识别"物理自动恢复后人工 review"）
- FE AlarmRecordList.vue handleConfirmHandle：源是 auto_resolved 提交 `alarm_status='resolved'`，否则提交 `'acked'`
- 闭环：Critical auto_resolved 留 Pending → staff 点 Handle → backend coerce → 进 Resolved tab，operation='auto_resolved' + handler 双标记

**alarmBell counter ↔ Pending list 对齐**（同次落地）：QueryCardAlarmState SQL 改 `Critical(0/1/2): active OR (auto_resolved AND handler IS NULL); Error/Warning(3/4): active`，与 Pending list 谓词完全一致；acked 不计入 alarmBell（与 Pending 一致 — acked=task done）。owl-common 改完 cardagg/data/qinglan/sleepace/sensor 5 个服务全 restart。

**alarm 锚定卡规则**（同次落地，事件级语义）：
- 不变：alarm.card_id 是 trigger 瞬间 GiST LPM snapshot，**device 迁走 alarm 不动**（留作历史，新 card 看不到、原 card 仍能查 Resolved tab）
- 新增：cards row 真正删除 → 同 card_id 上所有非终态 alarm（active/acked/auto_resolved）自动 UPDATE 'expired'。"空间集合改变 → card_id 失指 → expired"
- owl-common 加 `ExpireAlarmsByCardID` (单卡精确) + `ExpireAlarmsByCardPrefix` (`<<=` 批量)；移除旧 `ExpireAlarmsByDeviceAddrs`（按 device_addr 过严，迁走 device 漏过）
- 三处 card 删除路径全接：`postgres_card.DeleteCard` / `DeleteCardsByUnit` / `card_reconcile.applyDiffs`（事务内调 tx）

来源：用户 2026-05-22 实测 + 纠正 / [[alarm_status_no_acked_auto_resolved]] / rule.md A3

---

### ~~S8. card_reconcile has_bed projection bug~~（2026-05-22 Walk A 修复）

Q1 实测 pfmb02 has_bed = TRUE ✓。详 Walk A 落地清单 §1。

---

### ~~S9. `find_card_by_device_addr` 函数~~（2026-05-23 audit done）

audit 结论：
- 函数 [50_cards.sql:227-230](owlRD/dbv2/50_cards.sql#L227-L230) 已改为 `SELECT card_id FROM devices WHERE device_addr = addr LIMIT 1`，依赖 devices.card_id（denormalized）
- devices.card_id 由 [card_reconcile.applyDiffs:355-360](owlBack/wisefido-data/internal/service/card_reconcile.go#L355-L360) 按 mask DESC sort 后逐 anchor 更新，最具体优先
- 实测 101 五设备 LPM 完美：room radar→/88 Bedroom / bed device→/88 Bedroom / Bathroom radar→/80
- 无需 fix

---

## Card 模块 5 文件清单（2026-05-22 Walk A 后盘点）

> 主线 #3 的 "重写 card / device 模块" 在 Walk A 落地后 scope 已收紧为**只动写路径** —— 4 个 service + 1 个算法文件并非全部都需要重写。下面是 5 个文件各自的职责定位、IPv6 touch points、去留判定，作为今后调整的参照表。

```
              ┌──────────────────────────────┐
              │  card_allowed_provider.go    │  权限：WHO 能看 WHICH 卡
              │  (378 lines)                  │
              └──────────────┬────────────────┘
                             ↓ CardList
              ┌──────────────────────────────┐
              │  card_static_service.go      │  读：HTTP CardStatic 装配
              │  (723 lines)                  │
              └──────────────────────────────┘
                             ↑ 读
                             │
              ┌──────────────────────────────┐    ┌──────────────────────────┐
              │  card_sync_service.go        │ →  │  card_reconcile.go       │
              │  (154 lines, Walk A 后)       │    │  (760 lines)              │
              │  service 外壳 / globalCardSync│    │  ReconcileCards 算法层    │
              └──────────────┬────────────────┘    └──────────────────────────┘
                             ↓ emit CardChange
              ┌──────────────────────────────┐
              │  card_realtime_service.go    │  推送：SSE 长连接 + CardList diff
              │  (~1270 lines)                │
              └──────────────────────────────┘
```

### 1. `card_allowed_provider.go` — 留

**职责**：根据 user role 算"可见卡 ID 列表"。
- Family role → 只看自家 resident 所在 unit 的卡（`filterCardsForResident` LPM 反查 cards.resident_id）
- Staff role (Admin / Manager / Nurse / Caregiver) → 按 branch 分组（`cardIDsByUnit` 走 cards.unit_id /80）
- Share unit 特殊处理（`ActiveBedcardIDsByUnitShared`）
- 返回 `CardList` (按 /56 branch 分组)

**IPv6 touch**：tenantID /48 INET / unitID /80 LPM cards.card_id <<= unit_id / branchID /56 分组 / resident_id /128 反查

**判定**：单一职责清晰、跟其他服务零代码重复、与 IPv6 prefix 模型天然契合。**no rewrite needed**。

### 2. `card_static_service.go` — 留

**职责**：HTTP 读路径，把 cards 行装配成 `CardStatic`（含 unit/room/bed/devices/residents/caregivers）。
- `GetCardList` 分页 + 权限过滤
- `GetCardInfo` / `GetCardsByCardIDs` 单卡 / 批量
- `enrichResidentsAndDevices` + `fillDevicesV3`（350 行 / 总 723 行）— device LPM + resident + device sibling 决断
- `GetCardCaregivers` 聚合 resident_caregivers
- `resolveCardID` 接受 `card_dns` 短码 / IPv6 CIDR 双形态

**IPv6 touch**：card_id INET / device_addr INET / bed_id INET / room_id INET / unit_id INET；LPM `d.device_addr <<= bed_id <<= room_id <<= unit_id`

**判定**：装配复杂度是 FE 业务必须，不是债。**留**。fillDevicesV3 偏大但是真复杂度。

### 3. `card_sync_service.go` — 留（Walk A 后已 154 行 service 外壳）

**职责**：写路径瘦壳 + 全局 hook。Walk A 后只剩：
- `PublishConfigCardReset` / `SetReconcileDeps` — wiring
- `CleanupOrphanCards` — 启动期清孤儿卡
- `ClearAllCards` — 灾难恢复（dormant，无 caller）
- `emitCardChange` — CloudEvent emit helper
- `globalCardSync` + `InitGlobalCardSync` — 包级单例 → unit_service / device_service hook 用

**IPv6 touch**：card_id INET（间接）—— 实际 SQL 都在 card_reconcile.go

**判定**：跟 card_reconcile.go 一起构成写路径，自身已 154 行无 stub。**留**。

### 4. `card_reconcile.go` — 留（Walk A 后已稳定）

**职责**：cards 表唯一权威写入入口（算法层）。
- `ReconcileCards` 入口（scope=INET CIDR）
- `buildExpected` → `scanBedAnchors`（Walk A 新增）+ `queryUnitAnchors` + `applyUnitSplitRule`
- `loadCurrent` + `applyDiffs` + `upsertCard`（含 has_bed ActiveBed 计算）
- `resolveResidentForCard` / `isPublicUnit` / `lookupResidentNick`
- `emitDiffs` + `syncDDNSForDiff` + `publishDiff` + `ensureDDNSForExpected`

**IPv6 touch**：贯穿全文件 —— scope /48 / unit /80 / room /88 / bed /96 anchor + device LPM

**判定**：Walk A 已落地，算法层独立性好。**留**。760 行可考虑拆 emit/DDNS 到独立 file（nice-to-have 不是必须）。

### 5. `card_realtime_service.go` — 留（2026-05-22 已 A1+A2+A3 清理）

**职责**：SSE 长连接 + 每 user CardList diff 推送。
- SSE 连接生命周期（registerSSE / activateSSE / unregisterSSE / supersedePrevious）
- `StartStatusFanout` — `StatusEvent` 推送到各 SSE
- `UpdateCardList` / `UpdateByBranch` / `diffCardList` — CardList 变化 diff 然后推 CardChange
- `getCardList` / `cardIndex` per-user cache
- `pull-realtime` endpoint
- `SubscribeRealtimeStream` (单卡 SSE) / `SubscribeCardsStream` (多卡 SSE) / `InitSSE` / `UpdateSSEView` / `UpdateSSEWatch`
- `GetCardStatus` (Detail 页 30s polling) + `enrichDeviceStatus`

**IPv6 touch**：cardID INET 字符串（cache key / index key），不直接做 LPM；enrichDeviceStatus SQL 走 /80 unit pool

**判定**：~1270 行偏大但是 4 个并列子模块（SSE/connection / Realtime pull / Status fanout / Card list diff）的合理体量。**留**。瘦身（拆 sse_connection.go / cardlist_diff.go / fanout.go）是 nice-to-have。

~~**遗留 B1**：`enrichDeviceStatus.entry["device_id"]` map field~~（2026-05-22 commit `ca52e4f` S6 已删 — FE 全工程不读该字段，map outer key 已是 device_addr 文本）。

---

## 工作流约定

- 每次会话进入前先看 "主线 TODO" 状态；主线只能按 #1 → #2 → #3 顺序推进
- 走支线时**新建一行**记录在 §支线 末尾，不挪主线条目位置
- 完成项打 `~~删除线~~` 不直接删，保留审计回溯
- schema 改动一律先改 [owlRD/dbv2/](owlRD/dbv2/) CREATE 给用户审，再 ALTER + Go（按 [[feedback_schema_review_via_dbv2]]）

测试账号：
plamform-admin:  sysadmin/ChangeMe@123
tenant_admin:      admin/Ts123@123
branch_manager: demo/Demo@2026
nurse: dvn1:Ts123@123
caregeiver: dvc1:Ts123@123  
family: f01:Ts123@123