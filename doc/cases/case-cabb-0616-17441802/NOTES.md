# case-cabb-0616-17441802 — Findings

room `fd00:0:3:411:1:200::/88`  tz=Asia/Shanghai(UTC+8)  window 17:44:00–18:02:30
两人(lid1 进门走动后坐 / lid2 17:45:27 进入站立),最终离场。**无真摔**(负样本)。

复现: `tools/Xsensorv1/replay-case.sh doc/cases/case-cabb-0616-17441802 1`
逐帧信念态见 `track_event_timeline.txt`(SELECTED + EXPANDED A/B 块)。

---

## Finding 1 — stillDiscount 移入 emission 的验证(commit 7a1ec6a,本次改动)

z/pose 直立证据从 floor 的 still 折扣移到 belief emission 直接压 SFallen。本案验证:

- **活动期 SFall 被显著压低**(前向滤波累积):17:45:27 `0.107→0.012`(~9×);17:44:50 walk `0.026→0.003`;17:47:27 sit `0.221→0.192`(supSit×0.8 温和)。
- **无误报**,正确收敛 `top=Left`、SLeft=0.972。
- 累积发生在**身份稳定的 lid1**(见 Finding 2),验证可靠。
- ⚠️ 本案是退出负样本,**验不到 FN**。`z≥80 误压真摔` 的 FN 风险须 **cd2b** replay 确认(cd2b firmware `position_z=105` 出现在 pose=6 Lying)。

副作用记录:站立静止 + z∈[30,80) 现在 emission(只压 z≥80,不压 pose=Stand)与 floor(折扣已摘)两头都不压,纯靠 floor raw 计时兜——FN-safe 方向,但比旧码早 ~80s。

---

## Finding 2 — 🔴 ExitRoom 后 logicID churn(既有身份层缺口,独立于本次改动)

**现象**: belief 层整数 LogicID(`dbn.lid`)在 **17:47:29 起暴增 lid 3→145**(几乎每帧一个新号),持续到窗尾。

**分段**:
- `17:44:37–17:47:27`(ExitRoom 前): 身份**稳定**,全程只有 `lid1`(坐,sb 持续涨)/ `lid2`(站)。无 switch-id。
- `17:47:29+`(ExitRoom@17:47:27 后): **每帧新 logicID**。

**机理**: 坐着的人(lid1)ExitRoom 离场后,其 "absent 续算" 轨(冻结在 `(-40,290)`)+ 现存轨的 present/absent 交替上报,使 census 的"最小作功距离关联"每帧都判 raw 距最近现存 track `>AssocCm` → **发新 logicID**。

**影响(为何本案无害但是潜在 FN)**:
- belief 滤波**按 LogicID 做 map key**(`engine.go: filters map[int]*belief.Filter`,新号 → `NewFilter` 起始 Prior=Empty 0.85)。
- 每个 churn 出的新号 = **全新滤波、SFall 从 0 起** → 这正是 ExitRoom 后 SFall 恒=0 的原因(churn 轨永远累积不起来)。
- 本案: churn 时房间已 `top=Left`、SFall=0,**不影响输出**。
- ⚠️ **潜在 FN**: 若**真摔发生在此类 churn 期间**,SFallen 每帧被重置 → 永远攒不到阈 → **漏报**。即 `[[w3_3_realness_wired_rebirth_fn]]` 的 "rebirth FN 根因=入口丢 track_id" 的另一触发路径:**ExitRoom 后续算轨 + present/absent 交替 → AssocCm 误拆**。

**归属**: census 关联鲁棒性缺口,**与 stillDiscount→emission 改动正交**(身份/census 层未碰)。建议单开排查:
1. ExitRoom 离场轨应及时 drop / 不再参与 AssocCm 关联(避免 stale 冻结点干扰现存轨关联)。
2. present/absent 交替帧的关联应容忍(absent 帧不应触发现存轨重拆号)。

证据: `track_event_timeline.txt` EXPANDED A(17:45 稳定 lid1/lid2)对比 EXPANDED B(17:47:29 起 churn)。
