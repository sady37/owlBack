package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// StreamMessage Redis Streams 消息
type StreamMessage struct {
	Stream string
	ID     string
	Values map[string]interface{}
}

// PublishToStream 发布消息到 Redis Streams
// maxLen: Stream 最大长度，超过此长度时自动删除最旧的消息。0 表示不限制长度
// retentionSeconds: 数据保留秒数，超过此秒数的消息会被自动删除。0 表示不限制时间（需要 Redis 6.2+）
func PublishToStream(ctx context.Context, client *redis.Client, stream string, values map[string]interface{}, maxLen int64, retentionSeconds int) (string, error) {
	// 将 values 转换为 Redis Streams 格式
	streamValues := make(map[string]interface{})
	for k, v := range values {
		// 将值转换为字符串
		var strValue string
		switch val := v.(type) {
		case string:
			strValue = val
		case []byte:
			strValue = string(val)
		case int:
			strValue = fmt.Sprintf("%d", val)
		case int32:
			strValue = fmt.Sprintf("%d", val)
		case int64:
			strValue = fmt.Sprintf("%d", val)
		case float32:
			strValue = fmt.Sprintf("%f", val)
		case float64:
			strValue = fmt.Sprintf("%f", val)
		case bool:
			if val {
				strValue = "true"
			} else {
				strValue = "false"
			}
		default:
			// 尝试 JSON 序列化
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			strValue = string(jsonBytes)
		}
		streamValues[k] = strValue
	}

	// 构建 XAddArgs
	args := &redis.XAddArgs{
		Stream: stream,
		Values: streamValues,
	}

	// 如果设置了 maxLen，限制 Stream 长度（使用 ~ 近似值，性能更好）
	if maxLen > 0 {
		args.MaxLen = maxLen
		args.Approx = true // 使用近似值，性能更好
	}

	// 如果设置了 retentionSeconds，基于时间清理旧消息（需要 Redis 6.2+）
	if retentionSeconds > 0 {
		// 计算最小 ID：当前时间 - 保留秒数（转换为毫秒时间戳）
		cutoffTime := time.Now().Add(-time.Duration(retentionSeconds) * time.Second)
		minID := fmt.Sprintf("%d-0", cutoffTime.UnixMilli())
		args.MinID = minID
		args.Approx = true // 使用近似值，性能更好
	}

	// 使用 XADD 命令添加消息
	id, err := client.XAdd(ctx, args).Result()

	return id, err
}

// PublishJSONToStream 发布 JSON 消息到 Redis Streams
// maxLen: Stream 最大长度，超过此长度时自动删除最旧的消息。0 表示不限制长度
// retentionSeconds: 数据保留秒数，超过此秒数的消息会被自动删除。0 表示不限制时间（需要 Redis 6.2+）
func PublishJSONToStream(ctx context.Context, client *redis.Client, stream string, data interface{}, maxLen int64, retentionSeconds int) (string, error) {
	// 序列化为 JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// 发布到 Streams
	return PublishToStream(ctx, client, stream, map[string]interface{}{
		"data":      string(jsonBytes),
		"timestamp": time.Now().Unix(),
	}, maxLen, retentionSeconds)
}

// ReadFromStream 从 Redis Streams 读取消息（默认 block 5 秒）
func ReadFromStream(ctx context.Context, client *redis.Client, stream string, consumerGroup string, consumer string, count int64) ([]StreamMessage, error) {
	return ReadFromStreamWithBlock(ctx, client, stream, consumerGroup, consumer, count, 5*time.Second)
}

// ReadFromStreamWithBlock 从 Redis Streams 读取消息，可指定 block 时长
func ReadFromStreamWithBlock(ctx context.Context, client *redis.Client, stream string, consumerGroup string, consumer string, count int64, block time.Duration) ([]StreamMessage, error) {
	streams, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    block,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return []StreamMessage{}, nil
		}
		return nil, err
	}

	var messages []StreamMessage
	for _, stream := range streams {
		for _, msg := range stream.Messages {
			messages = append(messages, StreamMessage{
				Stream: stream.Stream,
				ID:     msg.ID,
				Values: msg.Values,
			})
		}
	}

	return messages, nil
}

// ReadFromMultipleStreamsWithBlock 从多个 stream 读取消息（支持 XREADGROUP）
func ReadFromMultipleStreamsWithBlock(ctx context.Context, client *redis.Client, streamNames []string, consumerGroup string, consumer string, count int64, block time.Duration) ([]StreamMessage, error) {
	// 构建 Streams 参数：[stream1, stream2, ..., >, >, ...]
	// go-redis XReadGroup 要求先列所有 stream 名，再列对应的 ID
	streamArgs := make([]string, 0, len(streamNames)*2)
	for _, streamName := range streamNames {
		streamArgs = append(streamArgs, streamName)
	}
	for range streamNames {
		streamArgs = append(streamArgs, ">")
	}

	streams, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumer,
		Streams:  streamArgs,
		Count:    count,
		Block:    block,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return []StreamMessage{}, nil
		}
		return nil, err
	}

	var messages []StreamMessage
	for _, stream := range streams {
		for _, msg := range stream.Messages {
			messages = append(messages, StreamMessage{
				Stream: stream.Stream,
				ID:     msg.ID,
				Values: msg.Values,
			})
		}
	}

	return messages, nil
}

// AllStreams 返回系统中所有 stream 定义，供启动时批量操作使用。
func AllStreams() []StreamDefinition {
	return []StreamDefinition{
		StreamMonitor,
		StreamEvent,
		StreamAlarm,
		StreamAuth,
		StreamOther,
		StreamIoTCard,
		StreamConfigAlarmDevice,
		StreamConfigAlarmProcess,
		StreamConfigCard,
		StreamProbeDevice,
		StreamCardRealTime,
		StreamCardStatus,
		StreamCardUpdate,
	}
}

// TrimAllStreamsToNow 将所有 stream 裁剪到当前时间点，丢弃重启前的积压消息。
func TrimAllStreamsToNow(ctx context.Context, client *redis.Client) map[string]string {
	minID := fmt.Sprintf("%d-0", time.Now().UnixMilli())
	results := make(map[string]string, len(AllStreams()))

	for _, sd := range AllStreams() {
		trimmed, err := client.XTrimMinID(ctx, sd.Name, minID).Result()
		if err != nil {
			results[sd.Name] = fmt.Sprintf("error: %v", err)
		} else {
			results[sd.Name] = fmt.Sprintf("trimmed %d", trimmed)
		}
	}
	return results
}

// CreateConsumerGroup 创建消费者组
func CreateConsumerGroup(ctx context.Context, client *redis.Client, stream string, groupName string) error {
	// 尝试创建消费者组，如果已存在则忽略错误
	// 注意：redis/v8 的 XGroupCreate 不支持 MkStream 参数
	// 如果 stream 不存在，先创建 stream（通过发送一条临时消息）
	err := client.XGroupCreate(ctx, stream, groupName, "0").Err()

	// 如果错误是 "BUSYGROUP"，说明组已存在，这是正常的
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		// 检查是否是 Stream 不存在的错误（多种可能的错误信息）
		errStr := err.Error()
		isStreamNotExist := strings.Contains(errStr, "requires the key to exist") ||
			strings.Contains(errStr, "no such key") ||
			strings.Contains(errStr, "NOGROUP") ||
			strings.Contains(errStr, "ERR The XGROUP subcommand requires")

		if isStreamNotExist {
			// Stream 不存在，先创建一个临时消息来创建 stream
			msgID, createErr := client.XAdd(ctx, &redis.XAddArgs{
				Stream: stream,
				Values: map[string]interface{}{"init": "true"},
			}).Result()
			if createErr != nil {
				return fmt.Errorf("failed to create stream: %w", createErr)
			}
			// 删除临时消息
			client.XDel(ctx, stream, msgID)
			// 再次尝试创建消费者组
			err = client.XGroupCreate(ctx, stream, groupName, "0").Err()
			if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
