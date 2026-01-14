# Radar 报警处理逻辑验证文档

## 一、验证目标

验证 `wisefido-radar` 中的报警处理逻辑是否正确实现，包括：
1. 报警使能配置查询是否正确
2. 事件/统计数据到报警类型的映射是否正确
3. Stream 分离逻辑是否正确（event/stat → alarm stream）
4. 元数据字段处理是否正确（null 值处理）

## 二、验证环境准备

### 2.1 数据库准备

需要在 PostgreSQL 数据库中准备测试数据：

```sql
-- 1. 创建测试租户
INSERT INTO tenants (tenant_id, tenant_name) 
VALUES ('test-tenant-001', 'Test Tenant') 
ON CONFLICT DO NOTHING;

-- 2. 创建测试设备
INSERT INTO devices (device_id, tenant_id, serial_number, uid, device_name, status)
VALUES 
  ('test-device-001', 'test-tenant-001', 'TEST001', 'TEST_UID_001', 'Test Radar 1', 'online'),
  ('test-device-002', 'test-tenant-001', 'TEST002', 'TEST_UID_002', 'Test Radar 2', 'online')
ON CONFLICT DO NOTHING;

-- 3. 创建设备报警配置（alarm_device）
-- 设备 1：启用 Fall 和 OfflineAlarm，禁用其他
INSERT INTO alarm_device (device_id, tenant_id, monitor_config)
VALUES (
  'test-device-001',
  'test-tenant-001',
  '{
    "alarms": {
      "Fall": {"level": "EMERGENCY", "enabled": true},
      "SuspectedFall": {"level": "WARNING", "enabled": false},
      "OfflineAlarm": {"level": "WARNING", "enabled": true},
      "PoorReception": {"level": "WARNING", "enabled": false},
      "Radar_AbnormalHeartRate": {"level": "EMERGENCY", "enabled": false},
      "VitalsWeak": {"level": "EMERGENCY", "enabled": true}
    }
  }'::jsonb
)
ON CONFLICT (device_id) DO UPDATE SET monitor_config = EXCLUDED.monitor_config;

-- 设备 2：所有报警都禁用
INSERT INTO alarm_device (device_id, tenant_id, monitor_config)
VALUES (
  'test-device-002',
  'test-tenant-001',
  '{
    "alarms": {
      "Fall": {"level": "EMERGENCY", "enabled": false},
      "OfflineAlarm": {"level": "WARNING", "enabled": false}
    }
  }'::jsonb
)
ON CONFLICT (device_id) DO UPDATE SET monitor_config = EXCLUDED.monitor_config;
```

### 2.2 代码文件位置

需要验证的关键文件：
- `owlBack/wisefido-radar/internal/repository/device.go` - 报警使能配置查询
- `owlBack/wisefido-radar/internal/alarm/device_alarm_handler.go` - 报警处理器
- `owlBack/wisefido-radar/internal/consumer/mqtt_consumer.go` - MQTT 消费者（Stream 分离逻辑）

## 三、验证步骤

### 3.1 验证报警使能配置查询

#### 测试用例 1：查询已配置设备的报警使能
```go
// 测试代码位置：可以创建测试文件或使用 Go 测试
// 文件：owlBack/wisefido-radar/internal/repository/device_test.go

func TestGetAlarmEnablement_DeviceWithConfig(t *testing.T) {
    // 1. 初始化 DeviceRepository
    // 2. 调用 GetAlarmEnablement(ctx, "test-tenant-001", "test-device-001")
    // 3. 验证返回结果：
    //    - enablement["Fall"] == true
    //    - enablement["SuspectedFall"] == false
    //    - enablement["OfflineAlarm"] == true
    //    - enablement["PoorReception"] == false
}
```

**预期结果**：
- `enablement["Fall"]` = `true`
- `enablement["SuspectedFall"]` = `false`
- `enablement["OfflineAlarm"]` = `true`
- `enablement["PoorReception"]` = `false`
- `enablement["VitalsWeak"]` = `true`

#### 测试用例 2：查询未配置设备的报警使能
```go
func TestGetAlarmEnablement_DeviceWithoutConfig(t *testing.T) {
    // 1. 调用 GetAlarmEnablement(ctx, "test-tenant-001", "non-existent-device")
    // 2. 验证返回结果：空的 AlarmEnablement map（所有报警类型都未启用）
}
```

**预期结果**：
- 返回空的 `AlarmEnablement` map（`len(enablement) == 0`）
- 所有报警类型查询都返回 `false`

### 3.2 验证事件到报警类型映射

#### 测试用例 3：Event type=2 (pose) 映射
```go
// 测试代码位置：owlBack/wisefido-radar/internal/repository/device_test.go

func TestGetPossibleAlarmTypesFromEvent_Pose(t *testing.T) {
    // 输入：event 数据，category="pose"
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "pose",
            "pose": "Fall",
            "track_id": 1,
        },
    }
    
    // 调用
    possibleAlarms := repository.GetPossibleAlarmTypesFromEvent(dataValue)
    
    // 验证
    // 预期：包含 "Fall", "SuspectedFall", "SittingOnGround"
}
```

**预期结果**：
- `possibleAlarms` 包含：`"Fall"`, `"SuspectedFall"`, `"SittingOnGround"`

#### 测试用例 4：Event type=5 (isOnline) 映射
```go
func TestGetPossibleAlarmTypesFromEvent_IsOnline(t *testing.T) {
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "isOnline",
            "device_status": "offline",
        },
    }
    
    possibleAlarms := repository.GetPossibleAlarmTypesFromEvent(dataValue)
    
    // 验证：包含 "OfflineAlarm"
}
```

**预期结果**：
- `possibleAlarms` 包含：`"OfflineAlarm"`

#### 测试用例 5：Event type=7 (signal_poor) 映射
```go
func TestGetPossibleAlarmTypesFromEvent_SignalPoor(t *testing.T) {
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "signal_poor",
            "recovery": "signal_poor",
        },
    }
    
    possibleAlarms := repository.GetPossibleAlarmTypesFromEvent(dataValue)
    
    // 验证：包含 "PoorReception"
}
```

**预期结果**：
- `possibleAlarms` 包含：`"PoorReception"`

#### 测试用例 6：Event type=9 (other, alarmType="1") 映射
```go
func TestGetPossibleAlarmTypesFromEvent_Other_AlarmType1(t *testing.T) {
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "other",
            "alarmType": "1",
        },
    }
    
    possibleAlarms := repository.GetPossibleAlarmTypesFromEvent(dataValue)
    
    // 验证：包含 "Radar_LeftBed"
}
```

**预期结果**：
- `possibleAlarms` 包含：`"Radar_LeftBed"`

### 3.3 验证统计数据到报警类型映射

#### 测试用例 7：Stat sleep category - heart_state 异常
```go
func TestGetPossibleAlarmTypesFromStat_HeartStateAbnormal(t *testing.T) {
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "sleep",
            "heart_state": "Heart rate high",
            "heart_rate": 120,
        },
    }
    
    possibleAlarms := repository.GetPossibleAlarmTypesFromStat(dataValue)
    
    // 验证：包含 "Radar_AbnormalHeartRate"
}
```

**预期结果**：
- `possibleAlarms` 包含：`"Radar_AbnormalHeartRate"`

#### 测试用例 8：Stat sleep category - breath_state 异常
```go
func TestGetPossibleAlarmTypesFromStat_BreathStateAbnormal(t *testing.T) {
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "sleep",
            "breath_state": "Breath rate low",
        },
    }
    
    possibleAlarms := repository.GetPossibleAlarmTypesFromStat(dataValue)
    
    // 验证：包含 "Radar_AbnormalRespiratoryRate"
}
```

**预期结果**：
- `possibleAlarms` 包含：`"Radar_AbnormalRespiratoryRate"`

#### 测试用例 9：Stat sleep category - breath_state Apnea
```go
func TestGetPossibleAlarmTypesFromStat_BreathStateApnea(t *testing.T) {
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "sleep",
            "breath_state": "Apnea",
        },
    }
    
    possibleAlarms := repository.GetPossibleAlarmTypesFromStat(dataValue)
    
    // 验证：包含 "Radar_ApneaHypopnea"
}
```

**预期结果**：
- `possibleAlarms` 包含：`"Radar_ApneaHypopnea"`

#### 测试用例 10：Stat sleep category - vital_signs_state weak
```go
func TestGetPossibleAlarmTypesFromStat_VitalSignsWeak(t *testing.T) {
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "sleep",
            "vital_signs_state": "Vital signs weak",
        },
    }
    
    possibleAlarms := repository.GetPossibleAlarmTypesFromStat(dataValue)
    
    // 验证：包含 "VitalsWeak"
}
```

**预期结果**：
- `possibleAlarms` 包含：`"VitalsWeak"`

### 3.4 验证报警处理器逻辑

#### 测试用例 11：ShouldPublishAsAlarm - 启用报警
```go
// 测试代码位置：owlBack/wisefido-radar/internal/alarm/device_alarm_handler_test.go

func TestShouldPublishAsAlarm_EnabledAlarm(t *testing.T) {
    // 1. 准备：设备 test-device-001 已配置 Fall 报警启用
    // 2. 输入：event 数据，category="pose"
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "pose",
            "pose": "Fall",
        },
    }
    
    // 3. 调用
    shouldPublish, possibleAlarms, err := handler.ShouldPublishAsAlarm(
        ctx,
        "test-tenant-001",
        "test-device-001",
        "event",
        dataValue,
    )
    
    // 4. 验证
    //    - shouldPublish == true
    //    - possibleAlarms 包含 "Fall"
    //    - err == nil
}
```

**预期结果**：
- `shouldPublish` = `true`
- `possibleAlarms` 包含 `"Fall"`
- `err` = `nil`

#### 测试用例 12：ShouldPublishAsAlarm - 禁用报警
```go
func TestShouldPublishAsAlarm_DisabledAlarm(t *testing.T) {
    // 1. 准备：设备 test-device-002 已配置 Fall 报警禁用
    // 2. 输入：event 数据，category="pose"
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "pose",
            "pose": "Fall",
        },
    }
    
    // 3. 调用
    shouldPublish, possibleAlarms, err := handler.ShouldPublishAsAlarm(
        ctx,
        "test-tenant-001",
        "test-device-002",
        "event",
        dataValue,
    )
    
    // 4. 验证
    //    - shouldPublish == false
    //    - possibleAlarms 包含 "Fall"
    //    - err == nil
}
```

**预期结果**：
- `shouldPublish` = `false`
- `possibleAlarms` 包含 `"Fall"`
- `err` = `nil`

#### 测试用例 13：ShouldPublishAsAlarm - 多个可能报警类型，部分启用
```go
func TestShouldPublishAsAlarm_PartialEnabled(t *testing.T) {
    // 1. 准备：设备 test-device-001
    //    - Fall: enabled=true
    //    - SuspectedFall: enabled=false
    // 2. 输入：event 数据，category="pose"（可能触发 Fall 或 SuspectedFall）
    dataValue := []interface{}{
        map[string]interface{}{
            "category": "pose",
            "pose": "FallSuspected",
        },
    }
    
    // 3. 调用
    shouldPublish, possibleAlarms, err := handler.ShouldPublishAsAlarm(...)
    
    // 4. 验证
    //    - shouldPublish == true（因为 Fall 已启用）
    //    - possibleAlarms 包含 "Fall" 和 "SuspectedFall"
}
```

**预期结果**：
- `shouldPublish` = `true`（因为至少有一个报警类型已启用）
- `possibleAlarms` 包含 `"Fall"` 和 `"SuspectedFall"`

### 3.5 验证 Stream 分离逻辑

#### 测试用例 14：MQTT Consumer - Event 发布为 Alarm
```go
// 测试代码位置：owlBack/wisefido-radar/internal/consumer/mqtt_consumer_test.go
// 注意：这需要模拟 MQTT 消息和 Redis Stream 发布

func TestMQTTConsumer_EventPublishAsAlarm(t *testing.T) {
    // 1. 准备：
    //    - 设备 test-device-001，Fall 报警已启用
    //    - 模拟 MQTT 消息：topic="/prefix/event/productId/TEST_UID_001/post"
    //    - payload: {"cmd": "event", "type": 2, "data": [...]}
    
    // 2. 调用 handleMessage
    
    // 3. 验证：
    //    - 发布到 "iot:alarm:stream"（不是 "iot:event:stream"）
    //    - encodedData["topic_type"] == "alarm"
    //    - encodedData["device_id"] == "test-device-001"
    //    - encodedData["tenant_id"] == "test-tenant-001"
    //    - encodedData["branch_id"] == nil
    //    - encodedData["room_id"] == nil 或 字符串值
}
```

**预期结果**：
- 发布到 `"iot:alarm:stream"`
- `encodedData["topic_type"]` = `"alarm"`
- 所有元数据字段都设置（不存在则为 `null`）

#### 测试用例 15：MQTT Consumer - Event 发布为普通 Event
```go
func TestMQTTConsumer_EventPublishAsNormalEvent(t *testing.T) {
    // 1. 准备：
    //    - 设备 test-device-002，所有报警都禁用
    //    - 模拟 MQTT 消息：topic="/prefix/event/productId/TEST_UID_002/post"
    
    // 2. 调用 handleMessage
    
    // 3. 验证：
    //    - 发布到 "iot:event:stream"（不是 "iot:alarm:stream"）
    //    - encodedData["topic_type"] == "event"
}
```

**预期结果**：
- 发布到 `"iot:event:stream"`
- `encodedData["topic_type"]` = `"event"`

#### 测试用例 16：MQTT Consumer - Stat 发布为 Alarm
```go
func TestMQTTConsumer_StatPublishAsAlarm(t *testing.T) {
    // 1. 准备：
    //    - 设备 test-device-001，VitalsWeak 报警已启用
    //    - 模拟 MQTT 消息：topic="/prefix/stat/productId/TEST_UID_001/post"
    //    - payload: {"cmd": "stat", "data": {"sleep": {...}, "vital_signs_state": "Vital signs weak"}}
    
    // 2. 调用 handleMessage
    
    // 3. 验证：
    //    - 发布到 "iot:alarm:stream"
    //    - encodedData["topic_type"] == "alarm"
}
```

**预期结果**：
- 发布到 `"iot:alarm:stream"`
- `encodedData["topic_type"]` = `"alarm"`

#### 测试用例 17：MQTT Consumer - Monitor 不检查报警
```go
func TestMQTTConsumer_MonitorNoAlarmCheck(t *testing.T) {
    // 1. 准备：
    //    - 模拟 MQTT 消息：topic="/prefix/monitor/productId/TEST_UID_001/post"
    
    // 2. 调用 handleMessage
    
    // 3. 验证：
    //    - 发布到 "iot:monitor:stream"（不检查报警）
    //    - encodedData["topic_type"] == "monitor"
    //    - 不调用 alarmHandler.ShouldPublishAsAlarm
}
```

**预期结果**：
- 发布到 `"iot:monitor:stream"`
- `encodedData["topic_type"]` = `"monitor"`
- 不调用报警检查逻辑

### 3.6 验证元数据字段处理

#### 测试用例 18：元数据字段 null 值处理
```go
func TestMQTTConsumer_MetadataNullHandling(t *testing.T) {
    // 1. 准备：设备没有绑定信息（BoundBedID 和 BoundRoomID 为 nil）
    
    // 2. 调用 handleMessage
    
    // 3. 验证 encodedData 中的字段：
    //    - device_id: 字符串值（不为 null）
    //    - tenant_id: 字符串值（不为 null）
    //    - branch_id: null（不是空字符串）
    //    - building_id: null
    //    - unit_id: null
    //    - room_id: null（因为 BoundRoomID 为 nil）
    //    - bed_id: null（因为 BoundBedID 为 nil）
}
```

**预期结果**：
- `encodedData["device_id"]` = 字符串（不为 `null`）
- `encodedData["tenant_id"]` = 字符串（不为 `null`）
- `encodedData["branch_id"]` = `null`（不是空字符串 `""`）
- `encodedData["building_id"]` = `null`
- `encodedData["unit_id"]` = `null`
- `encodedData["room_id"]` = `null`（如果 `BoundRoomID` 为 `nil`）
- `encodedData["bed_id"]` = `null`（如果 `BoundBedID` 为 `nil`）

## 四、验证检查清单

### 4.1 代码逻辑检查

- [ ] `GetAlarmEnablement` 正确解析 `monitor_config.alarms` JSON
- [ ] `GetAlarmEnablement` 正确处理设备未配置的情况（返回空 map）
- [ ] `GetPossibleAlarmTypesFromEvent` 正确映射所有 event category
- [ ] `GetPossibleAlarmTypesFromStat` 正确映射所有 stat category 和状态字段
- [ ] `ShouldPublishAsAlarm` 正确判断是否应该发布为报警
- [ ] `ShouldPublishAsAlarm` 正确处理查询失败的情况（返回 false，不发布为报警）
- [ ] MQTT Consumer 正确调用 `ShouldPublishAsAlarm`（仅对 event 和 stat 类型）
- [ ] MQTT Consumer 正确设置 `topic_type`（"alarm" 或原类型）
- [ ] MQTT Consumer 正确选择输出 stream（`iot:alarm:stream` 或原 stream）
- [ ] 所有元数据字段都设置，不存在则为 `null`（不是空字符串）

### 4.2 边界情况检查

- [ ] 设备不存在时的处理（查询失败，不发布为报警）
- [ ] `monitor_config` JSON 格式错误时的处理（返回空 map）
- [ ] `alarms` 字段不存在时的处理（返回空 map）
- [ ] `enabled` 字段不存在时的处理（视为 `false`）
- [ ] `data_value` 为空数组时的处理（返回空报警类型列表）
- [ ] `data_value` 为对象（非数组）时的处理（转换为数组处理）

### 4.3 数据一致性检查

- [ ] 报警类型名称与 `RadarAlarmTypes` 列表一致
- [ ] 报警类型名称与 `alarm_device.monitor_config.alarms` 中的 key 一致
- [ ] 报警类型名称与前端 `AlarmCloud.vue` 中的 `ALARM_ITEM_ORDER.Radar` 一致
- [ ] Stream 名称与标准格式文档一致（`iot:alarm:stream`, `iot:event:stream`, `iot:stat:stream`）

## 五、运行验证

### 5.1 单元测试

```bash
cd /home/wisefido/owl-project/owlBack/wisefido-radar

# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/repository -v
go test ./internal/alarm -v
go test ./internal/consumer -v
```

### 5.2 集成测试

需要启动完整的服务进行端到端测试：

```bash
# 1. 启动 PostgreSQL 数据库
# 2. 启动 Redis
# 3. 准备测试数据（执行 2.1 中的 SQL）
# 4. 启动 wisefido-radar 服务
# 5. 发送模拟 MQTT 消息
# 6. 检查 Redis Stream 中的消息
```

### 5.3 手动验证步骤

1. **准备测试数据**：执行 2.1 中的 SQL 语句
2. **启动服务**：启动 `wisefido-radar` 服务
3. **发送 MQTT 消息**：使用 MQTT 客户端发送测试消息
4. **检查 Redis Stream**：
   ```bash
   # 检查 alarm stream
   redis-cli XREAD COUNT 10 STREAMS iot:alarm:stream 0
   
   # 检查 event stream
   redis-cli XREAD COUNT 10 STREAMS iot:event:stream 0
   
   # 检查 stat stream
   redis-cli XREAD COUNT 10 STREAMS iot:stat:stream 0
   ```
5. **验证消息格式**：检查消息中的 `topic_type` 和元数据字段

## 六、常见问题排查

### 6.1 报警未发布到 alarm stream

**可能原因**：
1. 报警类型未在 `alarm_device.monitor_config.alarms` 中配置
2. `enabled` 字段为 `false` 或不存在
3. `GetPossibleAlarmTypesFromEvent/Stat` 返回空列表
4. 数据库查询失败

**排查步骤**：
1. 检查数据库中的 `alarm_device` 配置
2. 检查日志中的 `possible_alarm_types`
3. 检查日志中的错误信息

### 6.2 元数据字段不是 null

**可能原因**：
1. `getStringOrNull` 或 `getStringOrNullPtr` 函数逻辑错误
2. 字段被设置为空字符串而不是 `nil`

**排查步骤**：
1. 检查 `encodedData` 中的字段值
2. 验证辅助函数的实现

### 6.3 报警类型映射不正确

**可能原因**：
1. `GetPossibleAlarmTypesFromEvent/Stat` 中的映射逻辑错误
2. `category` 或状态字段值不匹配

**排查步骤**：
1. 检查 `data_value` 中的 `category` 和状态字段
2. 验证映射函数中的 switch case 语句

## 七、验证报告模板

验证完成后，请填写以下报告：

```
## 验证报告

### 验证日期
[日期]

### 验证人员/AI
[名称]

### 验证结果

#### 1. 报警使能配置查询
- [ ] 测试用例 1：通过/失败
- [ ] 测试用例 2：通过/失败

#### 2. 事件到报警类型映射
- [ ] 测试用例 3-6：通过/失败

#### 3. 统计数据到报警类型映射
- [ ] 测试用例 7-10：通过/失败

#### 4. 报警处理器逻辑
- [ ] 测试用例 11-13：通过/失败

#### 5. Stream 分离逻辑
- [ ] 测试用例 14-17：通过/失败

#### 6. 元数据字段处理
- [ ] 测试用例 18：通过/失败

### 发现的问题
[列出发现的问题]

### 建议
[提出改进建议]
```

## 八、参考文档

- 计划文档：`/home/wisefido/.cursor/plans/radar_报警处理实施计划_c3db4bcd.plan.md`
- 标准格式文档：`owlBack/owl-common/encode/RADAR_REDIS_STREAM_FORMAT_STANDARD.md`
- 数据库表结构：`owlRD/db/21_alarm_device.sql`
- 前端报警配置：`owlFront/src/views/alarm/AlarmCloud.vue`
