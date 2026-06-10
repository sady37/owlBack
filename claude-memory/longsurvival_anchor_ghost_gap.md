---
name: LongSurvivalAnchor 吃掉 ghost 翻转
description: PR-5.3 LongSurvivalAnchor (5min) 守卫拦下 motion_symmetry/mirror 强信号的 ghost 翻转，导致旧 ghost track 不被改饱和度
type: project
originSessionId: 3d9a6ccb-8326-4fc0-96ba-8ec961837774
---
track_manager.go:1039 行 `!ts.LongSurvivalAnchored && !ts.StartupGrace` 守卫：track 存活 ≥300s 后被强制锚定 Real，之后即使 GhostPenalty 累积到 80（包括 motion_symmetry/mirror 强信号命中）也**不会翻 Ghost、不 emit verdict**。结果：旧 ghost track 在画布上一直保持原 firmware confidence=80 的满饱和度，前端无法降饱和度。

**Why:** 2026-05-06 排查 9923003AB17F (kitchen) Denver 11:26 时 P1 ghost 没改色 — Monitor 工具实测看到两次 motion_symmetry_hit 行为差异：track_id=2 (LongSurvivalAnchored=true) penalty=80 后无 ai_emit；track_id=3 (新生 pending) penalty=80 后 ai_emit + cardagg 51s 持续 apply 80→20。证实守卫有效但拦错了对象。

**How to apply:** 修复方向 — LongSurvivalAnchor 应只对**弱 ghost 信号**（30s 静止 / no_enter_pair / 低 score）有保护，**强信号**（motion_symmetric_with_real_track / mirror_image_of_real_track）应能 override 锚定翻 ghost。改 1039 行守卫为白名单：仅当 BirthReason 不是 motion_symmetric/mirror 时才生效。修前可参考 case fixture 导出方式记录 11:26 vs 11:41 这两个对照样本。
