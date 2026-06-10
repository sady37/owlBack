---
name: number_people=0 ExitRoom 兜底（PR-C）
description: firmware 部分场景不发 ExitRoom，用 number_people=0 兜底跳过 lost_fall pending；实测 D523 早 track=88 36-44ms，三例 100%
type: project
originSessionId: 28e63f8a-abf3-408d-b6c4-76fe69672bff
---
2026-05-02 PR-C：解决 D523 firmware 不发 ExitRoom 导致 lost_fall 误报的工程兜底。

## 实测时序（3 例独立样本，零反例）

| 离场点 | number_people=0 | track_id=88 | 间隔 |
|---|---|---|---|
| D523 00:22:00 | 00:22:01.538 | 00:22:01.578 | +40ms |
| D523 00:35:48 | 00:35:48.870 | 00:35:48.906 | +36ms |
| Kitchen 00:36:54 | 00:36:54.037 | 00:36:54.081 | +44ms |

**number_people=0 一律领先 track_id=88 36-44ms。** 选 number_people 不选 track=88 的原因：① 早 40ms ② 同在 iot:event:stream（轻量）③ 语义干净（"屋内人数→0"），track=88 只是 firmware 心跳产物。

## 实现要点

- 新字段 `TrackState.LastObservedMs`（[track.go:124](../../../owl/owlBack/wisefido-ai/internal/roomengine/track.go#L124)）：仅 PushPoint 更新，miss tick 不动；与 LastUpdateMs 区分（后者每 tick 都刷）
- 新字段 `TrackManager.lastNumberPeopleZeroMs`：记录最近一次 number_people=0 时间
- `RecordRadarEvent` 收到 `EventName="NumberPeople" && NumberPeople==0` → 取消已存在 pendingLostFalls（同 ExitRoom 路径）
- 入池前检查（[track_manager.go:922](../../../owl/owlBack/wisefido-ai/internal/roomengine/track_manager.go#L922)）三道关卡：
  1. `ts.FrozenRunStart == 0`：track 失锁前未进入 frozen 残影状态
  2. `lastNumberPeopleZeroMs > 0 && ts.LastObservedMs > 0`
  3. `|lastNumberPeopleZeroMs − ts.LastObservedMs| ≤ NumberPeopleZeroFallbackMs(60s)`
  - 关键：比对基准是 `ts.LastObservedMs` 不是 `nowMs`——nowMs 比 number_people=0 晚 ~10s（MaxMissCount=10），落不进窗口
- 新参数 `FallRulesParam.Lost.NumberPeopleZeroFallbackMs = 60000`（60s，对齐用户"t1+60s 前到达视为正常离房"语义）

## frozen ↔ number_people=0 互斥（关键设计约束）

firmware 状态机里两者互斥：
- frozen 期间 firmware 维持"屋内有人"判定，持续发同帧（pose/x/y 字面相同）→ **绝不会发 number_people=0**
- number_people=0 ↔ firmware 自己确认屋空 → 必然不在 frozen 状态

若两者同现说明是 firmware-frozen 残影**结束后**的新状态（CD2B 类盲区返回 case：人入盲区 → frozen 5.5min → firmware 给 up → number_people=0）。这种情况下 number_people=0 的语义已经不是"正常离房"，而是"firmware 终于放弃残影"——人可能仍在盲区里跌倒，必须保留 pending 等 birth-recovery 取消或正常超时报警，**不能误抑制**。

## "多少算是噪音" — 用户对齐 2026-05-02

frame-level：**0 字节差异**（不放宽 frameSignature）。
- frameSignature 7 字段（TrackID/X/Y/Z/Pose/TrackConfidence/RemainingTime）任一变化即判 reset
- 不放宽因为：① CD2B/D5F7 真 frozen 就是 byte-equal，放宽模糊语义 ② 放宽 X/Y 容忍会把低速移动（如床上翻身）当 frozen，frozen credit 错算缩短 lost_fall wait

time-level：**也不做**（讨论过 6min lookback，回滚了）。
- 没有实测数据证明 "frozen + 抖动 + 给 up" 这种 case 真实存在
- 多 1 个 sticky 字段 + 1 套时间判断 = 多一处可能 bug 的代码路径
- YAGNI——等 prod 真出现 frozen 漏报再加，当前严格 `FrozenRunStart > 0` 已覆盖所有已知样本（CD2B / D5F7 都是纯 byte-equal）

适用边界（PR-C 不接管的 case）：
- **倒地后偶尔被检测到（间歇可见）**：firmware 一直在发 track 帧（即使是残影），engine 段 2 不触发，track 不消失，**根本不走 PR-C 路径**——这是 firmware pose 误判问题
  - bathroom 场景：靠 `still_fall` 兜（pose=stand 静止 ≥15min）
  - 非 bathroom 场景：暂时无解，待将来"firmware Fall + 持续静止"组合规则覆盖
- **真 frozen + 中间 1-2 帧抖动 + firmware 给 up**：理论存在，实测无样本，先不写预防性代码
- `ParseRadarTrackEvents` 接受 `number_people` envelope category，规整为 `EventName="NumberPeople"` 内部判断

## 单元测试覆盖

`track_manager_silent_fall_test.go` 末尾 6 个测试：
- `TestLostFall_NumberPeopleZeroSkipsPendingCreation` — 入池前命中窗口 → 跳过创建
- `TestLostFall_NumberPeopleZeroCancelsExisting` — 入池后到达（non-frozen）→ 取消已存在
- `TestLostFall_NumberPeopleZeroOutsideWindow` — 远早于失踪（>60s）→ 不抑制，正常入池
- `TestLostFall_FrozenOverridesNumberPeopleZeroSkip` — frozen+number_people=0 同现 → 不跳过（CD2B 盲区返回保护）
- `TestLostFall_FrozenPendingNotCancelledByNumberPeopleZero` — frozen pending 不被 number_people=0 取消
- `TestParseRadarTrackEvents_NumberPeople` — envelope 解析

测试 helper：`settleAtPos(tm, tid, x, y, z, pose, startTms, frames)` 喂 N 帧同位置让 Kalman 速度收敛——避免 runFramesUntilReal 留下高速度后 PredictOnly 把消失点甩出屋外（CellAt=nil → checkLostFall 返 false）。其他 lost-fall 测试也可复用。

**Why**：D523 同台设备一晚两次离场（00:22 / 00:35）都 firmware 没发 ExitRoom，导致两个 false-positive lost_fall（00:30:27 / 00:41:02）。根因在 firmware enter_zone 配置不全，但工程上 number_people=0 是稳定可观测的兜底信号。

**How to apply**：未来如果其它 firmware 也出现"number_people=0 早于 track=88"模式，本兜底自动生效；如果某天 firmware 时序反转（先 track=88 再 number_people=0），cancel-existing 路径兜底兼容。
