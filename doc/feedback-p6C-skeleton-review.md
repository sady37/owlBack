# feedback-p6C 附件 — 评审组 C 对 A 阶段 1 骨架真 code 的独立审查

> 审查对象：commit `6baa31f`（A 执笔，`tools/Xsensorv1/`）。
> 审查方式：pull 真 code，对照 [[feedback-p6C-acceptance]] T1–T5 + 独立查规格未覆盖的隐患。
> 角色：C 审 A 写的 code（C 未参与执笔，独立性成立）。
> 工具链注：容器限 go 1.22（go.dev 不在白名单），临时降版编译验证，未改 A 的 go.mod（1.25.0）。

---

## 总判：**骨架通过，质量超出验收要求；一个结构问题须修，两个小点建议。**

T1–T5 全过（实测，非纸面）。A 的实现**优于 C 的参考稿**——C 参考是线性域，A 自己上了 log 域 + LogSumExp，满足 DBN-Zone-Room §7 数值稳定要求。核心不变量（ε≪λ 复现 staleness、§C vac 吸收）实测兑现。

---

## 一、验收结果（T1–T5 实测）

| 验收 | 结果 | 实测 |
|---|---|---|
| T1 Σα=1 守恒 | ✅ | log 域 LogNormalize，nb=0/1/2/3 多步后 Σexp(α)=1 |
| T2 单床退化 + numBeds 字段 | ✅ | size 9/18/36/72；`NumBeds()` 显式持有（B1 闭合） |
| T3 maxBeds panic | ✅ | -1/4/99 panic，maxBeds=3 合法 |
| T4 D-2 ε≪λ staleness | ✅ | **离线后陈旧 occ 14 步(≈14s)蒸发到 vac 主导，< 30s 窗**；在线对照不蒸发 |
| T5 §C vac 吸收 | ✅ | 空房离线 120s 后 P(occ)≈0，不伪占用 |

**D-2 是关键验收**：ε=1e-2、λ=0.05，半衰期≈14s，陈旧 occ 在 30s 窗内蒸发——"一个不等式 ε≪λ 替代所有 staleness/TTL"实证兑现，cd2b 漏报根因（陈旧 occ 永不衰减）在新架构治本。

## 二、C 认可 A 超出要求的两处（记功，非挑刺）

1. **log 域 + LogSumExp（A 自加，C 参考无）**：`filter.go` 全程 log 域，`LogSumExp` 防下溢（跳过 <-50 项防 subnormal）、`logP(p≤0)=-Inf` 边界干净。这是 §7 要求、C 验收规格**未强制**但 A 主动做对的——为阶段 2 的极小 $\varepsilon_{art}$（§E）log 域诊断预留了数值空间。**A 比 ground truth 要求更前一步。**
2. **logKernel 预存**：T_B 核 log 化预存（`makeLogKernel`），Predict 内层不重复 log，性能正确。

## 三、🔴 必修：两套 belief 包重复提交（结构问题）

commit `6baa31f` 把 belief 包提交进**两个路径**，内容不同：
- `internal/belief/`（6 文件）
- `internal/roomengine/belief/`（6 文件，commit stat 列为正本）

**问题**：①包名都是 `belief`，并存造成 import 歧义、审查对象不明；②两套内容不同（疑早期试写 vs 定稿），后续维护会 drift；③C 实测的是 `roomengine/belief/`（与 Tsensor 路径一致），`internal/belief/` 来源/用途不明。

**C 要求**：确认正本（建议 `roomengine/belief/`，与 Tsensor `internal/roomengine/belief/` 路径一致、便于阶段 4 对照 diff），**删除另一套**。这是 D-1 baseline 对照纪律的延伸——审查对象必须唯一。

## 四、🟡 建议（不阻塞，阶段 2 前处理）

1. **`bedOnline` 长度与 numBeds 的契约未断言**：`Predict(online bedOnline)` 内 `if j < len(online)` 容错，但 `len(online) != numBeds` 时静默用 false（离线）。建议加断言或文档明确契约——否则阶段 2 adapter 传错长度会静默当离线，掩盖 bug。
2. **参数是形态占位，须在阶段 2 标定前显式标记**：`defaultBedAxisParams` 的 ε=1e-2/λ=0.05 是 feedback-p6C §5 的单 case 量级锚，**非标定值**。注释已写，但建议 A 在进 emission/coupling 前确认这些不被下游当权威值硬编码（铁律：定形态、不定参数）。

## 五、阶段 2 放行判据（C 立场）

- 🔴 三（两套包）修了 → 审查对象唯一 → 骨架正式通过。
- 阶段 2（emission Φ/§D gate、Ψ/§E mixture、transition/§A neighbor）**仍按硬序**：neighbor transition 等 A 的 ρ_xroom 方程；emission/coupling 的非 neighbor 部分可在骨架修后起。

**C 净判：A 骨架数值正确、质量超 C 参考；修掉两套包重复即正式通过。B 可独立复审。**
