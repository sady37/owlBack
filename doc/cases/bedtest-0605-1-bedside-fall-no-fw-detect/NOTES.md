# bedtest-0605-1 — 床边跌倒 / 固件没判出 / 段4c 被自救取消 / 无告警

- 设备: radar 9D8A326309E7 (9e7) + sleepad BM87224700978 (978)
- 窗口: 2026-06-05 11:37:50–11:46:30 MDT = 17:37:50–17:46:30 UTC
- 数据: `test_record.txt`（脚本 `scripts/export_bed_test_record.sh` 生成）

## 时间线要点
- 17:38:49 radar 固件 InBed → 17:40:36 sleepad InBed（接触式权威晚锁 ~107s）
- 17:45:50 sleepad LeftBed（人翻下床、跌在床边地面）
- 17:46:34 radar LeftBed（窗口边缘）/ 17:46:43 ExitRoom + number_people=0

## 关键结论
1. **固件全程没判出跌倒**：radar pose 全程 ≈ 3（走），从未 pose=2/5。`alarm_events=0`、`sensor_decision_log=0`。
2. **HR/RR 全来自 sleepad**；radar.heart 不带数值，radar event HR/RR 恒 -1。
3. **段 4c（silent-leftbed-fall）武装后被取消**：sleepad LeftBed 起算等待窗 60s（有HR/RR+单人），
   但人在 44s 起身、53s 走出房间（ExitRoom 17:46:43），裁决点（17:46:50）命中 `roomLedgerEmpty`
   → cancel "exited_room" → **零记录**。
4. **缺口**：60s 是二元"恢复了吗"闸，自救型床边跌倒（=下次救不回的先行指标）连 event 都不留。
   建议：cancel 时若 z≈0 床邻域驻留 ≥15–20s，补发低 severity `bedside_fall_self_recovered` event。
   详见记忆 `silent_leftbed_fall_recovery_window_gap`。

## 与 #2 对照
同样床边跌倒，**固件能否看见 pose=5 决定报不报**。#2（fw-detect）固件判出了 → EMERG。
