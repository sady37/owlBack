package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"time"

	rediscommon "owl-common/redis"
	"wisefido-radar/internal/config"
)

func main() {
	// 命令行参数
	var (
		source    = flag.String("source", "redis", "数据源: redis 或 db")
		stream    = flag.String("stream", "radar:monitor:stream", "Redis Stream 名称 (仅当 source=redis 时使用，支持 radar:*:stream 或 sleepace:*:stream)")
		count     = flag.Int64("count", 1, "读取的消息数量 (仅当 source=redis 时使用)")
		topicType = flag.String("topic-type", "", "topic_type (仅当 source=db 时使用，可选): monitor, stat, event, alarm (留空表示查询所有类型)")
		limit     = flag.Int("limit", 1, "读取的记录数量 (仅当 source=db 时使用)")
		decode    = flag.String("decode", "", "直接解码指定的 base64 track 字符串")
	)
	flag.Parse()

	ctx := context.Background()

	// 如果指定了 decode 参数，直接解码
	if *decode != "" {
		if err := decodeAndPrint(*decode); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	var radarDataList []RadarData

	switch *source {
	case "redis":
		// 从 Redis Streams 读取
		redisClient := rediscommon.NewRedisClient(&cfg.Redis)
		defer rediscommon.Close(redisClient)

		if err := rediscommon.Ping(ctx, redisClient); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to connect to Redis: %v\n", err)
			fmt.Fprintf(os.Stderr, "Hint: Please set REDIS_ADDR and REDIS_PASSWORD environment variables\n")
			fmt.Fprintf(os.Stderr, "Example: export REDIS_ADDR=127.0.0.1:6379\n")
			fmt.Fprintf(os.Stderr, "         export REDIS_PASSWORD=your_password\n")
			os.Exit(1)
		}

		radarDataList, err = ReadFromRedisStream(ctx, redisClient, *stream, *count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read from Redis Stream: %v\n", err)
			os.Exit(1)
		}

	case "db":
		// 从数据库读取
		dbRecords, err := ReadFromDatabase(ctx, &cfg.Database, *topicType, *limit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read from database: %v\n", err)
			os.Exit(1)
		}

		// 检查是否查询到记录
		if len(dbRecords) == 0 {
			filterMsg := ""
			if *topicType != "" {
				filterMsg = fmt.Sprintf(" for topic_type='%s'", *topicType)
			}
			fmt.Fprintf(os.Stderr, "No records found in database%s.\n", filterMsg)
			fmt.Fprintf(os.Stderr, "Possible reasons:\n")
			fmt.Fprintf(os.Stderr, "  1. Data has not been written to database yet\n")
			fmt.Fprintf(os.Stderr, "  2. iot-timeseries service may not be running\n")
			if *topicType != "" {
				fmt.Fprintf(os.Stderr, "  3. No data matches the topic_type filter\n")
			}
			os.Exit(1)
		}

		// 显示数据库记录详情（用于验证存储是否正常）
		filterMsg := ""
		if *topicType != "" {
			filterMsg = fmt.Sprintf(" (topic_type='%s')", *topicType)
		}
		fmt.Printf("=== Database Records%s ===\n", filterMsg)
		fmt.Printf("Found %d record(s)\n\n", len(dbRecords))
		
		for i, record := range dbRecords {
			fmt.Printf("--- Record %d ---\n", i+1)
			fmt.Printf("ID: %d\n", record.ID)
			fmt.Printf("Device ID: %s\n", record.DeviceID)
			fmt.Printf("Timestamp: %s\n", record.Timestamp.Format(time.RFC3339))
			fmt.Printf("Topic Type: %s\n", record.TopicType)
			fmt.Printf("Data Type: %s\n", record.DataType)
			
			// 显示标准字段
			if record.PositionX != nil {
				fmt.Printf("Position: X=%d cm, Y=%d cm, Z=%d cm\n", 
					*record.PositionX, *record.PositionY, *record.PositionZ)
			}
			if record.TrackingID != nil {
				fmt.Printf("Tracking ID: %d\n", *record.TrackingID)
			}
			if record.PostureDisplay != nil {
				fmt.Printf("Posture: %s (SNOMED: %s)\n", 
					*record.PostureDisplay, *record.PostureSNOMEDCode)
			}
			if record.EventDisplay != nil {
				fmt.Printf("Event: %s (SNOMED: %s)\n", 
					*record.EventDisplay, *record.EventSNOMEDCode)
			}
			if record.HeartRate != nil {
				fmt.Printf("Heart Rate: %d bpm\n", *record.HeartRate)
			}
			if record.RespiratoryRate != nil {
				fmt.Printf("Respiratory Rate: %d /min\n", *record.RespiratoryRate)
			}
			if record.SleepStateDisplay != nil {
				fmt.Printf("Sleep State: %s (SNOMED: %s)\n", 
					*record.SleepStateDisplay, *record.SleepStateSNOMEDCode)
			}
			
			// 显示 raw_original
			if record.RawOriginal != nil {
				fmt.Printf("Raw Original: %s\n", *record.RawOriginal)
			}
			
			// 显示 metadata（如果有）
			if record.Metadata != nil {
				fmt.Printf("Metadata: %s\n", *record.Metadata)
			}
			
			fmt.Println()
		}
		
		fmt.Printf("Note: Database stores converted standard values, not raw track data.\n")
		fmt.Printf("To decode track data, use: --source redis --stream radar:<topic_type>:stream\n")
		os.Exit(0)
		fmt.Printf("=== Database Records%s ===\n", filterMsg)
		fmt.Printf("Found %d record(s)\n\n", len(dbRecords))
		
		for i, record := range dbRecords {
			fmt.Printf("--- Record %d ---\n", i+1)
			fmt.Printf("ID: %d\n", record.ID)
			fmt.Printf("Device ID: %s\n", record.DeviceID)
			fmt.Printf("Timestamp: %s\n", record.Timestamp.Format(time.RFC3339))
			fmt.Printf("Topic Type: %s\n", record.TopicType)
			fmt.Printf("Data Type: %s\n", record.DataType)
			
			// 显示标准字段
			if record.PositionX != nil {
				fmt.Printf("Position: X=%d cm, Y=%d cm, Z=%d cm\n", 
					*record.PositionX, *record.PositionY, *record.PositionZ)
			}
			if record.TrackingID != nil {
				fmt.Printf("Tracking ID: %d\n", *record.TrackingID)
			}
			if record.PostureDisplay != nil {
				fmt.Printf("Posture: %s (SNOMED: %s)\n", 
					*record.PostureDisplay, *record.PostureSNOMEDCode)
			}
			if record.EventDisplay != nil {
				fmt.Printf("Event: %s (SNOMED: %s)\n", 
					*record.EventDisplay, *record.EventSNOMEDCode)
			}
			if record.HeartRate != nil {
				fmt.Printf("Heart Rate: %d bpm\n", *record.HeartRate)
			}
			if record.RespiratoryRate != nil {
				fmt.Printf("Respiratory Rate: %d /min\n", *record.RespiratoryRate)
			}
			if record.SleepStateDisplay != nil {
				fmt.Printf("Sleep State: %s (SNOMED: %s)\n", 
					*record.SleepStateDisplay, *record.SleepStateSNOMEDCode)
			}
			
			// 显示 raw_original
			if record.RawOriginal != nil {
				fmt.Printf("Raw Original: %s\n", *record.RawOriginal)
			}
			
			// 显示 metadata（如果有）
			if record.Metadata != nil {
				fmt.Printf("Metadata: %s\n", *record.Metadata)
			}
			
			fmt.Println()
		}
		
		fmt.Printf("Note: Database stores converted standard values, not raw track data.\n")
		fmt.Printf("To decode track data, use: --source redis --stream radar:<topic_type>:stream\n")
		os.Exit(0)

	default:
		fmt.Fprintf(os.Stderr, "Invalid source: %s (must be 'redis' or 'db')\n", *source)
		os.Exit(1)
	}

	// 注意：radarDataList 只在 source=redis 时使用
	// source=db 时已经在 case "db" 中处理并退出了
	if len(radarDataList) == 0 {
		fmt.Println("No data found")
		return
	}

	// 打印解码后的数据（仅用于 Redis Stream 数据）
	for i, radarData := range radarDataList {
		if i > 0 {
			fmt.Println()
		}
		printDecodedData(radarData)
	}
}

// decodeAndPrint 解码并打印指定的 base64 track 字符串
// 智能判断是 monitor track 还是 stat track
func decodeAndPrint(base64Track string) error {
	hexStr, err := FormatTrackBytes(base64Track)
	if err != nil {
		return err
	}

	fmt.Printf("=== Track Decoded ===\n")
	fmt.Printf("Track (Base64): %s\n", base64Track)
	fmt.Printf("Track (Hex): %s\n", hexStr)
	fmt.Println()

	// 解码字节数组以检查第一个字节
	trackBytes, err := base64.StdEncoding.DecodeString(base64Track)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(trackBytes) == 0 {
		return fmt.Errorf("empty track data")
	}

	firstByte := int(trackBytes[0])

	// 判断逻辑：
	// - stat track: 第一个字节通常是 1 或 2（版本号），且长度固定为 16 字节
	// - monitor track: 第一个字节是 0-7（人员编号）或 88（无人），长度是 16 的倍数
	if len(trackBytes) == 16 && (firstByte == 1 || firstByte == 2) {
		// 很可能是 stat track
		statTrackData, err := DecodeStatTrack(base64Track)
		if err == nil {
			fmt.Println("Decoded as Stat Track:")
			printStatTrackData(statTrackData)
			return nil
		}
	}

	// 尝试作为 monitor track 解码
	trackDataList, err := DecodeMonitorTrack(base64Track)
	if err == nil && len(trackDataList) > 0 {
		fmt.Println("Decoded as Monitor Track:")
		for i, trackData := range trackDataList {
			printTrackData(i, trackData)
		}
		return nil
	}

	// 如果 monitor track 解码失败，再次尝试 stat track
	statTrackData, err := DecodeStatTrack(base64Track)
	if err == nil {
		fmt.Println("Decoded as Stat Track:")
		printStatTrackData(statTrackData)
		return nil
	}

	return fmt.Errorf("failed to decode as both monitor and stat track")
}

// printDecodedData 打印解码后的数据
func printDecodedData(radarData RadarData) {
	fmt.Printf("=== %s Track Decoded ===\n", radarData.TopicType)
	fmt.Printf("Device ID: %s\n", radarData.DeviceID)
	fmt.Printf("Timestamp: %s\n", radarData.Timestamp.Format(time.RFC3339))
	fmt.Printf("Topic Type: %s\n", radarData.TopicType)
	fmt.Println()

	if radarData.Track == "" {
		fmt.Println("Warning: No track data found")
		return
	}

	hexStr, err := FormatTrackBytes(radarData.Track)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to format track bytes: %v\n", err)
	} else {
		fmt.Printf("Track (Base64): %s\n", radarData.Track)
		fmt.Printf("Track (Hex): %s\n", hexStr)
		fmt.Println()
	}

	// 根据 topic_type 选择不同的解码函数
	switch radarData.TopicType {
	case "monitor":
		// 解码 monitor track
		trackDataList, err := DecodeMonitorTrack(radarData.Track)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to decode monitor track: %v\n", err)
			return
		}
		for i, trackData := range trackDataList {
			printTrackData(i, trackData)
		}

	case "stat":
		// 解码 stat track
		statTrackData, err := DecodeStatTrack(radarData.Track)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to decode stat track: %v\n", err)
			return
		}
		printStatTrackData(statTrackData)

	case "event":
		fmt.Println("Event data does not contain track field")
		fmt.Println("Event data structure is different from monitor/stat")

	default:
		// 尝试作为 monitor track 解码
		trackDataList, err := DecodeMonitorTrack(radarData.Track)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to decode track: %v\n", err)
			return
		}
		for i, trackData := range trackDataList {
			printTrackData(i, trackData)
		}
	}
}

// printTrackData 打印单个 monitor track 数据
func printTrackData(index int, trackData TrackData) {
	personLabel := fmt.Sprintf("Person %d", index)
	if trackData.TargetID == 88 {
		personLabel = "No Person (target_id=88)"
	}

	fmt.Printf("%s:\n", personLabel)
	fmt.Printf("  target_id: %d", trackData.TargetID)
	if trackData.TargetID == 88 {
		fmt.Printf(" (无人)")
	}
	fmt.Println()

	// position_x 和 position_y 需要显示原始值和转换后的值
	posXOriginal := trackData.PositionX / 10
	posYOriginal := trackData.PositionY / 10
	fmt.Printf("  position_x: %d cm (原始值: %d dm)\n", trackData.PositionX, posXOriginal)
	fmt.Printf("  position_y: %d cm (原始值: %d dm)\n", trackData.PositionY, posYOriginal)
	fmt.Printf("  position_z: %d cm\n", trackData.PositionZ)
	fmt.Printf("  remaining_time: %d sec\n", trackData.RemainingTime)
	fmt.Printf("  pose: %d (%s)\n", trackData.Pose, GetPoseDisplay(trackData.Pose))
	fmt.Printf("  event: %d (%s)\n", trackData.Event, GetEventDisplay(trackData.Event))
	fmt.Printf("  area_id: %d\n", trackData.AreaID)
}

// printStatTrackData 打印 stat track 数据
func printStatTrackData(statTrackData *StatTrackData) {
	fmt.Printf("Statistics Track Data:\n")
	fmt.Printf("  version: %d\n", statTrackData.Version)
	fmt.Printf("  people_count: %d\n", statTrackData.PeopleCount)
	
	// 行走距离显示原始值（米）和转换后的值（厘米）
	walkDistanceM := statTrackData.WalkDistance / 100
	fmt.Printf("  walk_distance: %d cm (原始值: %d m)\n", statTrackData.WalkDistance, walkDistanceM)
	fmt.Printf("  walk_duration: %d sec\n", statTrackData.WalkDuration)
	fmt.Printf("  sit_duration: %d sec (未开放使用)\n", statTrackData.SitDuration)
	fmt.Printf("  lie_duration: %d sec\n", statTrackData.LieDuration)
	fmt.Printf("  stand_duration: %d sec\n", statTrackData.StandDuration)
	fmt.Printf("  multi_person_duration: %d sec\n", statTrackData.MultiPersonDuration)
}
