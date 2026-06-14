# 评审 B：Xsensorv1 骨架审查意见

审查对象：`tools/Xsensorv1/internal/roomengine/belief/`（joint.go, bed_axis.go, mm.go, filter.go, filter_test.go）
对照基准：`doc/DBN-Zone-Room.md`（联合占用模型设计）
审查日期：2026-06-14

## 总评

骨架数学结构正确，log 域数值方案合理，T_B 门控逻辑与设计文档 §6 对准。三条发现需要阻塞合入，两条需要在 Phase 2 启动前裁定。

---

## 阻塞项（必须修才能进 Phase 2）

### B1. 包级全局状态导致并发竞态

**位置**：joint.go:7 `var numBeds int` + InitJoint()

生产 engine 按 room 起 goroutine。A 房 1 床、B 房 2 床时，InitJoint(1) 和 InitJoint(2) 互相覆盖。nBC() 在竞态下对 A 房返回 2、对 B 房返回 4 或 2——取决于调度时序。JointLen() 随之漂移，切片越界 panic 或静默计算错房间。

**修法**：numBeds 移入 Filter 结构体，nBC/JointLen 改为 Filter 方法。JointVector 不感知床数（它只是 []float64，长度由调用方保证）。InitJoint 删除。

### B2. Φ 接口装不下位置依赖的联合发射

**位置**：filter.go:98 Correct(logPhiS []float64, logPhiB [][2]float64, logPsi []float64)

设计文档 §5 的 pose 似然取决于 (x,y) 落入哪个区域，而"哪个区域"的解释依赖 B：床垫占时 bed-area lying=AtBed，床垫空时同一位置 lying=F。这是 S×B 的交叉项——Φ 必须是二维表 Φ[s][b]，不是 S 轴和 B 轴分轴加的两个一维表。

当前接口把 `logPhiS[s]` 和 `logPhiB[j][bj]` 分轴叠加，无法表达 Φ(AtBed, occ) ≫ Φ(F, occ) 这类 S-B 联合约束。Phase 2 写 emission.go 时会撞墙：pose 观测需要同时看到 S 和 B 的联合配置才能正确路由似然。

**修法**：Correct 的 Φ 参数改为 `logPhi JointVector`（与 α 同形），emission.go 直接填充 Φ[s][b] 全表。Φ_S/Φ_B 分轴分解作为内部优化保留在 emission.go 内，不暴露给 Correct。

---

## 裁定项（Phase 2 启动前需要决策，非阻塞）

### J1. T_B_unobs 的对称弛豫：vacant→occupied 速率是否应该与 occupied→vacant 相同

**位置**：bed_axis.go:34-39 K_unobs = [[1-λ, λ], [λ, 1-λ]], λ=0.15

离线时，vacant 也以每步 15% 概率「弛豫」到 occupied。稳态是 [0.5, 0.5]——数学上是「无信息」。但操作语义有歧义：空房 + 无 entry 事件 + 床垫离线，~7 秒后 P(occupied) = 0.5。这 0.5 不是「床可能被占了」——它是「我们不知道占没占」。

设计文档 §6 只写了「向无信息弛豫」，没指定目标分布。两个选项：

- **选项 A（当前）**：对称矩阵，静止分布 [0.5, 0.5]。含义：「我们完全不知道」。优点：诚实。缺点：空房离线时 P(occupied) 无理由地从 0 升到 0.5，可能触发下游对「床区有人」的误判。
- **选项 B（非对称弛豫）**：vac→vac = 1, vac→occ = 0；occ→vac = λ, occ→occ = 1-λ。含义：vacant 是吸收态，occupied 泄漏到 vacant。优点：空房离线不会凭空造占用。缺点：不对称意味着 vacant 是「特权态」——如果床垫离线时人真的上床了，模型永远学不到 occupied。

**推荐 B**。理由：一个人上床是主动事件，应由接触传感器（在线时）或雷达（pose=Lying + 床区位置）驱动，不应由「离线弛豫」驱动。occupied→vacant 的泄漏是治 cd2b 陈旧占用的核心机制，反向不需要。

### J2. λ=0.15 的半衰期是否与房间动态匹配

**位置**：bed_axis.go:15

λ=0.15 意味着占用信念每步有 15% 概率翻转。1Hz 帧率下，半衰期 ≈ 4-5 步（~5 秒）。一个离线 30 秒的床垫，P(occupied) 从 0.99 降到 ~0.04。

这个速率的人体语义：「床垫离线 30 秒后，我们几乎确信床空了」。对于熟睡中的人（可能完全静止数分钟，床垫持续在线），这是合理的——离线本身就是强信号。但对于短暂网络抖动（3-5 秒断连），λ=0.15 可能过于激进——一个网络 hiccup 就把 P(occupied) 从 0.99 砍到 0.8，恢复后需要多帧重建。

**建议**：Phase 3 标定阶段用 cd2b 回放数据验证 λ 的灵敏度曲线。骨架阶段 λ=0.15 作为初始值保留，但标记为待标定项（非固定常数）。

---

## 非阻塞项（质量改进，不阻塞合入）

### N1. Predict 内层 buildLogTBCol 被重复计算 9×
filter.go:64 — 列仅依赖 bDst 不依赖 sDst。提到 sDst 循环外。节约 9× 计算量，nBC=8 时每帧省 ~1000 次浮点加。

### N2. MarginalB 每帧分配临时切片
joint.go:126 — append 无预分配。调试 probe 调用时建议预分配 `make([]float64, 0, numStates*nBC/2)`。

### N3. LogSumExp 缺 exp 下溢跳过
joint.go:82 — 建议 `if v-max < -50 { continue }`，防极端 log 差值触发 subnormal。

### N4. 退化测试测的是恒等式
filter_test.go:37-42 — 两个均匀先验比较 P(Fallen) = 1/9 = 1/9。真正的退化验证应跑相同观测序列后比较。

### N5. jointIdx/decomposeIdx 未使用
joint.go:57-64 — 死代码，删掉或在 filter.go Predict 中替换手动索引计算。

---

## 已验证正确的项

- LogSumExp 数值稳定性（含 all--inf 守卫）✅
- 单床退化 JointLen=18 ✅
- 零床退化 JointLen=9 ✅
- Σα=1 在 Predict 和 StepNeutral 后保持 ✅
- T_B online 高自持（ε=0.01）✅
- T_B offline 衰减（λ=0.15）✅
- S/B 边缘化求和为 1 ✅

---

## 对 Phase 2 的影响

B2 修完后，emission.go 的接口是 `func EmitJoint(obs Observation, mm MMTable) JointVector`（返回与 α 同形的 log Φ 全表）。pose 似然直接填 Φ[s][b] 的交叉项——不再需要分轴分解暴露给 Correct。Ψ（coupling.go）同理：`func Compatibility(kappa []float64) JointVector`。

B1 修完后，Filter 持 numBeds，多房间并发安全。

J1 如选 B（推荐），bed_axis.go 的 K_unobs 改为 `[[1, 0], [λ, 1-λ]]`——单行修改，不影响骨架其余部分。
