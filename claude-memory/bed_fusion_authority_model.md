---
name: bed_fusion_authority_model
description: 2026-06-04 sleepad/radar 床态融合权威模型(Layer0身份/Layer1三档置信含抗EMI/Layer2 cancel→fire vanish)+ silent_leftbed 真义 + P0 alarm_level 契约;已部署未commit
metadata: 
  node_type: memory
  type: project
  originSessionId: e8750e7d-da7e-421d-98f1-fc1e56c285ed
---

2026-06-04 卧室跌倒实测(201/Yang.R/radar 9D8A32A1CD2B + sleepad BM87224601641)驱动的两件治本,均**已部署 owlback.sensor/cardagg 未 commit(main)**。

## silent_leftbed 真义(用户校正)
不是"雷达看到地上的人",而是 **sleepad LeftBed 给"雷达冻结/精度甩偏"做权威对账**:sleepad(接触式)说离床,但 radar 还在床边显示 track(冻结残影 或 精度把摔倒的人定位进床)→ 矛盾=跌倒。sleepad 是消歧锚点。frozen-static track 进不了 lost_fall(被 MovingPrecondition 静止≥60s 守卫挡掉,治 CABB FP),所以唯有 sleepad LeftBed 能锚定这类。

## 床态融合权威模型(用户定的 spec,落在 track_manager BedSession,不碰 zoneengine 状态管道)
- **Layer 0 床身份**:单床房`bedCount==1`→radar+sleepad 必同床(RecordRadarEvent 内 radar InBed 无视时差记 RadarInBedInSessionMs);多床房→±15s 同向事件(同上床/同下床)才算同床。**±15s 不再是武装闸,而是最高置信档/多床身份关联器**(删了 scanSilentFallLeftBed 的 `if RadarInBedConfirmedMs==0 continue` 硬闸,sleepad 单独武装=不对称"未报上床不否定上床")。
- **Layer 1 三档融合置信**(`bedInBedConfidence()`,LeftBed 时 latch 到 `BedSession.InBedConfidence`):radar 印证在床→`min(sleepad90+radar70,100)`;radar 看到人非在床→`sleepad90`;**radar 全程无 track→`max(90−70,30)=30` 抗震动/EMI 假上床,地板 30 不归零**(真人静卧雷达也可能无回波)。
- **Layer 2 cancel→fire 四级**(`scanSilentFallLeftBed` 窗末):near-bed frozen track→fire(human-bed cell 豁免)/ `anyActiveTrack()` 房内有人→cancel / `roomLedgerEmpty()`(ExitRoom 过门,不信 np=0)→cancel / **都没+`InBedConfidence>=50`+有雷达归因→`engine_silent_leftbed_vanished`**。

## vanish-fire 治本的盲区
"干净摔进床后死角"(无椅子遮挡物):track 静止在床→lost_fall 的 MovingPrecondition 守卫判正常卧床不报;radar 丢轨后房内无 track。以前 silent_leftbed 窗末无 near-bed track 就无声 cancel。现在=sleepad LeftBed + 无 track + 无 ExitRoom + 置信足 → vanish-fire。低置信(EMI)走 else 抑制。新增字段 RadarSawTrackMs/RadarDeviceAddr(latchRadarTrackForBedSessions 每 Tick latch);新测试 VanishFires*/VanishSuppressedByEMI* 全绿。

## P0 alarm_level 契约(同轮治本)
roomengine alarm 路径([engine.go publishAIMessage](../../../owl/owlBack/wisefido-sensor/internal/roomengine/engine.go))漏发 `alarm_level`→cardagg [alarm_router.go:187](../../../owl/owlBack/wisefido-cardagg/internal/consumer/alarm_router.go) `level=="" 静默丢`(无日志)→Fall 引擎报了但 UI 全黑。zonealarm 的 mergeTriggerData 本就发 alarm_level=范本。修:① roomengine alarm 路径用 `alarm.Registry.DefaultLevel` 补 `alarm_level`(字段名沿用既成契约,非 EventItem 的 event_level);② cardagg 普通分支加 DefaultLevel 兜底+真丢时 Warn 留痕。**坑**:device_config 存的 Fall level 是非规范 "CRITICAL"(canonical 是 EMERG/ALERT),导入器待规范。绑卡/缓存其实没问题(AngleException 能入库证明路由通),是 alarm_level 缺失致丢。

关联 [[layout_authority_ai_correction_model]] [[belief_state_rule_engine_reframe]] [[feedback_signal_loss_lost_track_not_suppressible]] [[fall_fp_roots_and_todo]]
