---
name: layout_drift_wall_expansion_policy
description: 2026-06-04 观察+策略：layout 画图精度误差致真人 track 漂到墙外(CABB P0)；若持续穿墙→外扩 wall+radar border 含进漂移点，上限 30cm，绝不超雷达最大边界
metadata: 
  node_type: memory
  type: project
  originSessionId: 1098084b-61f0-4d91-b068-162fde7f2974
---

2026-06-04 用户从 CABB(4D8710D5CABB)实时画面观察 + 拍板的 layout 校正策略。

**观察**：layout 画图本身有**精度/误差**——CABB 画面里真人 track **P0(track_0)跑到墙(wall 多边形)外侧**(门 Enter 那一侧右边)。这不是真人走出房间，是**墙多边形画小了 / radar border 画得不够准**导致真人贴墙活动时 track 落在画出来的墙外。

**策略(用户定)**：
- **先记录、观察**，不立即改 layout。
- **若后续 P0 持续在 wall 穿越**(反复落墙外)→ 判定为画图误差 → **修正 layout**：把 **wall 多边形 + radar border 向外扩**，把漂移的 track 点包含进来。
- **硬上限**：外扩**最大 30cm**。
- **铁律**：外扩**绝不能超过雷达的最大边界值**(FOV/射程物理极限)——雷达本就看不到最大边界外，墙扩过去无意义且错误。

**判别**：持续穿墙(同一真人 track 反复落墙外固定区) = 画图误差，可小幅外扩；偶发 / 真走出房间 = 不扩。

**与静止反射体自学习的关联**([[belief_state_rule_engine_reframe]] 里 P2 后的 static_reflector.go Phase A)：静止反射体检测用"**近墙 ≤40cm**"作学习判据；wall 画不准会给"近墙"分类加噪声。两件事都指向同一底盘问题=**layout wall 多边形精度**。static_reflector commit 5aa0267。

**落地状态**：仅记录策略 + 观察，**未实现穿墙检测器**(需要它来攒"持续穿越"证据时再加 log-only 计数；用户要的是先记下)。layout 是人工真相，校正经 wisefido-data API/FE，不在 sensor 擅改。关联 [[radar_layout_device_invariant]](InstallMod/Height/Boundary 必 ==firmware 不漂移)、[[radar_geometry_point_invariant]]。
