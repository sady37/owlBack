# Ghost 反射伪迹检测：原理、适用边界与判据迁移

状态：**设计/原理底稿**。用途：(1) 厘清当前 ghost 反射检测的四条防线各自的物理原理；(2) 从数学上证明 mirror_detect 的**适用边界**——它只对「单一稳定镜面」有效，对「离散多反射体」结构性失效；(3) 给出治理方向：把判据权重从「几何对称」迁到「共生 + 生命体征」这些**对几何不敏感**的轴上。

关联：[[mirror_detect_single_mirror_boundary]]（本文核心结论的记忆索引）、[[reflection_cells_areareflector_static_phaseA]]（mirror_detect 升格 AreaReflector / static Phase A）、[[teleport_interference_purge_mechanism]]（≥200 瞬移 PURGE）、[[realness_axis_redefined_real_vs_mirror]]（realness 轴 = Real vs Mirror，主职 N_r 排 ghost）、[[realness_never_vetoes_fall]]（铁律：realness 绝不否决 fall）、[[two_radar_mirror_gate_samedevice]]（镜像门控同设备）、[[gatelist_retired_ghost_single_source_census]]（census 单源 PReal）、[[cell_dbn_timescales_stillbox_single_source]]（still-box 单源）、[[radar_hr_rr_bed_enter_gated]]（radar vital gated）。

---

## 第一部分 · 物理本质

雷达伪迹（ghost）的物理根因是**多径反射**：雷达波打到金属/玻璃/镜面后弹一次再返回，固件把这条二次路径的回波当成一个**额外的人**，生成一条假 track。这条假 track 的位置，就是真人关于那面「镜子」的**镜像点**。

所以判 ghost 的核心命题始终是同一句：**这两条 track 是不是一对镜像？** 当前系统有四条防线在从不同角度回答它。

| 防线 | 文件 | 抓什么形态的 ghost | 依赖 | 接线状态 (6-27) |
|---|---|---|---|---|
| **Mirror 检测** | [mirror_detect.go](../wisefido-sensor/internal/roomengine/mirror_detect.go) | 单一稳定镜面 + 连续运动 | 几何对称（轴 + 速率同步） | ✅ 生效，命中升 AreaDeny |
| **IsReflection** | [census.go](../wisefido-sensor/internal/roomengine/adapter/census.go) `reflectSep()` | 出生瞬间、墙外镜像 | 几何（连线穿墙） | ✅ 生效，喂 realness 观测 |
| **Teleport purge** | [teleport_interference.go](../wisefido-sensor/internal/roomengine/teleport_interference.go) | 跳变 / 瞬移 | 运动学（人类速度上限） | ✅ **生效，真 PURGE 删轨**（track_manager.go:1158/1200/1329，无 flag 默认开） |
| **Static reflector** | [static_reflector.go](../wisefido-sensor/internal/roomengine/static_reflector.go) | 静止金属反射体 | 持久性（出生位 + 久驻 + 近墙） | ⚠️ **Phase A，只 log 不升格**（scan 在 track_manager.go:1326 每帧跑，MarkStaticReflector 只 `Count++`，无 ≥3 升 AreaDeny；待真 case 验标点后放 Phase B） |

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

### 6.4 三检测器分工（动作 / 确定性 / 归属）

| 触发 | 动作 | 确定性 | 归属文件 |
|---|---|---|---|
| **teleport 跳变** | hard **PURGE** 删轨 | 运动学（确定）| `teleport_interference.go` |
| **interfer 出生** | **soft 压制 + 可撤销**（不进 still-box / 掉 $N_r$）| 空间先验（概率）| `static_reflector.go`（fast-path）|
| **static 持久** | **cell 学习 → AreaReflector/AreaDeny** | 持久性（慢确认）| `static_reflector.go` |

三者**正交叠加，不是三选一**：一条轨可被 interfer-born 出生软压，之后若又跳 ≥200 仍被 teleport 硬 purge。它们：

- **共享**一个 fall-precursor 撤销闸（pose∈fall / Δz drop → 撤销 / 绝不 purge）——FN 安全网。
- **共享**一个出口 sink（artifact → 不进 still-box + 掉出 $N_r$），全落 realness 轴。

**interfer-born 归 static 这一支、不归 teleport**：血缘看动作——它和 static 同源于「born-here + confinement（不走就是伪迹，走了就撤）」，和 teleport 的运动学跳变是两码事。差别只在 **发现 vs 利用**：static = 对**未知** cell 做 5min 持久性证明、**学**出反射标签；interfer-born = **已知** interfer 标签下、**出生即软压**。别把它揉进 static 的 5min 发现路径，当成 fast-path 放旁边、共享 confinement-revocation 代码即可。

### 6.5 为什么不走 belief 的 SLeft 态

曾考虑把这些 artifact 信号路由进 belief 9 态的 `SLeft`（离场态），**否决**：

- **语义错配**：`SLeft` = 人朝门走出去、房腾空（由 ExitRoom / hand-off 驱动）。artifact 是「**从来就不是人**」= realness 轴，不是「离场」轴。phantom-only 的房归宿是 `SEmpty`，不是 `SLeft`。
- **作用域错**：artifact 是 per-track，`SLeft` 是 per-room。floor 被 `exitL < exitFlipLogOdds` gate——抬全房离场信念会**给全房关 floor**，1 真人 + 1 伪迹共存时会压掉真人的 floor 网。per-track 的 PURGE/软压只抹掉幻影自己的 still 贡献，**共存真摔的网完好**。
- **决定性**：PURGE/软压是**外科切除**（per-track），SLeft 是**全房麻醉**（per-room）；floor 是盲摔最后一道网，拿幻影派生的 SLeft 去 gate 它 = 捅洞。
- **其实已在正确的层统一**：census `Nr()` 是 **present-only + PReal≥0.5** 过滤。teleport 删轨 → $N_r$ 自动减；static Phase B → AreaReflector/Deny + 出生即 ghost → PReal 被压 → 掉 $N_r$；interfer-born → ghost → 掉 $N_r$。三条**已汇到同一出口**（realness/$N_r$），无需新建 SLeft。铁律 [[realness_never_vetoes_fall]]：artifact 只削 $N_r$、绝不碰 fall 否决；SLeft 恰恰会碰。

---

## 第七部分 · 落地路线

两步走，按"先利用已知标签、再放开自动学习"的风险次序：

### Phase 1 — 开始 interfer-born
实现第六部分的 per-track interfer-born 软压制（**三闸 + 可撤销**），并**删掉 [floor.go:69-72](../wisefido-sensor/internal/roomengine/belief/floor.go) 的 interfer 无条件豁免**（interfer → default 12min，只留 reflector 豁免）。
- 它**利用已存在的 interfer 标签**（人 pin / AreaType 声明），风险低，立刻解 6.1 的 FN。
- 落点：track_manager 出生钩子盖 provisional ghost → still-box 入口闸 + 掉 $N_r$；放 `static_reflector.go` fast-path，共享 confinement-revocation。

### Phase 2 — 启动 static reflector + interfer-born
把 static reflector 从 **Phase A（只 log）放开到 Phase B**（cell 命中 ≥ `StaticReflectorPromoteThreshold` 升 AreaReflector/AreaDeny，[[reflection_cells_areareflector_static_phaseA]]），与 interfer-born **联动上线**。
- 二者是 **发现 → 利用** 的闭环：static reflector **发现并标**出新的反射/干扰 cell，这些新标签**正是 interfer-born 的输入**。
- 先 Phase 1 后 Phase 2 的理由：interfer-born 吃既有标签、可独立验证收益；static Phase B 是更激进的自动改 cell（[[cell_dbn_timescales_stillbox_single_source]] 单源约束下），待 interfer-born 验稳、且 static 标点经真 case 核准后再放，并与 interfer-born 同时上以形成闭环。

---

## 第八部分 · 待办

- [ ] **Phase 1**：实现 interfer-born 软压制（三闸 + 可撤销）+ 删 floor interfer 豁免。
- [ ] **Phase 2**：放开 static reflector Phase B + 与 interfer-born 联动（发现→利用闭环）。
- [ ] **审计 census 判 ghost 时几何项（IsReflection）vs 共生项的权重占比**——确认是否过度依赖几何、而 co-existence/vital 没吃满。
- [ ] 评估把「持续无 vital 的 co-existing track」接入 census 作 ghost 一票（依赖 [[radar_hr_rr_bed_enter_gated]] 的覆盖范围）。
