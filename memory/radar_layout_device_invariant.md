---
name: Radar Layout/Device 不变量
description: Radar 的 InstallMod / Height / Boundary 三项，layout (sessionStorage + config_versions) 必须永远 == device firmware；不允许漂移
type: feedback
originSessionId: c2d0b96a-ef02-4ee6-9673-3a25811aadfc
---
Radar 的 **InstallMod / Height / Boundary** 三项配置，layout 侧（sessionStorage + 后端 config_versions）必须永远等于 device firmware 真值。areas（区域）不强制（语义复杂，写失败时不 cover）。

**Why**：历史 LaySave / IoTSave 双轨架构允许两侧独立写入——用户在 UI 改 install_model=2 后只点 LaySave，config_versions 已是 2，但 firmware 仍是 1，导致 angle alarm 持续异常但 UI 显示已修。这种漂移用户看不见、health_check 看 firmware 真值、UI 看 layout 视图，两边永远对不上。2026-05-04 调试 Radar_25A859B8333B 的 angle alarm 时确认是这个根因。

**How to apply**：

1. **唯一入口**：所有 InstallMod/Height/Boundary 修改必须经 `unifiedSave`（Toolbar.vue Save 按钮），不要给用户保留只写 layout 不写设备 / 只写设备不写 layout 的路径。
2. **写入成功后必须 Query + 重存 layout**：firmware 可能 silent clamp（如 boundary 截到设备允许范围），Query 拿设备真值再 saveCanvas 一次，把真值落到 sessionStorage / 后端，layout 才与设备同步。
3. **写入失败的 cover 红线**：非区域类（installModel/installHeight/rectangle）任何一项写失败 → Query 设备 → 用真值覆盖 layout 对应字段 → 重存 layout。区域类失败不 cover（多 area 联动语义复杂，机械覆盖会丢用户编辑）。
4. **首次绑定**：baseline.queriedAt 缺失 → 视为 first-bind，UI 当权威全量推到设备（不 auto-Query），完成后走 Step 2 维持不变量。
5. **新增 radar 配置字段时检查这个不变量是否需要扩展**：如果该字段在 firmware 有真源，加入 cover 列表；纯 layout 视觉字段（rotation/canvas 标注等）不强制。

**实现**：[owlFront/src/components/Radar/Toolbar.vue](owlFront/src/components/Radar/Toolbar.vue) `unifiedSave` (~line 2416) + `buildFullWriteCommands`。
