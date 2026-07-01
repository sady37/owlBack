# Ghost 反射伪迹检测：原理、适用边界与判据迁移

状态：**设计/原理底稿**。用途：(1) 厘清当前 ghost 反射检测的四条防线各自的物理原理；(2) 从数学上证明 mirror_detect 的**适用边界**——它只对「单一稳定镜面」有效，对「离散多反射体」结构性失效；(3) 给出治理方向：把判据权重从「几何对称」迁到「共生 + 生命体征」这些**对几何不敏感**的轴上。

关联：[[mirror_detect_single_mirror_boundary]]（本文核心结论的记忆索引）、[[reflection_cells_areareflector_static_phaseA]]（mirror_detect 升 AreaDeny / static 6-29 已放 Phase B 升 AreaReflector）、[[teleport_interference_purge_mechanism]]（≥200 瞬移 PURGE）、[[realness_axis_redefined_real_vs_mirror]]（realness 轴 = Real vs Mirror，主职 N_r 排 ghost）、[[realness_never_vetoes_fall]]（铁律：realness 绝不否决 fall）、[[two_radar_mirror_gate_samedevice]]（镜像门控同设备）、[[gatelist_retired_ghost_single_source_census]]（census 单源 PReal）、[[cell_dbn_timescales_stillbox_single_source]]（still-box 单源）、[[radar_hr_rr_bed_enter_gated]]（radar vital gated）。

---

## 第一部分 · 物理本质

雷达伪迹（ghost）的物理根因是**多径反射**：雷达波打到金属/玻璃/镜面后弹一次再返回，固件把这条二次路径的回波当成一个**额外的人**，生成一条假 track。这条假 track 的位置，就是真人关于那面「镜子」的**镜像点**。

所以判 ghost 的核心命题始终是同一句：**这两条 track 是不是一对镜像？** 当前系统有四条防线在从不同角度回答它。

| 防线 | 文件 | 抓什么形态的 ghost | 依赖 | 接线状态 (6-29) |
|---|---|---|---|---|
| **Mirror 检测** | [mirror_detect.go](../wisefido-sensor/internal/roomengine/mirror_detect.go) | 单一稳定镜面 + 连续运动 | 几何对称（轴 + 速率同步） | ✅ 生效，命中升 AreaDeny（其 ghost 有真人伴随 → realness 可抓） |
| **IsReflection** | [census.go](../wisefido-sensor/internal/roomengine/adapter/census.go) `reflectSep()` | 出生瞬间、墙外镜像 | 几何（连线穿墙） | ✅ 生效，喂 realness 观测 |
| **Teleport purge** | [teleport_interference.go](../wisefido-sensor/internal/roomengine/teleport_interference.go) | 跳变 / 瞬移 | 运动学（人类速度上限） | ✅ **生效，真 PURGE 删轨**（track_manager.go:1158/1200/1329，无 flag 默认开） |
| **Static reflector** | [static_reflector.go](../wisefido-sensor/internal/roomengine/static_reflector.go) | 静止金属反射体 | 持久性（出生位 + 久驻 + 近墙） | ✅ **生效（6-29 放 Phase B）**：scan 每帧跑（track_manager.go:1602），三签名累计跨独立 episode ≥3（`StaticReflectorPromoteThreshold`）→ `MarkStaticReflector` 升 cell→**AreaReflector**（拿 floor 整片豁免）；demote 由 cell_learning hasRealActivity 兜底 |
| **Split ghost** | [split_ghost.go](../wisefido-sensor/internal/roomengine/split_ghost.go) | 单人走墙/干扰区 firmware 分裂的并存幽灵 | 时机（出生贴 present 邻轨 + lost 窗末打分） | ✅ **生效，已接 track_manager**（出生 `stampSplitObservingIfCoexist` / 每帧 `updateSplitWalkLatch` / lost `maybePurgeSplitGhost`，hard PURGE 复用 teleport interferenceSuppress） |

---

## 第二部分 · Mirror 检测的原理（轴对称）

[mirror_detect.go](../wisefido-sensor/internal/roomengine/mirror_detect.go) 的 `evaluateMirrorPair()` 取两条 track 的 5 帧配对样本，验证三个「镜像不变量」。其几何灵魂是**轴对称**：若 A、B 真是一对镜像，必存在一条稳定的**对称轴（镜面线）**，A、B 互为关于它的反射。

单一固定镜面 M 是一个**等距仿射变换** $R_M$（反射矩阵 + 平移）：

$$\text{ghost} = R_M \cdot \text{real}$$

这个变换有两个恒等性质，正是三不变量的来源：

| 不变量 | 几何含义（轴对称的必然推论） | 阈值 |
|---|---|---|
| **I1 方向恒定** | 镜面线固定 ⇒ 连线向量 $(B-A)$ 的**方向**必恒定 | 角度圆周标准差 ≤ 15° (`mirrorVAngleStdRad=0.26`) |
| **I2 中点共线 + 正交** | A、B 关于轴对称 ⇒ **中点必落在轴上**，且 $(B-A) \perp$ 轴 | 中点 RMSE ≤ 20cm；正交 cos ≤ 0.20 |
| **I3 同步等速** | 等距变换 ⇒ $\lvert v_{\text{ghost}}\rvert = \lvert v_{\text{real}}\rvert$（速率守恒、方向翻转） | 速率差比 ≤ 40% (`mirrorSpeedRatioMax`) |

> ⚠️ 关键澄清：I3 抓的是 **$\lvert V_{\text{real}}\rvert = \lvert V_{\text{ghost}}\rvert$（速率相等）**，**不是矢量相等**——方向被轴镜像翻转了。

**判定流程**：用 PCA 把所有帧中点拟合成一条直线（`fitLineDir`），这条线就是估计出的**对称轴**；再验 $(B-A) \perp$ 轴（I2 正交）+ 方向恒定（I1）+ 速率同步（I3）。三条全过 ⇒ 镜像对。

**谁是 ghost**：tiebreaker 用物理——多径路径更长 ⇒ **距雷达更远者是 ghost**（5 帧平均距离）。命中后算反射弹射点累加 grid，cell 命中 ≥3 次升格 `AreaDeny`（见 [[reflection_cells_areareflector_static_phaseA]]）。

---

## 第三部分 · 适用边界：单镜面能抓、多镜面必崩（数学根因）

这是本文档的核心结论。**mirror_detect 的三不变量全部建立在「单一等距变换 $R_M$」假设上。一旦镜面切换，变换本身在变，所有不变量结构性崩塌——不是阈值不够好，是模型假设被违反，调参原则上无解。**

### 3.1 场景：离散多反射体

现实里反射体常常**不是一面连续墙镜**，而是散布房间各处的离散金属物件 $M_1, M_2, M_3 \dots$。雷达每秒一帧采样，每帧激活的反射体可能不同：

```
t   : 真人经 M1 反射 → ghost 出现在 address_1
t+1 : 真人经 M1 反射 → ghost 出现在 address_2
t+1': 真人经 M3 反射 → ghost 出现在 address_3   ← 反射体切换
```

观测到的 ghost track 被关联成一条，但它实际在**不同反射体的镜像点之间离散跳变**。帧间位移 distance 时大时小（$d_{3,1}$ vs $d_{3,2}$ 无规律），既不连续，也无稳定轴 → 看起来像「随机游走」。

### 3.2 为什么 V_real == V_ghost 也救不了

单镜面：$\text{ghost}_t = R_M \cdot \text{real}_t$，变换恒定，所以速率守恒（I3 成立）。

多镜面：

$$\text{ghost}_t = R_{M_1}\cdot\text{real}_t, \qquad \text{ghost}_{t+1} = R_{M_3}\cdot\text{real}_{t+1}$$

ghost 的**视速度** $= R_{M_3}\cdot\text{real}_{t+1} - R_{M_1}\cdot\text{real}_t$。因为 $R_{M_1} \neq R_{M_3}$，这个差**不等于任何** $R\cdot v_{\text{real}}$。

> **结构性崩塌**：变换本身在帧间改变 ⇒ 等距性丢失 ⇒ 速率不再守恒（I3 废）、轴不再存在（I2 废）、方向不再恒定（I1 废）。三不变量同时失效，且是**确定性失效，不是噪声**。

极端情形：real 几乎没动，但 $M_1 \to M_3$ 切换让 ghost 从房间一头跳到另一头。

**∴ 任何建立在「连续性 / 对称轴 / 速率同步」上的几何判据，在多镜面下原则上无解。** 离散随机游走的 ghost，纯几何上无法与一个乱走的真人区分。

---

## 第四部分 · 治理方向：迁到对几何不敏感的不变量

既然几何轴失效，就必须找**不依赖单一变换、不依赖轨迹连续**的不变量。物理上 ghost 仍有三个逃不掉的性质：

### ① 瞬移（随机游走的尖峰那一面）—— 几何的破绽
镜面切换时 ghost 单帧位移爆表、超人类极限，命中 [[teleport_interference_purge_mechanism]] 的 ≥200cm PURGE。它不要求轴、不要求同步，只要求「人不可能这么快」。
- **互补关系**：mirror_detect 收**平滑连续**的镜面 ghost；teleport purge 收**跳变**的多镜面 ghost。离散跳变恰是后者的破绽。
- ⚠️ **盲区**：当 $M_1$、$M_3$ 镜像点恰好接近时，那一帧位移很小，不瞬移、不被 purge。所以只能收「跳得远」的，收不全。

### ② 共生性（co-existence）—— 镜面无关的本职判据
ghost 的存在**绑定真人**：real 在则 ghost 可能在，real 消失则 ghost 必消失，且 ghost 永远凑不齐独立行人的「出生→行走→离场」闭环。这是 census（[[gatelist_retired_ghost_single_source_census]]、[[realness_axis_redefined_real_vs_mirror]]）该承担的本职——
> **它不问「你长得对不对称」，问「你和谁共生、你能不能独立存在」。**

多镜面随机游走的 ghost，几何上与乱走真人无法区分，但在**共生轴**上仍露馅：它从不独自出现、从不有自己的独立存活闭环。

### ③ 生命体征（vital）—— 最硬、不碰运动
反射回波是真人回波的二次路径，**没有独立心跳 / 呼吸**。真人有 HR/RR，ghost 没有。这条完全不碰几何、不碰运动，是对「随机游走 ghost」最干净的判据。已接 radar HR/RR（[[radar_hr_rr_bed_enter_gated]]，enterBed gated），值得把「持续测不到 vital 的 co-existing track」纳入 ghost 一票。

---

## 第五部分 · 架构结论

把 mirror_detect 的定位摆正：**它是「单一稳定镜面」专用的强判据，不该指望它覆盖离散多反射体**。多镜面是它设计假设外的场景，期待它处理是错配。正确的分层：

```
                          ghost 判定
        ┌──────────────┬──────────────┬──────────────┐
   mirror_detect      teleport      co-existence      vital
   (单镜面/连续)       (跳变尖峰)     (共生/独立性)     (无心跳)
   几何强判据          几何破绽       几何无关          运动无关
   适用边界=单镜面     收离散跳的      收随机游走的       收一切静态ghost
        │               ↑互补          ↑本职(应增重)     ↑兜底(应接入)
```

**核心洞察**：随机游走的多镜面 ghost，在**几何轴**上确实无解，但在**共生轴**和**生命体征轴**上无所遁形。治理方向不是给 mirror_detect 加更多几何 trick，而是**把 census 判据权重从「几何对称（IsReflection）」迁到「共生 + vital」**。

### 5.1 Fall 风险视角的安全网
即便某个 ghost 几何上排不掉，按铁律 [[realness_never_vetoes_fall]] **realness 永不否决 fall**，它最多影响占用人数 $N_r$，不会压掉真摔。真正的危害只在「≥2 track 时 ghost 让系统误判房里有 2 人、从而激进否决真摔」——而那条否决路径本就该用 **co-existence 置信度**而非几何（[[two_radar_mirror_gate_samedevice]]）。所以把判据迁到共生轴，连这个残余风险也一并解了。

---

## 第六部分 · per-track 空间先验 ghost（birth-provenance）

第四部分的三条不变量都在「房级 / 几何级」回答 ghost。但有一类伪迹两条都抓不到：**空房里的孤迹干扰**（帘动、风扇杂波）。

- **几何（mirror_detect）抓不到**：它常是离散多反射体的随机游走，第三部分已证结构性无解。
- **共生（co-existence）抓不到**：co-existence realness 只在 **≥2 轨互为镜像**时判别；一条**孤迹** phantom（无真人陪它）PReal ≡ 1，共生轴上看它就是个独立的人（[[realness_axis_redefined_real_vs_mirror]]、[[engine_aggregation_floor_gate_f1_occupancy]]）。

所以需要第三条轴：**出生地空间先验**——一条 track **诞生在已知干扰 cell** 里，本身就是伪迹信号，与共生、与几何都无关。

### 6.1 要解决的 FN：interfer 区当前被整片豁免

当前 [floor.go:69-72](../wisefido-sensor/internal/roomengine/belief/floor.go) 对 `areaInterfer` 无条件 `return false`——**整个干扰区的 floor 网全关**。这压住了帘动幻影，但**抬高了 FN**：一个真人走进干扰区摔倒，floor 网也对他关着 → 漏报。

根因是把豁免挂在**区域**上（位置维度），而风险其实挂在**是不是伪迹**上（身份维度）。

### 6.2 修法：判别上移到出生时刻，用 still-box 入口做唯一闸

不在 floor 层按区域豁免，而在出生时刻（track_manager，`BirthPos` 与 `grid.CellAt()` 都现成）给伪迹盖一个 **provisional ghost** → **不进 still-box** + 掉出 $N_r$。然后 floor 层的 interfer 豁免**整条删掉**，interfer 退回 default 12min。

- born-elsewhere 真人在干扰区摔 → ghost score 低 → still-box 正常累 → **floor 12min 网恢复**（解 FN）。
- born-in-interfer 帘动幻影 → ghost → StillSec 恒 0 → floor 无从 fire（保 FP 抑制），**不需要任何区域豁免**。

盖戳条件（**三闸全满足**才盖）：

1. `CellAt(BirthPos).AreaType == areaInterfer`
2. **fresh 出生，无 lid 继承**（排 churn 原地重生，[[logicid_unified_census_refs_tm]] 全局 lid）。
3. `minDist(BirthPos, 其它非ghost真轨) ≥ 120cm`（隔离）。

撤销（**任一触发即撤 ghost、still-box 恢复**）：

- 走出 interfer cell / 位移超 confinement 半径（表现人类位移）。
- 出现**摔倒前兆**（pose ∈ fall / Δz 骤降）——这是把唯一窄 FN 口堵死的安全网。

> **三闸的物理**：真·干扰幻影的位置只由干扰源决定、与人在哪无关，所以它没理由诞生在真人旁边——隔离 = 独立伪迹签名，贴人（<120cm）= 暧昧、缩手（FN-safe）。
> **120 vs 50**：同人保号是 ≤50cm（[[track_identity_logicid_ghost_track2_scope]]）；120 故意更宽，给真人周围留更大的「别压」缓冲带。"其它轨"须限定为**非 ghost 真轨**，否则两个挨近的帘动幻影会互相打掩护。
> 残余 FN 被收窄到几乎不存在的一格：真人恰在帘边、固件干净重生把他原地新生、120cm 内无其它真轨、且在走出去之前就摔——四条同时成立才漏；而帘动幻影满足前三条、永远走不出去，稳被压住。

### 6.3 为什么是软压制、不能 purge（决定性）

teleport 能硬 PURGE，因为它有**历史**：一条轨跳 ≥200cm，有清晰 before/after，可在删除那刻用 pose 闸判定"这不是摔"、**确定**了再删。

interfer-born 的触发是**出生**——那一刻**没有 before**，拿不到出生时确定性，只能**边看边撤**。这逼出软路径：

> purge 删了就没了。若那条 born-in-interfer 其实是 churn 原地重生的**真人、且就摔在区里不走动**，purge 会反复删他 → still 永不累 → **FN**。软压制留着轨，才能在 pose 演化成 fall 时撤销 ghost、恢复 still-box，把这个口堵上。purge 做不到，删了就没 pose 可观察。

confidence 也不对等：teleport = 运动学不可能（近确定），interfer-born = 空间先验（概率）。**低确定性 + 出生无历史 ⇒ 结构上就不能是 purge**。

### 6.4 五检测器分工（动作 / 确定性 / 归属）

| 触发 | 动作 | 确定性 | 归属文件 |
|---|---|---|---|
| **teleport 跳变** | hard **PURGE** 删轨 | 运动学（确定）| `teleport_interference.go` |
| **interfer 出生** | **soft 压制 + 可撤销**（不进 still-box / 掉 $N_r$）| 空间先验（概率）| `static_reflector.go`（fast-path）|
| **static 持久** | **cell 学习 → AreaReflector**（拿 floor 整片豁免；非 AreaDeny——孤迹无人时 realness 抓不到）| 持久性（慢确认）| `static_reflector.go` |
| **split 分裂** | **soft 压制 `SplitGhostSinceMs`**（vote→spliter 静止门→group→赖锚点者零 still-box，不删轨）| 时机（概率）| `split_ghost.go`（vote-based，已重写）|
| **ExitRoom 残迹** | hard **delete**（离场耦合 ≤30s + ①interfer-born/②split/③门距 + pose/dz 闸）| 离场硬证据 | `track_manager.go` `exitCoupledLostResidual` |

五者**正交叠加，不是五选一**：一条轨可被 interfer-born 出生软压，之后若又跳 ≥200 仍被 teleport 硬 purge，离场又被 exitCoupledLostResidual 删。它们：

- **共享**一个 fall-precursor 撤销闸（pose∈fall / Δz drop → 撤销 / 绝不 purge）——FN 安全网。
- **共享**一个出口 sink（artifact → 不进 still-box + 掉出 $N_r$），全落 realness 轴。

**interfer-born 归 static 这一支、不归 teleport**：血缘看动作——它和 static 同源于「born-here + confinement（不走就是伪迹，走了就撤）」，和 teleport 的运动学跳变是两码事。差别只在 **发现 vs 利用**：static = 对**未知** cell 做 5min 持久性证明、**学**出反射标签；interfer-born = **已知** interfer 标签下、**出生即软压**。别把它揉进 static 的 5min 发现路径，当成 fast-path 放旁边、共享 confinement-revocation 代码即可。

### 6.5 为什么不走 belief 的 SLeft 态

曾考虑把这些 artifact 信号路由进 belief 9 态的 `SLeft`（离场态），**否决**：

- **语义错配**：`SLeft` = 人朝门走出去、房腾空（由 ExitRoom / hand-off 驱动）。artifact 是「**从来就不是人**」= realness 轴，不是「离场」轴。phantom-only 的房归宿是 `SEmpty`，不是 `SLeft`。
- **作用域错**：artifact 是 per-track，`SLeft` 是 per-room。floor 被 `exitL < exitFlipLogOdds` gate——抬全房离场信念会**给全房关 floor**，1 真人 + 1 伪迹共存时会压掉真人的 floor 网。per-track 的 PURGE/软压只抹掉幻影自己的 still 贡献，**共存真摔的网完好**。
- **决定性**：PURGE/软压是**外科切除**（per-track），SLeft 是**全房麻醉**（per-room）；floor 是盲摔最后一道网，拿幻影派生的 SLeft 去 gate 它 = 捅洞。
- **其实已在正确的层统一**：census `Nr()` 是 **present-only + PReal≥0.5** 过滤。teleport 删轨 → $N_r$ 自动减；static Phase B → AreaReflector → floor 整片豁免（孤迹金属幻影无伴随真人，realness 抓不到，靠 floor 豁免压 FP）；interfer-born → ghost → 掉 $N_r$。多条**汇到同一出口**（realness/$N_r$ 或 floor 豁免），无需新建 SLeft。铁律 [[realness_never_vetoes_fall]]：artifact 只削 $N_r$、绝不碰 fall 否决；SLeft 恰恰会碰。

### 6.5-b · GhostLeftLogOdds — 离场者的滞留 ghost（present 静止，**走 SLeft**，与 §6.5 反射伪迹恰相反）

§6.5 说反射伪迹（mirror / interfer / static）「**从来不是人**」→ realness / `SEmpty`、**不走** SLeft。但有**第三类**长得像、语义却相反：**离场者的滞留 ghost**——一个真人**走了**，固件仍在原位吐**冻结的 present 轨**。它「**曾是人、现已离场**」，归宿正是 `SLeft`（不是 SEmpty、不是 realness）。这不是"路由 artifact 进 SLeft"（§6.5 否决的那件事），而是"离场语义本就归 SLeft"。

- **判据（`GhostLeftLogOdds`，[track_manager.go:1290](../tools/Xsensorv1/internal/roomengine/track_manager.go)）**：**present** 静止占用轨，在 room-empty 信号（ExitRoom / np=0 / tid=88）**之后**，按
  - **门距分桶** $P_{born}$：门口逗留（<110cm）→ 负分不划（护门口真人）；房深处滞留（≥140cm）→ 正分；
  - × **时间门** $g(\Delta t)$：距 room-empty 信号 Δt 的乘性门（<15s 太早=给真人重获机会=0；[15,45s] 线性升至满权；>45s 陈旧=0 退回 floor 兜底）
  → 累 SLeft 对数几率（单发≈门距分，封顶 `ghostLeftMaxLogOdds=8.0`，攒够顶过 `absorbedThresh` → absorbed-drop purge）。
- **硬门否决真摔**：注入前查 pose/Δz——**露过倒地前兆（fall pose / z 骤降）→ 不注入**（真摔者绝不被软清；[[realness_never_vetoes_fall]] 的 SLeft 版护栏）。**这是 FN-safe 命门**。
- **与 §6.5 不冲突**：反射伪迹 = per-track PURGE（外科切除，realness 轴）；GhostLeft = per-track SLeft 注入（离场轴，belief）。**两者都 per-track、都不全房麻醉**——§6.5 顾虑的"SLeft 全房关 floor、压掉共存真摔"，在这里被 **per-track absorbed-drop 只撤自己那条轨**化解（不动同房其它 track 的 floor 网）。
- **同设备 scope**：ghost 反射必与投射它的真人同设备，故 GhostLeft 按 same-device 判（[[per_device_np_latch_no_crossradar_clobber]]）。
- **🔴 适用边界 = 仅 present 轨**：人若**丢轨**（固件转 no-target、present=False）→ GhostLeft **不适用**。此时走 [DBN-Zone-Room §O](DBN-Zone-Room.md) 的 belief-Empty 房级路径——**单人 lost 盲区真摔的 FN 风险在那条路径，GhostLeft 覆盖不到**（b197-0630 实证：08:07:41 丢轨转 present=False → GhostLeft 失效 → 靠 belief 漂 Empty + F1 占用 gate 压 floor，而"站着"和"摔了"同样漂 Empty，不可分）。

---

## 第六·二部分 · split-ghost：单人走到墙/干扰区时的分裂收治

代码：[split_ghost.go](../wisefido-sensor/internal/roomengine/split_ghost.go)（fire-neutral，track-lifecycle 轴）。状态：✅ 已实现并接入 track_manager，两棵树 build/vet/gofmt 全绿。

> ⚠️ **已重写为 vote-based（下文 6.6–6.11 描述的早期"≤80cm 邻轨 / 窗末 score / hard purge"设计已被取代）**。现行三段式：
> **step1** np↑ 出新 offspring(无 EnterRoom)→ t1 各轨投最近 t0(前一tick)→ 得票最多者=**spliter**；
> **spliter 静止门**（仅 spliter：t-1,t-2 逐轴 dx,dy≤10 = 干扰源不动）；
> **step2** 投了 spliter 的轨入 `split_group`，锚点=spliter 前一tick位；
> **step3**（~2s 节流）走开者(净位移>200)=real(`EverWalkedOut` 不可逆)、赖锚点(≤10cm)者=ghost → **soft 压 `SplitGhostSinceMs`**（零 StillBoxSec、不删轨=无 churn、可撤销、露倒地相即解压锁 real）。下文 6.6–6.10 的**场景/安全性论证仍成立**，仅"删除判据"那段(6.9)的 hard-purge 动作换成 soft 压制。

### 6.6 要解决的场景：tid 会交换的并存分裂

室内仅 1 人，走近墙边 / interfer zone（帘、扇、金属外凸）时，firmware 偶把一条真轨**分裂成 2~3 条并存 track**（1 real + 1~2 ghost）。前述四类检测器都不覆盖它：

- **mirror** 要「贴真人同步运动」的 5 帧不变量；split ghost 常钉死墙边、不同步。
- **static** 要 300s 存活；split 几秒钟就要判。
- **teleport** 要 ≥200cm/s 单 tick 跳变；split 是「贴住原处冒出来」不是「跳过去」。
- **interfer-born** 只软压 floor 不删；split 要删并存伴幽灵以防 FP。

**核心难点：tid 会交换。** firmware 可能把 **tid 留给即将解散的 ghost**，真人换**新 tid** 走开。故 [[split_no_tid_priority]]：**绝不能按 tid 大小 / 出生先后认 real**，只能纯行为证伪。

### 6.7 为什么「只留一个」是安全的：FN 兜底不在本层（对象逻辑真值表）

直觉上「分裂现场只保留一条、删其余」是危险的。但真值表表明：**无论删的是真 A 还是 ghost，真人摔倒永不漏，因为 FN 兜底责任根本不在 split 层**，而在两条独立线（firmware / tFloor）。

| 情况 | 谁兜底 | 在哪条独立线 |
|---|---|---|
| ① A 是真人 | 真对了 | DBN 正常 |
| ② A 倒，真人 fall，**firmware fire** | firmware pose5 | **独立 alarm stream**（engine `runAlarmLoop`）——本层删 `tm.tracks` map 碰不到它 |
| ③ A 倒，真人 **silent fall**（fw 未 fire）| tFloor | A 这真人扑地静止 → 攒满 floor 在对的位置兜底 |
| ④ A 倒，真人 **exitRoom**（np=0），A lost | teleport / np=0 | 沿离场逻辑 purge，本层不另处理 |

### 6.8 ③ 的命门与漏洞补丁：真 A 绝不被 latch / A 游走防误锁

- **③ 命门：真·1 人 A 绝不被 `StillBoxSec=0` latch。** interfer-born/FloorArtifact 的动作是「留轨 + 抹 floor」；split 是**相反语义**——若误把真 A 也 latch 抹 floor，③ 的 silent-fall 兜底被自己关掉 → 漏报。故 **split 观察轨屏蔽 `maybeLatchFloorArtifact`**（已在 track_manager lost 循环按 `SplitObservingSinceMs==0` gate）。
- **漏洞补丁：A 游走被误当 real。** 若真 A 在 interfer zone 内小范围游走，会被误锁成「活的 real」→ 真人 silent fall lost 后系统仍以为「还有人在」→ 漏。补丁 [[split_ghost_walk_anchored_score]]：A 的「游走」必须被「**锚定在 interfer zone 的程度**」打折——ghost 游走 = confine 内贴 interfer；real walk = **走出 confine、远离 interfer**。
- **关键 FN 锁 [[split_everwalkedout_irreversible]]：`EverWalkedOut` 一旦置位不可逆。** 净位移曾 > walk-out(200) = 真人 → 永久 latch real，**后续 still / near-interfer 不得翻案**。`pFallReal≡1.0` 同理：正证据必 latch，不被后续 still 翻。

### 6.9 删除判据与 FN 硬闸

只删 **lost** 的轨（**present 真人永不删**）：

```
lost(过 presence coast) ∧ ghost_score ≥ promote(100) ∧ gate1NoPrecursorFrozen
  → PURGE + interferenceSuppress 反压（born 点位 drift，防 fw ~30s 滞报 churn）
```

- **FN 硬闸 `gate1NoPrecursorFrozen`**（复用 teleport 闸1）：末帧防 fall-precursor / Z 骤降 → **绝不删**（墙边摔者交 floor 兜底）。
- **「等」的省算 [[split_defer_score_to_window_end]]**：多数 ghost 在 window(10s) 内自然 lost 时解。观察期只维护廉价的 `EverWalkedOut` latch（1 次 distInt，无 grid 扫描）；沉贵的 interfer 距离打分推迟到 lost 时一次性算（`splitGhostScoreNow` 的 O(rad²) grid 探测仅此一次）。
- **latch 节流 [[split_latch_step_hold_safe]]**：walk-out latch per-track 节流（`splitLatchIntervalMs=2s`）。FN-safe 依据：walk-out 的真净位移 `distInt(born, now)` 是**持续保持**量而非脉冲——真人走出 200cm 后停下，净位移持续 >200，任何后续扫描都能捕到，不漏 latch。

### 6.10 已否决的两个激进简化

- ❌ **「10 秒后只处理 firmware fire」**：silent fall 真定义就是 fw 不 fire，此规等于把 ③ 的 tFloor 兜底关掉。
- ❌ **「滤掉 bed/lying zone 的 fire」**：床边 / lying zone 边缘跌倒（老人高发跌倒之一）会被误滤，踩 FN-safe 红线。

### 6.11 确定性梯度上的正确位置

split 的「等」不是通用技巧，而是**安全梯度上那个位置的正确选择**——每个检测器的等待尺度已匹配其证据类型：

| 检测器 | 证据类型 | 等待尺度 | 处置 |
|---|---|---|---|
| teleport | 运动学（确定）| **即时**（单 tick）| hard PURGE |
| split | 时机（概率）| **窗末**（10s）| hard PURGE（仅 lost）|
| static | 持久（慢确认）| **300s** | cell 学习 → AreaReflector |
| interfer-born | 空间先验（概率）| **出生即** | 软压 + 可撤销 |

split 复用 teleport 的「hard PURGE + 反压重建」动作，但触发是**出生分裂 + 窗末时机**而非运动学跳变；独立 `split_ghost.go`，与 teleport/static fast-path 并存，共享 fall-precursor 闸（闸1）与出口 sink（删轨 → $N_r$ 自动减）。

---

## 第七部分 · 落地路线

两步走，按"先利用已知标签、再放开自动学习"的风险次序：

### Phase 1 — interfer-born（✅ 已落地）
第六部分的 per-track interfer-born 软压制（**三闸 + 可撤销**）已实现：出生钩子 `stampInterferBornIfIsolated`（落 interfer cell + 出生孤立 ≥120cm）+ `revokeInterferBornIfMoved`（走出区/摔倒前兆即撤）；并已**删掉 floor.go 的 interfer 无条件豁免**——[floor.go:69-72](../wisefido-sensor/internal/roomengine/belief/floor.go) 现只留 `areaReflector` 整片豁免，interfer 走 default 12min（杂波区走进去摔的真人不再被漏）。

### Phase 2 — static reflector Phase B（✅ 6-29 已放开，与 interfer-born 联动）
static reflector 已从 Phase A（只 log）放开到 Phase B：三签名跨独立 episode 累 ≥ `StaticReflectorPromoteThreshold`(3) → `MarkStaticReflector` 升 cell→**AreaReflector**+SourceLearned（[grid.go](../wisefido-sensor/internal/roomengine/grid.go)），拿 floor 整片豁免；demote 由 [cell_learning.go](../wisefido-sensor/internal/roomengine/cell_learning.go) `hasRealActivity` 兜底（真人走过 → 退回 Unknown）。
- **升 AreaReflector 而非 AreaDeny**：金属幻影常孤迹（无人在房），realness co-existence 抓不到（需 ≥2 轨）；`tFloorFor(AreaDeny)` 走 default 12min 等于不压，唯一能压住 FP 的是 areaReflector 的 floor 整片豁免。mirror 仍升 AreaDeny（其 ghost 有真人伴随，realness 可抓）。
- 二者是 **发现 → 利用** 的闭环：static reflector **发现并标**出新的反射 cell，这些新标签可作 interfer-born 的输入（[[cell_dbn_timescales_stillbox_single_source]] 单源约束下）。
- ⚠️ **残留 FN 须真 case 盯**：近墙 AreaReflector 整片豁免，若真人正好摔在已学的金属点（扑地静止不 traverse → demote 不触发）→ floor 漏。demote 只对"走过"自纠，"摔在原地"是该豁免的固有二义。

---

## 第九部分 · 各 ghost 检出后怎么处置 × 在场场景（有真人 / 无真人 / ExitRoom）

**核心洞**：ghost 的 FP 危险**不均匀分布在三个场景上**——有真人时多重压制基本兜住，**无真人(空房)时孤 ghost 才是命门**（witness/多人闸全失效），ExitRoom 是从"有人"切到"无人"的**相变时刻**、最易留残迹。B17F 13:57 那次 FP 正是「空房 + 离场残迹 coast 11min」漏在这条缝里。

### 9.1 有真人在场（≥1 真人 present 且在动）

多重压制叠加，ghost FP 基本被兜住：

- **witness 先验**（`witnessNearby`，半径 200cm）：有**移动**真人在旁 → 直接 `StillBoxRunStart=0` 清 ghost 的 still-box（旁人会主动呼救，独留计时火报=FP）。**前提是真人在动**（焊死条件③）——静止旁观者不算 witness。
- **floor-artifact latch**（`maybeLatchFloorArtifact`）：immature(<5s) ghost + **有活轨共存** → `FloorArtifactSinceMs` 零其 still-box。
- **realness `N_r`**（census）：≥2 轨时 ghost 若被 mirror/co-existence 判低 PReal → 不计入人数（防把镜像当第 2 人改 fall 决策）。
- **split soft 压**：split_group 的 ghost 成员 → `SplitGhostSinceMs` 零 still-box，real 成员走开 latch 保护。

> 残留风险：**贴墙 ghost + 真人 >200cm 或不动** → witness 够不着（见 9.2 命门的前奏）。

### 9.2 无真人（firmware np=0）—— 命门

**先厘清物理**：**firmware 报任何一条 track（哪怕 byte-frozen ghost）→ np≥1**；**np=0 ⟺ firmware 一条 track 都不报**。所以"空房里有个 present ghost"**不存在**——会喂 floor 的那条 ghost **永远是 sensor coast 出来的 lost 轨**（firmware 早把它丢成 np=0，sensor 用冻结坐标续 still-box）。

**witness 空、多人 floor-artifact latch 不触发**（都要"有 present 真人"）→ **孤 coast 轨无人压**。只剩**不依赖共存**的几条：

- **interfer-born**（`InterferBornSinceMs`）：出生在 interfer cell + 出生时孤立 → 零 still-box。**孤轨也管用**（不靠 co-existence）。
- **static reflector → AreaReflector**：cell 学成反射 → floor 整片豁免。**孤迹金属唯一能压它的**（realness 对孤轨 PReal≡1 抓不到）。
- **split soft 压**：若该 ghost 是 split_group 成员且赖锚点 → `SplitGhostSinceMs`。
- **B `ghostLeftSuppress`/exitL**：设备报屋空(`deviceEmptySince`) + 静止轨在房深处(门距分) → exitL log-odds 压 floor。

> 🔴 **np=0 绝不能拿来删 coast 轨**：firmware 丢一条轨可能是**人走了**，也可能是**人摔在盲区/降功率被跟丢**——`np=0` 分不开这两者。删了 = 漏掉盲区真摔（lost-carry 续存正是为此 FN-safe 存在）。所以**删的触发只认 ExitRoom（定向穿越离场区的硬证据，确证是"走"非"摔"）**，不认 np=0（见 9.3）。

> **这就是 13:57 FP 的洞**：byte-frozen 残迹**既非 interfer-born、又（旧版）没被 split 抓、static 没学到** → 上面全不命中 → 孤 coast 轨满 tFloor 火报。**floor 安全网"无人也从 0 重暖机补火"的设计**（[[witness_prior_suppress_stillbox_floor]]）对真盲摔是兜底、对孤 ghost 就是 FP 源——而**正因 np=0 可能是盲区摔，这道兜底不能简单关掉**，只能靠 born-signature + ExitRoom 耦合精准删。

### 9.3 ExitRoom 时（有人→无人 相变）

人走光的**硬证据时刻**，最易留 coast 残迹。三条处置：

- **`exitCoupledLostResidual`（新，治本）**：本设备 ExitRoom 与某轨失锁**耦合 ≤30s** + （①interfer-born / ②split-group / ③门距分 `85*d/150>65`）+ pose∉{fall,sitGnd,walking,sit} + 末2tick dz≤40 → **直接 delete，不 coast**。把"离场残迹续 11min 喂 floor"从根上断掉。
- **B exitL**：屋空 + 静止深处 → 压 floor（**不删轨**，软压；与新 delete 互补：delete 治"该走的残迹"，exitL 兜"没删但该压的"）。
- **belief `SLeft`**：ExitRoom 硬证据 → `exitL ≥ exitFlipLogOdds` → **per-room** gate floor（§6.5：全房麻醉，仅在确证离场时用，不拿 per-track 幻影派生）。

> FN 闸（三场景共享）：露**倒地前兆 / Δz 下坠 / lying/sit/walking** 的轨**绝不删/压**——真人贴干扰源、在门口、刚摔倒都被这道闸接住（真盲摔照常走 tFloor 兜底）。

### 9.4 处置矩阵（速查）

| 机制 \ 场景 | 有真人在动 | 空房(np=0) | ExitRoom |
|---|---|---|---|
| witness 清 still | ✅ 主力 | ❌ 无 witness | ❌ |
| floor-artifact latch | ✅（多人+immature）| ❌（要共存）| ❌ |
| realness $N_r$ | ✅（≥2 轨判低 PReal）| ⚠️ 孤轨 PReal≡1 抓不到 | — |
| interfer-born 软压 | ✅ | ✅（不靠共存）| ✅ |
| static→AreaReflector | ✅ | ✅（孤迹金属唯一解）| ✅ |
| split 软压 | ✅ | ✅（若 group 成员）| ✅ |
| B exitL 压 floor | — | ✅（屋空深处）| ✅ |
| **exitCoupledLostResidual 删** | — | ❌ **np=0 不删（怕盲区摔）** | ✅ **治本（认 ExitRoom）** |

**读法**：
- **空房(np=0) 那列最稀**——孤 coast 轨只剩 born-signature(interfer/static/split) + exitL 软压；**删不敢删**（np=0 可能是盲区真摔，删=漏报）。
- **治本只能落在 ExitRoom 那列**：定向离场硬证据确证"走非摔"，才敢 delete。无自身 ExitRoom、只 np=0 掉下来的 coast 残迹，靠"耦合最近一条 ExitRoom ≤30s"接（B17F t2 即如此）；耦不上的孤静止轨**仍是 FN-safe 代价下的残留 FP 风险**（继续真 case 盯，必要时上 byte-frozen 0-抖动识别——0 抖动是物理上人不可能、可与盲区摔区分）。

---

## 第十部分 · lost_track 判据扩展：live-frozen residual + EverGhost 闩（d5f7-0630 驱动）

case：`case-d5f7-0630-19021918`，room `fd00:0:3:111:3:300`（浴室，layout 已声明 `mirror`(Interfere)+`Reflector (marked)`(MetalCan)），ghost = lid `D5F711158099`（下称 58099），真人 = lid `...36288`。生产 19:15:50 报 Fall-Cloud，replay 复现机制（窗口太短未 fire，按规则 #3 只验机制）。

本部分扩展 §6.4 的 `exitCoupledLostResidual` 与 §9 的三场景处置——现行删除全部 gated 在**「轨已 lost（固件停报）」**上，58099 暴露了一条绕过全部防线的新缝。

### 10.1 新洞：present-but-byte-frozen residual —— 固件不停报，永远进不了 lost 分支

58099 出生 19:11:58（np 1→2，**无 EnterRoom**），census 共存期正确判镜像（p_mirror 0.42→**0.94**，p_real 压到 **0.06**）。真人 19:16:02 ExitRoom 离场后，固件**继续逐帧照发** 58099——byte 冻结在 (180,−40)、Δ=0、np 掉到 1（**不是 0**）——直到 19:18:37 才消失，其间 still_sec 一路涨到 **167s** 还在涨（导出窗末），生产里就此撞 tFloor。

**代码级根因**（[track_manager.go:1514](../tools/Xsensorv1/internal/roomengine/track_manager.go)）：

```go
for id, ts := range tm.tracks {
    if activeIDs[id] { continue }   // ← 本帧被固件观测到 = 跳过整段驱逐
    ... exitCoupledLostResidual / witnessNearby / lostStillCarry 全在此下 ...
}
```

`exitCoupledLostResidual`、witness-lost 清、lostStillCarry 驱逐**全部在 `!activeIDs`（未观测）分支内**。58099 每帧 `activeIDs[id]==true`（固件照发），`continue` 直接跳过 → **一条驱逐都不求值**。§9.2 说「喂 floor 的永远是 sensor coast 的 lost 轨」——d5f7 是**反例**：一条**固件持续上报的 present 冻结 ghost**（np≥1）也能磨 still-box 喂 floor。这正是待办 `byte-frozen 0-抖动识别` 要收的形态，且必须在 **present 分支**判，不能只在 lost 分支。

### 10.2 第二个洞：软压随「孤镜面转正」蒸发 —— 判 ghost 必须取 ever/peak

58099 的 p_real 弧线：born 0.58 → 共存期 **0.06**（census 咬死镜像）→ 真人一走 → **秒跳 1.00 并锁死**（[[realness_axis_redefined_real_vs_mirror]] 孤镜面 p_real≡1，mirror 配对需 ≥2 轨，孤轨配不上对）。

**唯一识破它的信号（0.06），恰在它变危险（孤轨磨 floor）的瞬间蒸发。** 现行 split/interfer/mirror 软压都是**「当前状态成立才压」**——孤轨转正后条件消失、压制解除。故 fire 闸**不能信 fire-时刻的瞬时 p_real**，必须信**历史峰值**。

> **EverGhost 闩**：共存期一旦 `p_mirror ≥ 0.8`（持续 ≥N tick）且**无出生证**（born 无 EnterRoom）→ 置 `EverGhostSinceMs`（单调、**不随 p_real 翻 1.0 而清**）。fire 闸只查这个闩。撤销唯一口：`EverWalkedOut`（净位移>200，镜像不会脱离本体独走）或摔倒前兆（pose/Δz，[[realness_never_vetoes_fall]] 护栏）。

### 10.3 判据各轴的 FN 权重（**定稿判据见 §10.3-b，本节只给各轴的 FN 依据**）

58099 是**多轴擦线选手**——每条现行硬闸都差 10~15cm：门距 100cm（<115 驱逐阈、<120 先验 floor 阈）、离真人轨 108cm（>80 split 阈、<120 interfer 阈，落**正中死区**）。故单调任一阈值都误伤真人；§10.3-b 的联合判据即为此。下表列各轴 FN 权重（供 §10.5 量化引用）：

| 轴 | 语义 | 反-FN 作用 |
|---|---|---|
| 无出生证（born ghost>50） | 出生无 EnterRoom + born p_mirror>0.5 | **主闸**：真人必走门→需固件漏发进门（~3%）|
| Max ghost>80（EverGhost 闩）| 生涯峰值 p_mirror>0.8 | **主闸**：真人被判持续镜像需巧合几何对称（~2%）|
| R<120 | 净位移<120cm=赖锚点没走开 | 弱：摔倒本就原地（~70%）|
| dz≤40 | 末2tick 垂直≤40cm=无掉高 | 软：固件 z 会撒谎（~20%）|
| pose∉{fall,lying} | 无摔姿 | 软：固件 pose 会撒谎（~30%）|

### 10.3-b `lost_fall delete` 定稿判据（用户定，逐字）

> 权威判据，按此实现。`delete = 条件1 ∧ 条件2 ∧ 条件3`。

**条件1（生命周期，二选一）**：
```
ExitRoom ≤30s  lost_track
  或
ExitRoom == null  &&  全程 np == 1
```

**条件2（出生/ghost 证据，四选一）**：
```
1. interfer-born
2. split
3. EnterRoom == null  &&  born_ghost 门距分 = 80*d/150 >= 60      (d = 出生地→最近门 cm；阈 d ≥ 112.5)
4. census p_mirror ≥ 0.8 且 sustained ≥3 tick      (非峰值——真镜像帧帧成立，两真人巧合并排走的尖峰撑不过 3 帧)
```

**条件3（FN 硬闸，全满足）**：
```
pose ∉ {2, 5, 6, 7, 8}      /* 2 susfall, 5 fall, 6 lying, 7/8 sitOnGround */
&&  dz  <= 40
&&  R   <= 120
```

> 对照现行 `exitCoupledLostResidual`（§9.3）：条件1 的 `ExitRoom≤30s lost` = A 分支（已生效）；`ExitRoom==null && 全程np==1` = B 分支（**新增**：无离场硬证据、单人全程的孤影自生自灭，present∨lost 都查，治 58099 那种 present-frozen 永不 lost）。条件2 的 ①②③ 已在 exitCoupledLostResidual，④ `Max ghost>80`（EverGhost 闩）为**新增**。条件3 与现行 pose/dz 闸一致，另加 `R≤120`（净位移 confined）。

### 10.4 三场景处置（A/B/C 对应 §9 的 有人/空房/ExitRoom）

| 场景 | 定义 | 处置 | FOR≈ |
|---|---|---|---|
| **A** | 无出生证 + **ExitRoom 耦合** lost | **删**（= 现行 `exitCoupledLostResidual`，扩到 present-frozen：EverGhost 命中即删）；留 `EverWalkedOut` 撤销 | ~10⁻³ |
| **B** | 空房无 EnterRoom 凭空 born → 无 ExitRoom 直接 vanish **+ 有配对源轨** | **无条件删**：守恒论证——realness 守恒于配对源，删「影」留「本体」，真人当场尚存 | ~10⁻⁴ |
| **C** | 有人时消失，但**无配对 / 在场者是旁观** | **走 witness 5 条焊死**（[[witness_prior_suppress_stillbox_floor]]），**不无脑删**——在场者可能是旁观、消失的可能是第二真人摔倒 | 显著↑ |

**B 的守恒论证**（为何可无条件删）：`Max ghost>80` 本身要求存在配对轨 → 58099 是某真人轨的反射；census 把 95% 判给源、5% 判给影；删 5% 的影、留 95% 的本体，**真人一条不丢**。「如果 b 也是真人」三重兜底：① 真摔多半 dz>40/pose=fall → 违判据不删；② 两独立真人撑不起 mirror>80；③ 静默摔也有在场真人当 witness。

**边界（别混轴）**：B 只管「配得上对的影」。**孤幽灵**（真孤轨、无配对）算不出 mirror>80、进不了 B，归 §6 出生地先验轴（[[mirror_detect_single_mirror_boundary]] 孤迹 PReal≡1）。

### 10.5 FN 量化（术语对齐）

混淆矩阵（阳性=发火报警）：**FP=假阳性/误报**（ghost 不删→floor 误火，要压的）；**FN=假阴性/漏报**（把真摔当 ghost 删）。Pa/Pb 是后验 **FOR（假遗漏率）= P(真摔 | 已判ghost删) = 1−NPV**。

删除判据被真摔全部命中的联合概率（各闸连乘）：

$$\text{FNR} \approx \underbrace{0.03}_{\text{无门}}\times\underbrace{0.02}_{\text{镜像}}\times\underbrace{0.7}_{\text{confined}}\times\underbrace{0.2}_{dz}\times\underbrace{0.3}_{pose} \approx 2.5\times10^{-5}$$

保守版（软闸 dz/pose 不给信用，固件对静默摔全程撒谎）：`0.03×0.02×0.7 ≈ 4×10⁻⁴`。故 **FNR ∈ [10⁻⁵, 10⁻⁴]**，主闸是前两条（无门×镜像=6×10⁻⁴）。车队换算（1000 老人×~1.5 真摔/年）：**每 ~7 年 fleet 级漏 1 次**（带宽 1.7~27 年）。

> ⚠️ 结构量级估计非实测——[[fall_data_is_artificial_test]] 标不出「固件漏进门率 / 镜像误标率 / pose 撒谎率」三个关键输入。落地焊死须拿生产 event_log/monitor_stream 回测这三项（待办新增）。

### 10.6 与现行五检测器的关系（扩展非替换）

- **扩 `exitCoupledLostResidual`**：触发从「lost」放宽到「present ∨ lost」——EverGhost 命中的 present-frozen 轨也走同款删除（现行只在 lost 分支）。
- **扩软压为闩**：split/interfer/mirror 的软压结论一旦达 `Max ghost>80`+无出生证，升格 `EverGhostSinceMs`（不蒸发），治「孤镜面转正解除软压」。
- **共享**同一 fall-precursor 撤销闸（pose∈fall/Δz drop 绝不删）与出口 sink（still-box 零 / 掉 $N_r$），与 §6.4 五检测器正交叠加。

---

## 第八部分 · 待办

- [x] **Phase 1**：interfer-born 软压制（三闸 + 可撤销）+ 删 floor interfer 豁免——已落地。
- [x] **Phase 2**：static reflector Phase B（升 AreaReflector）+ demote 兜底覆盖 reflector——6-29 已放开。
- [x] **split-ghost vote-based 重写**（第六·二部分）：vote→spliter 静止门→group→step3 soft 压 `SplitGhostSinceMs`，已接 track_manager，两棵树全绿。
- [x] **exitCoupledLostResidual**（第九·三）：ExitRoom 耦合 ≤30s + ①interfer/②split/③门距分 + pose/dz → 删离场残迹；replay 验 faller still 封 83s(was 739)、fire 0。
- [ ] **byte-frozen 0-抖动识别 / EverGhost 闩**（判据见第十部分，d5f7-0630 驱动）：① `EverGhostSinceMs` 单调闩（共存期 p_mirror≥0.8+无出生证，不随 p_real 翻 1.0 而清）；② fire 闸在 **present 分支**（非只 lost）查闩 → 5 条 AND 判据命中即删/零 still-box；③ A/B 删、C 走 witness。空房孤静止轨残留 FP 的兜底。
- [ ] **FN 三输入回测脚本**（第十·五）：扫生产 event_log/monitor_stream 数出「固件漏 EnterRoom 率 / 镜像误标率 / pose 撒谎率」→ 把 FNR 从结构估计（~10⁻⁴）变实测。
- [ ] **审计 census 判 ghost 时几何项（IsReflection）vs 共生项的权重占比**——确认是否过度依赖几何、而 co-existence/vital 没吃满。
- [ ] 评估把「持续无 vital 的 co-existing track」接入 census 作 ghost 一票（依赖 [[radar_hr_rr_bed_enter_gated]] 的覆盖范围）。
- [ ] static reflector 近墙 AreaReflector 整片豁免的残留 FN（摔在已学金属点）——真 case 盯，必要时收窄为按身份压而非整片豁免。
