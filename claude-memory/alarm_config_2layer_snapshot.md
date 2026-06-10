---
name: Alarm 配置 2-layer 快照模型
description: tenant/device 两层 alarm 配置；device 首次读时拷 tenant 快照，之后独立；不做 longest-prefix-match 实时继承
type: project
originSessionId: d5909d17-97c2-4293-a282-0c841f2e9982
---
Alarm/monitor 配置在 v2 spatial_config 中按 IPv6 prefix 分层存储，**仅 2 层**（不要 branch 中间层）：

| 层 | prefix masklen | 角色 | 写权限 |
|---|---|---|---|
| tenant | `/48` (e.g. `fd00:0:3::/48`) | tenant_admin | AlarmCloud.vue (`/admin/api/v1/alarm-cloud`) |
| device | `/128` (设备完整地址) | manager/nurse | device monitor settings page |

**Why:** 用户决策（2026-05-11）— branch 层无业务价值；2 层够覆盖"机构默认 + 个别设备特调"两个场景。

**How to apply:** 实现 device 层读取逻辑时：
1. 先 `SELECT FROM spatial_config WHERE spatial_prefix=<device_ipv6/128> AND config_key='alarm.cloud_config'`
2. 行存在 → 直接返回该 device 自己的 config（**不**回退看 tenant）
3. 行不存在 → 读 tenant prefix (`/48`) 的同 key；存在则当作"种子默认值"返回；不存在则用 `owl-common/alarm.GetDefaultAlarmItems()` 硬编码
4. **不要**在读取时实时合并/继承 tenant 与 device — 用户原话："一旦本层设定，就不受上层影响"
5. 写 device 配置 = upsert 该 device prefix 的 row（独立行）

关键反约束：
- ❌ 不实现 longest-prefix-match 解析（spatial_config 自带的能力本期不用）
- ❌ device 创建时**不要**自动派生快照写入（懒派生 — 只有用户实际改了 device 配置才写行）
- ❌ tenant 改动**不**级联到已存在的 device 行（明确隔离）

实施进度：
- ✅ Phase 1a tenant 层：postgres_alarm_cloud.go 改 spatial_config（commit 待）
- ⏸ Phase 1b device 层：等 1a 验证后做，涉及 radar-monitor-settings.vue / sleepace-monitor-settings.vue + device_monitor_settings_service v2 重写
