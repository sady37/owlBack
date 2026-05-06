---
name: qinglan radar online 检测路径
description: 设备上线判定走 path A（autoSubscribeOnFirstMessage inline OfflineRecover），不要再加无证据 fan-out
type: feedback
originSessionId: c2d0b96a-ef02-4ee6-9673-3a25811aadfc
---
qinglan 判设备上线只走一条路：[autoSubscribeOnFirstMessage](owlBack/wisefido-qinglan/internal/subscriber/device_subscription_manager.go) 在收到设备首条 MQTT 创建 subscription 后**立即 inline publish OfflineRecover** 到 iot:alarm:stream。不要再加任何"按 qinglan 自己订阅的 topic 列表 fan-out"的兜底路径。

**Why**：2026-05-04 修过一次 path B fan-out 的假阳 bug（mqtt_consumer.publishOnlineForConnectedAfterStartup，启动 +1min 扫 subscribedTopics 给所有允许设备发 OfflineRecover）。subscribedTopics 是 qinglan 自己订阅了哪些 topic（`subscribeAllAccessibleDevices` 启动时把所有 allow_access=true 设备的 topic 都订了），与"设备真在发包"完全无关。后果：每次 qinglan 重启 +60s，所有离线设备（如长期不发 MQTT 的 Radar_0777）都被假阳标 online，UI 显示在线但实际无任何数据，cardagg `offline_recover.handle` 里 set_online_ok 是无条件的，stale OfflineRecover 一进来就会把卡片刷成 online。

path B 当初被加的真原因是 path A 有个时序 bug：UpdateLastSeenByType 在 sub 不存在时早 return，OfflineRecover 直到第二条 MQTT 才触发——心跳间隔 5min 的设备就要等 5min 才上线。修这个 bug 比加 fan-out 兜底正确：把 OfflineRecover 移到 autoSubscribeOnFirstMessage 创建 sub 的同一刻，"被调用 = 收到了首条 MQTT" 是实证级证据。修完后真在线设备 0.4-30s 内上线，真离线设备永不假阳。

**How to apply**：

1. **不要恢复 publishOnlineForConnectedAfterStartup 或类似无差别 fan-out**——任何"启动 +N 秒批量发 OfflineRecover"都会重新引入假阳。设备上线必须有"收到了它的 MQTT 包"作为实证。
2. **不要把 subscribedTopics 当成"在线设备列表"**——它是 qinglan 自己监听的 topic 集合，跟设备状态无关。"在线设备"看 `subscriptionsByUID[uid]` 里有 `LastSeen != zero` 的记录。
3. **改 path A 时小心 LastSeen 时序**：autoSubscribeOnFirstMessage 创建新 sub 时把 LastSeen 设为 now（不是 zero），否则 UpdateLastSeenByType 下次仍会按"首条"再发一次 OfflineRecover——重复发不影响 cardagg（幂等），但日志会乱。
4. **cardagg `offline_recover.handle` 应只在有 active offline alarm 时 set_online**（治本，跨服务 backlog 待做）：当前无条件 set_online，让任何 stale OfflineRecover 都能把卡片标 online，是这个假阳 bug 放大的根源。
5. **慢上线诉求要新加路径时**：选项是「主动 prop read 探活 + 看响应」（同 healthCheckMonitor 但更频繁），不是「无证据 blast OfflineRecover」。

**关联实现**：[device_subscription_manager.go autoSubscribeOnFirstMessage](owlBack/wisefido-qinglan/internal/subscriber/device_subscription_manager.go) inline publish 段；[mqtt_consumer.go Start](owlBack/wisefido-qinglan/internal/consumer/mqtt_consumer.go) 已删除 publishOnlineForConnectedAfterStartup。
