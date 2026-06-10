---
name: v2-cutover-type-assert-silent-drop
description: v2 service 缺方法 → main.go type assertion 默默失败 → 整段下发不执行；改 v2 service 时检查 main.go 用 if-ok 注入的所有可选接口
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 0ef86edd-98ed-4293-bb19-0c02b8af18a4
---

v2 service 把可选 setter 当 optional interface 通过 type assertion 注入时，缺方法 → assertion 失败 → 注入静默跳过 → 调用方拿到 nil gateway → 全部下发被吞，但**没有任何 error/warn 日志**。

**Why:** 2026-05-14 sleepad 设备配置看起来 save 成功（FE 绿色 toast），但厂家侧值从未变。根因是 v2 `DeviceMonitorSettingsV2Service` 没实现 `SetSleepaceGatewayClient`，[main.go:357-363](../../../owl/owlBack/wisefido-data/cmd/wisefido-data/main.go#L357-L363) 的 `if svc, ok := ...interface{...}); ok` assertion 失败，gateway 没注入，`UpdateDeviceMonitorSettings` 硬编码 `device_write: false`。FE 字段名又错配（v2 改成了 `database_write` 但 FE 还读 `db_write`），把 false 当 undefined → 永远显示 "DB OK"。

**How to apply:**
- 改 v2 service 时 grep main.go 找 `interface { Set*Client }` 模式（也叫 ad-hoc capability detection）— 每一个都是必须实现的契约，缺了无报错
- 给注入路径加成功 log（`logger.Info("X client set for Y")`）；缺了的话 startup 应能看出
- FE/BE 字段名改名时双向 grep；命名升级（v1 短名→v2 全名）时 FE 跟着升级，不要 backend 退回 v1 名
- 半失败状态（DB OK + device fail）UI 必须 `message.warning` 而不是 `message.success`，否则用户看不到差异

相关：[[sleepace-vendor-param-limits]]（v2 cutover 把链路接回后才暴露的下游 quirk）
