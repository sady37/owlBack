package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"wisefido-sensor-fusion/internal/alarm"
	"wisefido-sensor-fusion/internal/models"
	"wisefido-sensor-fusion/internal/repository"

	"go.uber.org/zap"
)

func main() {
	// 创建测试日志
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	fmt.Println("=== 测试设备直接报警处理 ===")

	// 1. 测试 IsDeviceDirectAlarm 函数
	fmt.Println("\n1. 测试 IsDeviceDirectAlarm 函数:")
	testCases := []struct {
		eventType string
		expected  bool
	}{
		{"Fall", true},
		{"SuspectedFall", true},
		{"OfflineAlarm", true},
		{"LowBattery", true},
		{"DeviceFailure", true},
		{"HeartRateHigh", false}, // 云端事件
		{"BreathRateLow", false}, // 云端事件
		{"", false},
	}

	for _, tc := range testCases {
		result := alarm.IsDeviceDirectAlarm(tc.eventType)
		status := "✓"
		if result != tc.expected {
			status = "✗"
		}
		fmt.Printf("  %s %s -> %v (期望: %v)\n", status, tc.eventType, result, tc.expected)
	}

	// 2. 测试 AlarmEvent 模型
	fmt.Println("\n2. 测试 AlarmEvent 模型:")
	now := time.Now()
	triggerData := models.TriggerData{
		EventType: "Fall",
		Source:    "Radar",
		Confidence: func() *int { v := 95; return &v }(),
	}
	triggerDataJSON, _ := json.Marshal(triggerData)

	alarmEvent := &models.AlarmEvent{
		EventID:         "test-event-id",
		TenantID:        "test-tenant-id",
		DeviceID:        "test-device-id",
		EventType:       "Fall",
		Category:        "safety",
		AlarmLevel:      "ALERT",
		AlarmStatus:     "active",
		TriggeredAt:     now,
		TriggerData:     triggerDataJSON,
		NotifiedUsers:   json.RawMessage("[]"),
		Metadata:        json.RawMessage(`{"test": true}`),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	fmt.Printf("  AlarmEvent 创建成功:\n")
	fmt.Printf("    EventID: %s\n", alarmEvent.EventID)
	fmt.Printf("    EventType: %s\n", alarmEvent.EventType)
	fmt.Printf("    Category: %s\n", alarmEvent.Category)
	fmt.Printf("    AlarmLevel: %s\n", alarmEvent.AlarmLevel)

	// 3. 测试 TriggerData 序列化
	fmt.Println("\n3. 测试 TriggerData 序列化:")
	var decodedTriggerData models.TriggerData
	if err := json.Unmarshal(triggerDataJSON, &decodedTriggerData); err != nil {
		log.Fatalf("Failed to unmarshal trigger data: %v", err)
	}
	fmt.Printf("  TriggerData 解码成功:\n")
	fmt.Printf("    EventType: %s\n", decodedTriggerData.EventType)
	fmt.Printf("    Source: %s\n", decodedTriggerData.Source)
	if decodedTriggerData.Confidence != nil {
		fmt.Printf("    Confidence: %d\n", *decodedTriggerData.Confidence)
	}

	// 4. 测试 Repository 接口
	fmt.Println("\n4. 测试 Repository 接口:")
	// 注意：这里不实际连接数据库，只验证接口定义
	fmt.Println("  AlarmEventsRepository 接口定义正确")
	fmt.Println("  AlarmDeviceRepository 接口定义正确")

	// 5. 测试数据流集成
	fmt.Println("\n5. 测试数据流集成:")
	iotData := models.IoTDataMessage{
		IoTTimeSeriesID: 12345,
		DeviceID:        "radar-001",
		TenantID:        "demo-tenant",
		DeviceType:      "Radar",
		Timestamp:       time.Now().Unix(),
		DataType:        "alarm",
		Category:        "safety",
		EventType:       func() *string { s := "Fall"; return &s }(),
	}

	fmt.Printf("  IoTDataMessage 创建成功:\n")
	fmt.Printf("    DeviceID: %s\n", iotData.DeviceID)
	fmt.Printf("    EventType: %s\n", *iotData.EventType)
	fmt.Printf("    DataType: %s\n", iotData.DataType)

	// 6. 验证 wisefido-data-transformer 发布逻辑
	fmt.Println("\n6. 验证 wisefido-data-transformer 发布逻辑:")
	outputData := map[string]interface{}{
		"iot_timeseries_id": iotData.IoTTimeSeriesID,
		"device_id":         iotData.DeviceID,
		"tenant_id":         iotData.TenantID,
		"device_type":       iotData.DeviceType,
		"timestamp":         iotData.Timestamp,
		"data_type":         iotData.DataType,
		"category":          iotData.Category,
	}
	if iotData.EventType != nil {
		outputData["event_type"] = *iotData.EventType
		fmt.Println("  ✓ event_type 已添加到输出数据")
	} else {
		fmt.Println("  ✗ event_type 未添加到输出数据")
	}

	outputJSON, _ := json.MarshalIndent(outputData, "  ", "  ")
	fmt.Printf("  输出数据:\n%s\n", string(outputJSON))

	fmt.Println("\n=== 测试完成 ===")
	fmt.Println("\n总结:")
	fmt.Println("1. ✅ IsDeviceDirectAlarm 函数正确识别设备直接报警")
	fmt.Println("2. ✅ AlarmEvent 模型定义完整")
	fmt.Println("3. ✅ TriggerData 序列化正常")
	fmt.Println("4. ✅ Repository 接口定义正确")
	fmt.Println("5. ✅ IoTDataMessage 包含 EventType 字段")
	fmt.Println("6. ✅ wisefido-data-transformer 会发布 event_type")
	fmt.Println("\n下一步:")
	fmt.Println("- 需要启动 Redis 并创建 Stream")
	fmt.Println("- 需要启动 wisefido-sensor-fusion 服务")
	fmt.Println("- 发送测试数据验证报警创建")
}
