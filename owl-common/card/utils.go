package card

import (
	//"encoding/json"
	//"fmt"
	"strconv"
)

// ResidentJSON resident JSON format (for cards.residents JSONB field)
type ResidentJSON struct {
	ResidentID string `json:"resident_id"`
	Nickname   string `json:"nickname"`
}

// DeviceJSON device JSON format (for cards.devices JSONB field).
// 写 cards.devices 时：DeviceJSON 需含 device_uid；组装 DeviceInfo 的查询（如 GetDevicesByBed、GetUnboundDevicesByUnit）从 devices 表选 d.device_uid 填入。
type DeviceJSON struct {
	DeviceID    string  `json:"device_id"`
	DeviceUID   string  `json:"device_uid,omitempty"`  // 与 card-overview、GetCardDevices 对齐
	DeviceCode  string  `json:"device_code,omitempty"` // 与 card-overview、GetCardDevices 对齐
	DeviceName  string  `json:"device_name"`
	DeviceType  string  `json:"device_type"`
	DeviceModel string  `json:"device_model"`
	BedID       *string `json:"bed_id,omitempty"`
	BedName     *string `json:"bed_name,omitempty"`
	RoomID      *string `json:"room_id,omitempty"`
	RoomName    *string `json:"room_name,omitempty"`
	UnitID      string  `json:"unit_id"`
}

/*
// ConvertDevicesToJSON converts device list to JSON（含 device_uid、device_code，与 card-overview、GetCardDevices 对齐）
func ConvertDevicesToJSON(devices []DeviceInfo) ([]byte, error) {
	var deviceJSONs []DeviceJSON
	for _, device := range devices {
		var deviceTypeStr string
		if device.DeviceType != nil {
			deviceTypeStr = fmt.Sprint(device.DeviceType)
		}
		deviceJSONs = append(deviceJSONs, DeviceJSON{
			DeviceID:    device.DeviceID,
			DeviceUID:   device.DeviceUID,
			DeviceCode:  device.DeviceCode,
			DeviceName:  device.DeviceName,
			DeviceType:  deviceTypeStr,
			DeviceModel: device.DeviceModel,
			BedID:       device.BoundBedID,

			RoomID:      device.BoundRoomID,

		})
	}
	return json.Marshal(deviceJSONs)
}

// ConvertResidentsToJSON converts resident list to JSON
func ConvertResidentsToJSON(residents []ResidentInfo) ([]byte, error) {
	var residentJSONs []ResidentJSON
	for _, resident := range residents {
		residentJSONs = append(residentJSONs, ResidentJSON{
			ResidentID: resident.ResidentID,
			Nickname:   resident.Nickname,
		})
	}
	return json.Marshal(residentJSONs)
}
*/

func poseStringToInt(pose string) int {
	if pose == "" {
		return 0
	}

	// 直接支持数字字符串
	if n, err := strconv.Atoi(pose); err == nil {
		if n >= 0 && n <= 11 {
			return n
		}
	}

	poseMap := map[string]int{
		"Initialization":           0,
		"Walking":                  1,
		"SuspectedFall":            2,
		"Sitting position":         3,
		"Standing position":        4,
		"Fall":                     5,
		"Lying position":           6,
		"SuspectedSittingOnGround": 7,
		"SittingOnGround":          8,
		"Sitting up in bed":        9,
		"SuspectedBedSitUp":        10,
		"BedSitUp":                 11,
	}

	if v, ok := poseMap[pose]; ok {
		return v
	}
	return 0
}
