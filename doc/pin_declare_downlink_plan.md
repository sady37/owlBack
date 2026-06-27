# 区域编排 / declare 下发 / 失败告警 — 后端化重构方案

> 6-25 拍定方向：**DB = 唯一真值（SoT），device = 投影**。`wisefido-data.SaveRoomLayout(canvas)` 是 layout 的**单一写入口**。
> area-id 编排 + canvas→device-model 转换 + declare 全量重构 + 下发 + 失败告警 **全部下沉后端**；**FE 只画 canvas**（只 PUT 几何，不算 area_id / area_type / declare，不发 MQTT）。
> feedback 改走 data 的 `SaveRoomLayout`（顺手修掉 sensor 直连库违规 + 消除生效盲区）。

---

## 0. 一句话大白话

现在「房间里哪块是床、哪块是门、哪块要当噪声屏蔽」这套活，是**前端浏览器**算完、再由前端**直接命令雷达**的。问题：真值散在前端会话里、雷达被前端绕过后端直接捅、feedback 又偷偷绕过写入口改库。本方案把这套活**整体搬回后端**：前端只负责"画图"，画完存库；后端读库→编排→翻译成雷达听得懂的指令→下发→下发失败/超 16 块就报"设备故障"。

---

## 1. 现状（as-is，带 file:line）

### 1.1 FE 当前承担了全部编排（要搬走）
- `owlFront/src/components/Radar/Toolbar.vue:2950 unifiedSave()` — 保存编排：先存 layout，再算 radar diff 直接发 MQTT。
- `owlFront/src/utils/radar/radarUtils.ts:709 updateRadarAreas()` — **FE 分配 area_id**（`assignAreaIds` 复用旧 id，0–15 粘性）。
- `radarUtils.ts:727-730` / `Toolbar.vue:3215-3241` — **FE 映射 area_type**（typeName→0-9，再 `firmwareDeclareAreaType()` 塌到 0-5；Bed 含雷达→5 否则 2）。
- `radarUtils.ts:733` — **FE 做几何换算**：canvas 坐标 → 雷达本地系 (h,v,z)，含 angle 旋转。
- `radarUtils.ts:715` — **只 `source='Human'` 进 declare**；Feedback/Learned 存 canvas 但**不下发固件**。
- `Toolbar.vue:3220-3241` — FE 拼 `declare_area` 字符串并经 `executeRadarCommands()` **直接发 MQTT** 给设备。

### 1.2 data 的 SaveRoomLayout 现在只是"存画布"
- `owlBack/wisefido-data/internal/service/radar_install_service.go:371-432 SaveRoomLayout(ctx, tenant, roomID, configData)`
  - 规范化 prefix(/80,/88,/128) → SHA256 去重 → UPSERT `room_visual_layout(canvas, canvas_hash, version++)`。
  - `detectAndNotifyFeedbackVetoes()`（/128 上 Feedback 删除→POST sensor `/roomengine/cell/veto`）。
  - `PublishConfigChanged("update", [prefix])` → 发 `config:card` → sensor `ReloadRooms()`。
  - **不生成 declare_area、不下发设备。**
- HTTP 入口 `radar_handler.go:391-429 PutRoomLayout`（PUT `/.../room/:roomId/layout`）。

### 1.3 qinglan 才是真正的下发腿（cap 在这里，但静默）
- `owlBack/wisefido-qinglan/internal/decode/radar_encode.go:359 const maxDeclareAreas = 16`。
- `sortCapDeclareAreaMaps()/sortCapDeclareAreaString()`(356-418) — 按 `declareAreaKeepPriority` 排序后**静默截断**到 16，仅 `log.Printf` 警告，**不报警**。
- `internal/service/radar_service.go:260-294 SetDeviceProperties` — 多区时逐块 300ms 下发 MQTT `cmd=update,data.declare_area`；**单块失败无处理**。
- 设备线协议 area_type 是**数字 0-5**。

### 1.4 sensor 直连库违规（要随手修）
- `internal/roomengine/feedback.go:206` — feedback ingest **直接读 `alarm_events`**。
- `feedback.go:405-414 appendFeedbackObject()` — **直接 `UPDATE room_visual_layout`**（绕过 SaveRoomLayout）。
- `internal/roomengine/layout_load.go:29 / 133` — 直接读 `room_visual_layout` / `beds`。
- `internal/service/alarm_enablement.go:~150` — 直接读 `device_config`。
- HTTP 入口 `cmd/wisefido-sensor/veto_http.go:73-93` `/roomengine/feedback/ingest`。

### 1.5 DeviceFailure 已注册但从不触发
- `owl-common/alarm/alarm.go:59,175 AlarmTypeDeviceFailure`（Level=Err、`DedupWhileActive`、HIPAA 强审计、`alarmGateSkip` 绕过 enablement）。
- `wisefido-qinglan/internal/consumer/stream_publisher.go:67-104 PublishAlarm` 会发，但**无人在 declare 失败/超 16 时调它**。

---

## 2. 权威路由矩阵（6-25 拍定）

权威枚举：`owl-common/observation/fields.go:281-310`（AreaType 0-9 + `Name()`）。固件线只认 **0-5**。

**核心原则**：
1. **不 pin = 啥都不动**（不碰 layout / cell / firmware）。主动权在护工/士，没 pin 即不存在。
2. **三层各自独立**，按 type 分流：layout 永远写（SoT）；cell 看引擎是否需要；firmware 只下发固件认的码（0-5）。
3. **Enter / Bed / MonitorBed 恒下发**（结构性，人工画不经 feedback）；**Chair / AreaBlind 永不下发固件**。

| 对象 | pin 行为 | layout | cell(引擎grid) | firmware 线码 | 大白话 |
|---|---|---|---|---|---|
| **Enter** | **恒 pin** | ✓ | ✓ | **4「门」** | 门结构性，永远声明 |
| **Interfer** | pin→全 / 不pin→啥都不动 | ✓ | ✓ | **3「噪」** | 帘/扇/植物抖动，固件屏蔽（本次新增下发）|
| **Reflector** | pin→全 / 不pin→啥都不动 | ✓ | ✓ | **3「噪」** | 镜/金属反射，固件屏蔽 |
| **Chair** | pin→layout+cell / 不pin→不动 | ✓ | ✓ | **✗ 不下发** | 引擎当坐区(Sit)防误报久坐摔，固件不认 |
| **Recliner** | pin→全 / 不pin→不动 | ✓ | ✓ | **1 Deny** | 较固定家具，pin 时下发屏蔽 |
| **AreaBlind** | pin→仅 layout / 不pin→不动 | ✓ | — | **✗ 不下发** | 盲区雷达看不见，下发无意义 |
| **Bed**（不含雷达） | **恒下发**（人工画，不经 feedback）| ✓ | ✓ | **2** | 接触床，结构性 |
| **Bed/MonitorBed**（含雷达）| **恒下发**（人工画，不经 feedback）| ✓ | ✓ | **5** | 监护床（生命体征+床事件）|

> **关键行为变更**：①今天 FE 把 Interfer(6) 当"其它"丢弃(`-1` 不下发)，本方案 **Interfer→噪3 下发固件**（消盲区/降假阳）。②今天 FE 只 `source='Human'` 下发，本方案 **pin 即权威下发**（不 pin 啥都不动）。
>
> **三层名词**：`layout`=`room_visual_layout` canvas（SoT）；`cell`=引擎 grid AreaType（经 config:card→ReloadRooms→StampPrior 染色）；`firmware`=declare_area 下发设备。同一 pin 动作由 `SaveRoomLayout` 单入口自动分流三层。

cap 优先级（沿用 `declareAreaKeepPriority`）：删除 > 床(2/5) > 门(4) > 噪(3) > 自定义(1) > 其它。

### 2.1 typeName 级下发策略表（实现权威 / 对齐真实 FURNITURE_CONFIGS）

> 编排器必须按 **typeName**（非 typeValue）决定下发——因为 Chair(Sit7) 与 Recliner(Lying8)/Furniture(Deny1) 引擎语义都折叠到固件 1，固件无法区分，但 Chair 要停发、后两者要发。

| typeName | typeValue | layout | cell | **firmware 码** | pin 语义 |
|---|---|---|---|---|---|
| Bed | 2 | ✓ | ✓ | **2** | 恒下发（人工画不经 feedback）|
| MonitorBed（Bed∩Radar 派生）| 5 | ✓ | ✓ | **5** | 恒下发 |
| Enter | 4 | ✓ | ✓ | **4** | 恒下发 |
| MetalCan / Mirror | 3 Reflector | ✓ | ✓ | **3** | pin 才发 |
| Curtain / WheelChair | 6 Interfer | ✓ | ✓ | **3** | pin 才发 |
| Recliner | 8 Lying | ✓ | ✓ | **1** | pin 才发（较固定）|
| Furniture（通用）| 1 Deny | ✓ | ✓ | **1** | pin 才发 |
| **Chair** | 7 Sit | ✓ | ✓ | **✗ 不下发** | 🔴 唯一行为变更（今天发1）；**可移动**——固定声明会框住过时位置造真盲；停发后靠引擎 Sit 区 90min floor |
| BlindArea | 9 Active | ✓ | — | **✗ 不下发** | 雷达不可见 |
| Wall | −1 | ✓ | — | **✗ 不下发** | 仅画布 |

**唯一固件行为变更 = Chair 停发**；其余 typeName 与今天 FE 一致。结构变更 = ①编排搬后端 ②feedback-pin 也下发(D1) ③Chair 停发。

> 实现：编排器替换 FE 的 `firmwareDeclareAreaType(typeValue)` 折叠为 **per-typeName 策略查表**（上表）。`pin 语义`：Bed/MonitorBed/Enter 恒发；其余「pin 才发」= source∈{Human, Feedback-pinned}；不 pin 啥都不动。

---

## 3. 目标架构（target）

```
FE (Toolbar.vue)                wisefido-data                         wisefido-qinglan         device
  画 canvas 几何 ───PUT layout──▶ SaveRoomLayout(canvas)  [单写入口]
  (无 area_id/type/declare,         │
   无 MQTT)                          ├─ UPSERT room_visual_layout (SoT)
                                     ├─ 【新】Orchestrator:
                                     │    canvas→dm: 顶点 angle 旋转 + 本地系(h,v,z)
                                     │    分配/复用 area_id (读旧 layout 粘性 0-15)
                                     │    映射 area_type→线码(0-5, Interfer→3)
                                     │    全量重构 declare set
                                     ├─ PublishConfigChanged(config:card)→ sensor ReloadRooms
                                     └─ 【新】下发委托 ──HTTP──▶ SetDeviceProperties
                                                                    final cap=16 + 逐块下发 ──▶ declare_area
                                                                    超16 / 下发失败(NAK/超时)
                                                                       └─▶ PublishAlarm(DeviceFailure)

sensor feedback ──HTTP──▶ data /internal/layout/feedback ─▶ SaveRoomLayout(stamp 反馈对象) ─▶ 同上全链
  (不再直连库)                (data 读 alarm_events 回传 payload)
```

核心三句：
1. **canvas 是 SoT 的输入，`room_visual_layout` 是 SoT，device 上的 declare_area 是投影**——任何时候以库重算可重建设备状态。
2. **编排只在 data 一处**：area_id 粘性、area_type 线码映射、几何旋转换算、declare 全量重构。
3. **下发 + 失败判定只在 qinglan 一处**：final 16-cap、逐块下发、ACK/超时监控、超限或失败 → DeviceFailure。

---

## 4. 关键决策点（待你拍板）

### D1. ✅ 已拍定（6-25）：pin = 权威，按 type 自动分流
- **不是"自动学 vs 手动"，而是 pin vs 不 pin**。护工/士 **pin 这个动作本身即权威决定**（等同 SourceHuman），**不 pin = 啥都不动**。
- pin 后按 §2 矩阵分三层：固件认的码(0-5)下发固件，引擎类型(Chair/Sit 等)只刷 cell，AreaBlind 只写 layout。
- 我之前"自动学错→真盲区(R2)"的担忧**不适用 pin**——人手动 pin 不存在自动误屏蔽。R2 仅约束 cell 自动学(Shadow)，那条腿**永不下发固件**（记忆 `cell_area_learning_scoring_design`）。
- **关闭项**：原"feedback 自动下发"的疑虑作废，改为本矩阵。

### D2. 16-cap 的"超限"在哪判、谁报 DeviceFailure？
- 选项 A：data 编排时已知逻辑区总数，超 16 立即 `DeviceFailure` + 回 FE 提示（用户当场知道画多了）。
- 选项 B：data 全量下发，qinglan final cap 截断后若发生截断 → `DeviceFailure`（你的原话"超16…→qinglan 发"）。
- **推荐 A+B 都做**：data 预检超 16 → 同步回 FE 让用户删区（最佳体验）；qinglan 兜底 final cap + 一旦真截断/下发失败 → `DeviceFailure`（运维可见，HIPAA 审计）。**FE 不再静默丢区**。

### D3. canvas→dm 的几何换算放哪？
- 必含 **angle 旋转**（记忆 `radar_area_vertices_must_apply_angle`：顶点 PRE-rotation 需绕中心转）+ canvas→(h,v,z)。
- **推荐**：抽到 `owl-common`（如 `owl-common/spatial/declare`），data 编排与 qinglan 解析共用一份，杜绝 FE/后端两套漂移。

### D4. area_id 粘性来源？
- 固件 slot 0-15 一旦变更即重 declare（抖动）。今天 FE 复用旧 id 保粘性。
- **推荐**：data 编排时**读现有 `room_visual_layout` 的 areaId**，diff：旧区保号、新区取空闲号、删区释放号。粘性真值落库（本就在 canvas.objects[].areaId）。

---

## 5. 分阶段施工

> 红线（记忆 `validate_real_case_no_unit_tests`）：**禁 unit test / 替身 harness，只跑真 case 验证**。每阶段 `build + vet` 绿、真设备/replay 验。

### T1 — 后端编排器骨架（data，shadow 不下发）✅✅ 已验证通过（6-25）
> 已落 `wisefido-data/internal/service/declare_orchestrator.go`（FE 几何全量移植 + §2.1 per-typeName 策略），`SaveRoomLayout` 内 `shadowLogDeclare` 接通；`go build`+`go vet` 绿。
> **运行期实测对账**（b17f Kitchen，真 FE save，log `/home/wisefido/owl/log/wisefido-data.log`）：FE 下发 6 区，后端逐区比对——
> - 几何(R1)：6 对象顶点**逐字节对齐**（含被 skip 的 Chair 算出的几何）→ 移植零漂移。
> - area_id(R3)：`{1,2,4,6,9,10}` 吻合 FE（避开旧占用 `{0,3,5,7,8,11}`）→ `assignAreaIds` 正确。
> - 策略：3 Enter(4)+2 Furniture(1) 照发；唯一分歧 = 可移动 Chair（FE 发 type1 / 后端 `skip policy:not-sent`），后端 count=5 比 FE 少 1。
> 结论：编排器与现网 FE 等价（仅 Chair 按设计停发）→ **R1/R3 清，可进 T2 cutover**。
- 新 `wisefido-data/internal/service/declare_orchestrator.go`：输入 canvas objects → 输出逻辑 declare set（area_id 粘性 + area_type 0-5 线码 + 旋转后顶点 cm）。
- **port 自 FE**（必须逐字对齐，R1）：
  - `toRadarCoordinate`(radarUtils.ts:94)：平移到雷达中心 → 逆旋转 `angle+corn(-45°)` → `h=-x, v=y`。
  - `getObjectVerticesInRadar`(591)：canvas 顶点(含对象自转)→雷达系→`round/10*10`→按 v 升序、同 v 按 h 升序排。
  - `assignAreaIds`(690)：复用旧 areas 的 id，新区取 0-15 首个空位（粘性，R3/R4）。
  - 类型：`AREA_TYPE_MAP[typeName]`→`firmwareDeclareAreaType()`折叠 0-5；Bed 含雷达→5 否则 2；按 §2 矩阵 Chair/AreaBlind→不下发。
- 输出对接 qinglan 既有腿：产出 `area_{i}_id/type/x1..y4`(cm) 交 `buildDeclareAreaFromConfigCm`（已存在，/10→dm + cap + 下发）。**data 不重做 cap/下发**。
- `SaveRoomLayout` 内调用，仅 `log` 输出 declare set 与现网 FE 产物对比；**不下发**。
- 验：同一 canvas，后端 declare set == 现网 FE `declare_area`（逐区 id/type/顶点对齐）。
- **几何换算暂放 data 内**（不先抽 owl-common，避免过早抽象）；T5 若 qinglan 也需再抽 D3。

### T2a（已落 6-25）— FE 先停发 Chair
> `firmwareDeclareAreaType` 把 Sit(7,Chair 可移动)→`AREA_TYPE_NOT_SENT`（原 →1）；Lying(8,Recliner)/Deny(1,Furniture) 仍 →1。复用既有 `isDeviceAreaType`(0-5) 过滤，Chair 不进 declare。根因=type1 雷达不处理该区，固定声明框住已挪走的椅子位置→漏检。FE(Vite HMR)与后端 shadow 对齐：两边都不下发 Chair。

### T2 — 下发腿接通（data→qinglan），关掉 FE 下发 ✅ 已落+部署（6-25）
> **决策**：save 即下发（projection 随 SoT）；`SaveRoomLayout` 返回结构体带 warnings/下发状态（3 端统一显示）。
> **data**（`declare_orchestrator.go` + `radar_install_service.go`）：`SaveRoomLayout` 持久化后 → `applyDeclare`：编排 new/old set（顺序 id 0..N-1）→ 16-cap 按优先级 + drops 转 warning → `downlinkDeclares` 每雷达 `device_addr`→`GetDeviceUID`→拼 `declare_area`（cm 串，删旧尾部 `[N..M-1]` type0 + 加新 0..N-1）→ `qinglanClient.SetDeviceProperties`（失败接 T3a DeviceFailure）。返回 `SaveRoomLayoutResult{Warnings,Downlink[]}`，HTTP `Ok(result)`。
> **FE**（`Toolbar.vue`）：两处 builder 的 `declare_area` 命令块**整段删除**（含 `>16` 警告、deltaOps、`buildDeclareAreaStringFromAreas`），只留 installModel/height/boundary 下发 + 画 canvas + PUT。
> **单位**：cm 串走 qinglan declare_area 字符串路径（不 /10，与现网 FE 一致，已验设备收 cm）。**id**：顺序 0..N-1（弃 sticky，FE 越界 id 16+ bug 消失）。
> build+vet 绿，data 部署 23:16；**R4**：FE 须硬刷新后再 save（否则旧 FE 仍下发=双写）。待验：真 save 看 data log `declare downlink ok` + 设备收 capped-16。
> ⏳ 小尾巴：FE 读 PUT 响应 `warnings` 弹 toast（现警告在 data 响应+log，FE 未显示）；FE 未用 import `buildDeclareAreaStringFromAreas/DECLARE_AREA_MAX` 待清。

### ~~T2 原计划~~ —— **clean cutover（无灰度双路）**
> ⚠️ owlBack 规则 #1.2 禁 `if v2 else v1` 双路/shim。**不设运行期开关**；靠 T1 shadow 已证后端==FE 产物后，**一刀切**。
- data 新增「下发委托」：编排器产出 `area_{i}_*`（cm）→ HTTP → qinglan `SetDeviceProperties`（已存在 `radar_service.go:260`，cm→dm /10 + cap + 逐块下发都已就绪）。
- **FE 改造（你中断的那个 save）**：`Toolbar.vue:unifiedSave` **删** `updateRadarAreas`/`assignAreaIds`/`buildDeclareAreaString`/`executeRadarCommands(declare_area)`，**只留 PUT layout 几何**；`radarUtils.ts:709-752`(`updateRadarAreas`) + 相关编排函数**整体下线删除**（不留 deprecated）。
- 切换原子性：同一 PR 内 data 接通 + FE 删 MQTT 下发，避免双写覆盖 declare（风险 R4）。

### T3 — qinglan 失败告警（DeviceFailure）
**T3a（已落 6-25）— write-firmware 失败 → DeviceFailure**
> `radar_service.go SetDeviceProperties` 顶部加 `defer`：任一失败路径（MQTT 发布失败 / `waitForResponse` 10s 超时 / 设备返回非 200 码）→ `publishDeviceFailure`。
> ACK/超时判定本就存在（`sendOneSetProperties` 已 `waitForResponse` + 判 code），缺的只是失败时发告警——已补。
> `publishDeviceFailure`：`GetDeviceStoreInfo` 反查 addr → `NewSingleItemMessage(addr,"","Radar",ts,"alarm",DeviceFailure,data)`（data 带 `device_failure=1`+`device_code`+`reason`）→ `streamPublisher.PublishAlarm`。
> DeviceFailure 在 `alarmGateSkip` → 绕 enablement gate + HIPAA 强审计 + `DedupWhileActive` 防刷。唯一调用方=雷达配置写 HTTP（OTA 走另一条 MQTTPublisher，不误触）。build+vet 绿，qinglan 已部署。
> **验**：对离线/NAK 设备触发配置写 → `iot:alarm:stream` 出 DeviceFailure。

**T3b（待办）— 超 16 截断 → DeviceFailure**
- `radar_encode.go sortCapDeclareAreaMaps/String` 现静默截断（仅 log）。需：发生截断时回传 truncated 标志 → 发 DeviceFailure（payload 带丢弃区列表）。
- data 预检超 16（D2-A）同步回 FE 422 + 区清单。
- 注：超 16 不算"写失败"（写仍成功 16 区），是独立触发，与 T3a 分开。

### T4 — feedback 改走 SaveRoomLayout（修违规 + 消盲）
- sensor `feedback.go:405 appendFeedbackObject` 直写库**删除** → 改 HTTP 调 data 新接口 `/internal/layout/feedback`（stamp 反馈对象进 canvas → 复用 `SaveRoomLayout` 全链 → 触发 T2 下发，按 D1 规则）。
- sensor `feedback.go:206` 直读 `alarm_events`**删除** → 由 data 在 `/roomengine/feedback/ingest` 触发前回传 payload（sensor 不连库，记忆 `sensor_asks_data_sync_not_db`）。
- 顺修 `layout_load.go:29/133`、`alarm_enablement.go` 直连库 → 走 data 只读 API（可独立排期，非本次阻塞）。
- 验：E598 钉 Sit/噪声区即时生效（关掉 `feedback_pin_zone_reload_gated_not_live` 的 reload 盲窗）；sensor 无 Postgres 连接。

### T5 — 收口
- 删 FE 区域编排死码；删 qinglan 静默截断 log-only 分支；文档对齐。

---

## 6. 风险与红线
- **R1 几何漂移**：FE→后端 port 旋转/坐标若不一致 = 区域整体偏移 → 床/门错位 → 系统性漏报。T1 必须逐区 diff 对齐现网产物再推进。
- **R2 feedback 下发屏蔽区**（D1）：固件侧屏蔽引擎救不回 → 学错即真盲区。务必 cap 让位 + Human 神圣 + source 标记 + 上限。
- **R3 area_id 抖动**：粘性丢失 → 每次 save 重 declare → 设备频繁重配。D4 必须读旧 id 复用。
- **R4 双写期**：T2 灰度期 FE/后端不可同时下发（否则互相覆盖 declare）。开关二选一硬隔离。
- 红线：禁 unit test，真 case 验；sensor 禁直连 Postgres（Redis 可）。

---

## 7. 验证清单
- [ ] T1：后端 declare set 与现网 FE `declare_area` 逐区对齐（id/type/旋转顶点）。
- [ ] T2：FE save 只产生一条 PUT layout，无 MQTT declare；设备 declare_area 与库一致。
- [ ] T3：17 区 / 离线设备 → DeviceFailure 上 `iot:alarm:stream` + card 可见 + 丢弃清单正确。
- [ ] T4：feedback 钉区即时生效无 reload 盲窗；`ss`/连接审计确认 sensor 无 Postgres 连接。
- [ ] 真设备回归：床/门/噪声区位置零漂移，床占用/摔倒链路零回归。

---

## 8. 关联记忆
`areatype_unified_single_source`（AreaType 单源 0-9）· `radar_area_vertices_must_apply_angle`（顶点必旋转）· `sensor_asks_data_sync_not_db`（sensor 不连库）· `feedback_pin_zone_reload_gated_not_live`（reload 盲窗/SourceHuman 护栏）· `feedback_position_from_payload_handoff`（feedback 改事件触发已上线）· `bugB_arealying_handoff`（9 值枚举/qinglan 下发核对）。

---

## 9. 决策补充（6-26）

### 9.0 进度快照
- **T2 cutover 已落+部署**：data 单一下发入口（编排+16cap+dm 串+删[N..15]清残留+失败 DeviceFailure+截断 warning）。FE 删自身 declare 下发。
- **C 方案 zone status 已落**：`GET /room/:id/zones`（意图态）+ `GET /room/:id/zones/verify`（意图∪设备 diff，✓/X）。RadarStatusBar 区域行 = `db:意图 iot:✓/X 设备RAW dm`。db=意图标签，iot=设备原值（dm 不×10）。
- **T4 已落（part1+2）**：computeDeclareSet 收 `source!=Learned`（Human+Feedback pin 都下发）；sensor `appendFeedbackObject` 改 POST `data /internal/radar/feedback-object`（原子 append+即时下发+config:card），不再直写 room_visual_layout。feedback 对象名统一 `(pin)`。

### 9.1 读写规则（拍定，修订旧"sensor 不连库"铁律）
- **写 = 每域唯一写入口（硬规则）**：layout→data（SaveRoomLayout/AppendFeedbackObject）；**cell/28 snapshot→sensor 自己的域（保留直写）**；alarm→iot。
- **读 = 务实**：自己域直读；跨域**优先 data 推/问**但不禁止（读不产生写冲突，危害小；schema 耦合才是成本）。
- ⇒ 28/cell 直读直写**不算违规**；layout 读长期走 data（不急）；alarm_events 读 → 见 9.2。

### 9.2 ✅ alarm_events 读 → data 推 payload（待做）
- sensor 读 alarm_events 是取反馈事件 payload（operation/position-canvas坐标/notes）。
- **改法**：data handle 告警后 POST sensor `/roomengine/feedback/ingest` 时**直接带上 payload**（现在只带 event_id）→ sensor 不再回读 alarm_events。零额外跳、零 schema 耦合。

### 9.3 ✅ FE 去掉 AreaID 编排（待做）
- 删 `assignAreaIds` + `updateRadarAreas` 的 area_id/cap/declare 赋值；**保留几何转换**（`toRadarCoordinate`/`getObjectVerticesInRadar`，实时 track overlay 用）。area_id 彻底归 BE。
- 理由：BE 全删/全 update + 自定 area_id(0..N-1)，FE 那套（复用旧 id/越界 16+）已和 BE 不一致，两套必 drift。

### 9.4 FE 死 import 清理 → **放最后**
- 指删没被引用的 `import { buildDeclareAreaStringFromAreas, DECLARE_AREA_MAX }`（非 import 功能）。
- ⚠️ **导出→导入 layout→手工 Edit/bind/save 流程可能仍用 `buildDeclareAreaStringFromAreas`**——**先确认 import 功能正常再清**，否则不动。

### 9.5 ✅ boundary/height/install_model 迁 data（待做）；设备操作不动
- **迁 data**：boundary / height / install_model（layout 派生的安装配置，和 areas 同源同流；可一并收掉 FE Step2/3 设备写逻辑）。
- **不动**：重启 / 校准 / 改 WiFi（一次性设备操作，非配置 drift 源，留 FE→qinglan）。

### 9.6 待办队列（按序）
1. 9.2 alarm_events → data 推 payload。
2. 9.3 FE 去 AreaID 编排。
3. 9.5 boundary/height/install_model 迁 data。
4. 9.4 清 FE 死 import（确认 import 功能后）。

---

## 10. T6 — handle/feedback 编排彻底搬 data + sensor 被动化（6-26 拍定，待实施）

> 9.0/T4 只把 `appendFeedbackObject` 改成 POST data，**业务编排仍在 sensor**（`AlarmFeedbackIngester` 全套）。T6 把 handle/feedback 业务**整体搬 data**，sensor 退成"只被 data 调的刷/清 grid 被动腿"，并修掉 E598「resize 后 chair 不生效」。

### 10.1 一句话
护士 handle+PIN 的全部业务（解析勾选、生成 layout 对象、下发、刷 grid、清 grid）**只在 data 一处**；sensor 不再有任何 feedback 逻辑，只暴露 `stamp`/`veto` 两个无判断的被动接口。

### 10.2 正确链路（target）
```
FE handle+PIN ──PUT handle──▶ data（唯一 handle 业务，HandleAlarmEvent 后接 handlePinZones）:
  解析 remarks marker（parseConditions 子集 + spec 表，port 自 sensor）
  ┌ PIN 类(sit/lounge/reflector/interfere/blind/enter):
  │   positionFromPayload(TriggerData) → toCanvasCoordinate → fire 心正方形(rotate=0, half=20)
  │   ① AppendFeedbackObject  → layout UPSERT + ② applyDeclare(firmware) + config:card
  │   ② firmware: firmwarePolicy 现状（Chair/Blind skip · Recliner→Deny1 · Enter4 · Reflector/Interfer3）
  │   Q2: 把生成的 rect 对象(坐标+名字)写入 alarm_events.metadata
  │   ③ POST sensor /roomengine/cell/stamp {device_addr, rect, area_type, conf}
  └ sticky veto 类(Never auto-suppress):
      POST sensor /roomengine/cell/veto {device_addr, x, y, sticky=true}

sensor（被动腿，零业务判断）:
  POST /roomengine/cell/stamp → RoomForDevice → engine.StampPriorRect(rect, area, conf)
  POST /roomengine/cell/veto  → VetoCell → ClearNonHumanLearnedZone(+sticky 时 MarkLearnBlocked)
```

### 10.3 已拍定决策（本轮闭合）
- **Q1 完整 rect 来源**：handle 时只有 fire 点 → 在 layout 生成「fire 为心、rotate=0 正方形」（初始 40×40，沿用 `feedbackObjectHalfCm=20`）；完整 90×70 由用户事后 FE resize+save 经 `SaveRoomLayout` 产生。**「stamp 跟随 layout 写入」**：`/cell/stamp` 调用做成 data 内共享函数，`AppendFeedbackObject` 与 `SaveRoomLayout` **都调** → resize 后即时刷完整 rect = **E598 修复**（不再等重启 RegisterRoom）。
- **Q2 marker→data**：选 A。data handle 解析 remarks marker 生成 layout 对象；并在 handle 后 **append 人读行到 `alarm_events.notes`**（notes-only 单源，不写 metadata 避双写）：`fall(cx,cy) add <Type>(x1,y1;x2,y2;x3,y3;x4,y4)`（canvas 坐标，4 角同 layout 顶点序）。
- **Q2b 坐标转换时机（6-26）**：fire→canvas **统一在 data handle 自转**（`toCanvasCoordinate`），不在存 fire 时转。理由：`alarm_events.payload.position_*` 全腿统一保持 raw（固件腿 cardagg AlarmRouter 从 MonitorBuffer 补 raw，无 layout；DBN 腿 sensor 有 mount 但只 DBN 转→两腿不一致更糟）；canvas 派生收口 data 一处（layout SoT）。sensor `evidence.fire` 虽已 canvas，但跨域读+按 track 非 event_id 关联，不用。
- **Q3 cell-counter 学习回路**：`routeFeedback`（RestZoneConfirmed/GhostCount/MarkRestZone/IncrFakeAlarm）**整条删除**，不迁移。pin 区改纯靠 layout+StampPrior，不再 counter 累积。
- **A clear/veto 保留**：`/roomengine/cell/veto` + `VetoCell`/`ClearNonHumanLearnedZone`/`MarkLearnBlocked` 全保留。两条清除腿都由 data 驱动通知 sensor 清回 `AreaUnknown`：(1) data 删 layout 对象且 `source==Feedback` → `detectAndNotifyFeedbackVetoes`（**已存在**于 SaveRoomLayout，保留）；(2) handle 勾「Never auto-suppress(sticky)」→ data 解析 `stickyVetoMark` → 走 veto 端点带 sticky 标志 → `MarkLearnBlocked`。
- **B Recliner firmware**：Recliner-pin 下发 `AreaDeny`(1，家具装饰无人占用)。= `firmwarePolicy` 现状，**②不改**。

### 10.4 改动清单（file:line）

**sensor — 删业务，留被动腿**
- 删：`internal/roomengine/feedback.go` 的 `AlarmFeedbackIngester`/`IngestFeedback`/`processOne`/`routeFeedback`/`pinFeedbackZone`/`appendFeedbackObject`/`parseConditions`/`feedbackObjectSpec` 表（`feedbackSitObject` 等）。
- 删：`cmd/wisefido-sensor/veto_http.go` 的 `/roomengine/feedback/ingest` 路由+处理体。
- 新增：`POST /roomengine/cell/stamp` `{device_addr, rect{x1,y1,x2,y2}, area_type, conf}` → `RoomForDevice` → `engine.StampPriorRect`。纯被动。
- 改：`/roomengine/cell/veto` body 加 `sticky bool`；`engine.go:571 VetoCell` 加 sticky 入参 → true 时 `MarkLearnBlocked`（原在 routeFeedback verified 分支，随删迁此）。
- 保留：`engine.go StampPriorRect`/`grid.go SetPrior`/`cell.go ClearNonHumanLearnedZone`/`MarkLearnBlocked`/`isSuppressiveArea`/`LearnBlocked` 持久化(persist v9)。

**data — 吸收 handle 业务**
- 改：`internal/service/alarm_event_service.go HandleAlarmEvent`（~L780）DB update 后接 `handlePinZones(ctx, event, remarks, operation)`。
- 新增：marker 常量 + `parseConditions` 子集 + `positionFromPayload`（port 自 sensor feedback.go），spec 表（typeName/area/conf）。
- 复用：`declare_orchestrator.go:195 toCanvasCoordinate`（radar→canvas，mount 取自 room canvas 的 radar 对象）+ `AppendFeedbackObject`（①+②+config:card）。
- 新增：data 内共享 `stampSensorCells`（POST `/cell/stamp`），`AppendFeedbackObject` 与 `SaveRoomLayout` 都调（Q1）。
- 新增：handle 时把 rect 对象写 `alarm_events.metadata`（Q2）。
- 删：`internal/service/sensor_feedback_client.go`（`notifySensorFeedback` + `/feedback/ingest` 调用）。
- 保留：`sensor_veto_client.go detectAndNotifyFeedbackVetoes`/`notifySensorVeto`（A-1）；新增 sticky-veto 触发（A-2，从 handle remarks）。

**不改**：`firmwarePolicy`（B）、`feedbackObjectHalfCm=20`、优先级 gate（按高风险非来源）、`StampPriorRect` grid==nil 静默 return（防单 room 卡死全局）。

### 10.5 验证
- E598 chair：resize→save 后 grid 完整 cell 立即 = Sit（不靠重启）；fire 点初始 pin 40×40 即时 Sit。
- sensor 无 `/roomengine/feedback/ingest`；`AlarmFeedbackIngester` 整套不存在；sensor 无任何 alarm/feedback 业务判断。
- handle sticky veto → 该 cell 清回 Unknown + LearnBlocked 跨重启保留。
- 摔倒链路零回归（真 case，禁 unit test）。

### 10.6 风险
- **R6-1 坐标系**：fire 点 raw(h,v,z) 在 data 内转 canvas 依赖 room canvas 里的 radar mount 对象；若该房 layout 缺 radar 对象 → `toCanvasCoordinate` 无 frame。需边界判：取不到 mount → 跳过 pin（log warn），不 panic。
- **R6-2 双写期**：sensor 删 `AlarmFeedbackIngester` 与 data 接 `handlePinZones` 必须同批上线，否则 handle PIN 落空或双刷。
- **R6-3 clear 粒度**：veto 是 point/cell 级（`VetoCell x,y`），stamp 是 rect 级；保持现状不对齐（清按中心点，刷按矩形），符合既有语义。

### 10.7 实施状态（6-26，build+vet 绿，待真 case 验）
- **sensor**：删 `internal/roomengine/feedback.go` 整文件（`AlarmFeedbackIngester`/`parseConditions`/`routeFeedback`/`pinFeedbackZone`/`appendFeedbackObject`/`FeedbackEvent`/marker/spec/`positionFromPayload`）；删 `engine.go` 的 `feedbackIngester` 字段+wiring+`IngestFeedback`；删 `veto_http.go` 的 `/roomengine/feedback/ingest`。新增 `engine.StampPriorRectByDevice` + `POST /roomengine/cell/stamp`（被动）。`VetoCell` 加 `sticky` 入参（false 仅清 / true 额外 `MarkLearnBlocked`）。cell-counter 基建（`GhostCount/RestZoneConfirmed/RealFallCount` + persist v5 + room_svg overlay + playback `MarkRestZoneFeedback`）**保留**（非 routeFeedback 独占，replay/可视化仍用）。
- **data**：新增 `handle_pin_zones.go`（`RadarInstall.HandlePinZones`：解析 marker→`positionFromPayload`→`toCanvasCoordinate`(复用 FE 移植，零坐标漂移)→fire 心 40×40 方块→`AppendFeedbackObject`(①+②+config:card)→`appendPinNote` append 人读行到 `alarm_events.notes`(Q2 notes-only)→`notifySensorStamp`(③)；sticky→`notifySensorVeto(sticky=true)`)。`applyDeclare` 尾加 `stampCanvasCells`（**stamp 跟随 layout 写入**：`SaveRoomLayout`+`AppendFeedbackObject` 都走→FE resize 后即时刷完整 rect = **E598 修复**）。`sensor_veto_client.go` 加 `sticky`+`notifySensorStamp`。删 `sensor_feedback_client.go`（`notifySensorFeedback`/`/feedback/ingest`）。`alarmEventService` 注入 `radarInstall`，`HandleAlarmEvent` 改调 `HandlePinZones`。
- **未验/待办**：真 case（E598 chair resize→save 立即全 Sit 不靠重启）；R6-2 两服务同批部署；`HandlePinZones` 的 `tenantID` 形参暂未用（留作签名）；Xsensorv1/playback 镜像未做；orphaned cell-counter 基建按 rule 1.2 后续可清（涉 persist v5 schema，单独排期）。
