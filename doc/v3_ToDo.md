# sensor v3 ToDo — 大改造期 backlog

收集需要 v3 周期才能合理做的项。共同特征：要么牵涉**架构性重写**（如 logicID 真融合 / multi-target tracking），要么需要**长期数据积累**（30 天滑窗），要么是**多人共床 / 多 radar 真合并**这种 v2 单 person/单 radar 假设之外的扩展。

v2 阶段（P4 已收尾）保留这些设计能力的"占位接口"——v3 重做时直接接，不需重新 audit。

---

## J. Standing 双 radar AND 规则（v3 sensor logicID 真融合）

**Why**：v2 多 radar 同 room 时 Standing 取简单 max（[[target_state_per_device]] D5）；多个 radar 报站立可能同一个人也可能不同人——区分不了。v3 sensor logicID 真融合后，能用 AND 规则（多 radar 都报某区域有人才算）排除盲区 / 静坐误判。

**How to apply**：
- 等 v3 sensor 跨 device 真融合落地（建 logicID 上层 entity）
- Standing 判定改为 AND（多 radar 一致才升 standing_continuous_min）
- 需要 sensor 区分 "我看不到" vs "我看到没有"（盲区 vs 真无）

**前提依赖**：sensor logicID 融合层（v3 大改造）

来源：[[p4_next_session_prompt]] FU12

---

## K. 多人共床 cap 0..2 建模

**Why**：v2 单 person 假设——BedState.TrackNumber 只支 0/1。真实场景夫妻共床、家属探望临时上床等，需 cap 0..2（≥2 也归 2 防数值爆）。

**How to apply**：
- sensor bed FSM 加 multi-person count（Vacant→Occupied 时支持 count > 1 路径）
- Sleepad firmware 不分人，但 radar track 可计数；融合规则：sleepad InBed 决定 Occupied，radar count 决定 number
- BedState.TrackNumber 改为 uint8 但语义 cap 2
- cardagg max-merge 处理多 radar count 时简单 max（与 J 同步推进）

**前提依赖**：J（多 radar 真融合）或独立做但需 sensor 能区分多人

来源：[[cardagg_v1_to_v2_migration_audit]] §S5b deferred

---

## L2. Track-ID 重新关联（multi-target tracking 架构改造）

**Why**：当前 engine 直接用 `tm.tracks[radar_track_id]` 100% 信任 firmware ID。firmware 有时 split 同一人为多 ID（D5F7 case A），有时不同人复用同一 ID。

**How to apply**（标准 multi-target tracking）：
- Engine 维护内部 track ID 空间，与 radar ID 解耦
- 每帧 Hungarian / NN 匹配：observation → existing internal track（cost = Kalman 残差 + pose 一致性 + 时间）
- radar_track_id 只作辅助提示，不是身份
- 涉及面大：processFrameAt 段 1 重写 / TrackState 加 RadarTrackID 字段 / outputs map 改 internal id / 大量单测重写

**用户原话**："track_id 有时可能有错，但相信厂家也尽力处理过，我们处理也未必更佳" —— 倾向相信 firmware ID 直到证据反过来。

**前提依赖**：先把 L1 镜面 ghost 做掉看 ghost 率；ghost 主因是镜面还是 ID 错乱靠数据决定

来源：[[ghost_detection_todos]] TODO 2

---

## M. Walking speed health 4 扩展点

**Why**：老人步速是"第五大生命体征"（< 60 cm/s 跌倒高危，60-80 衰弱前兆，80-120 健康）。v2 只用 20 cm/s 做 Move/Stand 二值兜底（cell_learning.go::LearnParams.MoveSpeedCms），未利用更精细的速度分级。完整临床数据在 [`owlBack/doc/walking_speed_health.md`](walking_speed_health.md)。

**How to apply**（按优先级）：
1. **步速健康指数**：per-track EMA 速度 + 30 天滑窗趋势 → FrailtyTrendAlarm 告警家属
2. **慢走 + 高风险区 → fall 先验**：进卫生间速度 < 40 cm/s 时，fall 判定阈值放宽 2 倍
3. **Move 阈值二级化**：Cell 加 SlowMove/FastMove 计数，做 per-resident baseline
4. **> 150 cm/s 噪声过滤**：雷达多径，连续 N 帧才丢

**前提依赖**：(1)(3) 需 30 天数据积累；(2) 需 shadow mode 验证 P/R

来源：[[walking_speed_health]]

---

## v3 触发时机

- 上面任一项有客户投诉或运行数据强信号 → 启动 v3 周期
- 至少凑齐 J + K（多 radar 真融合 + 多人共床）才值得开 v3，单点改造不划算
- L2 可独立但要先看 L1 落地后的 ghost 残余率
- M (1)(3) 任何时候可启动但要等数据
