package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"owl-common/config"
	database "owl-common/database"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

// RadarData 雷达数据（用于 Redis Stream）
type RadarData struct {
	DeviceID  string
	Timestamp time.Time
	TopicType string
	Track     string // base64 编码的 track 字段
}

// DatabaseRecord 数据库记录（用于数据库查询）
type DatabaseRecord struct {
	ID                   int64
	DeviceID             string
	Timestamp            time.Time
	TopicType            string
	DataType             string
	PositionX            *int
	PositionY            *int
	PositionZ            *int
	TrackingID           *int
	PostureSNOMEDCode    *string
	PostureDisplay       *string
	EventType            *string
	EventSNOMEDCode      *string
	EventDisplay         *string
	HeartRate            *int
	RespiratoryRate      *int
	SleepStateSNOMEDCode *string
	SleepStateDisplay    *string
	RawOriginal          *string
	Metadata             *string
}

// ReadFromRedisStream 从 Redis Streams 读取数据
func ReadFromRedisStream(ctx context.Context, redisClient *redis.Client, streamName string, count int64) ([]RadarData, error) {
	// 使用 XREVRANGE 读取最新的 count 条消息
	messages, err := redisClient.XRevRangeN(ctx, streamName, "+", "-", count).Result()
	if err != nil {
		if err == redis.Nil {
			return []RadarData{}, nil
		}
		return nil, fmt.Errorf("failed to read from stream %s: %w", streamName, err)
	}

	var results []RadarData
	for _, msg := range messages {
		// 从 message 的 Values 中提取 data 字段（JSON 字符串）
		dataStr, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		// 解析 JSON
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			continue
		}

		// 提取字段
		deviceID, _ := data["device_id"].(string)
		topicType, _ := data["topic_type"].(string)

		// 提取 timestamp
		var timestamp time.Time
		if ts, ok := data["timestamp"]; ok {
			switch v := ts.(type) {
			case int64:
				timestamp = time.Unix(v, 0)
			case float64:
				timestamp = time.Unix(int64(v), 0)
			default:
				timestamp = time.Now()
			}
		} else {
			timestamp = time.Now()
		}

		// 提取 track 字段（可能在 data.data.track 中）
		var track string
		if dataField, ok := data["data"]; ok {
			if dataMap, ok := dataField.(map[string]interface{}); ok {
				if trackVal, ok := dataMap["track"].(string); ok {
					track = trackVal
				}
			}
		}

		if track == "" {
			continue // 跳过没有 track 字段的数据
		}

		results = append(results, RadarData{
			DeviceID:  deviceID,
			Timestamp: timestamp,
			TopicType: topicType,
			Track:     track,
		})
	}

	return results, nil
}

// ReadFromDatabase 从数据库读取数据
func ReadFromDatabase(ctx context.Context, dbConfig *config.DatabaseConfig, topicType string, limit int) ([]DatabaseRecord, error) {
	// 连接数据库
	db, err := database.NewPostgresDB(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// 构建查询 SQL
	// 如果 topicType 为空，查询所有类型的记录
	var query string
	var args []interface{}

	if topicType != "" {
		// 有 topic_type 过滤条件
		query = `
			SELECT 
				id,
				device_id,
				timestamp,
				data_type,
				radar_pos_x,
				radar_pos_y,
				radar_pos_z,
				tracking_id,
				posture_snomed_code,
				posture_display,
				event_type,
				event_snomed_code,
				event_display,
				heart_rate,
				respiratory_rate,
				sleep_state_snomed_code,
				sleep_state_display,
				CASE 
					WHEN raw_original IS NOT NULL 
					THEN convert_from(raw_original, 'UTF8')
					ELSE NULL
				END as raw_original,
				CASE 
					WHEN metadata IS NOT NULL 
					THEN metadata::text
					ELSE NULL
				END as metadata,
				CASE 
					WHEN raw_original IS NOT NULL 
					THEN convert_from(raw_original, 'UTF8')::jsonb->>'topic_type'
					ELSE NULL
				END as topic_type
			FROM iot_timeseries
			WHERE (
				raw_original IS NOT NULL 
				AND convert_from(raw_original, 'UTF8')::jsonb->>'topic_type' = $1
			)
			ORDER BY timestamp DESC
			LIMIT $2
		`
		args = []interface{}{topicType, limit}
	} else {
		// 查询所有类型的记录
		query = `
			SELECT 
				id,
				device_id,
				timestamp,
				data_type,
				radar_pos_x,
				radar_pos_y,
				radar_pos_z,
				tracking_id,
				posture_snomed_code,
				posture_display,
				event_type,
				event_snomed_code,
				event_display,
				heart_rate,
				respiratory_rate,
				sleep_state_snomed_code,
				sleep_state_display,
				CASE 
					WHEN raw_original IS NOT NULL 
					THEN convert_from(raw_original, 'UTF8')
					ELSE NULL
				END as raw_original,
				CASE 
					WHEN metadata IS NOT NULL 
					THEN metadata::text
					ELSE NULL
				END as metadata,
				CASE 
					WHEN raw_original IS NOT NULL 
					THEN convert_from(raw_original, 'UTF8')::jsonb->>'topic_type'
					ELSE NULL
				END as topic_type
			FROM iot_timeseries
			ORDER BY timestamp DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	defer rows.Close()

	var results []DatabaseRecord
	for rows.Next() {
		var record DatabaseRecord
		var posX, posY, posZ, trackingID sql.NullInt64
		var postureCode, postureDisplay, eventType, eventCode, eventDisplay sql.NullString
		var hr, rr sql.NullInt64
		var sleepCode, sleepDisplay sql.NullString
		var rawOriginal, metadata sql.NullString
		var topicTypeFromDB sql.NullString

		if err := rows.Scan(
			&record.ID,
			&record.DeviceID,
			&record.Timestamp,
			&record.DataType,
			&posX, &posY, &posZ,
			&trackingID,
			&postureCode, &postureDisplay,
			&eventType, &eventCode, &eventDisplay,
			&hr, &rr,
			&sleepCode, &sleepDisplay,
			&rawOriginal,
			&metadata,
			&topicTypeFromDB,
		); err != nil {
			continue
		}

		// 填充字段
		if posX.Valid {
			x := int(posX.Int64)
			record.PositionX = &x
		}
		if posY.Valid {
			y := int(posY.Int64)
			record.PositionY = &y
		}
		if posZ.Valid {
			z := int(posZ.Int64)
			record.PositionZ = &z
		}
		if trackingID.Valid {
			tid := int(trackingID.Int64)
			record.TrackingID = &tid
		}
		if postureCode.Valid {
			record.PostureSNOMEDCode = &postureCode.String
		}
		if postureDisplay.Valid {
			record.PostureDisplay = &postureDisplay.String
		}
		if eventType.Valid {
			record.EventType = &eventType.String
		}
		if eventCode.Valid {
			record.EventSNOMEDCode = &eventCode.String
		}
		if eventDisplay.Valid {
			record.EventDisplay = &eventDisplay.String
		}
		if hr.Valid {
			h := int(hr.Int64)
			record.HeartRate = &h
		}
		if rr.Valid {
			r := int(rr.Int64)
			record.RespiratoryRate = &r
		}
		if sleepCode.Valid {
			record.SleepStateSNOMEDCode = &sleepCode.String
		}
		if sleepDisplay.Valid {
			record.SleepStateDisplay = &sleepDisplay.String
		}
		if rawOriginal.Valid {
			record.RawOriginal = &rawOriginal.String
		}
		if metadata.Valid {
			record.Metadata = &metadata.String
		}
		if topicTypeFromDB.Valid {
			record.TopicType = topicTypeFromDB.String
		}

		results = append(results, record)
	}

	return results, nil
}
