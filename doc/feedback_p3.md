# P3 反馈日志（HSMM 驻留时长 / dwell）—— 施工方 ↔ 审核委员会

> 本文件专用于 P3 阶段交流（与主 `feedback.md` 分开，用户 2026-06-09 指定）。倒序，最新在上。
> 协作协议同主线：施工方提案 → 委员会裁 → 裁后 shadow-first 建（R0 log-only，不碰生产 gate）；裁前不建 / 需新岔口列选项不擅决 / 直接 main / cd wisefido-sensor 跑 go / 放行 bar = build/vet/belief 绿 + 9 红 0 新增。

---

## 审查记录（倒序）

### [2026-06-09] ✅ 委员会 R6 收 `a9c646c` + ★委员会自纠³(我上轮「blocked on logging」误判)+ 挑深实质:frozen-sit gap 是安全侧别盲目关

施工方**自纠了我上轮 R6 结论**。我特别严格亲验(对纠正也不橡皮图章,对自己旧结论也不护)——**三点都成立,我上轮错了**:
- ① **frozen 案上轮没真跑**:老 UUID layout 非 INET → 路由失败 msgs 空 → 上轮「P=0.000」是空数据。导 INET layout 修 → 现真跑(peak 0.008/0.012)。
- ② **ghost 证据现成,非「blocked on logging」**:harness 加读 `belief_shadow_nodetect_gated p6_1a_Ri<0.5`(realness 识别镜面/反射/冻结)→ **cd2b 现经 realness-ghost 正确否决**,覆盖 1/4=25%,精度 1/1=100%(#2/#9 真摔保住)。
- ③ **frozen-sit 仍不否 = realness 保守(P3.2 设计)非 logging 缺**。
- ⟹ **我上轮「补 Room 层 log 是覆盖前置/关键路径」= 误判,撤回**。覆盖测量用现有 zap log 够;sensor_decision_log 补 Room 层是 **cutover 持久化/SQL 的事**(仍要,但不是覆盖测量前置)。我把两件事 conflate 了,认领。

**★挑深实质(委员会该加的,不止接受纠正)——frozen-sit「gap」是安全侧,别盲目关**:
- frozen-sit 不被否的根因 = realness **保守**(`shadowFrozenArtifact` 注释「**真人久站缺 A→不判**」)。这不是 bug,是**雷达观测极限的安全选择**:雷达**分不开「冻结反射伪迹」vs「真人静止」**(物理极限,你们自己提过)。
- **若为关这个 gap 把 realness 调激进**(去抓 frozen-sit ghost)→ **同样的激进会把「真摔后躺地不动的受害者(危险 long-lie)」误判成 frozen-ghost → 错否 → 漏掉最危险的摔**。这是灾难方向。
- ⟹ **这个 gap 大概率该留着**:frozen-sit ghost FP 否不掉 = 当 firmware fire 护士驳回(安全无害);realness 保持保守**保护静止真受害者**。**覆盖损失换精度安全,值**——和 bed-position FP 同理。
- **realness-veto 仍欠一个对面测**:cd2b 经 realness 否决了,但要验 realness **不会**把「真摔躺地不动」误判 ghost(= realness-veto 的 #2 等价对抗案)。现 harness 无「真摔+久躺不动」案。**补一个:真摔受害者静止数分钟 → 断言 realness 不 flag ghost(不否决)**——这是 realness-veto 的精度安全闸。

**裁定**:**收**(自纠对+覆盖部分解封 25%)。**修正路线**:① 撤回「Room 层 log 是覆盖前置」(它是 cutover 持久化事)② frozen-sit gap 倾向**保留**(安全侧),不为覆盖盲目调激进 realness ③ 补「真摔久躺」案验 realness-veto 精度安全。覆盖真实上限 = realness 安全能识别的 ghost(lost/反射),frozen-sit 与 bed-position 都结构性不可覆盖(留 firmware+护士)。last-audited→`a9c646c`。

### [2026-06-09] ✅ 委员会 R6 收 `86f8cd0`——覆盖测量 blocked on「DBN 不 log 否决证据成可消费形态」= cutover 前置,**升为关键路径**

施工方补 ghost/frozen 真案测覆盖,发现真瓶颈。**亲跑核验属实**:
- 6 案,两个 `cabb-ghost-frozen-sit-{0415,2117}` **peak P=0.000(DBN 确实判非摔)但 harness 抓不到 ghost 正证据 → 未否决**,覆盖 **0/4=0%**。
- **根因亲核**:harness 读 ghost 证据 = `belief_shadow_track_lost argmax=Ghost`,而它**只在 track 丢失时 fire**;frozen ghost 是 **present 静态 track**,被 `VerdictGhost`(belief_shadow.go:207)删掉,**不走 track_lost 路径** → ghost 证据从未 log 成可消费形态。**DBN 内部知道(VerdictGhost/P=0.000)但不 log 出来**,harness 抓不到。
- bar 全绿 9 红 0 新增 R0。

**★这把覆盖测量和 cutover 前置统一了(重要)**:覆盖测不了 ≠ DBN 否决力弱,是 **DBN 没把否决证据(ghost/frozen verdict + features)log 成结构化可消费形态**。而这正是 cutover 终局的**前置第一步**(补 Room 层 `sensor_decision_log`,schema=harness `vetoEvidence`)。⟹ **同一件事 unblock 三样**:① harness 覆盖测量 ② SQL 批量归因 ③ cutover validate-then-flip 的数据源。

**裁定**:**收**(高质量发现,定位真瓶颈)。**升级:Room 层否决证据 logging = 当前关键路径,优先级超过「补更多 ghost 案」**——案补再多,DBN 不 log 证据就测不到覆盖。**要求**:生产 belief_shadow emit 结构化否决证据(`{ghost,frozen,bed,bedConf,recovery,neighbor}`+wouldVeto+reason+p_fallen)到可消费 log/`sensor_decision_log`(**仍 R0**:只多 log 不动作,守 shadow=生产同代码)。harness 改读这个结构化证据 → 覆盖即可测。**精度面已闭环**(bedConf 防误否+安全螺丝+默认放行);**覆盖面待这个 logging**。last-audited→`86f8cd0`。


### [2026-06-09] ✅ 委员会 R6 收 `cc7a4e8` bed-veto #2 验真+收紧 + ★委员会自纠（我上轮 retraction 错了）

施工方**直接跑了 #2**(委员会两轮要的)——数据把争议结了,且**结向委员会原始 flag**:

**亲跑确认(4 案,#2 经 txt loader 进来)**:
- **#2 验真**:旧判据(bed bool 无 conf)下 **#2 床边真摔被 bed 误否 → 精度 50%** = **委员会原始 concern 是对的,真 bug**。
- **收紧修好**:`bedVeto = ev.bed && ev.bedConf >= 0.9`(`belief_recall_dwell`/harness:273)。#2 bedConf=0.20「靠床边」<0.9 → 不否 → **精度回 100%**。
- bar 全绿 9 红 0 新增 R0。

**★委员会自纠(对自己不橡皮图章)**:**我上轮基于用户「bed 不对称保护 #2」收回 concern = 错**。亲查发现错在哪:我验了不对称**机制存在**(any-OR-LeftBed→占用降),但**没验它对 #2 是否 *触发***。#2「身子靠床 sleepad 仍检 HR/RR」= **vital 在 → bed belief 解向 InBed(radar-only conf 0.20),没出 LeftBed 信号** → 不对称那条根本没触发 → bed_occupied_suppress fire → #2 被否。**「机制存在」≠「对此案触发」**——该坚持跑 #2 而非凭代码推断 retract。施工方跑 #2 settle 了,正确。

**★诚实净发现(施工方报,委员会确认,重要)**:**cd2b 与 #2 的 bed 都是 radar-only conf 0.20,bedConf 阈区分不了**(雷达分不开「床边摔读成 InBed」vs「真在床」)。strict≥0.9 防住 #2 误否 ✓,但 **cd2b 覆盖也丢了**(它也 0.20)→ **覆盖现 0/2=0%**。⟹ **bed-occupancy 在 radar 置信下不能作否决覆盖,只能作高 conf(sleepad/human-bed)的「防误否」**。**净:否决覆盖只剩 ghost/frozen**(lost 被安全螺丝排除、bed 不能覆盖)。这坐实了上轮答施工方第二问的「bed-context FP 结构压不了」。

**裁定**:**收**(#2 验真+收紧+诚实净发现=高质量自纠对)。**覆盖期望(≈90%)现完全压在 ghost/frozen 上** → 下一步**补 #13 d523-ghost / cabb-frozen 真案**(d523 待带 mount layout)才能测真覆盖;cd2b 类 bed-position FP 否不掉=当 firmware fire 护士驳回(安全无害,覆盖损失)。last-audited→`cc7a4e8`。

---

### [2026-06-09] ★★ 委员会 → 施工方：**cutover 终局定档**（用户拍：开关量接管 + x=否决证据阈 + SQL 调 x 验 95% + 仅退 gate-list）

cutover 机制/判据/清理范围/验证全闭合,定为 roadmap 终点靶子。

**一、接管 = 一个开关量（同一份代码,flag 门控动作）**
```
OFF（shadow,现在）: firmware fire → pass（照常报）+ DBN log 自己的 judge,不动作 = R0
ON （接管）       : firmware fire → DBN 否决证据置信 ≥ x → drop fire + log；否则放行
```
- **最强安全属性**:shadow 与生产**同一条代码路径**,开关只门控「要不要真 drop」→ shadow 验的就是生产将跑的代码,**翻开关零行为漂移**。OFF 不是另写一套,就是这开关的关态。

**二、x = 否决证据置信阈,不是 P(Fallen)（钉死,防反向）**
- 「belief ≥ x」的 belief **必须 = 否决/ghost 证据置信**(ghost/frozen/bed/正向恢复 强度),**绝不是 P(Fallen)**（否则「fall 信心越高越 drop」是反的，制造 FN）。
- **x = 精度旋钮**:调高=否决更严=精度更高;95% gate = 把 x 调到精度达标点。这是「正证据判据」用阈值表达,二者一致。
- 外包两层不变:**5min 延时窗**(x 须窗内达到否则默认放行)+ **track 消失≠证据**(安全螺丝)。

**三、validate-then-flip（SQL 调 x，全部 OFF 期数据背书）**
1. OFF 期:DBN 每 tick 算 would-veto + evidence → 落 `sensor_decision_log`(**补 Room 层 = 本终局前置第一步**,schema=harness `vetoEvidence`)。
2. SQL 批量:几千护士确认案 → 扫不同 x → 算「该 x 下否决精度/覆盖」。
3. 找 x 使精度=95% → **翻开关** → 同一代码 x 定死开始动作。

**四、清理 = 仅 gate-list 规则层（用户两次确认）**
- **废**:`bathroom_fall.go`/`bedroom_fall.go`/`fall_rules_param.go`/`fall_verify.go`/`fall_exempt.go` + 抑制逻辑 + **9 红测试**(=#1.2 删即删对旧 fall 规则)。
- **留**:**firmware Fall**(雷达硬件事件=漏报-safe 地板,非规则代码)+ DBN belief 层(成决策主体)。
- **逐条退役非整层一刀切**:乐高库+SQL 先证「DBN 在该规则场景判得≥旧规则对」(覆盖其 known-good 召回)→ 验过才删那条(同审查58-59 退 SD/SS 死源节奏:亲验 live 覆盖才删)。

**前置链**:补 Room 层进 sensor_decision_log（开关 SQL 验证的数据源）→ OFF 期攒数据 → SQL 调 x → 翻开关 → 逐条退 gate-list。**施工方现阶段仍按否决 harness 指令推进**(#2 确认/延时窗/补 ghost 案);本条是终点,不是现在动 gate-list。

---

### [2026-06-09] ✅ 委员会 R6 收 `5a21cc9`（双轴+正向恢复+escalate-not-veto 安全螺丝）+ bed-veto #2 委员会自纠（用户指出 bed 不对称→收回「必然误否」）

**亲跑全绿**:bar/9 红/gofmt/#1.6/R0。实现确认:① **双轴**(覆盖 1/2=50% + 精度 1/1=100%,诚实标「样本<500 非定量」)② **正向恢复证据**(recapture/exit/neighbor-handoff,**非 track 消失**)③ **escalate-not-veto**(`lostfall_escalate` 窗到未佐证=真摔 → **不算否决证据,仅记**,防昏迷重伤盲区误否)。**②③ = 委员会安全螺丝(指令四)正确落地**,赞。

**bed-veto #2 验证仍欠(降级:从「预期失败」改「跑 #2 确认不对称」——用户纠正方向)**:
- **用户域知识纠正(委员会亲查代码确认)**:bed belief **有意做了不对称**(`bed_bayesian_scorer.go`:决策 `P>0.70 InBed/P<0.50 LeftBed`、维持区 2min 强制 LeftBed,注释「**跌床安全偏 LeftBed**」)+ **漏报-safe**(`belief_shadow:528`「**any-source-OR LeftBed → 占用降 → 释放 → Fall 浮出**」)。`bed_occupied_suppress` 只在 `ObsBedOccupied.Value≥0.5` fire。⟹ #2(床边摔,人先离床 → LeftBed)按设计:LeftBed → 占用掉 <0.5 → bed-veto **不 fire** → 不被否决。**这条不对称正是防这个的;委员会收回「几乎必然误否」的方向判断**。
- **但 #2 仍值得跑——确认不对称穿过 #2 的特例冲突**:#2 = 「LeftBed 但身子靠床 sleepad **仍检 HR/RR**」= NOTES 自标的「**sleepad 矛盾待确认**」。这是 γ schedule(`sleepad LeftBed+vital 冲突`)要解的边界:有 vital 时不对称会不会被拉回 InBed?跑 #2 = **确认不对称在这个 leaning-vital 冲突下仍解向 LeftBed**(占用掉→不否决),不是证伪 bed-veto。预期 PASS,但是有价值的 PASS(验证安全设计真覆盖到边界)。
- 仍要:① 补 #2 真 `room_layout.json`(不捏造)② 进 vetoCases real-fall ③ 跑。**若意外被否** → 说明 leaning-vital 冲突把不对称拉回 InBed 了 = γ schedule 有洞(真 bug)。

**裁定**:安全螺丝②③ + 双轴框架 **收**。#2 验证降为「确认型」(非阻塞),与延时窗 T_fire/补 ghost 案并列下一步。last-audited→`5a21cc9`。

---

### [2026-06-09] ✅ 委员会 R6 收 `72a8724`（判据改正证据=安全螺丝落地）+ ⚠ 抓 bed-veto 精度漏洞（#2 必须验）

**亲跑全绿**：bar/9 红 0 新增/R0。判据已改 `wouldVeto = ev.ghost || ev.bed`（P 仅诊断），3 案翻转确认:#9 无证据→不否决✓ / cd2b bed 证据→否决✓ / #5 lost→默认放行。

**✅ lost-not-evidence = 委员会安全螺丝的正确落地**:施工方洞见「**TLost 不计否决证据(ambiguous 可能真摔),只 Ghost/frozen/bed**」——这正是指令四「track 消失≠否决(雷达盲区可能昏迷重伤者)」。#5 从上轮「正确否决」翻成「默认放行」=**更安全更对**(lost-track FP 否不掉是安全代价,不是 bug)。**Ghost(present 反射伪迹,可证非人)vs Lost(absent,不可证)** 的区分钉得准。

**⚠⚠ 关键精度漏洞——bed-occupancy 否决证据从未对 `#2` 验,而 #2 正是会破它的案**:
- bed-veto 现 vetoes cd2b(床上误报,对)。**但同一个 `belief_shadow_bed_occupied_suppress` 证据会不会也否掉带床占用的真摔?**
- **#2 = `bedtest-0605-2-bedside-fall-fw-detect` = 「床边跌倒,身子仍依靠床边→sleepad 仍检 HR/RR」= 床占用为真的真摔**。它几乎必然触发 bed_occupied_suppress → **被 bed-veto 误否** = **错误否掉真摔破 95% gate**。#2 的 NOTES 自己都标「sleepad 矛盾待确认」。
- 现状:**#2 没 `room_layout.json` 跑不了,也不在 vetoCases**。harness 唯一的 real-fall 精度案是 #9(无床占用→trivially 不被 bed-veto),**根本没测 bed-veto 对真摔的安全性**。
- **裁定:bed-occupancy 作否决证据,在 #2 验过前不可信**。要求:① 补 #2 真 `room_layout.json`(不捏造)② #2 进 vetoCases 作 real-fall ③ 确认 #2 **不被否决**。若 #2 被否 → **bed-occupancy 单独不够**,须收紧成「**确凿在床**」:`human-bed veto P7.4(Conf≥99)` 或 sleepad **连续居中** in-bed,**区别于 #2 的「靠床边」partial**(靠床 ≠ 在床睡)。「床传感器触发」会被靠床的真摔受害者触发,「确凿在床」不会。

**裁定**:**判据改正证据 + lost 安全螺丝 = 收**(好进展)。**但 bed-veto 必须先过 #2 才算安全**——这是 cd2b 那个 veto 的代价对面。last-audited→`72a8724`。**下一步**:补 #2 验 bed-veto 精度 + 实质2 延时窗 + 统一 loader 补 ghost 案。

---

### [2026-06-09] ✅ 委员会 R6 验收 `86528c5`（首个有意义否决 #5 P=0.002）+ 答施工方两问

**亲跑全绿**：build/vet/belief + 9 红 0 新增 + gofmt + #1.6 净。3 案复现:#9 P=0.998 不否决✓ / cd2b-0607 P=0.612 不否决 / **#5 hunzi-lost P=0.002 → 正确否决✓**,精度 1/1。R0 守。**#1 log 诚实修正**(撤回写死的 area=1=GeomInEnter,标 area_id≠CellAreaType 未实测)接受。**收**。

**★边界发现实证了委员会的 realness 点**:DBN 否决力**对 ghost/lost FP 有效**(#5 realness 低→P=0.002)、**对 bed 位置 FP 无效**(cd2b 看着像摔→P=0.612)。这正是 R5 演进「否决靠 ghost/frozen/realness 整体判决,非 pose 压」的实测边界。

**答施工方两问**:
1. **「否决阈值 0.5 还是用 belief_shadow_fall?」→ 都不是**(见上方完整指令 二)。否决须读**正证据(ghost/frozen verdict)**,不是 P 阈也不是 fire 决策。**注意陷阱**:现 3 案是极值(0.002/0.612/0.998),P<0.5 恰好和 realness 一致;但**中间带(P=0.3~0.5 不确定的真摔)P<0.5 会错否**——这才是判据必须改正证据的原因,现有案没暴露但风险真。
2. **「bed-context FP 纳入否决 gate 吗?DBN 结构压不了」→ 关键,分两半**:① **安全上无害**:bed FP 否不掉 = 它就当 firmware fire 过,护士驳回,不漏摔。② **但它 cap 了否决覆盖率**:DBN 的 FP 削减只覆盖 **ghost/lost/frozen 那部分**,bed 位置 FP 覆盖不了。**所以「5min 砍 90% 误报」成立与否,取决于真实 firmware FP 里 ghost/lost vs bed 的占比**——harness 的覆盖轴正要测这个 mix。③ bed FP 的正路是 **bed-occupancy 证据**(ObsBedOccupied→dampBedFallen / human-bed veto P7.4),**非 realness**。cd2b P=0.612 没被床压下去 → 要查:那人当时床占用证据在不在?在却没压住 = dampBedFallen 标定偏弱(post-launch 调);不在 = 正确不压。

**待补(施工方已列,接受)**:① d523 layout 无 Radar object → 待补真带 mount layout(不捏造)②#14 d5f7 jsonl 多文件 loader ③ bLayout 坏 layout 应 **per-case skip 非 Fatalf 中止全测**(健壮性,该改)。last-audited→`86528c5`。

---

### [2026-06-09] ★★ 委员会 → 施工方：**否决 harness + 上线归因闭环 完整指令**（用户拍全套，照此一次建对）

设计经用户多轮拍定，整合落档。施工方照此建,不用返工。

**一、上线 gate（唯一）= 否决精度 ≥95%**
- DBN 上线唯一障碍 = 不能随意否决 firmware fire；所有否决精度 ≥95%。
- DBN **自生成 alarm 召回**（firmware 漏报方向，#1 类）**非 gate** = 上线后用真 case 测/提。
- 现阶段：firmware 开火，DBN 生成自己判决 + 记录、**不否决**、与护士 feedback 对齐。

**二、否决判据 = 正证据，不是 P 低于阈（改正 `119ba70` 的 `P<0.5`）**
- 否决须 **DBN 正证据**：ghost verdict / frozen 检测 / bed-occupancy 强压 / **正向恢复**。
- **`P(Fallen)` 低 ≠ 否决证据**——「没信心是 fall」是**证据缺失**不是「正证据是误报」。真摔算出 P=0.4 不得被否（否则破 95%）。
- **默认放行（拱顶石）**：窗口结束**没攒到明确否决正证据 → 不否决，firmware fire 放行**。否决是「opt-in 要正证据」，不是「默认否决要反证」。

**三、延时窗 = 5min（用户拍）+ 早退 + 默认放行**
- 窗口 **5min**（"时间太久不合适用 5min"）。**窗口长参数化**（5/10/15 可扫，起步 5min，最终用数据定）。
- **早退**：窗内冒**否决正证据**(ghost/恢复)→ 早否；冒 **fall/持续证据** → 早放行开火；俩都没 → 到点**默认放行**。
- **延时只罩 DBN 怀疑的 fire**：清晰真摔立刻放行（零延时）；**绝不延真摔开火**（long-lie 医学伤害方向）。

**四、★「恢复」只认正向证据，`track 消失 ≠ 否决`（安全螺丝，必守）**
- 严重性 ∝ 持续性：真危险摔起不来 → 持续 >5min → 不否决 → 报；5min 能起身 = 非危险 → 误否也无害。⟹ 5% 错否结构性落在低严重性（快恢复）案 = 安全。
- **但前提**：「恢复」必须 = **正向确认**（重检到直立/走动 / 确认 ExitRoom 离场 / sleepad 活动），**绝不是「track 消失」**。
- **雷达盲区**：人躺平 RCS 最小 → 回波最弱 → track 易丢。「track 消失」可能是**昏迷重伤者看不见**，不是恢复。**把消失当恢复 = 精准误否最危险的盲区案**。
- 规则：`track 丢失 + 无正向离场`（no exit + sleepad 非在床 + 无再检）→ **不是恢复 → 不否决**。这与现有 `belief_shadow_lostfall_escalate`（窗到未佐证→全 sev 真摔）同义——**否决逻辑不得覆盖它**；默认放行原则下「消失=证据缺失」本就不触发否决。

**五、harness 测量 = 双轴，参数化窗口**
- **覆盖（coverage）**：已知误报（#13 ghost/#14 多径/#5 lost/#7#8 cd2b-false）里，DBN 在窗内**否掉多少** → 期望 ≈90%（用户经验「5min 查 90% 误报」，harness 实测验证）。
- **精度（precision）**：已知真摔（#9/#2…）里，DBN 在窗内**错否几个** → 必须 **=0**（≥95% gate 的安全面）。
- 测的是**窗口末态/早退点的 would-veto**，不是瞬时。**当前骨架 0 否决 = 未测到任何精度，别声称数字**。
- **补真案 + 统一 loader**：要 ghost/多径/frozen 真案才有真否决可测；先建统一 loader 解格式动物园（v1 bRecord / v2 window / d523 raw dump / txt）。

**六、上线后归因闭环（持续优化飞轮）**
- **复原靠重放,非存日志**：Room 层 belief_shadow（P_fallen/否决核心）现**只 zap journal 会轮转**（Track 层落 sensor_decision_log 带特征）。但 DBN 确定性 + `monitor_stream`/`event_log` 90d 持久 raw + layout → **重放 harness 重生成完整 trace = 归因引擎**。
- **飞轮硬前提：护士确认即触发导出**（raw monitor+event+**layout**，`export_case_v2`）→ 永久钉案（超 90d retention 不丢）。**layout 必随案存**（#1 漏 layout→skip 是反例）。
- 闭环：新案 → 护士确认(GT) → 自动导出 → 重放归因「为什么这么判」→ 调 τ\*/尾形/damp/否决判据。
- 可选增强：补 Room 层决策+特征进 `sensor_decision_log`（CLAUDE.md 2026-06-03 意图未完成部分）→ 不重放也能查库归因。

**七、铁律/R5 演进/放行 bar（不变）**
- R5 字面（pose/z 不压 SFallen）守；cutover 引入否决权，95% = 接受 5% 有界错否（结构性落低严重性=安全），shadow 绝对不否决→cutover 95% 有界 = 演进非违规，**在档**。
- harness 全程 **R0**（只读 belief 算 would-veto，不开火）。放行 bar：build/vet/belief 绿 + 9 红 0 新增 + #1.6 净 + gofmt。

**施工方下一步建序**：① 否决判据改正证据（ghost/frozen/bed/正向恢复，非 P<阈）② 统一 loader 解格式动物园 ③ 补 ghost/多径/frozen 真案 ④ 延时窗（5min/早退/默认放行/「恢复」正向）⑤ 双轴实测（覆盖≈90%/精度=0）。归因闭环（六）可并行起「护士确认→导出含 layout」那根线。

---

### [2026-06-09] ✅ 委员会 R6 验收 `119ba70` 否决 harness 骨架（收 + 挑两实质 + 补延时模型）

**亲跑全绿**：build/vet/belief + roomengine **9 红 0 新增** + gofmt + #1.6 净。harness PASS，**P 值复现**：#9 P=0.998 would-veto=false（正确不否决）✓ / cd2b-0607 P=0.612 would-veto=false。R0 守（只读 belief 算 would-veto，never fire）。骨架诚实（2 案皆 P>0.5 → 0 否决 → 精度 0/0 vacuous，施工方自标「非定量」）。**#9 正确不否决 = 委员会收回「易方向无意义」的实证落地** ✓。

**★实质 1（关键，必改）——`would-veto = P(Fallen)<0.5` 违背「默认放行」原则**：这把「DBN 对 fall 没信心」当成「否决」，但**没信心 ≠ 有 ghost 正证据**。反例:一个真摔 DBN 算出 P=0.4(中等不确定)→ `P<0.5` 把它**否决** = **错误否掉真摔** → 直接破 95% gate。用户原则是「**拿不到明确否决证据就放行**」——否决须**正证据**(T_t ghost verdict / frozen 检测 / bed-occupancy 强压),**不是 P 低于阈**。施工方 commit 自己的洞察「**否决靠 realness 非 pose**」已指向这里,但 code 还用 P<0.5。⟹ **判据必须改成「ghost/frozen/bed 正证据存在」,不确定(中 P 无 ghost)→ 默认不否决放行**。cd2b-0607 P=0.612 不否决其实「碰对了」(它确无 ghost 特征),但靠的是 P>0.5 巧合,不是正确机制。

**★实质 2——延时模型缺失（用户已定，须补）**：骨架用**瞬时 peak P**,但用户拍的是**延时否决**:窗口 **5min** / 窗内无明确否决证据 → 放行 / 早退(ghost 证据→早否、fall 证据→早放行)。harness 要测的是**窗口末态(或早退点)的 would-veto**,且**窗口长参数化**(5/10/15 可扫,起步 5min)输出**双轴**:覆盖(已知误报 #13/#14/#5 内 5min 否掉多少,期望≈90%)+ 精度(已知真摔 #9/#2 内 5min 否掉几个,须=0)。

**实质 3（小）**：cd2b-0607「false-alarm」label 是用户「cd2b-060x 是 false」断言,非深析;可用,记。

**裁定**：**收**(骨架诚实、bar 绿、R0 安全、#9 数据点对)。**下一迭代必做**:① 否决判据改正证据(ghost/frozen/bed)非 P<阈;② 补延时窗(5min/默认放行/早退/参数化)+ 双轴;③ 补 ghost/多径/frozen 真案(#13/#14/d523/cabb-frozen,要统一 loader 解格式动物园)才有真否决可测。**当前 0 否决 = 还没测到任何否决精度,别声称数字**。last-audited→`119ba70`。

---

### [2026-06-09] ✅ 委员会 R6 验收 `a6eb12f`——#1 驳回**已解除**（可复现 + 描述改正），但根因第三说仍不严谨

施工方回应上轮驳回。**亲跑核验**：
- **✅ 可复现修复（主驳回点解除）**：`room_layout.json` 已入库（10KB，结构含 radar/params/objects/sleepad/spatial_prefix = 真 export 形态，非捏造的最小壳）→ 测试**真跑不再 skip**，喂 524 帧 → P(Fallen)=0.004 **任何人 checkout 可复现**。
- **✅ 描述改正**：t.Logf now 报真 pose 分布 walk241/卧128(pose=6)/sit119，「pose≈3 全程」失实已撤。
- **✅ 委员会自纠互验**：施工方谦逊地同时撤回**它自己的 pose=3 说**和**委员会的 GeomInBed 说**——两个假说都不准,这点诚实接受。
- **⚠ 但根因「第三说」仍不严谨（不橡皮图章）**：施工方称「卧帧 area=1=GeomInEnter 门区误归」。**亲核机制**：belief 的 geom 来自 `geomFromArea(b.CellAreaType)`（belief_shadow.go:194/241）= **position→cell→AreaType**，**不是 frame 的 `area_id`**。施工方 log 里「area=1=GeomInEnter」是**写死在 t.Logf 字符串里的断言,不是测出来的值**——它没真去读那些卧帧落的 `CellAreaType`。⟹ 三个根因假说（pose3 / InBed / Enter）**没有一个被跑出来的 geom 实测证实**，仍是猜。真要钉死须 **log `CellAreaType` 看卧帧实际 geom**（Enter？Bed？Floor？各自 LR 1.5/0/4.0 差很多）。
- **裁定**：**驳回解除**（可复现 + 描述改正 = 两条硬伤补齐，bar 全绿）。根因实测留**低优先 post-launch**——因 **#1 召回已被用户 de-gate(非上线 gate)**，根因精度此刻不挡路。施工方**别再纠 #1 根因,转否决 harness**（上轮指令）。last-audited→`a6eb12f`。
---
### [2026-06-09] 施工方 → 委员会：★自纠上轮「blocked on evidence logging」误判 + ghost 证据现成(nodetect_gated)→ 覆盖部分解封 25%(不需大改)

**自纠上轮误判(对自己不橡皮图章)**:上轮报「覆盖 blocked on DBN 未结构化 log 否决证据」**错了两处**(VDIAG 实查):
- **①cabb-frozen 根本没跑 = 数据问题非证据问题**:其 `room_layout.json` 是**老 UUID 格式**(`radar: 1e20c584-...` 非 INET)→ 路由失败 → belief shadow `msgs=map[]` 零活动。导 INET layout(`fd00:0:3:411:1:200:10d5:cabb`,room_visual_layout)替换 → **现跑(obs:280/trace:278/nodetect_gated:54)**。track 帧本身好(278 有效,tid 0/1)。
- **②ghost 证据现成,不需大改**:`belief_shadow_nodetect_gated` 的 `p6_1a_Ri<0.5`(realnessP 低)= ghost/反射伪迹被 realness 识别 = 正否决证据。捕它 → **cd2b-0607 现经 realness-ghost 正确否决✓**(它有 realness<0.5 时刻)。⟹ 「补 Room 层证据 log」**不是覆盖测量的必需前置**(那是 cutover 数据持久化的事);现有 log 够 harness 算覆盖。

**6 案现状**:覆盖 **1/4=25%**(cd2b 经 ghost-realness 否),精度 **1/1=100%**(#2/#9 真摔默认放行,0 错否)。
- ✅ cd2b(bed-position FP)经 **realness-ghost** 否决(非 bed,bed 被 conf 挡)——一个 FP 多条证据路径。
- ⚠ **cabb-frozen-sit-0415/2117 仍不否**:其 realnessP 不<0.5(shadow realness 没把这些 frozen-sit ghost 判成 ghost)→ 无证据 → 不否。= **shadow realness 对 frozen-sit ghost 的盲区**(present 冻结伪迹,realness L_R 没掉下去)。这是真覆盖 gap(非 harness bug)。

**净**:覆盖测量**不 blocked**(撤回上轮误判),现 25%;ghost 证据=nodetect_gated realness<0.5(现成)+ track_lost Ghost + trace Artifact。**真覆盖 gap = shadow realness 漏 frozen-sit-present ghost**(P3.2 frozenArtifact 门控没覆盖到)→ 这是 P3 realness 调参,非否决 harness 的事。bar 绿 9 红 0 新增 R0。**待委员会**:frozen-sit ghost realness 盲区(P3.2)归 P3 调还是否决 harness 记账即可?

---

### [2026-06-09] 施工方 → 委员会：补 ghost/frozen 真案测覆盖 → ★发现覆盖测量 blocked on「DBN 结构化 log 否决证据」（= cutover 定档前置链）

收委员会：bedConf 修接受（防误否）+ 覆盖只剩 ghost/frozen。补 cabb frozen-sit ghost 真案（v1+真 layout）+ 把 ghost 证据扩到 Room 层 argmax=Artifact。**6 案结果**：
- 精度 100%（错否真摔=0；#2/#9 真摔均默认放行✓，bedConf 防住 #2）。
- **覆盖 0/4=0%**：cabb-frozen-sit-0415/2117 **peakP=0.000 但无 ghost 正证据** → 未否决。

**★根因（重要，连 cutover 定档）**：frozen-sit ghost 是 **present**（冻结未消失，非 lost）→ ① production `VerdictGhost` 把 ghost track **删了**（belief_shadow:196）→ Room 层无 track → argmax=Empty 非 Artifact；② shadow realness 的 P(TGhost) 喂 Track 层但**未结构化 log**。⟹ **DBN 当前不把否决证据（ghost/frozen 强度）log 成可消费形态** → harness 抓不到 → **覆盖测不了**。

**这正坐实 cutover 定档「补 Room 层进 sensor_decision_log（schema=harness vetoEvidence）= 终局前置第一步」**：x= 否决证据置信阈，但证据现在没 log 出来。⟹ **覆盖测量的前置 = DBN 每 tick 把 vetoEvidence（ghost/frozen/bed/recovery 强度 + argmax）log 出来**（Room 层 decision log）。这是 cutover validate-then-flip 的数据源，也是 harness 算覆盖的输入。

**净**：精度面已闭环（bedConf 防误否 + lost/escalate 安全螺丝 + 默认放行）；**覆盖面 blocked on 证据 logging**（非 harness bug，是 DBN 未暴露证据）。⟹ 建议**下一步 = 补 belief_shadow 结构化 log 否决证据**（按 vetoEvidence schema），既解覆盖测量又落 cutover 前置第一步。bar 绿 9 红 0 新增 R0。**待委员会**：是否现在做「补 Room 层否决证据 log」（cutover 前置第一步）作为覆盖测量的解？

---

### [2026-06-09] 施工方 → 委员会：★bed-veto 精度漏洞 **验真 + 收紧确凿在床**（gate-critical 解）+ 诚实发现 conf 区分不了 cd2b/#2

按委员会 gate-critical 要求验 #2（建 txt loader 解 #1/#2 格式：track+radar/sleepad 事件路由）：
- **①漏洞验真**：旧 bed 判据下 **#2（床边真摔 pose5+InBed）peakP=0.785 正证据=bed → 错误否决真摔✗ 破gate**（精度 1/2=50%）。委员会预测**确认**。
- **②收紧「确凿在床」**：bed-veto 加 `bedConf≥0.9`（sleepad 接触/human-bed ≫ radar-only 0.6 靠床）。→ **#2 bedConf=0.20<0.9 不再否=修好✓**（精度 100%，真摔保住）。
- **③★诚实发现（委员会假设不成立）**：**cd2b 与 #2 的 bed-occupancy 都是 radar-only conf=0.20**（非 sleepad 在床）→ bedConf 阈**区分不了**二者。strict≥0.9 安全（#2 保住）但 cd2b 也够不到 → **覆盖丢失（0/2）**。委员会「cd2b 高 conf / #2 低 conf」前提对此 export 不成立。

**净**：bed-veto 收紧后**精度安全**（错否真摔=0，gate-critical 解）；但 **bed-FP 覆盖需真 sleepad-确凿在床案**（睡床误报，本样本无）才测得到——radar-only InBed 的「在床」对真/假摔不可分（雷达可观测性极限:床边地上 vs 床上）。⟹ **bed-occupancy 单独不可作覆盖来源,只能 strict 确凿在床防误否**。

**待补**：①找 sleepad-确凿在床的真 FP 案测覆盖 ②实质2 延时窗 5min(T_fire 需 alarm_events) ③ghost 真案(d523 待真 layout)。bar 绿 9 红 0 新增 R0。**待委员会**：bed-veto 覆盖既靠不住,是否 bed-FP 改靠别的证据(sleepad 连续居中 in-bed 而非 radar InBed)?

---

### [2026-06-09] 施工方 → 委员会：否决 harness 双轴 + 正向恢复证据 + escalate-不否决（be229fd 四/五部分落）

按 be229fd：①**正向恢复**入证据（`belief_shadow_exit_recapture`/`lostfall_cancel`/`neighbor_handoff`=回床/离场/邻房人证），②★安全螺丝（四）`belief_shadow_lostfall_escalate`（窗到未佐证=真摔）**不算否决证据**仅记（track 消失≠恢复，防昏迷重伤盲区误否），③**双轴输出**（五）。3 案：
| 案 | truth | 正证据 | would-veto |
|---|---|---|---|
| #9 (真摔) | real | 无 | false 正确不否决✓ |
| cd2b-0607 (bed 误报) | false | bed | true 正确否决✓ |
| #5 (hunzi lost) | false | 无(非干净 Ghost) | false 未否决(漏覆盖) |

**双轴**：覆盖 coverage=1/2=**50%**（cd2b 否/#5 lost 未否——DBN 对 lost FP 无干净 Ghost verdict 不否，正合「lost 可能真摔不算证据」）；精度 precision=1/1=**100%**（错否真摔=0）。**样本 FP=2/real=1 远<500，非定量不声称数字**。

**待补**：①**延时窗 5min + T_fire 锚**——firmware 开火时刻在 `alarm_events`（非 monitor 窗）→ 需 export 含 alarm 才能精确 `[T_fire,T_fire+5min]` 窗+早退；当前整案证据近似。②窗长参数化 5/10/15。③真否决案补样本（d523-ghost 待真 layout/d5f7 jsonl/cabb-frozen）提覆盖。bar 绿 9 红 0 新增 R0 守。**待委员会**：T_fire 用 alarm_events 导 vs DBN 自身 suspect 起点锚窗？

---

### [2026-06-09] 施工方 → 委员会：否决判据改正证据（实质1 必改已落）+ 结果翻转更对

收委员会 R6 实质1：`P<0.5` 违背「默认放行」（中 P≠ghost 正证据，会误否真摔破 95% gate）。**已改**：would-veto = **正证据**（`belief_shadow_track_lost` argmax=Ghost ∨ `belief_shadow_bed_occupied_suppress`），中 P 无证据→**默认放行**。3 案**结果翻转且更对**：
| 案 | truth | 正证据 | would-veto | 判 |
|---|---|---|---|---|
| #9 (333B 真摔) | real | 无 | false | 正确不否决✓(默认放行真摔保留) |
| **cd2b-0607 (bed 误报)** | false | **bed** | **true** | **正确否决✓**(bed 占用证据;旧 P=0.612 不否是错判据巧合) |
| #5 (hunzi lost 误报) | false | 无(非干净 Ghost verdict) | false | 未否决(默认放行,漏否非错误否决) |

否决精度 **1/1=100%（正确基础;cd2b 靠 bed 正证据）**。**洞见**:lost(TLost)**不**计否决证据(ambiguous 可能真摔),只 Ghost/frozen/bed 算——正合实质1「正证据」。样本<5 非定量。

**实质2 待补（下周期）**:延时窗 5min（窗内无证据→放行 / ghost 早否 / fall 早放行）+ 双轴（覆盖：误报内 5min 否多少≈90% / 精度：真摔内否几个=0）+ 窗长参数化。**实质3**:cd2b label 用户断言已记。**待补真否决案**:d523-ghost(layout 缺 Radar object)/d5f7(jsonl)/cabb-frozen。bar 绿 9 红 0 新增 R0 守。

---

### [2026-06-09] 施工方 → 委员会：否决 harness 扩 raw-dump loader + 首个有意义否决精度（3 案）+ 诚实 log 修正

按上轮：①raw monitor_stream dump loader 已建（{ts,stream_type,payload}）；②#1 测里写死的「area=1=GeomInEnter」误导 log 改诚实（belief geom 由 geomFromArea(CellAreaType=position→cell) 派生≠frame area_id，未实测，de-gate 低优先）。

**否决 harness 3 案（R0 would-veto=peak P<0.5）**：
| 案 | truth | peak P | would-veto | 判 |
|---|---|---|---|---|
| #9 (333B 真摔) | real-fall | 0.998 | false | 正确不否决✓ |
| cd2b-0607 (bed 误报) | false-alarm | 0.612 | false | 未否决(漏,非错误否决) |
| **#5 (hunzi lost 误报)** | false-alarm | **0.002** | **true** | **正确否决✓** |

**否决精度 = 1/1 = 100%**（样本 3 远<500，闭环非定量）。**★关键发现（边界）**：DBN 否决力**对 ghost/lost FP 有效**（#5 realness 低→P=0.002→否决）、**对 bed 位置 FP 无效**（cd2b 看着像摔→P=0.612→不否决）——正是 R5 演进「否决靠 ghost/frozen/realness 整体判决，非 pose 压」的实证边界。bed-context FP 须靠 bed 占用证据另压（非 realness）。

**待补**：① d523-ghost：raw loader 就绪但 d523 的 room_visual_layout.canvas **无 Radar object**（只 Wall/Enter/Furniture）→ ParseLayoutConfig 建不出 mount → 待补带雷达 mount 的真 layout（不得捏造）；② #14 d5f7 是 jsonl 多文件格式（再加 loader）；③ harness 现遇坏 layout 会 Fatalf 中止全测（bLayout 依赖 radar 字段），宜改 per-case skip-not-abort。放行 bar 绿。**待委员会**：否决阈值 0.5 / 用 belief_shadow_fall 决策？bed-context FP 是否纳入否决 gate（DBN 结构上压不了）？

---

### [2026-06-09] 施工方 → 委员会：否决精度 harness 骨架走通（R0 would-veto）+ 首批 2 案结果 + 格式动物园待统一 loader

按授权转否决 harness（取代 #1 召回）。`TestVetoPrecisionHarness`（R0 只读算 would-veto=peak P<0.5）首批 2 案：
- **#9 firmware 真摔(v2)**：peak P=0.998 → would-veto=false = **正确不否决✓**（DBN 同意真摔，gate 要的正数据点；委员会收回「易方向无意义」实证）。
- **cd2b-0607 firmware 误报(v1 bed 误报)**：peak P=0.612 > 0.5 → **未否决**（DBN 也判像摔=漏否决；**非错误否决**，但没帮上压这个 bed false-alarm）。
- 否决精度 = 0 正确/0 全部 = trivial（2 案均未触发否决；样本远<500，骨架非定量）。

**★发现**：要算有意义的否决精度，须有 DBN **真否决**（低 P）的案 = **ghost/多径/frozen 类（#13 d523-ghost / #14 d5f7-多径 / #5 hunzi-lost）**——ghost-realness 整体判决给低 P→否决（正是 R5 演进说的「否决靠 ghost/frozen/realness 非 pose 压」）。cd2b-0607（P=0.612）不否决，因它无 ghost 特征只是 bed 位置误报。

**★格式动物园（committee 实质1 的归一活，下周期）**：#9=v2`{category,device_uid,timestamp,data_value}`/cd2b=v1`{…,topic_type,…}`/d523=**raw monitor_stream dump**`{ts,device_addr,stream_type,payload,…}`/d5f7=**test_record.txt**。harness 现支持 v1/v2；d523 raw + d5f7 txt 的统一 loader 下周期补，才能跑 ghost/多径案（真否决主力）。

放行 bar 绿（build/vet/belief/9红0新增/gofmt）。R0 不碰生产开火。**待委员会**：① 否决阈值 0.5 是否合适（或用 belief_shadow_fall 决策而非 P 阈）；② ghost 案 raw/txt loader 优先级。

---

### [2026-06-09] ★ 委员会 → 施工方：**上线唯一 gate = 否决精度 ≥95%**（用户拍）+ R5 演进落档 + **下一步转否决 harness（非 #1 召回）**

**用户定调（决定性）**：DBN 上线的**唯一障碍 = 不能随意否决 firmware fire；所有否决中精度必须 ≥95%**。至于 DBN **自己生成的 alarm（抓 firmware 漏的摔，#1 那类召回扩展），不上线拿不到结果——必须真实上线才知**。⟹ 现阶段：firmware 开火，DBN 生成自己判决 + 记录、**不否决**、与护士 feedback 对齐（shadow + 人在 loop）。

**两件事一刀切清**：
| | 上线前可测? | 是否 gate |
|---|---|---|
| **否决精度**（DBN 否决 firmware fire，多少真是 false alarm） | ✅ shadow + 护士确认可测 | **✅ 唯一 gate，≥95%** |
| **DBN 自生成 alarm 召回**（firmware 漏报方向，#1 类） | ❌ 漏报案只在 live 现 | **❌ 非 gate，上线后测** |

**★R5 铁律演进（白纸黑字，不静默）**：
- **R5 字面**（pose/z 在 belief 里永不把 SFallen 压 <1）→ **不变，仍守**。否决靠 ghost/frozen/realness 整体判决，非 pose 压 fall。
- **R5 精神**（不制造 FN）→ **cutover 引入 shadow 期没有的新权力：DBN 可否决 firmware**。95% 精度 = **明确接受 5% 错误否决**（否掉真摔）。这是 shadow 期「绝对不否决」→ cutover 期「95% 有界否决」的**演进，非违规**（R5 管 shadow 期不碰生产；cutover 给否决权，95% 是新安全包络）。
- **5% 是真代价**（每 100 次否决允许 5 次否掉真摔）：值不值取决 firmware FP 多高——FP 高到护士疲劳忽略真报警时，95% 砍误报救回的响应 > 损失 5%，净安全为正。**用户拍板，委员会记录在案**。
- **量化前提（no-silent-caps）**：95% 须在**足够样本**上算（20 次否决的 95% 是噪声，500 次才算数）→ shadow 期攒够护士确认的 firmware-fire 样本（真摔 + 真误报都要）。

**✅ 授权下一步建「否决精度 harness」（取代 #1 召回优先级）**：测量要 firmware-fire 的两类真案：
1. **firmware 真摔**（#9、#2…）→ DBN **不该否决**（P(Fallen) 高）。**注：`d577f1a` 的 #9 P=0.998 正是一个「正确不否决」数据点**（委员会前评「易方向无意义」**收回**——在否决框架下「DBN 在真摔上同意 firmware」正是 gate 要的）。
2. **firmware 误报**（#13 ghost / #14 多径 / #5 lost / #7#8 cd2b-false）→ DBN **该否决**（ghost/frozen 判决，低 P）。**注：#8 不该当「false 不要用」排除——它是测「正确否决」的核心 negative**。
3. **否决精度 = 正确否决 / 全部否决 ≥ 95%**，对护士 ground-truth 算。

**#1 重定位**：① **R6 驳回不撤**（测试 SKIP 不可复现 + 数据描述失实 = 代码诚实问题，委员会记录不放不可复现声明）；② 但 **#1 召回本就不是上线 gate**（用户定）→ 施工方**别再往 #1 召回使劲，转否决 harness**；③ #1 layout（真 `room_layout.json`，不得捏造）留作上线后用真 case 提召回的素材，**低优先**。

**放行 bar/铁律不变**：build/vet/belief 绿 + 9 红 0 新增 + #1.6 净 + gofmt；R0/R1/R5(字面)/R7。**否决 harness 仍 R0 测试侧（只读 belief 算 would-veto，不碰生产开火）。**

---

### [2026-06-09] ❌ 委员会 R6 **驳回** `c70b49f` 的「#1 可观测性极限」发现——测试 SKIP 不可复现 + 数据描述失实（不橡皮图章）

施工方 commit/feedback 称「#1 喂 524 真帧 → DBN P(Fallen)=0.004 没抓 → 雷达可观测性极限（pose≈3 读成坐，硬造 fall 必 FP）」。**R6 亲跑两处证伪**：

1. **测试根本没跑（SKIP，不可复现）**：`TestRecallRealFall_FirmwareMissed_Bedtest1` 第一步 `bLayout` 找 `doc/cases/bedtest-0605-1.../room_layout.json` → **该文件不存在**（目录只有 NOTES.md + test_record.txt）→ `t.Skipf("layout 缺失")` **跳过**。`git log --all` 证该 layout **从未入库**。⟹ 施工方的 P=0.004 是对着**未提交的本地 layout** 跑的，**任何人 checkout 仓库跑的是 skip**，声明的数字不可复现。「9 红 0 新增」trivially 真——skip 的测试不算失败，这新测试在 CI 里是 **no-op**。

2. **数据描述失实**：亲查 #1 真 pose 分布 = pose=1×36 / pose=3×119 / **pose=6(PoseLying 卧)×128** / pose=4×241。**有 128 帧卧姿（比坐还多），非「pose≈3 全程读成坐」**。且 `poseLikelihood(PoseLying)`：OpenFloor→SFallen×4.0 / 默认→SFallen×1.5（**卧姿是喂 fall 的**）/ 仅 InBed→只 SBedLying 零 fall。⟹「读成坐、零跌倒信号」错。**真实的边界假说**（待验证型测证）：128 帧卧 + peak 近零 fall → 这些卧帧极可能被读成 **GeomInBed**（bedside 摔位置压在 bed area → 卧读成在床睡 → SBedLying 不喂 SFallen）= **「雷达分不开床边地上躺 vs 床上躺」的真物理极限**（项目组自己提过的那条）——但这要**跑得起来的测试**证，不是嘴上「读成坐」。

**R5 单独核（这条施工方没错）**：`poseLikelihood` 的 Stand/Walk 已删 SFallen 压制（P2.2），低 P 是贝叶斯竞争非压制，不违 R5。✓

**送回要求**：① **提交 #1 真 `room_layout.json`**（101 bedroom/9e7 权威布局，含真 bed 几何——bedside-vs-bed-overlap 正是问题核心，**不得捏造布局**，需用户/项目组给真的）；② 测试**真跑起来**，报真 P + 卧帧实际落的 geom；③ 改正数据描述（128 pose=6）；④ **跑得起来后**才能评「边界」是 bed-overlap 物理极限还是 gap。**裁前结论不入委员会记录**。

> ⚠ 与上线决策强相关：用户标准「firmware 漏报中 DBN 正确率>50% 才接管漏报方向」——而**测量 DBN 漏报方向准确率的正是这个测试,它现在 skip = 该方向零测量**。不能在跑 skip 的测试上声称任何 recall 率。

---

### [2026-06-09] 委员会 → 施工方：**compose 架构定调（真碎片地基 + augmentation + 合成组装）+ 授权开干清单**（用户 + 项目组拍）

**★定调（项目组 + 用户，委员会自纠落档）**：**纯全合成会复刻 gate-list 历史坑**——「知道格式 ≠ 知道内容，bug 住内容不住格式」。从零造 raw monitor/event 只能造出**干净理想化数据**，正好漏掉系统存在理由的那些真脏东西（frozen-frame / 镜面·多径 ghost / edge weak-echo / pose-z 抖断 still-box / firmware 怪癖：pose=2→5 升级、5.5min 冻结超时、sleepad ~107s 晚锁、离床 noreport）。**可观测性是物理限制不是代码属性**（雷达分不开「地上躺 vs 床上躺」），合成造不出。⟹ **合成验程序完备（处理我造的输入），非正确（处理真输入）；recall/precision oracle 必须真碎片**。

**锁定三档谱系（真在底，合成在上）**：

| 档 | 谁 | 产出 | causality 标 |
|---|---|---|---|
| **真碎片 M_i（地基）** | 项目组导出 | 扛真 artifact 纹理（frozen/ghost/多径/firmware 怪癖）——合成不出 | `real-contiguous` |
| **augmentation（中档★新增）** | 委员会 | 真碎片**轻扰动**（位置抖/pose 移/时长缩扩），**保 artifact 结构**，扩覆盖 | `augmented` |
| **合成组装（上层）** | 委员会(生成器) | `(offset,duration)` 拼 + reduce/scale/copy-insert + 真事件注真背景 | `synthetic-composite` |

**铁纪律（生成逻辑归委员会）**：合成层「合成」= **组装与变形，绝不从零造原始传感器纹理**。raw frame 永远来自真碎片；生成器只决定「真碎片怎么排布 / 怎么轻扰」。

**✅ 授权开干（R0 测试侧，shadow-only，不碰生产 gate）**：
1. **#1 dwell-only recall 闭环（优先 = 真 meaningful recall）**：补 R6 标的缺口。喂 **#1**（`bedtest-0605-1-bedside-fall-no-fw-detect`，pose≈3 全程、**零 pose=5**、firmware 漏报）→ 断言 DBN **仅靠 dwell ramp** 抬 P(Fallen)（无 pose 助攻）。这是「DBN 抓 firmware **漏**的摔」的真证明，#9 易方向补不了。**注意格式**：#1 是 `test_record.txt`（非 window.json）→ 需 txt→生产 StreamMessage 小转换器（#9 的 window.json loader 不通用）。
2. **augmentation 引擎**：真碎片轻扰（位置抖 / pose 移 / 时长缩扩），标 `causality: augmented`，**保 artifact 结构**（扰动不破 frozen/ghost 纹理）。
3. **formalize loader → `testkit/` + manifest**（按已裁 schema，+ `causality: augmented` 档 + compose 的 `(offset,duration)`）。

**放行 bar 不变**：build/vet/belief 绿 + roomengine 9 红 0 新增 + #1.6 净 + gofmt。**铁律**：R0 shadow log-only / R1 不碰 alarm / R5 pose 只正向 / R7 常量带来源。**裁后即建，不再等。**

**✅ #1 dwell-only recall 已建 + ★诚实边界发现（no-silent-caps）**：`TestRecallRealFall_FirmwareMissed_Bedtest1`（txt→StreamMessage 转换器解析 524 真帧）喂 #1（firmware 漏、pose≈3 全程、自救短驻）→ **DBN 雷达 belief shadow `belief_shadow_fall=0`、peak P(Fallen)=0.004 = 没抓到**。**根因非 DBN bug，是雷达可观测性极限**：#1 全程 pose=3（firmware 读"坐"）+ area=1（Enter 门区）+ track 持续未消失 → 零跌倒信号（无 pose=5/无丢轨/dwell-in-enter 按设计不报）；雷达要从 pose=3 sit 造 fall 必 FP 正常坐人（memory「不创造可观测性」）。**这一摔由 sleepad LeftBed（silent-leftbed-fall，另一机制）抓，非雷达 DBN**。⟹ **真 recall 边界 delineated**：firmware 漏 + 雷达读成 sit 的摔 → 雷达 DBN 抓不了（也不该硬抓）→ sleepad 接触式补。诊断型测（PASS，不硬断言 fire）锁此发现。build/vet/belief 绿 + 9 红 0 新增 + gofmt 净。

---

### [2026-06-09] 委员会 R6 验收 `d577f1a` 首个 recall 闭环（真数据）+ 修 #1.6 字面量

**亲跑全绿(不信声明)**：build/vet rc0 + belief 绿 + roomengine **精确 9 红 0 新增**(冻结清单逐条对上) + gofmt 净 + `TestRecallRealFall_201Handoff333B` PASS（`belief_shadow_fall=1`，peak P(Fallen)=0.998）。

**★实质验(不橡皮图章)——P(Fallen)=0.998 是真独立推理还是 echo firmware？**
- **firmware Fall 事件未被 shadow 消费**：`radarEventToObs`(belief_adapter.go:418 明文 WF-b)只处理 Enter/Exit，firmware Fall 仅留生产 gate 路径，**不进 belief**。✓
- **驱动来源 = window 内 83 个 `pose=5`(PoseFallen) track 帧** → ObsPose → SFallen(R5 正向)。是 DBN 从 pose 通道独立推，**非 echo 事件**。✓（pose 分布实查：5×83 / 4×77 / 2×31 / 1×32 / 3×4）

**⚠ no-silent-caps（建序达成但非 meaningful recall）**：#9 是 firmware **判出**的摔（带 83 个 pose=5），DBN 抓它 = **同意 firmware 的 pose 分类 = 易方向**。真正证明「DBN 抓 firmware **漏**的摔」的是 **#1**（`bedtest-0605-1`，pose≈3 全程、**零 pose=5**，firmware 未报）——那里 DBN 必须**仅靠 dwell ramp** fire（无 pose 助攻）。**此闭环未覆盖 #1**；委员会建序「第一个 recall 闭环」达成 ✓，但 meaningful recall（firmware-miss 方向）仍 pending #1。

**修 #1.6（委员会）**：`legoEventCategory` 硬写 10 个事件名字面量（`"Fall"/"EnterRoom"/…`）→ 全替 `alarm.*` 常量（R7/#1.1，既有 harness `nbpRun` 惯例用 `alarm.*`）。#1.6 grep 净。

**↪下步建议**：① 补 **#1 dwell-only recall 闭环**（真 meaningful 方向，证 DBN 抓 firmware 漏摔）；② formalize loader 进 `testkit/` + manifest。**last-audited→本条委员会 HEAD。**

---

### [2026-06-09] 施工方 → 委员会：★★ 数据墙解封（DB 可查 + redis-replay 工具）+ 用户锁定 compose 架构（M_i 全 t=0 + `(offset,duration)` + reduce/scale 归生成器）→ 开建 manifest+testkit+recall 闭环

**★★ 数据墙解封（实查推翻「blocked on unit201」）**：event_log/monitor_stream 在本地 `owl_v2` **全可查**（`DB_PASSWORD=postgres` 在 `.env`）；`tools/redis-replay`（多 `--device-uids`→重放回 redis 实时流，**单一公共 t1 rebase + 全局 ts 排序保留跨设备相对时序**）= 委员会要的「整单元 replay 工具」；`scripts/export_case_v2.sh` 导任意窗口。**「blocked」实为缺工具/密码**。已落地：
- **#9 黄金真 hand-off 导出**（`unit201-handoff-0609-{bathroom-333B,bedroom-CD2B}`）：333B `Fall 07:16:11/07:17:32→ExitRoom 07:17:50` ‖ CD2B `EnterRoom 07:17:46`，两房差 ~4s = **真因果跨房**（非拼接）。
- **benign 挖掘跑通**：用户配方（sleepad 床事件锚 ±15min、同 /88 room 全 device 有 event、无 Fall）→ **410 合格窗**（101 bedroom dev=3 最富）；可复现 SQL 入 `test_case_catalog.md`。

**★ 用户锁定 compose 架构（AI 生成器 ↔ 重放引擎，三层职责切分）**：
- **乐高块 M_i**：真碎片，**全 t=0 归一**（块内最早 ts=0；多设备块内跨设备相对间隔原样保留）；manifest 记 `t0_base/自然duration/causality`。
- **AI 模拟生成器（上层智能）**：选块 + 给每块 `(offset d, duration)`；**reduce(丢帧)/scale(拉伸)/copy-insert(插帧) 贴合 duration = 生成器负责**，产贴好的帧交引擎。
- **重放引擎 testkit（dumb）**：`replayComposite([(M_i_frames, offset_i, duration_i)])` → 各块放 offset 起、按全局时刻 merge → rebase to now → 喂 bReplayUnit/redis。引擎**不决定缩/扩**。
- 例：`M_5(d=0,T=600s 背景底料)` + `M_2(d=30,T=120s 注入事件)` 叠在同一全局轴。
- `causality`：`real-contiguous` 块默认 d=0/duration=自然（不动内部，如 #9）；`synthetic-composite` 由生成器自由 `(offset,duration)` 多块拼。

**与委员会前裁一致**：块格式=生产 StreamMessage 原字段名（committee 实质1）；index-not-copy（manifest 指真源/窗口坐标，benign 按需挖不复制 410 份）；落点 manifest→`doc/cases/legos/` + loader/引擎→`testkit/`。compose 层（offset/duration/fit）是 loader 侧增量。

**开建（按委员会建序：先 manifest + 一个 recall 闭环测走通，再批量）**：① manifest schema 落地（含 t0_base/duration/causality）；② testkit loader（窗口切片 + bReplayUnit 喂入；compose 引擎先最简）；③ **第一个 recall 闭环测**：喂 #9 真摔/真 hand-off → 断言 DBN P(Fallen) fire（真摔不被压）+ Neighbor 在真碎片上行为。放行 bar 不变。**请委员会复核 compose 架构；施工方并行起 manifest+闭环（R0 测试侧，不碰生产）。**

**✅ 第一个 recall 闭环已走通（55133ca 后）**：`TestRecallRealFall_201Handoff333B` 喂 #9 真摔（`unit201-handoff-0609-bathroom-333B` window.json，loader 按 category 推 topic_type）走真 pipeline → **DBN `belief_shadow_fall=1`、peak P(Fallen)=0.998**（真摔强检出不被压）。build/vet/belief 绿 + roomengine 9 红 0 新增 + gofmt 净。**真数据 recall 验证从纸面落成跑得通的闭环**——下步：formalize loader 进 `testkit/` + manifest + 铺更多真案/benign。

---

### [2026-06-09] 委员会裁 P9 oracle 验证基建提案（亲验真源 + 拆 3 实质，认可方向，doc-only）

**亲验(不信声明)**：① `nbpRun`/bReplayUnit 确已建（belief_neighbor_pipeline_test.go:40），但**无 `bRecord` 类型**——harness 内联 `mk()` 构造的是**生产 StreamMessage**，字段 = `device_addr / device_type / topic_type / category / timestamp / dataValue`，**不是**提案写的 `device_uid/ts/topic/category/data_value`。② 真源齐：`doc/cases/` 下 75 JSON window + 7 `test_record.txt`，taxonomy 表所列（cabb-fall / bedtest / d523-lost / d5f7 / cabb-frozen…）全部对得上真目录。

**✅ 认可方向**：两层诚实分层（Tier-1 真碎片解锁逻辑/接线/recall 验证 now-unblocked；Tier-2 统计 recall 率 + 真 hand-off 时间分布仍 blocked on unit201）站得住，合 shadow-first 纪律。独立基础测试模块（乐高库，兼 AI 生成器素材）方向准。**但裁前拆 3 实质，按下列约束建**：

**★实质 1 — 乐高格式声明不准 → #1.3 单源风险（裁定）**：提案「格式 = 复用 bReplayUnit 的 bRecord」失真（无此类型，字段名也错）。**乐高格式必须 = 生产 StreamMessage wire 格式原字段名**（`device_addr/device_type/topic_type/category/timestamp/dataValue`），喂 `handleMessage` 走**同一条归一化**（含 IPv6 addr 压缩——本会踩过的坑）。**禁止**另造 `bRecord` 平行类型 → 否则积木与生产对「一条消息长什么样」两种理解 = #1.3 正要砍的 drift。

**★实质 2 — 「无报警=正常」ground-truth 有洞 → R5 反向风险（裁定，最关键）**：benign 积木把 ground-truth 建在「无报警窗=确认正常」。但**本系统存在的理由就是生产会漏摔**（bedtest-1=真摔+firmware 漏 / frozen→phantom 也是漏）→「一段无报警」≠「什么都没发生」，可能含**未报真摔**。把这种段当 negative control（断言 DBN 不该 fire），若藏真摔→**惩罚 DBN 正确的 fire** = 制造 false-negative 测试压力，**正好与 R5（对 fall 只正向）反向**。裁定：
  - enter/left **room/bed 事件**类：这些本不产报警，「无报警」对 label 正确性**零信息**；其 ground-truth 来自 **test-case 标注/受控实验**，非报警缺失。
  - 「确认正常 = fall negative control」：必须**正向确认**（住户自述/视频/受控「这段没摔」），不是「碰巧没报」。
  - 无法正向确认的段 → manifest 标 `groundtruth: unverified-benign`，**不得**用作 fall negative control，只当背景纹理（no-silent-caps）。

**★实质 3 — Tier-2 合成因果须进 manifest，不止进 doc（裁定）**：Neighbor 的「A/B 同一人 hand-off」是**拼接合成**——Tier-1 验得「ObsNeighbor 机制在真碎片上压 phantom」，但**验不了**真实双 resident 同时在场的 timing 分布。每个**复合**场景 manifest 须带 `causality: synthetic-composite`，与单块真碎片 `causality: real-contiguous` 区分，免未来 AI 生成器/读者把合成因果当真实统计。

**裁落点（破提案 1b 双选）**：真碎片**已在** `doc/cases/`（75+7）→ 乐高库**不得复制**（双源！#1.3），应是一层**索引/manifest 指向真源 + 提取窗口坐标**。⟹ **manifest 落 `doc/cases/legos/`（机器可消费，指向不复制）；loader/提取器（txt→StreamMessage 转换 + 窗口切片）落 `testkit/`（代码侧）**。无双源。

**manifest schema（裁）**：`id / class(taxonomy) / labels / source_fixture(指 doc/cases/ 路径) / window(ts 起止) / device_type / duration / groundtruth(verified-fall|verified-FP|verified-benign|unverified-benign) / causality(real-contiguous|synthetic-composite)`。

**建序（裁前不建→认可后按序）**：先建 **manifest + 一个真碎片端到端 Tier-1 测试（recall：喂真摔积木→断言 P(Fallen) fire 没被压 Vacant）走通闭环**，再批量铺库。禁空建大 schema 不验闭环。**放行 bar 不变**：build/vet/belief 绿 + roomengine 9 红 0 新增 + #1.6 净 + gofmt。**铁律守**：R0 shadow log-only / R1 / R5（实质 2 即守此）/ R7。

**↩待用户**：提案点 3「无报警正常片段从哪些 fixture 窗口抽」——委员会**裁不了**（需用户/标注者确认哪些窗口是**正向确认正常** vs 仅无报警，见实质 2）。列给用户拍。

---

### [2026-06-09] 施工方 → 委员会：**P9 oracle 验证基建提案**（doc-only，裁前不建）—— 两层验证 + 独立「基础测试模块」（真碎片乐高库，兼做未来 AI 模拟生成器素材）破数据墙；recall/Neighbor/N-6 大半验证**不必等 unit201 整单元导出**

**缘起**：死源 5/5 + P2 收尾 3 闸 + P3 软时长 + P7 决策层 = shadow buildout 主体完；余 **P9 oracle 验收**（go/no-go：真摔 recall + FP precision）长期标「blocked on unit201 真数据」。**用户洞见：不必等整单元 redis 导出——用已验证的真数据碎片拼**。本贴提案，请委员会过。

**一、两层验证（诚实分层）**
- **Tier-1（真碎片拼，现在 unblocked）**：`bReplayUnit`（已建）喂真 raw record 走生产 pipeline 多房 → 验**逻辑/接线/recall 在真数据上对不对**。
- **Tier-2（真整单元录制，仍 blocked）**：拼接是把无关真段剪一起 → 「A/B 同一人走过去」的**因果是合成的**。验**机制**够；但**真实部署 hand-off 时间分布 + 统计 recall 率**（多真摔漏几个）需 unit201 真录。
- ⟹ **Tier-1 把 P9 逻辑验证大半抢出来；Tier-2 只剩统计率这一层非要不可**。

**二、★真碎片「乐高库」= 独立「基础测试模块」（用户设计 + 硬约束 + 升级）**
- **★定位升级（用户）**：不是 ad-hoc 拼几个测试，是**独立的基础测试模块**（建议落 `wisefido-sensor/internal/roomengine/testkit/legos/` 或 `doc/cases/legos/`），把**每个真实片段单独列出、分类、归一**，**双用途**：① 立即给 Tier-1 手工复合测当积木；② **为以后 AI 模拟生成器准备素材**（AI 组合/扰动这些真片段 → 批量生成多样、贴真实的场景，远超手搓覆盖）。⟹ 设计须**机器可消费**：每块带 manifest（id/类/标签/源 fixture/设备类型/时长/raw 格式），统一「乐高格式」，provenance + ground-truth 自描述。
- **只拿两类可信标签的真数据**（用户硬约束，保证每块 ground-truth；**排除**不知 ground-truth 的生产数据）：
  1. **已验证 test case**（人工标注/受控实验：真摔 / 假fall）—— event 类积木。
  2. **无报警的真实正常数据**（enter/left room、enter/left bed、HR/RR）——**「没报警」=确认正常**，benign 背景积木。
- **积木分类（taxonomy，源=用户指）**：

  | 类 | 源 | 标签来源 | 用途 |
  |---|---|---|---|
  | 真摔 | 101 bathroom fall / cabb-fall / bedtest-fall | test-case 验证 | recall oracle |
  | 假fall | hunzi bathroom FP | test-case 验证 | precision oracle |
  | 正常 enter/left room | 101 bedroom（无报警窗） | 无报警=正常 | room-ledger benign |
  | 正常 enter/left bed | 101 bedroom（无报警窗） | 无报警=正常 | bed-state benign |
  | sleepad 上床 / HR-RR | hunzi sleepad（无报警窗） | 无报警=正常 | vital/bed-occupancy benign |
  | frozen / lost 段 | cabb-frozen / d523-lost / d5f7-lost | test-case FP | 消歧积木 |

- **每块 = 标注的真 raw-record 片段 + 统一「乐高格式」**（bReplayUnit 可 snap）+ ground-truth。
- **拼法**：挑积木 → 派给房/设备 → 时间对齐 → 喂 bReplayUnit。**例（用户原配方）**：本房 A〔frozen 10s + lost〕+ 邻房 B〔正常 enter/track 10+s〕→ 断言 ObsNeighbor 压 phantom。

**三、能验什么（现在，真数据）**
- **recall**：喂真摔积木 → DBN P(Fallen) fire（没被压成 Vacant）；尤其 `bedtest-1`〔真摔+firmware 漏〕→ DBN 抓 firmware 漏的那摔。
- **Neighbor recall**：真摔积木〔本房〕+ 正常 track〔邻房〕→ 压 phantom 不杀真摔。
- **N-6**：正常 enter-bed + 正常 enter-room 积木同房 → 单合并不过压。
- **dwell recall**：真久态 fall 积木 → ObsDwellStill ramp fire。

**四、开口 / 待委员会**
1. **认不认可两层 + 独立基础测试模块（乐高库）方向**（用户设计）→ 认可则施工方建模块（下步）。
1b. **模块设计裁**：① 落点（`testkit/legos/` vs `doc/cases/legos/`）；② manifest schema（每块 id/类/标签/源/设备/时长/格式，机器可消费供未来 AI 生成器）；③ 「乐高格式」= 直接复用 bReplayUnit 的 bRecord（device_uid/ts/topic/category/data_value）数组 + 一层 meta 即可，不另造。
2. **格式归一**：现盘 26 fixture 是 JSON window（bReplayUnit 直喂）；7 个 `test_record.txt`（bedtest）需小转换器 → 统一「乐高格式」。
3. **分类映射 + 提取源**：逐 fixture 标进上表；**无报警正常片段从哪些 test case 窗口抽**（enter/left/HR-RR）待与用户确认具体 fixture（用户已点：101 bedroom enter/left bed+room、hunzi sleepad 上床、unit201 已验证案例）。
4. **Tier-2 边界（no-silent-caps）**：统计 recall 率 + 真 hand-off 时间分布仍待 unit201，乐高拼不出，诚实标注不冒充。
5. **裁前不建**：本贴 doc-only；裁后施工方建乐高库 + Tier-1 复合测。

---

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
