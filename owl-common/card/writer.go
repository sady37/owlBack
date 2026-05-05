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

// normalizeBedStateForWrite 写入前仅规范 DurationSec：<0 时置为 -1（未设置）。
func normalizeBedStateForWrite(bs *BedState) {
	if bs == nil {
		return
	}
	if bs.DurationSec < 0 {
		bs.DurationSec = BedStateDurationNotSet
	}
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
	if status.BathRoomState != nil {
		b, _ := json.Marshal(status.BathRoomState)
		hsetArgs = append(hsetArgs, "bathroom_state", string(b))
		changedFields = append(changedFields, "bathroom_state")
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
	if status.BathRoomState != nil {
		b, _ := json.Marshal(status.BathRoomState)
		args = append(args, "bathroom_state", string(b))
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
	if status.Message != nil {
		b, _ := json.Marshal(status.Message)
		args = append(args, "message", string(b))
	}

	if len(args) == 0 {
		return nil
	}
	return w.client.HSet(ctx, hashKey, args...).Err()
}

// WriteDeviceStatus 写连通性维度（offline / last_seen_ms / identity）到 device:status:{deviceID} 独立 Hash（Phase A）。
// 不发 stream（device-level 状态变化频率很低，由 cardagg 健康检测/告警自驱）。
//
// 故意**不**写 signal_poor / angle_abnormal / sensor_detached —— 这 3 个字段由 cardagg alarm_handler
// 在收到对应 alarm/recover 时通过 PatchDeviceFlag 单独维护。每轮 derive 调用 WriteDeviceStatus 时若全量
// HSet 这些字段会把刚刚 alarm_handler 写入的 1 重置回 0（因为 BuildDeviceStatus 不读 alarm flags 真值，
// struct 里都是 zero value），导致 device:status hash 始终看不到 degraded 标志。
func (w *Writer) WriteDeviceStatus(ctx context.Context, status *DeviceStatus) error {
	if status == nil || status.DeviceID == "" {
		return nil
	}
	hashKey := DeviceStatusHashKey(status.DeviceID)
	now := time.Now().UnixMilli()
	if status.UpdatedAt == 0 {
		status.UpdatedAt = now
	}
	args := []interface{}{
		"device_id", status.DeviceID,
		"device_uid", status.DeviceUID,
		"device_type", status.DeviceType,
		"updated_at", fmt.Sprintf("%d", status.UpdatedAt),
		"last_seen_ms", fmt.Sprintf("%d", status.LastSeenMs),
		"offline", fmt.Sprintf("%d", status.Offline),
	}
	return w.client.HSet(ctx, hashKey, args...).Err()
}

// PatchDeviceStatus 增量更新 device:status:{deviceID} Hash 的部分字段（仅写传入的非空 string/int 字段）。
// fields 中的 key 应为 Hash 字段名（"offline" / "signal_poor" / "last_seen_ms" 等）。
func (w *Writer) PatchDeviceStatus(ctx context.Context, deviceID string, fields map[string]interface{}) error {
	if deviceID == "" || len(fields) == 0 {
		return nil
	}
	args := make([]interface{}, 0, len(fields)*2+2)
	args = append(args, "device_id", deviceID)
	for k, v := range fields {
		args = append(args, k, v)
	}
	if _, ok := fields["updated_at"]; !ok {
		args = append(args, "updated_at", fmt.Sprintf("%d", time.Now().UnixMilli()))
	}
	return w.client.HSet(ctx, DeviceStatusHashKey(deviceID), args...).Err()
}

// SetDeviceOnline 改写为 device:status:{deviceID} Hash 的部分更新（Phase A）。
// 上线：offline=0 + last_seen_ms=now；下线：offline=1 + 保留 device_id 用于追踪。
func (w *Writer) SetDeviceOnline(ctx context.Context, deviceID, deviceUID, deviceType string, online bool) error {
	if deviceID == "" {
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
	return w.PatchDeviceStatus(ctx, deviceID, fields)
}

// DeleteDeviceStatus 删除 device:status:{deviceID} Hash（设备从 cards 解绑/卸载时调用）。
func (w *Writer) DeleteDeviceStatus(ctx context.Context, deviceID string) error {
	if deviceID == "" {
		return nil
	}
	return w.client.Del(ctx, DeviceStatusHashKey(deviceID)).Err()
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
func (w *Writer) PublishMonitor(ctx context.Context, cardID, deviceUID string, data map[string]any) error {
	if len(data) == 0 || cardID == "" {
		return nil
	}

	dataJSON, err := json.Marshal(data)
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
