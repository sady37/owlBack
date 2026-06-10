---
name: feedback-no-dynamic-threshold-modulation
description: 铁律 — WeakBio / VitalTrendLevel 等派生状态信号禁止进 alarm 决策路径（verdict/severity/routing/dedup/timing）；客户/firmware 阈值更不能被 runtime 派生信号修改
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 0566c6c8-ddbd-4e2d-b8b6-79fef2a5a715
---

**派生状态信号（WeakBio score / VitalTrendLevel / 任何聚合/趋势/累积值）禁止影响 alarm 决策。** 仅可用于：
- ✅ FE 展示（横条 / 趋势曲线）
- ✅ 审计 / 复盘 log
- ✅ 医护理解病程的辅助信息

不可：
- ❌ 独立触发 alarm（早期就拍板）
- ❌ 修改 verdict 标签（ghost/suspect/real）
- ❌ 修改 severity（Notice/Warn/Critical/Fatal）
- ❌ 修改 routing 决策
- ❌ 修改 dedup / fire / cancel 时序

**Why**：客户 / 医生 / 护士 / 运维看到 verdict=real 时必须能相信"real 就是 real"；看到 Critical 时必须能相信"Critical 就是 Critical"。一旦允许某条派生信号让 X→Y，整套报警语义对外不再一致——同样物理证据，A 客户进 real 路径，B 客户进 suspect 路径，差别只是体征长期弱不弱。整套报警约定崩塌。

**How to apply**：
- 看到 "X 时 verdict 强制 Y" / "X 时 severity 升 Z" / "X 时缩短/跳过 verifier" 这类设计，**先反问：X 是 firmware 上报的离散事件 / 人工配置的静态值，还是 runtime 派生信号？派生信号一律拒绝**
- 我（assistant）反复被挡 4 次仍换变种 = 系统性盲点，命中第 5 次必须停手不再讨论
- 已落地的违规要么主动反查回退，要么由用户拍板留下作"已知例外"（见下）

## 区分两类阈值

| 类别 | 例子 | 允许 runtime 派生信号改吗 |
|---|---|---|
| Firmware / 客户面阈值 | `spatial_config alarm.cloud_config` 里的 HR/RR/fall 灵敏度 | ❌ 绝对禁止（不仅是派生信号，runtime 任何路径都不允许；只能 layout/admin/API 人工配置）|
| 人工预设的"客户合约"阈值 | layout 里 cfg.Beds/Sits/Furnitures 等矩形定义 | ❌ runtime 不可改（详 [[v2_cutover_rules]] R-001）|
| Sensor 内部启发式常量 | `FallVerifyRealThreshold=70` / `GhostPenaltyThreshold=80` / `bedroomLostFallSilentSec*` | ⚠️ 调常量本身 OK（git 改源码 PR review）；但 **不能** 通过 runtime 派生信号自动调 |

## 历史违规记录（防变种再试）

- ❌ 早期：WeakBio 跨 80 触发独立 Critical escalation alarm — 被砍
- ❌ cardagg AlarmRouter Warn→Critical 提级（WeakBio≥80）— 被挡，未实施
- ❌ zonealarm Supervisor LeftBed/Stay DurationSec 缩短（WeakBio≥80）— 被挡，未实施
- ⚠️ **sensor fall verifier WeakBio≥80 force real**（commit `15ba836`）—
  实施了；2026-05-19 评估实际影响面 = sensor 内 derived verdict 标签 derived→derived，**不动 firmware alarm / firmware 阈值 / 客户配置**；breakdown 留 audit；用户拍板**保留不回退，标为已知此一处例外**
- 教训：assistant 想"用 WeakBio 让 fall 更宽容/更敏感"的冲动反复出现；4 次后必须停。新代码再要复用这个 pattern，必须用户重新评估，不允许直接复制 `15ba836` 模式扩散。

相关：[[v2_cutover_rules]] R-001 不动 v1 业务逻辑/功能；[[target_state_weak_bio_signal_design]] WeakBio 风险描述符设计。
