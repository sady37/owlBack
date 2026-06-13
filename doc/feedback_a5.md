# A5 反馈日志(委员会侧)— belief 全空间区域占用重写 + cd2b 床边真摔 FN

> **QA 分离(2026-06-12,因 git 同步写冲突)**:P5 主题拆成两文件各写各的——
> **本文件 `feedback_a5.md` = 委员会侧**(裁决/审查记录/工单);项目组侧提案/问题/根因写 `feedback_q5.md`。
> 两侧不写同一文件 → push 不再非-fast-forward 撞车。倒序,最新在上。

---

## 2026-06-13 — 委员会裁决:Q5 施工方案批准

### 裁决结论

**9 态设计批准。** 状态空间:

| idx | 态 | 裁决 |
|-----|-----|------|
| 0 | Empty | 通过。同旧 SEmpty。 |
| 1 | Bed | 通过。area=Bed + sleepad InBed。 |
| 2 | Sit | 通过。area=Sit,久坐正常,dwell tail=90min(不 ramp)。 |
| 3 | OpenFloor | 通过。area=Active/Deny,站/走,dwell tail=8min。 |
| 4 | Bath | 通过。area=Toilet/Shower,dwell tail=15min。 |
| 5 | Fallen | 通过。index 不变→Decider 不动。近吸收(自持 0.99)。 |
| 6 | BlindRest | 通过。从 rest 区(床/卫浴)进盲,无 Left 逃生阀。 |
| 7 | BlindOpen | 通过。从开阔区进盲,reachable-exit 裁 Left vs Fallen。 |
| 8 | Left | 通过。ExitRoom/门区 hand-off,→收敛 Empty。 |

**删 SArtifact/STransition/SBedRestless 批准。** ghost 全归 Track 层(Conf*=P(Real)退火),Room 层不设伪迹态。不确定性由 Blind+normalize 承载。

**Blind 入口条件化批准。** 盲区无几何(layout/sensor 都不画),态索引=消失前 last-seen 区(可观测),非"在盲哪"(不可观测)。Bed/Bath→BlindRest;OpenFloor→BlindOpen。

**unit 联动批准。** 三类机制:①同房多雷达 union(absence 取并集)②跨房 neighbor hand-off(偏分监控铁律)③sleepad↔radar 权威。scope 由 `owl-common/spatial/relmatrix.go` MM 矩阵界定。

**bed 融合权威(sleepad>radar)裁定为必要前置。** cd2b 被子把 Bed→Blind 转移遮住,只能靠 sleepad LeftBed 与 radar 床 track 矛盾推出。与状态空间重写正交,需并行修。施工顺序:bed 融合先修→cd2b fire 验证→以此为 oracle 建测试套→重写 belief→跑回归。

### 工单

- [ ] **工单 1**:bed scorer/ BedOccupancyState 修 sleepad>radar 权威(sleepad LeftBed 须盖过 radar 被子 InBed)。前置,不依赖 belief 重写。
- [ ] **工单 2**:belief 重写——state.go/model.go/likelihood.go 按批准 9 态重写;observation.go 加 zone 字段/删 ObsTrackPresent;adapter 重写;belief_shadow.go 删 lost-fall sweep;belief.go/track.go 不动。
- [ ] **工单 3**:oracle 测试套重建。旧 r5_calibration_lock_test/belief_test/fall_reason_test 按新 9 态全废,须先建已知 TP(fire)/FP(不 fire)oracle。
- [ ] **工单 4**:cd2b replay 单测——完成判据=cd2b fixture 在 ~22:25 fire。
