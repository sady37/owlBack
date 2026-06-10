---
name: Track.ToFieldMap 零值 bed_status 泄漏（已全局清理）
description: 2026-05-19 commit 5940756 — ToFieldMap 改 omit-when-zero；qinglan 局部手写 map 撤回；5 Tier1 case lock 契约
type: project
originSessionId: 2540d0c0-e995-418d-a1a4-3e2d3075ad3a
---

## 已全局清理（2026-05-19 commit `5940756` 起步，commit `73304b5` 终态）

终态：`Track.BedStatus *int` 三态 — `nil = 未知/不适用`、`*0 = 在床`、`*1 = 离床`。与同 struct 的 PositionX/Y/Z 一致 idiomatic Go pattern。

**为什么不用 -1 sentinel**：Go 零值 0 会与"在床"语义碰撞；任何 `observation.Track{}` 字面量默认值就是"在床"，雷达 caller 需纪律性地显式写 `BedStatus: -1` 否则 silent bug 回归。`*int` 用 nil 表示"未知"，零值 nil 天然安全，不需要 caller discipline。

中间步骤（已被覆盖）：先把 ToFieldMap 改 omit-when-zero（5940756），后用 audit 发现 radar/heartbeat Track 默认 0 仍可能被误写为"在床"，遂改 *int 三态（73304b5）。

原"0 始终写入"是因为误以为 sleepace 0=在床要走这条路径，实地 audit 后确认所有真正需要 0/1 语义的路径都不走 ToFieldMap：

- sleepace 床事件 → 直接 build 手写 map（bed_status 从 decoded vital map 顶层带）
- silent_fall alarm → engine.publishAIMessage 内 `if BedStatus != 0` 已经 omit-when-zero
- 唯一 ToFieldMap caller（qinglan radar heartbeat + sleepace device heartbeat）都是 bed-agnostic，强制写 0 是污染下游协议的噪声

## 验证

- 5 Tier1 case 在 `owl-common/observation/track_test.go`：omit zero / write non-zero / round-trip 0 / round-trip 1 / 其它 int 字段保持 omit
- 撤回 qinglan publishRadarMonitorHeartbeat 局部手写 map，回归 canonical ToFieldMap 单一路径
- FromFieldMap 读不到键默认 0，round-trip 语义安全

## 关联

- 来源 backlog [[p4_next_session_prompt]] item N（H 之前未做、I 已完工、N 本次完工）
