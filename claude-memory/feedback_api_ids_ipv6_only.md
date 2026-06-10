---
name: feedback-api-ids-ipv6-only
description: server 内部彻底 drop DeviceID/device_id；DeviceAddr (INET /128) 是唯一 device 寻址；外部厂家 ID 仅在 boundary adapter 翻译，不进数据契约
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 3ceef7bd-08fc-42c7-b60b-a4d85fe2134e
---

**2026-05-25 拍板（推翻 5-20 dual-ID 方案）**：server 全栈 drop DeviceID / device_id Go identifier 和 JSON 字段；所有 device 寻址用 DeviceAddr (INET /128 canonical text)。

**Why**：5-20 试过 dual ID（`DeviceID string` UUID 对外 + `DeviceIPv6 string` 内部）。同一周这条路径暴露 3 处 silent bug 全来自命名 drift：
1. `alarm_event_handler.go` map key 写 `"device_id"` 但 FE 期望 `device_addr` → 列表 RePlay 按钮不显示
2. URL query `scope_device_id` vs FE 发 `scope_device_addr` → 单设备精确 scope 失效
3. cards.devices JSONB key `device_id` vs owl-common 写入 `device_addr` → resident 权限路径 0 命中走 fallback

dual ID 表面积 = 漂移源头数；删掉一份命名直接消除整类 drift。

**How to apply**：
- ❌ Go struct 不写 `DeviceID string`；一律 `DeviceAddr string`
- ❌ JSON 字段不写 `"device_id"`；一律 `"device_addr"`
- ❌ URL query 不收 `device_id` / `scope_device_id` / `device_ids`；用 `device_addr` / `scope_device_addr` / `device_addrs`
- ❌ JSONB key 不写 `device_id`；一律 `device_addr`
- ❌ DB column 已是 `device_addr INET NOT NULL`；旧的 `device_id UUID` 列退役不留快照
- ✅ 反向归属（device → unit/room/bed/branch/tenant）一律 IPv6 prefix-mask：`set_masklen(device_addr, 48/56/80/88/96)` 或 `INET <<=`；不走 JOIN 链 — 详 [[feedback_ipv6_prefix_set_membership]]
- ✅ 直接按 device 查：`WHERE device_addr = $X::INET`，跟以前一样直
- ✅ 外部厂家 ID（sleepace logMAC、第三方 SDK userId）= **device_uid** (MAC 类)，仅在 vendor adapter (sleepace_gateway_client / OEM SDK) 内使用 + boundary 翻译；**不**作为 server↔FE / cross-service 数据契约字段
- ✅ DB device_factory_meta 的 dfm.device_id UUID legacy 主键存在 = 历史，不再向上扩散到业务层

**关联**：
- [[device_ipv6_single_trip]] — 一次性切换的部署纪律（本 memory 是终态契约）
- [[feedback_ipv6_prefix_set_membership]] — 反向归属用 prefix-mask 不用 JOIN
- [[sleepace_userid_is_device_uid]] — sleepace userId == device_uid (MAC) 不是 device_id；vendor adapter 翻译
- [[unbound_device_card_id_fallback]] — device 充 card_id 兜底的 hack 一并清理
