---
name: silent_leftbed_fall_recovery_window_gap
description: 2026-06-05 自救型床边跌倒被 60s 恢复窗 cancel→零记录;实测 09e7/0978 案例;建议 cancel 也留低severity event
metadata:
  type: project
---

2026-06-05 用真实测试(radar 9e7 9D8A...09E7 + sleepad 978 BM87...0978,11:37-11:46 MDT=17:37-17:46 UTC)验证段 4c `scanSilentFallLeftBed` 的一个缺口。

**机制(已存在,正是"跌床"设计)**:sleepad LeftBed + radar 仍在 Bed 邻域(≤100cm)= silent/bedside fall。`track_manager.go:1730` 段 4c。等待窗 `WaitVitalSec=60s`(有HR/RR+单人)/ `WaitNoVitalSec=120s`;`BedNeighborhood=100cm`(fall_rules_param.go:145-147)。裁决:床邻域有 track→FIRE(`ReasonSleepadRadarConflict`,Risk=100 CRITICAL);房内有 track 不在邻域→cancel active_track_in_room;`roomLedgerEmpty`(ExitRoom)→cancel exited_room;InBedConf≥50且房内无track→vanish-fire(摔进床后死角)。

**这次为什么没报(精确)**:17:45:50 sleepad LeftBed(此前 HR=68 单人)→ 等待窗=60s → 17:46:50 才裁决。但人 17:46:34 起身(radar LeftBed)、**17:46:43 ExitRoom+np=0**(roomLedger 空)→ 裁决时命中 `roomLedgerEmpty` → **cancel "exited_room"**。即:人在床边地面(z=0,43帧实测 z0=36/43,贴床边 sd≈25,未走开)躺了 44s,**在 60s 裁决窗到期前自己爬起走出房间** → 归为正常离开,**连 event 都不留**。

**缺口本质**:60s 窗是二元"恢复了吗"闸;一次真实 44s 跌地因自救被抹掉。老人护理里**能自救的床边跌倒=下一次救不回的先行指标**;cancel-on-exit 分不清"稳稳起身"vs"摔疼硬撑挪出"。

**雷达精度分型(用户)对应判据**:phantom(椅子/被子/无人,无微动)→几十秒丢静止 track→裁决时床邻域无 track→正确不报;真人跌床(有微动 z=0 贴床边)→track 不丢→躺满 60s 会 FIRE。判据本身能分真人/家具,缺的是自救情形的留痕。

**建议(推方向2,未实施)**:cancel 也留痕——冲突(z≈0+床邻域)持续 ≥15-20s 后才恢复的,即便 cancel 也发低severity `bedside_fall_self_recovered` event(非CRITICAL)供跌倒风险趋势;判别用 z≈0 贴地驻留时长。不动 alarm 阈值。

**待核实**:段 4c arm/cancel 走 `tm.logger.Info`(silent_fall_leftbed_cancelled)**不写 sensor_decision_log**,故窗口内 sdl=0 不能证明检测器跑没跑;需查 sensor journal 17:46:43-50 确认是"武装→exited_room cancel"而非 roomengine 没消费该 radar。关联 [[bed_fusion_authority_model]] [[fall_rules_three_classes]] [[lost_track_fall_detection_envelope_gate]]。
