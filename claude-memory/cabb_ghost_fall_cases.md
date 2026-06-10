---
name: CABB radar ghost+fall fixture
description: 4D8710D5CABB 24h 6 份 fixture (3 fall + 3 ghost-pure)，证据揭示 health-tick→commit pending 假报模式
type: project
originSessionId: aa45449c-a767-466b-9f73-a31dd0bef48e
---
设备 `4D8710D5CABB` (Radar_CABB, Denver-LakeW-C2601, room=`0af07f65-b4f3-465b-bbcf-21d2abc7f7a6`) 在 2026-05-03 ~ 05-04 出现 3 起 Fall + 多段 ghost track，已导出 6 份 fixture 到 `doc/cases/cabb-*`，总览 `doc/cases/cabb-2026-05-overview.md`。

**Why:** 这是首次完整记录到 **qinglan 10-min health_check tick 触发 AI engine commit pending lost-fall** 的因果链——B 事件 tick 后 4min commit、C 事件 tick 后 5s commit + spatial_jump=true。AngleException long-time onset 说明 qinglan 频繁重启（user 确认主动）。同时 CABB 的 z 通道劣化导致真人躺下后 frozen-at-z=0 ghost (A 事件根因)。

**How to apply:** 做 fall verifier / ghost detector 改进时拿 `cabb-fall-A-frozen-0016/` 验证 frozen-at-z=0 prune 规则；拿 `cabb-fall-B/C` 验证 "OfflineRecover 后 N 秒禁止 commit pending" 规则；拿 `cabb-ghost-pair-1307/` 等三份 ghost-only 验证双 track 距离+z=0 ghost-pair 检测。CABB 非 D523 firmware (firmware_version=`2.3-Jun 25 2025 11:33:44`)，`number_people=0 ExitRoom 兜底` 不适用，调试时不要假设它能拦下 lost-fall。
