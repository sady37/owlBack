# 评审 B：DBN-Zone-Room 联合占用模型 — 阶段 1 设计审查

审查对象：`doc/DBN-Zone-Room.md`（联合占用模型设计底稿）
对照基准：设计文档自身的方程自洽性 + 物理约束 + cd2b 已知事实
审查阶段：**阶段 1（设计审查）**——不涉及代码、不涉及接口、不涉及实现选择
审查日期：2026-06-14

## 总评

设计骨架正确：$(S,\{B^j\})$ 联合隐空间、$\Phi$ 按 attachment 分轴、$\Psi$ 相容势耦合、$T$ 纯因子化、log 域数值方案——这些方向都是对的。§9 三态验证表完整覆盖了 cd2b 的三条路径。四条发现需要 A 在设计层裁定后才能进实现，一条需要标定阶段验证。

---

## 阻塞项（设计层未闭合，不能进实现）

### B1. $K^{\text{obs}}$ 的 vac→occ 参数 $\mu$ 定义后从未约束

**位置**：§6，方程 $K^{\text{obs}}:\ \text{occ}\to\text{occ}=1-\varepsilon,\ \text{vac}\to\text{occ}=\mu$

$\varepsilon$ 在上下文中有明确语义——在线时床态翻转是稀有事件（上下床），$\varepsilon\ll1$。但 $\mu$ 没有任何约束：它可以等于 $\varepsilon$（进入床和离开床同样稀有），可以等于 0（床垫在线时只能检测离开不能检测进入），也可以等于某个中间值。更关键的是 $\mu$ 与 $\varepsilon$ 的关系——如果 $\mu\gg\varepsilon$（进床比离床容易），则 $B$ 链有向 occupied 的漂移；如果 $\mu=0$，则床垫在线时 vacant 是吸收态。

$\mu$ 的值直接决定：床垫在线时，一个人上床后 $P(B=\text{occ})$ 的收敛速度。没有这个参数的行为语义，$T_B$ 在在线模式下的表现是未定义的。

**建议**：明确 $\mu$ 的物理语义和数量级，或说明为什么对称（$\mu=\varepsilon$）是合理默认。

### B2. $K^{\text{unobs}}_\lambda$「向无信息弛豫」的目标分布未指定

**位置**：§6，$K^{\text{unobs}}_\lambda:\ \text{向无信息弛豫, 速率}\ \lambda$

「无信息」有两种不等价的定义：
- **定义 A（均匀）**：$P(B)\to[0.5, 0.5]$。离线足够久后 vacant 和 occupied 等概率。空房离线 → P(occupied) 从 0 升到 0.5。
- **定义 B（泄漏）**：仅 occupied 泄漏到 vacant，vacant 不动。vacant 是吸收态。空房离线 → P(occupied) 保持 0。但床垫离线时人上床 → 模型永远学不到 occupied。

两种定义在 $\varepsilon\ll\lambda$（陈旧占用快速蒸发）这个核心机制上完全一致——差异只在 vacant→occupied 方向。cd2b 只依赖 occ→vac 方向（陈旧 InBed 消失），所以两种定义都能解 cd2b。差异出现在：空房 + 床垫长期离线 + 无 entry 事件时，P(occupied) 的行为。

§6 的方程写的是矩阵形式，暗示 $2\times2$ 满矩阵——如果意图是定义 B，方程应写为 $K^{\text{unobs}}_\lambda = \begin{pmatrix}1 & \lambda \\ 0 & 1-\lambda\end{pmatrix}$。

**建议**：显式写出 $K^{\text{unobs}}_\lambda$ 的矩阵形式，或声明定义 A/定义 B 的选择及理由。

### B3. HR/RR 非对称似然的正面/反面不对称隐含 FP 风险

**位置**：§5，$\ell_{\text{hrrr}}(\text{absent}\mid\text{AtBed}) = 1/L_{\text{hr}} < 1$

设计意图是：absent 主动否决 AtBed，迫使信念离开 AtBed 去找替代解释（$F$/Sit/$O$）。当人真的不在床时（cd2b 离线场景），这是正确的。但当人在床、HR/RR 因非床原因缺失时（mmWave 探测距离/姿态/遮挡），absent 同样否决 AtBed——把人从 AtBed 推到 $F$。HR/RR 的假阴性直接转化为 fall 假阳性。

这是非对称似然的固有取舍——present 的信息量越大，absent 的误伤面也越大。文档 §8 的风险不对称裁决（单人独处时 $C_{\text{FN}}\uparrow$）可以兜底——宁可误发也不漏真摔。但这条取舍在设计文档里没有显式标注。

**建议**：在 §5 的 HR/RR 含义段显式标注「HR/RR absent ≠ 人不在床」的假阴性风险，以及它与 §8 风险不对称裁决的关系（宁可误发）。

### B4. $\Psi$ 多床混合用加权和而非乘积的语义选择

**位置**：§4，$\Psi_t(S,\{B^j\}) = \sum_j a_j \tilde\psi_j(S, B^j) + a_\varnothing$

每个 $\tilde\psi_j$ 只依赖 $S$ 和 $B^j$——不依赖其他床。当 $|\mathcal B|>1$ 时，联合配置 $(S, B^1, B^2)$ 的相容性是各床相容性的加权**和**（再除以归一化）。这是 mixture 语义：人以概率 $a_j$ 归属床 $j$，在该床上按 $\tilde\psi_j$ 评价相容性。

乘积形式 $\Psi = \prod_j \tilde\psi_j$ 会是另一种选择：人对所有床同时相容。当前 mixture 形式的物理直觉是「人只能在一张床上」，乘积形式的物理直觉是「人对每张床独立相容」。文档 §4 的「软归属 $a_j$ 代床归属隐维」解释了为什么用 mixture 而不用乘积——床归属不作为独立隐维进状态空间，而是作为软权重边缘化——但没解释为什么 mixture 比乘积更适合这个边缘化。

这不是错误，但实现时如果误用乘积，多床场景的 $\Psi$ 行为会完全不同。

**建议**：在 §4 显式标注 mixture-vs-product 的选择及一行理由。

---

## 澄清项（设计文档内有矛盾或未定义，不阻塞实现但需标注）

### C1. $o_j$ 记号在 §2 和 §4 间漂移

§2 定义 $o_b(r,s)$ 为雷达与睡眠器的 overlap 标量。§4 的 $\psi_{\text{phys}}$ 表里出现 $o_j$（下标从 $(r,s)$ 变成 $j$）。如果 $o_j$ 是 $o_b(r,s_j)$ 的简写（床 $j$ 对应的 overlap），应声明。如果是不同的量，需要定义。

### C2. $\text{covers}(r,\cdot)$ 的点参数未指定

§5 的 $w_{\text{pose}} = \text{covers}(r,\cdot)$——点代表什么？对全部床取 max？对当前最近床？房间整体覆盖率？在单床场景下无歧义，但在多床场景下，pose 似然的权重取决于这个选择。

### C3. $\varepsilon_{\text{art}}$ 的值域未约束

§4 的 $\psi_{\text{phys}}(F, \text{occ}) = \varepsilon_{\text{art}}$ 描述为「极小」。$\varepsilon_{\text{art}}$ 太小会导致 §7 的 log 域 $\log \varepsilon_{\text{art}}$ 是一个很大的负数，在 Correct 阶段与 $\Phi$ 的正向似然相加时可能被淹没。$\varepsilon_{\text{art}}$ 太大则 $F$+occ 组合压不死——床上翻身的 pose=Fallen 误读会穿透。

文档没有给 $\varepsilon_{\text{art}}$ 的数量级（$10^{-2}$? $10^{-3}$? $10^{-6}$?）。这在设计层不是阻塞项——具体值在标定阶段定——但标定前必须有一个初始值，而初始值的数量级需要在设计层确定（与 $L_{\text{in}}$ 的数量级和 log 域的数值精度联合考虑）。

---

## 标定依赖项（设计正确，但正确性以未测物理量为前提）

### S1. $\delta_{\text{pad/floor}}$ 的跨 case 稳定性

文档 §8 把 $\delta_{\text{pad/floor}}$ 标为「唯一悬空输入」，这是诚实的。已有一个 case（cd2b-0604）测出 $\delta\approx1$ nat（边际可分离）。这证明「这个 case 的摔点位置可分」。但文档 §9 场景 3 声称模型能解「sleepad 离线摔」，前提是 $g^{xy}$ 几何似然能提供足够的 $F$-vs-AtBed 判别力——而这个判别力的物理基础正是 $\delta$。如果 $\delta$ 在其他床型/其他摔落点降到 $\approx0$，场景 3 的 claim 需要降级为「依赖风险兜底」。

**建议**：在 §9 场景 3 的机制描述中加注：该路径的有效性取决于 $\delta\gg0$；$\delta\approx0$ 时场景 3 退化为 §8 的不可判+风险兜底路径。

---

## 对照 §9 三态验证表的完整性检查

| 场景 | 依赖的设计要素 | 要素在文档中是否定义 |
|---|---|---|
| 正常睡 | $\ell_s(\text{InBed}\mid\text{occ})\gg1$, $\psi_{\text{phys}}(F,\text{occ})=\varepsilon_{\text{art}}$ | ✅ 定义完整 |
| 在线摔 | $\ell_s(\text{LeftBed}\mid\text{vac})\gg1$, $\psi_{\text{phys}}(F,\text{vac})=1$ | ✅ 定义完整 |
| 离线摔 | $\ell_s\equiv1$（离线中性）, HR/RR absent 压 AtBed, $\Psi$ 拉 $B\to$vac, $\lambda$ 漏 occ | ✅ 机制完整，但性能依赖 $\delta_{\text{pad/floor}}$（S1）和 HR/RR 假阴性率（B3） |

---

## 对 A 的汇总

| # | 类型 | 描述 | A 需要做的 |
|---|---|---|---|
| B1 | 阻塞 | $\mu$ 参数未约束 | 给定 $\mu$ 的值/范围/语义 |
| B2 | 阻塞 | $K^{\text{unobs}}$ 目标分布未指定 | 选 A（均匀）或 B（泄漏），写进 §6 |
| B3 | 阻塞 | HR/RR 非对称似然的 FP 风险未标注 | 在 §5 加一行风险标注 |
| B4 | 阻塞 | $\Psi$ mixture-vs-product 未声明 | 在 §4 加一行选择理由 |
| C1-C3 | 澄清 | 记号漂移、点参数、$\varepsilon_{\text{art}}$ 量级 | 改文档或标 TODO |
| S1 | 标定 | $\delta$ 跨 case 稳定性 | 多 case 验证或标注退化条件 |
