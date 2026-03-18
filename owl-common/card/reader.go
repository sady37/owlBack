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
	vals, err := r.client.HMGet(ctx, HashKey(cardID),
		"target", "room_state", "bathroom_state", "bed_state", "device_status", "alarm_state", "message",
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
		var ds map[string]*DeviceStatus
		if json.Unmarshal([]byte(s), &ds) == nil {
			status.DeviceStatus = ds
		}
	}
	if s, ok := vals[5].(string); ok && s != "" && s != "{}" {
		var as AlarmState
		if json.Unmarshal([]byte(s), &as) == nil {
			status.AlarmState = &as
		}
	}
	if s, ok := vals[6].(string); ok && s != "" && s != "{}" {
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

func (r *Reader) ReadDeviceStatusBatch(ctx context.Context, cardIDs []string) (map[string]map[string]*DeviceStatus, error) {
	if len(cardIDs) == 0 {
		return nil, nil
	}
	seen := make(map[string]struct{}, len(cardIDs))
	unique := make([]string, 0, len(cardIDs))
	for _, k := range cardIDs {
		if k == "" {
			continue
		}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		unique = append(unique, k)
	}
	if len(unique) == 0 {
		return nil, nil
	}
	pipe := r.client.Pipeline()
	cmds := make([]*redis.StringCmd, 0, len(unique))
	for _, k := range unique {
		cmds = append(cmds, pipe.HGet(ctx, HashKey(k), "device_status"))
	}
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}
	out := make(map[string]map[string]*DeviceStatus, len(unique))
	for i, k := range unique {
		raw, e := cmds[i].Result()
		if e != nil || raw == "" || raw == "{}" {
			out[k] = nil
			continue
		}
		var m map[string]*DeviceStatus
		if json.Unmarshal([]byte(raw), &m) != nil {
			out[k] = nil
			continue
		}
		out[k] = m
	}
	return out, nil
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
