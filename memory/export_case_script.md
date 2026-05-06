---
name: Case fixture export script
description: 统一脚本 scripts/export_case.sh 导出 ghost/fall fixture 到 doc/cases/<name>/，格式与 d5f7-ghost 一致
type: reference
originSessionId: 48426fae-acda-4d8d-ae6b-946fa83fcc2e
---
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
