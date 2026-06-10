---
name: sleepace_timezone_vendor_ignores_bind
description: "sleepace 报告时间错根因=owl provisioning 传错时区(-25200默认)非厂家丢弃;治本=ResyncDeviceTimezone 重传 unit offset"
metadata:
  node_type: memory
  type: project
  originSessionId: bd3ad766-aa20-4995-af63-01417bda5e91
---

2026-06-03 查清 sleepace sleep report 时间错(差 15h,设备在东八区却显示 -25200)根因 + 治本。**注意:本条已纠正早前一个错判。**

**根因(反编译厂家 jar 2025-09 版字节码验证):**
- sleepace mysql `sleepace_pro_user.time_info(userId,timeZone秒)` 是**每用户时区表,驱动 `history_summary_*.timezone` 与报告切天锚点**;铁证 userId 399977 time_info=-21600 → 其报告 timezone 全 -21600。
- **bind 会消费 timezone 并更新 time_info**(早前误判"厂家丢弃",已纠正):`DeviceBindController.bind` 读请求 `timezone` 字段 → `UserDeviceServiceImpl.bindDevice(...,int timezone,...)`:`if (传入tz != time_info.getTimeZone()) { setTimeZone(传入tz); TimeInfoService.replace(); }`。即**传入时区与已存不同就更新**。
- 所以"rebind 改不了时区"是因为 **owl 每次都传同一个 -25200**(判等跳过),不是厂家丢弃。time_info 的 -25200 也不是 JVM 默认(本机 JVM=PDT raw -28800),而是 **owl 传过去的 -25200**(`DefaultTimezoneOffsetSeconds=-25200` 全局 Denver fallback / wisefido-sleepace cfg `timezone:-25200`)。
- TimeInfoServiceImpl.queryByUserId 对全新用户才用 `TimeZone.getDefault().getRawOffset()` 兜底;bind 路径会覆盖。AuthController(第三方 App 登录)也能写,但 `app_id_key` 回调表为空未配,且非必需。

**治本(全程 owl 侧,不需厂家/不写 mysql/不做显示层 hack):** owl 已有 `ResyncDeviceTimezone`([device_monitor_settings_service.go:717-728](../../../owl/owlBack/wisefido-data/internal/service/device_monitor_settings_service.go))——按 device 的 unit IANA(owl_v2 `units.timezone`,两台=Asia/Shanghai)算 `IANAToOffsetSeconds`→+28800,再 InitializeDevice→bind 下发;bindDevice 见 +28800≠-25200 → 更新 time_info → 次日报告 timezone=+28800、锚点对。多时区天然支持(每台传自己 unit offset)。端点 POST `/internal/sleepace/device/{addr}/resync-timezone`。
**连带坑:** ResyncDeviceTimezone 走完整 init 级联,会把 `report_upload_time` 又推成 wisefido-sleepace cfg 的 **10**(test/dev yaml 都=10),覆盖 tenant 的 `DefaultSleepReportTime=8`;实测该设备 05-12 还是 8、06-01 变 10。需同时把 cfg 改 8 或让级联取 tenant 值,再 resync。

关联 [[sleepace_userid_is_device_uid]] [[sleep_report_storage]]。

**2026-06-03 已落地+实测(commit 9e652b3):**
- **DST 含夏令时本就对**:`IANAToOffsetSeconds` 用 `time.LoadLocation`+`time.Now().Zone()` = DST-aware;实测 Denver=-21600(MDT)/Shanghai=28800/LA=-25200(PDT);**201(BM87224601641,unit America/Denver)实绑 -21600**。
- **统一执行动作(用户铁律:自动/手动只是触发,动作必须同一份)**:`SleepaceDstRebindScheduler` 重构成调 `TimezoneResyncer.ResyncDeviceTimezone`(=Setting save / 手动端点同一个),不再自己内联 InitializeDevice。调度器只管 WHEN;main.go 传 deviceMonitorSettingsService。
- **启动即纠偏 startup_corrective**:旧逻辑只在 DST 切换当天重绑→错过当天(owl 宕机/新装漂移)要等半年;新增启动无条件重绑全 sleepad 到当前 offset,每日 tick 仍只切换日触发。**退避重试**(data 先于 sleepace gateway 起好则全失败→等就绪重试 5 次)。
- **report_upload_time**:sleepace-**dev.yaml** 10→8(run script `-env dev` 跑 dev.yaml,test.yaml=10 不用);cfg 仅 bind 时兜底初值,长期应由 wisefido-data per-branch 每日重算(换算到 sleepace 服务器 LA 时钟)own。
- **遗留**:批量重绑厂家偶发 500(每轮~4/10 瞬时,轮替,下次 save/切换补);report_upload_time 治本=让 wisefido-sleepace bind 别无条件用 cfg 覆盖。
