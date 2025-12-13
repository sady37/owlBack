// Package transformer 提供数据转换功能
// 
// 将原始设备数据转换为标准化格式，包括：
// - SNOMED CT 编码映射
// - FHIR Category 分类
// - 单位标准化
// - 数据验证和清洗
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

// SleepaceTransformer Sleepace 数据转换器
// 
// 负责将 Sleepace 设备的原始数据转换为标准化格式。
// 
// 转换内容：
// - 生命体征：心率、呼吸率（过滤无效值 0/255）
// - 床状态：0=在床 → SNOMED "370998004", 1=离床 → SNOMED "424287000"
// - 睡眠阶段：0=清醒, 1=浅睡眠, 2=深睡眠, 3=REM睡眠 → SNOMED 编码
// - 行为事件：坐起、翻身、体动等
// - FHIR Category：根据数据内容自动分类（vital-signs 或 activity）
type SleepaceTransformer struct {
	snomedRepo *repository.SNOMEDRepository // SNOMED CT 映射仓库
	logger     *zap.Logger                   // 日志记录器
}

// NewSleepaceTransformer 创建 Sleepace 数据转换器
func NewSleepaceTransformer(snomedRepo *repository.SNOMEDRepository, logger *zap.Logger) *SleepaceTransformer {
	return &SleepaceTransformer{
		snomedRepo: snomedRepo,
		logger:     logger,
	}
}

// Transform 转换 Sleepace 原始数据为标准格式
// 
// 该方法将 Sleepace 设备的原始数据（JSON 格式）转换为标准化的数据结构，
// 包括 SNOMED CT 编码映射、FHIR Category 分类等。
// 
// 转换流程：
// 1. 解析原始数据（JSON）
// 2. 转换生命体征数据（心率、呼吸率）
// 3. 转换床状态数据（在床/离床）
// 4. 转换睡眠阶段数据（清醒/浅睡眠/深睡眠/REM睡眠）
// 5. 转换行为事件数据（坐起、翻身、体动）
// 6. 确定 FHIR Category（根据数据内容）
// 
// 参数:
//   - rawData: 原始设备数据，包含 device_id、tenant_id、raw_data 等
// 
// 返回:
//   - *models.StandardizedData: 标准化后的数据，包含 SNOMED 编码、FHIR Category 等
//   - error: 如果转换过程中发生错误（如数据格式错误、映射失败等）
// 
// 示例:
//   stdData, err := transformer.Transform(rawData)
//   if err != nil {
//       return nil, fmt.Errorf("转换失败: %w", err)
//   }
func (t *SleepaceTransformer) Transform(rawData *models.RawDeviceData) (*models.StandardizedData, error) {
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
	
	// 转换生命体征数据（Sleepace 的主要数据）
	if err := t.transformVitalSigns(rawData.RawData, stdData); err != nil {
		t.logger.Warn("Failed to transform vital signs", zap.Error(err))
	}
	
	// 转换床状态数据
	if err := t.transformBedStatus(rawData.RawData, stdData); err != nil {
		t.logger.Warn("Failed to transform bed status", zap.Error(err))
	}
	
	// 转换睡眠阶段数据
	if err := t.transformSleepStage(rawData.RawData, stdData); err != nil {
		t.logger.Warn("Failed to transform sleep stage", zap.Error(err))
	}
	
	// 转换行为事件数据
	if err := t.transformBehaviorEvents(rawData.RawData, stdData); err != nil {
		t.logger.Warn("Failed to transform behavior events", zap.Error(err))
	}
	
	// 确定 category（根据转换后的数据）
	t.determineCategory(stdData)
	
	return stdData, nil
}

// transformVitalSigns 转换生命体征数据
func (t *SleepaceTransformer) transformVitalSigns(rawData map[string]interface{}, stdData *models.StandardizedData) error {
	// 心率
	if hr, ok := rawData["heart"]; ok {
		if heartRate, err := parseIntSleepace(hr); err == nil {
			// 过滤无效值（0 或 255 表示无效）
			if heartRate > 0 && heartRate < 255 {
				stdData.HeartRate = &heartRate
				hrCode := "364075005"
				hrDisplay := "Heart rate"
				stdData.HeartRateCode = &hrCode
				stdData.HeartRateDisplay = &hrDisplay
			}
		}
	}
	
	// 呼吸率
	if br, ok := rawData["breath"]; ok {
		if breathRate, err := parseIntSleepace(br); err == nil {
			// 过滤无效值（0 或 255 表示无效）
			if breathRate > 0 && breathRate < 255 {
				stdData.RespiratoryRate = &breathRate
				rrCode := "86290005"
				rrDisplay := "Respiratory rate"
				stdData.RespiratoryRateCode = &rrCode
				stdData.RespiratoryRateDisplay = &rrDisplay
			}
		}
	}
	
	return nil
}

// transformBedStatus 转换床状态数据
func (t *SleepaceTransformer) transformBedStatus(rawData map[string]interface{}, stdData *models.StandardizedData) error {
	// bedStatus: 0=在床, 1=离床
	if bedStatus, ok := rawData["bedStatus"]; ok {
		status, err := parseIntSleepace(bedStatus)
		if err != nil {
			return err
		}
		
		// 映射到 SNOMED CT
		// 0 = 在床 (On bed) -> 370998004 / 248569007
		// 1 = 离床 (Left bed) -> 424287000
		var bedStatusCode, bedStatusDisplay string
		if status == 0 {
			bedStatusCode = "370998004" // On bed
			bedStatusDisplay = "On bed"
		} else if status == 1 {
			bedStatusCode = "424287000" // Left bed
			bedStatusDisplay = "Left bed"
		}
		
		if bedStatusCode != "" {
			stdData.BedStatusSNOMEDCode = &bedStatusCode
			stdData.BedStatusDisplay = &bedStatusDisplay
		}
	}
	
	return nil
}

// transformSleepStage 转换睡眠阶段数据
func (t *SleepaceTransformer) transformSleepStage(rawData map[string]interface{}, stdData *models.StandardizedData) error {
	// sleepStage: 0=清醒, 1=浅睡眠, 2=深睡眠, 3=REM睡眠
	// 注意：Sleepace 的 sleepStage 可能在不同字段中，需要根据实际数据格式调整
	if sleepStage, ok := rawData["sleepStage"]; ok {
		stage, err := parseIntSleepace(sleepStage)
		if err != nil {
			return err
		}
		
		// 映射到 SNOMED CT
		// 0 = 清醒 (Awake) -> 248220002
		// 1 = 浅睡眠 (Light sleep) -> 248232005
		// 2 = 深睡眠 (Deep sleep) -> 248233000
		// 3 = REM睡眠 (REM sleep) -> 248234006
		var sleepStateCode, sleepStateDisplay string
		switch stage {
		case 0:
			sleepStateCode = "248220002"
			sleepStateDisplay = "Awake"
		case 1:
			sleepStateCode = "248232005"
			sleepStateDisplay = "Light sleep"
		case 2:
			sleepStateCode = "248233000"
			sleepStateDisplay = "Deep sleep"
		case 3:
			sleepStateCode = "248234006"
			sleepStateDisplay = "REM sleep"
		}
		
		if sleepStateCode != "" {
			stdData.SleepStateSNOMEDCode = &sleepStateCode
			stdData.SleepStateDisplay = &sleepStateDisplay
		}
	}
	
	return nil
}

// transformBehaviorEvents 转换行为事件数据
func (t *SleepaceTransformer) transformBehaviorEvents(rawData map[string]interface{}, stdData *models.StandardizedData) error {
	// 转换行为事件
	// sitUp: 床上坐起 -> 422256002 / 40199007
	if sitUp, ok := rawData["sitUp"]; ok {
		if val, err := parseIntSleepace(sitUp); err == nil && val > 0 {
			eventType := "BED_SIT_UP"
			eventCode := "422256002"
			eventDisplay := "Sitting up in bed"
			stdData.EventType = &eventType
			stdData.EventSNOMEDCode = &eventCode
			stdData.EventDisplay = &eventDisplay
		}
	}
	
	// turnOver: 翻身
	if turnOver, ok := rawData["turnOver"]; ok {
		if val, err := parseIntSleepace(turnOver); err == nil && val > 0 {
			// 翻身事件通常不单独记录，而是作为姿态变化的一部分
			// 如果需要记录，可以使用相应的事件类型
		}
	}
	
	// bodyMove: 体动
	if bodyMove, ok := rawData["bodyMove"]; ok {
		if val, err := parseIntSleepace(bodyMove); err == nil && val > 0 {
			// 体动事件通常不单独记录，而是作为姿态变化的一部分
		}
	}
	
	return nil
}

// determineCategory 确定 FHIR Category
func (t *SleepaceTransformer) determineCategory(stdData *models.StandardizedData) {
	// 如果有生命体征数据，category 为 vital-signs
	if stdData.HeartRate != nil || stdData.RespiratoryRate != nil {
		stdData.Category = "vital-signs"
		return
	}
	
	// 如果有睡眠状态数据，category 为 activity
	if stdData.SleepStateSNOMEDCode != nil {
		stdData.Category = "activity"
		return
	}
	
	// 如果有床状态数据，category 为 activity
	if stdData.BedStatusSNOMEDCode != nil {
		stdData.Category = "activity"
		return
	}
	
	// 如果有事件数据，category 为 activity
	if stdData.EventType != nil {
		stdData.Category = "activity"
		return
	}
	
	// 默认 category
	stdData.Category = "activity"
}

// parseIntSleepace 解析整数（Sleepace 专用，避免与 radar.go 中的 parseInt 冲突）
func parseIntSleepace(v interface{}) (int, error) {
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

