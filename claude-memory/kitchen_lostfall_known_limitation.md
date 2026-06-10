---
name: Kitchen lost-fall 已知误报
description: Kitchen long-stand cooking → lost-fall 误报；elder care 场景不适用（老人不下厨）
type: project
originSessionId: 08872106-789f-48c2-8a3f-bcb3bd7092ef
---
Kitchen 长时间站立做饭场景会触发 lost-fall 误报：
- 人站在 counter/stove 前 15-60min 不挪动，雷达失锁（pose=Stand 久）
- 5min 等待期无新 track + 无 ExitRoom → 报 lost-fall
- 实测：D523 bookroom 3 天 47 次（boundary 修复前），Kitchen 3 天 11 次

**Why:** 老人 elder care 场景下厨房不是主要监控区（家属/护工代劳，老人进厨房通常短暂）。该误报仅在"老人独居 + 自己做饭 + 厨房有 radar"才显著。

**How to apply:** 不需要算法层修复。部署时：
- 老人独居 + 厨房 radar：layout 显式标 Counter/Stove 为 Furniture (AreaDeny)
- 或关闭厨房 still-fall stay-alarm
- 养老机构多人厨房：不部署 radar，集中在卧室/卫生间/客厅

详见 owlBack/doc/AI_fall_detect.md 第 16.1 节。
