# wisefido-data v2 ToDo — card / device 重写 backlog

起源：Mom 设备 unbind 失败（`devices_pkey` PK collision）→ 揭出 producer/consumer `has_bed` 反向 drift + card 寿命系在 device 上 + `device_ipv6` 当 PK 在 unbind 时被 `reset_device_prefix` 重写 → fixture MAC 冲突。决定 data 模块（特别 card / device 两大子模块）彻底重写。

**主线 = 必须按顺序做完的核心改造。**
**支线 = 顺手发现但独立可单做的修补；做支线时禁止改动主线条目，避免上次"从支线返回后主线全丢"的事故复发。**

---

## 主线 TODO

### 1. 重新规划 card 创建 / 维护规则

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

**Walk A vs Walk B 待定**：
- Walk A：每 bed 永远 /96，/88 / /80 维持 device 触发（推荐 / 保守）
- Walk B：每空间都建对应卡，display 层挑粒度（彻底但 orphan 多）

**前提依赖**：无（先行）

来源：本会话 §card_reconcile 诊断 / [50_cards.sql:10-11](owlRD/dbv2/50_cards.sql#L10-L11) 文档已改 / [[card_sync_service / has_bed drift]]

---

### ~~2. device_id 转 IPv6 化~~（2026-05-22 完成）

**完成方式**：Phase 2 一刀切 — `device_uid VARCHAR(50) logMAC` (identity) + `device_addr INET /128` (业务寻址) 两词收口。`device_id UUID` + `device_ipv6` 全栈退役。

详 [[phase2_device_uid_collapse_progress]] memory；rule.md §D1-D8；migration `owlRD/dbv2/migrations/2026-05-21_phase2_device_uid_collapse.sql`；8 hotfix commit (ffce64b / 05cdb00 / 2011a88 / 3e1e823 / f9e21c8 / f36832b / d6f25ea)。

---

### 3. 重写 wisefido-data card / device 模块

**Why**：
- #1 + #2 落地需要 data 内部 card/device 写路径成片改，逐函数 short-circuit 已被证明不可行（[[v2_cutover_lessons]]）
- 既有 stub 路径（card_sync_service）+ 事实路径（card_reconcile）并存 → 必须二选一统一
- repository / service / handler 三层都有遗留 v1/v2 双轨残骸

**How to apply**：
- 先 #1 + #2 dbv2 CREATE 改完用户审过（按 [[feedback_schema_review_via_dbv2]]）
- 整段重写不逐函数 short-circuit（按 [[v2_cutover_lessons]]）
- 单写主入口：`card_sync_service`（card 生命周期） / `device_service`（device 业务态）
- 边界：cards 表 100% data 写；devices 表外部写点（ipam / qinglan）同窗口跟字段，但不重写其代码结构
- 下游只读模块（cardagg / sensor / iot / sleepace / qinglan）跟着改字段名，不改结构

**重写涉及文件清单**：
- [owlBack/wisefido-data/internal/domain/](owlBack/wisefido-data/internal/domain/) ：`card.go` / `device.go` / `card_sync.go`
- [owlBack/wisefido-data/internal/service/](owlBack/wisefido-data/internal/service/) ：`card_sync_service.go` / `card_reconcile.go` / `device_service.go` / `device_store_service.go` 全部
- [owlBack/wisefido-data/internal/repository/](owlBack/wisefido-data/internal/repository/) ：`postgres_devices.go` / `postgres_cards.go` / `postgres_device_store.go` / `cards_repository.go`
- [owlBack/wisefido-data/internal/http/](owlBack/wisefido-data/internal/http/) ：`device_handler.go` / `unit_handler.go` / `vital_focus_handler.go` / `radar_handler.go`

**前提依赖**：#1 + #2

来源：本会话 §三层 stub/inverted 现象 / [[v2_cutover_lessons]] / [[feedback_producer_first]]

---

## 支线 TODO（不掺主线 — 单独可做 / 单独可取舍）

各项独立。做支线时**不要**修改、拆并、删除主线条目；返回主线就按原 #1 / #2 / #3 顺序接续。

### S1. `card_sync_service.upsertSpaceCard` stub 清理

[card_sync_service.go:226-309](owlBack/wisefido-data/internal/service/card_sync_service.go#L226-L309) 整段 `SELECT 1 -- v2.5 cards INSERT stubbed` no-op。重写时整段删，不留 deprecated wrapper（按 [[v2_cutover_lessons]] 整段重写规则）。

### S2. cardagg 过时变量名 / 注释

- [card_display_builder.go:21](owlBack/wisefido-cardagg/internal/consumer/card_display_builder.go#L21) `hasBedDevice = ... 物理上有 sleepad 床设备` —— 现在 `has_bed` 改 ActiveBed 后语义对了，注释保留 OK；变量名 `hasBedDevice` 也仍准确
- 其他文件 grep "active_bed" 残留同步改 "bed_card"（schema doc 已统一）

### S3. devices_pkey 测试 fixture MAC 重复

种子数据里 17 个设备末 32 bit 全 `::1`，#2 落地前任何 unbind 都 PK 冲突。**主线 #2 完成后此问题自然消失**（device_ipv6 不再重写）；在那之前要测 unbind 路径只能临时手动改 MAC。

### S4. Overview.vue 加 monitor=off 卡片过滤

按本会话讨论：SSE 全发，FE Overview 页 favor 视图过滤 `card.devices.some(d => d.monitoring_enabled)`。详 [Overview.vue:466 周边](owlFront/src/views/monitoring/overview/Overview.vue#L466)。

需要 `/vital-focus/cards` 接口响应里 card 必带 `devices[].monitoring_enabled` —— 先 verify 一下接口是否已带。

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

### S8. card_reconcile has_bed projection bug（2026-05-22 用户报）

**Why**：用户实测 Card `pfmb02` (Denver/A/101/GuestRoom/BedA) 有 bed + 已 bind device (`BM87224601903` 现 addr `fd00:0:3:111:3:201::1`)，但 has_bed 显示错。这是主线 #1 的具体案例 — Z-002 producer/consumer drift 仍有遗留。

**How to apply**：
- 在主线 #1 card_reconcile 重写时一并修：has_bed 走 ActiveBed (EXISTS device JOIN bed)
- 修完后跑 `card pfmb02` 验证 has_bed = TRUE

来源：用户 2026-05-22 实测 / [[z002_has_bed_producer_consumer_drift]] / 重复主线 #1 中已记的 has_bed drift（保留此项作 specific repro fixture）

---

### S9. `find_card_by_device_addr` 函数

[50_cards.sql:212-217](owlRD/dbv2/50_cards.sql#L212-L217) 当前实现 `SELECT card_id FROM devices WHERE device_ipv6 = addr`。主线 #2 device_ipv6 改稳定身份后，此函数语义和入参不变（仍按 device IPv6 查），但要确认 alarm 流里传的 device_addr 是新 IPv6 还是空间编码 IPv6。

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