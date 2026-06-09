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

**⟹ 「P3」真实范围有三种可能（逐条列优劣 + 施工方推荐，请委员会 + 用户拍，不擅决）**

**P3-A：宣告 HSMM 软时长已建 → P3 收口（余 cutover 待数据）**
- 内容：确认 ObsDwellStill 已覆盖所有 fall-relevant 久态区 → P3 per gate_vs_dbn 标「已完成」，更新陈旧标注；真正剩的是 cutover（删生产 gate 硬阈，**数据-blocked P9**）。
- **优**：诚实对齐现状（HSMM ramp 确已建）；零新代码、零回归风险；立即把 roadmap 标注纠正、避免重复造轮子；与委员会「P2-P7 buildout 完」口径一致。
- **劣**：可能**漏掉残留缺口**（若 ObsDwellStill 没覆盖某 fall-relevant 久态区，直接收口=把缺口当已完成，同审查58 SD double-count 的反面风险——没查 live 覆盖就下结论）。**故 A 不能空口宣告,须先做覆盖审计(=判别 A vs B 的前置)**。

**P3-B：内化残留 dwell 硬阈（若审计发现缺口）**
- 内容：ObsDwellStill 现只发 toilet(900s)+open(480s×tol)；gate-list 另有 DenyZone(Still.DenyZoneSec)/ RestZone/Walkway(Lost.*WaitSec) 等硬阈。审计判这些是**故意不报**（DenyZone=餐桌保 Walk 设计意图[[dining_table_walk_intentional]] / bed 久卧正常）还是**真 fall-relevant 缺口**，缺口则内化（同 P2 闸内化，shadow-first）。
- **优**：真治本（补上 ObsDwellStill 没覆盖的久态-fall）；与 P2 闸内化同模式、风险可控；为 cutover parity 补齐（gate 有阈、shadow 无→cutover 会漏报）。
- **劣**：多数残留很可能是**故意不报**（DenyZone/bed 设计意图），真缺口可能为零=做了审计发现无活干；且「往 ObsDwellStill 加区」有把**设计意图的不报区误变成报**的漏报反向风险（餐桌误报老人吃饭），须逐区核 fall-relevance 不能盲加。

**P3-C：真 HSMM 升级（时长建进转移矩阵 A，v3）**
- 内容：当前 ObsDwellStill 是**观测层生存 ramp**（survival function on observation）；真 HSMM = **转移矩阵 A 的显式 per-state duration 分布**（停留越久离开概率随分布变，非几何分布固定率）。room_belief_state_machine §5「v2 再上 HSMM」/§9「留 v3」。
- **优**：理论最完备（Murphy DBN 完全体）；治「久态」最彻底（duration 分布带方差，比单一 scale ramp 更贴真实停留统计）。
- **劣**：**工程重、roadmap 明列 deferred**；现有 ramp 已拿到 80% 收益（零方差→软化已治硬悬崖），A 层显式时长是边际精度提升；且**标定需真停留时长分布数据**（部分 blocked on 真机统计）；ROI 在 cutover/recall 都没过之前偏低（先把已建的验穿比再升级模型更优先）。

**★施工方推荐：先做 P3-A/B 覆盖审计（doc-only，不需裁即可做=纯验证非建），用审计结果收敛到 A 或 B；P3-C 暂不做（deferred，ROI 低于先验证已建的）。**
- 理由：A 不能空口收口（劣中已述），B 不能盲加（漏报风险），**两者的判别都靠同一次覆盖审计**——审计是 unblocked 的诚实第一步，做完事实自己说话：全覆盖→A 收口（纠标注）；有真 fall-relevant 缺口→B 内化。C 是更大的 v3，**不建议现在开**（先让已建的 HSMM ramp 走完 cutover/recall 验证，证明值得了再上 A 层显式时长）。

**待委员会/用户拍**：
1. **认不认可「先覆盖审计再定 A/B」**（施工方推荐）——若认可，我即跑审计（doc-only，落本文件）。
2. **P3-C（A 层显式时长 v3）现在做还是 deferred**——施工方推荐 deferred（ROI 低于先验证已建）。
3. 审计若发现 B 类真缺口，再逐区裁 fall-relevance（防把设计意图不报区误变报）。

**待处理问题清单（P3 open items）**：
- [ ] P3 范围裁定（A/B/C，待委员会+用户）
- [ ] ObsDwellStill 覆盖审计（施工方可即做，doc-only）：列全 fall-relevant 久态区 × 是否已发 ObsDwellStill × 未发的是否故意
- [ ] gate_vs_dbn_shadow.html「P3 未开工」陈旧标注待纠（A 收口时一并改）
- [ ] cutover（删 gate 硬阈，gate→DBN 判定权移交）—— 数据-blocked P9，非本阶段
- [ ] 裁前不建。
