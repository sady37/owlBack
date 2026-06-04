# Feedback 驱动的 lying/sit 学习 + Layout 权威模型 + AI object 下发策略

> 分支：`feat/feedback-lying-learning`。本文是需求+规格，驱动 backend(feedback.go/cell engine) + FE(alarm Handle 表单) 两侧改动。
> 关联：[[alarm_handle_form_by_category]]、[[feedback_work_on_main_no_branches]]、`feedback.go`、`cell.go`、
> static_reflector.go(金属反射体自学习)、[[layout_drift_wall_expansion_policy]]。
> 起于 2026-06-04 关于"layout 人工真相是否值得商酌"的讨论。

---

## 1. Layout 权威模型（推翻"layout=人工真相不可改"）

**原则**：人类无精确工具画 layout，人画的是**不精确先验**，不是真相。次序应为：

```
人工绘制(粗先验) → AI 修正(默认接受) → 人类明确否决(sticky，永久不再绘入该点)
```

代码已有骨架：cell `Source ∈ {SourceHuman, SourceLearned, Unset}`；自学习产 SourceLearned 且默认生效；
demote-on-real-activity 自纠。差两块才到此模型：
- **sticky 否决**：新增 `SourceHumanVeto`（或 cell `LearnBlocked` 标志）——人否决某点 → 学习器**永久跳过**，跨重启保留。= "下次不再绘入该点"。
- **AI 可修正人画的不精确几何**（现 `SourceHuman 不覆盖` 太死）。

**两类修正分级**（命门）：
- **加性 / 几何**（补金属点、扩墙含漂移）= 低风险 → **默认接受**，靠 demote 自愈 + veto 纠偏。
- **会抑制跌倒的修正**（lying/deny/bed）= 高风险（漏报不可逆）→ **不默认接受**，走 §3 人类 feedback 门控。
- **翻人类语义标签**（改人标的 bed/Enter）= 不默认翻，**只加不改**。

provenance/HIPAA：每 cell 带 Source（human/learned/vetoed）= "人画 X→AI 修 Y→人否决回 X" 全审计链。

---

## 2. AI object 下发策略（firmware 会消费 declared areas）

| 类别 | 下发? | 理由 |
|---|---|---|
| **AI 新生成 object**（lying/deny/metal 等 fall-affecting） | **否**。做 server 端 overlay，DBN 在**原始雷达数据**上判 | 下发会改 firmware 区域标注/处理 → 污染我们用来**验证这个猜测**的原始数据；server 端可逆、可审计、不动 firmware；契合 DBN/shadow 纪律 |
| **AI 修改既有几何（仅外扩）**（墙/边界外扩含漂移点） | **可下发** | **只 enlarge 永不 shrink**——扩只纳入更多检测不排除真目标；缩会漏真目标。配合 ≤30cm、≤雷达最大边界（[[layout_drift_wall_expansion_policy]]） |

**铁律**：下发的 AI 几何修正**只许 enlarge 不许 shrink**。

---

## 3. Feedback 驱动 lying/sit 学习（安全关键，§1 的"会抑制跌倒"类）

cell engine 学 lying/sit 会**抑制该处跌倒报警** → 不能自动学，必须人类 alarm feedback 门控。
现有 `feedback.go AlarmFeedbackIngester` 已做骨架（从 alarm_events 拉 false_alarm/verified 喂 cell counter），本节细化。

### 3.1 verified（人类确认**真**跌倒）
- **擦除该处 AI/feedback 学到的 lying/rest-zone**（`SourceLearned`）→ 相当于没学过。**最强安全律**：真摔证明该处抑制是错的，即刻擦除。
- **人工画的 bed（SourceHuman）不擦**（刻意语义 + 床上真摔=摔下床/塌陷真实存在；擦了反漏）。但**记录该真摔**（`cell.RealFallCount++`）作 forensic；若同一人工 bed 反复 verified 真摔 → 触发"该 bed 标注存疑"人工复审（不自动擦）。
- 实现：verified Fall → 该 cell 若 `RestZoneConfirmed/lying` 且 `Source==SourceLearned` → 清空回 Unknown；`RealFallCount++` 始终记。

### 3.2 false_alarm（误报）→ 按 reason 路由

| reason | 动作 | 语义 |
|---|---|---|
| Sit on Chair / **Short** Sofa | `RestZoneConfirmed++`（sit） | 人真坐在此 → 强化 sit 区 |
| **Behind Chair / Table** | `RestZoneConfirmed++`（sit） | 人在家具后被遮，仍是坐区 |
| Sit in Wheelchair | `RestZoneConfirmed++`（sit；mobile decay 自滤） | |
| Error Pose Detection / Electric / AC Interference / Out of Detection Range / Other / Unknown | **不加 sit/lying 学习计数器** | 传感器误差/干扰非空间属性，不污染 lying/sit 学习（Electric/AC 仍可进独立 `GhostCount`，那是 ghost 学习非 lying） |
| **Lying Lounge Chair / Long Sofa** | **立即交互询问**（见 3.3） | 躺=床级抑制，比坐强，需人确认 permanent 才学 lying |
| Other / Unknown | 正常标注（`FakeAlarmCount++` 或跳过，不进 lying/sit） | |

**关键区分**：**坐类**（椅/短沙发/轮椅/家具后）= 渐进 sit 计数自动累；**躺类**（躺椅/长沙发）= 抑制更强（床级），**必须人确认** permanent 才学 lying。

### 3.3 Lying Lounge Chair / Long Sofa —— confirm 后立即交互
- **长期放置？现在更新 layout？** 是 → `MarkRestZoneByFeedback(lying)` 该 cell（永久学 lying）。
- **临时放置？2H 内不报警？** 是 → `cell.FallSuppressUntilMs = now + 2h`（**机制已存在**，非 layout 改）；否 → 不动。

---

## 4. Backend vs FE 拆分

**Backend 可独立做（本分支）**：
- §3.1 verified-fall 擦除 `SourceLearned` lying（人工 bed 不擦 + `RealFallCount++` 记录真摔）。
- §3.2 reason→counter 路由细化：坐类→RestZoneConfirmed、no-learn 类不进 sit/lying。
- §3.3 临时 2H 抑制：`FallSuppressUntilMs` 路径（reason="Lying Lounge/Long Sofa" + 选"临时2H"时置位）。
- cell 学到 lying/metal/deny 不下发，仅 server overlay（§2）；几何外扩下发只 enlarge（§2，与 [[layout_drift_wall_expansion_policy]] 合流）。

**FE 依赖（另开，producer-first：FE 先落 reason 选项与交互）**：
- alarm Handle 表单新增 reason 选项：`Behind Chair/Table`、区分 `Short Sofa` vs `Long Sofa/Lounge Chair`、`Out of Detection Range`、`Error Pose Detection`（对齐 [[alarm_handle_form_by_category]]）。
- §3.3 Lounge/Long Sofa 的**交互问询 UI**（permanent→更新 layout / 临时→2H 抑制 二选一）。
- §1 sticky 否决的 FE 入口（人否决 AI 修正点 → 写 veto）。

**注**：backend 路由可先就位（认这些 reason 字符串），但在 FE 真正发出这些 reason 前不会被触发；reason 字符串契约以 FE/alarm.go 常量为准，禁字面量复写（CLAUDE.md 1.1）。

---

## 5. Source 分级（gate-list 实施 2026-06-04）

`Source` 加 `SourceFeedback=4`，信任序 **Human > Feedback > Learned**：
| Source | 来源 | verified 真摔 | 可被覆盖 |
|---|---|---|---|
| `SourceHuman` | FE 刻意画(SetPrior) / interactive-confirmed lying | **不擦**(+RealFallCount++ 记录, 反复→人工复审) | 否(神圣) |
| `SourceFeedback` | alarm feedback / 观测刷新(MarkRestZoneByFeedback: sit / auto-learn AreaSit) | **擦回 Unknown** | 可被 Human 覆盖 |
| `SourceLearned` | 纯自动(mirror/static-reflector/AutoDeny) | **擦回 Unknown** | 可被 Human/Feedback 覆盖 |

落地：`MarkRestZoneByFeedback→SourceFeedback`(原 SourceHuman);`Cell.ClearNonHumanLearnedZone()` verified 真摔擦
非 Human 的 {AreaBed/Sit/Toilet/Shower/Deny};`feedback.go` verified 分支调它(route `cleared_non_human_zone`);
MarkMirrorBounce 晋升不覆盖 SourceFeedback。**exempt 行为零变**(`fall_exempt` 要 SourceHuman+Conf99，feedback
本就 Conf95 不 exempt)。`fall_exempt` 的 Conf99-hack 现可由 SourceFeedback 简化(留 backlog，本次不动避风险)。

## 6. belief 侧后续（gate-list 完成后做，用户 2026-06-04 定序）

把 `Source` provenance 带进 belief **输入层**作 **Conf/信任权重**(非硬标签)——`geomFromGrid` 的 InBed/InToilet 观测
Conf 按 Source 缩放(Human~0.95 / Feedback~0.6 / Learned~0.4)。理由：
- DBN-native 表达 Human>Feedback>Learned，**软先验替硬 exempt 闸**——强跌倒证据(kinematics/firmware)能盖过弱学习先验，不漏真摔;
- **不违 [[feedback_no_dynamic_threshold_modulation]]**：是已有几何输入(InBed)的**可靠度元数据**，非派生风险调制器;
- cutover 时 DBN 用此**吸收** gate-list 的 `fall_exempt` 硬 Source 判定(一处表达)。
落点：`belief_input_normalization.md` 的 `ObsBedOccupied`/geom 观测 Conf。**本次 gate-list 不含此项**。

## 7. 决策记录（2026-06-04 用户拍板）
- 人工 bed 不擦；verified fall 记录 `RealFallCount++`（反复真摔触发人工复审，不自动擦）。
- AI 新生成 fall-affecting object 不下发（overlay+DBN）；几何外扩可下发（只 enlarge）。
- 坐类渐进自学；躺类（Lounge/Long Sofa）人确认 permanent 才学 lying，临时走 2H 抑制。
- Source 三分(+SourceFeedback)；verified 真摔擦非 Human rest/deny，保 FE bed。
- gate-list 在 main 实施(本次)；belief Conf 加权(§6)gate-list 后做；FE 部分另开新会话。
