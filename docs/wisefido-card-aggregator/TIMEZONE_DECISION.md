# 时间字段处理（UTC vs 单元时区）决策

## ✅ 结论

- **存储/传输统一使用 UTC**（推荐：Unix epoch 秒/毫秒 或 RFC3339 UTC）
- **展示时按 unit 的 IANA timezone 格式化**
  - `units.timezone` 字段在 `owlRD/db/05_units.sql` 已定义（例如 `"America/Los_Angeles"`）

## 为什么这样做

- **一致性**：后端所有服务（TimescaleDB/Redis/日志）统一用 UTC，避免跨服务时区歧义
- **可追溯**：排查问题时只需要对齐 UTC 时间线
- **前端体验**：同一套 UTC 数据可在不同地点用不同 timezone 展示

## 在本系统中的落地建议

### 1) Redis realtime/alarm/full 缓存
- `vital-focus:card:{card_id}:realtime`：保留源数据时间戳为 UTC（epoch 或 RFC3339）
- `vital-focus:card:{card_id}:alarms`：`triggered_at` 使用 UTC（epoch 或 RFC3339）
- `vital-focus:card:{card_id}:full`：原则同上

### 2) `bed_status_timestamp` / `status_duration` 这类“展示字段”
- **建议在 API 层（wisefido-data）计算**：
  - API 层可以一次性拿到 `units.timezone` 并做统一格式化
  - 前端只负责展示，不需要再计算时长/时区
- 如果暂时没有 API 层，也可以：
  - **先在 full 缓存里保存 UTC epoch**（例如 `bed_status_timestamp_utc`），前端用 `units.timezone` 格式化

## 需要确认/后续补充

- `wisefido-data` 未实现前：前端拿不到 `units.timezone` 的话，需要先通过 DB/API 获得 timezone
- 若要完全对齐前端接口（`bed_status_timestamp` 为字符串）：建议等 API 层实现时统一输出格式


