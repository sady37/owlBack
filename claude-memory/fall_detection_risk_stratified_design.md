---
name: fall_detection_risk_stratified_design
description: fall检测统领原则=一切看风险;ghost=co-existence只多track;否决激进度按人在场分层(独处永发/有人激进)
metadata: 
  node_type: memory
  type: project
  originSessionId: f86f523b-5bda-4c73-ac49-9b2cf61a4b1c
---

2026-06-09 用户拍板 fall 检测+否决体系的**统领设计原则:一切看风险**。完整模型:

**两正交轴(先判真伪,再判摔型)**:
- **轴1 真伪(ghost)= co-existence,只在 ≥2 track 时做**。ghost=真人的反射→必有共存 partner→**1 track 根本不进 ghost 分支(=真人,永发)**。判据始终 co-existence(镜像/运动对称,需另一 Real partner;gate-list `checkMirrorSymmetry` 条件#1=房间有另一 Real track,做对了)。**绝不用单 track 的静止/realness 判 ghost**——DBN 现在的 `shadowFrozenArtifact(单 ts)`/`realLO`/harness realness-ghost 是**病根**(把孤立躺地真受害者误判 ghost=long-lie 灾难)。
- **轴2 摔型(still/lost/moving)= 延时换准确率(平衡术),只对 real track**。dwell 越久 fallLR 越高(still)/丢轨恢复窗(lost)/倒姿确认窗(moving)。`survival.go fallLRFromDwell`+`TauConfirm{CFP:55,CFN:45}` 是平衡旋钮。**轴1 必须前置**——否则 ghost 静止永不动,被 dwell 时间喂成假 still-fall。

**★风险分层(核心)——ghost 否决激进度 ∝ 人在场**:
- **1 track 独处 = 最高风险(摔了没人救)→ 全发,阈值=∞,绝不否**。
- **≥2 track 有人在场 = 低风险(误否真摔也有人能发现/帮)→ 激进用 ghost 置信度否(放低阈值大胆砍)**。≥2 track 不只"ghost 可检测",更是"**误否代价低**"(另一真人是安全网)。

**这正面解「误报太多=无用」**:误报大头在多 track 场景(那里否得起,风险低),独处少数真摔保护到底——激进砍 FP 与保护独处真摔**用风险分开,不矛盾**。

**How to apply**:
- cutover veto 现 `coExist:=len(bases)>=2 && VerdictGhost`(委员会 6dafdc6 收;belief_shadow.go),方向对但应从二元升**风险分层 ghost 置信度阈**(≥2 放低激进/1 track ∞)。
- endgame(删 gate-list 前置):DBN 自己的 co-existence ghost 检测替掉单 track frozenGhost/realLO,按 track 数/人在场调阈。
- recovery-veto 同纳风险框架:独处自救保守(留告警),有人在激进。
- 关联 [[silent_leftbed_fall_recovery_window_gap]] / [[partial_monitoring_fall_suppression_law]] / [[fall_data_is_artificial_test]];cutover 5 条安全包络见 feedback_p3 6c376e4。
