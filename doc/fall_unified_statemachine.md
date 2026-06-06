# Fall 检测统一重构 — 复核稿

目的:把当前散落的 8 条 fall 路径收敛成 **silent / lost / moving 三类 + 触发表**,
基于 60s tick 的 track 运动态状态机。本文 = 现状盘点 + 拟定设计,供 codex 复核。

---

## 1. 当前逻辑分支(代码位置)

### 1.1 编排 / dispatch
| 位置 | 作用 |
|---|---|
| `engine.go:949` `isBathroom := roomType == card.RoomTypeBathroom` | 房型分流 |
| `engine.go:988` `e.bathroomFall.Evaluate(roomID, bases, nowMs)` | bathroom 房入 BathroomFallRules |
| `engine.go:1001` `e.bedroomFall.Evaluate(roomID, bases, beds, nowMs)` | 非 bathroom 入 BedroomFallRules |
| `track_manager.go:1446` `checkLostFall(ts)`(ProcessFrame 内) | per-track 失锁 → arm pending lost-fall |
| `track_manager.go:1460` `pendingLostFalls[id] = ...` | arm |
| `track_manager.go:1610` 段 4b 扫 `pendingLostFalls` → fire | lost-fall 超时发射 |
| `track_manager.go:1758` `scanSilentFallLeftBed(nowMs)` | silent_fall(下床)发射 |
| `track_manager.go:1028` ExitRoom → cancel pending lost-fall | 取消 |

### 1.2 静止 / 失锁原语
| 机制 | 位置 | 说明 |
|---|---|---|
| StillSec(逐帧) | `scoreMovement` `track_manager.go:2717`,阈 `StillThreshCm=15`(`track.go:199`) | 逐帧位移<15cm,任一帧≥15cm 瞬清。**脆**。 |
| **still-box(滚动盒)** | `updateContinuousIndicators` `track_manager.go`,`BoxRangeWithinMs` `track.go`,`StillBoxCm=50`(`fall_rules_param.go:162`) | 30s 滚动窗 per-axis 50×50 方框。**已改用此为 StillBoxSec**。 |
| RegionStatic | `updateRegionStatic` `track_manager.go` | 15cm region + 90% 容忍 + 50cm 硬 reset(未接 fall) |
| lost-fall wait | `lostFallWaitMs` `track_manager.go:3028` | cell-area 5~60min + still-box credit |

### 1.3 八条 AI fall 路径(publish 统一 `alarm.Fall`,靠 `reason` 区分)
| # | 函数 | 触发信号(arming) | 阈值 | reason | 拟归类 |
|---|---|---|---|---|---|
| 1 | `scanSilentFallLeftBed`(track_manager.go:1868) | sleepad InBed≥5min + LeftBed(事件) | 等 60/120s + radar≤100cm 邻域 | `sleepad_radar_conflict` | **silent**(下床) |
| 2 | `checkLostFall`(2983)+`lostFallWaitMs`(3028)段4b | track 消失(失锁) | cell-area 5~60min + still-box credit | `lost_track` | **lost / moving**(未分) |
| 3 | bathroom `evaluateStillFall`(bathroom_fall.go:287) | 浴室占用 + cell Toilet/Shower + pose{0,4} | StillBoxSec≥10/12min | `bathroom_still` | **silent**(浴室) |
| 4 | bathroom `evaluateBedsideFall`(328) | 浴室占用 + 90s grace + 非桶/淋浴 | StillBoxSec≥8min | `bathroom_long_static` | **silent**(浴室) |
| 5 | bathroom `evaluateLostFallStrong`(406) | 浴室活track==0≥30s + Count≥1 | 30s | `suite_person_completely_lost…` | **lost**(浴室) |
| 6 | bathroom `evaluateLostFallWeak`(480) | 浴室 SuitePerson static | StillBoxSec≥7min | `suite_person_silent_with_ghost…` | **silent**(浴室) |
| 7 | bedroom `evaluateBedsideFall`(bedroom_fall.go:232) | LeftBed latched + 床边≤100cm | StillBoxSec≥15min | `bedroom_bedside_static` | **silent**(下床) ←与#1 重叠 |
| 8 | bedroom `evaluateLostFall`(295) | bedroom SuitePerson idle | AreaBed 2h/Sit 30min/Active 5min | `bedroom_person_silent` | **silent**(久坐久躺) |

(firmware pose=5 直发 = Device_ALARM,非 AI 路径,不在本重构内。)

### 1.4 现状问题
- #1 与 #7 重叠(都"下床后床边静止"),一个 sleepad 驱动 + 100cm,一个 radar + 100cm + 15min,口径不一。
- #2(lost)把"silent后消失"和"moving中消失"混在一起,**缺 moving-fall 区分**。
- 6 条本质同为"静止 still-box≥阈",被拆 6 个函数 + 6 个 reason + 散落距离闸。
- 距离闸(#1/#7 的 ≤100cm)散落、口径不一,且会漏(135cm 真摔)。

---

## 2. 拟定替换设计

### 2.1 统一状态机(60s tick,per track)
按 track 运动态三分支:
```
STILL-BOX (track在, 30s滚动50×50盒内不动)
   ├ still_duration += 60
   └ 按【触发表】判 still-fall:  continue / fire

LOSING (本tick track消失)
   ├ prev=STILL  且 still_duration≥阈 → pending LOST-fall    (silent→消失)
   └ prev=MOVING                      → pending MOVING-fall  (走着→消失)

MOVING (track在, 连贯合理速度位移)
   ├ 多人 / ghost 校正
   └ 出现连贯活动 = 人活着 → CANCEL pending(still/lost/moving)  (恢复兜底)
```
- lost vs moving 由"消失前 prev state"区分(新增 moving-fall)。
- cancel 用"出现连贯活动 track"替代散落的距离闸 / ExitRoom 兜底。

### 2.2 三类定义
| 类 | 定义 | 现路径来源 |
|---|---|---|
| **silent-fall** | 人静止不动(track在) | 1,3,4,6,7,8 合并 |
| **lost-fall** | silent 后 track 消失 | 2(prev=still),5 |
| **moving-fall** | moving 中突然消失 | 2(prev=moving)拆出 |

### 2.3 silent-fall 触发表(取代 6 个独立函数)
| case | arming 信号 | 类型 | still_duration 阈值 |
|---|---|---|---|
| **leftBed(下床)** | InBed≥5min + LeftBed | 事件型 | 短(~1-3min,起身高危窗) |
| bathroom | 浴室占用 + 90s grace | 占用型 | 桶/淋浴 10-15min / 任意 8min |
| stay-alarm | 运维显式开启 | 占用型 | 配置 |
| default(卧室/客厅 normal) | — | — | 不报(正常久坐久站) |
- 同一内核"判 still_duration ≥ 阈",case 只配 arming + 阈值。
- **去掉距离闸**:触发已隐含位置上下文(下床=床边,占用=该房);定位错(质心甩偏)由 cancel 兜底,不卡距离。

### 2.4 still 判据(已落地基础)
- still = **30s 滚动 50×50 per-axis 方框**(`BoxRangeWithinMs ≤ StillBoxCm=50`);
- 理由:人体宽 ~40cm,质心在体内游走;<体宽 的阈值(旧 30/对角线)把真摔躺着误判成"动"。
- fall 判据已从逐帧 `StillSec` 切到 `StillBoxSec`。

### 2.5 输出矩阵(每 tick 状态向量)
| 维度 | 取值 |
|---|---|
| Track 态 | move / stand / sit / lying / inBed / **fall** |
| Bed | 无人 / 有人 |
| Room | 0人 / 1人 / 多人 |

决策矩阵(运动态 × 上下文 → 输出):
| 运动态＼上下文 | bed有人 | leftBed后/床边 | 浴室占用 | room normal |
|---|---|---|---|---|
| STILL 久 | InBed/lying(正常) | **silent-fall(短阈)** | **silent-fall(8-15min)** | 正常不报 |
| STILL→LOSING | — | lost-fall | lost-fall | lost-fall |
| MOVING→LOSING | — | moving-fall | moving-fall | moving-fall |
| MOVING | move(cancel) | cancel | cancel | move(cancel) |

---

## 3. 待复核要点(请 codex 重点看)
1. **prev-state 记忆**:LOSING 分支要可靠区分 prev=STILL vs prev=MOVING —— 现有 `History`/`StillBoxRunStart` 是否够支撑?是否需要显式 prevMotion 字段。
2. **cancel 充分性**:用"出现连贯活动 track"取代 100cm 距离闸 + ExitRoom 兜底,是否覆盖现有所有 cancel 场景(多人入室 / 同房他雷达 / 门区 exit / np=0)。
3. **silent default 不报**:卧室/客厅正常久坐久站不报,靠 case 缺省 —— 会不会漏"非下床、非浴室"的真摔(如客厅站着晕倒)?是否要 stay-alarm 兜这类。
4. **bathroom 占用 vs 事件**:占用型 arming(90s grace)与事件型(LeftBed)统一进触发表后,grace/dedup 语义是否保留正确。
5. **#2 lost 拆分**:现 lost-fall 的 cell-area wait + still-box credit + 距离闸 + 同房对账,拆成 lost/moving 后这些守卫如何归位。
6. **firmware 直发**:pose=5 Device_ALARM 与本 AI 三类的去重/优先级。
