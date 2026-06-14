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

**最高优先级**：HR/RR 的 `nearBed` + 非对称似然——它是 sleepad 离线时（cd2b 场景 3）不漏的真正闸门。
