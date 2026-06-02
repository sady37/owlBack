# Room Belief State Machine — 房间级人状态信念估计器（设计文档）

**定位**：用一层**显式概率信念估计**替代当前 gate-list 规则引擎，治本解决跌倒误报"打地鼠"。
独立 PR，分阶段 shadow-mode 迁移。本文件自包含。

**关系**：这是 [[bed_bayesian_review.md]] 的**推广** —— 床占用（单变量 InBed）的贝叶斯 scorer
已在生产验证可行（log-odds 递归 + 证据簇 LR + 非对称阈值 + maintain 区）。本文档把同一范式
从"床的二态"提升到"房间里一个人的多态"。**范式已被证明，本 PR 只是把它扩到 person-state。**

`revision: 1` ・ `created: 2026-05-31`

---

## 1 为什么做：本质问题不是管道，是缺一层世界状态估计

### 1.1 打地鼠的真因（3 层）

| # | 症状层 | 真因 |
|---|---|---|
| 1 | person_silent 在 9h 前离开的 ghost 上报倒地 | **没有"人可能已离开"这个显式状态** —— 系统世界观里只有"在/不在、动/不动"，没有"去了我看不见的地方（床/盲区/门外）" |
| 2 | 每修一个误报，缝换个地方再漏 | 判定是 **gate-list（条件清单）不是状态机**：`if A or B or C ...`，OR 越加越长，**构造上开放、永不封闭** |
| 3 | 证据停了系统却继续推断 | **"证据缺失/stale"没被建模** —— track 冻住后 census 假设"人还在原地不动"硬撑计时器，9h 后误报 |

**根因 = 缺一层"对有限隐状态的显式信念估计"**，夹在原始多源观测和决策之间。

### 1.2 track 消失的二义性（核心）

"track 消失"在数据上**无法区分**这几种原因，但语义完全相反：

| track 消失的真实原因 | 正确动作 |
|---|---|
| 人走出门 / 进了雷达盲区（床、09E7 覆盖区） | **不报**（人没事，离开了） |
| 人原地静止，雷达失去多普勒丢目标 | **候选倒地**（person_silent 存在的唯一目的） |
| ghost 散了 | 不报 |

今天的系统对所有"消失"用同一套 lost_fall pending 处理 → 必然在某一类上误报或漏报。
**要解二义性，必须知道"在哪消失的、消失前什么状态、别的源怎么说"** —— 这正是 belief 向量
天然携带的信息。

---

## 2 数学模型（直接复用床贝叶斯，升维到多状态）

### 2.1 床模型是这个模型的一维特例

[[bed_bayesian_review.md]] 的床 scorer：隐状态 = {InBed, ¬InBed}（2 态），
log-odds 递归 `L_t = clamp(L_{t-1} + ΔL_t)`，证据簇 max-merge + 簇间累加。

本模型：隐状态 = N 个 person-state（见 §3），belief 是 **Δ(S) 上的概率向量** b（不是单个
log-odds 标量）。床模型 = 本模型在 S={Bed-Lying, ¬Bed} 上的投影。**两者数学同源**，床的
LR 表可直接作为本模型 `P(o|s)` 中"sleepad → Bed-Lying"那几列的标定来源。

### 2.2 belief 向量与两步更新

```
b_t ∈ Δ(S)          # 概率向量，Σ b_t[s] = 1，永远是分布不是确定状态

每帧两步（= 贝叶斯滤波 / HMM forward 算法）：

① 时间推进（prediction）:   b ← A · b
   A[i][j] = P(s_t=j | s_{t-1}=i)，转移矩阵，编码物理常识
   （Bed-Lying 不瞬变 Stand 要经 Transition；Fallen 不自愈成 Walk）

② 观测更新（correction）:    b ← normalize( diag(P(o|s)) · b )
   每条观测 o 按"状态 s 下看到 o 的似然 P(o|s)"重加权每个状态分量
```

### 2.3 缺证据 = 命门（这一条治掉整类 bug）

| 场景 | 今天 | belief 模型 |
|---|---|---|
| track 冻住，无新可信观测 | census 假设"人还在原地不动"，计时器照走 → 9h 误报 | 该源的 `P(o|s) = I`（单位阵），**不更新观测**；b 只由 A + **其他源** 维持 |

9h case 在 belief 模型里的演化：
1. D523 track 冻住 → D523 这路停止更新，不强推任何状态
2. belief 是**房间/suite 级**，还有别的源：sleepad 报 InBed、09E7 在床区见 vital
3. A 里"Stand/Walk → 经床区边界/门 → 离开 D523 视野"有概率
4. b 自然演化为 `P(Left/Bed-elsewhere) ↑，P(Fallen-in-D523) → ≈0`
5. **person_silent 永不 fire，因为它问 P(Fallen)，而那概率从没起来**

"证据停了却拿旧假设当真相" 这一整类 bug **构造上消失**。

### 2.4 决策层（非对称代价，复用床的思路）

belief 不直接 = 报警。决策 = belief × 代价矩阵：

```
报 Fall 当且仅当  P(Floor-Fallen) > θ_fire   （θ 高，FN cost ≫ FP cost）
b 摊平（max 分量 < θ_uncertain）→ 不报 / 升级人工复核，不瞎猜
```

跌倒 FN（漏报真跌倒）代价 ≫ FP（误报），所以阈值非对称 —— 与床模型
`P>0.70 InBed / P<0.50 LeftBed / 中间维持` 同构。

---

## 3 隐状态集 S（一个房间一个人）

```
S0  Empty             房内无人
S1  Bed-Lying         在床躺（含睡眠）
S2  Bed-Restless      床上翻动/坐起
S3  Sit               坐（椅/沙发/马桶）
S4  Stand-Walk        站立或行走
S5  Floor-Fallen      倒地  ← 唯一 fire Fall 的目标态
S6  Transition        过渡中（状态间，吸收瞬时不确定）
S7  Left-via-Door     从门离开（→ 收敛回 Empty）
S8  Artifact          ghost / 反射 / 设备伪迹
```

设计要点：
- **S0 Empty + S7 Left 是第一类公民** —— 今天系统缺的就是这两个，9h bug 的直接原因。
- **位置语义不进 S，由 grid 提供**（在 bed-polygon ∩ / 在 Enter 区 ∩ / 开阔地板）作为
  `P(o|s)` 的条件。这依赖 ⑤ 已加载的 layout 几何（[[fall_fp_roots_and_todo]] 根因 A）。
- **多人 / 多 device**：v1 先做"一房一人"（与今天 census 一致）；多人留 per-track belief +
  suite 级聚合（[[cardagg_sensor_split.md]] census 已是 suiteID-key，天然衔接）。

---

## 4 观测模型 P(o|s)（4 个传感器源）

每个源给出一组 `P(observation | state)` 的标定行。缺该源 → 该行 = I（不更新，§2.3）。

| 源 | 观测 o | 主要区分的状态 | 标定来源 |
|---|---|---|---|
| **radar pose+kinematics** | firmware pose（1walk/2suspect/4stand/5fall/6lie）+ Δz + 位移 + cell 语义 | S1/S3/S4/S5 + S8(运动学矛盾→ghost) | [[fall_rules_three_classes]] 现行阈值转似然 |
| **sleepad** | InBed event + HR/RR/body_move/turn_over | S1/S2 vs 其余 | **直接取 [[bed_bayesian_review.md]] §2.2 LR 表** |
| **firmware fall qualification** | pose2→5 升级（30~90s 自带） | S5 高似然 | [[firmware_fall_qualification]] Option C |
| **time-context prior** | 时段（夜/昼）+ 房型（bathroom 标准 StandingContinuousMin） | 调 prior，非硬观测 | [[fall_fp_roots_and_todo]] ④ |

关键：**派生信号（WeakBio/趋势/聚合）禁进**（[[feedback_no_dynamic_threshold_modulation]] 铁律）——
它们不是独立观测，进了会双重计数。belief 只吃**原始物理观测**。

---

## 5 转移矩阵 A（白盒 —— 每条目可读可改）

A[i][j] = 从状态 i 到 j 的一步转移概率。编码物理/生理常识，**没有一个数字不可解释**：

```
            →Empty →BedLy →BedRe →Sit →Stand →Fallen →Trans →Left →Artifact
Empty        高     0      0     低   低     ~0     低    0     低
Bed-Lying    0      高     中    0    0      低★   中    0     0
Stand-Walk   0      0      0     中   高     中★   中    中    低
Floor-Fallen 0      0      0     0    低     高     低    0     0    ← 不自愈
Left-via-Door 高    0      0     0    0      0      0     高    0    ← 收敛 Empty
...
★ = safety-critical 条目：进 Fallen 的概率。调这一个数 = 调灵敏度，集中、可审计。
```

**HSMM 扩展（治"久态"误报）**：标准 HMM 转移是几何分布（隐含"每帧固定概率离开"）。
现行 bathroom 10min / bedroom 2h 硬阈值 = HSMM 显式状态时长的**零方差退化版**。
A 升级为 HSMM（每状态带 duration 分布）可把硬阈值变成软的"停留越久越可疑"，更鲁棒。
v1 先用 HMM + 硬 duration gate 复现现状，v2 再上 HSMM。

---

## 5.5 多实体耦合 — 元素之间的转移关系

§5 的 A 只管**单个实体自己**从状态 i 到 j（时间转移）。但真实约束跨实体：
"roomB 有人 → roomA risk 降"、"Bed-Enter → 间接进证 Room 没人倒地"。
这类**实体间关系**单实体 HMM 表达不了 —— 它的世界里没有 roomB / bed 这些**别的变量**。
这正是 gate-list 漏风最多处（跨实体推断全靠零散 gate，§见 fall_rule_inventory 冲突 1/3/4）。

### 5.5.1 两种"转移关系"必须分清

| 类型 | 例子 | 数学对象 | 单实体 A 能否表达 |
|---|---|---|---|
| **时间转移**（同实体，t→t+1） | roomA 的人 Stand→Walk | 转移矩阵 A（§5） | ✓ 能 |
| **实体耦合**（同时刻，实体间） | roomB 有人→roomA risk 降；bed 有人→room 地板无人 | **约束 / 守恒项** | ✗ **不能** |

你的两个例子都是第二种。**关键设计：耦合不进单实体的 A**（A 永远只管自己，否则状态空间
指数爆炸），**而进每帧第③步的约束投影算子 C**（见 5.5.4）。复杂度隔离在一个明确的约束层。

### 5.5.2 弱耦合 — 邻居 belief 当一条观测（先做）

最轻量、不改单实体引擎：roomB 的 belief 输出，作 roomA 的一条 `ObsNeighbor` 观测，
进 roomA 的 P(o|s)：

```
P( roomB_有人 | roomA=Empty )   = 高    ← 人在 B，A 大概率空
P( roomB_有人 | roomA=Fallen )  = 低    ← 人在 B，A 不太可能有人倒地
```

这就表达"roomB 有人 → roomA risk 降"：roomB 占用信念通过似然项**压低 roomA 的 P(Fallen)**。
在 [[belief_input_normalization.md]] schema 里就是加一条 ObsKind：

```
ObsNeighbor{ Source: roomB, Value: P(roomB 有人), Conf: roomB belief 的确定度 }
```

优点：零改引擎，roomA 照常单实体 forward，多吃一条邻居观测。
缺点：单向快照耦合，不保证全局守恒（可能 A、B 同时被推成有人）。**v1 先做这个**，已能修一大类跨房误报。

### 5.5.3 中耦合 — suite 级人数守恒（治本你的两个例子）

你的例子本质是**守恒律**：一个 suite 就这么几个人，A 多一个、别处必少一个。
把单实体 belief 升成 **suite 级联合分布**，加守恒约束：

```
suite 状态 = (roomA, roomB, bed, ...) 各 belief 的联合
约束: Σ over 所有 room/bed 的 P(有人) ≈ 已知总人数 N

→ Bed-Enter 不只更新 bed，它通过守恒把 room 地板的 P(有人) 拉低
  （人进床，地板就少一个 → roomFloor 的 P(Fallen) 自动降）
```

这直接表达"**Bed Enter/Exit 间接进证 Room 状态**"：不是 room 自己观测到，而是
**守恒约束从 bed 传导过来**。也是 9h bug 的治本：D523"地板有人倒地"的信念，被
"sleepad 报床上有人 + 总人数=1"的守恒**强行压下去**（人就一个、在床上、地板不可能还有一个倒着）。

**宿主天然就是 census**：census 已是 suiteID-key 的人数管理（[[cardagg_sensor_split.md]]）——
把它从"计数器"升成"suite 联合 belief 的归一化器"，不新建实体。

### 5.5.4 统一表达 — 每帧三步（而非两步）

```
单实体（§2 今天的 HMM）:
  ① b ← A·b              时间转移
  ② b ← diag(P(o|s))·b   观测更新

多实体（加耦合）:
  ① 各实体 b_i ← A_i·b_i        各自时间转移（不变）
  ② 各实体 b_i ← diag(P(o|s))·b_i  各自观测更新（不变）
  ③ 联合归一化: 对 suite 内所有 b_i 施加约束投影算子 C   ← 新增，表达耦合
```

第③步的 **C = 约束投影算子**，就是"元素间转移关系"的数学落点：
- **弱耦合版（5.5.2）**：C = 把邻居 belief 当观测，等价于第②步多喂一条 ObsNeighbor
- **中耦合版（5.5.3）**：C = 人数守恒的投影，在 census 层归一化各 b_i

注意：**C 不改单实体 A**（A 永远只管自己），保持单实体引擎简洁不爆炸，把跨实体复杂度
隔离在 C 这一层。

### 5.5.5 对照你的两个例子

| 你的话 | 模型表达 | 哪一层 |
|---|---|---|
| Bed Enter/Exit 间接进证 Room 状态 | bed belief → 人数守恒 C → 压低 room 地板 P(有人/Fallen) | 中耦合（census 守恒） |
| roomB 有人 → roomA risk 降 | roomB 占用作 roomA 的 ObsNeighbor，似然压低 P(Fallen) | 弱耦合（先做） |

两个都不靠"再加一条 gate"，而是同一个守恒/耦合机制的不同投影 —— **这就是 belief 封死缝的关键：
跨实体推断有了统一的数学宿主（约束层 C），不再散落成无穷无尽的 gate**（fall_rule_inventory
冲突 1/3/4 全是"跨实体推断靠零散 gate + 人工同步"，C 是它们的统一解）。

### 5.5.6 强耦合 — 完整 DBN（远期，不现在做）

每个实体一个节点，节点间画有向边（bed→room、room↔room、door→两侧 room），边上是条件
概率表 = Murphy DBN 完全体。完备但工程重，留作 HSMM 之后的 v3。**v1 只做弱耦合（5.5.2），
v2 上中耦合守恒（5.5.3）。**

---

## 6 与现状一致 = 可证（用户硬约束）

| 约束 | 做法 |
|---|---|
| 现行每条硬 gate = 概率模型确定性角点 | A、P(o|s) 条目先取 **0/1**，把新模型**初始化到精确复现今天判定**的角点 |
| 逐步标定 | 再把 0/1 换成标定概率（似然反推自 [[fall_rules_three_classes]] / 床 LR 表） |
| "一致" = 与**已验证正确结果**一致 | 用 `doc/cases/` 当**回归 oracle**（见 §7），不是与现行每行代码字面一致 —— gate-list 缝处本就自相矛盾，字面一致会把 bug 搬过来 |
| 迁移安全 | **shadow mode**：旁路并行跑、记录新旧分歧、逐条 triage、全是"修真 bug"才 cutover（对齐 [[v2_cutover_lessons]] / producer-first） |

---

## 7 回归 oracle（已有 case fixture）

`doc/cases/` 现成素材，每个有人工验证的正确标签：

| fixture | 期望 belief 行为 | 验证 |
|---|---|---|
| **John.Y 9h person_silent**（D523 无床 + census ghost） | P(Fallen) 从不起来；P(Left-to-bed) 升 | **必须修对**（今天误报） |
| **CABB lost_track**（浴室门口丢 track） | track 丢在 Enter 区 → P(Left) 升，P(Fallen) 低 | **必须修对**（今天误报） |
| cabb-fall-A/B/C（真跌倒） | P(Fallen) > θ_fire | **必须仍对**（今天判对） |
| d5f7 / cabb ghost 段 | P(Artifact) 主导 | **必须仍对** |
| bayes-test-0524 | 床贝叶斯回归 | 床投影一致性 |

最小闭环目标：**复现现状 → 只修 9h+CABB → 其余零回归**。证明这一个闭环成立，
即同时证明"可一致 + 可治本"，再谈推广。

---

## 8 诚实边界（必须分清，否则新框架在矩阵条目上继续打地鼠）

- **完备 = 程序完备**（任意 present/absent/stale 组合都有定义后验，漏风消失），
  **≠ 正确**（S 漏真实模式 / P(o|s) 标错仍会自信地错）。但失效模式变成"补一个状态 / 调一个
  条目" —— 局部、可观测、可回归，**不再是缝里悄悄漏**。
- **不创造可观测性**：雷达分不开"地上不动 vs 床上不动"时矩阵变不出信息 —— 但会输出**恰当的
  不确定度**让决策不报 / 升级人工，而不是瞎猜误报。
- **工作量不消失，是搬家**：从"枚举规则组合，随条件指数爆炸" → "枚举有限状态 + 每传感器
  P(o|s)，随状态**线性**"。

---

## 9 落地第一步（本 PR scope）

1. **设计冻结**：定 S（§3，~9 态）+ 4 个 P(o|s)（§4）+ 一个 A（§5），产出**现行 gate→矩阵条目
   对照表**（每条硬 gate 映射到哪些 0/1 角点条目）。
2. **核心实现**：`internal/roomengine/belief/` 新包 —— belief 向量 + forward 更新 + 缺证据 I 处理。
   复用床 scorer 的 log-odds/LR 工具（[[bed_bayesian_review.md]] §4 实现位置）。
3. **shadow 接线**：在 publishTrackStatuses 旁路跑 belief，**只 log 不 fire**，记录与现行 fall
   判定的分歧。
4. **3-case 闭环**：CABB + John.Y + 一个真跌倒，跑"复现现状→只修 9h/CABB→零回归"。
   注：John.Y 9h 的治本依赖 §5.5.3 中耦合守恒（"床上有人+总人数=1→地板 P(Fallen)=0"），
   v1 最小闭环可先用 §5.5.2 弱耦合（sleepad InBed 作 ObsNeighbor 压低 P(Fallen)）近似，
   验证方向成立后 v2 再上守恒。
5. 闭环过了再谈：扩 case 集、HSMM、中耦合守恒（§5.5.3）、多人、cutover。

**不在本 PR**：HSMM、中耦合守恒（§5.5.3）、强耦合 DBN（§5.5.6）、多人 belief、cutover（删
gate-list）。本 PR 单实体 belief + §5.5.2 弱耦合，先证一个最小闭环。

---

## 10 参考锚点（成熟范式，非自创）

- Thrun/Burgard/Fox, *Probabilistic Robotics* (2005) — belief / Bayes filter 逐字教科书
- Rabiner, *HMM tutorial* (1989) — forward 算法
- Murphy, *Dynamic Bayesian Networks* (2002) — 多变量扩展
- HSMM（explicit-duration HMM）— 治"久态"误报，现行硬阈值是其零方差退化版

商用跌倒检测多走深度学习黑箱；本方案**故意选不时髦但满足"可解释 + 完备 + 能注入人类规则"
约束的贝叶斯支** —— 与床 scorer 同源，范式已在生产验证。

---

## 11 下次会话入口

读本文件 + [[bed_bayesian_review.md]]（同源范式）+ [[fall_fp_roots_and_todo]]（根因+两 fixture）。
开工 = §9 第 1 步：写 gate→矩阵条目对照表。关联记忆 [[belief_state_rule_engine_reframe]]。
