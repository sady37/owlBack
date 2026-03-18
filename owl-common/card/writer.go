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
	if status.DeviceStatus != nil {
		b, _ := json.Marshal(status.DeviceStatus)
		hsetArgs = append(hsetArgs, "device_status", string(b))
		changedFields = append(changedFields, "device_status")
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
	if status.DeviceStatus != nil {
		b, _ := json.Marshal(status.DeviceStatus)
		args = append(args, "device_status", string(b))
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

// SetDeviceOnline does read-modify-write on the device_status JSON in card:state Hash.
// 上线：置 Offline=0；下线：从 device_status 中删除该设备。
// Map key 仅用 deviceID（HIPAA）；不用 device_uid 兜底。
func (w *Writer) SetDeviceOnline(ctx context.Context, cardID, deviceID, deviceUID, deviceType string, online bool) error {
	if cardID == "" || deviceID == "" {
		return nil
	}
	key := deviceID
	hashKey := HashKey(cardID)

	raw, err := w.client.HGet(ctx, hashKey, "device_status").Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("read device_status: %w", err)
	}

	devices := make(map[string]*DeviceStatus)
	if raw != "" && raw != "{}" {
		if json.Unmarshal([]byte(raw), &devices) != nil {
			devices = make(map[string]*DeviceStatus)
		}
	}

	if !online {
		// 下线：删除该设备
		delete(devices, key)
		if deviceUID != "" {
			for k, d := range devices {
				if d != nil && d.DeviceUID == deviceUID {
					delete(devices, k)
					break
				}
			}
		}
	} else {
		// 上线：置 Offline=0
		ds := devices[key]
		if ds == nil {
			ds = devices[deviceUID]
			if ds != nil && deviceUID != "" {
				delete(devices, deviceUID)
				devices[key] = ds
			}
		}
		if ds == nil {
			ds = &DeviceStatus{DeviceUID: deviceUID, DeviceID: deviceID, DeviceType: deviceType}
			devices[key] = ds
		}
		if deviceID != "" {
			ds.DeviceID = deviceID
		}
		if ds.Offline == 0 {
			return nil
		}
		ds.Offline = 0
		ds.UpdatedAt = time.Now().UnixMilli()
	}

	b, err := json.Marshal(devices)
	if err != nil {
		return fmt.Errorf("marshal device_status: %w", err)
	}

	status := &CardStatus{
		CardID:       cardID,
		DeviceStatus: devices,
	}
	statusJSON, _ := json.Marshal(status)

	pipe := w.client.TxPipeline()
	pipe.HSet(ctx, hashKey, "device_status", string(b))
	pipe.XAdd(ctx, &redis.XAddArgs{
		Stream: StatusStreamName(),
		MaxLen: w.statusMaxLen,
		Approx: true,
		Values: map[string]interface{}{
			MsgTypeKey: MsgTypeEvent,
			"card_id": cardID,
			"fields":  "device_status",
			"data":    string(statusJSON),
		},
	})
	_, err = pipe.Exec(ctx)
	return err
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
