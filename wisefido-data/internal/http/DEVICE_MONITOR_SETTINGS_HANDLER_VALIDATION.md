# DeviceMonitorSettings Handler 响应格式验证

本文档旨在验证新实现的 `DeviceMonitorSettingsHandler` 的 HTTP 响应格式是否与旧的 `StubHandler.SettingsMonitor` 保持一致。

## 验证目标

- `GET /settings/api/v1/monitor/sleepace/:deviceId` (GetDeviceMonitorSettings)
- `PUT /settings/api/v1/monitor/sleepace/:deviceId` (UpdateDeviceMonitorSettings)
- `GET /settings/api/v1/monitor/radar/:deviceId` (GetDeviceMonitorSettings)
- `PUT /settings/api/v1/monitor/radar/:deviceId` (UpdateDeviceMonitorSettings)

确保：
1. 响应状态码一致（200 OK）
2. 响应体结构一致（`Ok` 包装，`code` 和 `message` 字段）
3. 数据字段命名一致（`snake_case`）
4. 数据类型和内容一致
5. 可选字段在不存在时不会返回

---

## 1. `GET /settings/api/v1/monitor/sleepace/:deviceId` (GetDeviceMonitorSettings)

### 旧 Handler (StubHandler.SettingsMonitor) 响应示例

```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "left_bed_start_hour": 0,
    "left_bed_start_minute": 0,
    "left_bed_end_hour": 0,
    "left_bed_end_minute": 0,
    "left_bed_duration": 0,
    "left_bed_alarm_level": "disabled",
    "min_heart_rate": 0,
    "heart_rate_slow_duration": 0,
    "heart_rate_slow_alarm_level": "disabled",
    "max_heart_rate": 0,
    "heart_rate_fast_duration": 0,
    "heart_rate_fast_alarm_level": "disabled",
    "min_breath_rate": 0,
    "breath_rate_slow_duration": 0,
    "breath_rate_slow_alarm_level": "disabled",
    "max_breath_rate": 0,
    "breath_rate_fast_duration": 0,
    "breath_rate_fast_alarm_level": "disabled",
    "breath_pause_duration": 0,
    "breath_pause_alarm_level": "disabled",
    "body_move_duration": 0,
    "body_move_alarm_level": "disabled",
    "nobody_move_duration": 0,
    "nobody_move_alarm_level": "disabled",
    "no_turn_over_duration": 0,
    "no_turn_over_alarm_level": "disabled",
    "situp_alarm_level": "disabled",
    "onbed_duration": 0,
    "onbed_alarm_level": "disabled",
    "fall_alarm_level": "disabled"
  }
}
```

### 新 Handler (DeviceMonitorSettingsHandler.GetDeviceMonitorSettings) 响应示例

```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "left_bed_start_hour": 0,
    "left_bed_start_minute": 0,
    "left_bed_end_hour": 0,
    "left_bed_end_minute": 0,
    "left_bed_duration": 0,
    "left_bed_alarm_level": "disabled",
    "min_heart_rate": 0,
    "heart_rate_slow_duration": 0,
    "heart_rate_slow_alarm_level": "disabled",
    "max_heart_rate": 0,
    "heart_rate_fast_duration": 0,
    "heart_rate_fast_alarm_level": "disabled",
    "min_breath_rate": 0,
    "breath_rate_slow_duration": 0,
    "breath_rate_slow_alarm_level": "disabled",
    "max_breath_rate": 0,
    "breath_rate_fast_duration": 0,
    "breath_rate_fast_alarm_level": "disabled",
    "breath_pause_duration": 0,
    "breath_pause_alarm_level": "disabled",
    "body_move_duration": 0,
    "body_move_alarm_level": "disabled",
    "nobody_move_duration": 0,
    "nobody_move_alarm_level": "disabled",
    "no_turn_over_duration": 0,
    "no_turn_over_alarm_level": "disabled",
    "situp_alarm_level": "disabled",
    "onbed_duration": 0,
    "onbed_alarm_level": "disabled",
    "fall_alarm_level": "disabled"
  }
}
```

**对比结果**：
- ✅ `code` 和 `message` 一致
- ✅ `result` 结构一致，包含所有 Sleepace 配置字段
- ✅ 字段命名一致（`snake_case`）
- ✅ 数据类型一致（整数和字符串）
- ✅ 默认值一致（0 和 "disabled"）

---

## 2. `GET /settings/api/v1/monitor/radar/:deviceId` (GetDeviceMonitorSettings)

### 旧 Handler (StubHandler.SettingsMonitor) 响应示例

```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "radar_function_mode": 0,
    "suspected_fall_duration": 0,
    "fall_alarm_level": "disabled",
    "posture_detection_alarm_level": "disabled",
    "sitting_on_ground_duration": 0,
    "sitting_on_ground_alarm_level": "disabled",
    "stay_detection_duration": 0,
    "stay_alarm_level": "disabled",
    "leave_detection_start_hour": 0,
    "leave_detection_start_minute": 0,
    "leave_detection_end_hour": 0,
    "leave_detection_end_minute": 0,
    "leave_detection_duration": 0,
    "leave_alarm_level": "disabled",
    "lower_heart_rate": 0,
    "heart_rate_slow_alarm_level": "disabled",
    "upper_heart_rate": 0,
    "heart_rate_fast_alarm_level": "disabled",
    "lower_breath_rate": 0,
    "breath_rate_slow_alarm_level": "disabled",
    "upper_breath_rate": 0,
    "breath_rate_fast_alarm_level": "disabled",
    "breath_pause_alarm_level": "disabled",
    "weak_vital_duration": 0,
    "weak_vital_sensitivity": 0,
    "weak_vital_alarm_level": "disabled",
    "inactivity_alarm_level": "disabled"
  }
}
```

### 新 Handler (DeviceMonitorSettingsHandler.GetDeviceMonitorSettings) 响应示例

```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "radar_function_mode": 0,
    "suspected_fall_duration": 0,
    "fall_alarm_level": "disabled",
    "posture_detection_alarm_level": "disabled",
    "sitting_on_ground_duration": 0,
    "sitting_on_ground_alarm_level": "disabled",
    "stay_detection_duration": 0,
    "stay_alarm_level": "disabled",
    "leave_detection_start_hour": 0,
    "leave_detection_start_minute": 0,
    "leave_detection_end_hour": 0,
    "leave_detection_end_minute": 0,
    "leave_detection_duration": 0,
    "leave_alarm_level": "disabled",
    "lower_heart_rate": 0,
    "heart_rate_slow_alarm_level": "disabled",
    "upper_heart_rate": 0,
    "heart_rate_fast_alarm_level": "disabled",
    "lower_breath_rate": 0,
    "breath_rate_slow_alarm_level": "disabled",
    "upper_breath_rate": 0,
    "breath_rate_fast_alarm_level": "disabled",
    "breath_pause_alarm_level": "disabled",
    "weak_vital_duration": 0,
    "weak_vital_sensitivity": 0,
    "weak_vital_alarm_level": "disabled",
    "inactivity_alarm_level": "disabled"
  }
}
```

**对比结果**：
- ✅ `code` 和 `message` 一致
- ✅ `result` 结构一致，包含所有 Radar 配置字段
- ✅ 字段命名一致（`snake_case`）
- ✅ 数据类型一致（整数和字符串）
- ✅ 默认值一致（0 和 "disabled"）

---

## 3. `PUT /settings/api/v1/monitor/sleepace/:deviceId` (UpdateDeviceMonitorSettings)

### 旧 Handler (StubHandler.SettingsMonitor) 响应示例

**注意**：旧 Handler 的 PUT 方法返回的是与 GET 相同的配置项结构（这是 stub 实现的行为）。

```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "left_bed_start_hour": 0,
    "left_bed_start_minute": 0,
    // ... 所有配置字段
  }
}
```

### 新 Handler (DeviceMonitorSettingsHandler.UpdateDeviceMonitorSettings) 响应示例

```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "success": true
  }
}
```

**对比结果**：
- ⚠️ **响应格式不一致**：旧 Handler 返回配置项，新 Handler 返回 `{"success": true}`
- **分析**：
  - 旧 Handler 的 PUT 响应是 stub 实现，返回配置项是为了避免前端报错（不是真实行为）
  - 新 Handler 的 PUT 响应符合项目规范（与 `DeviceHandler.UpdateDevice`、`AlarmCloudHandler.UpdateAlarmCloudConfig` 等一致）
  - 新 Handler 的响应格式更符合 RESTful 规范（更新操作返回成功标志）
  - **结论**：这是预期的行为差异，新 Handler 的响应格式是正确的
  - **可选改进**：如果前端期望 PUT 返回更新后的配置，可以修改新 Handler 在成功后再次调用 GET 并返回配置

---

## 4. `PUT /settings/api/v1/monitor/radar/:deviceId` (UpdateDeviceMonitorSettings)

### 旧 Handler (StubHandler.SettingsMonitor) 响应示例

与 Sleepace 相同，返回配置项结构。

### 新 Handler (DeviceMonitorSettingsHandler.UpdateDeviceMonitorSettings) 响应示例

```json
{
  "code": 2000,
  "type": "success",
  "message": "ok",
  "result": {
    "success": true
  }
}
```

**对比结果**：
- ⚠️ **响应格式不一致**：与 Sleepace 相同的问题

---

## 5. 错误响应格式

### 旧 Handler 错误响应

旧 Handler 在 stub 模式下不会返回错误（总是返回默认配置）。

### 新 Handler 错误响应

```json
{
  "code": -1,
  "type": "error",
  "message": "error message here",
  "result": null
}
```

**对比结果**：
- ✅ 错误响应格式符合标准（使用 `Fail` 函数）
- ✅ 新 Handler 提供了更好的错误处理（设备不存在、类型不匹配等）

---

## 总结

### ✅ 完全兼容的部分

1. **GET 请求响应格式**：
   - Sleepace 和 Radar 的 GET 响应格式完全一致
   - 所有配置字段都正确返回
   - 字段命名、数据类型、默认值都一致

2. **错误响应格式**：
   - 使用标准的 `Fail` 函数
   - 错误信息清晰明确

### ⚠️ 需要确认的部分

1. **PUT 请求响应格式**：
   - 旧 Handler：返回配置项（stub 实现）
   - 新 Handler：返回 `{"success": true}`
   - **建议**：如果前端需要 PUT 返回更新后的配置，可以修改新 Handler 在成功后再次调用 GET 并返回配置

### 参考其他 Handler 的 PUT 响应格式

项目中其他 Handler 的 PUT 响应格式：

1. **DeviceHandler.UpdateDevice**：返回 `Ok(map[string]any{"success": true})`
2. **AlarmCloudHandler.UpdateAlarmCloudConfig**：返回 `Ok(map[string]any{"success": true})`
3. **TenantsHandler**：返回 `Ok(map[string]any{"success": true})`

**结论**：新 Handler 的 PUT 响应格式（返回 `{"success": true}`）与项目中其他 Handler 保持一致，符合项目规范。

### 可选的改进（如果前端需要）

如果前端期望 PUT 返回更新后的配置（与旧 stub Handler 行为一致），可以修改 `UpdateDeviceMonitorSettings` 方法：

```go
// 更新后返回更新后的配置
resp, err := h.deviceMonitorSettingsService.UpdateDeviceMonitorSettings(ctx, req)
if err != nil {
    // ... 错误处理
}

// 获取更新后的配置
getReq := service.GetDeviceMonitorSettingsRequest{
    TenantID:   tenantID,
    DeviceID:   deviceID,
    DeviceType: deviceType,
}
getResp, err := h.deviceMonitorSettingsService.GetDeviceMonitorSettings(ctx, getReq)
if err != nil {
    // ... 错误处理
}

writeJSON(w, http.StatusOK, Ok(getResp.Settings))
```

**注意**：这会导致额外的数据库查询，如果前端不需要，建议保持当前的 `{"success": true}` 响应格式。

---

## 验证结论

- ✅ **GET 请求**：完全兼容，响应格式一致
- ⚠️ **PUT 请求**：响应格式不同，但更符合 RESTful 规范
- ✅ **错误处理**：新 Handler 提供了更好的错误处理
- ✅ **字段命名**：完全一致（`snake_case`）
- ✅ **数据类型**：完全一致

**建议**：如果前端代码期望 PUT 返回配置项，可以考虑修改新 Handler 返回更新后的配置，以保持完全兼容。

