# 评审 B：A 阶段 1 骨架真 code 独立审查

审查对象：commit `a092bb7`（`tools/Xsensorv1/internal/roomengine/belief/`）
对照基准：`doc/DBN-Zone-Room.md` §1/§6/§7/§C + `doc/feedback-p6C-acceptance.md` T1-T5
审查阶段：阶段 2（实现审查）——对照设计文档方程，查代码是否忠实实现
角色：B 审 A 写的 code（B 未参与执笔，独立性成立）

---

## 总判：通过。T1-T5 实测全过，§7 方程实现正确，B1/P-5/§C 约束全部落地。

与 C 独立验证一致：log 域 + LogSumExp 满足 §7 数值稳定要求，ε≪λ staleness 复现实证兑现。一条可优化、一条与 C 共识。

---

## 一、验收实测

| 验收 | 结果 | 注 |
|---|---|---|
| T1 Σα=1 | ✅ | 0/1/2/3 床各 50 步，`exp(LogSumExp(α))` = 1 |
| T2 退化 | ✅ | size 9/18/36/72；NumBeds() 显式；nb=0 MarginalS ≡ Prior |
| T3 maxBeds | ✅ | -1/4/99 panic；3 合法 |
| T4 D-2 staleness | ✅ | 离线 14 步跌破 0.5（λ=0.05，半衰期≈14s < 30s）；在线对照不蒸发 |
| T5 vac 吸收 | ✅ | 空房离线 120s P(occ)≈0（§C vac→occ=0） |

---

## 二、方程→代码逐条对照

### §7 联合滤波

设计：$$\bar\alpha_t = \sum_{S',\{B'^j\}} T_S(S|S')\prod_j T_{B^j}(B^j|B'^j)\,\alpha_{t-1}$$

代码 `filter.go Predict`：
- `logA[sFrom][sTo]` = log T_S ✅
- `Σ_j logK[j][bedOf(bFrom,j)][bedOf(bTo,j)]` = Σ_j log T_B^j（因子化）✅
- `logAdd` 累积 → LogSumExp 归一化 ✅
- T_S 不被 B 调制（logA 预存不变）✅

设计：$$\alpha_t \propto \Psi_t \cdot \Phi_t \cdot \bar\alpha_t$$

代码 `filter.go Correct`：元素级 `alpha[i] += logPsi[i] + logPhi[i]` → LogNormalize ✅
阶段 1 nil = 中性（log 0 → 线性 1）✅

### §6 转移纯因子化

- T_S 与 T_B 各自独立预存，Predict 内只做 log 域加 ✅
- 耦合不进 T（Ψ 在 Correct 施）✅

### §C K_unobs 单向泄漏

`bed_axis.go kUnobs()`：
- `BOcc → BVac = λ, BOcc = 1-λ` ✅
- `BVac → BVac = 1, BOcc = 0`（vac 吸收）✅

### B1 numBeds 显式持有

`JointSpace.numBeds` 字段（非包级全局），`Filter.NumBeds()` 返回 ✅

### P-5 maxBeds 硬 bound

`NewJointSpace` panic 超界，`maxBeds=3` ✅

---

## 三、C 未发现的一项（非阻塞，阶段 2 前可优化）

### Predict 内层 logB 重复计算

`filter.go Predict`：对每个 (bFrom, bTo, sFrom, sTo) 四元组，内层循环体重新计算 `Σ_j logK[j][...][...]`。此和仅依赖 (bFrom, bTo)，不依赖 S。|B|=3 时 5184 次内层执行 × 3 次浮点加 = ~15k ops/Predict，可忽略。但结构上更干净的写法：

```go
// 预存 bmaskN×bmaskN 的 log T_B 表（在 sTo/sFrom 循环外）
logTB := make([][]float64, js.bmaskN)
for bFrom := 0; bFrom < js.bmaskN; bFrom++ {
    logTB[bFrom] = make([]float64, js.bmaskN)
    for bTo := 0; bTo < js.bmaskN; bTo++ {
        for j := 0; j < nb; j++ {
            logTB[bFrom][bTo] += logK[j][bedOf(bFrom,j)][bedOf(bTo,j)]
        }
    }
}
```

然后内层用 `logTB[bFrom][bTo]` 替代逐 bed 循环。语义等价，S 循环体从 O(numBeds·nBC²) 降到 O(nBC²)，且逻辑更清晰——T_B 因子化是 Predict 的第一步，不应嵌在最内层。

**建议**：阶段 2 重构，不阻塞阶段 1 通过。

---

## 四、与 C 共识的一项

`bedOnline` 长度与 `numBeds` 的契约未断言：`len(online) != numBeds` 时静默当离线。C 已标，B 认同——阶段 2 adapter 接上后，传错长度是隐蔽 bug。建议 `Predict` 入口加 `if len(online) != nb { panic }` 或至少文档标注。

---

## 五、放行判据

- 骨架通过 ✅
- 两套包重复：C 发现、A 已修 ✅
- 非阻塞优化 + bedOnline 契约：阶段 2 处理

**B 净判：骨架数值正确、方程实现忠实、核心不变量实证兑现。放行进阶段 2。**
