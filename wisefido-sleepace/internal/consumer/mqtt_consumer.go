package consumer

import (
	"context"
	"encoding/json"
	"net/netip"
	"strconv"
	"sync"
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

func businessAccessApproved(s string) bool {
	switch s {
	case "approved", "enable":
		return true
	default:
		return false
	}
}

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

	// dedup: topic+deviceId+dataKey+timestamp → seen
	dedupMu sync.Mutex
	dedup   map[string]struct{}

	// 设备健康类（Offline/SensorDetached）transition dedup：仅 0↔1 跳变时下发，
	// 防止 sleepace 云若把状态当心跳重发或 MQTT 网络抖动重投导致 alarm_events 被周期灌行。
	// key=deviceUID, value: 0=ok, 1=fault, -1=尚未观测过（首次=0 静默初始化，不发"全清"假恢复）。
	deviceHealthMu       sync.Mutex
	prevConnFault        map[string]int // 0=online, 1=offline
	prevSensorFailure    map[string]int // 0=已插上, 1=脱落
}

type rawMsg struct {
	topic   string
	payload []byte
}

func NewMQTTConsumer(publisher *StreamPublisher, reportSvc *service.ReportService, statusTracker *service.DeviceStatusTracker, logger *zap.Logger) *MQTTConsumer {
	return &MQTTConsumer{
		publisher:         publisher,
		reportService:     reportSvc,
		statusTracker:     statusTracker,
		logger:            logger,
		msgCh:             make(chan *rawMsg, 1000),
		lastBedStatus:     make(map[string]int),
		dedup:             make(map[string]struct{}),
		prevConnFault:     make(map[string]int),
		prevSensorFailure: make(map[string]int),
	}
}

// shouldPublishDeviceHealth 判断设备健康类事件（Offline/SensorDetached）是否需要下发。
//
// 返回 true 表示这是一次真正的状态跳变（或首次故障），调用方应正常 publish；
// 返回 false 表示状态未变（或首次=健康），调用方应跳过 publish。
//
// 状态规则：
//   - 首次观测 + value=0（健康）：静默初始化为 0，返回 false（不发"全清"假恢复）
//   - 首次观测 + value=1（故障）：记录为 1，返回 true（发 onset）
//   - 已记录 + 值未变：返回 false
//   - 已记录 + 值跳变：更新缓存，返回 true
//
// cache=&c.prevConnFault 或 &c.prevSensorFailure；deviceUID 为 key；fault=1 表示故障。
func (c *MQTTConsumer) shouldPublishDeviceHealth(cache map[string]int, deviceUID string, fault int) bool {
	c.deviceHealthMu.Lock()
	defer c.deviceHealthMu.Unlock()
	last, seen := cache[deviceUID]
	if !seen {
		if fault == 0 {
			cache[deviceUID] = 0
			return false
		}
		cache[deviceUID] = 1
		return true
	}
	if last == fault {
		return false
	}
	cache[deviceUID] = fault
	return true
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
	if isSleepaceVerboseLog() {
		c.logger.Debug("[MQTT_RX]",
			zap.String("topic", msg.topic),
			zap.ByteString("payload", msg.payload))
	}

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

// isDuplicate returns true if this message was already seen (dedup by deviceId+dataKey+timestamp).
func (c *MQTTConsumer) isDuplicate(m *ReceivedMessage) bool {
	key := m.DeviceID + "|" + m.DataKey + "|" + strconv.FormatInt(m.TimeStamp, 10)
	c.dedupMu.Lock()
	defer c.dedupMu.Unlock()
	if _, ok := c.dedup[key]; ok {
		return true
	}
	c.dedup[key] = struct{}{}
	// Simple cap: if map grows too large, reset it
	if len(c.dedup) > 100000 {
		c.dedup = make(map[string]struct{})
		c.dedup[key] = struct{}{}
	}
	return false
}

// dispatch routes a single ReceivedMessage by constructing IoTStreamMessage and publishing to iot: streams.
// 权限策略：
// - allow_access 且 business_access 为 approved|enable => canIoT
// - monitoring_enabled=TRUE => canMonitor
// - monitor=off（canMonitor=false）时，仅保留连通性最小信号（connectionStatus 与 monitor 心跳），不发布生理/行为 event/alarm。
// MQTT 的 deviceId 首次可能为 device_uid、后续可能为 device_code，统一传 Resolve(deviceKey) 由 DB 解析；三元组由 DeviceBaseline 带回。device_code 以 DB 为准，空则不兜底 MQTT 的 deviceId，便于暴露配置/数据问题。
func (c *MQTTConsumer) dispatch(ctx context.Context, m *ReceivedMessage) {
	if c.isDuplicate(m) {
		c.logger.Debug("duplicate message skipped", zap.String("device_id", m.DeviceID), zap.String("dataKey", m.DataKey))
		return
	}
	tenantID, _, _, cardID, deviceID, deviceUID, deviceCode, deviceType, allowAccess, businessAccess, monitoringEnabled, addr := c.publisher.Resolve(ctx, m.DeviceID)
	_ = tenantID
	_ = deviceID
	_ = deviceUID
	canIoT := allowAccess && businessAccessApproved(businessAccess)
	canMonitor := canIoT && monitoringEnabled
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
		// 首次上连（在线）时拉取 deviceVersion，与 device_store.firmware_version 比对，不一致则写入 ota_target_firmware_version（须 DB 有 device_code，不兜底 MQTT deviceId）
		if online && c.cardDB != nil && c.sleepaceAPI != nil {
			if deviceCode == "" {
				c.logger.Warn("skip device_store version sync: empty device_code", zap.String("mqtt_deviceId", m.DeviceID), zap.String("device_uid", deviceUID))
			} else {
				go c.syncDeviceStoreVersion(context.Background(), deviceCode)
			}
		}
		// iot:alarm 仅使用 device_id (UUID)；未入 device_store 或无 IoT 权限则不发布
		if deviceID == "" {
			c.logger.Debug("connectionStatus skip alarm (device not in device_store)", zap.String("device_uid", deviceUID))
		} else if !canIoT {
			c.logger.Debug("connectionStatus skip alarm (permission)", zap.String("device_uid", deviceUID), zap.Bool("allow_access", allowAccess), zap.String("business_access", businessAccess))
		} else {
			fault := 0
			if !online {
				fault = 1
			}
			// transition-only：值未变 / 首次=online 都不下发，避免 sleepace 云重发心跳灌爆 alarm_events
			if !c.shouldPublishDeviceHealth(c.prevConnFault, deviceUID, fault) {
				c.logger.Debug("connectionStatus dedup (no transition)", zap.String("device_uid", deviceUID), zap.Bool("online", online))
			} else {
				tsMs := toMillis(ts)
				streamCID := cardID
				var eventName string
				var item observation.EventItem
				if online {
					eventName = alarm.AlarmTypeOfflineRecover
					item = observation.NewEventItem(tsMs, "end")
				} else {
					eventName = alarm.AlarmTypeOffline
					item = observation.NewEventItem(tsMs, "start")
				}
				item.TrackID = observation.TrackDevice
				alarmData, _ := observation.EventItemToDataMap(&item)
				if alarmData == nil {
					alarmData = make(map[string]any)
				}
				if online {
					alarmData[observation.FieldOffline] = 0
				} else {
					alarmData[observation.FieldOffline] = 1
				}
				msg := redis.NewIoTStreamMessageWithData(addr, streamCID, deviceType, nowMs, "alarm", eventName, alarmData)
				if err := c.publisher.PublishAlarm(ctx, msg); err != nil {
					c.logger.Warn("connectionStatus publish alarm failed", zap.String("device_uid", deviceUID), zap.Bool("online", online), zap.Error(err))
				}
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
		if canMonitor {
			msg := redis.NewIoTStreamMessageWithData(addr, cardID, deviceType, ts, "monitor", observation.CategoryTrack, data)
			c.publisher.PublishMonitor(ctx, msg)
			c.logger.Debug("realtime received",
				zap.String("device_uid", deviceUID),
				zap.String("device_id", deviceID),
				zap.Int("bed_status", d.BedStatus),
				zap.Int("heart", d.Heart),
				zap.Int("breath", d.Breath),
				zap.Int("init_status", d.InitStatus))
		} else if canIoT {
			hb := observation.Track{TrackID: observation.TrackDevice, TrackConfidence: sleepaceTrackPoseConf}
			hbData := hb.ToFieldMap()
			hbMsg := redis.NewIoTStreamMessageWithData(addr, cardID, deviceType, ts, "monitor", observation.CategoryHeart, hbData)
			_ = c.publisher.PublishMonitor(ctx, hbMsg)
		}

		// bedStatus 实为状态量：仅 monitor=on 时参与 InBed/LeftBed 事件；
		// monitor=off 时仅保留上面的心跳（track_id=11, category=heart），不再发床态事件。
		if canMonitor && d.InitStatus == 0 && d.BedStatus != sleepaceStatusInvalid {
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
				evItem := observation.NewEventItem(ts, "start")
				evItem.TrackID = d.LeftRight
				evData, _ := observation.EventItemToDataMap(&evItem)
				if evData == nil {
					evData = make(map[string]any)
				}
				evData[observation.FieldBedStatus] = d.BedStatus // int: 0=在床, 1=离床
				evMsg := redis.NewIoTStreamMessageWithData(addr, cardID, deviceType, nowMs, "event", categoryBed, evData)
				if canIoT {
					if err := c.publisher.PublishEvent(ctx, evMsg); err != nil {
						c.logger.Warn("publish bedStatus change event", zap.String("device_uid", deviceUID), zap.Int("bedStatus", d.BedStatus), zap.Error(err))
					} else {
						c.logger.Info("bedStatus change → event", zap.String("device_uid", deviceUID), zap.Int("leftRight", d.LeftRight), zap.String("category", categoryBed))
					}
				}
			}
		}

	// inBedStatus：0=未离床 1=离床；8=算法初始化或状态未变，保持前一有效值，不写 stream。
	case "inBedStatus":
		if !canMonitor {
			return
		}
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
		item := observation.NewEventItem(ts, "instant")
		item.TrackID = d.LeftRight
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out[observation.FieldBedStatus] = d.InbedStatus
		if canIoT {
			c.publisher.PublishEvent(ctx, redis.NewIoTStreamMessageWithData(addr, cardID, deviceType, nowMs, "event", categoryInBed, out))
			c.logger.Info("inBedStatus", zap.String("category", categoryInBed), zap.String("device_uid", deviceUID), zap.Int("status", d.InbedStatus), zap.Int("lr", d.LeftRight))
		}

	case "sleepStage":
		if !canMonitor {
			return
		}
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
		item := observation.NewEventItem(ts, "instant")
		item.TrackID = d.LeftRight
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out[observation.FieldSleepStage] = stage
		msg := redis.NewIoTStreamMessageWithData(addr, cardID, deviceType, nowMs, "event", alarm.SleepStage, out)
		if canIoT {
			c.publisher.PublishEvent(ctx, msg)
			c.logger.Info("sleepStage", zap.String("category", alarm.SleepStage), zap.String("device_uid", deviceUID), zap.Int("stage", stage), zap.Int("lr", d.LeftRight))
		}

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
		// transition-only：sleepace 设备会周期重发 deviceSenSor，状态未变就跳过，避免 alarm_events 被刷
		if !c.shouldPublishDeviceHealth(c.prevSensorFailure, deviceUID, failureVal) {
			c.logger.Debug("deviceSenSor dedup (no transition)", zap.String("device_uid", deviceUID), zap.Bool("failure", failure))
			return
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
		item := observation.NewEventItem(ts, eventStatus)
		item.TrackID = observation.TrackDevice
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out[observation.FieldDeviceFailure] = failureVal
		if canIoT {
			msg := redis.NewIoTStreamMessageWithData(addr, cardID, deviceType, nowMs, "alarm", eventName, out)
			c.publisher.PublishAlarm(ctx, msg)
		}

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
		item := observation.NewEventItem(ts, eventStatus)
		item.TrackID = observation.TrackDevice
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out[observation.FieldDetached] = detachedVal
		if canIoT {
			msg := redis.NewIoTStreamMessageWithData(addr, cardID, deviceType, nowMs, "event", "pressureSensor", out)
			c.publisher.PublishEvent(ctx, msg)
		}

	case "analysis":
		if !canMonitor {
			return
		}
		var d AnalysisData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal analysis", zap.Error(err))
			return
		}
		// Sleepace userId 与 wisefido devices.device_id 一致；流与落库均以解析到的 deviceID 为准。
		streamUserID := d.UserId
		if deviceID != "" {
			streamUserID = deviceID
		}
		item := observation.NewEventItem(ts, "instant")
		item.TrackID = observation.TrackDevice
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out["user_id"] = streamUserID
		out["start_time"] = d.StartTime
		if deviceID != "" && canIoT {
			msg := redis.NewIoTStreamMessageWithData(addr, cardID, deviceType, nowMs, "event", "analysis", out)
			c.publisher.PublishEvent(ctx, msg)
		}
		if c.reportService != nil && deviceID != "" && canIoT {
			// 异步拉报告；形参在 go 启动前求值并拷贝，时间窗 = analysis.startTime+1 … 信封 timeStamp，did = devices.device_id
			go func(start, end int64, did string) {
				if err := c.reportService.DownloadAndSave(ctx, did, start, end); err != nil {
					c.logger.Error("download report failed", zap.String("device_id", did), zap.Error(err))
				}
			}(d.StartTime+1, m.TimeStamp, deviceID)
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
		item := observation.NewEventItem(ts, "instant")
		item.TrackID = observation.TrackDevice
		out, _ := observation.EventItemToDataMap(&item)
		if out == nil {
			out = make(map[string]any)
		}
		out[observation.FieldCommandAction] = "firmware_upgrade"
		out[observation.FieldCommandProgress] = progress
		out["length"] = d.Length
		out["offset"] = d.Offset
		if canIoT {
			msg := redis.NewIoTStreamMessageWithData(addr, cardID, deviceType, nowMs, "event", "upgradeProgress", out)
			c.publisher.PublishEvent(ctx, msg)
		}

	// 明确报警项：设备端已判定并上报告警，发 iot:alarm:stream；其余（InBed/LeftBed、SleepStage、pressureSensor 等）发 event，由后端判断。
	// alarmNotify：默认 single 单条，可通过 pushType 接口设 alarmDataType=array；handleMessage 已兼容单条/数组，此处每条为单条 data。
	case "alarmNotify":
		if deviceID == "" {
			return
		}
		if !canMonitor {
			return
		}
		var d AlarmNotifyData
		if err := json.Unmarshal(m.Data, &d); err != nil {
			c.logger.Error("unmarshal alarmNotify", zap.Error(err))
			return
		}
		c.handleAlarmNotify(ctx, addr, cardID, deviceID, deviceUID, deviceType, ts, nowMs, &d)

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
func (c *MQTTConsumer) handleAlarmNotify(ctx context.Context, addr netip.Addr, cardID, deviceID, deviceUID, deviceType string, tsPayload int64, nowMsEnvelope int64, d *AlarmNotifyData) {
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

	// Plan B：sleepace alarmNotify 所有字段都已经在 envelope/EventItem 顶层
	//   id            → EventItem.EventID
	//   type          → envelope.Category（via am.EventName 映射）
	//   status        → EventItem.EventStatus（"start"/"end"）
	//   userId        → envelope.SubjectEntity（gateway 已映射成 card_id）
	//   deviceId      → envelope.DeviceUID
	//   relieveReason → EventItem.EventReason
	//   relieveTime   → EventItem.EventEnd
	//   timestamp     → EventItem.EventSince
	// EventPayload 之前装的 raw 复本是 100% 冗余，删掉让下游直接读 envelope/EventItem。
	item := observation.NewEventItem(tsPayload, eventStatus)
	item.EventID = strconv.FormatInt(d.Id, 10)
	item.TrackID = am.TrackID
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
	msg := redis.NewIoTStreamMessageWithData(addr, cardID, deviceType, nowMsEnvelope, "alarm", am.EventName, out)
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
