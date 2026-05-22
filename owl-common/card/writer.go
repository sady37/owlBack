package card

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type Writer struct {
	client         *redis.Client
	statusMaxLen   int64
	realtimeMaxLen int64
}

func NewWriter(client *redis.Client, statusMaxLen, realtimeMaxLen int64) *Writer {
	return &Writer{client: client, statusMaxLen: statusMaxLen, realtimeMaxLen: realtimeMaxLen}
}

// normalizeBedStateForWrite 写入前的规范化（2026-05-20 起 BedState 改 per-field ts 模式后
// 无需 normalize；保留空函数兼容 caller）。
func normalizeBedStateForWrite(bs *BedState) {
	// DurationSec 已删除 — 消费者用 now - BedStatusTs 自己算
}

// WriteCardStatus atomically writes non-nil CardStatus blocks to Hash
// and publishes the full CardStatus JSON to card:status:stream.
func (w *Writer) WriteCardStatus(ctx context.Context, status *CardStatus) error {
	if status == nil || status.CardID == "" {
		return nil
	}

	hashKey := HashKey(status.CardID)
	hsetArgs := make([]interface{}, 0, 8)
	var changedFields []string

	if status.Target != nil {
		b, _ := json.Marshal(status.Target)
		hsetArgs = append(hsetArgs, "target", string(b))
		changedFields = append(changedFields, "target")
	}
	if status.RoomState != nil {
		b, _ := json.Marshal(status.RoomState)
		hsetArgs = append(hsetArgs, "room_state", string(b))
		changedFields = append(changedFields, "room_state")
	}
	if status.BedState != nil {
		normalizeBedStateForWrite(status.BedState)
		b, _ := json.Marshal(status.BedState)
		hsetArgs = append(hsetArgs, "bed_state", string(b))
		changedFields = append(changedFields, "bed_state")
	}
	if status.AlarmState != nil {
		b, _ := json.Marshal(status.AlarmState)
		hsetArgs = append(hsetArgs, "alarm_state", string(b))
		changedFields = append(changedFields, "alarm_state")
	}
	if status.Display != nil {
		b, _ := json.Marshal(status.Display)
		hsetArgs = append(hsetArgs, "display", string(b))
		changedFields = append(changedFields, "display")
	}
	if status.Message != nil {
		b, _ := json.Marshal(status.Message)
		hsetArgs = append(hsetArgs, "message", string(b))
		changedFields = append(changedFields, "message")
	}

	if len(changedFields) == 0 {
		return nil
	}

	statusJSON, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("marshal card status: %w", err)
	}

	pipe := w.client.TxPipeline()
	pipe.HSet(ctx, hashKey, hsetArgs...)
	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: StatusStreamName(),
		MaxLen: w.statusMaxLen,
		Approx: true,
		Values: map[string]interface{}{
			MsgTypeKey:  MsgTypeEvent,
			"card_id":   status.CardID,
			"fields":    strings.Join(changedFields, ","),
			"data":      string(statusJSON),
		},
	})

	_, err = pipe.Exec(ctx)
	return err
}

// WriteCardStatusSilent writes non-nil CardStatus blocks to Hash without stream notification.
func (w *Writer) WriteCardStatusSilent(ctx context.Context, status *CardStatus) error {
	if status == nil || status.CardID == "" {
		return nil
	}

	hashKey := HashKey(status.CardID)
	args := make([]interface{}, 0, 8)

	if status.Target != nil {
		b, _ := json.Marshal(status.Target)
		args = append(args, "target", string(b))
	}
	if status.RoomState != nil {
		b, _ := json.Marshal(status.RoomState)
		args = append(args, "room_state", string(b))
	}
	if status.BedState != nil {
		normalizeBedStateForWrite(status.BedState)
		b, _ := json.Marshal(status.BedState)
		args = append(args, "bed_state", string(b))
	}
	if status.AlarmState != nil {
		b, _ := json.Marshal(status.AlarmState)
		args = append(args, "alarm_state", string(b))
	}
	if status.Display != nil {
		b, _ := json.Marshal(status.Display)
		args = append(args, "display", string(b))
	}
	if status.Message != nil {
		b, _ := json.Marshal(status.Message)
		args = append(args, "message", string(b))
	}

	if len(args) == 0 {
		return nil
	}
	return w.client.HSet(ctx, hashKey, args...).Err()
}

// PatchDeviceStatus 增量更新 device:status:{ipv6} Hash 的部分字段（仅写传入的非空 string/int 字段）。
// deviceAddr = /128 IPv6 canonical host text。fields key 为 Hash 字段名（"offline" / "signal_poor" / "last_seen_ms" 等）。
func (w *Writer) PatchDeviceStatus(ctx context.Context, deviceAddr string, fields map[string]interface{}) error {
	if deviceAddr == "" || len(fields) == 0 {
		return nil
	}
	args := make([]interface{}, 0, len(fields)*2+2)
	args = append(args, "device_addr", deviceAddr)
	for k, v := range fields {
		args = append(args, k, v)
	}
	if _, ok := fields["updated_at"]; !ok {
		args = append(args, "updated_at", fmt.Sprintf("%d", time.Now().UnixMilli()))
	}
	return w.client.HSet(ctx, DeviceStatusHashKey(deviceAddr), args...).Err()
}

// SetDeviceOnline 改写为 device:status:{ipv6} Hash 的部分更新。
// 上线：offline=0 + last_seen_ms=now；下线：offline=1。
func (w *Writer) SetDeviceOnline(ctx context.Context, deviceAddr, deviceUID, deviceType string, online bool) error {
	if deviceAddr == "" {
		return nil
	}
	now := time.Now().UnixMilli()
	fields := map[string]interface{}{
		"device_uid":  deviceUID,
		"device_type": deviceType,
		"updated_at":  fmt.Sprintf("%d", now),
	}
	if online {
		fields["offline"] = "0"
		fields["last_seen_ms"] = fmt.Sprintf("%d", now)
	} else {
		fields["offline"] = "1"
	}
	return w.PatchDeviceStatus(ctx, deviceAddr, fields)
}

// DeleteDeviceStatus 删除 device:status:{ipv6} Hash（设备从 cards 解绑/卸载时调用）。
func (w *Writer) DeleteDeviceStatus(ctx context.Context, deviceAddr string) error {
	if deviceAddr == "" {
		return nil
	}
	return w.client.Del(ctx, DeviceStatusHashKey(deviceAddr)).Err()
}

// WriteAI writes AI-derived flat fields to Hash with TTL timestamps, then publishes notification.
func (w *Writer) WriteAI(ctx context.Context, cardID string, fields map[string]string) error {
	if len(fields) == 0 || cardID == "" {
		return nil
	}

	hashKey := HashKey(cardID)
	nowMs := fmt.Sprintf("%d", time.Now().UnixMilli())

	hsetArgs := make([]interface{}, 0, len(fields)*4)
	var fieldNames []string

	for k, v := range fields {
		sf := StateFieldLookup(k)
		if sf == nil || sf.Owner != OwnerAI {
			continue
		}
		hsetArgs = append(hsetArgs, k, v)
		fieldNames = append(fieldNames, k)
		if sf.HasTTL {
			hsetArgs = append(hsetArgs, TsKey(k), nowMs)
		}
	}

	if len(fieldNames) == 0 {
		return nil
	}

	pipe := w.client.TxPipeline()
	pipe.HSet(ctx, hashKey, hsetArgs...)
	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: StatusStreamName(),
		MaxLen: w.statusMaxLen,
		Approx: true,
		Values: map[string]interface{}{
			MsgTypeKey: MsgTypeEvent,
			"card_id": cardID,
			"fields":  strings.Join(fieldNames, ","),
		},
	})

	_, err := pipe.Exec(ctx)
	return err
}

// DeleteCardState removes the entire card:state:{card_id} Hash.
func (w *Writer) DeleteCardState(ctx context.Context, cardID string) error {
	if cardID == "" {
		return nil
	}
	return w.client.Del(ctx, HashKey(cardID)).Err()
}

// PublishMonitor sends ephemeral device data to card:realtime:stream (no Hash write).
//
// data 内自动注入 "ts_ms" = 当前 server unix ms（2026-05-17 增强）：
//   - 该值同时承载 "per-card 最新活跃时刻" 与 "server 时钟锚" 双语义
//   - 前端用此值更新 lastRealtimeTs (per-card) + applyServerClock (全局时钟同步)，省去独立对时心跳
//   - 1Hz monitor 流自带高频心跳，drift 几乎瞬时纠偏
// FE handleRealtimeData 已存在 "非 device key 过滤"逻辑（card_id/timestamp），ts_ms 同样被排除出 devices 迭代。
func (w *Writer) PublishMonitor(ctx context.Context, cardID, deviceUID string, data map[string]any) error {
	if len(data) == 0 || cardID == "" {
		return nil
	}

	// 注入 server timestamp（先 clone 避免污染调用方传入的 data 引用）
	enriched := make(map[string]any, len(data)+1)
	for k, v := range data {
		enriched[k] = v
	}
	enriched["ts_ms"] = time.Now().UnixMilli()

	dataJSON, err := json.Marshal(enriched)
	if err != nil {
		return fmt.Errorf("marshal monitor data: %w", err)
	}

	return w.client.XAdd(ctx, &redis.XAddArgs{
		Stream: RealtimeStreamName(),
		MaxLen: w.realtimeMaxLen,
		Approx: true,
		Values: map[string]interface{}{
			MsgTypeKey:   MsgTypeMonitor,
			"card_id":    cardID,
			"device_uid": deviceUID,
			"data":       string(dataJSON),
		},
	}).Err()
}
