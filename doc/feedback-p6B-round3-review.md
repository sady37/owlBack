# 评审 B：第 3 轮审查 — 阶段 3 decide（§8 期望损失裁决）

审查对象：commit `8a3dc09`（decide.go + decide_test.go + multibed_test.go + probe.go）
对照基准：`doc/DBN-Zone-Room.md` §8/§B + C §17 D1–D6 验收规格
审查方式：独立跑全 24 测试 + 逐行读实现 + 独立代数验算 + 控制流追踪
状态：阶段 2 实现审查，对照设计文档方程 + C 验收规格

---

## 总判：通过。D1–D6 全部忠实，控制流/边界条件干净，全 24 测试通过（含新增 4 个多床/probe 测试）。一项合同缺口建议，一项值得肯定的超规格。

---

## 一、独立实测

| 组 | 测试数 | 结果 | 注 |
|---|---|---|---|
| T1–T5 骨架 | 5 | ✅ | 零回归 |
| E1–E5 emission | 5 | ✅ | cd2b 0.9998 无退化 |
| C1–C5 coupling | 5 | ✅ | mixture 55× 无退化 |
| DEC1–5 decide | 5 | ✅ | 本次新增 |
| MB1–2 多床 | 2 | ✅ | 本次新增（§E mixture 三床/路由双床）|
| CoversMaxC2 | 1 | ✅ | 本次新增 |
| ProbeSnapshot | 1 | ✅ | 本次新增 |
| **合计** | **24** | **全过** | `-count=1` 无缓存 |

---

## 二、D1–D6 逐条核（读实现 + 独立代数验算 + 控制流追踪）

### D1 期望损失裁决（非 argmax）

```
decide.go:94  margin = pFallen*cfn - (1-pFallen)*cFP
decide.go:95  inst = margin > 0
```

- 方程 = $P^F \cdot C_{FN} - (1-P^F) \cdot C_{FP}$，与 §8 一致 ✅
- 严格 `> 0`（非 `≥ 0`）：精确中性不 fire，保守 ✅
- 例：cFN=2 时 min $P^F$ = 1/(1+2) = 0.333…，$P^F$=0.333 时 margin<0 不 fire。argmax 在此 $P^F$ 下无论风险因子均不 fire，而代价翻转在 cFN=15（独处+夜）下 min $P^F$=0.0625——代价不对称造成 ~5× 灵敏度差 ✅

### D2 Λ 绝不作 gate

```
控制流追踪：
  Step(nowMs, pFallen, lambda, rc)
    → cfn := d.p.cFN(rc)           // lambda 不参与
    → margin := ...                 // lambda 不参与
    → inst := margin > 0            // lambda 不参与
    → fireSinceMs 更新              // lambda 不参与
    → fired := ... >= tHoldMs       // lambda 不参与
    → Decision{Lambda: lambda, Identifiable: lambda > 3.0}  // 仅写 forensic 字段
```

`lambda` 在 `Step` 的 5 个决策步骤中**零次读取**；仅在返回的 `Decision` 结构体赋值两次（`Lambda`/`Identifiable`），且 `Identifiable` 注释："诊断标志，**不参与 fire 决策**"。DEC4 测试实证：Λ=1（全暗不可判）+ 高风险独处 $P^F$=0.1 → margin=1.35>0 → inst=true。Identifiable=false 不阻断。✅

### D3 C_FN 连续/单调/多人不归零

独立逐因子验算（代数追踪，非取测试值）：

| 因子 | 计算 | 单调性 | 连续/离散 |
|---|---|---|---|
| 独居 | `1 + aloneGain·min(aloneMins/aloneSatMin, 1)`，aloneGain=4，aloneSatMin=30 | ✅ 单调不减 | ✅ 连续线性饱和 |
| 夜间 | `×1.5`（若 Night）| ✅ | 布尔门控，乘性 |
| 失能 | `×1.5`（若 Disabled）| ✅ | 布尔门控，乘性 |
| 多人 | `×max(1/N, 0.3)`（若 PeopleCount>1）| ✅ N 增则降 | ✅ 连续 1/N + 下限 |

多人下限=0.3 不归零——`PeopleCount>1` 时才触发，单人不折扣 ✅
极端值验算：cFNBase=2, aloneSat=30min, night, disabled, alone=1 → cFN=2·(1+4·1)·1.5·1.5 = 22.5 ✅
单人无折扣：PeopleCount=1 → cFN 不乘折扣 ✅

### D4 C_FP 归一

`cFP = 1.0`，注释"基线（护士跑空腿）" ✅

### D5 cd2b 不回归

E5 测试独立重跑：P(Fallen)=0.9998 不变。decide 只在 $P^F$ 上做代价乘法，不修改上游信念。✅

### D6 取值非权威标注

`decideParams` 行 21："标定锚，非权威值，留 oracle"
`cFN()` 行 42："形态=框架…取值=标定"
`defaultDecideParams()` 行 33：每个字段注释带"基线/非权威/留 oracle" ✅

---

## 三、C 未标、B 独立发现

### B-3a. `AloneContinuousMin` 无下界守卫

`alone := rc.AloneContinuousMin / p.aloneSatMin`。若 adapter 误传负值（如时钟回拨），`alone < 0` → `1 + aloneGain * alone < 1`，$C_{FN}$ 被错误压低。`cFN()` 不验证输入——合同在 adapter 侧。建议：在 `cFN()` 入口加 `if alone < 0 { alone = 0 }` 一行守卫，成本零、防 adapter 静默错。**不阻塞通过，阶段 4 adapter 接上时处理。**

### B-3b. `PeopleCount=0` 与 `PeopleCount=1` 退化等价

`if rc.PeopleCount > 1` 跳过折扣——0 人和 1 人都不折扣。空房不该进 decide（$P^F$≈0），但语义上 `PeopleCount=0` 不应等价于独处。建议在 `RiskContext` 文档标注"0=未知/未初始化，当独处处理（保守）"。**不阻塞。**

---

## 四、值得肯定的超规格

`ComputeLambda`：log 域 LSE 差值，与 §7 log 域滤波一致，数值稳定——不是验收要求，A 自己做对。延续阶段 1 起 A 在数值稳定性上主动超 C 参考的一贯。

A 阶段 3 立场①（不可判无独立分支）在 `Step` 控制流里兑现——整条链路没有任何 `if lambda < threshold` 分支。这不是"测试侥幸"，是结构保证。

---

## 五、与 C §18 对照

| 验收点 | C 核 | B 核 | 一致 |
|---|---|---|---|
| D1 期望损失非 argmax | ✅ | ✅ 独立代数验算 min $P^F$ 阈值 | ✅ |
| D2 Λ 不作 gate | ✅ 控制流追踪 | ✅ 独立控制流追踪（5 步骤零读 lambda）| ✅ |
| D3 C_FN 连续/单调/不归零 | ✅ 逐因子核 | ✅ 独立代数验算各因子+极端值 | ✅ |
| D4 C_FP 归一 | ✅ | ✅ | ✅ |
| D5 cd2b 不回归 | ✅ | ✅ 独立重跑 E5=0.9998 | ✅ |
| D6 取值非权威 | ✅ | ✅ 逐行核实注释 | ✅ |
| 全量回归 | ✅ 全 20（当时） | ✅ 全 24（含新增 4） | ✅ |
| 额外 | ComputeLambda LogSumExp 记功 | AloneContinuousMin 守卫 + PeopleCount=0 语义 | 互补 |

**本轮 B/C 无分歧。** B 补两项 C 未标的边界条件（合同守卫 + PeopleCount=0 语义），均为非阻塞建议。

---

## 六、放行判据

- D1–D6 全部忠实 ✅
- 全 24 测试通过，零回归 ✅
- 控制流干净（Λ 零次参与决策）✅
- C_FN 逐因子代数验算正确 ✅
- 边界条件保守（margin>0 strict、fireSince 断开复位、PeopleCount≤1 当独处）✅

**净判：阶段 3 decide 通过。裁决链 Predict→Correct→Decide 闭合。**
