// iot_timeseries_repo.go — iot_timeseries 入库（wisefido-iot）；读回放见 wisefido-data：iot_playback_repo.go
//
// Package repository — iot_timeseries 入库与 owlRD/db/18_iot_timeseries.sql、owl-common IoTStreamMessage 对齐。
//
//   - Head 展开写入列：tenant_id, device_id, device_uid, device_type, card_id, "timestamp", topic_type, category,
//     branch_name, building_name, unit_name, room_name, bed_name（名称快照；流里若仍为 *_id 键则写入对应 *_name 列）。
//   - "timestamp"：流顶层 timestamp，Unix 毫秒（≥1e12 视为毫秒，否则视为秒×1000）。
//   - data_value：仅序列化流字段 dataValue（JSON 数组），原始进库，不包一层 object、不混入 data_key。
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	rediscommon "owl-common/redis"

	"go.uber.org/zap"
)

const timestampMsThreshold = 1_000_000_000_000 // >= 视为毫秒

func timestampMSFromData(data map[string]interface{}) int64 {
	ts, ok := data["timestamp"]
	if !ok {
		return time.Now().UnixMilli()
	}
	var n int64
	switch v := ts.(type) {
	case int64:
		n = v
	case float64:
		n = int64(v)
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return time.Now().UnixMilli()
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t.UnixMilli()
		}
		parsed, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return time.Now().UnixMilli()
		}
		n = parsed
	default:
		return time.Now().UnixMilli()
	}
	if n >= timestampMsThreshold {
		return n
	}
	if n > 0 {
		return n * 1000
	}
	return time.Now().UnixMilli()
}

func stringPtrFromData(data map[string]interface{}, key string) *string {
	s, ok := data[key].(string)
	if !ok || strings.TrimSpace(s) == "" {
		return nil
	}
	s = strings.TrimSpace(s)
	return &s
}

// snapshotName 优先 name 键，否则用 legacy 键（流仍可能带 branch_id 等，落入 *_name 列作字符串快照）。
func snapshotName(data map[string]interface{}, nameKey, legacyKey string) *string {
	if p := stringPtrFromData(data, nameKey); p != nil {
		return p
	}
	return stringPtrFromData(data, legacyKey)
}

// uuidPtrFromData 按优先级 keys 取 string 字段，验证是 UUID 格式后返回（非 UUID/空值返回 nil，让列保 NULL）。
func uuidPtrFromData(data map[string]interface{}, keys ...string) *string {
	for _, k := range keys {
		if p := stringPtrFromData(data, k); p != nil {
			// 简单 UUID 长度校验（36 字符 + 4 个 dash），非严格但够过滤数字 / 短串
			if len(*p) == 36 && (*p)[8] == '-' && (*p)[13] == '-' && (*p)[18] == '-' && (*p)[23] == '-' {
				return p
			}
		}
	}
	return nil
}

func inferTopicTypeFromDataKey(data map[string]interface{}) *string {
	dk, ok := data["data_key"].(string)
	if !ok || dk == "" {
		return nil
	}
	var tt string
	switch dk {
	case "alarmNotify":
		tt = "alarm"
	case "realtime":
		tt = "monitor"
	case "sleepStage":
		tt = "stat"
	default:
		return nil
	}
	return &tt
}

func inferCategoryFromDataValue(data map[string]interface{}) *string {
	dv, ok := data[rediscommon.DataValueKey]
	if !ok || dv == nil {
		return nil
	}
	firstMap := func(m interface{}) *string {
		mm, ok := m.(map[string]interface{})
		if !ok {
			return nil
		}
		if c, ok := mm[rediscommon.DataCategoryKey].(string); ok && c != "" {
			return &c
		}
		if c, ok := mm["category"].(string); ok && c != "" {
			return &c
		}
		return nil
	}
	switch v := dv.(type) {
	case []interface{}:
		if len(v) == 0 {
			return nil
		}
		return firstMap(v[0])
	case map[string]interface{}:
		return firstMap(v)
	default:
		return nil
	}
}

// marshalDataValueRaw 只序列化 dataValue，与流一致。
func marshalDataValueRaw(data map[string]interface{}) ([]byte, error) {
	dv, ok := data[rediscommon.DataValueKey]
	if !ok || dv == nil {
		return []byte("[]"), nil
	}
	b, err := json.Marshal(dv)
	if err != nil {
		return nil, fmt.Errorf("marshal data_value: %w", err)
	}
	return b, nil
}

// IoTTimeSeriesRepository IoT 时序数据仓库
type IoTTimeSeriesRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewIoTTimeSeriesRepository 创建 IoT 时序数据仓库
func NewIoTTimeSeriesRepository(db *sql.DB, logger *zap.Logger) *IoTTimeSeriesRepository {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &IoTTimeSeriesRepository{db: db, logger: logger}
}

// Insert 将 iot:*:stream 展开 map 写入 iot_timeseries。
func (r *IoTTimeSeriesRepository) Insert(data map[string]interface{}) (int64, error) {
	deviceID, _ := data["device_id"].(string)
	if deviceID == "" {
		if uid, ok := data["device_uid"].(string); ok && uid != "" && r.db != nil {
			if id, err := r.resolveDeviceIDByUID(uid); err == nil && id != "" {
				deviceID = id
				data["device_id"] = id
			}
		}
		if deviceID == "" {
			return 0, fmt.Errorf("missing required field: device_id")
		}
	}

	tenantID := strings.TrimSpace(streamStr(data["tenant_id"]))
	if tenantID == "" {
		return 0, fmt.Errorf("missing tenant_id")
	}

	deviceUID := strings.TrimSpace(streamStr(data["device_uid"]))
	if deviceUID == "" {
		return 0, fmt.Errorf("missing device_uid")
	}

	timestampMs := timestampMSFromData(data)

	var topicType, category *string
	if tt, ok := data["topic_type"].(string); ok && tt != "" {
		topicType = &tt
	} else if tt := inferTopicTypeFromDataKey(data); tt != nil {
		topicType = tt
	}
	if cat, ok := data["category"].(string); ok && cat != "" {
		category = &cat
	} else if cat := inferCategoryFromDataValue(data); cat != nil {
		category = cat
	}

	var deviceIDPtr *string
	if deviceID != "" {
		deviceIDPtr = &deviceID
	}

	deviceType := stringPtrFromData(data, "device_type")

	// Where (snapshot)：仅 alarm/event 流写 room_id（5W where 维度，低频但需法律级永久 snapshot）。
	// monitor 流是 hot path 高频时序（每秒每雷达 multi tracks），不写 room_id 以省存储；
	// 按 room 维度查 monitor 时 JOIN devices.bound_room_id 取最新（device 极少跨房间重绑）。
	// bed_id 同理仅 alarm/event 用。
	// 不再写 card_id（card 是视图，view-level ID 不入业务表）。
	var roomID, bedID *string
	if topicType != nil && (*topicType == "alarm" || *topicType == "event") {
		roomID = uuidPtrFromData(data, "semantic_location", "room_id")
		bedID = uuidPtrFromData(data, "bed_id")
	}

	branchName := snapshotName(data, "branch_name", "branch_id")
	buildingName := snapshotName(data, "building_name", "building_id")
	unitName := snapshotName(data, "unit_name", "unit_id")
	roomName := snapshotName(data, "room_name", "room_id")
	bedName := snapshotName(data, "bed_name", "bed_id")

	dataValueJSON, err := marshalDataValueRaw(data)
	if err != nil {
		return 0, err
	}

	const q = `
		INSERT INTO iot_timeseries (
			tenant_id, device_id, device_uid, device_type,
			"timestamp", topic_type, category,
			branch_name, building_name, unit_name, room_name, bed_name,
			room_id, bed_id,
			data_value
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id`

	var id int64
	err = r.db.QueryRow(q,
		tenantID, deviceIDPtr, deviceUID, deviceType,
		timestampMs, topicType, category,
		branchName, buildingName, unitName, roomName, bedName,
		roomID, bedID,
		dataValueJSON,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert iot_timeseries: %w", err)
	}

	r.logger.Debug("iot_timeseries inserted",
		zap.Int64("id", id),
		zap.String("device_id", deviceID),
		zap.Stringp("topic_type", topicType),
	)
	return id, nil
}

func streamStr(v interface{}) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (r *IoTTimeSeriesRepository) resolveDeviceIDByUID(deviceUID string) (string, error) {
	var deviceID string
	err := r.db.QueryRowContext(context.Background(),
		`SELECT device_id::text FROM devices WHERE device_uid = $1 LIMIT 1`, deviceUID).Scan(&deviceID)
	if err != nil {
		return "", err
	}
	return deviceID, nil
}

// GetDeviceLocation 获取设备位置信息（unit_id, room_id）
func (r *IoTTimeSeriesRepository) GetDeviceLocation(deviceID string) (unitID *string, roomID *string, err error) {
	const query = `
		SELECT u.unit_id, r.room_id
		FROM devices d
		LEFT JOIN beds b ON d.bound_bed_id = b.bed_id
		LEFT JOIN rooms r ON COALESCE(b.room_id, d.bound_room_id) = r.room_id
		LEFT JOIN units u ON r.unit_id = u.unit_id
		WHERE d.device_id = $1
		LIMIT 1`
	var uID, rID sql.NullString
	err = r.db.QueryRow(query, deviceID).Scan(&uID, &rID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("device not found: %s", deviceID)
		}
		return nil, nil, fmt.Errorf("query device location: %w", err)
	}
	if uID.Valid {
		unitID = &uID.String
	}
	if rID.Valid {
		roomID = &rID.String
	}
	return unitID, roomID, nil
}

// GetDeviceHardwareInfo 通过 device_id 关联 device_store
func (r *IoTTimeSeriesRepository) GetDeviceHardwareInfo(deviceID string) (deviceType, deviceModel, serialNumber, deviceUID *string, err error) {
	const query = `
		SELECT ds.device_type, ds.device_model, NULL::text, ds.device_uid
		FROM devices d
		JOIN device_store ds ON d.device_id = ds.device_id
		WHERE d.device_id = $1
		LIMIT 1`
	var dt, dm, sn, u sql.NullString
	err = r.db.QueryRow(query, deviceID).Scan(&dt, &dm, &sn, &u)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil, nil, fmt.Errorf("device not found: %s", deviceID)
		}
		return nil, nil, nil, nil, fmt.Errorf("query device hardware: %w", err)
	}
	if dt.Valid {
		deviceType = &dt.String
	}
	if dm.Valid {
		deviceModel = &dm.String
	}
	if sn.Valid {
		serialNumber = &sn.String
	}
	if u.Valid {
		deviceUID = &u.String
	}
	return deviceType, deviceModel, serialNumber, deviceUID, nil
}

// UpdateLocation 更新名称快照列（与当前表 unit_name/room_name 一致）
func (r *IoTTimeSeriesRepository) UpdateLocation(id int64, unitName *string, roomName *string) error {
	_, err := r.db.Exec(
		`UPDATE iot_timeseries SET unit_name = $1, room_name = $2 WHERE id = $3`,
		unitName, roomName, id,
	)
	if err != nil {
		return fmt.Errorf("update location: %w", err)
	}
	return nil
}

// InvalidateLocationCache 设备绑定变更时调用（占位）
func (r *IoTTimeSeriesRepository) InvalidateLocationCache(deviceID string) {
	r.logger.Debug("location cache invalidate (no-op)", zap.String("device_id", deviceID))
}
