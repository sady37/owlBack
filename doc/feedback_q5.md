# Q5 反馈日志(项目组侧)— belief 全空间区域占用重写 + cd2b 床边真摔 FN

> **QA 分离(2026-06-12,因 git 同步写冲突)**:P5 主题拆成两文件各写各的——
> **本文件 `feedback_q5.md` = 项目组侧**(提案/问题/根因/复算请求);委员会裁决/审查/工单写 `feedback_a5.md`。
> 两侧不写同一文件 → push 不再非-fast-forward 撞车。倒序,最新在上。
>
> **协作协议**:项目组提案 → 委员会裁 → 裁后建;裁前不建 / 需新岔口列选项不擅决 / 直接 main / cd wisefido-sensor 跑 go / 放行 bar = build/vet/belief 绿 + 0 FAIL。

---

## 当前状态(2026-06-12)

**定论(用户/架构师)**:DBN 自第一天(`b3a45ca`,2026-06-01)状态空间建错。重写方向 = 把 Room 层从「姿态 9 态」(位置不进状态,只作观测条件)换成「**全空间区域占用 × {直立, 倒地}**,含盲区一等态」;转移=空间运动;观测=区域有无 track;fall 从「占用 + 不可见或静止 + 未离场」涌现。`belief_shadow.go` 的 lost-fall sweep(MovingPrecondition 译反)删除变涌现。

**验证方法论**:用进程内 `belief_replay` 单测(干净可复现),不再 live 重放(bedSessions 跨 run 残留)。fixture = `case-cd2b-0604-16141631`,完成判据 = 该 fixture 必须 fire。

---

### 架构定性(项目组,待委员会印证)

两层各建不同的东西,正交:

- **Track 层**(`belief/track.go`,**保留**)= 观测质量滤波器,答「这条 blip 是真人还是反射」。状态 `T={None,Real,Ghost,JustLeft,Lost}`,5 态 HMM。数学核心 = co-existence 耦合 ρ(孤立 track 不可能 ghost——ghost=真人反射,必有共存 partner)。唯一对外输出 `P(Real)∈[0,1]`。
- **Room 层**(新空间模型,**重写**)= 状态估计器,答「人在哪个区、倒没倒」。状态 = Zone × {直立, 倒地}(或等效:空间 zone + `SFallen` 吸收态)。数学核心 = 空间转移阵 + 盲区一等态 + dwell 生存函数。
- **耦合(已落地)**:`belief_shadow.go:354` `o.Conf *= pReal`(`pReal=P(T=Real)`)。ghost→`P(Real)→0`→Conf→0→似然经 `temper` 退火成单位阵→不驱动空间推断。等价原始提案 §4.3 边缘化 `P(fall|y)=P(fall|real,y)·P(R=real|y)`。
- **工程经济**:保持 `numStates=9` → `Vector`/`normalize`/`Max`/`temper`/`Decider`/`belief.go` 全不动,重写收敛进 `state.go`(语义)+`model.go`(A)+`likelihood.go`(L)三文件(`SFallen` 仍须是 9 态之一,Decider 引用)。

**项目组提出的三处需拧紧(委员会待裁)**:
1. cd2b 的 fire **不来自 Track 层判 ghost**——co-existence 对孤立被子 ρ=0 判它 Real,反而保护它。fire 杠杆在 Room 层(sleepad 权威释放床占用 + 真人 absence + 未离场 → blind dwell ramp Fallen)。
2. absence 须定义在 realness 门控后的 track 上(「任何可见区无 `P(Real)>阈` 的 track」=有效缺失),而非「无 track」。
3. FP 安全性反转:旧 `A` 靠「→Fallen 极小,沉默不造跌倒」;新模型 blind-zone 久缺**必须**从沉默 ramp Fallen → 守门挪到 blind 进出纪律 + dwell 时序 + recapture/Exit cancel,且守[偏分监控铁律](只有极近两事件有向 hand-off 能排除 lost-fall)。

**盲区无几何 — 对上面架构的实质修正(2026-06-12,用户提出)**:layout(`room_visual_layout`/`radar.areas`)只编码雷达**覆盖到**的 area 对象;盲区=其补集,**任何数据源都不画它**;firmware 只报 track,不报「看不见」。所以**盲区不能是有几何、可定位的状态**。
- 重述:**Blind = 「有真人在场(近期证据)+ 没有任何 real track 解析出他 + 没有 Exit 事件」**——用否定定义,可观测性来自「track 从有到无 + 无 ExitRoom」,不来自几何。
- Near/Far/Bath 子态**不能**来自盲区几何(未知),只能来自**消失前最后一帧位置**(last-seen area_type + 门距,已知)→ 改成**入口条件化**:`Blind-from-Door`(高 exit 先验)/`Blind-from-OpenFloor`(高 fall 先验)/`Blind-from-Bath`。状态索引是「怎么进的盲」(可观测),不是「现在在盲哪」(不可观测)。
- cd2b 把此逼到极致:被子 track 把 Bed→Blind 转移**遮住**(连「track 没了」都观测不到)→ 该转移只能靠 **sleepad LeftBed 与 radar 床 track 矛盾**推出,对上原始提案 §7 诚实下限。

---

### cd2b 床边真摔 replay FN — 根因实证(供 CodeX 独立复算)

**fixture**:`doc/cases/case-cd2b-0604-16141631`(201 卧室,radar `cd2b` + sleepad `1641`,UTC 22:14–22:31)。
**oracle**:22:23:20 sleepad **LeftBed**(人跌床边、雷达不可见);之后 radar 持续在床报 lying(被子=静止反射伪装成在床的人)。`alarm.json=[]`(生产漏报)。**正确结果 = ~22:25 应 fire 一条 Fall**(被子是反射,不该按「在床睡」豁免)。
**replay 结果**(in-process belief replay,cd2b fixture):`belief_shadow_fall` 空 → `SFallen` 从未 confirm → **DBN 也 FN**。

**FN 代码链(file:line 可核)**:

1. **孤立被子 track 不被判 ghost,被当人。** `belief_shadow.go:349-354`:fall 证据 `o.Conf *= pReal`,注释明写 `real lone faller pReal≈1(孤立 ρ=0 不可能 ghost)→ 无折扣`。即 co-existence 对孤立被子判 Real → Conf 不降 → 被子 pose=lying@bed 正常驱动 Room。**与 ghost 无关。**
2. **床区躺→倒地的唯一出口 = BedReleased 分支。** `likelihood.go:161` `case inBed && !c.BedReleased:` → **SBedLying**;`likelihood.go:165` `(inBed && c.BedReleased)` → **SFallen** 候选。由 `Observation.BedReleased` 二选一。
3. **BedReleased 生产闸门**:`belief_shadow.go:361` `if bedReleased { o.BedReleased = true }`;`belief_shadow.go:229` `bedReleased := bedSt.BedConfidence > 0 && bedSt.BedStatus != 0`(0=InBed);`bedSt := tm.BedOccupancyState(nowMs)`(:228)。
4. **闸门输入错**:回放期 `BedOccupancyState` 返回 `BedStatus=0(InBed)/conf90` → `bedReleased=(90>0 && 0!=0)=false` → BedReleased 永不置位 → 走 SBedLying → P(Fallen) 不 ramp → FN。

**根因定位**:**不在** belief 层(D3 止血 `belief_shadow.go:359-368` 接线正确),**在上游 bed 占用融合 `tm.BedOccupancyState`**:sleepad 22:23:20 的权威 LeftBed 没能盖过 radar 被子的 InBed 证据,bed 贝叶斯 scorer 仍输出 InBed/conf90。契约:sleepad 接触式权威 > radar(见 `bed_fusion_authority_model` / `bed_stale_leftbed_vetoes_radar_inbed`)。

**请 CodeX 独立复算/证伪**:
- **(a)** FN 根是否就是 `BedOccupancyState` 在 22:23:20–22:30 窗内仍返回 `BedStatus=0/conf90`?即 sleepad LeftBed 为何未释放 bed_state——radar 被子 InBed 证据权重 ≥ sleepad LeftBed?还是 LeftBed 根本没喂进 scorer?
- **(b)** 若把 bed 融合修对(sleepad LeftBed 胜出 → `BedStatus!=0` → `bedReleased=true`),现有 D3 止血是否**即可**让 cd2b 在 ~22:25 fire?还是 pose/dwell ramp 时序不够、仍需别的证据?
- **(c)** 计划中「全空间区域占用×姿态」重写是否真能消掉这个 band-aid——还是只是把同一个「sleepad vs radar 谁定床占用」的权威问题从 BedReleased 闸门挪到 blind-zone 进入条件上(即重写**不**自动解 (a),仍依赖 bed 融合权威修对)?

**项目组判断(待 CodeX 印证)**:(a) 成立;(c) 重写**不**自动解 (a)——区域模型让 blind-zone 成一等态、删 MovingPrecondition 译反,但「状态离开 Bed 区」仍取决于 sleepad LeftBed 能否压过 radar 被子的床占用。故 **bed 占用融合权威(sleepad>radar)是 cd2b fire 的必要前置,与状态空间重写正交,两者都要做。**
