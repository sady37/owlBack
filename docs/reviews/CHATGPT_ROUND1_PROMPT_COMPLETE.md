# ChatGPT 第1轮审查提示词（完整版）

> **用途**: 直接提交给 ChatGPT 的完整提示词，包含所有代码

---

## 📋 提交给 ChatGPT 的完整提示词

请将以下内容完整复制并提交给 ChatGPT：

---

请审查以下 Go 代码，重点关注以下方面：

## 审查文件

### 1. sensor_fusion.go (核心融合逻辑)

这是传感器融合服务的核心文件，负责将多个设备（Sleepace + Radar）的数据融合为统一的实时数据格式。

### 2. card.go (数据访问层)

这是数据访问层，负责查询卡片和设备关联关系。

### 3. sleepace.go (数据转换)

这是数据转换器，负责将 Sleepace 原始数据转换为标准化格式。

---

## 审查重点

### 1. 代码质量
- 命名是否清晰、符合 Go 规范？
- 函数是否过长（建议 < 50 行）？
- 是否有重复代码？
- 代码结构是否清晰？

### 2. 潜在错误
- 是否有逻辑错误？
- 是否有边界条件未处理？
- 是否有空指针风险？
- 是否有数组越界风险？

### 3. 性能问题
- 是否有 N+1 查询问题？
- 是否有不必要的循环？
- 是否有内存泄漏风险？
- 是否有不必要的内存分配？

### 4. 并发安全
- 是否有数据竞争？
- 是否需要加锁？
- Context 使用是否正确？

### 5. 错误处理
- 所有错误是否都被处理？
- 错误信息是否有意义？
- 是否使用了 %w 包装错误？

### 6. 最佳实践
- 是否符合 Go 代码规范？
- 是否使用了合适的 Go 特性？
- 是否有更好的实现方式？

### 7. 安全性
- 是否有 SQL 注入风险？
- 是否有输入验证？
- 敏感信息是否泄露？

---

## 代码文件

### 文件 1: sensor_fusion.go

```go
// Package fusion 提供传感器融合功能
// 
// 主要功能：
// - 多传感器数据融合（HR/RR、姿态、床状态等）
// - 数据优先级处理（Sleepace 优先于 Radar）
// - 姿态数据合并（合并所有 Radar 设备的 tracking_id）
package fusion

import (
	"fmt"
	"time"
	"wisefido-sensor-fusion/internal/models"
	"wisefido-sensor-fusion/internal/repository"
	
	"go.uber.org/zap"
)

// SensorFusion 传感器融合器
// 
// 负责将多个设备的数据融合为统一的实时数据格式
// 融合规则：
// - HR/RR：优先 Sleepace，无数据则 Radar
// - 床状态/睡眠状态：优先 Sleepace
// - 姿态数据：合并所有 Radar 设备的 tracking_id
type SensorFusion struct {
	cardRepo *repository.CardRepository       // 卡片仓库，用于查询设备关联
	iotRepo  *repository.IoTTimeSeriesRepository // IoT 时序数据仓库，用于查询设备数据
	logger   *zap.Logger                     // 日志记录器
}

// NewSensorFusion 创建传感器融合器
func NewSensorFusion(
	cardRepo *repository.CardRepository,
	iotRepo *repository.IoTTimeSeriesRepository,
	logger *zap.Logger,
) *SensorFusion {
	return &SensorFusion{
		cardRepo: cardRepo,
		iotRepo:  iotRepo,
		logger:   logger,
	}
}

// FuseCardData 融合卡片的所有设备数据
// 
// 该方法从卡片关联的所有设备中收集最新数据，并按照优先级规则进行融合。
// 
// 融合规则：
// 1. HR/RR（心率/呼吸率）：优先使用 Sleepace 数据，如果 Sleepace 无数据则使用 Radar 数据
// 2. 床状态/睡眠状态：优先使用 Sleepace 数据，如果 Sleepace 无数据则使用 Radar 数据（如果有）
// 3. 姿态数据：合并所有 Radar 设备的 tracking_id（不跨设备去重）
// 
// 参数:
//   - cardID: 卡片 ID（UUID 格式）
// 
// 返回:
//   - *models.RealtimeData: 融合后的实时数据，包含心率、呼吸率、姿态等信息
//   - error: 如果融合过程中发生错误（如设备查询失败、数据获取失败等）
func (f *SensorFusion) FuseCardData(cardID string) (*models.RealtimeData, error) {
	// 1. 获取卡片关联的所有设备
	devices, err := f.cardRepo.GetCardDevices(cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card devices: %w", err)
	}
	
	if len(devices) == 0 {
		return nil, fmt.Errorf("no devices found for card: %s", cardID)
	}
	
	// 2. 收集所有设备的最新数据
	var sleepaceData []*models.IoTTimeSeries
	var radarData []*models.IoTTimeSeries
	
	for _, device := range devices {
		// 获取设备最新数据（最近 1 条）
		latestData, err := f.iotRepo.GetLatestByDeviceID(device.DeviceID, 1)
		if err != nil {
			f.logger.Warn("Failed to get latest data for device",
				zap.String("device_id", device.DeviceID),
				zap.Error(err),
			)
			continue
		}
		
		if len(latestData) == 0 {
			continue
		}
		
		// 设置设备类型
		deviceType, err := f.iotRepo.GetDeviceType(device.DeviceID)
		if err != nil {
			f.logger.Warn("Failed to get device type",
				zap.String("device_id", device.DeviceID),
				zap.Error(err),
			)
			continue
		}
		
		latestData[0].DeviceType = deviceType
		
		// 分类数据
		if deviceType == "Sleepace" {
			sleepaceData = append(sleepaceData, latestData[0])
		} else if deviceType == "Radar" {
			radarData = append(radarData, latestData[0])
		}
	}
	
	// 3. 融合数据
	result := &models.RealtimeData{
		Timestamp: time.Now().Unix(),
		Postures:  []models.Posture{},
	}
	
	// 3.1 融合 HR/RR（优先 Sleepace）
	f.fuseVitalSigns(sleepaceData, radarData, result)
	
	// 3.2 融合床状态和睡眠状态（优先 Sleepace）
	f.fuseBedAndSleepStatus(sleepaceData, radarData, result)
	
	// 3.3 融合姿态数据（来自所有 Radar 设备）
	f.fusePostures(radarData, result)
	
	return result, nil
}

// fuseVitalSigns 融合生命体征（HR/RR）
// 规则：优先 Sleepace，无数据则 Radar
func (f *SensorFusion) fuseVitalSigns(
	sleepaceData []*models.IoTTimeSeries,
	radarData []*models.IoTTimeSeries,
	result *models.RealtimeData,
) {
	// 优先使用 Sleepace 数据
	if len(sleepaceData) > 0 {
		for _, data := range sleepaceData {
			if data.HeartRate != nil {
				result.Heart = data.HeartRate
				result.HeartSource = "Sleepace"
				break
			}
		}
		for _, data := range sleepaceData {
			if data.RespiratoryRate != nil {
				result.Breath = data.RespiratoryRate
				result.BreathSource = "Sleepace"
				break
			}
		}
	}
	
	// 如果 Sleepace 没有数据，使用 Radar 数据
	if result.Heart == nil && len(radarData) > 0 {
		for _, data := range radarData {
			if data.HeartRate != nil {
				result.Heart = data.HeartRate
				result.HeartSource = "Radar"
				break
			}
		}
	}
	if result.Breath == nil && len(radarData) > 0 {
		for _, data := range radarData {
			if data.RespiratoryRate != nil {
				result.Breath = data.RespiratoryRate
				result.BreathSource = "Radar"
				break
			}
		}
	}
}

// fuseBedAndSleepStatus 融合床状态和睡眠状态
// 规则：优先 Sleepace
func (f *SensorFusion) fuseBedAndSleepStatus(
	sleepaceData []*models.IoTTimeSeries,
	radarData []*models.IoTTimeSeries,
	result *models.RealtimeData,
) {
	// 优先使用 Sleepace 数据
	if len(sleepaceData) > 0 {
		for _, data := range sleepaceData {
			if data.BedStatusSNOMEDCode != nil {
				result.BedStatus = data.BedStatusSNOMEDCode
				break
			}
		}
		for _, data := range sleepaceData {
			if data.SleepStateSNOMEDCode != nil {
				result.SleepStage = data.SleepStateSNOMEDCode
				break
			}
		}
	}
	
	// 如果 Sleepace 没有数据，使用 Radar 数据（如果有）
	if result.BedStatus == nil && len(radarData) > 0 {
		for _, data := range radarData {
			if data.BedStatusSNOMEDCode != nil {
				result.BedStatus = data.BedStatusSNOMEDCode
				break
			}
		}
	}
	if result.SleepStage == nil && len(radarData) > 0 {
		for _, data := range radarData {
			if data.SleepStateSNOMEDCode != nil {
				result.SleepStage = data.SleepStateSNOMEDCode
				break
			}
		}
	}
}

// fusePostures 融合姿态数据
// 规则：合并所有 Radar 设备的 tracking_id（不跨设备去重）
func (f *SensorFusion) fusePostures(
	radarData []*models.IoTTimeSeries,
	result *models.RealtimeData,
) {
	// 收集所有 Radar 设备的姿态数据
	trackingMap := make(map[string]*models.Posture)
	
	for _, data := range radarData {
		if data.TrackingID != nil && data.PostureSNOMEDCode != nil {
			trackingID := *data.TrackingID
			
			// 如果该 tracking_id 已存在，更新（使用最新的数据）
			// 注意：这里简化处理，直接使用最后一条数据
			// TODO: 实现时间戳比较逻辑，使用更新的数据
			
			posture := &models.Posture{
				TrackingID:    trackingID,
				PostureCode:   *data.PostureSNOMEDCode,
				PostureDisplay: "",
			}
			
			if data.PostureDisplay != nil {
				posture.PostureDisplay = *data.PostureDisplay
			}
			
			trackingMap[trackingID] = posture
		}
	}
	
	// 转换为列表
	for _, posture := range trackingMap {
		result.Postures = append(result.Postures, *posture)
	}
	
	result.PersonCount = len(result.Postures)
}
```

### 文件 2: card.go (相关部分)

```go
package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
)

// CardRepository 卡片仓库
type CardRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewCardRepository 创建卡片仓库
func NewCardRepository(db *sql.DB, logger *zap.Logger) *CardRepository {
	return &CardRepository{
		db:     db,
		logger: logger,
	}
}

// GetCardByDeviceID 根据设备ID获取关联的卡片
// 
// 该方法根据设备的绑定关系（绑定到 Bed 或 Room）查询对应的卡片。
// 
// 查询逻辑：
// 1. 根据 device_id 查询设备信息，获取 bound_bed_id 或 bound_room_id
// 2. 如果设备绑定到 bed（bound_bed_id IS NOT NULL）：
//    - 查询 ActiveBed 类型的卡片（cards.bed_id = bound_bed_id）
// 3. 如果设备绑定到 room（bound_room_id IS NOT NULL）：
//    - 查询 Location 类型的卡片（cards.unit_id = room.unit_id）
// 
// 注意：
// - 设备只能绑定到 Bed 或 Room 之一（互斥约束）
// - 如果设备未绑定或绑定关系不存在，返回错误
// 
// 参数:
//   - deviceID: 设备 ID（UUID 格式）
// 
// 返回:
//   - *CardInfo: 卡片信息，包含 card_id、card_type、tenant_id 等
//   - error: 如果设备不存在、未绑定或查询失败
func (r *CardRepository) GetCardByDeviceID(deviceID string) (*CardInfo, error) {
	query := `
		WITH device_info AS (
			SELECT 
				d.device_id,
				d.tenant_id,
				d.bound_bed_id,
				d.bound_room_id
			FROM devices d
			WHERE d.device_id = $1
		),
		bed_card AS (
			SELECT 
				c.card_id,
				c.tenant_id,
				c.card_type,
				c.bed_id,
				c.unit_id
			FROM cards c
			INNER JOIN device_info di ON c.bed_id = di.bound_bed_id
			WHERE di.bound_bed_id IS NOT NULL
			LIMIT 1
		),
		room_card AS (
			SELECT 
				c.card_id,
				c.tenant_id,
				c.card_type,
				c.bed_id,
				c.unit_id
			FROM cards c
			INNER JOIN device_info di ON c.unit_id = (
				SELECT r.unit_id FROM rooms r WHERE r.room_id = di.bound_room_id
			)
			WHERE di.bound_room_id IS NOT NULL
			LIMIT 1
		)
		SELECT card_id, tenant_id, card_type, bed_id, unit_id
		FROM bed_card
		UNION ALL
		SELECT card_id, tenant_id, card_type, bed_id, unit_id
		FROM room_card
		LIMIT 1
	`
	
	card := &CardInfo{}
	var bedID, unitID sql.NullString
	
	err := r.db.QueryRow(query, deviceID).Scan(
		&card.CardID,
		&card.TenantID,
		&card.CardType,
		&bedID,
		&unitID,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found for device: %s", deviceID)
		}
		return nil, fmt.Errorf("failed to query card: %w", err)
	}
	
	if bedID.Valid {
		card.BedID = &bedID.String
	}
	if unitID.Valid {
		card.UnitID = &unitID.String
	}
	
	return card, nil
}

// GetCardDevices 获取卡片关联的所有设备信息
// 
// 该方法从 cards 表的 devices JSONB 字段中提取设备信息。
func (r *CardRepository) GetCardDevices(cardID string) ([]DeviceInfo, error) {
	query := `
		SELECT devices
		FROM cards
		WHERE card_id = $1
	`
	
	var devicesJSON []byte
	err := r.db.QueryRow(query, cardID).Scan(&devicesJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("card not found: %s", cardID)
		}
		return nil, fmt.Errorf("failed to query card devices: %w", err)
	}
	
	var devices []DeviceInfo
	if err := json.Unmarshal(devicesJSON, &devices); err != nil {
		return nil, fmt.Errorf("failed to unmarshal devices JSON: %w", err)
	}
	
	return devices, nil
}

// CardInfo 卡片信息
type CardInfo struct {
	CardID   string
	TenantID string
	CardType string // "ActiveBed" 或 "Location"
	BedID    *string
	UnitID   *string
}

// DeviceInfo 设备信息（从 cards.devices JSONB 解析）
type DeviceInfo struct {
	DeviceID    string `json:"device_id"`
	DeviceName  string `json:"device_name"`
	DeviceType  string `json:"device_type"` // "Radar" 或 "Sleepace"
	DeviceModel string `json:"device_model"`
	BindingType string `json:"binding_type"` // "direct" 或 "indirect"
}
```

### 文件 3: sleepace.go (相关部分 - Transform 方法)

```go
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
	if sleepStage, ok := rawData["sleepStage"]; ok {
		stage, err := parseIntSleepace(sleepStage)
		if err != nil {
			return err
		}
		
		// 映射到 SNOMED CT
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
	// sitUp: 床上坐起
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
```

---

## 设计背景

### 传感器融合规则
1. **HR/RR（心率/呼吸率）**: 优先使用 Sleepace 数据，如果 Sleepace 无数据则使用 Radar 数据
2. **床状态/睡眠状态**: 优先使用 Sleepace 数据，如果 Sleepace 无数据则使用 Radar 数据（如果有）
3. **姿态数据**: 合并所有 Radar 设备的 tracking_id（不跨设备去重）

### 已知问题
- 当前实现存在 N+1 查询问题（在 `FuseCardData` 方法中，对每个设备单独查询）
- 姿态数据融合中，时间戳比较逻辑未实现（TODO 注释）

---

## 请提供

1. **发现的问题列表**（按严重性排序）
   - 问题描述
   - 位置（文件:行号）
   - 严重性（高/中/低）
   - 修复建议

2. **代码质量评分**（1-10 分）
   - 总体评分
   - 分项评分（代码质量、性能、安全性、可维护性、最佳实践）

3. **改进建议**
   - 高优先级建议
   - 中优先级建议

4. **总体评价**
   - 优点
   - 需要改进的地方

---

请详细审查并提供反馈。

---

## 📝 使用说明

1. **复制提示词**: 从 "请审查以下 Go 代码..." 开始，到 "请详细审查并提供反馈。" 结束
2. **提交给 ChatGPT**: 将完整提示词提交给 ChatGPT
3. **记录反馈**: 将 ChatGPT 的反馈记录到 `docs/reviews/chatgpt_round1_sensor_fusion.md`

---

**准备日期**: 2024-12-19  
**状态**: ✅ 完整版，可直接提交给 ChatGPT

