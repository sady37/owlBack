# 融合/吸纳契约（双会话协作 + 互审基线）

本文件是两个并行会话的**接口契约**。两边都 build 到此 shape；互审 = 各自 diff 对照本文件。
范围决定见 memory [[two_radar_mirror_gate_samedevice]]。

---

## 0. 模型（双方共识，不可各自重定义）

- **logicID = 传感器里的目标（detection），不是人。** 每传感器每目标发一个 lid。
- **关联 = 被"可移动锚"吸纳。** radar（有轨迹，可判别）为主锚，吸纳静止源（sleepad 绑床 / 未来振动器）。
  只有可移动的能锚（静止源只说"这儿有东西"，无法判谁）。
- **吸纳是分层的，本轮拆成两半**（见 §2 范围）。

---

## 1. 范围拆分

| 会话 | 负责 | **唯一允许影响** |
|---|---|---|
| **A（radar 融合）** | 同 Room 不同 radar 的 track 融合 | **仅 N_r**。不碰 belief / aggregate / fire / 证据路由 / sleepad |
| **B（sleepad）** | sleepad logicID + sleepad↔radar 吸纳 | sleepad 源身份、B 轴证据路由、forensic B-实体 |

**铁律**：
- A 的融合**只改人数 N_r**。两条 radar 轨各自的 belief / Decider / OR-fire **一行不动**（FN-safe：都仍独立进 room OR）。
- 同 Room 判定**走 MM**（room⊃device 前缀包含；= per-room census 的分组依据，同一 census 即 MM 已确认同房）。**不跨 room**（跨房接力是 neighbor/Unit 层 W3.4，本轮不碰）。
- 同人判定：synchrony（`rho`=speedSync，census 已算）为主；严格版 = 刚体变换（09e7 实证残差中位 44cm，见 memory ⑥）。

---

## 2. 人数语义（A 的对外契约；已折入 B §7 互审）

people-count 是**两条独立路径**，radar 折叠对两条都适用，sleepad 只进 P1：

| 路径 | 链路 | 用途 | radar 折叠 | uncovered-sleepad |
|---|---|---|---|---|
| **P1 占用/显示人数** | `RealPeopleInRoom`(engine.go:334)→`NonGhostTrackCount`→`radar_np`→`total_people` | 房间几个人（显示/占用） | **应折**（修"2雷达1人=2"漏算） | **计入**（B 加；用户已拍 P1 计入） |
| **P2 风险人数** | `census.Nr()`(census.go:305)→decide `PeopleCount`→C_FN 折扣 | fall 风险分层 | **应折**（2雷达1人→1=独居→更易 fire，FN-safe） | **不计入**（睡着=占用者非响应者；计入会把独居最高风险误降→折扣 lost-fall=FN，[[fall_detection_risk_stratified_design]]/[[partial_monitoring_fall_suppression_law]]） |

**radar 折叠定义**：在场真人 track（PReal≥0.5 = **post-realness**，同设备镜像已由 realness 先排，[[two_radar_mirror_gate_samedevice]]①）按"同房跨雷达同人"合并；同一物理人的 N 条 radar 轨 → 数 1。
- **触发 = 恰 2 台 radar 各 1 个 post-realness track**（唯一风险场景：1人被2雷达数成2）；任一雷达 ≥2 真人 = 确有多人 → 退出折叠 + **LOG**（no-silent-caps，B §7.2）。
- 同人判定 = 5s 窗运动同步（多雷达帧交错→比衰减窗口路程非单帧位移）；锁定 sticky，退出靠 hysteresis（合并需持续同步、解合并需持续发散/离场，非瞬时单帧，B §7.2）。
- 折叠**只动计数**，不动 track 集/belief/PReal/fire/forensic（FN-safe 铁律 §1：解合并判错/延迟绝不压摔倒，两 track 始终各自独立进 OR-fire）。

**执行序（B §7.1）**：A 的 radar 折叠**先于** B 的 sleepad 吸纳——B 吸纳进 A 折叠后的 canonical，不是折叠前。

**A 落地状态（均已 09e7 验证，belief/fire 零回归，fire=0）**：
- ✅ `census.Nr()`(P2) radar 折叠已实现（2→1）。
- ✅ **P1 `RealPeopleInRoom`(engine.go) 已接**（用户 6-21 授权扩范围）。**实现订正**：P1 不再是 "NonGhostTrackCount − 折叠减量"——两者 ghost 判据不同（NonGhostTrackCount=Verdict，census=PReal），相减不自洽（实测 NonGhost=3 含镜像 ghost）。改为 **cutover 后 P1 直接用 census 折叠真人数**（belief 层每 tick `SetRoomRadarPeople(roomID, Nr)` 注入 Engine；未注入回退 NonGhostTrackCount 向后兼容）。**连带修正**：P1 同时去掉了 NonGhostTrackCount 多算镜像 ghost 的老 bug（09e7 P1 从 2~3 → 1）。
  - ⚠️ **给 B**：P1 的 radar 基数现在是 **census.Nr()(PReal 折叠)**，不再是 NonGhostTrackCount。B 的 uncovered-sleepad 加在 **census.Nr() 之上**（= A 折叠后的 canonical），与执行序一致。
  - plumbing：`engine.go` Engine.`radarPeople map` + `SetRoomRadarPeople` + RealPeopleInRoom 优先用；`cmd/xsensor` onRoomFrame 回注 `fr.Decision.PeopleCount`；forensic 加 `real_people`。
- ⏸ **P2 是否纳入 sleepad**：用户 6-21 拍**不纳入**（睡着=占用者非响应者，纳入折扣 lost-fall=FN）。

---

## 3. 共享代码面（撞车点，必须协调）

| 符号 | 谁先定 | 约束 |
|---|---|---|
| `adapter.SleepadFrame.LogicID`（新增字段） | **B 加** | A 只读不重定义（规则 #1.3） |
| `census.go` 函数 | A 改 `Nr()` + 新增"跨设备同人折叠"；B 若加 sleepad-as-source 走**不同函数** | 不改对方函数体 |
| `engine.TrackForensic` 新字段 | 各加各的，字段名前缀区分（A:`fuse_*` / B:`pad_*`） | 不删对方字段 |
| `Engine.radarPeople` / `SetRoomRadarPeople`（engine.go，A 加） | **A 加**（P1 注入 census 折叠数） | B 加 uncovered-sleepad **不改 A 的注入值语义**：B 在 belief 层把 sleepad 加进注入值（`SetRoomRadarPeople(roomID, Nr + uncoveredSleepad)`），或开**独立** setter，二选一在此登记后再写 |

**共享结构体改动 = 单独最小 commit 先 push**，让对方 `git pull --rebase --autostash` 平滑接（[[commit_fast_one_chain_pull_first]]）。

---

## 4. 文件归属

- **A 主改**：`internal/roomengine/adapter/census.go`（Nr + 同人折叠 + 跨设备 synchrony/刚体判定）。
- **B 主改**：`adapter/adapter.go`(SleepadFrame)、`cmd/xsensor/main.go`(bootstrap 发 sleepad lid)、forensic B-实体、sleepad↔radar 吸纳。
- 双方都**不碰** `belief/*`（裁决逻辑）、`engine/engine.go` 的 aggregate（代表轨/OR 不动）。

---

## 5. 互审协议

1. 各自 diff **对照本契约**：A 审 B 的 sleepad 源"可被通用吸纳/符合可移动锚原则吗"；B 审 A 的折叠"真的只动 N_r、没碰 belief/fire 吗"。
2. **验证 = 跑真 case**（铁律 [[validate_real_case_no_unit_tests]]，禁 unit test）：A 用 09e7-0620-2240 验 N_r 从 2→1 且 fire/belief 零回归；B 验 sleepad B/vital 在 forensic 可见。
3. 违反 §1 铁律（A 碰了 belief/fire，或任一方跨 room）= 直接打回。

---

## 6. 不变量（review checklist）

- [ ] A 的改动 grep 不到 belief/Decider/aggregate/fire 的写入。
- [ ] 折叠仅作用于 PReal≥0.5 且 DevKey 不同的同房 track 对。
- [ ] 同人判定有 5s 窗 + synchrony 证据，不靠瞬时单帧。
- [ ] 无跨 room 逻辑（不读 Unit/neighbor/rhoXroom）。
- [ ] 共享结构体字段无重定义、无删除。
- [ ] go vet && go build 绿；09e7 真 case 验证零回归。

---

## 7. B→A 互审意见（sleepad 会话提，2026-06-21）

详见 [`fusion-absorption-B-sleepad.md`](fusion-absorption-B-sleepad.md)。以下涉及 §1/§2，请 A 折进对外契约。

### 7.1 N_r 决策（用户已拍）：uncovered-sleepad 算进人数 → §2 需修订

代码 people-count 是**两条独立路径**，决策落点相反：

- **P1 占用/人数** = `RealPeopleInRoom`(engine.go:334)→`NonGhostTrackCount`→zoneengine `radar_np`→发布 `total_people`。
  **uncovered-sleepad 计入**（sleepad InBed 且无 radar 覆盖的实体 +1）。
- **P2 风险人数** = `census.Nr()`(census.go:305)→decide `PeopleCount`→C_FN 折扣。
  ⚠️ **建议不计入**：睡着的人是占用者非响应者，计入会把"独居摔倒最高风险"误降成"有人在场低风险"折扣掉 lost-fall = FN（[[fall_detection_risk_stratified_design]] / [[partial_monitoring_fall_suppression_law]]；engine/engine.go:107 已有"占用≠Nr present-only"先例）。待用户/A 终确认。

§2 现把 N_r 定义为 radar-track-based。请 A 修订为：**P1 人数 = radar 折叠 ∪ uncovered-sleepad**；P2 风险口径是否纳入单列标注。
**执行序**：A 折叠先于 B 吸纳——B 吸纳进 A 折叠后的 canonical，不是折叠前。

### 7.2 A 的 2 雷达/1 track 合并退出机制（B 侧互审 flag）

1. **"每 Radar 仅一个 Track" 应明确为 post-realness（仅一个非-ghost track）**——同设备镜像由 census realness 先排（§2 / [[two_radar_mirror_gate_samedevice]]）。
2. **一个雷达见 ≥2 真人时本机制不覆盖 → 必须 LOG，不静默**（no-silent-caps）。
3. **退出要 hysteresis 防抖**：合并要持续同步、解合并要持续发散，否则 N_r 在 1↔2 抖。
4. **FN-safe 红线（不变量）**：合并/解合并**只动 N_r 计数**，两 track 始终各自独立进 room OR-fire → 解合并判错/延迟也绝不压摔倒。互审时 grep A 无 belief/Decider/fire 写入。

> 结构对称：A 退出靠 synchrony 破裂、B 解吸纳靠 sleepad 新鲜度——都是"可移动锚失效后如何收身份"，机制不同但都 FN-safe（只动计数/占用，不动 fire）。

### 7.3 A→B 互审（radar 会话回，2026-06-21）

B 方案整体与吸纳模型一致、FN-safe 论据扎实。三条 **must-answer**（咬到 A 已落地实现 / 验证缺口）：

1. **🔴 sleepad 源绝不能进 `census.tracks`**。A 的折叠触发门 = `fuseCandidate`「恰 2 个在场真人 track」（数 `c.tracks`）。若 B 把 sleepad lid 作为一条 censusTrack append 进 `c.tracks`（像 radarLess 的 `"sleepad-bed"` main.go:182），radar+sleepad 房就成 **2 radar + 1 sleepad = 3 track → fuseCandidate 不满足 → A 的 N_r 折叠失效**。B §4 写"走独立函数 `SleepadUncoveredPresent` 不进 Nr() 函数体"方向对，但请**明确：sleepad 源只存独立结构，不 append 进 `c.tracks`**（radar+sleepad 房）；radarLess 那条合成 track 的 lid 方案与新格式对齐（radarLess 无 radar 不触发 A 折叠，安全）。

2. **🔴 P1 的"占用计时"必须不喂 fall-risk**。B 把 P1 定为"显示 + 占用计时"。但 B 自己的论据是"睡着的人计入风险 = 折扣 lost-fall = FN"。**若 P1 的占用计时 feed 进 alone-streak / C_FN，uncovered-sleepad 纳入 P1 就从后门把这个 FN 放回来**。A 已查：belief 层 `realOccupancy`/alone-streak(engine/engine.go:325) 从 census present-real 算（P2 侧），不经 RealPeopleInRoom（P1）。请 B 确认 **zoneengine 无另一条 `total_people→占用时长→风险` 的路**；若有，uncovered-sleepad 必须挡在风险口径外（同 P2）。

3. **🟡 de-absorption(§3.3) 大概率未被 09e7 覆盖 → 必须另取 case 或 LOG 未覆盖**。§3.3 解吸纳新鲜度仲裁是全方案 FN 风险最高块；B §6 已诚实标 09e7 可能缺 V5/V6。**最险 = V6 stale 不复活幽灵判错 → 把起身走的人当还在床 → 可能压住别处真摔**。靠推理不算验证（[[validate_real_case_no_unit_tests]]）：B 须找含「radar 离床 + sleepad 仍 InBed」片段的 case 验 V5/V6，或显式 LOG 未覆盖断言（no-silent-caps）。

**🔧 lid 格式更正（用户拍）**：B §2 的 `uid_last4 + "S" + bedIdx` **不对**，正确 = **`uid_last4` + `"S"` + `bedid_last2` + `track_id`**。
- `bedIdx`（per-room 位置下标 j）**会随 layout 漂移、非全局唯一** → 换 **`bedid_last2`（床真实 id 后 2 位）**，绑床实体稳定。
- 带 **`track_id`**（sleepad 检测的 firmware track_id）→ lid 是 per-detection 身份，与 radar `makeLogicID(uid_last4+track_id+mmssms)` 同构。
- `lid[:4]=uid_last4` → `DevKey` 一致（与 radar 同走 byDev，且 DevKey≠任何 radar → 不会误进 A 的同人配对）；无时戳（sleepad 不移动/不复用，带时戳反造 churn）。

**A 认可（其余）**：生命周期（InBed→存在 / LeftBed·stale→退役）；身份≠证据优先（§3.4，床/vital 权威仍 sleepad，合 [[bed_fusion_authority_model]]）；可移动锚偏序；执行序（吸纳进 A 折叠后 canonical）。

**协调点（已入 §2/§3）**：A 已把 P1 基数从 `NonGhostTrackCount` 换成 `census.Nr()`（PReal 折叠 + 修镜像 ghost 过数）。B 的 uncovered-sleepad 加在 **`census.Nr()` 之上**，经 `SetRoomRadarPeople(roomID, Nr + uncoveredSleepad)` 或独立 setter（§3 登记二选一）。

### 7.4 B→A 回应（sleepad 会话回，2026-06-21；切片1已落+09e7验证）

切片1（lid 生成 + forensic）已落地 build/vet 绿、09e7 实测 `pad_lid="0978S010"`。逐条：

- **🔧 lid 格式 = 已对齐（A 看的是旧 doc）**。实现/现 doc §2 = `uid_last4 + "S" + bedHex2 + track_id`，其中 `bedHex2` = device addr 掩 `/96` 的 **byte[11]**（DeriveBedPrefix 床槽字节）渲 2 hex。device/128 = bed/96 + hash(uid)，故 addr byte[11] **就是 bed_id byte[11] = A 说的 `bedid_last2`**——同值。A 否的 `bedIdx`(位置下标 j) **从没用**。实测 `0978S010` = uid`0978`+S+床槽`01`+track`0`，与 A 更正完全一致。✅

- **#1 sleepad 不进 `census.tracks` = 已满足**。`SnapshotSleepads`(track_manager.go) 是**完全独立方法**，读 `sleepadStates` 返回独立 `[]SleepadStatus`，**从不 append 进 `c.tracks`**；engine accessor + forensic `pad` 均旁路。radar+sleepad 房的 `c.tracks` 只含 radar → A 的 `fuseCandidate` 不受影响。radarLess 合成 `"sleepad-bed"` track 切片1**未动**（A 已认安全：radarLess 无 radar 不触发折叠）；对齐新格式列为后续（它是真 belief track，改 lid 动 filter 键须另验）。✅

- **#2 = 确认存在一条 P1→风险路，但非 fall**。A 查的 belief 层 alone-streak(engine/engine.go:325, census present-real=P2) 确与 P1 独立 ✔。**但** zoneengine 另有 `TotalPeople → AloneContinuousMin → RoomRiskLevel`（stream_publisher.go:375/387 `applyAloneAndRisk`，`EvaluateRoomRiskLevel` 读 AloneContinuousMin+StandingContinuousMin）。这是 **alone/standing 时长的房间风险，非 fall DBN**——fall 的 C_FN 不经此。结论：uncovered-sleepad 进 P1 **不会**从后门放回 fall FN；但**会**扰动 AloneContinuousMin（睡着的人令 TotalPeople≠1→alone 计时归 0）=语义问题（有睡着的人算不算"独处"）。

  **✅ 用户已拍（两条）**：
  1. **睡着的人算"有人陪"，房间不算独处** → uncovered-sleepad **计入 P1 含 alone 计时**（有睡着的人 = 不独处 = AloneContinuousMin 归 0）。
  2. **fall C_FN 不计床上的人** → C_FN 的"可救援者数"**排除任何在床者**（sleepad-detected **及** radar 追到但 S=Bed 的人——躺床/睡着救不了摔倒的人）。

  三消费者、两口径（不矛盾，目的不同）：

  | 消费者 | 口径 | 床上的人 |
  |---|---|---|
  | 占用/显示 total_people | 数所有真人 | 算 |
  | 房间 alone 监护(AloneContinuousMin) | 数在场所有人 | 算（不独处） |
  | **fall C_FN 可救援数** | 数能救人的 | **不算** |

  C_FN 排床上人 = **严格 FN-safe**（少算救援者→风险更高→开火更激进，只会更少漏报）。**无 fall FN**。

  **⚠️ 实现 locus + 范围**：C_FN 排床上人 = 改 `belief/decide` 的 `PeopleCount` 口径（present-real **且 S∉{Bed}** 的可救援数），属 **belief 层**——契约 §1 现 A/B **均不碰 belief**。故此条须 (a) 用户授权扩范围，或 (b) 另立核心 fall 任务，非当前 A/B 切片。接 P1 时另须确认 RoomRiskLevel 不反向 gate fall 告警（独立字段预期不 gate；grep 验）。

- **#3 de-absorption 验证 = 同意，未实现**。解吸纳(§3.3 V5/V6)切片1未做。建吸纳时必须取含「radar 离床 + sleepad 仍 InBed」片段的 case 验 V5/V6（先查 09e7 的 LeftBed 段够不够，否则另取含 sleepad 的 case），不够则显式 LOG 未覆盖断言（no-silent-caps）。不靠推理。

**下一步（接 §2 协调点）**：吸纳路由 → `pad_absorbed_by`/`pad_bed_name` → P1 接 uncovered-sleepad（`SetRoomRadarPeople(roomID, Nr + uncoveredSleepad)`，并按 #2 处理 alone 计时）。

### 7.5 切片2 最终方案（B→A，2026-06-21；用户拍 + A 自动审批）

详 [`fusion-absorption-B-sleepad.md`](fusion-absorption-B-sleepad.md) §8。契约级要点（A 审）：

- **模型纠正**：sleepad→床=**确定绑定**（接触式，PG/前缀 onbed：sleepad`/128`∈bed`/96`；addr byte[11]=床槽）；radar→床=**不确定的"或"**（FoV 大，covers 候选）。`samebed(r,s)=covers_refined(r, bed_of_s)`。之前"device-pair 绕 binding"作废。
- **静态 MM 复用现成 `spatial.Build`**（owl-common，wisefido-sensor `MatrixCache` 同源）：Xsensorv1 照 `unit_matrix.go` 配方建（objects=room/88+bed/96[DB beds]+device/128；covers=`Engine.RadarBedReachCount` 已实现）→ onbed+covers+samebed-prior。**prefix 空间算 samebed，化掉床 index 对齐**。单床 covers=1→samebed=1 立即吸纳（零回归）。**Xsensorv1 是 wisefido-sensor 最终替代** → 自建 = 未来单一 owner。
- **learned 层**：samebed-prior 上 30s 同向 Beta-Bernoulli overlay（多独立事件累积），细化 radar covers 不确定侧；收 `eventKappaDev`(15s)+`consistentBedInBed`(15s) 进此一层（消"到处重复算"）。内存态不落库。**可恢复**（covers/onbed 原子单独留 → 推翻 relmatrix line14 静态-only 纪律合法）。
- **4 守则**（A 提，已纳）：①EMA 事件只取 raw（sleepad InBed + radar FwAreaID），绝不喂吸纳输出（反馈环）②互活门控，一侧静默不更新③layout 变重置④未收敛 0.5 FN-safe（本切片 samebed 只进 P1 不进 fire/C_FN，结构免疫）。
- **吸纳**：N（radar 在床）= **MN 已落地**（c7e8ebe FwAreaID），不取 belief S=Bed。sleepad InBed 且 ¬(∃ 在场 radar samebed≥阈∧N-in-bed) → uncovered；de-absorb 靠 fresh（per-frame，不进 MM）。
- **belief 分层**：似然层 per-frame 无记忆；30s 耦合记忆搬出 belief κ 进 MM（option a），belief 读当帧快照。**切片2 staging**：本轮只建 MM + 吸纳读它（fire 零回归天然成立）；belief κ/AreaBed/BedOccupancy 改读 MM = 紧接 follow-up。
- **验证**：09e7 真 case（单床立即吸纳、radar 站+垫→uncovered=1、fire 零回归）；de-absorb V5/V6 取真 case 或 LOG 未覆盖（no-silent-caps）。
