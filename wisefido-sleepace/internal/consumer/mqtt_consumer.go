package consumer

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.uber.org/zap"

	"owl-common/alarm"
	"owl-common/card"
	"owl-common/observation"
	"owl-common/redis"

	"wisefido-sleepace/internal/service"
)

// ReceivedMessage is the envelope from sleepace-service MQTT push.
type ReceivedMessage struct {
	DataKey   string          `json:"dataKey"`
	TimeStamp int64           `json:"timeStamp"`
	DeviceID  string          `json:"deviceId"`
	Data      json.RawMessage `json:"data"`
}

// ------ per-dataKey data structs ------

type ConnectionStatusData struct {
	ConnectionStatus int `json:"connectionStatus"`
}

// RealtimeData Sleepace realtime 单条。breath/heart 255=初始化；initStatus 1 时 breath/heart/turnOver/bodyMove/sitUp/bedStatus 均无效。temp/hum 当前无此功能，仅解析不取值。
type RealtimeData struct {
	LeftRight     int `json:"leftRight"`
	Breath        int `json:"breath"`        // 呼吸率，255=初始化
	Heart         int `json:"heart"`         // 心率，255=初始化
	TurnOver      int `json:"turnOver"`      // 翻身（BM8701-2 固件≤2.x）
	BodyMove      int `json:"bodyMove"`      // 体动（BM8701-2 ≥5.x 及 M901L）
	SitUp         int `json:"sitUp"`         // 0 非坐起 1 坐起 8 初始化/未变
	InitStatus    int `json:"initStatus"`    // 0 正常 1 设备初始化中（此时其他体征/床态无效）
	BedStatus     int `json:"bedStatus"`     // 厂家：0=未离床 1=离床 8=初始化或状态未变(不触发事件)
	SignalQuality int `json:"signalQuality"` // 0-5，5 最好（BM8701-2 ≥v6.22）
	Temp          int `json:"temp"`          // 当前无此功能，不取值
	Hum           int `json:"hum"`           // 当前无此功能，不取值
}

type SleepStageData struct {
	LeftRight  int `json:"leftRight"`
	SleepStage int `json:"sleepStage"`
}

type SensorData struct {
	Status int `json:"status"`
}

type AnalysisData struct {
	UserId    string `json:"userId"`
	StartTime int64  `json:"startTime"`
}

type UpgradeProgressData struct {
	Length int `json:"length"`
	Offset int `json:"offset"`
}

type AlarmNotifyData struct {
	Id            int64  `json:"id"`
	Type          string `json:"type"`
	UserId        string `json:"userId"`
	Status        int    `json:"status"`
	RelieveReason string `json:"relieveReason"`
	RelieveTime   int64  `json:"relieveTime"`
}

type InBedStatusData struct {
	LeftRight   int `json:"leftRight"`
	InbedStatus int `json:"inbedStatus"`
}

// Sleepace 取值约定：
// breath/heart: 255=初始化状态，不写入 stream。
// bedStatus/inbedStatus 厂家定义：0=未离床 1=离床；8=情况一(sitUp/bedStatus/sleepStage 全8)算法初始化，情况二(该字段8其余不全8)状态未变保持前一有效值。BM8701/M901L 建议用 bedStatus 查在离床。收到 8 不写 stream、不触发 InBed/LeftBed。
// sitUp/sleepStage: 0/1 同上语义；8 同上。收到 8 不写入 stream。
// initStatus: 0 正常 1 设备初始化中（breath/heart/turnOver/bodyMove/sitUp/bedStatus 均无效，仅写 initStatus 不写 track/bed）。
// signalQuality: 0-5 原始，网关归一化 0-100% 写入。
const (
	sleepaceStatusInvalid = 8   // 床态/坐起/睡眠阶段 8=无效或未变
	sleepaceVitalInvalid  = 255 // 心率/呼吸率 255=初始化
	sleepaceTrackPoseConf = 90  // 在床/坐起翻身时 TrackConfidence、PoseConfidence 固定值
)

// MQTTConsumer consumes messages from the sleepace MQTT topic,
// transforms them into observation.Messages, and publishes to Redis Streams.
type MQTTConsumer struct {
	publisher     *StreamPublisher
	reportService *service.ReportService
	statusTracker *service.DeviceStatusTracker
	logger        *zap.Logger

	cardDB      *card.CardDB         // 可选，用于首次上连时更新 device_store 版本
	sleepaceAPI *service.SleepaceAPI // 可选，用于首次上连时拉取 deviceVersion

	msgCh  chan *rawMsg
	wg     sync.WaitGroup
	cancel context.CancelFunc

	// realtime 里 bedStatus 实为状态量，仅在变化时发 InBed/LeftBed 到 event
	lastBedStatusMu sync.Mutex
	lastBedStatus   map[string]int // key = deviceUID+":"+leftRight, value 0 或 1

	realtimeLogCount atomic.Uint64 // 每 10 条 realtime 打一条 Info 日志
}

type rawMsg struct {
	topic   string
	payload []byte
}

func NewMQTTConsumer(publisher *StreamPublisher, reportSvc *service.ReportService, statusTracker *service.DeviceStatusTracker, logger *zap.Logger) *MQTTConsumer {
	return &MQTTConsumer{
		publisher:     publisher,
		reportService: reportSvc,
		statusTracker: statusTracker,
		logger:        logger,
		msgCh:         make(chan *rawMsg, 1000),
		lastBedStatus: make(map[string]int),
	}
}

// SetDeviceVersionSync 设置首次上连时同步 device_store 版本所需依赖（cardDB + sleepaceAPI）
func (c *MQTTConsumer) SetDeviceVersionSync(cardDB *card.CardDB, sleepaceAPI *service.SleepaceAPI) {
	c.cardDB = cardDB
	c.sleepaceAPI = sleepaceAPI
}

// syncDeviceStoreVersion 首次上连时拉取 deviceInfo。MQTT 的 deviceId 始终为 device_code，用其查厂家 API 得到 Version，与 device_store 比对并更新。
func (c *MQTTConsumer) syncDeviceStoreVersion(ctx context.Context, deviceCode string) {
	if deviceCode == "" {
		return
	}
	info, err := c.sleepaceAPI.GetDeviceInfoByDeviceId(deviceCode)
	if err != nil {
		c.logger.Debug("get device info for version sync", zap.String("device_code", deviceCode), zap.Error(err))
		return
	}
	if info == nil || info.Version == "" {
		return
	}
	if err := c.cardDB.UpdateDeviceStoreReportedVersion(ctx, deviceCode, info.Version); err != nil {
		c.logger.Warn("update device_store reported version", zap.String("device_code", deviceCode), zap.String("version", info.Version), zap.Error(err))
		return
	}
	c.logger.Info("device_store version synced", zap.String("device_code", deviceCode), zap.String("version", info.Version))
}

// MessageHandler returns the mqtt.MessageHandler to be used with mqtt.Client.Subscribe.
func (c *MQTTConsumer) MessageHandler() mqtt.MessageHandler {
	return func(_ mqtt.Client, msg mqtt.Message) {
		c.msgCh <- &rawMsg{topic: msg.Topic(), payload: msg.Payload()}
	}
}

// Start launches worker goroutines.
func (c *MQTTConsumer) Start(ctx context.Context, workers int) {
	ctx, c.cancel = context.WithCancel(ctx)
	for i := 0; i < workers; i++ {
		c.wg.Add(1)
		go c.worker(ctx)
	}
}

func (c *MQTTConsumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
}

func (c *MQTTConsumer) worker(ctx context.Context) {
	defer c.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c.msgCh:
			c.handleMessage(ctx, msg)
		}
	}
}

// handleMessage 解析 MQTT payload：支持数组（事件/实时/deviceSenSor 等多条）与单条（报警默认 single）。逐条 dispatch。
func (c *MQTTConsumer) handleMessage(ctx context.Context, msg *rawMsg) {
	var batch []*ReceivedMessage
	if err := json.Unmarshal(msg.payload, &batch); err != nil {
		// 报警默认 single 时为单对象，按单条解析后包装为单元素切片
		var single ReceivedMessage
		if err2 := json.Unmarshal(msg.payload, &single); err2 != nil {
			c.logger.Error("unmarshal mqtt payload (array or single)", zap.Error(err), zap.Error(err2))
			return
		}
		batch = []*ReceivedMessage{&single}
	}
	for _, m := range batch {
		c.dispatch(ctx, m)
	}
}

// dispatch routes a single ReceivedMessage by constructing IoTStreamMessage and publishing to iot: streams.
// MQTT 的 deviceId 首次可能为 device_uid、后续可能为 device_code，统一传 Resolve(deviceKey) 由 DB 解析；三元组由 DeviceCardMapping 带回，未命中时用 m.DeviceID 兜底 deviceCode。
func (c *MQTTConsumer) dispatch(ctx context.Context, m *ReceivedMessage) {
	tenantID, _, _, _, deviceID, deviceUID, deviceCode, deviceType := c.publisher.Resolve(ctx, m.DeviceID)
	if deviceCode == "" {
		deviceCode = m.DeviceID
	}
	if deviceType == "" {
		deviceType = "Sleepad" // 与 card_mapping.sleepadDeviceTypes 一致，勿改为 SleepPad/sleepad 以外
	}
	// iot: 流用 deviceID (UUID)、deviceType；未入 device_store 时 deviceID 为空
	ts := toMillis(m.TimeStamp)
	nowMs := time.Now().UnixMilli() // 仅 IotHead（stream 消息时间戳）用当前时间，避免设备 ts 过旧导致 cardagg 判 stale 丢弃；payload 内 EventSince/event_since 仍用 ts

	switch m.DataKey {
	// 明确报警项 → alarm：connectionStatus(离线/恢复)、deviceSenSor(SensorDetached)、alarmNotify
	case "connectionStatus":
		var d ConnectionStatusData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal connectionStatus", zap.Error(err))
			return
		}
		online := d.ConnectionStatus != 0
		// 导出 MQTT 格式：重启 Sleepad 后看此条日志即可（deviceId=MQTT 里的设备标识，data=connectionStatus 载荷）
		//c.logger.Info("connectionStatus",
		//	zap.String("device_uid", deviceUID),
		//	zap.Bool("online", online),
		//	zap.String("mqtt_deviceId", m.DeviceID),
		//	zap.String("mqtt_dataKey", m.DataKey),
		//	zap.ByteString("mqtt_data", m.Data))
		c.logger.Info("connectionStatus", zap.String("device_uid", deviceUID), zap.Bool("online", online))
		if c.statusTracker != nil {
			c.statusTracker.UpdateConnection(deviceUID, online)
		}
		// 首次上连（在线）时拉取 deviceVersion，与 device_store.firmware_version 比对，不一致则写入 ota_target_firmware_version
		if online && c.cardDB != nil && c.sleepaceAPI != nil {
			go c.syncDeviceStoreVersion(context.Background(), deviceCode)
		}
		// iot:alarm 仅使用 device_id (UUID)；未入 device_store 则不发布
		if deviceID == "" {
			c.logger.Debug("connectionStatus skip alarm (device not in device_store)", zap.String("device_uid", deviceUID))
		} else {
			tsMs := toMillis(ts)
			var eventName string
			var item observation.EventItem
			if online {
				eventName = alarm.AlarmTypeDeviceRecover
				item = observation.EventItem{
					DataCategory: eventName,
					EventName:    eventName,
					EventSince:   tsMs,
					EventStatus:  "start",
					TrackID:      observation.TrackDevice,
				}
			} else {
				eventName = alarm.AlarmTypeOffline
				item = observation.EventItem{
					DataCategory: eventName,
					EventName:    eventName,
					EventSince:   tsMs,
					EventStatus:  "start",
					TrackID:      observation.TrackDevice,
				}
			}
			data, _ := observation.EventItemToDataMap(&item)
			if data == nil {
				data = make(map[string]any)
			}
			data[observation.FieldEventName] = eventName
			data[redis.DataCategoryKey] = eventName
			if !online {
				data[observation.FieldOffline] = 1
			}
			msg := redis.NewIoTStreamMessageWithData(tenantID, "", deviceUID, deviceID, deviceType, nowMs, "alarm", eventName, data)
			err := c.publisher.PublishAlarm(ctx, msg)
			if err != nil {
				c.logger.Warn("connectionStatus publish alarm failed", zap.String("device_uid", deviceUID), zap.Bool("online", online), zap.Error(err))
			}
		}
	// realtime/实时：订阅接口约定为数组推送，handleMessage 已逐条 dispatch；此处每条对应一个 (deviceId, data)，data 内含 leftRight。
	case "realtime":
		if deviceID == "" {
			return
		}
		var d RealtimeData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal realtime", zap.Error(err))
			return
		}
		// signal_quality 0-5 → VitalConfidence=18*signal_quality（精细压力）；TrackConfidence/PoseConfidence=90（在床/坐起翻身，较大压力变化）
		vitalConf := d.SignalQuality * 18
		if vitalConf > sleepaceTrackPoseConf {
			vitalConf = sleepaceTrackPoseConf
		}

		if c.statusTracker != nil {
			c.statusTracker.UpdateRealtime(deviceUID, d.Heart, d.Breath, d.BedStatus, d.TurnOver, d.BodyMove, d.SitUp, d.InitStatus, vitalConf)
		}
		var data map[string]any
		if d.InitStatus == 0 {
			// 按 track 字段一一赋值
			data = make(map[string]any)
			data[observation.FieldTrackID] = d.LeftRight
			if d.Heart > 0 && d.Heart < sleepaceVitalInvalid {
				data[observation.FieldHeartRate] = d.Heart
			}
			if d.Breath > 0 && d.Breath < sleepaceVitalInvalid {
				data[observation.FieldRespiratoryRate] = d.Breath
			}
			data[observation.FieldBodyMove] = d.BodyMove
			data[observation.FieldTurnOver] = d.TurnOver
			if d.SitUp == 1 {
				data[observation.FieldPose] = observation.PoseBedSitUp
			}
			if d.BedStatus < sleepaceStatusInvalid {
				data[observation.FieldBedStatus] = d.BedStatus
			}
			data[observation.FieldTrackConfidence] = sleepaceTrackPoseConf
			data[observation.FieldPoseConfidence] = sleepaceTrackPoseConf
			data[observation.FieldVitalConfidence] = vitalConf
			data[observation.FieldInitStatus] = d.InitStatus
			if vitalConf > 0 {
				data[observation.FieldSignalQuality] = vitalConf
			}
		} else {
			data = map[string]any{observation.FieldInitStatus: d.InitStatus}
			if vitalConf > 0 {
				data[observation.FieldSignalQuality] = vitalConf
			}
		}
		msg := redis.NewIoTStreamMessageWithData(tenantID, "", deviceUID, deviceID, deviceType, ts, "monitor", observation.CategoryTrack, data)
		c.publisher.PublishMonitor(ctx, msg)
		n := c.realtimeLogCount.Add(1)
		if n%10 == 0 {
			c.logger.Info("realtime received (every 10)", zap.Uint64("count", n), zap.String("device_uid", deviceUID), zap.String("device_id", deviceID), zap.Int("bed_status", d.BedStatus))
		}

		// bedStatus 实为状态量：仅 0/1 参与；8=未变/初始化，不更新 lastBedStatus、不发事件
		if d.InitStatus == 0 && d.BedStatus != sleepaceStatusInvalid {
			key := deviceUID + ":" + strconv.Itoa(d.LeftRight)
			c.lastBedStatusMu.Lock()
			last := c.lastBedStatus[key]
			c.lastBedStatus[key] = d.BedStatus
			c.lastBedStatusMu.Unlock()
			if last != d.BedStatus {
				var categoryBed string
				if d.BedStatus == 0 {
					categoryBed = alarm.InBed // 0=未离床
				} else {
					categoryBed = alarm.LeftBed // 1=离床
				}
				evItem := observation.EventItem{
					DataCategory: categoryBed,
					EventName:    categoryBed,
					EventSince:   ts,
					EventStatus:  "start",
					TrackID:      d.LeftRight,
				}
				evData, _ := observation.EventItemToDataMap(&evItem)
				if evData == nil {
					evData = make(map[string]any)
				}
				evData[observation.FieldBedStatus] = d.BedStatus // int: 0=在床, 1=离床
				evMsg := redis.NewIoTStreamMessageWithData(tenantID, "", deviceUID, deviceID, deviceType, nowMs, "event", categoryBed, evData)
				evMsg.DeviceID = deviceID
				if err := c.publisher.PublishEvent(ctx, evMsg); err != nil {
					c.logger.Warn("publish bedStatus change event", zap.String("device_uid", deviceUID), zap.Int("bedStatus", d.BedStatus), zap.Error(err))
				} else {
					c.logger.Info("bedStatus change → event", zap.String("device_uid", deviceUID), zap.Int("leftRight", d.LeftRight), zap.String("category", categoryBed))
				}
			}
		}

	// inBedStatus：0=未离床 1=离床；8=算法初始化或状态未变，保持前一有效值，不写 stream。
	case "inBedStatus":
		var d InBedStatusData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal inBedStatus", zap.Error(err))
			return
		}
		if d.InbedStatus == 8 {
			return
		}
		categoryInBed := alarm.InBed // 0=未离床
		if d.InbedStatus == 1 {
			categoryInBed = alarm.LeftBed // 1=离床
		}
		data := map[string]any{
			redis.DataCategoryKey:      categoryInBed,
			observation.FieldTrackID:   d.LeftRight,
			observation.FieldBedStatus: d.InbedStatus,
		}
		payloadJSON, _ := json.Marshal(data)
		item := observation.EventItem{
			DataCategory: categoryInBed,
			EventName:    categoryInBed,
			EventSince:   ts,
			EventStatus:  "instant",
			TrackID:      d.LeftRight,
			EventPayload: string(payloadJSON),
		}
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out[observation.FieldBedStatus] = d.InbedStatus
		c.publisher.PublishEvent(ctx, redis.NewIoTStreamMessageWithData(tenantID, "", deviceUID, deviceID, deviceType, nowMs, "event", categoryInBed, out))
		c.logger.Info("inBedStatus", zap.String("category", categoryInBed), zap.String("device_uid", deviceUID), zap.Int("status", d.InbedStatus), zap.Int("lr", d.LeftRight))

	case "sleepStage":
		if deviceID == "" {
			return
		}
		var d SleepStageData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal sleepStage", zap.Error(err))
			return
		}
		// 统一 Sleepad 标准：1=awake 2=light 4=deep 8=unknown/初始；1/2/4/8 都发（8 表设备在线）
		stage := d.SleepStage
		if stage != 1 && stage != 2 && stage != 4 && stage != 8 {
			stage = 8
		}
		if c.statusTracker != nil {
			c.statusTracker.UpdateSleepStage(deviceUID, stage)
		}
		payloadData := map[string]any{
			redis.DataCategoryKey:       alarm.SleepStage,
			"event_name":                alarm.SleepStage,
			"event_since":               ts,
			"event_status":              "instant",
			"track_id":                  d.LeftRight,
			observation.FieldSleepStage: stage,
		}
		payloadJSON, _ := json.Marshal(payloadData)
		item := observation.EventItem{
			DataCategory: alarm.SleepStage,
			EventName:    alarm.SleepStage,
			EventSince:   ts,
			EventStatus:  "instant",
			TrackID:      d.LeftRight,
			EventPayload: string(payloadJSON),
		}
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out[observation.FieldSleepStage] = stage
		msg := redis.NewIoTStreamMessageWithData(tenantID, "", deviceUID, deviceID, deviceType, nowMs, "event", alarm.SleepStage, out)
		c.publisher.PublishEvent(ctx, msg)
		c.logger.Info("sleepStage", zap.String("category", alarm.SleepStage), zap.String("device_uid", deviceUID), zap.Int("stage", stage), zap.Int("lr", d.LeftRight))

		// deviceSenSor：文档 0=脱落 1=已经插上。每条消息可能包含多条在床状态记录，顶层为数组；handleMessage 已按数组逐条 dispatch，此处每条为单条 data { status }。EventItem + EventPayload。
	case "deviceSenSor":
		if deviceID == "" {
			return
		}
		var d SensorData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal deviceSenSor", zap.Error(err), zap.ByteString("raw_data", m.Data))
			return
		}
		// 排查误报：记录原始 payload 与解析出的 status（若设备/中间层字段名或类型不一致会导致 status=0 误触 SensorDetached）
		c.logger.Info("deviceSenSor", zap.String("device_uid", deviceUID), zap.Int("status", d.Status), zap.ByteString("raw_data", m.Data))
		failure := d.Status == 0 // 0=脱落 → 告警；1=已插上 → 恢复（仅做恢复，不产生新告警）
		if c.statusTracker != nil {
			c.statusTracker.UpdateSensor(deviceUID, failure, false)
		}
		failureVal := 0
		if failure {
			failureVal = 1
		}
		// status=1 必须发 SensorDetachedRecover，cardagg 只做 auto_resolve，不落新 SensorDetached
		eventName := alarm.SensorDetached
		if !failure {
			eventName = alarm.SensorDetachedRecover
		}
		eventStatus := "start"
		if !failure {
			eventStatus = "end"
		}
		payloadData := map[string]any{
			redis.DataCategoryKey:          alarm.SensorDetached,
			"event_name":                   eventName,
			"event_since":                  ts,
			"event_status":                 eventStatus,
			"track_id":                     observation.TrackDevice,
			observation.FieldDeviceFailure: failureVal,
		}
		payloadJSON, _ := json.Marshal(payloadData)
		item := observation.EventItem{
			DataCategory: alarm.SensorDetached,
			EventName:    eventName,
			EventSince:   ts,
			EventStatus:  eventStatus,
			TrackID:      observation.TrackDevice,
			EventPayload: string(payloadJSON),
		}
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out[observation.FieldDeviceFailure] = failureVal
		msg := redis.NewIoTStreamMessageWithData(tenantID, "", deviceUID, deviceID, deviceType, nowMs, "alarm", eventName, out)
		c.publisher.PublishAlarm(ctx, msg)

	case "pressureSenSor":
		if deviceID == "" {
			return
		}
		var d SensorData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal pressureSenSor", zap.Error(err))
			return
		}
		detached := d.Status != 0
		if c.statusTracker != nil {
			c.statusTracker.UpdateSensor(deviceUID, false, detached)
		}
		detachedVal := 0
		if detached {
			detachedVal = 1
		}
		eventStatus := "start"
		if !detached {
			eventStatus = "end"
		}
		payloadData := map[string]any{
			"dataCategory":            "pressureSensor",
			"event_name":              "pressureSensor",
			"event_since":             ts,
			"event_status":            eventStatus,
			"track_id":                observation.TrackDevice,
			observation.FieldDetached: detachedVal,
		}
		payloadJSON, _ := json.Marshal(payloadData)
		item := observation.EventItem{
			DataCategory: "pressureSensor",
			EventName:    "pressureSensor",
			EventSince:   ts,
			EventStatus:  eventStatus,
			TrackID:      observation.TrackDevice,
			EventPayload: string(payloadJSON),
		}
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out[observation.FieldDetached] = detachedVal
		msg := redis.NewIoTStreamMessageWithData(tenantID, "", deviceUID, deviceID, deviceType, nowMs, "event", "pressureSensor", out)
		c.publisher.PublishEvent(ctx, msg)

	case "analysis":
		var d AnalysisData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal analysis", zap.Error(err))
			return
		}
		payloadData := map[string]any{
			"dataCategory": "analysis",
			"event_name":   "analysis",
			"event_since":  ts,
			"event_status": "instant",
			"user_id":      d.UserId,
			"start_time":   d.StartTime,
		}
		payloadJSON, _ := json.Marshal(payloadData)
		item := observation.EventItem{
			DataCategory: "analysis",
			EventName:    "analysis",
			EventSince:   ts,
			EventStatus:  "instant",
			TrackID:      observation.TrackDevice,
			EventPayload: string(payloadJSON),
		}
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		if deviceID != "" {
			msg := redis.NewIoTStreamMessageWithData(tenantID, "", deviceUID, deviceID, deviceType, nowMs, "event", "analysis", out)
			c.publisher.PublishEvent(ctx, msg)
		}
		if c.reportService != nil {
			go func() {
				if err := c.reportService.DownloadAndSave(ctx, deviceCode, d.UserId, d.StartTime+1, m.TimeStamp); err != nil {
					c.logger.Error("download report failed", zap.String("device_code", deviceCode), zap.Error(err))
				}
			}()
		}

	case "upgradeProgress":
		if deviceID == "" {
			return
		}
		var d UpgradeProgressData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal upgradeProgress", zap.Error(err))
			return
		}
		progress := 0
		if d.Length > 0 {
			progress = d.Offset * 100 / d.Length
		}
		payloadData := map[string]any{
			"dataCategory":                   "upgradeProgress",
			"event_name":                     "upgradeProgress",
			"event_since":                    ts,
			"event_status":                   "instant",
			"track_id":                       observation.TrackDevice,
			observation.FieldCommandAction:   "firmware_upgrade",
			observation.FieldCommandProgress: progress,
			"length":                         d.Length,
			"offset":                         d.Offset,
		}
		payloadJSON, _ := json.Marshal(payloadData)
		item := observation.EventItem{
			DataCategory: "upgradeProgress",
			EventName:    "upgradeProgress",
			EventSince:   ts,
			EventStatus:  "instant",
			TrackID:      observation.TrackDevice,
			EventPayload: string(payloadJSON),
		}
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out[observation.FieldCommandAction] = "firmware_upgrade"
		out[observation.FieldCommandProgress] = progress
		msg := redis.NewIoTStreamMessageWithData(tenantID, "", deviceUID, deviceID, deviceType, nowMs, "event", "upgradeProgress", out)
		c.publisher.PublishEvent(ctx, msg)

	// 明确报警项：设备端已判定并上报告警，发 iot:alarm:stream；其余（InBed/LeftBed、SleepStage、pressureSensor 等）发 event，由后端判断。
	// alarmNotify：默认 single 单条，可通过 pushType 接口设 alarmDataType=array；handleMessage 已兼容单条/数组，此处每条为单条 data。
	case "alarmNotify":
		if deviceID == "" {
			return
		}
		var d AlarmNotifyData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal alarmNotify", zap.Error(err))
			return
		}
		c.handleAlarmNotify(ctx, tenantID, deviceID, deviceUID, deviceType, ts, nowMs, &d)

	default:
		c.logger.Warn("unknown dataKey", zap.String("dataKey", m.DataKey), zap.String("device_uid", deviceUID))
	}
}

type alarmMapping struct {
	EventName string
	TrackID   int
}

var mqttAlarmMap = map[string]alarmMapping{
	"alarmLeftBed":         {alarm.LeftBed, observation.TrackUnknownPerson},
	"alarmHeartRateFast":   {alarm.HeartRateAlertHigh, observation.TrackUnknownPerson},
	"alarmHeartRateSlow":   {alarm.HeartRateAlertLow, observation.TrackUnknownPerson},
	"alarmBreathRateFast":  {alarm.RespRateAlertHigh, observation.TrackUnknownPerson},
	"alarmBreathRateSlow":  {alarm.RespRateAlertLow, observation.TrackUnknownPerson},
	"alarmBreathRatePause": {alarm.ApneaHypopnea, observation.TrackUnknownPerson},
	"alarmBodymove":        {alarm.AbnormalBodyMovement, observation.TrackUnknownPerson},
	"alarmNoBodymove":      {alarm.NoBodyMove, observation.TrackUnknownPerson},
	"alarmNoTurnOver":      {alarm.NoTurnOver, observation.TrackUnknownPerson},
	"alarmBedSitup":        {alarm.BedSitUp, observation.TrackUnknownPerson},
	"alarmSitup":           {alarm.BedSitUp, observation.TrackUnknownPerson}, // 疑似坐起
	"alarmInBed":           {alarm.InBed, observation.TrackSpace},
	"alarmOnBed":           {alarm.InBed, observation.TrackSpace}, // 在床报警
	"alarmSensorFall":      {alarm.SensorDetached, observation.TrackDevice},
}

// handleAlarmNotify 处理 Sleepace 明确上报告警（alarmNotify），发布到 iot:alarm:stream，cardagg 直接落库。 DataKey:alarmNotify。tsPayload 用于 payload 内 EventSince，nowMsEnvelope 用于 IotHead 消息时间戳。
func (c *MQTTConsumer) handleAlarmNotify(ctx context.Context, tenantID, deviceID, deviceUID, deviceType string, tsPayload int64, nowMsEnvelope int64, d *AlarmNotifyData) {
	am, ok := mqttAlarmMap[d.Type]
	if !ok {
		am = alarmMapping{EventName: d.Type, TrackID: observation.TrackUnknownPerson}
	}

	// event_status 与 EventItem 一致，字符串：start/end，下游 cardagg 按 string 判断
	eventStatus := "start"
	if d.Status != 0 {
		eventStatus = "end"
	}

	if c.statusTracker != nil {
		active := ""
		if eventStatus == "start" {
			active = am.EventName
		}
		c.statusTracker.UpdateAlarm(deviceID, active)
	}

	payloadData := map[string]any{
		"id":             d.Id,
		"type":           d.Type,
		"user_id":        d.UserId,
		"status":         d.Status,
		"relieve_reason": d.RelieveReason,
		"relieve_time":   d.RelieveTime,
	}
	payloadJSON, _ := json.Marshal(payloadData)
	item := observation.EventItem{
		DataCategory: am.EventName,
		EventID:      strconv.FormatInt(d.Id, 10),
		EventName:    am.EventName,
		EventSince:   tsPayload,
		EventStatus:  eventStatus,
		TrackID:      am.TrackID,
		EventPayload: string(payloadJSON),
	}
	if d.RelieveReason != "" {
		item.EventReason = d.RelieveReason
	}
	if d.RelieveTime > 0 {
		item.EventEnd = toMillis(d.RelieveTime)
	}
	out, err := observation.EventItemToDataMap(&item)
	if err != nil {
		c.logger.Warn("EventItemToDataMap failed", zap.String("device_id", deviceID), zap.Error(err))
		return
	}
	if out == nil {
		out = make(map[string]any)
	}
	msg := redis.NewIoTStreamMessageWithData(tenantID, "", deviceUID, deviceID, deviceType, nowMsEnvelope, "alarm", am.EventName, out)
	c.publisher.PublishAlarm(ctx, msg)
}

// toMillis ensures timestamp is in milliseconds.
// Sleepace may send seconds (< 1e12) or milliseconds.
func toMillis(ts int64) int64 {
	if ts > 0 && ts < 1e12 {
		return ts * 1000
	}
	return ts
}
