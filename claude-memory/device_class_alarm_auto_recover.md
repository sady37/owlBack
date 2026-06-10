---
name: 设备类报警一律走 alarm 流 + auto-recover
description: 雷达 SignalPoor/AngleException 与 Offline/SensorDetached 同等处理；恢复信号自动解除 active alarm，不需人工标记
type: feedback
originSessionId: 2540d0c0-e995-418d-a1a4-3e2d3075ad3a
---
设备类（device-class）报警的统一原则：onset 走 `iot:alarm:stream` 入库报警，对应 `*Recover` 触发 `HandleRecoveryWithTypes` 自动解除 active alarm。**不走 event 流，不需要人工标记 false_alarm 关闭。**

## Why
- 设备健康类信号（离线/信号差/角度异常/传感器脱落）由 firmware/health-check 周期采样，状态翻转就是真实状态变化，不需人工核实
- 走 event 流后由上层"去抖判定后再决定报不报"原本是合理设计意图，但实测上层（cardagg）从未实现去抖，事件被悄悄吞掉等于黑洞
- 走 alarm 流 + 周期 onset/recover 配对的设计简单可见，直接由用户/UI 看到状态翻转

## How to apply（2026-05-03 修复路径，作为模板）

雷达侧（qinglan）：
- [`health_check.go publishDeviceAlarm directAlarm` 白名单](owlBack/wisefido-qinglan/internal/subscriber/health_check.go) 必须包含 onset + recover 两端
- 已纳入：`AlarmTypeOffline / OfflineRecover / SensorDetached / SensorDetachedRecover / SignalPoor / SingalPoorRecover / AngleException / AngleExceptionRecover`

cardagg 侧（alarm_handler.go）：
- onset case 模板：`if enabled && level != "" { PersistAlarmAndPublish(payload, eventName, level) }`
- recover case 模板：`HandleRecoveryWithTypes(payload, []string{对应的 onset 类型})`（对无 active 是 noop，幂等）

cardagg 侧（event_handler.go）：
- **不**为这些类型在 event_handler 留处理 case；旧消息走 default 静默丢弃即可
- 之前 `metaCache.UpdateStatus("angle_abnormal","1")` 那种状态更新链路是死代码（IsDeviceHealthy 函数 0 调用方），删除安全

sleepace 侧：已经对了（mqtt_consumer connectionStatus / deviceSenSor、health_check connectionStatus 都通过 `Publisher.PublishAlarm`，无需改）

## 阈值待复审
- WiFi RSSI ≤ -88 SignalPoor — 偏宽松（-88 已临界断线），考虑收紧到 -80
- 雷达墙角支架 angle 范围 -60 ≤ Y ≤ -45 — 现场实测 D523 稳定 -63.89（超 -60 仅 4°，安装现场普遍偏陡），考虑放宽到 -65 ≤ Y ≤ -45
