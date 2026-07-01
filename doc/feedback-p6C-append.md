
# §NN 增补（C 提案 — ghost 打分统一为「运动学存疑轴」三档布尔 + 与 lost_fall 整合；含 C 五次振荡的诚实记账 + G 三处纠偏）

> 本节是 C 委员会**独立设计提案**（非 A 的落码复审），起于「split 打分怎么设计」，经五轮振荡 + G（Review 组同侪）三处纠偏后收敛。行号/常量经 grep 核实（clone @2e4878a，`tools/Xsensorv1/`），**行号以本节为准**（C 前几轮口头引错的已订正）。**结论待架构师裁**——C 不自行落码。

## 〇、TL;DR

- ghost 打分**不引入浮点 score / 乘性 K**，回到「定罪是布尔」。统一为一根**运动学存疑轴**（离锚点净距三档）：`≤50 GLUED` / `50~200 HOLD` / `≥200 WALKOUT`，配现有 3-tick 棘轮（`ghostSustainRun`, :481-483）。**无除法、无比值、无 oracle 阈**。
- **HOLD 的正身 = 两消费者反向处置**（G）：压制轴 HOLD→继续压（FP 侧廉价可逆，信得宽）；删除轴 HOLD→否决删除（FN 侧昂贵不可逆，信得严）。HOLD 里的轨两头同时安全，是 limbo，非折中。
- 与 lost_fall 整合：三档轴同时兼任 §10.3-b「条件2 坐实」与「条件3 R≤120 闸」——本是同一根「钉死程度」轴被拆成两判据，合并。**整合范围只动 lost 分支。**
- **present-frozen 归压制轴，不进删除轴**（G 纠偏）：present-frozen 的 FP 是 floor 误火，压制轴（GhostActive→StillBoxSec=0, :1077）已中和；删 present 轨=固件重报 churn，违背「不删轨靠压制」自定原则。C 上一轮提的「present∨lost 双分支 + neighbor coast 断言」整块**消解，不需要**。
- 该轴统一 **interfer-born / split / 门距分** 三路；**mirror 留几何轴**（无固定锚点），保留自有 `BoxRange≤120`（`mirrorResidualConfineCm`, :610）。

## 一、C 的五次振荡（诚实记账，这是本节该留的东西）

C 在「ghost 打分 / 整合」上反复五次，**同一根因**：

| # | C 的提议 | 错在哪 | 谁纠 |
|---|---|---|---|
| 1 | 新造 `SplitGhostEverMs` latch | 多余，违背 ghost 生命周期统一 | 自查 |
| 2 | `MaxGhostPreSplit` 回滚快照 | 违背「ghost score 对 lid 生命周期统一、单调只升」 | 架构师原则 |
| 3 | 把 branch② 读 `SplitObservingSinceMs` 定性为 bug | 忽略条件3 pose/dz 硬闸兜底——FN 上安全，非 bug（修法见四.4） | 架构师原则 |
| 4 | `静止时间/max(净距−50,0)` 比值 + oracle 阈 X | 违反 C 自己「定罪是布尔」；比值在 50~200 带「往 real 拉」= partial-清罪 = 破坏自己要的单调；X 标不了（[[fall_data_is_artificial_test]]） | **G** |
| 5 | 把 present-frozen 提到删除腿（present∨lost 双分支） | present-frozen FP 归**压制轴**非删除轴；删 present=churn，违背 C 早轮自定「不删轨靠 ghost 压 stillbox」 | **G** |

**根因（五次一以贯之）**：C 每遇一个 FP/FN 焦虑，就想在**当前正讨论的那根轴上加动作**，忘了它归属别的轴。#1-4 是想让 ghost 轴自己「区分真人」（区分真人在 pose/dz 硬闸、walkout latch、floor/fire 独立线，§6.7）；#5 是想让删除轴管 present-frozen 的 FP（那归压制轴）。**正解永远是克制：让当前轴只做它该做的粗动作，把焦虑交给归属它的正交轴。**

## 二、为什么单浮点标量必然矛盾 + HOLD 的正身

### 2.1 单标量的数学不相容
想让一标量同答两不相容问题：**Q_FP**「现在该不该压/删?」要能**降**（真人一动就松）；**Q_history**「lid 生涯像不像 ghost?」要**只升**（否则 §10.2 蒸发洞回来）。塞一个标量→必在「走开要不要降分」撞车。拆两个各自单调的量→矛盾消失。三档布尔是最简布尔形。

### 2.2 ★ HOLD = 两消费者反向处置，是设计正身（G 讲透，C 上一轮只写了删除侧）
**同一个 50~200 带，两消费者按各自成本不对称反着用**——这是 G 早轮「理由位掩码 / 两消费者按硬度各取」的兑现：

| 消费者 | 成本属性 | HOLD 处置 | 信任口径 |
|---|---|---|---|
| stillbox 压制（FP 侧） | 廉价、可逆 | **继续压**（单调 Convicted 穿过 HOLD） | 信得**宽** |
| lost_fall 删除（FN 侧） | 昂贵、不可逆 | **否决删除** | 信得**严** |

→ HOLD 里的轨：**被压 stillbox（无 FP）+ 不被删（无 FN）= 两头同时安全**。**HOLD 不是折中，是唯一对两轴都最安全的 limbo。** 同一根量，压制信得宽、删除信得严——这才是三档轴的设计正身。

## 三、G 版：运动学存疑轴三档（C 采纳，去除法/去比值/去 oracle）

**静止时长**已有载体 = 现有 3-tick 棘轮 `ghostSustainRun`（:481-483，`>=3 → MaxGhostSustained=true`）。**不拿它除漂移**。分开：静止时长走棘轮（布尔 sustain）；漂移只当**闸**（贴没贴锚点），不当分母。

| 离锚点净距 | 判定 | 机制 | 阈来源（确定性几何量） |
|---|---|---|---|
| **≤ 50** | GLUED → 喂 3-tick 棘轮，坐实即定罪 | 现有 ratchet，gated on glued | stillbox 半框（:1072 注释）；吸收 glue 10（:26）+ 噪声地板 |
| **50 ~ 200** | **HOLD**：不定罪、**也不清既有定罪** | 单调轴「存疑」态（二.2 两消费者反向处置） | — |
| **≥ 200** | WALKOUT → 永久 real | 现有 latch `SplitEverWalkedOut`（:128/:158） | `splitWalkOutCm`=200（:27） |

三阈 **10/50/200** 全确定性几何量，**无 oracle 参数**。

### 3.1 HOLD 优于 C 的「连续往 real 拉」（G 的核心，C 服）
50~200 带正确动作 = **什么都不做**：不定罪（可能刚要走开的真人）、**不清既有定罪**（可能是 §6.8 [[split_ghost_walk_anchored_score]] 慢游走 ghost：漂几十 cm 仍钉源附近）。单调轴上「存疑=不动作」既不误压也不误清。C 的比值非在这带给连续输出、方向还错（往 real 拉→partial-洗白慢游走 ghost=FP 源）。**HOLD 和「定性只升」自洽；C 的比值恰在这带破坏它自己要的单调。**

### 3.2 「离锚点净距」而非「离出生点净距」——★ 逐路标清锚点语义（G 提醒，防误读）
`splitWalkOutCm` 现判 `distInt(BirthPos, now)`（离**出生点**，:157）。改判**离锚点**：ghost 绕出生点画圈可满足「离出生点≥200」但从没离源→漏；离**锚点**则绕源画圈永远 ≤glue→正确判 ghost。**但「锚点」每路语义不同，不可当同义：**

| 路 | 锚点 = | 是否防「绕圈 ghost」 |
|---|---|---|
| **split** | spliter 前一tick 位（`SplitSpliterX/Y`, :124）| ✓ 真外部锚点，绕圈被挡 |
| **interfer** | interfer cell | ✓ 真外部锚点 |
| **门距 born** | **出生点**（= net-from-birth，正是 C 早轮反对的量）| ✗ **不防绕圈，也不需要**——born-ghost 是「出生即静止的反射点」（出生点≈锚点、本就不动），仅因不动才无害。**别误以为门距路径也拿到了绕圈保护。** |

## 四、与 lost_fall 整合（范围只动 lost 分支）

### 4.1 核心洞：三档轴 = §10.3-b 条件2 坐实 + 条件3-R，本是同一根轴被拆开
lost_fall 删除 = 条件1 ∧ 条件2 ∧ 条件3。核实现码：条件2④ 用 `MaxGhostSustained`（:2065）；条件3 R≤120 **只 mirror 路径有**（`confirmedMirrorResidual` :2071 `BoxRangeWithinMs ≤ 120`）；**exitCoupled ①interfer/②split/③门距 无 R 闸**（:2029-2050，只过 pose/dz :2037）。「条件2 像不像 ghost」与「条件3 R≤120 赖着没走」**测同一件事——钉死程度**。`≤50 GLUED` 一条同时满足二者 → 整合成 `kinematicGlued(ts,nowMs)→{GLUED,HOLD,WALKOUT}`。

### 4.2 整合后 lost_fall（伪码，只动 lost 分支）
```
delete(ts) =                                                          // 仍 gate 在 nowMs-LastObservedMs >= presenceCoastMs（:1576/:1585），lost-only
  条件1: (ExitRoom≤30s ∧ lost) ∨ (ExitRoom==null ∧ 全程np==1)        // exitCoupledLostMs=30_000, :604
  ∧ 条件3-pose: poseEvictable(LastPose) ∧ last2Dz ≤ 40               // exitLostMaxDz=40, :606；先过硬闸省算
  ∧ (
      // 运动学轴（三路，各自锚点，都 ∧ kinematicGlued==GLUED；HOLD/WALKOUT → 此支假 → 不删）
      // GLUED 兼任 条件2坐实 + 条件3-R；GLUED 的「坐实」= 现有 3-tick 棘轮 ghostSustainRun（:481-483）
        ( InterferBornSinceMs>0                    ∧ kinematicGlued(anchor=interfer源)==GLUED )   // ①interfer
      ∨ ( 门距分>65（85*d/150, d≥115cm, :2045）      ∧ kinematicGlued(anchor=出生点)==GLUED )        // ③门距
      ∨ ( kinematicGlued(anchor=spliter)==GLUED )                                                  // ②split：GLUED 独扛定罪，无额外前置项（4.4）
      ∨
      // 几何轴（mirror 独立，锚点不适用）:
      ( MaxGhostSustained ∧ MaxRoomNp≥2 ∧ BoxRangeWithinMs ≤ 120 )   // mirror 自有 R（:610）
    )
```

### 4.3 门距分公式真实常量（以码为准，G 核对 @2e4878a）
真实常量：`bornGhostDoorGain = 85.0`（:613）、`bornGhostDoorScaleCm = 150.0`（:614）、`bornGhostEvictThresh = 65.0`（:615）。判据 [:2045] `85*d/150 > 65` → **d > 114.7 ≈ 115cm**（对齐码注释 :2026「85*d/150 >65 ⟺ >~115cm」）。
⚠ C 早轮两次都抄错 gain：口头「60/112.5」→ 上一版「80/121.9」，均非码值。gain 是 **85 不是 80**，阈距 = **115cm 不是 121.9cm**。此处落在 lost_fall 删除阈上（FN-critical），必须用 115。

### 4.4 ★ Gap1 收口 = kinematicGlued **替换** SplitObservingSinceMs（不是并列，G 精确化）
branch②（:2040）现读 `SplitObservingSinceMs`（成员身份，含 walkout race 的赢家=真人）。应读 **`kinematicGlued(anchor=spliter)==GLUED`**——赖锚点者=GLUED=split 的定罪；`SplitEverWalkedOut`(=WALKOUT，:128 声明 / :150 倒地相路径 / :158 walkout 路径置位) 天然把赢家挡在删除外。**GLUED 本身即 split 定罪，不再需要单独 membership 项** → Gap1 自然关闭。是替换，非并列。

### 4.5 ★ present-frozen 归压制轴，删除保持 lost-only（G 纠偏，砍掉 C 的「真正工作量」）
C 上一轮提「把 kinematicGlued+删除提到 present∨lost 都跑，好删 present-frozen」，并称这是整合真正工作量 + 唯一碰 neighbor 时序（要加 coast 断言）。**此处 G 挡下，成立：**

- **present-frozen 的 FP 是 floor 误火 → 归压制轴，不归删除轴。** byte-frozen 但固件在发→stillbox 攒→floor fire；**压制轴（GhostActive→StillBoxSec=0, :1077）已中和**，不需删。lost_fall 删除是给 **lost 残迹**（触发 lost-fall 告警）用的，**present 轨不走 lost-fall**。C 混了 FP 的两个出口（floor 误火 vs lost 残迹告警）。
- **删 present 轨 = 固件下一帧重报 = churn** —— 全仓反复躲的坑（[[teleport_interference_purge_mechanism]] 三闸、witness 用 ResetStillBox 不 EvictTrack、[[cd2b_present_coast_evict_purge]]）。且 C 早轮自定「不删轨，靠 ghost score 压 stillbox」——#5 又违背它。d5f7 89/90min present-frozen case 是**几何 SBed + 压制**治的，不是删的。

→ **净效果**：删除保持 lost-only（gate `presenceCoastMs`, :1576/:1585）；present-frozen 交压制轴；C 说的「present 分支扩展 + neighbor coast 断言」**整块消解，不需要**——风险最大、唯一碰 neighbor 时序那块，是被「压制已覆盖 present」消掉的，不是要去啃。**整合范围缩到只动 lost 分支：FN 更安（HOLD veto）、FP 压制轴管、不碰 neighbor、不引 churn。**

### 4.6 整合顺手关掉的现存不一致
1. exitCoupled ①②③ 无 R 闸（:2034/2040/2046）→ 统一加 `kinematicGlued==GLUED`，多一道 FN 保护（刚出生要走开的真人不再靠 pose 单独兜）。
2. Gap1（四.4）随 GLUED 替换自然关闭。
3. mirror 因锚点不适用**保留独立几何路径**（4.2 第二支）。

## 五、C 净判 + 待架构师裁

**采纳 G 版**（三档布尔 + 3-tick 棘轮 + HOLD 存疑 + present-frozen 归压制轴），去 C 的除法/oracle/present 分支——可辩护性严格占优、守单调、范围最小、不碰 neighbor、不引 churn。

**本 case 两版一致**（G 验）：tid0 离锚点恒 0 ≤50→棘轮坐实→幽灵；tid1 142→206（5s 破 200）→walkout real。G 版少一除法、一 oracle 阈。

**待裁两项**（present 分支那项已消解，删）：
1. `kinematicGlued` 三态抽取 + lost_fall 4.2 重构（合并条件2/条件3-R，只动 lost 分支，mirror 留几何轴，branch② GLUED 替换 membership）——**范围与优先级**；
2. 逐路「锚点」语义确认（3.2：split/interfer 真外部锚点，门距=net-from-birth 仅因不动才无害）——**A 落码时别让三路误当同义**。

**C 不落码**，等架构师裁范围后再出验收规格（§69 模式）。
