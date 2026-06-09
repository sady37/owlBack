# P3 反馈日志（HSMM 驻留时长 / dwell）—— 施工方 ↔ 审核委员会

> 本文件专用于 P3 阶段交流（与主 `feedback.md` 分开，用户 2026-06-09 指定）。倒序，最新在上。
> 协作协议同主线：施工方提案 → 委员会裁 → 裁后 shadow-first 建（R0 log-only，不碰生产 gate）；裁前不建 / 需新岔口列选项不擅决 / 直接 main / cd wisefido-sensor 跑 go / 放行 bar = build/vet/belief 绿 + 9 红 0 新增。

---

## 审查记录（倒序）

### [2026-06-09] 施工方 → 委员会：**P2 检查通过 → 开 P3**（用户指令）→ ★P3 scope-clarifying 预审（doc-only，裁前不建）—— HSMM 软时长**已在 shadow 建好**，请裁 P3 真实范围

**P2 检查通过**：G-1 id-swap（`TObsLogicAlive`，施工方 009d794）+ G-2 空房账（`RoomLedgerEmpty`，委员会 c158023）+ G-3 距离闸（`lostFarFromRadar`，委员会 09039c4）三治本闸全内化进 DBN；build/vet/belief 绿 + roomengine 9 红 0 新增。**P2 真收尾**。

**★ P3 撞口径问题（先澄清再动手，同 P2）**：
- **gate_vs_dbn_shadow.html**：「P3 · HSMM 驻留时长 **未开工**」。
- **belief_dbn_impl_plan.md**：「**P3 = realness R_i**（已建：shadowTrackGhostness/realnessStep/poseZLock）/ **P4 = dwell HSMM**（散落硬阈→−ln S_vol ramp）」。
- **委员会 P7.5**：「DAG P2→{P3,P4}→**P7✅**→P9，shadow 决策层 P2-P7 buildout 完」→ 含义 P3/P4 已 buildout。

**亲查代码——HSMM 软时长(=用户口中的"P3")其实已建**：`ObsDwellStill`（belief/likelihood.go:30）= **生存函数 ramp `fallLR=1+(d/scale)²`（封顶 dwellFallCap=2.5）平滑取代硬悬崖**，按 geom 分 scale：
- `GeomInToilet`：scale=`dwellScaleToiletSec`=900s（对齐生产 ToiletShowerSec 15min）。
- `GeomOpenFloor`：scale=`dwellScaleOpenSec`=480s × ToleranceFactor（前置 Z_cell tolerance gate，被容忍久站的 cell 尾拉长）。
- `bed/enter/unknown`：lk(nil) 不报（久驻正常）。

= 这正是「现行 10min/2h/8min 硬阈值 = HSMM 零方差退化版 → 软化」的落地（impl_plan P4.1/P4.4 裁决⑮/⑱）。**gate_vs_dbn「P3 未开工」是陈旧标注**。

**⟹ 「P3」真实范围有三种可能,请委员会 + 用户拍(不擅决)**:
- **P3-A（已建，宣告收口）**：HSMM 软时长 = ObsDwellStill 已覆盖 fall-relevant 久态区（toilet/open）。则 P3 per gate_vs_dbn = **已完成**，余下唯 cutover（删 gate-list 硬阈，**数据-blocked P9**）。施工方 lean：**先确认这个**——若是，P3 无新建，本预审即收口，更新 gate_vs_dbn 标注。
- **P3-B（残留 dwell 区内化）**：ObsDwellStill 只发 toilet+open；gate-list 另有 **DenyZone**（Still.DenyZoneSec）/ rest/bed 等硬阈。待核：这些是**故意不报**（DenyZone=餐桌保 Walk 设计意图 / bed 久卧正常）还是**真覆盖缺口**。若有 fall-relevant 区的硬阈未进 ObsDwellStill → 内化（同 P2 闸内化）。
- **P3-C（真 HSMM 升级，v3）**：当前 ObsDwellStill 是**观测层生存 ramp**（survival function on observation），非**转移矩阵 A 的显式 per-state duration 分布**（Murphy DBN 完全体 HSMM）。room_belief_state_machine §5「v2 再上 HSMM」+ §9「HSMM 留 v3」=把时长建进 A（停留越久离开概率随分布变，非几何分布固定率）。这是**建模升级**，工程重，roadmap 明列 deferred。

**施工方倾向**：先走 **P3-A 确认**（亲查 ObsDwellStill 是否已覆盖所有 fall-relevant 久态区）→ 若全覆盖则 P3 收口（更新陈旧标注，余 cutover 待数据）；若 P3-B 发现缺口则内化；P3-C（A 层显式时长）是更大的 v3，需用户单独拍是否现在做。

**待委员会/用户**：① 裁「P3」指 A/B/C 哪个；② 若 A/B，施工方做覆盖审计（doc-only 先行）；若 C，需单独立项设计。**裁前不建**。
