---
name: qinglan-restart-kills-monitor-sub
description: 运维地雷 — 重启 owlBack/qinglan 后约1小时雷达 monitor(轨迹)流全静默；event 流正常；恢复=重下实时订阅
metadata: 
  node_type: memory
  type: project
  originSessionId: bfa121c6-2088-4ca1-a827-9ad91206bc28
---

2026-06-01 实战排查:重启 owlBack 后约 1 小时,**所有青澜雷达的 monitor(radar.track 轨迹)流齐刷刷静默**,但 **event 流(EnterRoom/number_people/activity)完全正常**,sleepad HR/RR 也正常。

**根因**:青澜雷达的实时 monitor 是**按需订阅、1 小时时长**(`device_subscription_manager`:`monitorMaxAge=1h`/`defaultDuration=3600`,周期续订);event 是固件常推、不受订阅闸。qinglan **一重启就丢内存订阅状态(`monitorSubTime` map 清空),又不主动给已连接的雷达重订** → 雷达靠重启前的旧 1h 订阅续命,各自到期后掉 monitor,没人续。铁证:B197 重启前最后续订 02:43:47Z +3600s = 21:43:47Z 到期,其真实 track 正好 21:44 停。**不是编译/版本问题**(二进制没变),不是 monitoring_enabled 闸(DB 全 true;若是该闸会发 `track_id=11` 占位,实际是雷达自己的 `track_id=88` 无人 keepalive),不是 MQTT broker(sleepad+event 都正常)。

**关键鉴别**:monitor 死但 event 活 = 实时订阅过期(不是 broker/网络/电源)。看 `monitor_stream` 里 track_id:`0-8`=真人 / `88`=雷达无人keepalive(monitor订阅在) / `11`=qinglan monitoring-off占位 / 完全无 track 行=订阅过期或没订。

**立即恢复**(已验证有效):`POST http://127.0.0.1:8081/api/v1/radar/devices/{uid}/subscribe` body `{"content":0,"duration":3600}`(content 0=轨迹+vital)。遍历所有 `access AND monitoring_enabled` 的 Radar 逐个下发即可(60 台全成,~20s 后真实帧回流)。**但这是临时的,新订阅也 1h**。

**真正根因 + 已修(commit 1b7ac48, 2026-06-01)**:`device_subscription_manager.go` 有个**专为重启写的 `restoreAuthenticatedDeviceSubscriptions`**(从 DB 查所有 access+Radar,重建内存订阅记录,首条 MQTT 消息触发 monitor 命令),但 **`Start()` 从没调用它**(孤儿:grep 0 调用、日志 0 次)。`Start()` 只起 heartbeatMonitor/subscriptionRenewal/healthCheckMonitor,而 subscriptionRenewal 只续内存里已有记录(重启后空)→ 没人重订。修复=`Start()` 末尾加 `go m.restoreAuthenticatedDeviceSubscriptions(ctx)`,实测重启后 restore 触发、count=60、monitor 流恢复。以后重启不再隔 1h 静默。详 [[deploy_mode_by_domain]] [[qinglan_online_detection]]。
