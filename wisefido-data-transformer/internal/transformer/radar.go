package transformer

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"wisefido-data-transformer/internal/models"
	"wisefido-data-transformer/internal/repository"
	
	"go.uber.org/zap"
)

// RadarTransformer 雷达数据转换器
type RadarTransformer struct {
	snomedRepo *repository.SNOMEDRepository
	logger     *zap.Logger
}

// NewRadarTransformer 创建雷达数据转换器
func NewRadarTransformer(snomedRepo *repository.SNOMEDRepository, logger *zap.Logger) *RadarTransformer {
	return &RadarTransformer{
		snomedRepo: snomedRepo,
		logger:     logger,
	}
}

// Transform 转换雷达原始数据为标准格式
func (t *RadarTransformer) Transform(rawData *models.RawDeviceData) (*models.StandardizedData, error) {
	stdData := &models.StandardizedData{
		TenantID:  rawData.TenantID,
		DeviceID:  rawData.DeviceID,
		Timestamp: time.Unix(rawData.Timestamp, 0),
		DataType:  "observation", // 默认为 observation，告警事件由 alarm 服务判断
	}
	
	// 序列化原始数据
	rawOriginal, err := json.Marshal(rawData.RawData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal raw data: %w", err)
	}
	stdData.RawOriginal = rawOriginal
	
	// 转换轨迹数据
	if err := t.transformTrackingData(rawData.RawData, stdData); err != nil {
		t.logger.Warn("Failed to transform tracking data", zap.Error(err))
	}
	
	// 转换姿态数据
	if err := t.transformPosture(rawData.RawData, stdData); err != nil {
		t.logger.Warn("Failed to transform posture", zap.Error(err))
	}
	
	// 转换生命体征数据
	if err := t.transformVitalSigns(rawData.RawData, stdData); err != nil {
		t.logger.Warn("Failed to transform vital signs", zap.Error(err))
	}
	
	// 转换事件数据
	if err := t.transformEvent(rawData.RawData, stdData); err != nil {
		t.logger.Warn("Failed to transform event", zap.Error(err))
	}
	
	// 确定 category（根据转换后的数据）
	t.determineCategory(stdData)
	
	return stdData, nil
}

// transformTrackingData 转换轨迹数据
func (t *RadarTransformer) transformTrackingData(rawData map[string]interface{}, stdData *models.StandardizedData) error {
	// 提取 tracking_id
	if trackingID, ok := rawData["tracking_id"]; ok {
		if id, err := parseInt(trackingID); err == nil {
			stdData.TrackingID = &id
		}
	}
	
	// 提取位置坐标（单位转换：dm -> cm）
	if posX, ok := rawData["position_x"]; ok {
		if x, err := parseInt(posX); err == nil {
			// 假设原始数据是 dm，转换为 cm
			xCm := x * 10
			stdData.RadarPosX = &xCm
		}
	}
	
	if posY, ok := rawData["position_y"]; ok {
		if y, err := parseInt(posY); err == nil {
			yCm := y * 10
			stdData.RadarPosY = &yCm
		}
	}
	
	if posZ, ok := rawData["position_z"]; ok {
		if z, err := parseInt(posZ); err == nil {
			zCm := z * 10
			stdData.RadarPosZ = &zCm
		}
	}
	
	return nil
}

// transformPosture 转换姿态数据
func (t *RadarTransformer) transformPosture(rawData map[string]interface{}, stdData *models.StandardizedData) error {
	// 提取姿态值
	postureValue, ok := rawData["posture"]
	if !ok {
		return nil // 没有姿态数据
	}
	
	// 转换为字符串
	postureStr := fmt.Sprintf("%v", postureValue)
	
	// 查询 SNOMED 映射
	mapping, err := t.snomedRepo.GetPostureMapping(postureStr, nil)
	if err != nil {
		return fmt.Errorf("failed to get posture mapping: %w", err)
	}
	
	stdData.PostureSNOMEDCode = mapping.SNOMEDCode
	stdData.PostureDisplay = &mapping.SNOMEDDisplay
	
	return nil
}

// transformVitalSigns 转换生命体征数据
func (t *RadarTransformer) transformVitalSigns(rawData map[string]interface{}, stdData *models.StandardizedData) error {
	// 心率
	if hr, ok := rawData["heart_rate"]; ok {
		if heartRate, err := parseInt(hr); err == nil {
			stdData.HeartRate = &heartRate
			// 固定值：心率测量项目的 SNOMED CT 编码
			hrCode := "364075005"
			hrDisplay := "Heart rate"
			stdData.HeartRateCode = &hrCode
			stdData.HeartRateDisplay = &hrDisplay
		}
	}
	
	// 呼吸率
	if br, ok := rawData["breath_rate"]; ok {
		if breathRate, err := parseInt(br); err == nil {
			stdData.RespiratoryRate = &breathRate
			// 固定值：呼吸频率测量项目的 SNOMED CT 编码
			rrCode := "86290005"
			rrDisplay := "Respiratory rate"
			stdData.RespiratoryRateCode = &rrCode
			stdData.RespiratoryRateDisplay = &rrDisplay
		}
	}
	
	return nil
}

// transformEvent 转换事件数据
func (t *RadarTransformer) transformEvent(rawData map[string]interface{}, stdData *models.StandardizedData) error {
	// 提取事件类型（如果有）
	// 注意：事件类型可能是原始值，需要映射到标准事件类型
	// 这里假设原始数据中已经有标准事件类型，或者需要根据其他字段判断
	
	// 示例：如果有 event_type 字段
	if eventType, ok := rawData["event_type"]; ok {
		eventTypeStr := fmt.Sprintf("%v", eventType)
		
		// 查询事件映射
		mapping, err := t.snomedRepo.GetEventMapping(eventTypeStr)
		if err != nil {
			// 如果映射不存在，可能是标准事件类型，直接使用
			stdData.EventType = &eventTypeStr
		} else {
			stdData.EventType = &eventTypeStr
			stdData.EventSNOMEDCode = mapping.SNOMEDCode
			stdData.EventDisplay = &mapping.SNOMEDDisplay
		}
	}
	
	// 提取区域 ID
	if areaID, ok := rawData["area_id"]; ok {
		if id, err := parseInt(areaID); err == nil {
			stdData.AreaID = &id
		}
	}
	
	return nil
}

// determineCategory 确定 FHIR Category
func (t *RadarTransformer) determineCategory(stdData *models.StandardizedData) {
	// 如果有生命体征数据，category 为 vital-signs
	if stdData.HeartRate != nil || stdData.RespiratoryRate != nil {
		stdData.Category = "vital-signs"
		return
	}
	
	// 如果有姿态数据，category 为 activity
	if stdData.PostureSNOMEDCode != nil {
		// 从姿态映射中获取 category（已经在 GetPostureMapping 中获取）
		// 这里需要从映射中获取，但为了简化，先使用 activity
		stdData.Category = "activity"
		return
	}
	
	// 如果有事件数据，根据事件类型确定 category
	if stdData.EventType != nil {
		// 查询事件映射获取 category
		mapping, err := t.snomedRepo.GetEventMapping(*stdData.EventType)
		if err == nil {
			stdData.Category = mapping.Category
		} else {
			stdData.Category = "activity" // 默认
		}
		return
	}
	
	// 默认 category
	stdData.Category = "activity"
}

// parseInt 解析整数
func parseInt(v interface{}) (int, error) {
	switch val := v.(type) {
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case float64:
		return int(val), nil
	case string:
		return strconv.Atoi(val)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

