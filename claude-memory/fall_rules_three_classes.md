---
name: 跌倒检测三类规则（still / silent / lost）+ cell history integral + frozen detect
description: roomengine 跌倒规则分类与自适应阈值 — still/silent/lost 三类 + cell 级回归学习 + lazy frozen 检测；常量集中 FallRulesParam（Go 原生 const 风格）
type: project
originSessionId: 08872106-789f-48c2-8a3f-bcb3bd7092ef
---

2026-04-27 用户对齐三类跌倒检测规则 + cell 级自适应 + lazy frozen 检测设计。

## 三类规则总览

| 规则 | 可见性 | 行为 | 兜底反证 | 时间窗 | 状态 |
|------|--------|------|----------|--------|------|
| still fall | 一直可见 | stand 静止 | bathroom 语义（cell ∪ room）+ Stay alarm 启用 | **risk-time 15min（严）/ non-risk 18min（宽）** | 待实现 |
| silent fall | 突然消失 | 床上方遮挡 | sleepad+radar 融合 | 120s（无 HR/RR）/ 60s（有 HR/RR + 单人） | **新设计待重构**（旧实现 track_manager.go:997-1042，需替换） |
| lost fall | 突然消失 | 屋内任意处 | 无 ExitRoom + 离门 >1m + 房间无其他人 | **按 cell areaType 分**：床/沙发 1h；toilet/shower 同 still；deny/其它 5min | Phase 1 已奠基（cell history），完整逻辑待实现 |

## 关键命名澄清

**risk-time / non-risk-time**（不是 day/night）：
- risk-time = 高风险时段（默认夜间 23:30-07:30，可由 `RiskTimeConfig` 覆盖）→ **15min（严）**：跌倒更易发、人手少
- non-risk-time = 非风险时段（默认白天）→ **18min（宽）**：×1.2 放宽
- 现有 cell.go:243-269 方向已对，Phase 1 已重构成 `StillTimeoutSec(isRiskTime bool)` + 引用 `FallRulesParam.Still`

## 时区（关键 bug）

`math_util.go::IsNightTime` 用 `time.UnixMilli(nowMs).Local()` —— `.Local()` 跟服务器时区，服务器是 UTC，所以现在算的是 UTC 23:30-07:30，对 Denver 单元等于错位 6 小时。

**解决方案 A（已选）**：threading IANA timezone 进 engine
- `RoomConfig` 加 `Timezone string` 字段
- engine 启动时从 `units.timezone` 加载（已有数据）
- `IsNightTime(nowMs, ianaTz)` 多个 tz 参数；输出统一转回 local 显示
- DST 自动处理（Go 的 `time.LoadLocation`）

## Still Fall

- pose=stand + 连续静止（|Δpos|<`StaticPosCm`=20cm 重置） ≥ effective_timeout
- 位置语义并集：**cell.Belief[0].Type ∈ {AreaToilet, AreaShower}** ∪ **room.name 含 `bathroom`/`restroom`/`toilet`（cardagg 提供）**
- Stay alarm 必须启用
- effective_timeout 来自 `cell.EffectiveStillTimeoutSec(isRiskTime)` —— 含 cell history 自适应放宽

## Lost Fall（核心，Phase 2 重点）

### 触发链

```
last_real_frame → frozen 持续 ≥ 25 帧（连续 25 同）→ frozen_run_start_ms 记录
                ↓
firmware tid=88（heartbeat 无人）or track 长期无更新 → engine 判 track 消失
                ↓
新建 pendingLostFall（room_id, frozen_run_start_ms, last_position）
                ↓
等待 effective_wait = baseWait(cell areaType) - frozen_credit
        其中 frozen_credit = frozen_duration / 2
                ↓
取消条件：① 新 track 出现 ② ExitRoom 事件 ③ room.NumberPeople ≥ 2
                ↓
超时仍未取消 → 报 Fall（verdict 未定按 real 计）
```

### Frozen 检测（lazy 反应式，不持续监听）

每帧 O(1)：
```
sig = hash(track_id, x, y, z, pose, track_confidence, remaining_time)
if sig == ts.LastFrameSig:
    ts.FrozenSameCount++
else:
    ts.FrozenSameCount = 0
    ts.FrozenRunStart = 0
ts.LastFrameSig = sig

if ts.FrozenSameCount >= 25 && ts.FrozenRunStart == 0:
    ts.FrozenRunStart = nowMs - 25*1000  // 回填到 25 帧前
```

判定：连续 25 帧字面 byte-equal（用户实测：5.5min frozen 期间所有字段 distinct combo 数=1，noise 不存在；25 阈值给极少数 firmware 抖动留余量）

### Lost-fall pending 时长

**按消失点 cell areaType**（在 `FallRulesParam.Lost` 中）：
- AreaBed / AreaSit：60min（睡觉/坐着雷达丢 track 常见）
- AreaToilet / AreaShower：与 still fall 同
- AreaDeny / 其它：5min

**Frozen credit**（关键）：
- track 消失时若 `ts.FrozenRunStart > 0`：frozen_duration = nowMs - FrozenRunStart
- frozen_credit = frozen_duration / 2（一半计入等待已过）
- effective_wait = baseWait - frozen_credit（最少 60s 兜底）

### 取消 pending

1. 新 track 出生（已有 cancelPendingByBirth，扩展支持 lost）
2. ExitRoom 事件（engine.go:520 当前 dormant 待启）
3. **room.NumberPeople ≥ 2**（待加）
4. **新 track 出生时 → 检查 room 是否有 pending lost-fall** → 取消 + 新 track verdict 直接 VerdictReal（绕过 ghost 检查；这是漏报场景下"人从盲区返回"的预期行为）

## Birth 整合到 Kalman 域（连续 ghost 信号）

`TrackState` 已有 `BirthPos TimedPoint{X, Y, TMs}`。但 `birthScore()` 是一次性的，整个 track 生命周期不再回头看。

**问题**：firmware 复用 track_id（D5F7 case 17:26:42 track 0 从 (160,-10) 跳到 (40,0) 再跳回 (190,-50)）→ 出生检查通过 + 瞬时 Kalman 残差小（相邻帧位置不离谱），但累计位移除以累计时间隐含速度不可能。

**新增 2 个连续指标**（每帧 O(1) 维护）：
```go
type TrackState struct {
    // ... existing
    MaxKalmanResidual         float64  // 峰值残差（Mahalanobis-like）
    MaxImpliedSpeedFromBirth  int      // cm/s, max over life of dist(current, birth)/age
}
```

每帧 Kalman update 后：
```go
ageSec := float64(nowMs - ts.BirthPos.TMs) / 1000
if ageSec >= 1.0 {
    implied := distInt(x, y, ts.BirthPos.X, ts.BirthPos.Y) / int(ageSec)
    if implied > ts.MaxImpliedSpeedFromBirth { ts.MaxImpliedSpeedFromBirth = implied }
}
if ts.Kalman.LastResidual > ts.MaxKalmanResidual { ts.MaxKalmanResidual = ts.Kalman.LastResidual }
```

**速度二档判定**（FallRulesParam.Lost）：

| 隐含速度 | 判定 | 解释 |
|---|---|---|
| > 200cm/s（ImpossibleSpeedCm） | **硬 ghost**，无 EnterRoom 也判定 | 老人最快 100-150cm/s，>200 必是 track-id 跳变 |
| 100-200cm/s + 无 EnterRoom | **软 ghost**（probation） | 健康成人快走可达，但需入场证据反证 |
| 100-200cm/s + 有 EnterRoom | 真人加成 | 冲进来或快走 |
| < 100cm/s（SuspectSpeedCm） | 默认真人 | 老人 60-80cm/s 步速正常活动 |

**用法**：
- D5F7 case：track 0 从 (160,-10) 跳到 (40,0) 距离 120cm 跨 0 秒 → 隐含速度爆表 → 硬 ghost
- Lost-fall 加权因子：跳过的 track，frozen_credit ×0.5（更敏感）

## Birth verdict 延迟终判（多流时序兜底）

**问题**：track 帧（monitor stream）与 EnterRoom（event stream）是两条独立流，event stream 偶发延迟 1-3s。
出生瞬时 EnterRoom 还在路上 → birthScore=0 → 误判 ghost。

**方案 A（已选）**：birth 时仅打初步分，verdict 留 Pending；`BirthFinalGraceMs=2000` 后重算。

```go
// at birth time T
ts.BirthScore = preliminaryScore  // backward window: EnterRoom in [T-3s, T]
ts.Verdict = VerdictPending
ts.BirthFinalDeadline = T + FallRulesParam.Lost.BirthFinalGraceMs

// at each tick
if nowMs >= ts.BirthFinalDeadline && ts.Verdict == VerdictPending {
    ts.BirthScore = recomputeWithFullWindow(ts.BirthPos, nowMs)  // window 扩展到 [T-3s, nowMs]
    if ts.BirthScore < ScoreGhostTh {
        ts.Verdict = VerdictGhost
    } else if ts.BirthScore >= ScoreConfirmTh {
        ts.Verdict = VerdictReal
    }
    // 中间分继续 Pending 由后续累计观测
}
```

**与 silent fall 的兼容**：silent fall 60s 窗口里 2s 占 3%，可接受；Pending 期间不立即触发 silent fall 流程。

## Cell History Integral（三类反馈）

每个 cell 累积**三类**反馈自动调节告警敏感度：

**A. FakeAlarmCount** —— 人工标 false_alarm
- alarm_events.handle_type='false_alarm' → 溯源 alarm 触发位置（triggered_at + device → iot_timeseries 90s lookback）→ cell.IncrFakeAlarm

**B. ToleratedStillCount** —— 自然观察长时静止无 alarm
- track LongStillReported=true 后自己走了（已 Phase 1 实现于 track_manager.go：从静止恢复时 → MarkToleratedStill）

**C. BlindSpotRecoveryCount**（新增） —— 人从盲区返回
- 当新 track 出生时 room 有 pending lost-fall → cancel + cell.IncrBlindSpotRecovery
- 该 cell 学到自己是「人能不能突然出现」的语义
- 未来 track 在该 cell 出生时，birth-score 加成（避免重复误判 ghost）

**决策放宽**（仅 still fall）：
```
tolerance_score = min((FakeAlarmCount + ToleratedStillCount) / threshold, 1.0)
effective_timeout = cell.StillTimeoutSec(isRiskTime) × (1 + (MaxToleranceFactor - 1) × tolerance_score)
```
封顶 MaxToleranceFactor=2.0；衰减半衰期 30 天（用 EventSec 复用现有 Decay）

BlindSpotRecoveryCount 单独用：未来 lost-fall 触发时，若在多次出现 BlindSpotRecovery 的 cell 附近消失 → 等待时间额外延长（这里就是已知盲区出口）

## Why Ghost 风险可控

- ghost 几乎只在已有真人 track 时由多径/镜面反射产生
- 单人独居：屋内单 track 消失 → 必是真人
- 双人场景：1 个 lost 时另 1 个仍在 → room.NumberPeople ≥ 2 已自动取消 pending
- 结论：lost-fall「不确定按 real 报」工程上安全

## FallRulesParam（Go const 风格，已实现 Phase 1）

`wisefido-ai/internal/roomengine/fall_rules_param.go`：

```go
var FallRulesParam = fallRulesParam{
    Still: stillFallParam{
        ToiletShowerSec:   15 * 60,  // risk-time 基线
        DenyZoneSec:       5 * 60,
        DefaultSec:        8 * 60,
        NonRiskTimeFactor: 1.2,      // non-risk × 1.2 = 18min
        StaticPosCm:       20,
    },
    Lost: lostFallParam{
        RestZoneWaitSec:    60 * 60,
        DenyZoneWaitSec:    5 * 60,
        WalkwayWaitSec:     5 * 60,
        ExitDistMinCm:      100,
        SpatialJumpFactor:  0.5,
        // 待加（Phase 2）：
        FrozenSameThreshold:    25,    // 连续 25 帧字面相同
        ImpossibleSpeedCm:      200,   // 硬 ghost 阈值
        SuspectSpeedCm:         100,   // 软 ghost（需 EnterRoom 反证）
        BirthFinalGraceMs:      2000,  // birth 终判延迟 2s，给 event stream 缓冲
        EffectiveWaitFloorSec:  60,    // 兜底最少等 60s
    },
    CellHistory: cellHistoryParam{
        FakeAlarmThreshold:           3,
        ToleratedStillThreshold:      5,
        MaxToleranceFactor:           2.0,
        DecayHalfLifeDays:            30,
        BlindSpotRecoveryThreshold:   2,   // 待加
    },
}
```

不用 yaml；Go const 风格（var + 大写命名 + 不修改约定），改参数 → 重编 → 部署。

## How to Apply（Phase 2 实施清单）

1. **修 timezone bug**（前提）：RoomConfig + IsNightTime 多 IANA 参数
2. **engine 真正消费 EnterRoom/ExitRoom**（engine.go:520 dormant，三类规则共同依赖）
3. **TrackState 加 LastFrameSig / FrozenRunStart / FrozenSameCount + 每帧维护**
4. **TrackState 加 MaxKalmanResidual / MaxImpliedSpeedFromBirth + 每帧维护**
5. **Lost-fall pending 池**：基于现有 silent fall pending 扩展 ruleType；按 cell type 分时长；frozen credit 计算
6. **Birth 时检查 room pending lost-fall**：cancel + verdict 直接 real + cell.IncrBlindSpotRecovery
7. **room.NumberPeople ≥ 2 → 取消 pending**（cardagg 已有 number_people 事件）
8. **接入 cardagg room name** + Still fall 触发器
9. **alarm_events.handle_type='false_alarm' 反馈链**（DB 轮询方案 A，回填 + 增量同源）

## Live Fixture

[case_lostfall_cd2b_11351148](../../../owl/owlBack/doc/cases/case_lostfall_cd2b_11351148/) —— 9D8A32A1CD2B 11:35-11:48 MDT 实测：
- 11:39:35 sleepad LeftBed + radar 最后真实 track
- 11:39:35-11:45:06 frozen 5.5min（337 帧字面相同 `(-90, 320, 0, pose=4, tc=60)`，发送频率 992ms ≈ 1Hz）
- 11:45:07 firmware tid=88 放弃
- 11:47:08 人重新出现（盲区返回）
- 全程 alarm_events 无 Fall 记录（漏报）

按本设计计算：frozen credit = 165s，walkway baseWait = 300s，effective_wait = 135s，从 firmware giveup 11:45:07 起算 → 11:47:22 触发；人 11:47:08 返回 → 14s 前 cancel → 不报�