# 工单 — split-coupled 冻结残迹 floor FP（d5f7 0702）

## 现象
- 设备 E598A2ACD5F7 / Denver-101，07-02 报 `dbn.Fall` band=floor。
- incident(stillbox 起) 18:37:33 → fire 18:57:43（coast ~20min 满 tfloor）。
- 真值：**误报**。房里只有 1 个真人，摔倒未发生。

## 根因（已用长窗 replay 实证）
时间线：
```
18:37:15  number_people=2      firmware split（真人 + 反射）
18:37:15  tid1(D5F713715552) 出生 area255，z 起伏 42~87（真人身高）
18:38:04  firmware 丢 tid0，np 2→1；引擎 coast tid0 残迹
18:45:48  tid1 ExitRoom（真人跨门）+ np→0
18:57:43  tid0 残迹 coast 满 → 生产 floor 补火
```
- **tid0 (D5F703543822)** = 带 EnterRoom 出生证、z≡0 的贴地反射，被固件冻在 (-170,270)/(−150,220) 后丢弃，引擎独自 coast。
- **tid1** 才是真人（有 z 身高、18:45:48 触发 ExitRoom——只有真人能跨门）。
- census pR 因 **EnterBorn 钳制**把 tid0 判 real(1.0)、tid1 判 ghost(0.11)——反了。但 **floor 用 StillBoxSec 不用 pR**，pR 判反不直接致 fire。

## 关键代码事实
- `split_ghost.go`：vote→spliter(tid0)→group→step3 三档（≥200 WALKOUT real / ≤50 GLUED 3-tick 坐实 ghost / 中间 HOLD）。本 case **已 convict tid0**（`split_ghost_convicted` anchor_dist=41 → `SplitGhostSinceMs` 置位，单调）。
- `track_manager.go:1213-1214`：`SplitGhostSinceMs>0 → base.StillBoxSec=0`（**suppress**，留轨不删，可逆：walkout/倒地相解压）。coast 期 `updateSplitGhost` 不跑 → 标记不清 → **suppress 撑过整个 coast**。
- `exitCoupledLostResidual`（track_manager.go:2202）分支②：`SplitObservingSinceMs>0 && kinematicGlued(SplitSpliter)==GLUED` → **purge（delete）**。要求 `exitMs≠0 && |exitMs−LastObservedMs|≤30s`。

## replay 验证（含窗口订正）
⚠️ **窗口选取教训**：生产 alarm 的 fire lid = `D5F702620120`，按 makeLogicID 解码 = 出生 **18:26:20**。
第一次窗口 18:35 起 **漏掉该轨出生** → replay 冷启给它发了别的身份 `D5F703543822`（带 18:35:43 EnterRoom）
= **无效验证**（apples-to-oranges）。**fire lid 编码出生时刻，窗口必须早于它。**

订正窗口 **case-d5f7-0702-18241858（18:24→18:58，8x）**：
- `split_ghost_convicted tid=0 lid=D5F702620120 dist=41` —— **与生产 fire 完全同一 lid**。
- 18:57:43（生产 fire 时刻）：tid0 `present=False, still=0, pR=0.26, band=no` → **不 fire**。
- ⟹ 当前代码对**这条精确轨** split-convict + `line 1213` suppress（still=0）→ 压住。靠 **suppress 非 purge**（轨到 18:58 仍在）。
- 附注：生产 room=Bathroom、tfloor=1200(20min)；fire cell_area=9、坐标(-170,310) 与 replay 一致。

## ⚠️ 任务1 复查反转：不是部署差，是 split-conviction knife-edge
生产日志 `wisefido-sensor.log`（UTC=Denver+6）实锤：
- `00:37:15.552Z split_group_formed` d5f7 anchor=(-170,270) offspring=1 → **生产代码在场、形成了 split group**。
- **事发窗 00:37–00:58 UTC 内 d5f7 零 `split_ghost_convicted`**；fire lid D5F702620120 无 convicted 记录。
- 代码 diff：`split_ghost.go` 两树逐字相同、line1213 两侧都在；事发窗(15:00–18:40)无 roomengine 提交；事发进程 16:26 起(HEAD=0d5d3f6 含 suppress)。
⟹ **同代码同数据，生产 fire 是因 conviction 从没发生**（非缺 suppress 代码）。
**根因=knife-edge**：anchor(-170,270) 离 tid0 落点(-150~-160,220) 恰 51~54cm，卡在 `splitGluedCm=50` 边界。冷启 replay Kalman 探到 41cm(GLUED)→3-tick convict→suppress；生产暖态 Kalman 停 51~54cm(HOLD)→棘轮反复断→永不 convict→不 suppress→fire。
⟹ **line1213 convict-based suppress 不可靠**（冷/暖态翻转）→ 任务2 purge 兜底=真正的修；并建议硬化 conviction（见任务3）。

## 两种补救（勿混）
| | 做法 | 适用 | 可逆 |
|---|---|---|---|
| suppress | StillBoxSec=0，留轨 | **已 convict** 的 split ghost（line 1213）| 可逆 |
| purge | delete 删轨 | exit-coupled 残迹 / **未 convict** 的 split 成员 | 不可逆（ExitRoom+np→0 撑腰）|

---

## 任务 1 — ✅ 已定位(反转)：非部署差，是 knife-edge conviction
见上"复查反转"。结论：代码在场，生产形成 split group 但 knife-edge 未 convict → 不 suppress → fire。
⟹ 无部署动作；conviction-based suppress 不可靠 → 转任务 2(purge 兜底)+任务 3(硬化 conviction)。

## 任务 2 — ✅ 已实现+验证：prong 3 房空 purge split 残迹
覆盖空洞：split 成员未 convict（knife-edge）或 exit 前久驻 lost → line1213/exitCoupled 都兜不到 → 裸 coast → fire。**只能 purge**。

**实现**（两树 1:1，track_manager.go）：
- 新函数 `splitResidualRoomEmpty(ts,nowMs)`：`SplitObservingSinceMs>0 && !SplitEverWalkedOut && d.np==0 && d.npTs>0 && nowMs-d.npTs≥splitResidualEmptyGraceMs(15s) && poseEvictable && last2Dz≤exitLostMaxDz` → true。
  - **不卡 GLUED**（残迹赖锚点 51~54cm=HOLD 也收）；**不依赖 exitMs 30s 耦合**（用 np→0 触发，治 7min-gap）。
- lost-loop 在 confirmedMirrorResidual 之后接入：`presenceCoast + splitResidualRoomEmpty → delete + log split_residual_room_empty_purge`。
- 新常量 `splitResidualEmptyGraceMs=15_000`。
- 未走 exitCoupledLostResidual 分支②（其 exitMs 30s 耦合结构不适合），另立独立函数。
- FN 安全：SplitEverWalkedOut 绝不删 / np→0 硬确认(防"2 真人 1 出门 1 摔") / pose·dz 硬闸。

**验证**（case-d5f7-0702-18241858，8x，两树 build OK）：
- `split_residual_room_empty_purge` 触发，lid=`D5F702620120`（生产 fire 精确身份）。
- tid0 于 18:45:51（np→0 18:45:49 后）从 target 消失（无 prong3 时赖到 18:58:47）；fire=0。
- ⟹ 生产"未 convict"情形下，prong3 会在 ~18:46 删 tid0 → 18:57:43 fire 从根消失。
- ⚠️ 待补 FN 验证：需一个"split group 内真人摔倒"case 确认 prong3 不误删（pose/dz+np→0 闸应挡住，未实测）。

## 任务 3 — 硬化 split conviction（治 knife-edge，可选但推荐）
问题：anchor 距残迹落点 ~50cm 时，冷/暖 Kalman 翻转 GLUED/HOLD → convict 与否随机 → suppress 不可靠。
候选（择一，需 replay 验不引 FN）：
- [ ] GLUED 阈从硬 50cm 改**滞回**（进 convict 用 ≤50，维持用 ≤70）——防边界 flap；或
- [ ] HOLD 也喂**慢棘轮**（HOLD 累计 N tick 无 walkout 也软坐实），使"赖在 51~54cm 不走"最终也被判 ghost；或
- [ ] convict 距离基准用**冻结坐标滑窗中位数**而非当前 Kalman 帧，去除冷/暖态抖动。
- 注：任务 3 治"suppress 该发生却没发生"；任务 2 治"suppress 没发生时的 floor 兜底"。两者叠加 = 双保险，任务 2 优先(FN-safe purge 立即止血)。

## 边界（本工单不覆盖）
- 非 split 的 WALKOUT-frozen 冻结残迹（离 birth≥200 被判永久 real + 无 ExitRoom）仍是独立盲区，另立工单。
- census pR 的 EnterBorn 钳制判反（tid0=real）是 realness 轴问题，只影响 lost-fall/realness 路径，不影响本 floor 路；如需治 realness 轴另议（split group 解 EnterBorn 钳 + 对称 ghost 积分 + z 身高证据）。
