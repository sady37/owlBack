# P3 反馈日志（HSMM 驻留时长 / dwell）—— 施工方 ↔ 审核委员会

> 本文件专用于 P3 阶段交流（与主 `feedback.md` 分开，用户 2026-06-09 指定）。倒序，最新在上。
> 协作协议同主线：施工方提案 → 委员会裁 → 裁后 shadow-first 建（R0 log-only，不碰生产 gate）；裁前不建 / 需新岔口列选项不擅决 / 直接 main / cd wisefido-sensor 跑 go / 放行 bar = build/vet/belief 绿 + 9 红 0 新增。

---

## 审查记录（倒序）

### [2026-06-09 00:40 MDT] 委员会执行 P3 首增量（收敛施工方预审 `3e3fc5c`）— 裁 A/B/C + P4.1 formalize(`belief/survival.go` 单源)+ P4.3 夜间短尾落地 `272a48e`

**第四次收敛**：委员会建 P3 时施工方独立推 `3e3fc5c`，**同建 `doc/feedback_p3.md`** + 同观察「P4.1 ObsDwellStill 软时长已建，gate_vs_dbn『P3 未开工』陈旧」+ 列 A/B/C 三岔口。委员会已建 = 执行该 scope。

**✅ 裁施工方 A/B/C 三岔口**（亲查代码）：
- **P3-A（已建）= 准，且 formalize 落地**：ObsDwellStill 软时长确已建（`1+(d/scale)²`）。委员会**抽成 `belief/survival.go` 单源模块**（#1.3）：`dwellTailFor(zone)` per-zone 尾 + `fallLRFromDwell(dwellSec,tolMult,zone,night)`，likelihood 改调（等价复现日间，零回归）。gate_vs_dbn『P3 未开工』标注**应更新为『软时长已建』**（施工方后续改）。
- **P3-B（残留 dwell 区）= 无 fall-relevant 缺口（亲查）**：`geomFromArea`（belief_adapter:84）`AreaDeny→GeomOpenFloor` → **DenyZone 已走 open 8min 尾报**（非缺口）。bed/enter/unknown 故意不报（久卧正常，保守防裸 ramp FP）。impl_plan 的「deny 专属 5min 尾」需 AreaType 粒度（现用 Geom coarser）= **P4.2 zone 选档**事，非 gap，留 P4.2。
- **P3-C（真 HSMM，显式 per-state duration 进转移矩阵 A）= deferred v3**：roadmap 明列（room_belief_state_machine §9 留 v3），需用户单独立项。**不现在做**。

**★ 委员会顺手做 P4.3（impl_plan §3 P4.3，genuinely 未开工，非 freelancing）**：夜间短尾——`Observation.Night` + 夜间 `scale×dwellNightTailMult(0.7)`=更快 ramp（3am 静止 8min 比 3pm 更可疑，治久态误报核心）。night 由 beliefShadowTick `IsNightTime(nowMs,tm.timezone)` 算（P7.2 已有，复用）→ radarFrameAdapter（仅 shadow 调 R0-safe）→ ObsDwellStill.Night。镜像已落地 P7.2（context→cost 调）。

**测试** `survival_test`：等价复现（toilet/open 日间=旧 inline）/ zone 表（bed/enter/unknown→1.0）/ 夜短尾（night LR>day 单调封顶）/ tolerance 拉长尾。**放行 bar**：build/vet 净 + belief 绿 + roomengine **9 红 0 新增** + #1.6 净 + gofmt。**铁律守**：R0 shadow log-only / R1 / R5（dwell 久静抬 fall 正向；Z 不涉）/ R7 常量带来源待 P9 收紧尾形。

**P3 状态**：A 收口（formalize 落地）+ P4.3 夜尾建 ✅；B 无缺口（deny 已走 open，deny 专属尾留 P4.2）；C（真 HSMM）deferred 待用户立项。**P3「软时长」主体完成**，余 cutover（删 gate 硬阈）数据-blocked P9。

---

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
