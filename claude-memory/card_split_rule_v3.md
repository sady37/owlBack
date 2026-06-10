---
name: card-split-rule-v3
description: "[REVIVED 2026-05-21 PM] card split 按 structural Bed count 决定 (space DDL 一次定型)；has_bed = ActiveBed (sensor 运行时 flag)；C8 的 structural-has_bed 当日撤销"
metadata: 
  node_type: memory
  type: project
  originSessionId: 921ec0a5-71d7-4de9-a51c-4f68586607a5
---

> ⚠️ **2026-05-23 三轮迭代终版** — Walk A → 简化版 37e413a → 终版（含 /88）。三种实卡 /80 + /88 + /96，2-step + M + lazy /80。
> split/merge 由 unit 内**结构性 bed 总数**决定；`has_bed` flag 仍 ActiveBed 语义（C3 — 给 sensor 用）。
> 权威规则：[owlBack/doc/rule.md](../../../owl/owlBack/doc/rule.md) §C C1 / C2 / C9 / C10。
> 实现：[card_reconcile.go buildExpected](../../../owl/owlBack/wisefido-data/internal/service/card_reconcile.go) + [card_static_service.go LATERAL JOIN](../../../owl/owlBack/wisefido-data/internal/service/card_static_service.go)。

---

## Split / Merge 规则（终版）

```
N = bedCount_in_unit          (整 unit 床数 /96 候选)
M = noBedRoomCount_in_unit    (不含 bed 的 room 数 /88 候选)

Step 0: unit 无 room → 无卡
Step 1: Unit
  N ≤ 1 → /80 only (merge，吸所有)
  N > 1 → split：M > 0 时预创建 /80
Step 2: 每个含 bed 的 room
  Room.N ≤ 1 → /88 卡 (吸 bed + room device)
  Room.N > 1 → N 张 /96 + room device → /80 (M=0 时 lazy create /80)
```

三种实卡 **/80 + /88 + /96**。/88 保留以维系 same-room 拓扑（Bedroom radar 与 BedA 设备同卡显示）。

**device/alarm 路由**：PG GiST `<<=` LPM 自动 — device <<= 现存 /96 落 /96；<<= 现存 /88 落 /88；否则 /80。

**Address 显示**（[card_static_service.go LATERAL](../../../owl/owlBack/wisefido-data/internal/service/card_static_service.go)）：按 card_id mask 分级查 room/bed —
- /80 → room=NULL bed=NULL → "Denver/A/101"
- /88 → room=exact bed=NULL → "Denver/A/101/Bedroom"
- /96 → room=parent /88 bed=exact → "Denver/A/101/Bedroom/BedA"

**Producer 红利**：alarm 无对应 /96 时自然汇 /88 或 /80；`alarm_events.producer` 列保留发起设备身份。

**核心原则**：Bed count 是 structural（不是 ActiveBed），shape 一次定型；device 绑/解绑、非 bed device 有无**永不**影响卡集合（C9+C10）。room space DDL 影响 /80 / /88 寿命，bed space DDL 跨阈值（unit N=1↔2，room N=1↔2）触发 /80↔/96 / /88↔/96 切换。

**S1.down.left 显示** 用 active child /88 卡的 `room_name`（不读 /80 自己的 room），不受 /80 LATERAL null 影响。

## has_bed 语义

`has_bed` = ActiveBed = scope 下任一 bed 已绑 device 才 TRUE（C3）。
**Why:** 这是给 sensor 消费的运行时 flag（"scope 内有没有可用 bed 设备"），不是结构标记。
**触发更新**：device 绑/解绑事件（C10 snapshot 路径），不是 space DDL。

## Z-002 修订

Producer 对齐 consumer（has_bed = ActiveBed），不是反过来。
`card_reconcile.go` 写 has_bed 改走 device 绑定状态；删 `HasBedBoundRadar` workaround（data_v2_todo）。

## 关联
- [[phase1_qinglan_v2_done]]
