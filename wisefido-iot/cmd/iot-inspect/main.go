package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	commonconfig "owl-common/config"
	database "owl-common/database"
	rediscommon "owl-common/redis"
	"wisefido-iot/internal/config"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

func main() {
	var (
		source    = flag.String("source", "", "数据源: redis, db, config (必选)")
		stream    = flag.String("stream", "", "Redis Stream 名称 (仅当 source=redis 时使用): iot:monitor:stream, iot:stat:stream, iot:event:stream, iot:alarm:stream")
		count     = flag.Int64("count", 1, "读取的消息数量 (仅当 source=redis 时使用)")
		topicType = flag.String("topic-type", "", "topic_type (仅当 source=db 时使用，可选): monitor, stat, event, alarm (留空表示查询所有类型)")
		deviceID  = flag.String("device-id", "", "设备ID (仅当 source=db 时使用，可选)")
		limit     = flag.Int("limit", 10, "读取的记录数量 (仅当 source=db 时使用)")
		format    = flag.String("format", "pretty", "输出格式: pretty, json (仅当 source=db 时使用)")
	)
	flag.Parse()

	if *source == "" {
		fmt.Fprintf(os.Stderr, "Error: --source is required (redis, db, or config)\n")
		flag.Usage()
		os.Exit(1)
	}

	ctx := context.Background()

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	switch *source {
	case "redis":
		if *stream == "" {
			fmt.Fprintf(os.Stderr, "Error: --stream is required when source=redis\n")
			fmt.Fprintf(os.Stderr, "Available streams: iot:monitor:stream, iot:stat:stream, iot:event:stream, iot:alarm:stream\n")
			os.Exit(1)
		}

		redisClient := rediscommon.NewRedisClient(&cfg.Redis)
		defer rediscommon.Close(redisClient)

		if err := rediscommon.Ping(ctx, redisClient); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to Redis: %v\n", err)
			fmt.Fprintf(os.Stderr, "Hint: Please set REDIS_ADDR and REDIS_PASSWORD environment variables\n")
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
		dbRecords, err := ReadFromDatabase(ctx, &cfg.Database, *topicType, *deviceID, *limit)
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
			os.Exit(1)
		}

		if *format == "json" {
			printDBRecordsJSON(dbRecords)
		} else {
			printDBRecords(dbRecords)
		}

	case "config":
		printConfig(cfg)

	default:
		fmt.Fprintf(os.Stderr, "Invalid source: %s (must be 'redis', 'db', or 'config')\n", *source)
		os.Exit(1)
	}
}

// StreamMessage Redis Stream 消息
type StreamMessage struct {
	ID        string
	Timestamp time.Time
	Data      map[string]interface{}
}

// DatabaseRecord 数据库记录（与 owlRD/db/18_iot_timeseries.sql 一致）
type DatabaseRecord struct {
	ID           int64
	TenantID     string
	DeviceID     string
	DeviceUID    *string
	TimestampMs  int64
	TopicType    *string
	Category     *string
	DataValue    interface{}
	BranchName   *string
	BuildingName *string
	UnitName     *string
	RoomName     *string
	BedName      *string
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
		dataStr, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			continue
		}

		var timestamp time.Time
		if ts, ok := data["timestamp"]; ok {
			switch v := ts.(type) {
			case int64:
				timestamp = time.Unix(v, 0)
			case float64:
				timestamp = time.Unix(int64(v), 0)
			case string:
				if t, err := time.Parse(time.RFC3339, v); err == nil {
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

	query = `
		SELECT 
			id,
			tenant_id::text,
			device_id,
			device_uid,
			"timestamp",
			topic_type,
			category,
			data_value,
			branch_name,
			building_name,
			unit_name,
			room_name,
			bed_name
		FROM iot_timeseries
		WHERE 1=1
	`

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

	query += ` ORDER BY "timestamp" DESC`
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
		var tenantID, deviceID sql.NullString
		var deviceUID, topicTypeVal, category sql.NullString
		var dataValueJSON []byte
		var branchName, buildingName, unitName, roomName, bedName sql.NullString

		if err := rows.Scan(
			&record.ID,
			&tenantID,
			&deviceID,
			&deviceUID,
			&record.TimestampMs,
			&topicTypeVal,
			&category,
			&dataValueJSON,
			&branchName,
			&buildingName,
			&unitName,
			&roomName,
			&bedName,
		); err != nil {
			continue
		}

		if tenantID.Valid {
			record.TenantID = tenantID.String
		}
		if deviceID.Valid {
			record.DeviceID = deviceID.String
		}
		if deviceUID.Valid {
			s := deviceUID.String
			record.DeviceUID = &s
		}
		if topicTypeVal.Valid {
			s := topicTypeVal.String
			record.TopicType = &s
		}
		if category.Valid {
			s := category.String
			record.Category = &s
		}
		if branchName.Valid {
			s := branchName.String
			record.BranchName = &s
		}
		if buildingName.Valid {
			s := buildingName.String
			record.BuildingName = &s
		}
		if unitName.Valid {
			s := unitName.String
			record.UnitName = &s
		}
		if roomName.Valid {
			s := roomName.String
			record.RoomName = &s
		}
		if bedName.Valid {
			s := bedName.String
			record.BedName = &s
		}

		if len(dataValueJSON) > 0 {
			_ = json.Unmarshal(dataValueJSON, &record.DataValue)
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
	if topicType, ok := msg.Data["topic_type"].(string); ok {
		fmt.Printf("Topic Type: %s\n", topicType)
	}
	if category, ok := msg.Data["category"].(string); ok {
		fmt.Printf("Category: %s\n", category)
	}
	fmt.Println()

	fmt.Println("Data Values (JSON):")
	dataJSON, _ := json.MarshalIndent(msg.Data, "", "  ")
	fmt.Println(string(dataJSON))
}

// printDBRecords 打印数据库记录（pretty format）
func printDBRecords(records []DatabaseRecord) {
	fmt.Printf("=== Database Records ===\n")
	fmt.Printf("Found %d record(s)\n\n", len(records))

	for i, record := range records {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("--- Record %d ---\n", i+1)
		fmt.Printf("ID: %d\n", record.ID)
		if record.TenantID != "" {
			fmt.Printf("Tenant ID: %s\n", record.TenantID)
		}
		fmt.Printf("Device ID: %s\n", record.DeviceID)
		if record.DeviceUID != nil {
			fmt.Printf("Device UID: %s\n", *record.DeviceUID)
		}
		fmt.Printf("Timestamp (ms): %d\n", record.TimestampMs)
		if record.TopicType != nil {
			fmt.Printf("Topic Type: %s\n", *record.TopicType)
		}
		if record.Category != nil {
			fmt.Printf("Category: %s\n", *record.Category)
		}
		if record.BranchName != nil {
			fmt.Printf("Branch name: %s\n", *record.BranchName)
		}
		if record.BuildingName != nil {
			fmt.Printf("Building name: %s\n", *record.BuildingName)
		}
		if record.UnitName != nil {
			fmt.Printf("Unit name: %s\n", *record.UnitName)
		}
		if record.RoomName != nil {
			fmt.Printf("Room name: %s\n", *record.RoomName)
		}
		if record.BedName != nil {
			fmt.Printf("Bed name: %s\n", *record.BedName)
		}
		fmt.Println()

		fmt.Println("data_value (JSON):")
		dataJSON, _ := json.MarshalIndent(record.DataValue, "", "  ")
		fmt.Println(string(dataJSON))
	}
}

// printDBRecordsJSON 打印数据库记录（JSON format）
func printDBRecordsJSON(records []DatabaseRecord) {
	output := map[string]interface{}{
		"count":   len(records),
		"records": records,
	}
	jsonData, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(jsonData))
}

// printConfig 打印配置信息
func printConfig(cfg *config.Config) {
	fmt.Println("=== IoT Service Configuration ===")
	fmt.Println()

	fmt.Println("Database:")
	fmt.Printf("  Host: %s\n", cfg.Database.Host)
	fmt.Printf("  Port: %d\n", cfg.Database.Port)
	fmt.Printf("  User: %s\n", cfg.Database.User)
	fmt.Printf("  Database: %s\n", cfg.Database.Database)
	fmt.Printf("  SSL Mode: %s\n", cfg.Database.SSLMode)
	fmt.Println()

	fmt.Println("Redis:")
	fmt.Printf("  Addr: %s\n", cfg.Redis.Addr)
	fmt.Printf("  DB: %d\n", cfg.Redis.DB)
	fmt.Println()

	fmt.Println("Streams:")
	fmt.Printf("  Monitor: %s\n", cfg.Streams.Monitor)
	fmt.Printf("  Stat: %s\n", cfg.Streams.Stat)
	fmt.Printf("  Event: %s\n", cfg.Streams.Event)
	fmt.Printf("  Alarm: %s\n", cfg.Streams.Alarm)
	fmt.Println()

	fmt.Println("Consumer:")
	fmt.Printf("  Group: %s\n", cfg.ConsumerGroup)
	fmt.Printf("  Name: %s\n", cfg.ConsumerName)
	fmt.Printf("  Batch Size: %d\n", cfg.BatchSize)
	fmt.Println()

	fmt.Println("Log:")
	fmt.Printf("  Level: %s\n", cfg.Log.Level)
	fmt.Printf("  Format: %s\n", cfg.Log.Format)
}
