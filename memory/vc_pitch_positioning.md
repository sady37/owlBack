---
name: VC Pitch Positioning
description: Wisefido 路演核心叙事——11 条护城河 / 4 大 bucket / 3 个反共识赌注 / TAM 分层；写 pitch deck、投资人邮件、BD 话术时的统一框架
type: project
originSessionId: fea71b02-998e-48f0-886e-5278bec8a8ea
---
# Wisefido VC 路演核心叙事框架

## One-Liner（开场 hook）
> "Cameras own home security. **No one owns radar elder care.** A $20B market with zero incumbent."

## 公司基础（合规叙事）
- Wisefido, Inc. — Denver, CO 注册
- USA-only（HIPAA / HITRUST / SOC 2 / CMS / 州层合规）
- B2B / B2G only — 老人/家属对评分零可见
- 三类客户：① 大型养老连锁 + PACE（首选 Denver InnovAge） ② 州 Medicaid HCBS（首选 CO HCPF） ③ 商业健康险/再保（post-VC）

## 11 条护城河（4 bucket 分层）

### A. Sensing Moat — "We see what no one else can see"
1. 零穿戴/零摄像头 — 唯一可进卧室+卫生间
2. 浴室黑匣子破壁（StayFSM）
3. Zero-Damage 安装（10 秒重部署）

### B. AI Moat — "Every install makes the model smarter"
4. 多源融合反干扰 + Activity Window（雷达+Sleepace）
5. RoomEngine 自学习 cell-level 自适应
6. Tier-S 五大独家长期信号（recovery / frailty velocity / stability / multi-signal resonance / solo-living vitality）

### C. Architecture & Data Moat — "Contrarian bet that compounds"
7. **All-cloud 反共识架构**：放弃边缘 AI/复杂 MCU；5× 迭代速度，80% 降功耗，BoM $40
8. **原始点云全部入湖** — 唯一专有 elder-radar 数据集
9. **6–8 月 time-to-data 窗口** — 资本无法压缩（销售周期 3 月 + per-home baseline 收敛 3-5 月）

### D. Business & Regulatory Moat — "We own the data exit"
10. **Care-Not-Treatment** — 绕开 FDA 510(k)，省 5-10× R&D 成本
11. **B2B-only signed-report monopoly** — Wisefido 是唯一 HIPAA 合规数据出口

## 3 个反共识赌注（VC 最爱）
- **赌注 1：放弃边缘计算** — 行业 all-in MCU/edge AI，我们 all-cloud → 迭代速度 + 数据完整沉淀
- **赌注 2：放弃 to-C / 放弃医疗** — 不做家属 app + 不做医疗诊断 → 同时绕开 FTC + FDA
- **赌注 3：放弃摄像头** — 红海里没机会，雷达蓝海无 incumbent

## 市场分层（TAM 金字塔）
| 阶段 | 客户 | 规模 | 时点 |
|---|---|---|---|
| SOM | Denver 三角 PACE/SNF pilot | $10–50M | 2026 |
| SAM | 全美养老连锁 + 州 Medicaid HCBS | $3–5B/年 | Series A 后 |
| TAM | 居家独居老人 to-C / 子女付费 | **$20–40B** | Series B 后 |

雷达养老 ≈ 摄像头市场 1/5，但**无 incumbent**（三道壁垒同时挡住所有人：硬件 + 信号处理 + 老人场景数据）

## Tagline 候选
- "The only sensor allowed in every room of an aging American home."
- "Ambient elder care — invisible to residents, indispensable to operators."
- "Radar-native, cloud-first, HIPAA-signed. The data exit for aging-in-place."

## Self-Learning AI 精准措辞（DD 不会翻车的版本）

### 我们方案的"成分表"（哪些是真 AI / 哪些是经典）
| 模块 | 技术本质 | 算 AI 吗 |
|---|---|---|
| 多设备融合（雷达+Sleepace） | 传感器融合/贝叶斯 | 行业惯例算 AI |
| Kalman 跟踪 | 经典控制理论 1960 | ❌ |
| Ghost 检测（birth-position+速度） | 规则+阈值 | ❌ |
| **Cell-level 自适应阈值（假报历史学习）** | 在线学习/经验贝叶斯 | ✅ 真 AI |
| **z 分布聚类→沙发/床语义** | 无监督学习 | ✅ 真 AI |
| **RoomEngine 网格自学习房间几何** | 无监督空间学习 | ✅ 真 AI |
| **每户独立基线** | 统计学习/个性化建模 | ✅ 真 AI |
| StayFSM | 有限状态机 | ❌ |

### ✅ 安全用词（pitch / DD 都过得去）
- Multi-modal sensor fusion AI
- Self-learning / adaptive AI
- **Fleet-learning algorithms** ← 最 sexy，绑定 #9 time-to-data
- Per-home unsupervised spatial cognition
- Online learning with cell-level adaptation
- AI-powered fall detection
- AI radar

### ❌ 千万别说（DD 必死 + 监管风险）
- Deep learning / Neural network（没在用）
- LLM-powered / Foundation model（完全不沾边）
- Predictive AI for 急性医学事件（触发 FDA 510k 审查）
- AI diagnosis（踩 FDA 红线 + 违反自己 Care-Not-Treatment 原则）

### Closer 杀招（把 AI 和 #9 time-to-data 绑定）
> "Classical signal processing alone cannot solve radar multipath in elder homes. We layered a self-learning AI on top — every cell in every room learns its own false-alarm signature. **Our AI gets sharper while a copycat's stays dull.**"

## Traction Story（写邮件/deck 用）
elder care 4-unit pilot + 即将上门免费 OTA 升级（StayFSM 浴室高危识别 + 双源 Activity Window 跌倒检测 + 合规报告自动生成）

## 已成型的产物
- `/home/wisefido/owl/owlBack/doc/wisefido.md` — 投递用 markdown（用户 2026-05-03 决定不再单独做 1-page exec summary，直接用此 md）

## 用户对配套材料的判断（2026-05-03）
- **不做 1-page Exec Summary** — 硬件+AI 产品面谈+实测的杠杆远大于纸面材料
- **不预先做 unit economics 详算** — 等 VC 真有意愿再细谈，那时还能根据 VC 偏好（ARR multiple vs LTV/CAC）调维度
- **核心 deck = wisefido.md + 现场 demo**，其他材料按需触发

## Why: 用户 2026-05-02 准备投递 Rockies Venture Club（rockiesventureclub.org/pitch）

## How to apply:
- 写 pitch deck / 投资人邮件 / BD 话术时，按 4 bucket 框架展开，不要散开 11 条
- 强调 3 个反共识赌注（这是技术 VC 的兴奋点）
- TAM 一定用三层金字塔（SOM/SAM/TAM）讲，不要平铺
- 永远把 #9 time-to-data 作为 closer——"capital cannot compress this"
- 严守 Care-Not-Treatment 边界，不要被诱导讲"医疗诊断/疾病预测"
- AI 措辞按上面"成分表"严格区分，对外只用安全词清单
- 不要主动产出"投递配套材料"（exec summary / detailed financials），除非用户明确 ask 或 VC 已表达意愿
