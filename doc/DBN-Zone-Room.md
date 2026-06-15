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
- **人数 $N_r$ / 真伪 ghost**：同构造换轴——$S^{(i)}$ 多份，$N_r=\sum_i\mathbb 1[S^{(i)}\notin\{E,L\}]$，realness $T^{(i)}$ 由跨 track $\rho$（与 $\kappa$ 同源的共现/镜面）耦合，$P(\text{ghost})$ 从 co-existence 涌现。与本床轴正交，同一滤波形式。

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

记号：本房 $r$、同 unit 兄弟房集 $\mathcal N(r)$；本房 real-present track 消失时刻 $t^{\text{lost}}_r$（进 Blind 的丢轨时刻）；兄弟房 hand-off 落点事件时刻 $t^{\text{arr}}_{r'}$（EnterRoom ∨ InBed 翻转）；滞后

$$\Delta_{r'}=t^{\text{arr}}_{r'}-t^{\text{lost}}_r\qquad(\Delta>0=\text{先走后到}=\text{有向命中})$$

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
| ① | ghost 轴 → neighbor 轴 | $q_{r'}$ 吃兄弟房 $P_{r'}(\text{real-present})\cdot(1{-}P(\text{ghost}))$——§10 房内 ghost 后验喂 §A 房间 hand-off；兄弟房 ghost 不算落点 |
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

### 整合后轴/裁决一览

| 轴 | 证据 | 状态 |
|---|---|---|
| 空间（§5 $g^{xy}$）| 雷达 XY | 已内化；δ 标定、脆弱（[[feedback-p6C]] 附）|
| 时间（dwell 符号）| 久静 × cell 容忍 | 已内化（survival）；cell 容忍可靠性 = 框架前提 |
| 裁决（§8 $C_{FN}$）| 风险因子 | 主框架；$C_{FN}$ 连续代价函数须 decide 落地 |
| 跨房（§A neighbor ρ_xroom）| 兄弟房**有向** hand-off | 框架命题 + **方程已落（§A.1 ρ_xroom / §A.2 $T_S$ 门控 / §A.3 §10 接口）**；曲线参数待标定 |
