# DBN：Zone 与 Room 的关系 + 联合占用模型方程

状态：**设计底稿**。用途：(1) 厘清 zoneengine（Zone）与 roomengine/DBN（Room）的职责边界与当前耦合的错误；(2) 给出把床占用 $O_b$、人数 $N_r$ 从「硬外部输入」改为「联合隐变量」的完整数学模型与每条方程的含义。
关联：[[belief_dbn_signal_map]]（节点/发射/转移总映射）、[[zone_engine]]（ZoneState 三态状态机）、[[DBN-cd2b]]（cd2b case 测试）、[[mm_relationship_matrix]]、[[belief_redesign_fullspace_mandate]]、[[partial_monitoring_fall_suppression_law]]。

---

## 第一部分 · Zone 与 Room 的差异

### 1. 两者各算什么

| | **Zone（zoneengine）** | **Room（roomengine / DBN）** |
|---|---|---|
| 产物 | **ZoneState**：床/房/卫生间的三态 `Vacant/Occupied/Leaving` + 人数 | **S_room**：人在全空间的 9 态占用/姿态信念（含 Fallen） |
| 性质 | 多源融合后的**硬结论**（状态机 + 各自的小贝叶斯） | 逐帧**联合后验** $P(S\mid o_{1:t})$（HMM forward） |
| 时间尺度 | 秒（事件驱动） | 逐帧 ~1Hz |
| 写者 | sensor zoneengine（单源真相，下游只读） | roomengine belief 层 |
| 入口 | `wisefido-sensor/internal/zoneengine/` | `wisefido-sensor/internal/roomengine/belief/` |

Zone 回答「**这张床/这个房现在占没占**」；Room 回答「**人此刻在哪、是什么姿态、是不是摔了**」。

### 2. 当前的耦合：Zone 的结论被 Room 当硬观测吃掉（这是 bug）

蓝图（[[belief_dbn_signal_map]] §1）里 $O_b$（床占用）、$N_r$（人数）是 DBN 的**隐节点**，应与 $S\_room$ **联合推断**。但当前实现把它们做成了 DBN 的**外部观测源**：

```
[zoneengine]                                  [roomengine / DBN]
 O_b 床占用(自跑 bed_bayesian) ──BedReleased(硬布尔)──► S_room 似然分支
 N_r 人数(硬计数)              ──ObsNumberPeople(硬)──►
```

- zoneengine 自己跑床贝叶斯，产出 `BedOccupied/LeftBed` **硬结论**；DBN 在 `likelihood.go` 里按 `BedReleased∈{0,1}` 走似然分支（`inBed && !BedReleased → SBed`，SFallen 零分）。
- **第一步的不确定性在第二步丢失**——DBN 看到的只是 `true/false`，无法质疑 $O_b$，因为 $O_b$ 不是联合隐变量。

### 3. 为什么错：cd2b 漏报是这个错误在不同层面的复现

实证（[[DBN-cd2b]]、bed 融合审计）：一次床边真摔被判 `SBedLying` 而非 `SFallen`。逐层追下来：

1. 似然层 `likelihood.go:151` 硬路由 `inBed && !BedReleased → SBed only`（SFallen 零分）；
2. `BedReleased` 永假，因为 `BedOccupancyState` 的床占用退化成 **radar InBed** 撑着（sleepad 1641 当时整设备离线，最后一条 InBed 后近 3h 无上报）；
3. radar 的 LeftBed 是 **zone-based**（离开床区多边形才发），人摔在床区**内** → 不发 LeftBed → 床态卡 occupied。

> 关键结论：cd2b **不是路由 bug**（1641 与 cd2b radar 经 IPv6 /88 前缀解析到**同一个 room**，MM/绑定正确），而是 **identifiability + 硬外挂**问题——接触权威（sleepad）离线后，把「上次在床」的陈旧硬结论当此刻真相，违反 [[partial_monitoring_fall_suppression_law]]。任何在硬 $O_b$ 上打的补丁都会再漏。

### 4. 修法：$O_b$ 进联合隐空间，MM 当耦合，全空间推断

不再有 zoneengine→DBN 的硬结论交接。所有传感器（sleepad/radar/vital）都是联合隐变量 $(S,B)$ 的**观测**；床占用 $B$ 在滤波器内被推断，不确定性全程传播：

$$\underbrace{P(S_t,B_t\mid o_{1:t})}_{\text{联合推断}}\quad\text{取代}\quad \underbrace{P(S_t\mid \hat O_b^{\text{硬}},o_{1:t})}_{\text{当前:吃硬结论}}$$

Zone 不消失——它仍产出 ZoneState 供 card/告警/离床漫游等产品面（[[device_config_vs_spatial_config_split]]）；变的是 **DBN 不再依赖它的结论，而是消费原始证据**。下面给完整模型。

---

## 第二部分 · 联合占用模型方程与含义

记号桥接：$S$=S_room（人态，含 Fallen $F$）；$B^j$=第 $j$ 床的 $O_b$；多 track 时 $N_r=\sum_i\mathbb 1[S^{(i)}\notin\{E,L\}]$。Cell engine 仍是外生只读慢层（[[cell_dbn_timescales_stillbox_single_source]]），不进本联合推断。

### 1. 隐状态

$$J_t=\Big(S_t,\ \{B^j_t\}_{j\in\mathcal B}\Big),\quad S_t\in\{E,\text{AtBed},Sit,O,Ba,F,BR,BO,L\},\quad B^j_t\in\{0,1\}$$

**含义**：把「人在哪/什么姿态」($S$) 与「每张床垫占没占」($B^j$) 做成**两根正交隐轴**的笛卡尔积（基数 $9\cdot2^{|\mathcal B|}$）。$B$ 不再折进 $S$、也不再外挂——它和 $S$ 一起被联合滤波。床归属不设独立隐维，用 §4 的软权重 $a_j$ 边缘化，避免状态爆炸。

### 2. MM 几何原语（cell 计数，[[mm_relationship_matrix]]）

床 $b$ 由 cell 集 $C_b$ 定义。设备对床的支撑集 $R(r,b)$（雷达可定位的 cell）、$P(s,b)$（垫可感知的 cell）：

$$\text{covers}(r,b)=\tfrac{|R(r,b)|}{|C_b|},\quad \text{onbed}(s,b)=\tfrac{|P(s,b)|}{|C_b|}\in\{0,1\},\quad o_b(r,s)=\tfrac{|R(r,b)\cap P(s,b)|}{|C_b|}$$

**含义**：`covers`=雷达覆盖这张床多少（连续）；`onbed`=这个垫是不是这张床的（二元，垫严格绑一床）；`overlap` $o_b$=雷达视野与垫足迹的**实际交叠**（不是 `min(covers,onbed)` 那种上界代理）。三标量一次 cell 枚举产出，$M$ 只存这三个。

### 3. 同床耦合置信 $\kappa$（几何冷启 + 事件驱动时间修，**无 max**）

$$\kappa^{(0)}_{r,s_j}=\text{covers}(r,j)\cdot\text{onbed}(s_j,j)\quad(\text{分级冷启})$$

更新（互活门控）：任一设备在 $t_e$ 产床事件，且 $r$ 与 $s_j$ 在 $[t_e\pm K]$ **均在线**时——

$$\boxed{\ \kappa_{r,s_j}\leftarrow(1-\gamma)\,\kappa_{r,s_j}+\gamma\,\mathbb 1\big[\text{另一设备在 }K\text{ 内有匹配床事件}\big]\ }\quad(\text{否则不更新})$$

**含义**：$\kappa$=「雷达 $r$ 和睡眠器 $s_j$ 看的是不是同一张床」的置信。几何只播种初值，**时间在线修正且可升可降**——这是你的「两事件越近、越频繁，逻辑关系越强」的数学化（带遗忘 Beta–Bernoulli，几何作先验伪计数）。
- **不用 `max(几何, 时间)`**：max 只增不减，几何一旦重叠就钉死 1，时间再也压不下来 → 多床消歧失效。
- **互活门控**：只在双方都在线才更新。对方离线/沉默 → 不更新（无信息），不能把「对方没发」误读成「不同床」。
- $K$=动作+延迟窗（~5–10s），固定；只调遗忘率 $\gamma$。

### 4. 软床归属 $a_j$ + 片内相容势 $\Psi$

几何 XY 似然 $g^{xy}_j$（雷达对床的可分性：能分→尖峰；只知在某床→床间均匀；看不见→0）。归属：

$$a_j=\frac{\kappa_{r,s_j}\,g^{xy}_j}{\sum_{j'}\kappa_{r,s_{j'}}\,g^{xy}_{j'}+c_\varnothing},\qquad a_\varnothing=\frac{c_\varnothing}{(\cdot)}$$

物理相容表（固定，$\varepsilon_{\text{art}}$=被子残留极小概率）：

$$\psi_{\text{phys}}(S,B^j)=\begin{array}{c|cc}
S & B^j{=}\text{occ} & B^j{=}\text{vac}\\\hline
\text{AtBed} & 1 & 1-o_j\\
F & \varepsilon_{\text{art}} & 1\\
\text{else} & 1 & 1
\end{array}$$

$\kappa$ 退火向中性 + 按归属混合：

$$\tilde\psi_j(S,B^j)=\kappa_{r,s_j}\,\psi_{\text{phys}}(S,B^j)+(1-\kappa_{r,s_j}),\qquad \boxed{\ \Psi_t(S,\{B^j\})=\sum_{j}a_j\,\tilde\psi_j(S,B^j)+a_\varnothing\ }$$

**含义**：$\Psi$ 是同一帧内 $S$ 与 $B$ 两轴的**相容势**（不是转移）。
- $\psi_{\text{phys}}$ 由**物理**定：人在床→垫占（相容 1）；人在床但垫空=床内头/脚端错位（$1-o_j$）；**摔倒→离垫→垫空（相容 1），摔倒却垫占只能是被子残留（$\varepsilon_{\text{art}}$）**。Fallen 行不许被 $\kappa$ 覆盖。
- $\kappa$ 只**调耦合强度**：$\kappa\to0$（不确定同床）→ $\tilde\psi\to1$ 中性（诚实无知，不强绑）；$\kappa\to1$ → 全物理表。
- 远离任何床 → $g^{xy}_j\to0$ → $a_j\to0$ → $\Psi$ 中性：开阔地跌倒不牵涉床垫（**靠轨迹位置 $g^{xy}$ 门控，非显式势惩罚**；这也是为什么 `else=1` 不罚 $O$+occ 仍自洽）。
- 几何 `overlap` 与时间 `co-occur` 是**同一个量 $\kappa$ 的两种估计**；但 `overlap` 还另含床内对齐 $o_j$，时间估不出——「时间替代几何」只对同床归属 $\kappa$ 成立。

**$\kappa$ 衰减 FN-safe 的承重前提（2026-06-20，A+C 对码核实）**：曾担心「安静熟睡者无床事件 → $\kappa$ 衰减 → $a_j$ 降 → SBed 保护掉 = FN」，欲加「无事件不衰减」的 gate。核实后**收回，gate 不需要**——因为 SBed 有两类来源，$\kappa$ 只进其一：
- **(a) 雷达直接抬 SBed，全 $\kappa$-free**：M×N(firmware 床 area_id)/HR-RR(近床 present)/pose=Lying，均在 `radarLogS` 直接 `addLogLk`，不过 $\kappa$。
- **(b) 睡垫 B 轴 → $\Psi$ 耦合 → 拉 $S$ 向 SBed**：才过 $\kappa$（进 $a_j$ 与 $\tilde\psi=\kappa\psi_{\text{phys}}+(1-\kappa)$ **两处**，非单一线性权重）。

SBed 在 $T$ 层有**自衰减 ≈0.80/tick**（model.go `SBed` 行自保持 $80/99.75$，~20% 流出；transition 层，**不分设备、不碰 $\kappa$**）。维持 = 「证据每帧抬 vs 自衰减」平衡：证据在→稳高位，证据没→几 tick 掉光（这本身就是正确的 FN-safe，不需主动清，亦答 §6 的「$B$/态无 →Empty 转移也不会永久滞留」）。$\kappa$ 衰减只削 (b)；真要保护的睡者**总有一条 (a) 的 $\kappa$-free 源撑 SBed**（雷达可见→HR/RR+firmware 床；睡垫-only→合成 bed-track 自带 pose=Lying+firmware 床），故 $\kappa$ 衰减**饿不死**他。结论：$\kappa$ = **纯归属权重**（多床多人时决定睡垫床信息贴哪个 track、贴多强），升降 FN-safe，wire UpdateKappa 不碰 SBed 维持。

> ⚠️ **承重不变量**：「不要 gate」的安全**建立在**——*每个该保护的在床者都至少有一条 $\kappa$-free 直接 SBed 源（M×N / HR-RR / pose=Lying）*。现成立（firmware HR/RR-on-enterBed [[radar_hr_rr_bed_enter_gated]] + M×N + 合成 track pose=Lying 三重兜底）。**若哪天这条破了**（深睡 HR/RR 断流 / firmware 床 id 对真睡者漏报 / 合成 track 不再标 pose=Lying），(b) 路即承重，$\kappa$ 衰减就会 FN，gate 须回。这是承重**假设**非铁律，勿无声拆掉三重兜底之一。

### 5. 联合发射 $\Phi$（按 attachment 分轴，离线=中性）

$$\Phi_t=\underbrace{\prod_{j}\ell_{s_j}\!\big(o^{s_j}_t\mid B^j\big)^{w_{s_j}}}_{\text{接触}\to B^j}\cdot\underbrace{\prod_{c}\ell_{c}\!\big(o^{c}_t\mid S\big)^{w_c}}_{\text{雷达 pose}/z/\text{dwell}/\text{hrrr}\to S}$$

$$w_{s_j}=\text{onbed}(s_j,j),\qquad w_{\text{pose}}=\text{covers}(r,\cdot),\qquad w_{\text{hrrr}}=\text{covers}(r,\cdot)\cdot\mathbb 1[\text{nearBed}_t]$$

接触核：

$$\ell_{s_j}(\text{InBed}\mid\text{occ})=L_{\text{in}}\!\gg\!1,\ \ \ell_{s_j}(\text{LeftBed}\mid\text{vac})=L_{\text{left}}\!\gg\!1,\ \ \ell_{s_j}(\text{InBed}\mid\text{vac})=\ell_{s_j}(\text{LeftBed}\mid\text{occ})=1$$

pose 核（二义，刻意）：

$$\ell_{\text{pose}}(\text{lying}\mid\text{AtBed})>1,\qquad \ell_{\text{pose}}(\text{lying}\mid F)>1$$

HR/RR 核（**非对称**，present/absent 都带信息）：

$$\ell_{\text{hrrr}}(\text{present}\mid\text{AtBed})=L_{\text{hr}}>1,\quad \ell_{\text{hrrr}}(\text{present}\mid\neg\text{AtBed})=\tfrac1{L_{\text{hr}}}$$
$$\ell_{\text{hrrr}}(\text{absent}\mid\text{AtBed})=\tfrac1{L_{\text{hr}}},\quad \ell_{\text{hrrr}}(\text{absent}\mid\neg\text{AtBed})=1$$

dwell 核（显式，暴露 identifiability 负担）：

$$\ell_{\text{dwell}}(\text{still}{\ge}\tau\mid F)=\ell_{\text{dwell}}(\text{still}{\ge}\tau\mid\text{AtBed})=D>1,\qquad \ell_{\text{dwell}}(\text{still}\mid O/Sit/\text{walk})<1$$

**含义**：
- **按 attachment 分轴**：接触（InBed/LeftBed/垫 vital）测**床垫** → 挂 $B^j$；雷达（pose/z/dwell/HR-RR）测**人** → 挂 $S$。雷达**测不到垫压**，HR/RR 对 $B$ 的影响经 §4 的 $\Psi$ 涌现，不直接挂 $B$。
- **离线 = 中性**：传感器无上报 $o=\varnothing\Rightarrow\ell\equiv1$（零信息），**不冻结、不永久否决**——这一行消解 1641 离线的硬 occ 钉死。
- **HR/RR 用 nearBed 门控（非 enterBed）+ 非对称似然**（最关键修正）：`enterBed` 是 firmware 返 HR/RR 的**门控状态**，人一离床即退出、HR/RR 转 absent；若按 $\mathbb 1[\text{enterBed}]$ 加权，会在「HR/RR 消失=离床」这个判别信号出现的瞬间把它扔掉。改 `nearBed`（空间邻域、连续）后：present 持续印证 AtBed，**absent 持续否决 AtBed**（不指定替代，$F$/Sit/$O$ 由 pose/dwell/z 分）。
- **dwell 不分 $F$/AtBed**：摔在地和睡在床都「久静止」→ 二者 dwell 都 $>1$；dwell 只分「静止占用 vs 活动」。$F$-vs-AtBed 的判别**全压在 $B$、HR/RR、位置**上——这正是 sleepad 离线 + HR/RR 若被门掉时的危险点。

### 6. 转移 $T$（纯因子化 + $B$ 观测门控自持）

$$T(J\mid J')=T_S(S\mid S')\cdot\prod_j T_{B^j}(B^j\mid B'^j),\qquad T_S=\text{P5 的 }9\times9\ \text{propensity}\ (\textbf{不被 }B'\textbf{ 调制})$$

$$T_{B^j}(b\mid b')=\rho^j_t\,K^{\text{obs}}(b\mid b')+(1-\rho^j_t)\,K^{\text{unobs}}_\lambda(b\mid b'),\qquad \rho^j_t=\mathbb 1[s_j\text{ 在线}]$$

$$K^{\text{obs}}:\ \text{occ}\!\to\!\text{occ}=1-\varepsilon,\ \ \text{vac}\!\to\!\text{occ}=\mu;\qquad K^{\text{unobs}}_\lambda:\ \text{向无信息弛豫,速率}\ \lambda,\quad \boxed{\varepsilon\ll\lambda}$$

**含义**：两条链各自转移，耦合**只在 $\Psi$**（不进 $T$，防双施）。$B$ 的自持被**观测可用性 $\rho$ 门控**：
- 在线（$\rho{=}1$）→ 高自持 $1-\varepsilon$（上下床稀有，防抖）；
- 离线（$\rho{=}0$）→ 以 $\lambda\gg\varepsilon$ 快速弛豫到无信息。**陈旧 InBed 不再永久挂占用**——一个不等式 $\varepsilon\ll\lambda$ 换掉所有 staleness/TTL 补丁。
- 注：$T_S(F\mid\text{AtBed})$ 不依赖 $B'$，「床空更易摔下」的抑制由 $\Psi$ 的 $\psi_{\text{phys}}(F,\text{occ})=\varepsilon_{\text{art}}$ 承担（相容性而非转移率；联合概率结果等价，log 域诊断时两者贡献需分列）。

### 7. 联合滤波（log 域）

$$\bar\alpha_t(S,\{B^j\})=\sum_{S',\{B'^j\}}T_S(S\mid S')\prod_j T_{B^j}(B^j\mid B'^j)\,\alpha_{t-1}(S',\{B'^j\})$$

$$\boxed{\ \alpha_t(S,\{B^j\})\ \propto\ \Psi_t\cdot\Phi_t\cdot\bar\alpha_t,\qquad \textstyle\sum_{S,\{B^j\}}\alpha_t=1\ }$$

**含义**：标准 HMM Predict/Correct，状态空间是 $(S,\{B^j\})$ 的乘积。Predict 可按 $S'$ 分解先内层求 $B$ 链（降常数因子）。$\varepsilon_{\text{art}}$ 极小、$L_{\text{in}}$ 大倍数 → **全程 log 域 + 归一化用 log-sum-exp**（数值稳定）。$|\mathcal B|\le3$（实际家庭）全表即可：$|\mathcal B|{=}2\Rightarrow36$ 态，$|\mathcal B|{=}3\Rightarrow72$ 态，轻量。

### 8. 读出与裁决（含可识别性兜底）

$$P^{F}_t=\sum_{\{B^j\}}\alpha_t(F,\{B^j\})$$

$$\Lambda_t=\frac{\sum_{\{B^j\}}\Psi_t\,\Phi_t\,(F,\{B^j\})}{\sum_{\{B^j\}}\Psi_t\,\Phi_t\,(\text{AtBed},\{B^j\})}\ :\quad \text{informative}\Rightarrow\Lambda\gg1;\quad \text{全暗}\Rightarrow\Lambda\to1$$

$$\boxed{\ \text{fire}\iff P^F_t\,C_{\text{FN}}(\text{risk})>(1-P^F_t)\,C_{\text{FP}}\ \ \text{持续}\ \ge T_{\text{hold}}\ },\qquad C_{\text{FN}}\uparrow\ (\text{单住户独处})$$

$$\delta_{\text{pad/floor}}=D_{\mathrm{KL}}\!\big(P(\text{XY}\mid\text{on-pad})\,\|\,P(\text{XY}\mid\text{床沿地条})\big)\quad(\text{实测前置量})$$

**含义**：
- $\Lambda_t$=当前帧证据能否分开「床边摔」与「睡床上」的似然比。sleepad 在线或 HR/RR armed+absent → $\Lambda\gg1$ 可判；**全暗**（sleepad 离线 ∧ ¬nearBed ∧ $\delta\approx0$）→ $\Lambda\to1$ **物理不可判**。
- 不可判时 belief 据实保持两可，**裁决交风险不对称**：单住户独处 = 最高风险 → $C_{\text{FN}}$ 大 → 低 $P^F$ 也发。安全来自「**诚实的不确定 + 风险不对称**」，不来自假装确信的耦合（[[fall_detection_risk_stratified_design]]）。
- $\delta_{\text{pad/floor}}$=雷达原始 XY 能否分「在垫上 vs 床沿地条」。$\delta\gg0$→ 离线也可判、模型解得了 cd2b；$\delta\approx0$→ 只能风险兜底。**这是声称「修好 cd2b 离线态」前必须实测的唯一悬空输入。**
- $T_{\text{hold}}$（如 90s）防瞬时噪声；与 $\varepsilon$ 有交互（$B$ 慢翻转可能延迟 $F$ 信念），但 sleepad LeftBed（发射）与离线 $\lambda$ 都**绕过**慢 $\varepsilon$，延迟风险仅限「纯无证据」态——须在 cd2b replay 验证。

### 9. cd2b 三态验证（同一组方程，无补丁）

| 场景 | 机制 | 结果 |
|---|---|---|
| 正常睡 | $\ell_s(\text{InBed}\mid\text{occ})$ 强 → $B$ occ；$\psi_{\text{phys}}(F,\text{occ}){=}\varepsilon_{\text{art}}$ 压 $F$；且 $F$ 需正证据，撤否决不造跌倒 | 不误火 |
| 摔下床（sleepad 在线） | $\ell_s(\text{LeftBed}\mid\text{vac})$ 强 → $B$ vac；$\psi_{\text{phys}}(F,\text{vac}){=}1$ → $F$ 通道全开 | 浮出 |
| 摔下床（sleepad 离线=1641） | $\ell_s\equiv1$；HR/RR 灭经 $\ell_{\text{hrrr}}(\text{absent}\mid\text{AtBed}){=}1/L_{\text{hr}}$ 压 AtBed（nearBed armed）→ $S{\to}F$；$\Psi$ 经 $\psi_{\text{phys}}(F,\text{vac}){\gg}\psi_{\text{phys}}(F,\text{occ})$ 拉 $B{\to}$vac；$\lambda$ 漏率 + $\rho{=}0$ 不让陈旧 occ 冻结 | 浮出 |

**不靠 sleepad 路由、不靠 covers/min 代理、不把 HR/RR 挂 $B$、不靠任何专用补丁。**

### 10. 退化与正交扩展

- **单床退化**：$|\mathcal B|{=}1\Rightarrow a_1\!\approx\!1$、$B$ 向量降标量 → 回 18 态 $(S,B)$。
- **一雷达多床**：$(r,s_1)/(r,s_2)$ 各维护 $\kappa_{r,s_1}/\kappa_{r,s_2}$，$\Psi$ 对各 $B^j$ 按各自 $\kappa$ 耦合；时间共现在几何分不开时把跌倒证据**路由到正确的床**。
- **人数 $N_r$ / 真伪 ghost**：同构造换轴——$S^{(i)}$ 多份，$N_r=\sum_i\mathbb 1[S^{(i)}\notin\{E,L\}]$，realness $T^{(i)}$ 由跨 track $\rho$（与 $\kappa$ 同源的共现/镜面）耦合，$P(\text{ghost})$ 从 co-existence 涌现。与本床轴正交，同一滤波形式。（**realness 三类 → 两类 {Real, Mirror}、主职 = $N_r$ 排除 ghost、Static 溶解，见 §G**）

### 11. radar↔sleepad logicID 融合定论（2026-06-20，A+C 同步，对码收敛）

**设备 logicID 不硬合并、按相关性软摊**：每设备各有 track_id/logicID（雷达 track 会 switch 故用 logicID）。跨设备不强行归一（只能归一个=脆），而用相关性 $\text{Con}\in[0,1]$ 软分配——$\text{Con}{=}1$ 同一人完全合并，不确定按概率分。**Con 即本文 $\kappa/a_j/\Psi$，非新机制**：$\text{Con}_{ad}=\kappa_a(\text{房级})\cdot g^{xy}_a(\text{track }d)$，钉具体 track 靠 $g^{xy}$、$\kappa$ 管 sleepad 整体耦合强度。Fall 只由雷达 logicID 判，sleepad 经 Con 抬雷达 logicID 的 SBed（case3 平分 $=\sum_j a_j$ mixture，已在 `LogPsi`）。**case1**：$\text{win}_{15}(\text{InBed}_{\text{sleepad }a},\text{InBed}_{\text{radar }d})\to\text{Con}_{ad}{=}1$（雷达↔sleepad 共发，非两 sleepad，仍被 $g^{xy}$ 门控）。

**真正要新建的只有两件**（cd2b 真解收敛于此，其余现成结构在干活）：

1. **$\kappa$ 变活（wire `UpdateKappa`）** —— 现 orphan（零调用），$\kappa$ 死在几何冷启、只 $g^{xy}$ 半边活。落点：每帧每房在 per-track `LogPsi` loop 前**无条件**调（末尾必调 `UpdateKappa`，防再成 orphan）。**$\kappa$ 两条腿喂，各管各 $\gamma$（C feedback_a5 定稿：双腿不可合并，根在 $\gamma$ 矛盾）**。空间分不开哪张床时**靠时间共现破平局**——雷达只知 `pos`+`是床`，推不出 BedA/BedB（几何平局 $g^{xy}_A{\approx}g^{xy}_B$）；但 sleepad$_A$ InBed 事件与 radar track 入床区**同刻**（±15s）→ 绑同人（靠时间不靠位置）：
- **强化腿**：同向共跳 15s deferred 窗（sleepad 与 radar 床事件同向、窗内）→ 强 $\gamma_{\text{evt}}{=}0.2$，建立关联（时间绑同人）。钉②：只在"sleepad 在床 ∧ radar 持续别床"=真矛盾→强降；"sleepad 跳入 + radar 持续在床（agree）"→不强更新，交维持腿。
- **维持腿**：持续共态 `agree`=sleepad InBed ∧ radar `FwAreaID==`床 $j$（per-frame）→ 弱 $\gamma_{\text{hold}}{\approx}0.01$，维持 $\kappa$ 抗衰减（熟睡持续在床保持高）。`live`=双在线∧至少一方占用（both-vacant→冻结）。
- **钉① lost 边界**：`radarInBed` 不 gate Online（用 M×N 冻结末值），sleepad 半实时把关——人真离床→sleepad LeftBed→agree/matched=F→$\kappa$ 衰减，冻结 N 不误维持。
- 一个 $\gamma$ 做不到"建立要大步快锁、维持要小步稳不抖"→ 两腿不可合并。两腿同更 $c.\kappa$ 叠加：事件快跳、持续慢托。$\kappa$=**纯归属权重**，不碰 SBed 维持（承重不变量见 §4：SBed 由"证据抬 vs 0.80 自衰减"平衡决定，真睡者总有 $\kappa$-free 直接源撑，故 $\kappa$ 升降 FN-safe、**无需防睡眠者 gate**）。

2. **LeftBed → SOpenFloor（组件③，cd2b 主路径）** —— 现 emission 的 LeftBed 只推 $B{\to}$vac，需补：在 $S$ 轴**主动满幅**抬 SOpenFloor，按 $a_j$ 归属（$g^{xy}$ 几何门控：邻床 LeftBed 对在别床 track $g^{xy}{\approx}0\to a_j{\approx}0$ 不串=多住户假摔被几何自然兜住）。落点按 track 态：在场→SOpenFloor / lost→SBlindRest / 先 Open 再 lost→SBlindOpen。「满幅」= 不被 Con 二次折扣（$a_j$ 已分过概率），靠 SOpenFloor 单位量级 $\gg$ SBed 做"进床慢离床快"不对称。

**不用单建的两件**（我曾当补丁加、被砍回——「框架不是补丁」）：

- **latch（牙齿）删**：「LeftBed 一次性 vs 反射每帧抬 SBed」的持久压制**已是结构自带**——$B$ 轴在线核 `kObs` $B\text{vac}{\to}B\text{vac}{=}1{-}\mu{\approx}0.99$（每帧 99% 黏住）、$B\text{vac}{\to}B\text{occ}{=}\mu{\approx}0.01$（只 InBed 似然 $L_{\text{in}}{=}20$ 能扳回），**正是「咬住、只 InBed 能松」**；离线走 `kUnobs` $\lambda$ 漏（半衰期 ~14s）。配 $\Psi(\text{SBed},\text{vac}){=}1{-}o_j$ 每帧压 SBed。LeftBed 的效果经 $B$ 轴是**持久、每帧施加**的，非一次性。
- **floor 不改（⑤）**：`floor.go` `StillSec >= tFloorFor(area)` 是 **belief-独立的天花板**（不看 $P(\text{Fallen})$/不看状态，只 raw 时长 + area + 可观测豁免），专为「belief 判错时」兜底。**「状态驱动 / per-state tFloor」作废**（会把兜底焊回它该兜的 belief 上，纵深防御塌成一层）。床边摔（床矩形外=open cell）→ floor 12min **独立**兜底，哪怕 belief 被 M×N 卡 SBed 也照响；床矩形内 + ③ 失败=残留，靠**实测验证③**（cd2b 重测）不靠改 floor。（M×N firmware 床每帧 $\kappa$-free 重抬 SBed 是 ③/$B$-vac 压制的对手，胜负=量级问题，留实测。）

> **大白话**：cd2b 真解 = ①让"睡垫和雷达是不是一个人"活起来（管贴给谁）+ ③ 人一离床立刻把"可能在地上"拉满（主动放行真摔）。牙齿是 $B$ 轴本来就有的，兜底 floor 本来就独立——都不用动。

---

## 净改动一览（相对当前实现）

| 改 | 内容 |
|---|---|
| 架构 | $O_b$/$N_r$ 从硬外部观测 → 联合隐变量；Zone 仍产 ZoneState 供产品面，DBN 不再吃其结论 |
| 状态 | $(S,\{B^j\})$ 联合，软归属 $a_j$ 代床归属隐维 |
| $\kappa$ | 事件驱动 EMA + 几何初值，**无 max**，互活门控；$K$ 固定调 $\gamma$ |
| $\Psi$ | 物理表经 $\kappa$ 退火、按 $a_j$ 混合、**只施一次**；Fallen 行物理常量不被覆盖 |
| $\Phi$ | 按 attachment 分轴；离线=中性；**HR/RR 用 nearBed + present/absent 非对称似然**；dwell 显式 |
| $T$ | 纯因子化 + $B$ 观测门控自持（$\varepsilon\ll\lambda$）；耦合不进 $T$ |
| 裁决 | $\Lambda_t$ 展开；不可判交风险不对称；前置量 $\delta_{\text{pad/floor}}$ 待实测 |
| 实现 | log 域 + log-sum-exp |

**cd2b 主解（本轮 P-2/P-3 修正旧叙事）**：§5 emission 位置似然 $g^{xy}$（δ≫0 雷达几何可判）；HR/RR 因两源在 cd2b 摔倒段皆空而**后置**（详第三部分 §D）。

---

## 第三部分 · 三卷评审整合（2026-06-14）

> 来源：P6 竞争式三卷评审（[[feedback-p6A]] 项目组 A / [[feedback-p6B]] 评审 B / [[feedback-p6C]] 评审 C）收敛回 ground truth。
> 整合纪律：**框架进本文档；标定（含单 case 实测初值）留 [[feedback-p6C]] 作依据**——本文档只写"待标定项 + 约束/方向/锚"，设计不被单 case 反向绑死。

### §A 新增第四轴：跨房 neighbor 耦合（C §9 发现，A 接受）

第二部分内化了 $O_b$（床）、$N_r$（同房多 track 人数）、ghost（realness），**漏了跨房 neighbor**——现役代码 `belief_neighbor.go` 的 `ObsNeighbor`（本房丢轨后查兄弟房 hand-off、消解 lost-fall 二义）是 belief 外算好的硬结论外挂，与已废的硬 $O_b$ 同病。

框架命题：neighbor 内化为**第四隐轴**，与 §10 ghost/$N_r$ 换轴同构——

| 轴 | 证据 | 耦合 |
|---|---|---|
| 床 $B^j$ | 接触 | → 床占用 |
| 人数 $S^{(i)}$ | 同房多 track | → $N_r$ |
| ghost $T^{(i)}$ | 跨 track ρ | → P(ghost)（**对称**共存）|
| **跨房 neighbor** | **兄弟房 hand-off ρ_xroom** | **→ 本房 $S$ 的 {Left,Empty} vs {Fallen} 路由（新增）**|

- 内化方式：邻房占用不再外挂 `ObsNeighbor` 标量，而作本房 $S$ 转移 $T_S$ 的**跨房门控发射**——`Real-present(本房丢轨) → {Left, 邻房 Real-present}` 转移概率由 ρ_xroom 驱动，belief **联合推断**"人挪去邻房 vs 本房真摔"。
- **有向性（A 校正 C 的"同构"）**：ρ_xroom 与 ghost ρ 仅在"跨实体共现耦合"层面同构；但 hand-off 是**有向时序**（同一人先走后到、新鲜度窗 + 方向），ghost co-existence 是**对称共存**。故 ρ_xroom 必须带**有向门控**（HandoffWindow×Jitter×方向），**不照搬 ghost ρ 的对称形式**。
- sole-resident 门从离散 `rc≠1` 硬 OFF 改**连续衰减**：ρ_xroom 按 resident 数衰减（单住户强、多住户弱不归零）+ 进 §8 的 $C_{FN}$（多住户邻房占用压 fall 的"信用"降，漏报代价仍在）。
- K（现 `dampNbrFallen=0.7`）从似然层固定 damp 常数 → ρ_xroom 驱动的转移概率，与 §3 κ、§10 ghost ρ 同为"几何/事件共现耦合"家族。

下面四式即 C 提出、A 补的方程开口：ρ_xroom 计算式（§A.1，**§38 规则①③ band-pass + 规则② track 守恒校正**）、$T_S$ 跨房门控转移（§A.2）、与 §10 正交扩展接口（§A.3）、时间窗随 unit 自适应（§A.4，**§38 规则④⑤ + §39 D=10min 定稿**）。标定初值（HandoffWindow 60s / Jitter 5s / K 上确界 0.7 / 源型可信度 sleepad 0.9·room-enter 0.8·radar-only 0.2）见 [[feedback-p6C]] §9·§38·§39；本节只立**框架与方向/符号**，曲线参数（$\tau_p,\delta_0,\tau_j,\beta$）留 oracle 标定（"band-pass 先升后降、峰 1-5s"的形状定，曲线不标定）。

#### §A.1 ρ_xroom 计算式（有向 hand-off 耦合）

记号：本房 $r$、同 unit 兄弟房集 $\mathcal N(r)$；本房 real-present track 消失时刻 $t^{\text{lost}}_r$（**取 last-observed，见下「lost 锚点」校正**）；兄弟房 hand-off 落点事件时刻 $t^{\text{arr}}_{r'}$（EnterRoom ∨ InBed 翻转）；滞后

$$\Delta_{r'}=t^{\text{arr}}_{r'}-t^{\text{lost}}_r\qquad(\Delta>0=\text{先走后到}=\text{有向命中})$$

**lost 锚点（§7.7 v2 意图 + 实测现状）**：

- **设计意图**：$t^{\text{lost}}_r$ 取该轨**最后一次真被观测在场的时戳**，**不是系统「确认丢轨／进 Blind」那一帧**。后者迟到又抖——要等若干帧不再匹配 + coast 跑完才落定，叠加跨流抖动（sleepad 比 radar 晚 38–807ms），会把**最常见的 1–3s 正常过门接力**压成 $\Delta\approx0$ 撞低端门（$\Delta\le0$ 拒 ／ band-pass $\Delta{=}0$ 谷）误杀 = FP。意图是把 $\Delta$ 拉回 $(0,W]$ 正区间让过门接力正常放行。

- **实现现状（已知偏差，决定保持不修）**：锚点取自 `lastSeenMs`，它在每个 census `Present==true` 帧刷新，而 `Present = nowMs − LastObservedMs < presenceCoastMs`（`track_manager.go:992`，**1200ms** coast 容差，本为吸收跨流漏帧）。故真实锚 = **最后一个 coast-present 帧**，比真观测**偏晚 ≤ presenceCoastMs ≈ 1.2s**。d523 实证（4X 重放可复现，数据时戳驱动与速度无关）：真观测 `13:35:05.387`（`D523.0 stand`）、引擎实锚 `13:35:06.445`（`tid=88` no-target coast 帧、位置冻结）、偏 **1058ms**；`pending_lost_ms` 即记此值。
  - **上界严格 = $\min(\text{presenceCoastMs},\ \text{帧间隔})$**：帧稀时（firmware no-target `tid=88` 周期 ~30s）coast 窗内无帧落入 → 锚误差归 0；那 ~30s 是**丢轨检测延迟**（下一帧何时到），与锚误差**正交**，不会让锚偏到 30s。
  - **方向 FN-safe，影响小 → 不修**：偏晚 → $\Delta$ 偏小 → 更易拒接力 → 不抑制 lost-fall → 倾向报警；代价仅"离房后 $<1.2$s 即在邻房冒头"的快接力多一次 FP，**不漏真摔**，且 $1.2\text{s}\ll W(90\text{–}150\text{s})/D(14\text{min})$。修法（锚改真 fresh 判据 `LastObservedMs==nowMs`，已存在于 `track_manager.go:939` 旁）成本小但收益仅消一个 FN-safe 系统偏差，**评估后保持现状**。

- **落点**：从该锚点**向后**在 $\mathcal N(r)$ 找 $\Delta N_r>0$（兄弟房「新增 $+1$ real track」的守恒重现，§A.1(b)）= hand-off 落点；**source-agnostic**：radar 新轨 ∨ sleepad InBed 合成轨同等算数（只决定快窗／慢窗，不决定算不算一个 gain）。

**(a) 有向新鲜度核**（与 ghost 对称核的关键分野——对 $\operatorname{sign}\Delta$ 不对称；**§38 规则③ band-pass 校正**：旧式 $e^{-\Delta/\tau_h}$ 在 $\Delta{=}0$ 取峰是**错的**——同 tick 两房 = 人没法瞬移过去 = 不可能是同一人 hand-off）：

$$w^{\text{dir}}(\Delta)=\begin{cases}
g\!\big((\Delta+\delta_0)/\tau_p\big) & 0\le\Delta\le W & (\text{先升后降；}\Delta{=}0\text{ 压低但非零，峰在 }\Delta{=}\tau_p{-}\delta_0\approx1\text{-}5\text{s})\\[2pt]
g(\delta_0/\tau_p)\,e^{\Delta/\tau_j} & -J\le\Delta<0 & (\text{抖动反向余量，}\tau_j\text{ 仅容时钟噪声})\\[2pt]
0 & \text{otherwise} & (\text{陈旧 / 真反向}\to\text{非 hand-off，不压 fall})
\end{cases}$$

其中 $g(x)=x\,e^{1-x}$（band-pass 基形：$x{=}0\to0$、峰 $=1$ at $x{=}1$、$x{>}1$ 衰减），$\delta_0$ 小偏移令 $w(0)>0$。**物理（规则①③）**：$\Delta{\approx}0$ 同 tick = 瞬移 = 可疑（压低）；$\Delta{=}1\text{-}5\text{s}$ = 老人正常走过去 = 峰（越近越强）；之后越久越可能别的事 → 衰减。$W$=HandoffWindow、$J$=Jitter（5s）。窗外 = 0：stale「上次在哪」证不了此刻在哪（人可能穿盲区真摔），铁律 [[partial_monitoring_fall_suppression_law]]。曲线（$\tau_p,\delta_0,\tau_j$）留 oracle form-anchor（铁律 [[fall_data_is_artificial_test]]），**"先升后降、峰 1-5s"这个形状是定的**。

**(b) 归因可信加权、去 ghost 的兄弟房占用**（belief 据此可**质疑**邻房归因，非吃算好的标量）：

$$q_{r'}=c_{\text{attr}}(r')\cdot P_{r'}(\text{real-present}),\qquad P_{r'}(\text{real-present})=\Big(\!\!\sum_{S\notin\{E,L\}}\!\!P_{r'}(S)\Big)\cdot\big(1-P_{r'}(\text{ghost})\big)$$

$c_{\text{attr}}$=源型可信度（sleepad InBed 接触式 0.9 / radar room-enter 过门事件 0.8 / radar-only 占用 0.2）。$P_{r'}(\text{real-present})$ 吃兄弟房**去 ghost 后**的占用后验（§A.3 接口①）——兄弟房一个 ghost 不算合法 hand-off 落点。

**§38 规则② track 守恒**（hand-off 权威信号，非裸"兄弟房有人"）：短时 unit track 数恒定，本房 $-1$ track ↔ 兄弟房 $+1$ track = 同一人挪过去（doc P6.5「人在 X 丢了必在别处冒出来」）。故 $P_{r'}(\text{real-present})$ 取的是兄弟房**新增 $+1$ real track 的守恒重现后验** $=P($丢的人在此重现$)$，载体复用 SuiteCensus/P_id 跨区 track 账（别另起）。

**(c) sole-resident 连续衰减**（替离散 `rc≠1` 硬 OFF）：

$$\eta(\text{rc})=e^{-\beta(\text{rc}-1)}\quad\Rightarrow\quad \text{rc}{=}1\!\Rightarrow\!\eta{=}1\ (\text{单住户强}),\ \ \text{rc}{>}1\!\Rightarrow\!\eta\in(0,1)\ (\text{弱但不归零})$$

**合成**（N-4 单合并：至多取最强命中，单 conf 不相乘；sole-resident「人只在一处」前提）：

$$\boxed{\ \rho^{\text{xr}}_t=\eta(\text{rc})\cdot\max_{r'\in\mathcal N(r)}\big[\,w^{\text{dir}}(\Delta_{r'})\cdot q_{r'}\,\big]\ \in[0,1)\ }$$

**有向性对照（A 校正 C 的「同构」）**：$\kappa$（§3）= 对称互活 co-liveness（双方在线即更新、无时序）；ghost $\rho$（§10）= 对称 co-existence（两 track 同步移动，交换 $i\!\leftrightarrow\!j$ 不变）；**$\rho^{\text{xr}}$ = 有向**——源房**先**丢轨、宿房**后**占用，交换源/宿是不同事件，$w^{\text{dir}}$ 对 $\operatorname{sign}\Delta$ 不对称。三者同属「几何/事件共现耦合」家族，但只有 neighbor 带时序方向，**不照搬对称核**。

#### §A.2 $T_S$ 跨房门控转移

ρ_xroom **仅在 lost-track**（本房 real-present 消失、即处于 Blind* 行）激活，把 Blind 行的 $\to F$ 倾向按 ρ_xroom 改向 $\to L$（人挪去邻房而非本房真摔）：

$$\tilde T_S(F\mid S')=(1-\rho^{\text{xr}}_t)\,T^0_S(F\mid S'),\qquad \tilde T_S(L\mid S')=T^0_S(L\mid S')+\rho^{\text{xr}}_t\,T^0_S(F\mid S')$$

$$S'\in\{\text{BlindRest},\text{BlindOpen}\};\quad \text{行其余项不变，改向后按行归一}$$

- $\rho^{\text{xr}}{=}0$（无新鲜有向 hand-off）→ 行不变 → Blind 照常 ramp Fallen（**lost-fall 本义保留 = 安全默认**：stale / 无事件 / 多住户归因弱 → 不抑制）。
- $\rho^{\text{xr}}{\to}1$（fresh sleepad hand-off + 单住户）→ Fallen 倾向整流入 Left → 不造 phantom fall。
- $K$（旧 `dampNbrFallen=0.7` 似然层固定 damp 常数）= 现 $\rho^{\text{xr}}$ 上确界（fresh·sleepad·单住户 $\approx0.9$）由证据涌现，从「标定常数」变「hand-off 证据驱动的转移概率」，与 §3 $\kappa$、§10 ghost $\rho$ 同源。
- 注：改的是 $T_S$ **行内** $F\!\leftrightarrow\!L$ 改向（耦合进转移，**不进** $\Psi$，与 §6「耦合只在 $\Psi$」并不冲突——那条约束针对**床轴 $B$ 与 $S$ 的同帧相容**；neighbor 是 $S$ 轴自身的跨房转移先验，属 $T_S$ 本职，非 $B$–$S$ 双施）。

#### §A.3 与 §10 正交扩展接口

| # | 接口 | 内容 |
|---|---|---|
| ① | ghost 轴 → neighbor 轴 | $q_{r'}$ 吃兄弟房 $P_{r'}(\text{real-present})\cdot(1{-}P(\text{ghost}))$——§10 房内 ghost 后验喂 §A 房间 hand-off；兄弟房 ghost 不算落点。**注（§G 澄清）**：此处刻意用 `real-present`（有没有真人可接），**非** $N_r$（几个）——hand-off 落点只问兄弟房能不能接一个人；$C_{FN}$ 折扣才读 $N_r$。两消费者要的量不同，非冲突 |
| ② | 不加隐维（状态空间不爆） | $B^j$/$S^{(i)}$/$T^{(i)}$ 是房内 $J$ 的**隐维复制**；$\rho^{\text{xr}}$ 是房**间** $T_S$ 的**转移耦合**，吃兄弟房 belief 读出标量，**不进本房 $J$ 基数**（$9\cdot2^{|\mathcal B|}$ 不变） |
| ③ | 同 census 双消费 | $\eta(\text{rc})$（§A.1）与 §8 $C_{FN}$ 的 resident 数同源：多住户既弱化 $\rho^{\text{xr}}$ 归因（belief 更不确定、$P^F$ 不被压死）**又**折扣 $C_{FN}$（漏报代价降不归零）——两层一致下拉，§B「期望损失主框架 + 证据层」分工保持：$\rho^{\text{xr}}$ 路由 belief（发生了什么），$C_{FN}$ 裁代价不对称（残余 $P^F$ 报不报） |

#### §A.4 时间窗随 unit 自适应（§38 规则④⑤ + §39 D=10min 定稿）

两个时间窗，两个相反的 unit 依赖（非固定 60s）：

**(1) hand-off 检测窗 $W$（规则④，随公共度收小）**：unit 越公共、陌生人流越大 → $W$ 越小。理由：陌生人多 → 「本房丢 + 兄弟房冒」的巧合共现增多 → 收窄窗防把无关陌生人误判 hand-off。
$$W(\text{publicness})=W_0\,(1-\alpha\cdot\text{publicness}),\quad \text{publicness}\in[0,1]\ (0{=}\text{私有 suite},1{=}\text{楼层公共，决定24 standalone})$$

**(2) 延迟裁决窗 $D$（规则⑤ + §39 定值，随覆盖差放长）**：lost-track 后等 $D$ 仍无 hand-off/守恒重现 → fire。覆盖差（设备不足、盲区多）→ 二义只能靠 neighbor → $D$ 放长。

$D$ **锚在静止消失门限**（同一物理时标，用户裁定）：静止门限和 $D$ 都在答「track 消失 = 走了 vs 真摔被降功率滤掉」的同一类二义；底下是同一时钟——雷达功率自适应静止过滤。故
$$D=\text{stillVanish}+(1-\text{coverage})\cdot\text{margin},\qquad \boxed{D_{\max}=8\text{min（静止门限）}+2\text{min（余量）}=10\text{min}}$$

**8 vs 10 = 不卡物理边界的层层留余量**：5min（物理：人静止被降功率滤掉）→ 8min 门限（+3min 防「刚站定就误判」）→ 10min 延迟窗（+2min 给「消失后再确认无迟到 hand-off」缓冲）。$D{=}8$min 是边界重合（窗结束 = 人刚被滤，无观察缓冲，最脆）；$D{=}10$min 错开 2min 留确认缓冲。下界 5min（覆盖好/人流大无需久等）、上界 10min（覆盖差/二义多）。$D$ 由 decide/lost-fall 层消费，非 $\rho^{\text{xr}}$ 本身。

### §B 裁决定位：期望损失是主框架，emission/dwell/neighbor 是证据层（C §7 修正）

§8 的 fire 条件 $P^F C_{FN}(\text{risk}) > (1-P^F)C_{FP}$ 是裁决层**主框架（恒在）**；§5 emission（含 $g^{xy}$）、dwell、§A neighbor 均为**供 $P^F$ 的证据层**，非与裁决并列的解。

- 二义性靠"代价不对称"裁决、非"把某值算更高"：cd2b 段 $P^F$ 卡 0.4、与 SBed 0.45 纠缠，argmax → SBed 赢 → 漏；期望损失 → 独居 $C_{FN}$ 巨大 → 0.4 翻转 → fire。**不需 $P^F$ 赢，只需代价翻转。**
- 框架约束：现 `risk_evaluator.go` 离散三档 RiskLevel **不能直接喂**此连续不等式；decide 层须有 $C_{FN}(\text{risk})$ 连续代价函数消费风险因子（独居/夜/人数/失能）。代价函数"存在且连续"是框架，其参数曲线是标定（[[feedback-p6C]] §8）。

### §C 转移 $K^{unobs}_\lambda$ 取定义 B（泄漏，vacant 吸收）（B2，A+C 锁定）

离线弛豫核取**定义 B（occ 泄漏到 vac、vac 吸收）**，沿用 §6 $K^{obs}$ 的箭头记法（**不用矩阵，避免行列惯例 + 状态顺序双重歧义**）：

$$K^{unobs}_\lambda:\quad \text{occ}\to\text{vac}=\lambda,\quad \text{occ}\to\text{occ}=1-\lambda,\quad \text{vac}\to\text{vac}=1,\quad \text{vac}\to\text{occ}=0\qquad(\varepsilon\ll\lambda)$$

理由：养老院"床垫离线期间无人上床"为常态；vac 吸收避免空房离线被弛豫成 P(occ)=0.5 伪占用（cd2b 根方向）。离线时真有人上床 → 靠 radar pose/位置经 §4 Ψ 涌现，不靠 $B$ 链自发升 occ。λ 半衰期 < 离线判定窗（~30s，复现原 staleness），标定见 [[feedback-p6C]] §5。

### §D §5 HR/RR：优先级后置 + absent 须 sleepad-online gate（P-2/P-3/B3）

**优先级后置（推翻第二部分末"HR/RR 最高优先级"旧叙事）**：cd2b 摔倒段 HR/RR 两源皆空（radar enter-gate 不返 + sleepad 离线），HR/RR 闸门对 cd2b 结构性零作用。**cd2b 主解是 §5 emission 位置似然 $g^{xy}$（δ≫0 雷达几何可判），不是 HR/RR、不是风险兜底。** HR/RR 服务"床附近仍有 vital"的别场景，优先级后置。

**absent 须 sleepad-online gate（B3）**：HR/RR absent 否决分支 `ℓ_hrrr(absent|AtBed)=1/L_hr` 须 gate 在"房内有独立在线 vital 源（sleepad）"之下。radar 自身 HR/RR absent **不得**作否决 AtBed 的证据——radar 在床位被 firmware enter-gate、结构性不返 vital（[[radar_hr_rr_bed_enter_gated]]），其 absent 是零信息非"不在床"。实证：cd2b 在床 558s 段 radar HR/RR 缺失 100%（573 帧 0 vital），不 gate 则每在床帧误推 F = 100% FP。

### §E §4 Ψ 取 mixture（加权和），理由 = FN-safe（B4，A+B+C 共识）

多床 $\Psi=\sum_j a_j\tilde\psi_j+a_\varnothing$（mixture）非 $\prod_j\tilde\psi_j$（product）。两点理由：

① **物理**：人只占一床（不能同时占两床），mixture 软归属语义正确。
② **真摔安全（FN-safe）**：product 下**任一床 occ 即把 F 乘到 $\varepsilon_{art}$**——人在 ambiguous 床位（$a_A\approx a_B\approx0.5$）摔、其中一床陈旧 occ 时 $\Psi(F)=\varepsilon_{art}\cdot1=\varepsilon_{art}$ 被压死 = **漏报（FN）**；mixture $\Psi(F)\approx0.5\varepsilon_{art}+0.5\approx0.5$ 保 F 竞争。养老院风险不对称下选 mixture 防漏报。

> 校正记录：C 初版（[[feedback-p6C]] §10）曾写"product → phantom fall（FP）"，因果反了——product 过度压 F 是 **FN** 风险；phantom fall（多床 occ 时 $a_\varnothing$ 保住 F 的误报）反是 mixture 特性。A+B 数值独立验证一致，C §11 已自我反省、作废 §10 B4 论证。

### §F μ=ε 对称默认 + 记号/量级（B1/C1/C3，标定锚）

- **μ=ε（对称）默认**：$K^{obs}$ 的 vac→occ 速率 μ 与 occ→occ 自持补 ε 取对称——无数据支持"进床比离床易"的非对称漂移（cd2b 413s 仅 1 翻转），无证据不引偏置；未来 case 若显示上床事件系统性多于离床再放开。
- **记号 C1**：§4 的 $o_j$ ≡ §2 的 $o_b(r,s_j)$（床 $j$ 的 overlap）。
- **$w_{\text{pose}}$ 取定 C2（B 第2轮裁定）**：§5 的 $w_{\text{pose}}/w_{\text{dwell}}/w_{\text{hrrr}}$ 的 $\text{covers}(r,\cdot)$ 多床下取 **$\max_j\text{covers}(r,j)$**（实现 `geom0Covers()`），理由 = **保持 $\Phi$ 的 S/B 分轴清洁**（雷达轴只挂 $S$、不按床 $j$ 分解，否则 AtBed 用对应床 covers、$F$ 用 max covers 会把床轴拖进 $S$ 发射）。已知代价：多床房**雷达 coverage 高估**（被全覆盖的床"借" coverage 给弱覆盖床 → 弱覆盖床的 pose/dwell 弱信号被满权放大，可能推高其 $F$/AtBed 信念）。**触发条件式退路**：后续多床 case 若出现此高估致 FP，再引入 per-state covers（AtBed↔对应床、$F$↔max）。当前单 case 无多床证据，保 max。
- **$\varepsilon_{art}$ 量级 C3**（标定）：用"在床段 pose=Lying 帧占比"反推（压得住床上翻身误读、不被 $\Phi$ 正向似然在 log 域淹没），与 $L_{in}$ 联合定，非凭空 $10^{-3}$。

### §G realness 轴重定 — Real vs Mirror（Static 溶解）+ $N_r$ 排除为主职 + 金属三场景分层（2026-06-14 框架对齐）

> 来源：项目组与架构师链上对齐（非单 case 反推）。修正 §10 的 realness 三类 {Real, Mirror, Static} → **两类 {Real, Mirror}**；并把 realness 的主职从「压 ghost 自己的摔」改正为「**把 ghost 排除出人数 $N_r$**」。两固件时标由架构师 domain 拍定，本节锚。

**一、Static 溶解（删第三类）**。孤立纯金属反射体这一类**不该活在本层**，两条独立理由：
1. **firmware 已在底下处理**：无人时 firmware **30s 内过滤纯金属回波**（= 无目标 ID=88 态，[[AI_health]]「1s 有人 / 30s 无人 ID=88」）。孤立静止金属不形成持续 track 传到本层——Static 类当初要防的场景，硬件已防死。
2. **静止判 Static = 病根**：用「困 BirthPos + 久静 + 近墙」判 Static，**分不开静止金属与摔倒静止的人**（签名同）。单 track 静止判 ghost = [[fall_detection_risk_stratified_design]] 的病根，会误抑制无走动的床边真摔。删。

**二、realness 主职 = $N_r$ 排除 ghost（非压 ghost 的摔）**。危险不在 ghost 自己 false-fire，在 **ghost 被当成第 2 个人 → $N_r$ 虚增 → 独处真人风险被误降**：decide（§8/§26）对 `PeopleCount>1` 折扣 $C_{FN}$（×1/N，仅 45–55% 两可窗，「有人代发现」）。真人其实独处却被当 2 人 → 该报的边界摔不报 → **FN**。故 realness 要保证 $N_r$ 只数**不同的真人**，排除 mirror（同人反射）/金属。`PRoomHasReal`（有没有真人）≠ $N_r$（几个真人）。**两量各有消费者、非二选一**（C §G 复审澄清）：$C_{FN}$ 折扣读 $N_r$「几个」；§A.1(b) neighbor hand-off 落点判定读 `PRoomHasReal`「兄弟房有没有真人可接」——各取所需，见 §A.3 接口①注。

**三、两轴 FN-safe 非对称默认（「一切看风险」落地）**。realness 不确定时，两根轴各自往**高 fall-risk** 倒：

| 轴 | 暧昧默认 | FN-safe 因 |
|---|---|---|
| 它自己的 fall（pFallReal 调制 SFallen 发射）| **不压 $\equiv1$**，仅正向 ghost 证据才 $<1$ | 万一真人在摔，别抑制 |
| 它当别人的「陪护」（$N_r$ 计数）| **不计入**，仅正向「独立真人」证据才计入 | 万一非真人，别替别人折 $C_{FN}$ |

> 正向证据：ghost = co-existence $\rho$（mirror 共动，§10/§A.1）或 cell-AreaDeny 拓扑（外生只读层）；独立真人 = 独立入口出生 + 独立运动史 / 非共现几何。**「无 ghost 证据」≠「是 ghost」**——realness.go 现有 leak 向均匀漂、零证据时凭空造 ghost 质量（$P(\text{real})$ 0.5→0.33），违「$P(\text{ghost})$ 从 co-existence 涌现」（§10），是**非框架残留，待码层改**。

**四、金属三场景分层（互不重叠，每场景都有人管）**：

| 场景 | 处理层 |
|---|---|
| 空房纯金属 | **firmware 30s 过滤**（无目标 ID=88，不归本层）|
| 有人 + 金属 | **co-existence（$\ge2$ track）→ 共现排除出 $N_r$**（Mirror/$\rho$ + 计数保守默认）|
| 静止/摔倒的人**消失**（~5min 降功率滤）| **Blind 子态 + neighbor $D{=}10$min 闸**（§A，同一物理时标）|

**五、两固件时标锚（架构师 domain 拍定，文档直接锚，非待标定）**：
- **30s**：无人时 firmware 过滤纯金属回波（= 无目标 ID=88 态）。
- **~5min**：静止的*人*被降功率滤掉（已锚在 §A / $D{=}10$min 的 5min 物理底）。
- 二者**不同场景**：30s 治无人金属，~5min 治静止人消失的二义；勿混。

**六、身份/连续性 + ghost 作用域**（订正：原写「软边缘化、非 logicID」是过度设计——track 关联用最近距离是标准做法，非 §47 那个病，改回最小作功距离 logicID）。
- **身份关联（哪条人影是哪个人）= 最小作功距离 logicID**：新 track 出现发新 logicID；平时各 track 分得开、logicID 跟随不变；仅两 track 相遇（≤50cm 可能认错）时按**最小作功距离**保号（新坐标离上一 tick 哪条 track 近，就接哪条的 logicID——人不瞬移，最近即同一人）。这是标准 track 关联**预处理**，**非** §47 担心的「硬结论外挂」病：那病是把占用**结论**硬塞进 DBN 当观测；track 关联只决定观测归哪条 track，不进隐状态。
- **消失续存 = $S^{(i)}$ 转移自持**：track 消失（摔倒静止降功率被滤）时，该 track 的 $S^{(i)}$ 经转移自持（Fallen 留 Fallen + Blind 携占用），不靠重关联续命。
- **ghost（mirror）仅在 track 数 $=2$ 时作用**：1 track 不进 mirror 分支；正好 2 track 跑 mirror 判别（co-existence $\rho$ + 反射几何，排除影子出 $N_r$）；**3+ track 不处理**（人多、风险低，不为它把 mirror 逻辑复杂化）。**（注：「1track 永发」只对 mirror/静止——单 track 运动伪迹 ghost 是另一分支、可判，见 §G七）**

**七、ghost 判别信号体系（§53 修正 + 数量×时间，非硬 gate）**。§G六 的「ghost 仅 track==2」是指 **mirror（成对 co-existence）**；ghost 判别另有**单 track** 分支与之正交。三分：

- **桶一 · 单 track 运动伪迹 ghost（census/logicID 层从 raw XY 即可算，不等 MM）**。信号 = ① 速度超室内合理上限（老人走不了那么快）② 轨迹跳跃/瞬移（固件 track-swap 伪迹）。**判据用数量×时间累积、非硬阈**：不设「speed≥X→ghost」硬 gate（标定陷阱，[[fall_data_is_artificial_test]]）；每帧异常**量**（超速幅度/跳跃幅度）随**时间累积**成连续 ghost 后验（同 mirror $m$Score、§3 $\kappa$ EMA 范式——「越持续/越频繁，信心越强」）。持续越快/跳越多 → ghost 信心越高；偶发一帧异常（噪声/真人快走一步）≈ 不判。
  - **FN-safe**：① 持续快移 = ghost 或护工，都在 fall 保护圈外，敢判（护工不靠系统兜底，真摔在场即被发现）；② 累积避免真人偶发噪声误判 ghost。
  - **细化「1track 永发」**：1track **静止**永不判 ghost（病根：分不开金属/摔倒静止真人）；1track **持续快移/瞬移**可判 ghost（FN-safe）。
  - **★ 独立分量 + 共生律不污染（§54 框架补）**：桶一伪迹 ghost 是 realness **独立分量**（独立 score，**非** mirror 的 $m$Score）、**对 `PRoomHasReal` 不贡献真人源**——伪迹是固件 track-swap/超速，**无真人源**（区别 mirror 必有真人源）。两类 ghost 下游贡献不同：**`pFallReal`（压自己的摔）两类都压、$N_r$ 两类都排**；但 **`PRoomHasReal`（房内有无真人）只 mirror 贡献、伪迹不贡献**。码层若把桶一塞进同一 $m$Score → 语义错（超速伪迹被 `PMirror` 报成某真人镜像）+ 共生律污染（伪迹被当蕴含真人 → 房内有真人后验虚高 → 扰动「丢真人 track 但镜像存活→fall 不抑制」）。
- **桶二 · 出生地判据（入口/墙外/金属点）= 几何依赖**。判「墙外」要墙多边形、「入口」要入口几何、「金属点」要 cell-AreaDeny——**非 census-now**，随 MM/cell/layout 接通才算（[[mm_relationship_matrix]]）。
- **桶三 · metal 的 still-box 久静判 = 病根，保持删**。静止分不开金属 vs 摔倒静止的真人（§G 一），请回来 = 复活 FN。metal 归 **firmware 30s（§G 五）+ cell-AreaDeny 拓扑**，不归 census 静止。

> §53 修正（C 复审 census 单点依赖）：现 census 只算 mirror 的 $\rho$、等 MM 的 IsReflection；**桶一**这套单 track 运动伪迹判据 census 层即可算、不依赖 MM——故「ghost 全押 MM」是**当前实现未兑现桶一**，非框架死结。补本节立框架后由 census 兑现累积式判据。

---

### 整合后轴/裁决一览

| 轴 | 证据 | 状态 |
|---|---|---|
| 空间（§5 $g^{xy}$）| 雷达 XY | 已内化；δ 标定、脆弱（[[feedback-p6C]] 附）|
| 时间（dwell 符号）| 久静 × cell 容忍 | 已内化（survival）；cell 容忍可靠性 = 框架前提 |
| 裁决（§8 $C_{FN}$）| 风险因子 | 主框架；$C_{FN}$ 连续代价函数须 decide 落地 |
| 跨房（§A neighbor ρ_xroom）| 兄弟房**有向** hand-off | 框架命题 + **方程已落（§A.1 ρ_xroom / §A.2 $T_S$ 门控 / §A.3 §10 接口）**；曲线参数待标定 |
| realness（§G Real vs Mirror）| mirror: co-existence $\rho$+反射几何（成对）／ 单 track 伪迹: speed·跳跃（数量×时间，§G七）／ 出生地·metal: 几何·拓扑 | **框架重定（§G）**：主职 = $N_r$ 排除 ghost；两轴 FN-safe 默认；Static 溶解；ghost 信号三分（§G七：桶一 census 兑现/桶二 几何依赖/桶三 病根删）；leak 凭空造 ghost 已码层删 |

---

### §H per-area 静止时长统计标定 — floor / emission 单源 $(\mu,\sigma)$（2026-06-18，用户定）

**问题**：「久静多久算摔」的时长门限原本三处各漂——floor `tFloor`、D/DU `udLenFor`、`ThresholdNonRest` 各拍各的数；且 emission 当初否决 still-box 进发射（全局硬阈 `stillTau=60` 致 d523 站立伪迹 $P^F=0.945$ FP，§十四/十五）。**根因 = 缺 per-area 标定**：一个全局阈对伪迹和真摔一视同仁。

**标定法（高斯分位）**：各区「正常停留时长」$\sim$ 高斯 $(\mu,\sigma)$，**异常阈 $=\mu+1.5\sigma$**（上侧分位）。
- $\pm 1.5\sigma$ 含 **86.64%**；只取**上侧**（停留太久=疑似摔；太短=人正常离开非异常）→ 单侧上尾 $P(Z>1.5)=6.68\%$ = **异常阈自带 FP 率**。
- 选 $1.5\sigma$ 而非 $2\sigma$（上尾 2.3%）：高危跌倒宁多扰、少漏（知情 FP）。
- $\times 1.5$（均值倍）$=\mu+1.5\sigma$ 仅当 $\sigma=\mu/3$（CV=1/3）；无真 $\sigma$ 时以此近似。

**per-area $(\mu,\sigma)$ 与数据来源**（铁律 [[fall_data_is_artificial_test]]：无真摔数据标定，留 oracle）：

| 区 | $\mu$ | $\sigma$ | 异常阈 $\mu+1.5\sigma$ | 依据 |
|---|---|---|---|---|
| **Bath**（toilet/shower）| 12min | **4min（真 σ）** | **18min** | 医学建议正常 5–10min；健康人 >20min 仅 0.5%($\approx\mu+2\sigma$) 反推 $\sigma\approx4$；老人便秘倾向取 $\mu=12$；18min 覆盖 ~80% 便秘、severe 20min($\mu+2\sigma$)。**文献硬支撑** |
| **Sit·Lying**（含 Bed）| 60min | 20min($\mu/3$) | **90min** | 文献久坐单次 MBD 11.7–16min、prolonged ≥30min；$\mu=60$ 取久坐久卧「容忍上限」(看电视/午睡，FN-safe 高于平均、低危别老打扰)。**保守** |
| **default**（开阔/站立）| 8min | 2.67min($\mu/3$) | **12min** | 无直接文献（非标准观测场景），类比久坐短 bout <10min。**经验，oracle 待真实数据调** |

数据来源：CNN「马桶<10min」· anorectal PMC12669168（>20min 0.5% 健康 vs 8.1% 病）· sedentary PMC8679788（MBD 11.7–16）· stroke PMC9166254（≥17min 高危）。

**三处单源 + emission 内化**：
- **floor / stillbox 计时器**（已落 `belief/floor.go`）：`tFloorFor(area,room)=μ+1.5σ`，`(μ,σ)` 由 `stillMuSigma` room×cell 保守合并（§I）。Bed 并入 Sit·Lying 档。**原独立 D/DU 决断窗已退役并入此**（§I 合体）——不再有 `udLenFor`/per-room deadline/房型分 D-DU。
- **emission 高斯 CDF**（规划，取代「floor 独立 fire + still 不进发射」）：$\text{SFallen 贡献}=\Phi\big((\text{StillboxSec}-\mu)/\sigma\big)$，正常值$\approx0.5$、异常阈$\approx0.93$。per-area $(\mu,\sigma)$ 完全决定曲线，$k$ 不再单独拍。**解 §十四/十五「still 进发射单向爬 FP」**——伪迹区 $(\mu,\sigma)$ 大则推得慢、到不了 0.85。
- **床边跌倒**：emission $\Phi(\cdot)$ 须配 sleepad 接触轴（真 InBed 压 SFallen / LeftBed 放行 radar 假阳 InBed，[[bed_stale_leftbed_vetoes_radar_inbed]]）——$(\mu,\sigma)$ 单独搞不定接触假阳。

**裁决统一**：emission 各状态后验赛跑 $\ge0.85$ 抢先发（belief，无二义性）；floor / stillbox 计时器 `StillSec≥tFloor` 时长兜底发 SFallen（CDF 被假 InBed/area 压住到不了 0.85 时）；二者同 stillbox / 同 `tFloorFor` / 同 `StillSec==0` 清（§I 合体）。

**落地进度（2026-06-18，commit fb0782d）**：emission 高斯 CDF **已落**（`belief/emission.go`，独立于 RadarOnline——lost 续算 StillSec 在消失态仍喂；权重 `lStill`=0.1 待标定）；`(μ,σ)` **room×cell 保守合并已落**（`stillMuSigma` 取 max μ，bathroom 未画 toilet 也用 bathsec）；**§I 合体已落**：D/DU 决断窗退役、floor 成唯一 stillbox 计时器、`tFloorFor` room×cell。实测 d5f7-0617：SFallen 0.291→0.979，fire=40（floor 7 + lost 33），首 floor-fire 18.3min（room×cell 生效），belief lost fire 主导。

---

### §I floor / D-DU 合体 — 统一 stillbox 计时器（2026-06-18，用户定；commit fb0782d 已落）

**洞察**：floor 与 D/DU **本是同一件事的两个化身**——都在答「静止/失踪 tFloor 久没人接管 → 兜底发 SFallen」。唯一区别是计时基准：floor 用 `StillSec`（静止起，含 present），D/DU 用 `lostMs`（失踪起）。① 的 lost 续算让 `StillSec` 跨 present/lost 不断——floor 的 `StillSec` 已天然覆盖 D/DU 的 lost 段，D/DU 多余。

**合体（落地形态）**：**D/DU 决断窗退役**，`belief/floor.go` 的 `FloorGuard`（per-track，`StillSec≥tFloorFor`→发 SFallen）成**唯一 stillbox 计时器**。`StillSec>0` 自然计时、`StillSec==0`（移动/回来）自然清零，**无独立 deadline 状态**（删 `engine/unit.go` 的 `udDeadline`/`udLenFor`/per-room timer）。

**合体后架构 — 一个 stillbox 喂两个消费者**：
```
        stillbox 时长 (per-track, 含 lost 续算)
              │
   ┌──────────┴──────────┐
 emission CDF         floor (= stillbox 计时器)
 连续推 SFallen→赛跑0.85   StillSec≥tFloor 保底发 SFallen
 (belief 抢先发, 所有 unit) (CDF 被假 InBed/area 压住、
                        到不了 0.85 时兜底)
   └──────────┬──────────┘
        同锚 tFloorFor(room×cell 保守) + StillSec==0 清
```
- `tFloorFor(area,room)=stillMuSigma(area,room) 的 μ+1.5σ`——floor 与 CDF 用同一 `(μ,σ)`，bathroom 未画 toilet 也用 bathsec(18min)，floor 不再 cell-only 12min 抢先于 CDF 的 18min；
- floor = 纯时长保底（**专治 CDF 被接触假阳/area redirect 压住、SFallen 到不了 0.85 的真摔**）。

**资源克制 — (b) belief 抢发照常、只砍兜底**（用户拍，§I/#1）：单房 unit 资源少、无邻房印证、lost 后 FP 风险剧增 → **floor 兜底腿单房不发**；但 emission CDF 抢发**不**受单房 gate（belief 所有 unit 照常）：
- **单房 unit 久躺真摔**：CDF 推 SFallen→0.85 抢发（强证据照报）；
- **单房 unit lost 不确定**：floor 兜底被砍（无邻房印证、FP 高时不保底）；
- **多房 unit**：CDF 抢发 + floor 兜底齐全。

**取消条件**（`engine/unit.go`，只作用 `band=="floor"` 兜底腿，不碰 belief 抢发的 lost/report）：
$$\text{撤 floor 兜底} \iff \text{单房 unit}(\text{len(rooms)==1}) \lor \rho_{xroom}>0(\text{hand-off 现身隔壁})$$
`StillSec==0`(移动/回来) 与 `exitL≥flip`(本人过门) 的撤销已在 `floor.go` 内（`StillSec<tFloor` 自然不发 / `exitL<flip` 条件挡），不在 unit 层重做。**房型差异（bathroom-D / 开放-UD）取消**——合体后统一为「单房克制 + hand-off 撤」，不再按房型分 timer。

**neighbor / hand-off 融入 — 一个量 `ρ_xroom` 两层，方向恒一致**：
- **belief 塑形层**（`GateBlindRow` F→L，**不变**）：`ρ>0` 把 blind track 转移 SFallen→SLeft，跟 emission CDF 推 SFallen **拉锯**——hand-off 成立则 SFallen 上不去（belief 抢发自然被压）；
- **floor 取消层**：`ρ>0` 撤 floor 兜底（人现身隔壁不兜本房）；
- **W 窗**（`HandoffWindowFor` 45/60/90s 按 unit_type）= hand-off **时序门**：lost 后 W 窗内邻房 EnterRoom 才算 hand-off（[[partial_monitoring_fall_suppression_law]]：只有极近两事件能排除 lost-fall，过窗当人没走、floor 照兜）；
- **两时标各管各**：W 窗(秒级)判「走没走邻房」、tFloor(分钟级)判「静止够久算摔」；ρ 同源→「belief 拉 SLeft」与「floor 撤」永不矛盾。

**与原草案的出入（实现修正）**：原 §I 设想「floor 被 D/DU 吸收、扩到单房（floor 退役后单房靠计时器）」；落地**反转**为 **D/DU 退役、floor(FloorGuard) 成计时器本体、单房不兜底**——因单房扩兜底会 FP 剧增（无邻房印证），违 D/DU 初衷；(b) 改成「belief 抢发照常、只砍兜底」既护住单房久躺真摔（CDF）、又克制单房 lost FP（不兜底）。

**实测（d5f7-0617 多房零回归）**：fire=40（floor 7 + lost 33），首 floor-fire 由 12.5min→18.3min（room×cell 生效），belief 抢发主导、floor 兜底退到 7 帧。

### §J 重复跌倒提前报 — per-logicID leaky 残余（2026-06-18，用户定 + **已落地**；contract 6A 其十六）

**来源**：333b-0618 firmware 实发 `pose=2(SuspectedFall)`/`pose=7(SuspectedSittingOnGround)`（PoseMap 注释说不发，实际发）。护理域："摔→起→再摔"是急性恶化红旗，要升级（第二次更敏感），不是每次站起清零。

**定位（用户 2026-06-18 重头拍）——只做"重复摔提前报"，第一次归 firmware**：
- **第一次/孤立摔 → firmware 负责**（pose=2→5@30~60s / pose=7→8@90s 确认报）。Xsensor **从不抢第一次**。
- Xsensor 唯一职责：**人刚摔过、又摔 → 抢在 firmware 前提前报**（补漏）。
- 这彻底避开「无动态阈值调制」争议——**我们根本不报单次/孤立摔**；调制源="刚摔过"=**同一跌倒风险轴**（跌倒史=临床第一预测因子，内禀），非 WeakBio 式正交外来信号；不改 firmware 阈、不否决 firmware（FN-safe 只增发不压制）。

**模型（per logicID 残余器，engine 层挂载，类比 FloorGuard）**：

fall episode = 连续处于 fall 族 pose $\{2,5,7,8\}$ 且 `PReal≥0.5` 真人；起身离族 → episode 结束。设本 episode 持续 $\text{FallSec}$、与上个 episode 间隔 $\Delta t$、前科残余 $R_{\text{prior}}=R\cdot e^{-\Delta t/\tau}$（进 episode 时一次性衰减）：

$$\text{credit}=\min\!\big(1,\ \tfrac{\text{FallSec}}{\text{Throu}}\big)\qquad \boxed{R_{\text{prior}}>0 \ \land\ R_{\text{prior}}+\text{credit}\ge 1 \ \Rightarrow\ \text{提前 fire}}$$

episode 结束烘进残余：$R \leftarrow R_{\text{prior}}+\text{credit}$，并 emit `self_recovered` 记录。

- **$R_{\text{prior}}>0$ 闸 = 第一次（$R=0$）永不触发**——天然交 firmware。只有"有前科"才可能提前。
- **隔久没摔** → $R$ 经 $e^{-\Delta t/\tau}$ 衰减回 0 → 又算第一次（firmware 负责）。
- **timing-only**：只提前 fire 时序，不改 fire 阈、不跳 severity 档（severity 随 R 升 = 待评，暂未做）。
- **不再走 belief/emission**：单次 pose-CDF 抬 SFall 的方案（曾试）已**撤回**（belief 层恢复干净）——单次归 firmware，belief 不需为单次摔加料。

**常数（`engine/repeat_fall.go`，留 oracle [[fall_data_is_artificial_test]]）**：
- `throuFallSec=60`（pose 2/5 族；firmware 保守端，用户拍）、`throuSitSec=90`（pose 7/8 族）。
- `repeatFallTauS=866`（残余漏衰减 τ；半衰期 ~10min = 急性聚集窗）。

**两条线分离（护理域）**：
- **派人线（dispatch）**：上式 $R\ge1$ → 提前 fire（band=`repeat`）。
- **记录线（incident log）**：每次 episode 结束（**含自救**）emit `self_recovered`（审计/FE/医护，铁律明列允许，**不受 $R$ 回落影响**——recovery cancel 派人，绝不抹记录）。当前落 xray forensic（`self_recovered`/`repeat_r` 字段）；推 `iot:event:stream`→event_log 待 publish 接线（随 cardagg wiring 一并，已 defer）。

**盲区衔接**：带 $R>0$（未了结）走出覆盖 → 进 [[partial_monitoring_fall_suppression_law]] 耐心窗，不被 stale recovery 抑制；卫浴超时另走 welfare-check（§I bathroom 段）。

**实现 + 验证（已落地 2026-06-18）**：
- 落点：`engine/repeat_fall.go`（`RepeatFallEscalator` per-logicID）+ `engine.go`（map 挂载/`Tick` present 帧调用/`dropTrack` 销毁/forensic `RepeatR`+`SelfRecovered`）+ `cmd/xsensor/main.go`（xray `repeat_r`/`self_recovered`）。belief/observation/adapter 已恢复干净（单次方案撤回）。
- **333b 4x 实测**（逐帧 R 轨迹）：第 1 段疑似摔(~15s) → $R:0\to0.244$（第一次 $R_{\text{prior}}=0$ **不报**，记 self_recovered）；间隔衰减 $0.244\to0.242$；第 2 段(~28s) $R_{\text{prior}}=0.242>0$ 进提前判、峰值 $0.703<1$ **不报**（没攒够），$R\to0.703$ 记 self_recovered。**fire=0**（333b 两段都不够久，正确不误报）。反推：若第 2 段 ≥45s 则 $0.242+45/60\ge1$ → ~45s 提前报（早于 firmware 60s）。
- **零回归**：escalator 只"加"火不"减"火，单次 fall $R_{\text{prior}}=0$ 永不触发 → cd2b 等单次 case 结构安全（escalator 不改既有 belief/floor/lost 火路）。

**真实跌倒时长文献核对（2026-06-22）——阈值 1.0 维持，不为"秒起"测试调低**：
- 真摔倒地时长 = **平均 14min（范围 2–59min）**；frail 老人 26 例真摔**无一能自起**、**50%+ 摔后起不来**（[PMC3850536](https://pmc.ncbi.nlm.nih.gov/articles/PMC3850536/) / [PMC2590903](https://pmc.ncbi.nlm.nih.gov/articles/PMC2590903/) / [PMC7905119](https://pmc.ncbi.nlm.nih.gov/articles/PMC7905119/)）。
- → **每次真摔 ≥2min ≫ throu 60s → credit 60s 内即封顶 1.0**：第 1 次交 firmware（≥60s firmware 本就报）；第 2 次 $R_{\text{prior}}≈0.5$–$0.9$ → **约 15–30s 提前报**。**阈 1.0 已能兜"刚摔又摔",无需调低。**
- **不调低的理由**：现实中"秒起"的是绊一下/快速坐下（stumble/sit）= **非跌倒** → 调低 1.0 会把它们误报（FP 暴涨，老人日均坐站数十次）。0621/333b/测试里"摔后秒起(11–14s)"→credit 仅 ~0.18 凑不够 = **人为测试不真实，非 bug**。文献"反复跌倒"=6 月内 >2 次（跨天/周），非几分钟内连摔。
- ⚠️ **验证 §J 须用"摔后躺 ≥30–60s 再起再摔"的真节奏 case**，不能用"秒起"测试（credit 凑不够 = 测试假象）。

---

### §K 直立证据「压/抬」分治 + 身份离场 evict — 2026-06-19 变更思路（A/C 互核；commit 7a1ec6a / def5940 / 748f4c3 / bdb9764）

本日四改一条主线：**把判断放对层 → 分到具体对象 → 腾出的质量要有去处 → FN 默认**。

**1. stillDiscount 从 floor 移入 emission（7a1ec6a）—— 单源不双压**
- 旧：z/pose 直立证据在 floor 折 still 时长（`stillDiscount`），emission 另有 ZBand 也压 → 同一 z 压两处 = 站立瘫倒过压漏报（**ZBand 原罪**）。
- 改：z/pose **单一归 emission**；floor 退回**纯 raw-still 计时器**（`StillSec=StillBoxSec`，不信 pose/z——pose/z 错报正是 floor 该兜的场景）。两层职责正交：emission = pose/z-aware 精判；floor = 不信标签的纯时间兜底。

**2. 「压 vs 抬」分治（bdb9764）—— 压 ⟺ 物理互斥，抬 ⟺ 治 Empty**
- **关键认知**：S 的「有人」态全**绑位置**（Bed/Sit/OpenFloor/Bath），无「在场·直立·未定位」格。只压 SFallen → 腾出质量归一漏进默认 **Empty** → present 真人判空房（`n_r=1` 但 `top=Empty` 自相矛盾）+ **Empty→Fallen 转移种子 = 0** 拖慢/漏真摔（FN）。**SOpenFloor 即「present·直立·未定位」兜底态**（复用，不必新增状态）。
- **判据（物理互斥，非「标签可靠性」）**：
  - **压 SFallen（<1，取 min）⟺ 与「倒地静止」物理互斥**：`z≥80`（身高硬测，质心高 ≠ 贴地）∨ `walk`（**运动 ⊥ 倒地静止**——摔者不动 → 不会被标 walking → 压不到真摔，故 walk 压**不要 z 门**纯标签可压）。
  - **抬（>1）⟺ 治 Empty / 安置**：在场直立（walk ∨ z≥80 ∨ `Standing`）→ 抬 SOpenFloor；`sit` → 抬 SSit。**standing/sit 是静态直立（与「静态倒地 / 刚摔未及标 Lying」共享静止 → 可混）→ 只抬不压**（z 未知不赌 fall，z<30 中性红线延续）。
  - walk∧z≥80：压 min 一次 / 抬 SOpenFloor 一次（**不叠**防过抬）。硬互斥的「压+抬」是 **feature**（互斥该压死，且压不到真摔）；软的「只抬」保边界报路 = FN-safe。
- 公式（log 域，权重 $w=\text{covers}$）：`logS[SFallen]+=w·ln(min(supWalk,supStand))`（压）；`logS[SOpenFloor]+=w·ln(redOpen)` / `logS[SSit]+=w·ln(redSit)`（抬）。进 `filter.Correct` 平滑单帧噪声，非乘递归 P。
- ⚠️ **walk 边界**：地板挣扎摔者（低 z + 乱动）可能误标 walk 被压 → FN；靠 floor 时长兜底（不看 pose）+ 挣扎停转 Lying 接住（有界瞬态，留 oracle 真挣扎数据验）。

**3. ExitRoom 后 logicID churn 根治（def5940）—— belief 离场判定要回传 track_manager**
- 根因：belief 状态驱动 drop（`!Present ∧ SLeft+SEmpty≥0.9`）只删 belief/census 的 logicID，但 **track_manager 12s coast（`trackEvictMaxMs`）仍每帧把已离场 track 当 base 重发** → census 无对应 logicID → 每帧重发新号 = churn（cabb `lid 3→145`）+ rebirth FN 隐患（[[w3_3_realness_wired_rebirth_fn]] 另一触发路径）。
- 改：镜像「fire→`ResetStillBox`」通道，加「belief drop→`tm.EvictTrack`」（立即删 tracks/outputs，停 coast re-feed）。FN-safe：摔倒高 SFallen 永不进 drop → faller 绝不被 evict；闪失重捕的 coast 不动（无离场证据 belief 不 drop）。白赚消掉 churn 重生轨攒出的雷达原点幽灵 floor FP。
- 区分：12s coast 对「瞬时丢轨重捕」是对的（不能动，否则真人闪一下碎轨）；对「确认离场」是错的（人真走不会重捕）。evict 只在 belief 确认离场时触发，绕过 coast 但不动它。

**4. floor 接触豁免收窄（748f4c3）—— 分到具体床，分不清照报**
- floor `contactInBed` 从「任一近床 InBed 豁免」收窄到「**唯一一张近床 ∧ InBed**」；近多张床 = 床归属分不清（`NearBedMask` 非互斥，摔者同时在两床 100cm）→ 不豁免照报（FN-safe）。单床 sleeper 豁免不变（无回归）。
- **认知**：「InBed→压摔」的「分具体床 + 分不清用概率（1/n）+ 用 MM」原理，**belief Ψ 早已实现**——`Ψ = Σ_j a_j ψ̃_j + a_∅` 是 **mixture 非 product**，`a_j = κ_j·g^xy_j` 归一 = MM 归属概率（1 床→a≈1=100%、2 床分不清→对半=1/n），注释明否决 product（「任一床 occ 压死 F = 漏报」）。floor 是二值离群，本次对齐为 FN-safe。

**贯穿主线（一句话）**：**分到具体对象**（具体 track / 具体床 / 具体态）→ **分不清用概率**（Ψ mixture / 压取 min）→ **腾出质量要有去处**（抬 SOpenFloor 不漏 Empty）→ **FN 默认**（软证据只抬不压、摔倒永不 evict、床分不清照报）。

**验证 + 待验**：cabb-0616（无床退出负样本）全绿——churn `lid 145+→2`、幽灵 FP 消、幸存站立 `top Empty→OpenFloor`（SOpen 0.61 / SEmpty 0.04）、SFall `0.20→0.01`、`fire 0`。🔴 **带床真摔（cd2b / 9e7）未 replay**：四关待过——① z≥80 误压真摔 FN（emission 压，从未过 FN 关）② 抬-redirect / 软站 / sit→SSit 带床场景 ③ floor 多床二义 ④ EvictTrack 带床+离场。全部 `sup*/red*` 是 form-anchor 留 oracle（[[fall_data_is_artificial_test]]）。

---

### §L risktime 纯时间轴 — 只缩短 floor tFloor，退出 C_FN（2026-06-22，用户定 + **已落地**）

**原则（用户拍）**：risktime（夜间）**只动"决策层的耐心/等多久兜底"，不动证据（$P^F$）、不进报警阈（C_FN）**。同一 risktime 不能既降阈又缩时间（双重计风险 → 过敏 FP）。语义自洽：夜间真正变的是"没人巡视、躺久没人发现"=**时间问题**，不是"这一下算不算摔"=证据问题。

**职责切分（按驱动因子分）**：

| 轴 | 管什么 | 驱动因子 | 动什么 |
|---|---|---|---|
| **C_FN → pFire** | 报不报（证据够+值不值得报） | 可救援性：people / alone / disabled + $P^F$ | 阈值高低 |
| **tFloor**（§I floor 兜底） | 等多久才兜底（静止多久判摔） | **risktime（夜间）** | 时间长短 |

**实现**：
- `tFloorFor(area, room, isRiskTime) = μ + k·σ`：白天 $k=1.5$ / 夜间 $k=0.5$（$μ,σ$ 不变=物理，只动风险容忍 $k$；**只缩短不延长** = FN-safe）。default 区：白天 720s / 夜间 560s。
- `Observation.IsRiskTime` ← `IsNightTime(nowMs, 房时区)`（main.go 算 + bootstrap 填 `roomTZ`；xray `risktime` 字段）。
- **退出 C_FN**：删 `RiskContext.Night` + `decideParams.nightMult`（原 `Census.Night` 从没填 = risktime 在 DBN 本来就死；现真算并喂 floor）。
- floor 即时兜底（`fg.Step` 够阈直接 `d.Fire=true`，不走 tHold）→ 夜间缩短立即生效。
- **天然不误吵睡觉**：floor 在 bed/deny 区不发，只对 floor/open/unknown 区生效 → 夜间缩短只作用"躺地/床边"。

**配置（care 政策，用户定）**：`risk_time` 夜间窗 = **21:30–7:00**（覆盖默认 23:30；老人 21–22 点入睡 + 起夜跌倒高危，23:30 太晚漏前半夜）。

**验证（09e7-0621，真 case）**：床边摔躺 11min，夜间口径 tFloor 720→**560s** → **22:26:44 floor 兜底报出**（人 22:29 起身，提前 2.5min）；白天（旧 23:30 配）零回归（risktime=false，与改前一致）；grep 确认 fire 判据无 risktime（只在 floor + forensic）。

**落点**：`belief/floor.go`（tFloorFor+k）、`belief/observation.go`（IsRiskTime）、`belief/decide.go`（删 Night/nightMult）、`adapter/adapter.go`（BuildObservation 传 night / 删 RiskContext.Night）、`engine/engine.go`（3 处传 night）、`cmd/xsensor/main.go`（IsNightTime+roomTZ+log）、`cmd/xsensor/bootstrap.go`（填 roomTZ）、`wisefido-sensor/config.yaml`（risk_time）。commits c63e46b / 0d6ddda / d32c48d。

---

### §M 老人步速规范 + 速度类参数标定核对（2026-06-22 文献）

**老人惯常（室内/community-dwelling）步速**：均值 **~1.2 m/s**（Rotterdam）；室内 ~1.11 m/s（均龄 69）；跌倒风险临界 **< 0.85–0.88 m/s**；衰弱/高危（>80 岁）**< 0.6–0.8 m/s**；71 岁起下降最明显。
→ **室内走 ≈ 80–120 cm/s，衰弱 < 85，极衰弱 < 60**（[Rotterdam](https://www.sciencedirect.com/science/article/pii/S0531556521004289) / [室内步速 PMC6840929](https://www.ncbi.nlm.nih.gov/pmc/articles/PMC6840929/) / [gait-speed norms](https://geriatrictoolkit.missouri.edu/gaitspeed/Gait-Speed-norms.doc)）。

**对照系统速度类参数**：

| 参数 | 现值 | 用途 | vs 老人步速 |
|---|---|---|---|
| `move_speed_cms` | 20 cm/s | Kalman 兜底判 Move（pose 不准时） | 远低于惯常 80–120 → 保守(救慢走老人),OK |
| `AssocCm` | 250 cm | 帧间最大关联位移 | 1Hz 下够宽（120cm/s 走 120cm/帧），OK |
| `fuseMoveThreshCm` | 40 cm | 融合"有意义移动"窗 | 80–120 cm/s 轻松超 → OK |
| 速度合理性 | >150 cm/s 减分 | track score | =1.5 m/s，正常老人不会触发，OK |
| **`birthMaxRealisticCm`** | **150 cm** | **ghost 出生地：出生位距门 ≤此 ∧ 近期 EnterRoom→真人；>此→偏 ghost** | ✅ 足够 |

**`birthMaxRealisticCm=150` 标定（核对：足够）**：监控/track 上报 **1Hz** → 门→首帧检测滞后 ≤1s → 正常老人 1 秒最多走 **~120cm**（1.2 m/s）。放宽到 **150cm = 1.25× 余量**（对 1.2 m/s）；对跌倒高危步速 **0.88 m/s → 150/88 ≈ 1.7×** 余量。**足够扫到一次正常人 + 留裕量，不偏紧。**
- （早前一版误把它当 2.5s track 周期算成 0.6 m/s 判"偏紧"——错；实际 1Hz，150 充分。）
- 用处：`track_manager.go:1585`（出生配对加分）/ `:2013`（出生宽限跳过）。全仓无其它 0.6 m/s / 60 cm/s 速度假设。
- 影响面仅出生 Real/ghost 评分 → N_r 人数（不门控 fire，realness 绝不否决摔，[[realness_never_vetoes_fall]]）= FN-safe。
