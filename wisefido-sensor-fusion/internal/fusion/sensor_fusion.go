// Package fusion 提供传感器融合功能
// 
// 主要功能：
// - 多传感器数据融合（处理所有卡片上的设备数据）
// - 融合条件：
//   - ActiveBed 卡片：同一床上同时有 Radar 和 Sleepace 设备（只融合 bed_id 有效且相同的设备）
//   - Location 卡片：同一卡片上同时有 Radar 和 Sleepace 设备（融合所有设备）
// - 融合内容：HR/RR、床状态/睡眠状态（优先 Sleepace）
// - 姿态数据：直接使用 Radar 数据（Sleepace 不提供姿态数据）
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
// 
// 融合条件：
// - ActiveBed 卡片：同一床上同时有 Radar 和 Sleepace 设备（只融合 bed_id 有效且相同的设备）
//   - 场景 A（门牌下只有 1 个 ActiveBed）：ActiveBed 卡片包含床上的设备（bed_id 有效）和未绑床的设备（bed_id 为 NULL）
//     - 只融合床上的设备（bed_id 有效且相同），未绑床的设备（bed_id 为 NULL）不参与融合
//   - 场景 B（门牌下有多个 ActiveBed）：ActiveBed 卡片只包含床上的设备（bed_id 有效）
//     - 融合床上的设备（bed_id 有效且相同）
// - Location 卡片：同一卡片上同时有 Radar 和 Sleepace 设备（融合所有设备，bed_id 为 NULL）
// - 所有卡片（ActiveBed 和 Location）都处理其设备数据
// 
// 融合规则：
// - HR/RR：优先 Sleepace，无数据则 Radar
// - 床状态/睡眠状态：优先 Sleepace，无数据则 Radar
// - 姿态数据：直接使用 Radar 数据（不是融合，Sleepace 不提供姿态数据）
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
// ⚠️ 重要依赖：
// - 本函数依赖 PostgreSQL cards 表，需要 wisefido-card-aggregator 服务先创建卡片
// - 通过 GetCardDevices 查询卡片绑定的设备列表（从 cards.devices JSONB 字段）
// - 如果 cards 表为空或卡片不存在，会返回错误
// - 当前 wisefido-card-aggregator 的卡片创建功能还未实现，需要优先实现
//
// 该方法从卡片关联的设备中收集最新数据，并按照优先级规则进行融合。
// 所有卡片（ActiveBed 和 Location）都处理其设备数据。
// 
// 融合条件：
// - 同一张卡片上同时有 Radar 和 Sleepace 设备（两者都提供 HR/RR，需要选择更好的数据源）
// 
// 融合规则（仅当同时有 Radar 和 Sleepace 时）：
// 1. HR/RR（心率/呼吸率）：优先使用 Sleepace 数据，如果 Sleepace 无数据则使用 Radar 数据
// 2. 床状态/睡眠状态：优先使用 Sleepace 数据，如果 Sleepace 无数据则使用 Radar 数据
// 
// 非融合情况：
// - 如果只有 Radar 或只有 Sleepace：直接使用该设备数据
// - 姿态数据：直接使用 Radar 数据（不是融合，Sleepace 不提供姿态数据）
// 
// 参数:
//   - tenantID: 租户 ID（UUID 格式）
//   - cardID: 卡片 ID（UUID 格式）
//   - cardType: 卡片类型（"ActiveBed" 或 "Location"）
// 
// 返回:
//   - *models.RealtimeData: 融合后的实时数据，包含心率、呼吸率、姿态等信息
//   - error: 如果融合过程中发生错误（如设备查询失败、数据获取失败等）
func (f *SensorFusion) FuseCardData(tenantID, cardID, cardType string) (*models.RealtimeData, error) {
	// 1. 获取卡片关联的所有设备
	devices, err := f.cardRepo.GetCardDevices(cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card devices: %w", err)
	}
	
	if len(devices) == 0 {
		return nil, fmt.Errorf("no devices found for card: %s", cardID)
	}
	
	// 2. 过滤设备类型和绑定关系：只查询 Radar 和 Sleepace 设备（其他设备不参与融合）
	// 融合规则：
	// - ActiveBed 卡片：只融合绑定到同一床上的设备（bed_id 有效且相同）
	//   - 如果 bed_id 有效，则所有 bed_id 相同的设备都是绑在同一床上的，应该融合
	//   - 如果 bed_id 为 NULL，则不参与融合（未绑床的设备）
	// - Location 卡片：融合所有设备（因为它们都是未绑床的设备，bed_id 为 NULL）
	var fusionDeviceIDs []string
	var bedIDForFusion *string // 用于 ActiveBed 卡片，记录第一个有效 bed_id
	
	for _, device := range devices {
		deviceType := device.DeviceType
		if deviceType == "Radar" || deviceType == "Sleepace" || deviceType == "SleepPad" {
			if cardType == "ActiveBed" {
				// ActiveBed 卡片：只融合绑定到同一床上的设备
				if device.BedID != nil && *device.BedID != "" {
					// 如果这是第一个有效 bed_id，记录它
					if bedIDForFusion == nil {
						bedIDForFusion = device.BedID
					}
					// 只融合 bed_id 相同的设备（绑定到同一床上的设备）
					if bedIDForFusion != nil && *device.BedID == *bedIDForFusion {
						fusionDeviceIDs = append(fusionDeviceIDs, device.DeviceID)
					}
				}
				// bed_id 为 NULL 的设备不参与融合（未绑床的设备）
			} else {
				// Location 卡片：融合所有设备（因为它们都是未绑床的设备，bed_id 为 NULL）
				fusionDeviceIDs = append(fusionDeviceIDs, device.DeviceID)
			}
		}
	}
	
	if len(fusionDeviceIDs) == 0 {
		return nil, fmt.Errorf("no Radar or Sleepace devices found for card: %s", cardID)
	}
	
	// 3. 批量获取 Radar 和 Sleepace 设备的最新数据（优化 N+1 查询）
	// 使用批量查询（每个设备获取最新1条数据）
	deviceDataMap, err := f.iotRepo.GetLatestByDeviceIDs(tenantID, fusionDeviceIDs, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest data for devices: %w", err)
	}
	
	// 4. 收集 Radar 和 Sleepace 设备的最新数据，并找到最大时间戳
	var sleepaceData []*models.IoTTimeSeries
	var radarData []*models.IoTTimeSeries
	var maxTimestamp time.Time
	
	// 创建设备ID到设备信息的映射（用于获取设备类型）
	deviceMap := make(map[string]*repository.DeviceInfo)
	for _, device := range devices {
		deviceType := device.DeviceType
		if deviceType == "Radar" || deviceType == "Sleepace" || deviceType == "SleepPad" {
			deviceMap[device.DeviceID] = &device
		}
	}
	
	for _, deviceID := range fusionDeviceIDs {
		latestData, ok := deviceDataMap[deviceID]
		if !ok || len(latestData) == 0 {
			f.logger.Warn("No data found for device",
				zap.String("device_id", deviceID),
			)
			continue
		}
		
		data := latestData[0]
		
		// 更新最大时间戳（使用数据的时间戳，而不是当前时间）
		if data.Timestamp.After(maxTimestamp) {
			maxTimestamp = data.Timestamp
		}
		
		// 确定设备类型（优先使用数据中的类型，否则使用设备信息中的类型）
		deviceType := data.DeviceType
		if deviceType == "" {
			if deviceInfo, ok := deviceMap[deviceID]; ok {
				deviceType = deviceInfo.DeviceType
			} else {
				// 降级方案：从设备表查询
				deviceTypeFromDB, err := f.iotRepo.GetDeviceType(tenantID, deviceID)
				if err != nil {
					f.logger.Warn("Failed to get device type",
						zap.String("device_id", deviceID),
						zap.Error(err),
					)
					continue
				}
				deviceType = deviceTypeFromDB
			}
		}
		
		// 分类数据
		if deviceType == "Sleepace" || deviceType == "SleepPad" {
			sleepaceData = append(sleepaceData, data)
		} else if deviceType == "Radar" {
			radarData = append(radarData, data)
		}
	}
	
	// 5. 判断是否需要融合 HR/RR 和床状态
	// 融合条件：
	// - ActiveBed 卡片：同一床上同时有 Radar 和 Sleepace 设备（只融合 bed_id 有效且相同的设备）
	// - Location 卡片：同一卡片上同时有 Radar 和 Sleepace 设备（融合所有设备，bed_id 为 NULL）
	// 注意：Location 卡片上的设备可能监测不同的人，理论上不应该融合 HR/RR
	// 但为了统一处理逻辑，如果 Location 卡片上同时有 Radar 和 Sleepace，也进行融合
	// （实际业务中，Location 卡片通常不会同时绑定 Radar 和 Sleepace）
	needFusion := len(sleepaceData) > 0 && len(radarData) > 0
	
	// 使用数据的时间戳（如果没有任何数据，使用当前时间作为降级）
	resultTimestamp := time.Now().Unix()
	if !maxTimestamp.IsZero() {
		resultTimestamp = maxTimestamp.Unix()
	}
	
	result := &models.RealtimeData{
		Timestamp: resultTimestamp, // 使用数据的时间戳，而不是 time.Now()
		Postures:  []models.Posture{},
	}
	
	// 6. 处理数据（所有卡片都处理其设备数据）
	if needFusion {
		// 需要融合：同一张卡片上同时有 Radar 和 Sleepace 设备
		// 选择更好的 HR/RR 和床状态数据源（优先 Sleepace）
		f.fuseVitalSigns(sleepaceData, radarData, result)
		f.fuseBedAndSleepStatus(sleepaceData, radarData, result)
	} else {
		// 不需要融合：直接使用设备数据
		// 如果只有 Sleepace，使用 Sleepace 数据
		if len(sleepaceData) > 0 {
			f.useSingleDeviceData(sleepaceData[0], result)
		}
		// 如果只有 Radar，使用 Radar 数据
		// 注意：对于 Location 卡片，如果有多个 Radar，只使用第一个 Radar 的 HR/RR/床状态
		if len(radarData) > 0 {
			f.useRadarDeviceData(radarData[0], result)
		}
	}
	
	// 7. 处理姿态数据（直接使用 Radar 数据，不是融合）
	// 注意：Sleepace 不提供姿态数据，所以直接使用 Radar 数据
	f.useRadarPostures(radarData, result)
	
	return result, nil
}

// useSingleDeviceData 使用单个 Sleepace 设备的数据（不需要融合的情况）
func (f *SensorFusion) useSingleDeviceData(data *models.IoTTimeSeries, result *models.RealtimeData) {
	if data.HeartRate != nil {
		result.Heart = data.HeartRate
		result.HeartSource = "Sleepace"
		timestamp := data.Timestamp.Unix()
		result.HeartTimestamp = &timestamp
	}
	if data.RespiratoryRate != nil {
		result.Breath = data.RespiratoryRate
		result.BreathSource = "Sleepace"
		timestamp := data.Timestamp.Unix()
		result.BreathTimestamp = &timestamp
	}
	if data.BedStatusSNOMEDCode != nil {
		result.BedStatus = data.BedStatusSNOMEDCode
		result.BedStatusSource = "Sleepace"
		timestamp := data.Timestamp.Unix()
		result.BedStatusTimestamp = &timestamp
	}
	if data.SleepStateSNOMEDCode != nil {
		result.SleepStage = data.SleepStateSNOMEDCode
		result.SleepStageSource = "Sleepace"
		timestamp := data.Timestamp.Unix()
		result.SleepStageTimestamp = &timestamp
	}
}

// useRadarDeviceData 使用单个 Radar 设备的数据（不需要融合的情况）
func (f *SensorFusion) useRadarDeviceData(data *models.IoTTimeSeries, result *models.RealtimeData) {
	if data.HeartRate != nil {
		result.Heart = data.HeartRate
		result.HeartSource = "Radar"
		timestamp := data.Timestamp.Unix()
		result.HeartTimestamp = &timestamp
	}
	if data.RespiratoryRate != nil {
		result.Breath = data.RespiratoryRate
		result.BreathSource = "Radar"
		timestamp := data.Timestamp.Unix()
		result.BreathTimestamp = &timestamp
	}
	if data.BedStatusSNOMEDCode != nil {
		result.BedStatus = data.BedStatusSNOMEDCode
		result.BedStatusSource = "Radar"
		timestamp := data.Timestamp.Unix()
		result.BedStatusTimestamp = &timestamp
	}
	if data.SleepStateSNOMEDCode != nil {
		result.SleepStage = data.SleepStateSNOMEDCode
		result.SleepStageSource = "Radar"
		timestamp := data.Timestamp.Unix()
		result.SleepStageTimestamp = &timestamp
	}
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
				timestamp := data.Timestamp.Unix()
				result.HeartTimestamp = &timestamp
				break
			}
		}
		for _, data := range sleepaceData {
			if data.RespiratoryRate != nil {
				result.Breath = data.RespiratoryRate
				result.BreathSource = "Sleepace"
				timestamp := data.Timestamp.Unix()
				result.BreathTimestamp = &timestamp
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
				timestamp := data.Timestamp.Unix()
				result.HeartTimestamp = &timestamp
				break
			}
		}
	}
	if result.Breath == nil && len(radarData) > 0 {
		for _, data := range radarData {
			if data.RespiratoryRate != nil {
				result.Breath = data.RespiratoryRate
				result.BreathSource = "Radar"
				timestamp := data.Timestamp.Unix()
				result.BreathTimestamp = &timestamp
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
				result.BedStatusSource = "Sleepace"
				timestamp := data.Timestamp.Unix()
				result.BedStatusTimestamp = &timestamp
				break
			}
		}
		for _, data := range sleepaceData {
			if data.SleepStateSNOMEDCode != nil {
				result.SleepStage = data.SleepStateSNOMEDCode
				result.SleepStageSource = "Sleepace"
				timestamp := data.Timestamp.Unix()
				result.SleepStageTimestamp = &timestamp
				break
			}
		}
	}
	
	// 如果 Sleepace 没有数据，使用 Radar 数据（如果有）
	if result.BedStatus == nil && len(radarData) > 0 {
		for _, data := range radarData {
			if data.BedStatusSNOMEDCode != nil {
				result.BedStatus = data.BedStatusSNOMEDCode
				result.BedStatusSource = "Radar"
				timestamp := data.Timestamp.Unix()
				result.BedStatusTimestamp = &timestamp
				break
			}
		}
	}
	if result.SleepStage == nil && len(radarData) > 0 {
		for _, data := range radarData {
			if data.SleepStateSNOMEDCode != nil {
				result.SleepStage = data.SleepStateSNOMEDCode
				result.SleepStageSource = "Radar"
				timestamp := data.Timestamp.Unix()
				result.SleepStageTimestamp = &timestamp
				break
			}
		}
	}
}

// useRadarPostures 使用 Radar 设备的姿态数据
// 注意：这不是"融合"，而是直接使用 Radar 数据（Sleepace 不提供姿态数据）
// 规则：收集所有 Radar 设备的 tracking_id（不跨设备去重）
// 如果同一个 tracking_id 有多条记录，使用时间戳最新的
func (f *SensorFusion) useRadarPostures(
	radarData []*models.IoTTimeSeries,
	result *models.RealtimeData,
) {
	// 收集所有 Radar 设备的姿态数据
	// key: tracking_id, value: 姿态数据和时间戳
	trackingMap := make(map[string]struct {
		posture  *models.Posture
		timestamp time.Time
	})
	
	for _, data := range radarData {
		if data.TrackingID != nil && data.PostureSNOMEDCode != nil {
			trackingID := *data.TrackingID
			
			posture := &models.Posture{
				TrackingID:    trackingID,
				PostureCode:   *data.PostureSNOMEDCode,
				PostureDisplay: "",
			}
			
			if data.PostureDisplay != nil {
				posture.PostureDisplay = *data.PostureDisplay
			}
			
			// 如果该 tracking_id 已存在，比较时间戳，使用更新的数据
			if existing, ok := trackingMap[trackingID]; ok {
				// 如果当前数据的时间戳更新，则替换
				if data.Timestamp.After(existing.timestamp) {
					trackingMap[trackingID] = struct {
						posture  *models.Posture
						timestamp time.Time
					}{
						posture:  posture,
						timestamp: data.Timestamp,
					}
				}
			} else {
				// 首次出现，直接添加
				trackingMap[trackingID] = struct {
					posture  *models.Posture
					timestamp time.Time
				}{
					posture:  posture,
					timestamp: data.Timestamp,
				}
			}
		}
	}
	
	// 转换为列表
	for _, entry := range trackingMap {
		result.Postures = append(result.Postures, *entry.posture)
	}
	
	result.PersonCount = len(result.Postures)
}


