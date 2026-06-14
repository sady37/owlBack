# feedback-p6C 附件 — 评审组 C 的阶段 1 骨架验收规格

> **角色边界（C 自我更正）**：production code（含阶段 1 骨架）执笔归项目组 A；C 定验收标准 + 独立审。
> C 此前一稿曾代写 joint/bed_axis/filter——**作废为执笔产物**，仅可作 A 的实现参考；正式骨架由 A 执笔、B/C 审。
> 本文件 = C 交给 A 的**验收规格**（断言级，A 实现后 B/C 据此审真 code）。

---

## 验收前提（不验则骨架不可信）

阶段 1 骨架 = 纯联合滤波，Ψ/Φ 中性占位（=1），Correct 为恒等。验**数值正确性**非行为。与 δ/neighbor 零依赖。

---

## T1 · 归一化守恒 Σα=1

**断言**：任意 `numBeds ∈ {0,1,2,3}`，初始 α 及连续 ≥50 步 Predict 后，`Σ_i α[i]` 与 1 的偏差 < 1e-9。

**构造**：`NewFilter(DefaultModel(), nb)`；online 全 true；循环 Predict(1) ×50，每步后查 ΣΑ。

**通过判据**：`|ΣΑ − 1| < 1e-9` 恒成立（含初始态）。失败 = 因子化 Predict 或 normalize 有质量泄漏。

---

## T2 · 单床退化基数 + numBeds 显式持有（B1）

**断言**：联合空间基数 = `numStates · 2^numBeds`：

| numBeds | 期望 size |
|---|---|
| 0 | 9（回 S 轴单链） |
| 1 | 18 |
| 2 | 36 |
| 3 | 72 |

**附加断言（B1 闭合）**：`Filter.NumBeds()`（或等价显式字段）必须返回构造时的 numBeds——**床数显式持有，不靠隐式推断**。

**退化正确性**：`numBeds=0` 时 `MarginalS(α)` 必须逐分量等于原单实体 `Belief` 行为（初始 == model.Prior，偏差 < 1e-9）。

---

## T3 · P-5 maxBeds 超界硬断言

**断言**：`maxBeds=3`（养老院房 ≤3–4 床上界）。

- `numBeds ∈ {-1, 4, 99}` → 构造（NewJointSpace/NewFilter）必须 **panic**（不静默膨胀）。
- `numBeds == maxBeds`（=3）→ **不 panic**（边界合法）。

**通过判据**：超界 recover 到非 nil；边界值无 panic。

---

## T4 · D-2 核心：ε≪λ 复现 30s staleness（cd2b 治本不变量）

**这是阶段 1 最关键验收——"一个不等式 ε≪λ 替代所有 staleness/TTL"的实证，也是 cd2b 漏报根因（陈旧 occ 永不衰减）在新架构治本的直接证据。**

**喂数序列**：
1. 单床（numBeds=1）。构造初始 occ 主导：质量集中到 `(SBed, bmask=1)`（人在床睡，床 occ），归一化。
2. 验初始 `P(B^0=occ) = MarginalB(α,0) ≥ 0.99`。
3. **sleepad 离线**：`online = {false}`。逐步 Predict（1Hz，1 步 = 1s），共 120 步，每步记 `P(occ)`。

**期望衰减曲线**（§C 单向泄漏核 K^unobs，λ 半衰期 ≈ ln2/λ）：
- `P(occ)` 单调下降（occ→vac=λ 泄漏，vac→occ=0 吸收 → 单向）。
- **跌破 0.5 的步数 `crossStep` 必须 ∈ (0, 30]**——陈旧 occ 在原 30s staleness 窗内蒸发到 vac 主导。
- λ 形态约束：默认 λ 使半衰期 < 离线判定窗（~30s）；occ 主导跌破 0.5 约 1 个半衰期。**值是标定（feedback-p6C §5），但"< 30 步跌破"是形态验收，必过。**

**在线对照组（证明蒸发归因离线核，非数值漂移）**：
- 同初始 occ 主导，`online = {true}`，Predict ×30。
- 期望 `P(occ)` 仍 ≥ 0.5（K^obs 高自持，ε 小，不蒸发）。
- 失败 = ε 太大（在线核也蒸发）→ ε≪λ 不成立。

**通过判据**：离线 `crossStep ∈ (0,30]` **且** 在线 30 步后 `P(occ) ≥ 0.5`。两者缺一即 D-2 不通过。

---

## T5 · §C vac 吸收态（单向泄漏正确性）

**断言**：空房（`(SEmpty, bmask=0)`，床 vac）+ 离线（online=false）+ Predict ×120 后：
- `P(B^0=occ) ≤ 0.05`（vac→occ=0 吸收 → 空房离线**不被弛豫成 0.5 伪占用**）。

**为什么单列**：这是 A 坚持"弃矩阵、用箭头记法"要守的核心——若 K^unobs 方向写反（occ 吸收/vac 泄漏），此断言失败 = cd2b 漏报原样重演。**T5 是 §C 箭头记法方向正确性的守门测试。**

---

## 验收总判据

T1–T5 全过 = 阶段 1 骨架数值正确，可进阶段 2（emission/coupling/transition）。
任一不过 = 骨架有数值缺陷，阶段 2 不得动。

**审查归属**：A 执笔骨架 → B/C 据本规格独立审真 code（C 不审自写代码，独立性保住）。
