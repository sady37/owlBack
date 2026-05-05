package card

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

type Reader struct {
	client *redis.Client
}

func NewReader(client *redis.Client) *Reader {
	return &Reader{client: client}
}

func (r *Reader) ReadCardStatus(ctx context.Context, cardID string) (*CardStatus, error) {
	// Phase A：device_status 已迁出 card:state，独立到 device:status:{deviceID}（详 ReadDeviceStatusByDeviceIDs）。
	vals, err := r.client.HMGet(ctx, HashKey(cardID),
		"target", "room_state", "bathroom_state", "bed_state", "alarm_state", "message",
	).Result()
	if err != nil {
		return nil, err
	}
	status := &CardStatus{CardID: cardID}
	if s, ok := vals[0].(string); ok && s != "" && s != "{}" {
		var t TargetState
		if json.Unmarshal([]byte(s), &t) == nil {
			status.Target = &t
		}
	}
	if s, ok := vals[1].(string); ok && s != "" && s != "{}" {
		var rs RoomState
		if json.Unmarshal([]byte(s), &rs) == nil {
			status.RoomState = &rs
		}
	}
	if s, ok := vals[2].(string); ok && s != "" && s != "{}" {
		var br BathRoomState
		if json.Unmarshal([]byte(s), &br) == nil {
			status.BathRoomState = &br
		}
	}
	if s, ok := vals[3].(string); ok && s != "" && s != "{}" {
		var bs BedState
		if json.Unmarshal([]byte(s), &bs) == nil {
			status.BedState = &bs
		}
	}
	if s, ok := vals[4].(string); ok && s != "" && s != "{}" {
		var as AlarmState
		if json.Unmarshal([]byte(s), &as) == nil {
			status.AlarmState = &as
		}
	}
	if s, ok := vals[5].(string); ok && s != "" && s != "{}" {
		var msg map[string]interface{}
		if json.Unmarshal([]byte(s), &msg) == nil {
			status.Message = msg
		}
	}
	return status, nil
}

func (r *Reader) ReadAIFields(ctx context.Context, cardID string) (map[string]string, error) {
	var aiKeys []string
	for _, f := range StateFields {
		if f.Owner == OwnerAI {
			aiKeys = append(aiKeys, f.Key)
			if f.HasTTL {
				aiKeys = append(aiKeys, TsKey(f.Key))
			}
		}
	}
	if len(aiKeys) == 0 {
		return nil, nil
	}
	vals, err := r.client.HMGet(ctx, HashKey(cardID), aiKeys...).Result()
	if err != nil {
		return nil, err
	}
	raw := make(map[string]string, len(aiKeys))
	for i, k := range aiKeys {
		if vals[i] != nil {
			raw[k] = fmt.Sprintf("%v", vals[i])
		}
	}
	now := time.Now().UnixMilli()
	result := make(map[string]string, len(StateFields))
	for _, f := range StateFields {
		if f.Owner != OwnerAI {
			continue
		}
		v, exists := raw[f.Key]
		if !exists {
			if f.Default != "" {
				result[f.Key] = f.Default
			}
			continue
		}
		if f.HasTTL {
			tsStr, ok := raw[TsKey(f.Key)]
			if !ok || ttlExpired(tsStr, now) {
				if f.Default != "" {
					result[f.Key] = f.Default
				}
				continue
			}
		}
		result[f.Key] = v
	}
	return result, nil
}

// ReadDeviceStatus 读单个设备状态（device:status:{deviceID} Hash）。
func (r *Reader) ReadDeviceStatus(ctx context.Context, deviceID string) (*DeviceStatus, error) {
	if deviceID == "" {
		return nil, nil
	}
	vals, err := r.client.HGetAll(ctx, DeviceStatusHashKey(deviceID)).Result()
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, nil
	}
	return parseDeviceStatusHash(deviceID, vals), nil
}

// ReadDeviceStatusByDeviceIDs 批量读多个设备状态（pipeline，按 deviceID 列表）。
// 用于卡片详情聚合：调用方先从 cards.devices JSONB 拿 deviceID 列表，再调本接口。
func (r *Reader) ReadDeviceStatusByDeviceIDs(ctx context.Context, deviceIDs []string) (map[string]*DeviceStatus, error) {
	if len(deviceIDs) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{}, len(deviceIDs))
	unique := make([]string, 0, len(deviceIDs))
	for _, d := range deviceIDs {
		if d == "" {
			continue
		}
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
		unique = append(unique, d)
	}
	if len(unique) == 0 {
		return nil, nil
	}
	pipe := r.client.Pipeline()
	cmds := make([]*redis.StringStringMapCmd, 0, len(unique))
	for _, d := range unique {
		cmds = append(cmds, pipe.HGetAll(ctx, DeviceStatusHashKey(d)))
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}
	out := make(map[string]*DeviceStatus, len(unique))
	for i, d := range unique {
		vals, e := cmds[i].Result()
		if e != nil || len(vals) == 0 {
			continue
		}
		out[d] = parseDeviceStatusHash(d, vals)
	}
	return out, nil
}

func parseDeviceStatusHash(deviceID string, vals map[string]string) *DeviceStatus {
	atoi := func(k string) int {
		s, ok := vals[k]
		if !ok || s == "" {
			return 0
		}
		v, _ := strconv.Atoi(s)
		return v
	}
	atoi64 := func(k string) int64 {
		s, ok := vals[k]
		if !ok || s == "" {
			return 0
		}
		v, _ := strconv.ParseInt(s, 10, 64)
		return v
	}
	got := vals["device_id"]
	if got == "" {
		got = deviceID
	}
	return &DeviceStatus{
		DeviceUID:      vals["device_uid"],
		DeviceID:       got,
		DeviceType:     vals["device_type"],
		UpdatedAt:      atoi64("updated_at"),
		LastSeenMs:     atoi64("last_seen_ms"),
		Offline:        atoi("offline"),
		SignalPoor:     atoi("signal_poor"),
		AngleAbnormal:  atoi("angle_abnormal"),
		SensorDetached: atoi("sensor_detached"),
	}
}

func (r *Reader) Exists(ctx context.Context, cardID string) (bool, error) {
	n, err := r.client.Exists(ctx, HashKey(cardID)).Result()
	return n > 0, err
}

func ttlExpired(tsStr string, nowMs int64) bool {
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return true
	}
	return nowMs-ts > AIFieldTTL.Milliseconds()
}
