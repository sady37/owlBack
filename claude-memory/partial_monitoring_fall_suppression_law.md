---
name: partial_monitoring_fall_suppression_law
description: "铁律——非全空间监控下,只有极近的两事件能排除二义 lost-fall;stale/durable 占用不可用,会漏真摔"
metadata: 
  node_type: memory
  type: feedback
  originSessionId: f86f523b-5bda-4c73-ac49-9b2cf61a4b1c
---

2026-06-08 用户拍死的安全铁律(#3 Neighbor wire 设计收口)。**系统不是全空间监控**(房间之间有盲区/过道未覆盖/firmware 漏 ExitRoom),所以**无法证明「人不在本房」**——"上次看到在邻房"≠"此刻还在邻房",人完全可能起身穿过盲区走到本房**真摔**。

**Why**:委员会(审查61)曾提"durable 占用不过期"(老人 8pm 进卧室躺床、12am 浴室鬼影摔→卧室还占着→压浴室假警)。用户一刀否决:这个"卧室还占着"可能**只是我们没看见他出去**,浴室那一摔可能是真的→拿 stale 占用去压=**压掉真摔=漏报**(最坏)。stale/durable「上次在哪」是**最危险的抑制**,不是最强证据。

**How to apply**(唯一能有效排除二义 lost-fall 的判据):
- **仅压二义性 lost-fall**:本房 track 消失、"摔了还是走了"从本房自己分不清的那一类(=CABB 冻结 ghost/John.Y lost_track 主类)。**已确证跌倒(firmware pose=5/有房内证据的 in-room fall)不碰**(对齐 R5-lock:ObsNeighbor 只在许可压制清单)。
- **判据=两事件极近 + 有向**:`邻房 EnterRoom/InBed 的 ts − 本房 st.lastSeenMs ∈ [−jitter(~5s), 窗(默认 60s,30-60 可调)]`。人**先**从本房消失、**后**在邻房出现(先后),−jitter 只给传感器时序抖动留余量。
- **窗是固定绝对阈,不随房间距离/邻接伸缩**(用户否决"窗=穿行时间":穿行越久=穿过盲区越多=摔机会越大,放宽窗等于在最该警惕时松手)。超窗→中间可能盲区真摔→**不排除,保留告警**。
- **stale/durable 一律不算**:邻房账"一直占用"/老人"久躺床"都证不了此刻在哪→不构成排除证据。
- **N-3 sole-resident 门仍留**:紧窗证"时间"(就这会儿挪过去),门证"是同一个人"(多 resident 进邻房的可能是另一人→不排除)。两者合起来才充分。
- **盲区诚实**:人走进未覆盖区→邻房无事件→无法证明→保留告警(可能 FP,绝不漏真摔)。partial monitoring 固有极限,不能假装知道。

落地:`FallRulesParam.Neighbor.{HandoffWindowMs,JitterMs}` + ObsNeighbor 只在 beliefShadowTick lost-track 路径喂。详见 [[neighbor_wire_build_spec]] / [[belief_state_rule_engine_reframe]]。**推翻审查61 durable 受理条件**,改记此版为终裁。
