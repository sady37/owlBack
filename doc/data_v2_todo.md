# wisefido-data v2 ToDo — card / device 重写 backlog

起源：Mom 设备 unbind 失败（`devices_pkey` PK collision）→ 揭出 producer/consumer `has_bed` 反向 drift + card 寿命系在 device 上 + `device_ipv6` 当 PK 在 unbind 时被 `reset_device_prefix` 重写 → fixture MAC 冲突。决定 data 模块（特别 card / device 两大子模块）彻底重写。

**主线 = 必须按顺序做完的核心改造。**
**支线 = 顺手发现但独立可单做的修补；做支线时禁止改动主线条目，避免上次"从支线返回后主线全丢"的事故复发。**

---

## 主线 TODO

### ~~1. 重新规划 card 创建 / 维护规则~~（2026-05-22 Walk A 完成）

**Why**：
- 当前 [card_reconcile.go:115-141 buildExpected](owlBack/wisefido-data/internal/service/card_reconcile.go#L115-L141) 全 device-driven，仅 `monitoring_enabled=TRUE` 的 device 才反推出 anchor → **空 bed / 空 room / 空 unit 无卡**
- 当前 [card_reconcile.go:405-412](owlBack/wisefido-data/internal/service/card_reconcile.go#L405-L412) `has_bed` 走 structural `EXISTS(beds...)`；但 cardagg consumer（[device_meta.go:103](owlBack/wisefido-cardagg/internal/service/device_meta.go#L103) / [card_display_builder.go:21](owlBack/wisefido-cardagg/internal/consumer/card_display_builder.go#L21) / alarm_router）按 ActiveBed 语义读 → **producer/consumer 反向 drift**
- card 寿命系 device 上违反空间集合本质；[50_cards.sql](owlRD/dbv2/50_cards.sql) 已改文档规则但 Go 还没跟
- [`card_sync_service.upsertSpaceCard`](owlBack/wisefido-data/internal/service/card_sync_service.go#L226) 是 stub，真正写入在 [`card_reconcile.go`](owlBack/wisefido-data/internal/service/card_reconcile.go) —— 两套 writer 路径并存

**How to apply**：
- **lifecycle == space lifecycle**：bed/room/unit 存在即建对应 /96 / /88 / /80 卡，删 space 才删卡；与 device/resident 双解耦
- **has_bed 改 ActiveBed**：`EXISTS(devices d JOIN beds b ON d.device_ipv6 <<= b.bed_id WHERE b.bed_id <<= card_id)`；设备 bind/unbind/monitor flip 触发重算
- **buildExpected 改空间驱动**：每 bed → /96；split rule 决定 /88 / /80 聚合粒度（device 仅作 split 输入，不作 anchor gate）
- **bind/unbind 不动 card 行集合**：只更新 `has_bed` snapshot + `devices.card_id` FK

**实际落地（Walk A）**：
- ✅ buildExpected 加 `scanBedAnchors` 一步，bed 全扫 /96 强进 expected
- ✅ upsertCard.has_bed 改 ActiveBed SQL: `EXISTS(devices d JOIN beds b ON d.device_addr <<= b.bed_id WHERE b.bed_id <<= card_id AND d.monitoring_enabled=TRUE)`
- ✅ 11 个 v2 no-op stub 全删（card_sync_service 496→154 行 / card_create_service.go 整文件删）
- ✅ 9 个 caller 改直调 ReconcileCards（unit_service / device_service / main.go）
- ✅ 50_cards.sql lifecycle doc 精确化（Walk A 拍板 /96 == bed 寿命 / /88 /80 仍 split rule + device 触发）
- ✅ FE Overview.vue 加 `isCardMonitored` 过滤
- ✅ 实测 Card pfmb02 has_bed = TRUE（tenant fd00:0:3 卡 11→17）

详 commit `feat(data/cards): Walk A`。Walk B 不上（瘦 /88 /80 由 device 触发的简洁性 > Walk B 全空间 orphan 卡的彻底性）。

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

### S2. cardagg 过时变量名 / 注释

- [card_display_builder.go:21](owlBack/wisefido-cardagg/internal/consumer/card_display_builder.go#L21) `hasBedDevice = ... 物理上有 sleepad 床设备` —— 现在 `has_bed` 改 ActiveBed 后语义对了，注释保留 OK；变量名 `hasBedDevice` 也仍准确
- 其他文件 grep "active_bed" 残留同步改 "bed_card"（schema doc 已统一）

### S3. devices_pkey 测试 fixture MAC 重复

种子数据里 17 个设备末 32 bit 全 `::1`，#2 落地前任何 unbind 都 PK 冲突。**主线 #2 完成后此问题自然消失**（device_ipv6 不再重写）；在那之前要测 unbind 路径只能临时手动改 MAC。

### ~~S4. Overview.vue 加 monitor=off 卡片过滤~~（2026-05-22 完成）

落地：[Overview.vue](owlFront/src/views/monitoring/overview/Overview.vue) 加 `isCardMonitored` helper，`effectiveCardIds` 末端 filter。API 已带 `devices[].monitoring_enabled`（card_types.go:139）无需补。

### S5. qinglan `UpdateDeviceMonitoring` 并入 `devices.access`

按 [[qinglan_sleepace_access_gate_consolidation]]，3 字段（`allow_access` / `business_access` / `monitoring_enabled`）合并成单一 `devices.access` bool。主线 #2 是 device identity 重构，跟 access gate 字段合并是两件事；可在 #3 重写窗口顺手做也可单独 PR。

### S6. SSE fan-out 维护点 audit

`card_realtime_service` 的 `cardIndex` / `userIndex` 维护点 grep 一遍是否有 device 字段引用（按主线 #2 字段重命名要扫一遍）。

### S7. Pending Alarm Records — Critical auto_resolved 漏入 list（2026-05-22 用户报）

**Why**：FE alarmBell（alarm_level 统计 icon）显示有未处理 Critical alarm，但点开 Pending Alarm Records list 为空。设计契约：
- `popAlarm` (浮动横幅) **不含** `operation='auto_resolved'`
- `alarmBell` (统计 icon) + `Pending Records` (list) **应含** Critical (`alarm_level <= 3`) 的 `auto_resolved`（[[alarm_status_no_acked_auto_resolved]] A3 规则）

当前 Pending list query 过严，把 Critical auto_resolved 也过滤掉了。

**How to apply**：
- 检查 alarm_event_service.go 里 Pending list 的 WHERE 子句
- 改成：`alarm_status IN ('active','acked') OR (operation='auto_resolved' AND handler IS NULL AND alarm_level <= 3)`
- 注意 popAlarm endpoint 保持现有逻辑不动

来源：用户 2026-05-22 实测 / [[alarm_status_no_acked_auto_resolved]] / rule.md A3

---

### ~~S8. card_reconcile has_bed projection bug~~（2026-05-22 Walk A 修复）

Q1 实测 pfmb02 has_bed = TRUE ✓。详 Walk A 落地清单 §1。

---

### S9. `find_card_by_device_addr` 函数

[50_cards.sql:212-217](owlRD/dbv2/50_cards.sql#L212-L217) 当前实现 `SELECT card_id FROM devices WHERE device_ipv6 = addr`。主线 #2 device_ipv6 改稳定身份后，此函数语义和入参不变（仍按 device IPv6 查），但要确认 alarm 流里传的 device_addr 是新 IPv6 还是空间编码 IPv6。

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

**遗留 B1（待 FE 同步窗口）**：`enrichDeviceStatus.entry["device_id"]` map field key 实际存的是 device_addr (IPv6 文本)，命名误导。改成 `device_addr` 需 FE 同步。

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