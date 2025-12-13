package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// StreamMessage Redis Streams 消息
type StreamMessage struct {
	Stream   string
	ID       string
	Values   map[string]interface{}
}

// PublishToStream 发布消息到 Redis Streams
func PublishToStream(ctx context.Context, client *redis.Client, stream string, values map[string]interface{}) (string, error) {
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

	// 使用 XADD 命令添加消息
	id, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: streamValues,
	}).Result()
	
	return id, err
}

// PublishJSONToStream 发布 JSON 消息到 Redis Streams
func PublishJSONToStream(ctx context.Context, client *redis.Client, stream string, data interface{}) (string, error) {
	// 序列化为 JSON
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	// 发布到 Streams
	return PublishToStream(ctx, client, stream, map[string]interface{}{
		"data":      string(jsonBytes),
		"timestamp": time.Now().Unix(),
	})
}

// ReadFromStream 从 Redis Streams 读取消息
func ReadFromStream(ctx context.Context, client *redis.Client, stream string, consumerGroup string, consumer string, count int64) ([]StreamMessage, error) {
	// 使用 XREADGROUP 命令读取消息
	streams, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: consumer,
		Streams:  []string{stream, ">"},
		Count:    count,
		Block:    time.Second * 5, // 阻塞 5 秒
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

// CreateConsumerGroup 创建消费者组
func CreateConsumerGroup(ctx context.Context, client *redis.Client, stream string, groupName string) error {
	// 尝试创建消费者组，如果已存在则忽略错误
	// 注意：redis/v8 的 XGroupCreate 不支持 MkStream 参数
	// 如果 stream 不存在，先创建 stream（通过发送一条临时消息）
	err := client.XGroupCreate(ctx, stream, groupName, "0").Err()
	
	// 如果错误是 "BUSYGROUP"，说明组已存在，这是正常的
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		// 如果 stream 不存在，尝试创建（通过发送一条消息）
		if err.Error() == "NOGROUP" || err.Error() == "no such key" {
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

