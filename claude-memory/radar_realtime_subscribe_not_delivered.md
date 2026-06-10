---
name: radar-realtime-subscribe-not-delivered
description: 已解(2026-06-02) — "雷达连着却 monitor down/前端全 offline" 真根因=device_subscription_manager 自死锁(重入 m.mu.Lock)，叠加 paho 收包被 handler 阻塞；非固件非实时订阅命令
metadata: 
  node_type: memory
  type: project
  originSessionId: bfa121c6-2088-4ca1-a827-9ad91206bc28
---

**2026-06-02 已定位并修复。** 现象:雷达连着 broker、在推数据(broker mosquitto_sub 能收到),但 qinglan 收 ~10s 就停 / 45s 规律掉线 / 前端全 offline。**之前两版猜测(实时订阅命令未送达 / hairpin)都不是总根。**

**真正总根 — `device_subscription_manager.go` 自死锁(预存 bug,非 MQTT)**:`EnablePeriodicSubscription`(原 444+defer465) 与 `SubscribeDevice`(原 267+defer287) 已持 `m.mu`(Lock+defer Unlock),却在 `if deviceAddr != "" && deviceAddr != oldDeviceAddr` 分支里**二次 `m.mu.Lock()`**。Go RWMutex 不可重入 → 同一 goroutine 二次 Lock = **永久自死锁**,把 m.mu 永久占死。触发条件=设备**重认证且 device_addr 变化**(地址平时不变,故偶发、跑一会儿才中招)。锁一死:dispatchWorker/handler 调 `UpdateLastSeenByType→m.mu.Lock()` 全卡住 → 消息处理冻结 → device_status 陈旧 → 前端全 offline。铁证=SIGQUIT goroutine dump:dispatchWorker 卡 :654、8 个 auth 卡 :445、持锁者在 :465。

**鉴别要点**:device_subscription_manager 主键=**device_uid**(`subscriptionsByUID`,稳定 logMAC);**device_addr**(`subscriptionsByID`,IPv6)只是次级反查索引,死锁分支就是更新这个次级索引时踩的。

**两个共病(同一根的不同表现 + 放大)**:
- 同步 handler 跑在 paho 收包 goroutine 里,handler 一卡(撞死锁/慢 DB)收包就停 → 读不到 PINGRESP → keepalive 饿死掉线;且 **paho 掉线检测+AutoReconnect 本身也依赖这个 goroutine,一并卡死 → 既不报 lost 也不重连**(故"掉了永不回")。
- qinglan 曾被 `effectiveMQTTConnect` 推上公网 hairpin(拨 RADAR_MQTT_SERVER),那条路径脆弱致最初 45s 掉。

**已修(3 处,实测:4min 零掉线、monitor 逐秒连续、跨重认证不冻、13 雷达全 online)**:
1. **删两处二次 `m.mu.Lock()`**(map 写已在外层锁内,直接写)——**治总根**。
2. `owl-common/mqtt/client.go`:handler 移出 paho 回调,**拷贝 payload + 入队**,单 `dispatchWorker` 串行执行 → 收包永不阻塞 → keepalive 健康、paho 原生 AutoReconnect/ResumeSubs 恢复可用。
3. `wisefido-qinglan/internal/mqtt/client.go`:`effectiveMQTTConnect` 删 hairpin 改拨逻辑,一律拨 `MQTT_BROKER` 本机(不动 RadarDeviceMQTT/不影响固件下发)。

**避坑教训**:中途我叠加"客户端重建重连 + 自制 keepalive"过度工程,引入更严重故障(540 goroutine 堆积、数据全断),后回退。**MQTT keepalive 数学**:客户端 `SetKeepAlive(30s)` → broker 按 MQTT 协议 1.5× = 45s 无包即断(broker 无此配置项,纯协议);雷达无人时 monitor=track_id=88 每 60s 一帧 > 45s,故空闲必须靠 paho PINGREQ 或抬 keepalive 到 60s。

**环境遗留(待还原)**:`.env LOG_LEVEL=debug`+`QINGLAN_VERBOSE_LOG=true`、`mosquitto.conf connection_messages=true`。**3 文件改动待 commit。** 后续清理项:本文件 deviceAddr 变量名混用 uid/addr。详 [[qinglan_restart_kills_monitor_sub]][[qinglan_online_detection]][[device_status_event_driven_refactor]]。
