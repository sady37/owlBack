---
name: alarm_handle_form_by_category
description: Alarm Handle 弹窗按 FHIR category 分支 Observed Conditions/False Reason；前端镜像后端 alarm.go 分类
metadata: 
  node_type: memory
  type: project
  originSessionId: a5f99561-076d-4aa0-8862-a3ee1ff10db3
---

2026-05-30 落地：Handle Alarm 弹窗的 Observed Conditions / False Alarm Reason 按 **FHIR category + event_type** 分支。

**规则（真相源 = 后端 owl-common/alarm/alarm.go `AlarmTypeToFHIRCategory`）：**
- **device 类**（Offline/OfflineRecover/DeviceFailure/DeviceRecover/SignalPoor(+Recover)/AngleException(+Recover)/SensorDetached(+Recover)/pressureSensor）→ **无 Observed Conditions、无 False Reason**，弹窗只剩 verdict + Remarks
- **非 device verified** → Observed Conditions **统一一套**（fall 5 项：Awake/Responding/Unresponsive 单选 + Visible Bleeding + Fall/Sitting 多选），不按 event_type 变
- **False Reason 按 event_type**（`FALSE_ALARM_REASONS_BY_EVENT_TYPE`）：Fall/SittingOnGround→雷达 8 项；vital（HeartRateAlert/RespRateAlert/ApneaHypopnea）→`VITAL_REASONS` 3 项(Electric-AC/Out-of-range/Other)；LeftBed→5 项；NoBodyMove/BedSitUp 各自一套

**前端实现（`owlFront/src/utils/alarm.ts`）：**
- 镜像后端表为 `ALARM_TYPE_TO_CATEGORY` + `getAlarmCategory()`（兜底 'safety'，同后端 `GetFHIRCategory`）
- `baseEventType()` 剥 `.High`/`.Low` 后缀再查表 —— 否则 `HeartRateAlert.High` 漏映射掉进 Unknown
- `isDeviceAlarm = (record.category || getAlarmCategory(event_type)) === 'device'`；record.category 是后端真相优先，缺失才派生

~~两处重复实现~~ → **2026-05-31 已收口为单一共享组件 `src/components/alarm/AlarmHandleModal.vue`**。AlarmRecordList + Overview 都 import 它,各自只负责"取数 + format 出展示字段(resident/address/event/device/occurred/alarmTime) + replay 动作 + handled 后刷新",组件负责 template/CSS/表单态/remarks 拼装/提交(`handleAlarmEventApi`,唯一 callsite)。props: visible(v-model)/alarm{event_id,event_type,alarm_status,category}/各展示字段/replayAvailable/replayLoading;emits: update:visible/handled/replay。**两边历史分歧已统一**:提交一律 `auto_resolved?resolved:acked`(Overview 旧硬编码 acked 已废)、Observed/CSS/verdict 按钮一致。新增 Handle 弹窗字段只改这一个组件。

**用户决定**：~~Apnea 从原独立 reasons 改并入 vital 三项~~ → **2026-05-31 拆回**：ApneaHypopnea 走接触式 sleepad 专属 `APNEA_REASONS`（5 项：Empty bed / Pet on bed / Pad misplacement / User awake / Unknown），不再用雷达口径 VITAL_REASONS（接触 pad 无"Out of Detection Range"语义）。Empty bed 与 Pet on bed **拆两个 key**（不同处置 + 不稀释训练信号，对齐 LEFT_BED/NO_BODY_MOVE 既有 petOnBed/emptyBedMissed 分开的模式）。HR/RR 仍用 VITAL_REASONS。

教训同 [[feedback_read_design_before_modify]]：这次先排版后才发现 Observed Conditions 控件根本没接线（verified 选项不进 remarks）+ 没按 category 分支 —— 改交互表单要先确认数据流/分支接线，UI 不能跑在逻辑前。

---

**Handle 权限规则（2026-05-30 拍板，只认 unit_property，与 B2B/B2C 脱钩）：**
- `unit_property == home`（居家）→ **family + 机构 staff 都能 handle**（业态=互助/有限服务护理：机构有限时段服务 + 家属共担，责任共担非排他，留痕 handler_name/handled_at 满足问责）
- `unit_property == facility`（住机构内）→ **仅机构 staff**（Admin/Manager/Nurse/Caregiver 白名单），family 只读（机构物理托管、排他全责）
- **不看 tenant.kind**：B2B 机构可服务居家老人（B2B+home 单元），那张卡仍按 home 放行 family。曾误判"B2B+home 让 family 处理违规"——前提是排他责任，互助共担模式下不成立。

**用户模型（纠正 stale 假设）：**
- **没有 resident 登录用户**；`resident` 只是机构里被照护人的信息记录，无 login。resident 当用户无意义（老人自己 Fall 收报警没用）。
- **family 才是真正用户**：为 resident 买看护服务 + 负责处理报警。role 字符串 = `"Family"`（后端 scope.go `IsFamily()` EqualFold）。
- 前端权限按 `role` 判（user.ts:56：userType 不用于权限）；`canHandleAlarm` 旧的 `userType==='resident'` 拦截是对着不存在的角色防御，应删，改成 `role==='Family' && unit_property==facility → deny`。

**落地（2026-05-30 完成）：**
- **执行层=后端** `checkHandlePermission`（alarm_event_service.go:1316）已按 unit_property 正确分流，无需改。
- **拦截层=前端** `utils/alarmPermission.ts`：`canHandleAlarm(unitProperty?, userStore)` —— `role==='family' && unit_property===1(facility)` → false，其余放行。role 取登录态 userStore（login 返回）；unit_property 由调用方从 **CardStatic.unit.unit_property**（local store / props）取传入，**不从 alarm API 拿**。AlarmRecordList 传 `props.unitProperty`。
- **不需要后端给任何新信息**：event_id + 显示字段每个入口本就有（能提交即证明有）；unit_property 在 CardStatic store；role 在登录 store。曾走弯路想加 `GET /:id` + 往 alarm 响应塞 `unit_property`，均**已回退**（owlBack 规则 1.2 不留无用代码）。教训：加权限/功能前先问"现有数据够不够"，别急着让后端造数据。
- Overview 的 Handle 不经 canHandleAlarm（仅 AlarmRecordList 用），其 family 拦截靠后端兜底；如需前端预隐再传 cardMap 的 unit_property。

**Observed Conditions 也做成 per-event-type（2026-05-31）**：utils/alarm.ts 加 `OBSERVED_CONDITIONS_BY_EVENT_TYPE` + `getObservedConditionOptions()`（与 False Reason 结构对称）。反应度单选(Awake/Responding/Unresponsive)组件内通用硬编码;复选项 = 通用项(`visibleBleeding` 置顶) + per-type，按 **状态→处置→转归** 排序;跨类同义观察复用同一 key(`foundOnFloor` 在 Fall/LeftBed/NoBodyMove/BedSitUp;`repositioned` 在 Apnea/NoBodyMove)便于聚合。组件 Observed 区改成动态 v-for 渲染(不再硬编码 visibleBleeding/fallSitting)。全是写 remarks 自由文本、无后端权威、不驱动分诊（[[feedback_no_dynamic_threshold_modulation]]）。

与 [[tenant_kind_b2b_b2c]] 区分：那条讲 Family 通用访问（建用户等），alarm-handle 专按 unit_property。

---

**AlarmState 新增字段必须三处同改（2026-05-31 踩坑）：** card 弹窗 AlarmDevice 显 `-`/无 AlarmTime，Detail 弹窗正常 → 根因不是后端也不是旧构建，是 **Overview.vue SSE 回调 `handleCardStatus` 把 `alarm_state` 逐字段重建**（~line 1438），新加的 `alerted_at`/`device_addr` 没进那个 literal → `statusMap.alarm_state` 永远缺这俩 → `getLatestPopAlarm` 读到 undefined。Detail 不走 SSE/statusMap（走完整 alarm_events DTO）所以不受影响，差异由此而来。后端全链路 OK（card_types.go AlarmState struct 有字段 + Redis hash 实测有 + SSE 整对象 `json.Marshal(CardStatus)` 原样转发不裁剪）。**铁律：往 owl-common AlarmState 加字段，必须同步 (1) FE `monitorModel.ts` `AlarmState` type (2) Overview.vue 那个逐字段 merge literal (3) 用到的读侧；漏 (2) 就 card 路径静默丢、Detail 仍正常，最难查。** 这是 [[feedback_store_raw_not_derived]] 之外的同类 drift：逐字段拷贝是 silent-drop 温床，能 spread 就别逐字段列。
