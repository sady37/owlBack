---
name: dev-stage-owlback
description: 开发阶段允许直接 stop 整个 owlback (所有子服务一起停)；重启时跨服务依赖更易暴露 wiring 问题
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 102f29bf-0317-4351-b343-2f5e6fe0d658
---

开发阶段，可以直接 `sudo systemctl stop owlback`（umbrella）把所有子服务一起停。

**Why**：单个子服务重启时，其他服务的连接/缓存/状态可能掩盖问题；整套一起重启 = 冷启动，跨服务 wiring 错（schema drift / Redis key 不一致 / IPv6 寻址 mismatch / consumer group 重建）会直接暴露在 startup 日志。

**How to apply**：
- 改了 owl-common 或动了跨服务约定时 → 停整套再启，不要单服务热重启
- 大型 cutover（v2 schema 切换 / 协议升级 / iot_timeseries 退役）→ 必须整套停
- 单服务小改 fix 仍可单独 restart（无跨服务影响时）
- 不需要每次问授权，直接停起

这是 [feedback_dev_restart_services](feedback_dev_restart_services.md) 的扩展：单服务 restart 是常规手段，整套停起是 cutover/wiring 验证手段。
