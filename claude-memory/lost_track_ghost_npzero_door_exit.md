---
name: lost_track_ghost_npzero_door_exit
description: 2026-06-01 bathroom lost_track 误报两段修复（ghost-last 抑制 + np=0×门区合取推断 exit）+ np=0 是确认证据非替代证据铁律 + belief 转移矩阵已对/只需调 np=0 likelihood
metadata: 
  node_type: memory
  type: project
  originSessionId: c178e90b-a72d-4034-ad46-9d5d1b452040
---

2026-06-01 修 D5F7(浴室镜面雷达, `fd00:0:3:111:3:300:a2ac:d5f7`, card `/80`) lost_track Fall 误报。根因实证：3 天 number_people 上升/下降沿各 19 次，但 EnterRoom 只 3 次 / ExitRoom 只 1 次 = **enter/out 事件大面积失效**；今天 07:27:51 np 已翻 0 但无 ExitRoom → 07:29:56 误报。

**坐标判据修正（别再用 V 越界判 ghost）**：FP track 在 X∈[-190,-120]（镜面 ghost 区）V=140-210；但用户专门在 101 staged 的**真跌倒** lost-1031 case V 也到 190、X∈[-120,-40]。boundary frontV/rearV 只是 UI 参考非物理 FOV，真人趴地 raw V 本就大。**真正判据是 ghost adjudicator 的学习型 ghost-zone/birth 分析，不是几何越界。**

**两段修复（均在 [bathroom_fall.go](../../../owl/owlBack/wisefido-sensor/internal/roomengine/bathroom_fall.go) `evaluateLostFallStrong`，已部署）：**
1. 方法2 = `lastBasesAllGhost`：失锁前最后观测全 `VerdictGhost` → 不 fire（把已判准的 ghost verdict 接进 lost-strong 决策路径）。
2. 门区 exit 推断 = `inferredDoorExit`：失锁前最后观测全「门区(AreaEnter) 或 ghost」**且**固件近期发过 number_people=0 → 推断真人正常走出 → 不 fire。两证据缺一不可。
   - np=0 时间由 [track_manager.go](../../../owl/owlBack/wisefido-sensor/internal/roomengine/track_manager.go) `LastNumberPeopleZeroMs` 经 engine `SetBathroomFallRules` lookup 注入；`EventNameNumberPeople` 常量化。

**铁律（用户两轮强调）：np=0 是确认证据(corroboration)不是替代证据(substitution)。** count=0 ≠ 人离开（金属垃圾桶/镜子→ghost、淋浴水气→信号衰减都会假报 np=0）。代码早有此规：zone_rules.yaml `leave_evidence 不配 number_people_0`；pendingLostFall 只 ExitRoom/多人入屋/birth-recovery 三路取消；旧 `NumberPeopleZeroFallbackMs` 是死配置。安全切分：镜面反射=out-of-bounds ghost(方法2 抑制)；水气衰减=in-bounds real/pending(不抑制，倒地照报)；frozen 与 np=0 在固件状态机互斥。

**转移矩阵问题（用户问"是否也要调"）：[belief](../../../owl/owlBack/wisefido-sensor/internal/roomengine/belief/) shadow 状态机的转移阵 [model.go](../../../owl/owlBack/wisefido-sensor/internal/roomengine/belief/model.go) 基本不用动**——→Empty 列已把不对称编码对：SStandWalk/SSit/SBed*/SFallen → Empty 全 0（占用态不能直接 teleport 到空），离开只经 SLeft(→E 85)/STransition(→E 6)/SArtifact(→E 30)；SFallen 近吸收(self 99, →E 0, 只有"返回/走动"主动反证能压)。**真正要调的是 likelihood**：[likelihood.go](../../../owl/owlBack/wisefido-sensor/internal/roomengine/belief/likelihood.go) `ObsNumberPeople`<0.5 现标"强证据空房"(SEmpty:6 + SFallen:0.3 强压)= 过度信任，应降为弱(SEmpty≈1.5) + SFallen 中性≈1（np=0 不得反驳已倒地）；强 Empty 拉力交给 ObsEnterExit<0(SLeft:8)。这样"np=0+门区"合取在贝叶斯层自动涌现，无需硬 if。**likelihood 重标定尚未实施（shadow-only，待确认）。**

**belief 状态机用 101 真实记录测试（TestReplayOracle 新增 2 case + d5f7-bathroom-fp-0601 fixture）结论**：
1. 裸 replay（无 ghost adjudicator，Verdict 恒 Real）→ belief 也 **confirm=true P=0.999** 误确认（与 gate 同病）；属保真局限，不断言（同 Hunzi-0530）。
2. 仅注入 ghost verdict 还不够——ghost obs 只在 track 在场时存在，track 消失后 lost-while-moving ramp 无对手地拉 Fallen。**必须再加「ghost 消失不发 lost-while-moving」guard**（= method-2 在 belief 侧同构）。
3. 加 guard 后 FP+ghost verdict → **confirm=false maxP=0.029** ✅，真跌倒仍 confirm=true，6 个对照 case 不变。
→ **结论：belief 与 gate 完全同构，治本 = ghost adjudicator verdict + 「ghost 消失≠倒地」两件套，缺一不可；np=0 对本 case 是 red herring，真防线是 ghost verdict。** 已把 guard 补进 production [belief_shadow.go](../../../owl/owlBack/wisefido-sensor/internal/roomengine/belief_shadow.go)（ghost 帧 `delete(sh.tracks)`，防 real→ghost→消失 误发；shadow log-only 已部署）。

转移矩阵/likelihood 重标定（np=0 降弱）**仍未做**——本 case 证明它不是关键路径，优先级降。

验证：7 lost-strong + 全 roomengine 套件（含 8 个 replay oracle case）绿，go vet 干净，sensor 已重启。抑制日志 `bathroom_lost_strong_suppressed_ghost` / `_door_exit`（带 last_cell_area/np_zero_ms）可盯几天核对。相关 [[fall_fp_roots_and_todo]] [[d5f7_ghost_cases]] [[belief_state_rule_engine_reframe]] [[feedback_no_dynamic_threshold_modulation]]。
