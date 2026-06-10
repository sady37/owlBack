---
name: dbn_cutover_state
description: DBN ghost数学重做+cutover wire完成;两可逆开关DBN_FIRE/DBN_VETO_RECOVERY默认OFF;cutover序列待用户翻+live+删gate-list
metadata: 
  node_type: memory
  type: project
  originSessionId: f86f523b-5bda-4c73-ac49-9b2cf61a4b1c
---

2026-06-09 本会话完成 fall 检测的 DBN 化重做 + cutover wiring(委员会逐项 R6 收下)。**未来会话:别重做 ghost 检测,它已 DBN-native。**

**ghost 检测已从 gate-list 规则彻底重做成 DBN 自有 co-existence 对称(纯数学涌现)**:
- P1①转移耦合(belief/track.go `PredictCoupled(ρ)`,ρ=max 共存 P(TReal),孤立 ρ=0→P(Ghost)=0=long-lie 结构安全)
- P1②对称发射(Ghostness 源=motion+mirror 对称,**不再用单 track frozenGhost/realLO 病根**)
- P1③Room fall 发射 ×P(Real)(ghost 喂不动 SFallen,belief_shadow.go)
- P1④`DecideTauCtx` 的 C_FN 吃 OthersPresent(风险分层从代价涌现:独处 τ*0.55 必发/有人 0.671 容抑制)
- P1-final:`dbnMotionSymmetryGhost`(紧贴<100cm+同向 cos>0.866)+`dbnMirrorSymmetryGhost`(反射面镜像位有 VerdictReal partner≤50cm,复用 `reflectAcrossMirror`+`tm.GetInterferes()` lock-free)→ ghost 检测面 100% DBN-native。详 [[fall_detection_risk_stratified_design]]。
- 残留 b.Verdict 3 处**非检测**:veto_evidence 诊断 emit + Room 层 :305 VerdictGhost delete(生产过滤,gate-list 删 PR 一并退役:302)+ partner 读 VerdictReal(track_manager 核心非 gate-list,删后仍在,正当)。

**两可逆开关(belief_shadow.go,默认 OFF=部署零变化)**:
- `DBN_FIRE=1` → DBN 真发推断 fall(union firmware∨DBN,firmware floor 不动)+ gate-list 推断 fall 短路(fireFallCore/fireFall 开头 `if dbnFireEnabled return`)。**R0 在翻 ON 时结束**。DBN 输出 4-tag:`dbn_silent/dbn_lost/dbn_moving/dbn_pose_lying`(reason enum 单源 fall_reason.go)。
- `DBN_VETO_RECOVERY=1` → recovery-veto(纯误火摔后即起→抑制;自救真摔倒地≥15s→不抑制留告警)。现 would-veto 只 log,实际 retraction 留 cutover。

**cutover 序列(待用户决策,非代码)**:① 翻 `DBN_FIRE=1`(秒可逆)② live 护士确认 DBN fire 对/分类合理/veto 安全 ③ 稳定后删 gate-list 码(bathroom_fall/bedroom_fall/fall_rules_param/fall_verify/fall_exempt + :305/:302 + 9 红测)④ wire 实际 retraction + 开 DBN_VETO_RECOVERY 压误报。分类准确率走 live 护士标注(非合成,[[fall_data_is_artificial_test]])。

**协作**:施工方↔委员会经 owlBack/doc/feedback_p3.md(倒序);全程 shadow R0+单测+9 红基线+cutover #9 真摔照常发。委员会 R6 严验(亲跑不信声明)。
