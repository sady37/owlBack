# 评审 B：第 3 轮审查 — 阶段 3 decide（§8 期望损失裁决）

审查对象：commit `444dab7`（decide.go + decide_test.go）
对照基准：`doc/DBN-Zone-Room.md` §8/§B
审查阶段：阶段 2（实现审查）— 对照设计文档方程，查代码忠实度

---

## 总判：通过。§8 期望损失主框架忠实实现，A 三立场全部兑现，DEC1–5 全过，T1–E5 零回归。

---

## 一、测试实测

| 组 | 测试 | 结果 |
|---|---|---|
| DEC1 | fire 条件 + T_hold 持续 + 断开复位 | ✅ |
| DEC2 | 风险分层（同 P^F=0.3：独处+夜 fire / 多人白天不 fire）| ✅ |
| DEC3 | 代价翻转非 argmax（P^F=0.4 < AtBed 0.45，独处 fire）| ✅ |
| DEC4 | 不可判无独立分支（Λ=1+高风险独处 P^F=0.1 仍 fire；Identifiable≠gate）| ✅ |
| DEC5 | ComputeLambda 诊断（全暗→1 / informative≫1）| ✅ |
| T1–E5 | 骨架/emission/coupling 回归 | ✅ 零退化 |

---

## 二、方程→代码逐条对照

### §8 fire 条件

设计：$$\text{fire} \iff P^F_t C_{FN}(\text{risk}) > (1-P^F_t) C_{FP} \text{ 持续} \ge T_{hold}$$

代码 `Decider.Step`：
- `margin = pFallen*cfn - (1-pFallen)*cFP` ✅
- `inst = margin > 0` ✅
- `fireSinceMs` 追踪持续，`>= tHoldMs` 后 `Fire=true` ✅
- `margin <= 0` 时 `fireSinceMs=0`（断开复位）✅

### §8 Λ_t 似然比

设计：$$\Lambda_t = \frac{\sum_{\{B^j\}} \Psi_t\Phi_t(F,\{B^j\})}{\sum_{\{B^j\}} \Psi_t\Phi_t(\text{AtBed},\{B^j\})}$$

代码 `ComputeLambda`：
- `fLogs[b] = logPsi[F,b] + logPhi[F,b]`，`LogSumExp` ✅
- `bLogs[b] = logPsi[AtBed,b] + logPhi[AtBed,b]`，`LogSumExp` ✅
- `exp(LogSumExp(fLogs) - LogSumExp(bLogs))` ✅

### §B 期望损失主框架

A 三立场全部落地：

1. **不可判不写独立分支** ✅ — `Identifiable` 仅诊断标注，`Step` 内 `lambda` 不参与 fire 决策。DEC4 验证：Λ=1 + 高风险独处 P^F=0.1 → 仍 fire。
2. **C_FN 只设保守 form-anchor** ✅ — `cFN()` 连续、各因子单调、多人折扣 floor=0.3 不归零。参数显式标注"标定锚，非权威值，留 oracle"。
3. **cd2b 主解不改变** ✅ — decide 是 δ≈0 不可判时的退路；cd2b（δ≫0）E5 已在 emission 解掉 0.9998。

### C_FN 连续代价函数

代码 `cFN(rc)`：
- 独居连续增益（线性 0→aloneSatMin=30min 饱和）✅ 非离散档
- 夜间倍数 ✅
- 失能倍数 ✅
- 多人 1/N 折扣 + floor 不归零 ✅ 非硬 OFF

---

## 三、与 C §8 三前提的对账

| C 前提 | decide.go 状态 |
|---|---|
| C_FN 形态是框架（连续非离散档）| ✅ `cFN()` 连续函数，消费同源风险因子 |
| C_FN 取值是标定（留 oracle）| ✅ 参数显式标注非权威 |
| cell 容忍可靠性（dwell 符号框架级）| 在 emission/survival 层，decide 不涉及 |

---

## 四、A 三立场验证

| 立场 | 代码证据 | 验证 |
|---|---|---|
| ① 不可判无独立分支 | `Step` 不读 `lambda` 做 gate；DEC4 实证 | ✅ |
| ② C_FN 只设保守形态 | `defaultDecideParams()` 注释"标定锚非权威值" | ✅ |
| ③ cd2b 主解不改变 | E5=0.9998 仍在 emission 层；decide 不加干预 | ✅ |

---

## 五、放行判据

- DEC1–5 全过 ✅
- T1–E5 回归零退化 ✅
- §8 方程忠实实现 ✅
- §B 主框架正确（不可判无独立分支）✅
- C_FN 连续代价函数落地（形态=框架）✅

**B 净判：阶段 3 decide 通过。裁决链从 Predict→Correct→Decide 已闭合。放行。**
