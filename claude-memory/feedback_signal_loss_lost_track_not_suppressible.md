---
name: feedback_signal_loss_lost_track_not_suppressible
description: 突然丢信号的 lost_track 不可抑制——真跌倒与信号丢失在雷达数据上不可分，保留告警
metadata: 
  node_type: memory
  type: feedback
  originSessionId: 9d69b66d-2014-414e-a786-d3a3658c2ecb
---

2026-06-03 用户拍板:**雷达突然丢信号的 lost_track 误报，不要做"抑制闸"。** 真跌倒同样表现为突然丢锁——真 Fall 与信号丢失在雷达数据本身**不可分**，人也判不了。标成 fall 是诚实且安全的（宁可误报不可漏报）。

**Why**:任何"识别信号丢失→抑制 lost_fall"的闸（如 CABB frozen-static 的"零方差冻结→抑制""z 掉 0""pose 无 2/5"等），都会把**一类可能的真跌倒**一起压掉 = 漏报，不可接受。这类不是软件可判的逻辑 bug，是雷达硬件/安装固有局限（ceiling 雷达丢近旁静止目标）。

**How to apply**:遇到"人没走(无 ExitRoom)+ 静止/近距丢锁"的 lost_track FP，**不抑制、不降级，保留 CRITICAL**。真正缓解只能走硬件层（调安装位/参数、加门磁/sleepad 多源佐证），不在 engine 里靠"判信号丢失"做抑制。**例外（可安全抑制的）仅限能用结构性证据证伪的**：① 同 logic_id 仍活在另一条 track（换ID游戏，[[lost_track_fall_detection_envelope_gate]] id-swap 闸）；② ExitRoom 过门空间证据证明人已走（空房账闸，治 MoM/D5F7 残影）。这两类有硬证据；CABB 这类纯丢锁无证据 → 不动。详见 doc/cases/cabb-bathroom-frozen-static-FP-0603/README.md。
