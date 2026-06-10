---
name: Case fixture export script
description: 两个标准 case 导出脚本 export_case.sh(JSON fixture) + export_bed_test_record.sh(人类可读 txt,带 tid 列)，都落 doc/cases/<name>/
type: reference
originSessionId: 48426fae-acda-4d8d-ae6b-946fa83fcc2e
---
两个标准脚本，都落 `owlBack/doc/cases/<name>/`，**禁止再手写新格式导出脚本**：

1. **`scripts/export_case.sh`** — JSON replay fixture（喂 belief_replay_test）。
2. **`scripts/export_bed_test_record.sh`** — 人类可读 txt 复盘/对账（radar.track + radar.heart
   + sleepad.track + event_log + alarm_events，逐秒排序，自动派生 >>> STATE bed/room 转换行）。
   2026-06-06 标准化：tid(track_id) 提成**独立列**（时间|dev|tid|类型|明细），表头带 tid 语义图例
   （0-8 真人 / 9 vital / 10 space即np事件 / 11 device 心跳帧 / 88 无目标 / - 派生态无轨）。
   用法：`./scripts/export_bed_test_record.sh <radar_uid> <sleepad_uid|-> "<start>" "<end>" --tz <TZ> [name]`
   → `doc/cases/<name>/test_record.txt`。无床垫传 `-`。临时草稿目录 export_1137_1146/ 非标准，勿沿用。

---
（以下为 export_case.sh 细节）

位置：`owlBack/scripts/export_case.sh`

输出目录：`owlBack/doc/cases/<case_name>/`，与既有 `d5f7-ghost/`、
`case_lostfall_cd2b_11351148/` 同级、同格式。

每个 case 目录两个文件：
- `room_layout.json` — 整行 rooms（room_id / room_name / layout_config）
- `<YYYY-MM-DD_HH-MM>_to_<HH-MM>_<TZ>.json` — `iot_timeseries` 行数组，
  字段 `[id, device_uid, device_id, timestamp(ms), topic_type, category, data_value]`

用法（`--tz` **必填**，否则 exit 1）：
```bash
./scripts/export_case.sh <device_uid> "<start>" "<end>" --tz <IANA_TZ> [<case_name>]
# Denver 设备：
./scripts/export_case.sh 9923003AB17F "2026-04-29 07:10:00" "2026-04-29 07:30:00" \
  --tz America/Denver 9923-kitchen-0429
```

**时区坑**：服务器 `$TZ` 是 PDT (UTC-7)，但 demo tenant 在 Denver MDT (UTC-6)，
差 1 小时。`iot_timeseries.timestamp` 是 UTC epoch ms（firmware event_since 直
传，wisefido-iot 不翻译），所以"7:10 本地"映射到 epoch-ms 必须用设备所在地时区。
脚本第一版没设防默认 LA，导致导出窗口错 1 小时（同 device 同段时间错查 110 行
vs 1788 行）。现在 --tz 强制必填。

**Why**：用户明确反对每次手写不同格式的导出脚本。所有未来 ghost / fall /
其他场景 fixture 都用这个脚本，输出落 `doc/cases/`。

**How to apply**：用户要求"把某段数据存为 case"时，直接调这个脚本，不要再写
新脚本或自定义路径（除非用户明确要求）。
