# Lost Fall 演进笔记 — PR-C: number_people=0 ExitRoom 兜底

## 最终版 PR-C 改动（去掉了 lookback 的过度设计）

### 数据结构（最小）
- `TrackState` 加 [`LastObservedMs int64`](../wisefido-sensor/internal/roomengine/track.go) — 这个保留，用于 number_people=0 时间窗对齐
- `FallRulesParam.Lost` 加 [`NumberPeopleZeroFallbackMs = 60000`](../wisefido-sensor/internal/roomengine/fall_rules_param.go)
- `RadarTrackEvent` 加 `NumberPeople int` 字段
- `ParseRadarTrackEvents` 接受 `number_people` envelope category

### 行为联动（三处）
1. **frozen 检测**：维持原样，未动 frameSignature
2. **`RecordRadarEvent` 收 NumberPeople=0**：取消 `p.FrozenStartMs == 0` 的 pending（frozen pending 保留）
3. **入池前 skip**：`ts.FrozenRunStart == 0 && |lastNumberPeopleZeroMs − LastObservedMs| ≤ 60s` → skip

### 6 个单测全绿
（去掉了 FrozenLookbackSurvivesSignatureNoise 和 FrozenLookbackExpired 两个）

## 噪音容忍的最终立场

| 维度 | 容忍度 | 原因 |
|---|---|---|
| **frame-level** | 0 字节 | 真 frozen 是 byte-equal，放宽会模糊语义、误判低速移动 |
| **time-level** | 不做 | YAGNI：实测无 frozen+抖动样本 |

如果以后 prod 真出现"frozen 期间被抖动 reset，导致 number_people=0 误抑制 lost_fall"的样本，到时候再加 lookback 不迟——而且那时会有真实数据指导窗口大小，不像现在拍脑袋 6min。

---

## 核心场景：无 ExitRoom 但有 number_people=0

### 场景 1：标准 D523 离场（实测 3 例，PR-C 主要解决的 case）

```
时序                                           engine 行为
─────────────────────────────────────────────────────────────────
00:22:00.526  最后真帧 track=0                ts.LastObservedMs = 22:00.526
                                              ts.FrozenRunStart = 0 (没经历过 frozen)

00:22:01.538  number_people=0 事件到达        tm.lastNumberPeopleZeroMs = 22:01.538
                                              （此时 pendingLostFalls 为空，没有可取消的）

00:22:01.578  track_id=88 心跳（被 ParseRadarTracks 过滤为空帧）

00:22:11      MaxMissCount > 10 触发判失锁     入池前检查：
              ↓                                ① FrozenRunStart == 0 ✓
                                               ② lastNumberPeopleZeroMs = 22:01.538 ✓
                                               ③ gap = |22:01.538 − 22:00.526| = 1s ≤ 60s ✓
                                               → 跳过 pending 创建 ✓

                                              结果：不入 lost_fall pending → 不会误报
```

**这就是 D523 昨晚两次离场（00:22 / 00:35:48）应该的行为**——之前没有 PR-C 时这两次都误报了 lost_fall（00:30:27 / 00:41:02）。

### 场景 2：firmware 时序反转（防御性兜底）

```
假设 firmware bug：track=88 先到，number_people=0 几秒后才到

00:22:01      track 失锁判定 → pending 创建（FrozenStartMs=0）
00:22:05      number_people=0 终于到达
              ↓
              RecordRadarEvent 路径：
                p.FrozenStartMs == 0 → cancel
              → 撤销 pending ✓
```

### 场景 3：CD2B 盲区返回（frozen 状态，PR-C 不应误抑制）

```
00:00     人入盲区（真消失）       ts.LastObservedMs 持续刷到当前帧
00:00~5.5min  firmware frozen 残影  ts.FrozenRunStart > 0
              整个期间 NO number_people=0（frozen ↔ number_people=0 互斥）

00:05:30  人重新出现 → 新 track 出生 → cancelPendingByBirth

[或者：人没回来（盲区跌倒）]
00:05:30  firmware 给 up → track=88 + number_people=0 同时到达
          ts.LastObservedMs ≈ 5:30 (last frozen frame)
          ts.FrozenRunStart > 0 ★ 关键

00:05:40  入池前检查：
          ① FrozenRunStart > 0 ✗  ★ frozen guard 命中
          → 不 skip，正常入 pending（带 FrozenStartMs > 0）

00:05:40  之后 number_people=0 事件如果再来：
          p.FrozenStartMs > 0 → 保留 pending ★

00:10:30  effective_wait 到 → lost_fall fire ✓ 漏报问题就此堵上
```

### 场景 4：远古 number_people=0（防过度抑制）

```
00:00:00  上一个人离开了，发了 number_people=0
00:00:05  新人进入屋子 → 新 track
00:00:05~02:00:00  新人活动 2 小时
02:00:00  新人在盲区跌倒，track 消失（无 ExitRoom，无 number_people=0）
          ↓
          入池前检查：
          ① FrozenRunStart == 0 ✓
          ② lastNumberPeopleZeroMs = 00:00:00 ✓
          ③ gap = |00:00:00 − 02:00:00| = 2h ≫ 60s ✗
          → 不 skip，正常入 pending → 5min 后 lost_fall fire ✓
```

---

## 一句话总结

**有 number_people=0 + 无 ExitRoom + 非 frozen + 60s 窗口内 → 视作正常离场（skip 或 cancel pending）；任一条件不满足 → 走原有 lost_fall 流程**。

PR-C 用最少代码改动（10 个文件改动总数）解决了 D523 误报问题，同时保护了 CD2B 类盲区返回场景不被误抑制。frozen + 噪声那种理论 case 没数据先不做，等 prod 真出现再加。

---

## 实测样本（PR-C 设计依据）

三例独立样本，验证 number_people=0 与 track_id=88 的相对时序：

| 离场点 | number_people=0 | track_id=88 | 间隔 |
|---|---|---|---|
| D523 @ 00:22:00 | 00:22:01.538 | 00:22:01.578 | +40ms |
| D523 @ 00:35:48 | 00:35:48.870 | 00:35:48.906 | +36ms |
| Kitchen @ 00:36:54 | 00:36:54.037 | 00:36:54.081 | +44ms |

**number_people=0 一律领先 track_id=88 36-44ms**。选 number_people 而非 track=88 的原因：
1. 早 40ms（抢跑 lost_fall pending 创建逻辑）
2. 同在 `iot:event:stream`（engine 已订阅，无需新增 monitor 流处理）
3. 语义干净（"屋内人数→0"），track=88 只是 firmware 心跳产物

## 测试覆盖

`track_manager_silent_fall_test.go` 末尾 6 个测试：

| 测试 | 场景 | FrozenRunStart | 60s 窗口内？ | 结果 |
|---|---|---|---|---|
| `SkipsPendingCreation` | 标准离场 | 0 | 是 | skip pending ✓ |
| `CancelsExisting` | 时序反转 | 0 | (pending 已建)| cancel pending ✓ |
| `OutsideWindow` | 远古 number_people=0 | 0 | 否（>60s）| 正常入池 ✓ |
| `FrozenOverridesNumberPeopleZeroSkip` | CD2B 盲区返回（入池前）| > 0 | 是 | 强制入池（frozen guard）✓ |
| `FrozenPendingNotCancelledByNumberPeopleZero` | CD2B 盲区返回（入池后）| (pending 已建 frozen)| - | frozen pending 保留 ✓ |
| `ParseRadarTrackEvents_NumberPeople` | 解析层 | - | - | envelope 字段对齐 ✓ |
