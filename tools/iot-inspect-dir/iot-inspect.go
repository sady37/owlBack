package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	commonconfig "owl-common/config"
	database "owl-common/database"
	rediscommon "owl-common/redis"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

func main() {
	var (
		source    = flag.String("source", "redis", "数据源: redis 或 db")
		stream    = flag.String("stream", "", "Redis Stream 名称 (仅当 source=redis 时使用，支持 iot:*:stream，如果未指定则根据 topic-type 自动设置)")
		count     = flag.Int64("count", 1, "读取的消息数量 (仅当 source=redis 时使用)")
		topicType = flag.String("topic-type", "", "topic_type: monitor, stat, event, alarm (当 source=redis 时自动设置 stream，当 source=db 时作为过滤条件)")
		deviceID  = flag.String("device-id", "", "设备ID (仅当 source=db 时使用，可选)")
		limit     = flag.Int("limit", 1, "读取的记录数量 (仅当 source=db 时使用)")
	)
	flag.Parse()

	ctx := context.Background()

	// 如果 source=redis 且指定了 topic-type，自动设置 stream 名称
	if *source == "redis" && *topicType != "" && *stream == "" {
		streamMap := map[string]string{
			"monitor": "iot:monitor:stream",
			"stat":    "iot:stat:stream",
			"event":   "iot:event:stream",
			"alarm":   "iot:alarm:stream",
		}
		if streamName, ok := streamMap[*topicType]; ok {
			*stream = streamName
		} else {
			fmt.Fprintf(os.Stderr, "Invalid topic-type: %s (must be: monitor, stat, event, alarm)\n", *topicType)
			os.Exit(1)
		}
	}

	// 如果 source=redis 且 stream 仍未设置，使用默认值
	if *source == "redis" && *stream == "" {
		*stream = "iot:monitor:stream"
	}

	// 加载配置（从环境变量，与 wisefido-iot 服务配置一致）
	dbConfig := &commonconfig.DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5433),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		Database: getEnv("DB_NAME", "owlrd"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	redisConfig := &commonconfig.RedisConfig{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", "TeLunSu-36kr"), // 默认密码，可通过环境变量覆盖
		DB:       0,
	}

	switch *source {
	case "redis":
		redisClient := rediscommon.NewRedisClient(redisConfig)
		defer rediscommon.Close(redisClient)

		if err := rediscommon.Ping(ctx, redisClient); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to Redis: %v\n", err)
			fmt.Fprintf(os.Stderr, "Hint: Please set REDIS_ADDR and REDIS_PASSWORD environment variables\n")
			fmt.Fprintf(os.Stderr, "Example: export REDIS_ADDR=127.0.0.1:6379\n")
			fmt.Fprintf(os.Stderr, "         export REDIS_PASSWORD=TeLunSu-36kr\n")
			os.Exit(1)
		}

		messages, err := ReadFromRedisStream(ctx, redisClient, *stream, *count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read from Redis Stream: %v\n", err)
			os.Exit(1)
		}

		if len(messages) == 0 {
			fmt.Println("No messages found in stream")
			return
		}

		for i, msg := range messages {
			if i > 0 {
				fmt.Println()
			}
			printStreamMessage(msg, *stream)
		}

	case "db":
		dbRecords, err := ReadFromDatabase(ctx, dbConfig, *topicType, *deviceID, *limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read from database: %v\n", err)
			os.Exit(1)
		}

		if len(dbRecords) == 0 {
			filterMsg := ""
			if *topicType != "" {
				filterMsg += fmt.Sprintf(" topic_type='%s'", *topicType)
			}
			if *deviceID != "" {
				filterMsg += fmt.Sprintf(" device_id='%s'", *deviceID)
			}
			if filterMsg != "" {
				filterMsg = " with filter" + filterMsg
			}
			fmt.Fprintf(os.Stderr, "No records found in database%s.\n", filterMsg)
			fmt.Fprintf(os.Stderr, "Possible reasons:\n")
			fmt.Fprintf(os.Stderr, "  1. Data has not been written to database yet\n")
			fmt.Fprintf(os.Stderr, "  2. wisefido-iot service may not be running\n")
			if *topicType != "" {
				fmt.Fprintf(os.Stderr, "  3. No data matches the topic_type filter\n")
				fmt.Fprintf(os.Stderr, "  4. Table must have data_values (JSONB); owlRD 18_iot_timeseries has no raw_original\n")
			}
			// 不退出，只是提示
			fmt.Fprintf(os.Stderr, "\nTip: Use --source redis to check if data is in Redis Stream\n")
			os.Exit(1)
		}

		printDBRecords(dbRecords)

	default:
		fmt.Fprintf(os.Stderr, "Invalid source: %s (must be 'redis' or 'db')\n", *source)
		os.Exit(1)
	}
}

// StreamMessage Redis Stream 消息
type StreamMessage struct {
	ID        string
	Timestamp time.Time
	Data      map[string]interface{}
}

// DatabaseRecord 数据库记录
type DatabaseRecord struct {
	ID         int64
	DeviceID   string
	DeviceUID  *string
	Timestamp  time.Time
	TopicType  *string
	Category   *string
	DataValues map[string]interface{}
	BranchID   *string
	BuildingID *string
	UnitID     *string
	RoomID     *string
	BedID      *string
}

// ReadFromRedisStream 从 Redis Stream 读取数据
func ReadFromRedisStream(ctx context.Context, redisClient *redis.Client, streamName string, count int64) ([]StreamMessage, error) {
	messages, err := redisClient.XRevRangeN(ctx, streamName, "+", "-", count).Result()
	if err != nil {
		if err == redis.Nil {
			return []StreamMessage{}, nil
		}
		return nil, fmt.Errorf("failed to read from stream %s: %w", streamName, err)
	}

	var results []StreamMessage
	for _, msg := range messages {
		// 将 Stream 消息的所有字段转换为 map
		data := make(map[string]interface{})
		for key, value := range msg.Values {
			// 处理字符串值
			if strValue, ok := value.(string); ok {
				// 如果值是 JSON 字符串（如 data_value），尝试解析
				if key == "data_value" {
					// 只有当 data_value 非空时才尝试解析 JSON
					if strValue != "" {
						var jsonData map[string]interface{}
						if err := json.Unmarshal([]byte(strValue), &jsonData); err == nil {
							// 将解析后的 JSON 数据合并到 data 中
							for k, v := range jsonData {
								data[k] = v
							}
						} else {
							// 如果解析失败，保留原始字符串
							data[key] = strValue
						}
					}
					// 如果 data_value 为空，不添加到 data 中（stat 心跳消息通常没有 data_value）
				} else {
					data[key] = strValue
				}
			} else {
				data[key] = value
			}
		}

		// 解析时间戳
		var timestamp time.Time
		if ts, ok := data["timestamp"]; ok {
			switch v := ts.(type) {
			case int64:
				timestamp = time.Unix(v, 0)
			case float64:
				timestamp = time.Unix(int64(v), 0)
			case string:
				// 尝试解析为数字字符串
				if tsInt, err := strconv.ParseInt(v, 10, 64); err == nil {
					timestamp = time.Unix(tsInt, 0)
				} else if t, err := time.Parse(time.RFC3339, v); err == nil {
					timestamp = t
				} else {
					timestamp = time.Now()
				}
			default:
				timestamp = time.Now()
			}
		} else {
			timestamp = time.Now()
		}

		results = append(results, StreamMessage{
			ID:        msg.ID,
			Timestamp: timestamp,
			Data:      data,
		})
	}

	return results, nil
}

// ReadFromDatabase 从数据库读取数据
func ReadFromDatabase(ctx context.Context, dbConfig *commonconfig.DatabaseConfig, topicType, deviceID string, limit int) ([]DatabaseRecord, error) {
	db, err := database.NewPostgresDB(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	var query string
	var args []interface{}
	argIndex := 1

	// 查询字段（与 owlRD 18_iot_timeseries、wisefido-iot 一致：data_values JSONB，无 raw_original）
	baseQuery := `
		SELECT 
			id,
			device_id,
			device_uid,
			timestamp,
			topic_type,
			category,
			data_values,
			branch_id,
			building_id,
			unit_id,
			room_id,
			bed_id
		FROM iot_timeseries
		WHERE 1=1
	`

	query = baseQuery

	if topicType != "" {
		query += fmt.Sprintf(" AND topic_type = $%d", argIndex)
		args = append(args, topicType)
		argIndex++
	}

	if deviceID != "" {
		query += fmt.Sprintf(" AND device_id = $%d", argIndex)
		args = append(args, deviceID)
		argIndex++
	}

	query += " ORDER BY timestamp DESC"
	query += fmt.Sprintf(" LIMIT $%d", argIndex)
	args = append(args, limit)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var results []DatabaseRecord
	for rows.Next() {
		var record DatabaseRecord
		var deviceID, deviceUID, topicTypeVal, categoryVal sql.NullString
		var dataValuesJSON []byte
		var branchID, buildingID, unitID, roomID, bedID sql.NullString

		if err := rows.Scan(
			&record.ID,
			&deviceID,
			&deviceUID,
			&record.Timestamp,
			&topicTypeVal,
			&categoryVal,
			&dataValuesJSON,
			&branchID,
			&buildingID,
			&unitID,
			&roomID,
			&bedID,
		); err != nil {
			continue
		}

		if deviceID.Valid {
			record.DeviceID = deviceID.String
		}
		if deviceUID.Valid {
			record.DeviceUID = &deviceUID.String
		}
		if topicTypeVal.Valid {
			record.TopicType = &topicTypeVal.String
		}
		if categoryVal.Valid {
			record.Category = &categoryVal.String
		}
		if branchID.Valid {
			record.BranchID = &branchID.String
		}
		if buildingID.Valid {
			record.BuildingID = &buildingID.String
		}
		if unitID.Valid {
			record.UnitID = &unitID.String
		}
		if roomID.Valid {
			record.RoomID = &roomID.String
		}
		if bedID.Valid {
			record.BedID = &bedID.String
		}

		if err := json.Unmarshal(dataValuesJSON, &record.DataValues); err != nil {
			record.DataValues = make(map[string]interface{})
		}

		results = append(results, record)
	}

	return results, nil
}

// printStreamMessage 打印 Stream 消息
func printStreamMessage(msg StreamMessage, streamName string) {
	fmt.Printf("=== Stream: %s ===\n", streamName)
	fmt.Printf("Message ID: %s\n", msg.ID)
	fmt.Printf("Timestamp: %s\n", msg.Timestamp.Format(time.RFC3339))
	fmt.Println()

	if deviceID, ok := msg.Data["device_id"].(string); ok {
		fmt.Printf("Device ID: %s\n", deviceID)
	}
	if deviceUID, ok := msg.Data["device_uid"].(string); ok {
		fmt.Printf("Device UID: %s\n", deviceUID)
	}
	if deviceType, ok := msg.Data["device_type"].(string); ok {
		fmt.Printf("Device Type: %s\n", deviceType)
	}
	if topicType, ok := msg.Data["topic_type"].(string); ok {
		fmt.Printf("Topic Type: %s\n", topicType)
	}
	var category string
	if cat, ok := msg.Data["category"].(string); ok {
		category = cat
		fmt.Printf("Category: %s\n", category)
	}
	fmt.Println()

	// 检查是否是空的心跳消息（stat 消息通常 data_value 为空或只有基础字段）
	var topicType string
	if tt, ok := msg.Data["topic_type"].(string); ok {
		topicType = tt
	}

	isEmptyHeartbeat := false
	if topicType == "stat" {
		// 检查 data 中是否只有基础字段，没有实际数据
		hasData := false
		for key := range msg.Data {
			if key != "device_id" && key != "device_uid" && key != "device_type" &&
				key != "tenant_id" && key != "timestamp" && key != "topic_type" &&
				key != "category" && key != "data_value" {
				hasData = true
				break
			}
		}
		// 如果 category 为空且没有其他数据，认为是心跳消息
		if category == "" && !hasData {
			isEmptyHeartbeat = true
		}
	}

	if isEmptyHeartbeat {
		fmt.Println("Data Values: (Empty heartbeat message - 60s keepalive)")
		fmt.Println("Note: Stat messages are used as keepalive, no actual data payload.")
	} else {
		fmt.Println("Data Values (JSON):")
		dataJSON, _ := json.MarshalIndent(msg.Data, "", "  ")
		fmt.Println(string(dataJSON))
	}
}

// printDBRecords 打印数据库记录
func printDBRecords(records []DatabaseRecord) {
	fmt.Printf("=== Database Records ===\n")
	filterMsg := ""
	if len(records) > 0 && records[0].TopicType != nil {
		filterMsg = fmt.Sprintf(" (topic_type='%s')", *records[0].TopicType)
	}
	fmt.Printf("Found %d record(s)%s\n\n", len(records), filterMsg)

	for i, record := range records {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("--- Record %d ---\n", i+1)
		fmt.Printf("ID: %d\n", record.ID)
		fmt.Printf("Device ID: %s\n", record.DeviceID)
		if record.DeviceUID != nil {
			fmt.Printf("Device UID: %s\n", *record.DeviceUID)
		}
		fmt.Printf("Timestamp: %s\n", record.Timestamp.Format(time.RFC3339))
		if record.TopicType != nil {
			fmt.Printf("Topic Type: %s\n", *record.TopicType)
		}
		if record.Category != nil {
			fmt.Printf("Category: %s\n", *record.Category)
		}
		if record.BranchID != nil {
			fmt.Printf("Branch ID: %s\n", *record.BranchID)
		}
		if record.BuildingID != nil {
			fmt.Printf("Building ID: %s\n", *record.BuildingID)
		}
		if record.UnitID != nil {
			fmt.Printf("Unit ID: %s\n", *record.UnitID)
		}
		if record.RoomID != nil {
			fmt.Printf("Room ID: %s\n", *record.RoomID)
		}
		if record.BedID != nil {
			fmt.Printf("Bed ID: %s\n", *record.BedID)
		}
		fmt.Println()

		fmt.Println("Data Values (JSON):")
		dataJSON, _ := json.MarshalIndent(record.DataValues, "", "  ")
		fmt.Println(string(dataJSON))
	}

	fmt.Println()
	fmt.Println("Note: Database stores converted standard values in data_values JSONB.")
	fmt.Println("To view raw stream data, use: --source redis --stream iot:<topic_type>:stream")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if port, err := strconv.Atoi(value); err == nil && port > 0 {
			return port
		}
	}
	return defaultValue
}
