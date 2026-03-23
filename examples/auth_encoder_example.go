package main

import (
	"encoding/json"
	"fmt"
	"owl-common/encode"
)

// 示例：展示如何使用 auth_encoder 生成 Redis auth stream
func main() {
	// 示例 1: 编码认证请求
	fmt.Println("=== Example 1: Encode Auth Request ===")
	encodeAuthRequestExample()

	fmt.Println("\n=== Example 2: Encode Auth Response (Success) ===")
	encodeAuthResponseSuccessExample()

	fmt.Println("\n=== Example 3: Encode Auth Response (Failure) ===")
	encodeAuthResponseFailureExample()
}

// 编码认证请求示例
func encodeAuthRequestExample() {
	// 模拟从 HTTP 请求获取的数据
	deviceID := "550e8400-e29b-41d4-a716-446655440000" // 系统内 UUID
	deviceUID := "E598A2ACD523"                        // 设备序列号
	deviceType := "Radar"
	remoteAddr := "10.0.0.187:57087"

	// 构建设备信息
	deviceInfo := encode.BuildAuthRequestFromHTTPRequest(
		"E598A2ACD523",         // uid
		1,                      // device_type (1 = Radar)
		"2.0",                  // mcu_hw
		"Dec 17 2025 10:22:19", // mcu_sw
		"2.3",                  // radar_hw
		"Jun 25 2025 11:33:44", // radar_sw
		remoteAddr,
	)

	// 编码为 Redis Stream 事件
	authRequest := encode.EncodeAuthRequest(deviceID, deviceUID, deviceType, remoteAddr, deviceInfo)

	// 验证事件
	if err := encode.ValidateAuthStreamEvent(authRequest); err != nil {
		fmt.Printf("Validation Error: %v\n", err)
		return
	}

	// 输出 JSON 格式
	jsonData, _ := json.MarshalIndent(authRequest, "", "  ")
	fmt.Println(string(jsonData))

	// 输出 Redis Stream 命令示例
	fmt.Printf("\nRedis Stream Command:\n")
	fmt.Printf("XADD iot:auth:stream * ")
	fmt.Printf("device_id %s device_type %s tenant_id %s timestamp %d topic_type auth ",
		authRequest.DeviceID, authRequest.DeviceType, authRequest.TenantID, authRequest.Timestamp)

	// data_value 需要 JSON 序列化
	dataValueJSON, _ := json.Marshal(authRequest.DataValue)
	fmt.Printf("data_value '%s'\n", string(dataValueJSON))
}

// 编码认证成功响应示例
func encodeAuthResponseSuccessExample() {
	// 模拟从认证服务获取的数据
	deviceID := "550e8400-e29b-41d4-a716-446655440000"
	deviceUID := "E598A2ACD523"
	deviceType := "Radar"
	tenantID := "00000000-0000-0000-0000-000000000002" // 示例：Unallocated
	authStatus := "success"
	mqttServer := "10.0.0.100"
	mqttPort := 8883
	logInfo := "Device authenticated successfully"

	// 编码为 Redis Stream 事件
	authResponse := encode.EncodeAuthResponse(
		deviceID,
		deviceUID,
		deviceType,
		tenantID,
		authStatus,
		mqttServer,
		mqttPort,
		logInfo,
	)

	// 验证事件
	if err := encode.ValidateAuthStreamEvent(authResponse); err != nil {
		fmt.Printf("Validation Error: %v\n", err)
		return
	}

	// 输出 JSON 格式
	jsonData, _ := json.MarshalIndent(authResponse, "", "  ")
	fmt.Println(string(jsonData))
}

// 编码认证失败响应示例
func encodeAuthResponseFailureExample() {
	// 模拟认证失败的情况
	deviceID := "550e8400-e29b-41d4-a716-446655440000"
	deviceUID := "E598A2ACD523"
	deviceType := "Radar"
	tenantID := "00000000-0000-0000-0000-000000000002"
	authStatus := "failure"
	logInfo := "Device not found in device_store"

	// 编码为 Redis Stream 事件
	authResponse := encode.EncodeAuthResponse(
		deviceID,
		deviceUID,
		deviceType,
		tenantID,
		authStatus,
		"", // 失败时无 MQTT 服务器
		0,  // 失败时无 MQTT 端口
		logInfo,
	)

	// 验证事件
	if err := encode.ValidateAuthStreamEvent(authResponse); err != nil {
		fmt.Printf("Validation Error: %v\n", err)
		return
	}

	// 输出 JSON 格式
	jsonData, _ := json.MarshalIndent(authResponse, "", "  ")
	fmt.Println(string(jsonData))
}
