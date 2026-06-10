---
name: sleepace-interval-scheduler
description: Sleepad 按 rest/nap 时段自动切换 realtime_interval (2s↔10s) 的 scheduler 设计；bound 过滤+IANA DST+vendor-unbound backoff+用户配置尊重四项要素
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 0ef86edd-98ed-4293-bb19-0c02b8af18a4
---

[SleepaceIntervalScheduler](../../../owl/owlBack/wisefido-data/internal/service/sleepace_realtime_interval_scheduler.go) 60s tick，按时段切 sleepad realtime_interval：
- rest 窗口 `[ResetTime.InBedTime-10min, ResetTime.OutBedTime+2min]` 跨午夜
- nap 窗口 `[NapTime.InBedTime-10min, NapTime.OutBedTime+2min]` 不跨午夜
- 窗口内 → 2s 高频，窗口外 → 10s 低频

**Why:** sleep monitoring 需要高频采样（睡眠分期、apnea 算法依赖密集 RR/HR 数据），但 24/7 都 2s 浪费 broker 流量 + 厂家 quota。原方案 bind 时硬下发 2s 全天不切，"在 rest 期高频"是设备应用最自然的边界，加 ±pad 避免临界跳变。

**How to apply:** 改 sleepace interval 策略 / 加新时段 / 新 vendor 同款 scheduler，复用这 4 个关键设计：
1. **bound 过滤**：listSleepadDevices 用 `INNER JOIN units` 排除库存（v2 dev 环境 61 台 sleepad 实际只 8 台 bound，未过滤会刷 53 条 "user not found" warning）
2. **vendor-unbound backoff**：bound to unit 但厂家未 InitializeDevice 的 device，捕获 `user not found (status: 5)` → Redis 1h TTL backoff，避免每 60s spam；`InvalidateUnbound(deviceID)` 已 wire 到 `device_store_service.applyDefaultSleepadRealtime` 成功 path（main.go `deviceStoreService.SetOnSleepadVendorBound(...)`）→ 重新 bind 后下次 60s tick 立即重新 push，不等 TTL
3. **timezone+DST**：unit-level IANA tz（如 Asia/Shanghai / America/Denver），`time.LoadLocation` + `now.In(loc)` 自动按当年 DST 解析；不要存 UTC 算偏移（DST 切换日会错 1h）。多 tenant/branch 不同 tz 同时跑无冲突
4. **用户配置尊重**：从 `spatial_config.alarm.device_config` 读 `SleepadSetting.realtime_interval`；≥ 10 视为用户主动选低频，scheduler 跳过该 device（不强制覆盖人工设置）

dedup: Redis hash `sleepace:realtime_interval` 存上次推送值，相同跳过 set 调用，减少厂家 API 量。

相关：
- [[sleepace-vendor-param-limits]] 厂家硬限范围（HR/BR），同样需要尊重
- [[short-code-alias-resolve-everywhere]] 同样 silent miss 模式（这次是"未 init device 持续 retry"对应"短码下游漏 resolve"）
