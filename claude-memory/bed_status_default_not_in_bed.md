---
name: bed-status-default-not-in-bed
description: 默认 bed_status=1 (NotInBed)；在床需要证据（Sleepace InBed event OR radar vital signs），缺省视为不在床
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 36f14658-2d5a-470c-b075-31726dab854a
---

`hasBedDevice=true` 但无 fresh BedState 时，cardagg 必须默认 bed_status=1 (NotInBed)，FE 渲 outofbed icon。

**Why**: "在床"是有创伤的判定（影响 alarm / sleep stage / vital interpretation），必须有证据 — Sleepace 直接 InBed event 或 radar 测到 vital signs。没证据就是不在床。lying-bed-black icon **专属** sleep_stage=0/8 的子状态，bed_state 未知不能借这个 icon 当 "no-data 默认"。

**How to apply**:
- BedStatus 数据流：sensor 融合 Sleepace event + radar vital → 写 BedState；cardagg builder 读 BedState
- 没读到 BedState 时，cardagg 强制填 `bed_status=1`（[card_display_builder.go:42-47](../../../owl/owlBack/wisefido-cardagg/internal/consumer/card_display_builder.go#L42)）
- FE icon 优先级：`bed_status === 1 → outofbed`，先判 bed_status，再判 sleep_stage
- icon → state 不可逆推：bed_state 是父，sleep_stage 是子

历史教训：2026-05-23 我曾把 lying-bed-black 当 no-data 默认，用户立刻纠正"先有 bedstate 才有 sleepstage"——bed_state 的语义优先级在 sleep_stage 之上，不能用 sleep_stage 的 fallback icon 替代 bed_state 未知。
