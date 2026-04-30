package config

import (
	"log"
	"os"
	"owl-common/config"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config 报警服务配置
type Config struct {
	Database config.DatabaseConfig `yaml:"database"`
	Redis    config.RedisConfig    `yaml:"redis"`

	Alarm struct {
		Cache struct {
			RealtimeKeyPrefix string `yaml:"realtime_key_prefix"`
			RealtimeSuffix    string `yaml:"realtime_suffix"`
			AlarmKeyPrefix    string `yaml:"alarm_key_prefix"`
			AlarmSuffix       string `yaml:"alarm_suffix"`
			AlarmTTL          int    `yaml:"alarm_ttl"`
			StateKeyPrefix    string `yaml:"state_key_prefix"`
		} `yaml:"cache"`
		PollInterval int `yaml:"poll_interval"`
		Evaluation   struct {
			BatchSize int `yaml:"batch_size"`
		} `yaml:"evaluation"`
		IoTStream struct {
			Enabled       bool   `yaml:"enabled"`
			Monitor       string `yaml:"monitor"`
			Stat          string `yaml:"stat"`
			Event         string `yaml:"event"`
			Alarm         string `yaml:"alarm"`
			ConsumerGroup string `yaml:"consumer_group"`
			ConsumerName  string `yaml:"consumer_name"`
			BatchSize     int64  `yaml:"batch_size"`
		} `yaml:"iot_stream"`
	} `yaml:"alarm"`

	Log struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`

	RoomEngine RoomEngineConfig `yaml:"roomengine"`

	AIPublish AIPublishConfig `yaml:"ai_publish"`
}

// AIPublishConfig 控制 AI 派生 event/alarm 的发布行为。
//
// 单点配置同时管两件事：
//  1. 推送模式（仅 log，还是 log + 推送 iot stream）
//  2. AI 节点序号 —— 进 device_type 后缀（如 "Radar.AI01"）+ ai.log 审计字段
//
// device_uid 不变（仍是源 radar 的业务主键），只在 device_type 里加 ".AI<node_id>"
// 后缀区分 AI 派生事件，原 sensor 类型保留以兼容 cardagg 现有路由。
type AIPublishConfig struct {
	// Mode 取值：
	//   "log"          仅写 ai.log（每条 emit 都带 ai_node + would_publish_to + published=false）
	//   "log&publish"  写 log + 推送 iot:event/alarm:stream（默认）
	// 任何模式 alarm 都正常 fire（"宁可误报不可漏报"原则），mode 仅控制 publish 路径。
	Mode string `yaml:"mode"`

	// NodeID AI 节点序号字符串（如 "01"、"02"）。
	// 驱动两处：
	//   - IoTStreamMessage.DeviceType = "<base>.AI" + NodeID，如 "Radar.AI01"
	//   - ai.log 每条 publish/skip 都带 ai_node="AI" + NodeID
	// 多 AI 实例横向扩展时每个进程取不同 NodeID。
	NodeID string `yaml:"node_id"`
}

// PublishModeLog 仅 log 不推流；PublishModeLogPublish log + 推流。
const (
	PublishModeLog        = "log"
	PublishModeLogPublish = "log&publish"
)

// PublishEnabled 是否实际推送到 redis stream。Mode 为 PublishModeLogPublish 时返回 true。
func (c AIPublishConfig) PublishEnabled() bool {
	return c.Mode == PublishModeLogPublish
}

// AINodeTag 返回 "AI" + NodeID（如 "AI01"），用于 device_type 后缀和 ai.log 审计字段。
func (c AIPublishConfig) AINodeTag() string {
	return "AI" + c.NodeID
}

// RoomEngineConfig wisefido-ai/internal/roomengine 运行时参数
// 每个字段都可在 config.yaml 中覆盖；未设置时由 setRoomEngineDefaults 填默认值
type RoomEngineConfig struct {
	Decay struct {
		ImmediateSec int `yaml:"immediate_sec"` // Real/Ghost/Flow/DwellEMA/Stand   默认 15 min
		WalkSec      int `yaml:"walk_sec"`      // ActiveType[Move]                 默认 3 d
		SitSec       int `yaml:"sit_sec"`       // ActiveType[Sit]                  默认 24 h
		LieSec       int `yaml:"lie_sec"`       // ActiveType[Lie]                  默认 7 d
		EventSec     int `yaml:"event_sec"`     // Traverse/Fall/Retract/Sleepad/Door/LongStill/LieAnomaly  默认 7 d
	} `yaml:"decay"`
	Learn struct {
		WalkActiveX10     int `yaml:"walk_active_x10"`     // 默认 20  (2s × 10)
		WalkTraverse      int `yaml:"walk_traverse"`       // 默认 2   (单 cell 累计穿越次数)
		SitActiveX10      int `yaml:"sit_active_x10"`      // 默认 150 (15s × 10)
		LieAnomalyX10     int `yaml:"lie_anomaly_x10"`     // 默认 50  (5s × 10)
		BedToleranceCm    int `yaml:"bed_tolerance_cm"`    // 默认 30
		ToiletToleranceCm int `yaml:"toilet_tolerance_cm"` // 默认 20
		ShowerToleranceCm int `yaml:"shower_tolerance_cm"` // 默认 20
		MoveSpeedCms      int `yaml:"move_speed_cms"`      // 默认 20  Kalman 速度阈值，pose 不准时兜底判 Move
		NearTraverseWalk  int `yaml:"near_traverse_walk"`  // 默认 5   邻居穿越阈值，本 cell 推断为 Walk（沿路径涂厚）
		NearTraverseDeny  int `yaml:"near_traverse_deny"`  // 默认 10  邻居穿越阈值，本 cell 完全无触碰时反转 Deny
	} `yaml:"learn"`
	Belief struct {
		DecayIntervalSec int  `yaml:"decay_interval_sec"` // 默认 3600 (1h)
		ScanIntervalSec  int  `yaml:"scan_interval_sec"`  // 默认 300  (5min)
		WinnerEvalSec    int  `yaml:"winner_eval_sec"`    // 默认 86400 (24h)
		ConfidenceFloor  int  `yaml:"confidence_floor"`   // 默认 60   (promoteCell 起跑)
		ConfidenceFull   int  `yaml:"confidence_full"`    // 默认 95   (promoteCell 上限)
		Cumulative       bool `yaml:"cumulative"`         // 默认 true (同 type 累积取 max，多日推向 100)
	} `yaml:"belief"`
	RiskTime struct {
		// 风险时段（夜间）开始/结束的本地时间，默认 23:30 - 07:30。
		// 用于 StillTimeoutSec 白天放宽（×1.2）和 R4（床边晕倒）窗口判定。
		NightStartH int `yaml:"night_start_h"` // 默认 23
		NightStartM int `yaml:"night_start_m"` // 默认 30
		NightEndH   int `yaml:"night_end_h"`   // 默认 7
		NightEndM   int `yaml:"night_end_m"`   // 默认 30
	} `yaml:"risk_time"`
	BedsideFall struct {
		// R4 床边晕倒：风险时段 LeftBed 后人未走开，床边静止 >阈值 → 报警
		WindowSec        int `yaml:"window_sec"`         // 默认 180 (LeftBed 后 N 秒内是"开窗期")
		BedsideMarginCm  int `yaml:"bedside_margin_cm"`  // 默认 100 (距 AreaBed cell 此值内 = 床边)
		StillTimeoutSec  int `yaml:"still_timeout_sec"`  // 默认 900 (静止 15 min 触发)
	} `yaml:"bedside_fall"`
	ParamSets []struct {
		Name   string  `yaml:"name"`
		Alpha  float64 `yaml:"alpha"`
		Beta   float64 `yaml:"beta"`
		FlipTh int     `yaml:"flip_th"`
	} `yaml:"paramsets"`
	Persist struct {
		Enabled             bool   `yaml:"enabled"`               // 默认 true
		SnapshotIntervalSec int    `yaml:"snapshot_interval_sec"` // 默认 300 (5min)
		Storage             string `yaml:"storage"`               // 默认 "postgres"
		Table               string `yaml:"table"`                 // 默认 "roomengine_grid_snapshot"
	} `yaml:"persist"`
}

// Load 加载配置
func Load() (*Config, error) {
	configPath := "config.yaml"
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		configPath = path
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("(info) config file not found, using environment variables: %v", err)
		return LoadFromEnv()
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Printf("(warn) failed to parse config file, falling back to env: %v", err)
		return LoadFromEnv()
	}

	cfg.setDefaults()
	return &cfg, nil
}

func (c *Config) setDefaults() {
	if c.Database.SSLMode == "" {
		c.Database.SSLMode = "disable"
	}
	if c.Alarm.Cache.RealtimeKeyPrefix == "" {
		c.Alarm.Cache.RealtimeKeyPrefix = "vital-focus:card:"
	}
	if c.Alarm.Cache.RealtimeSuffix == "" {
		c.Alarm.Cache.RealtimeSuffix = ":realtime"
	}
	if c.Alarm.Cache.AlarmKeyPrefix == "" {
		c.Alarm.Cache.AlarmKeyPrefix = "vital-focus:card:"
	}
	if c.Alarm.Cache.AlarmSuffix == "" {
		c.Alarm.Cache.AlarmSuffix = ":alarms"
	}
	if c.Alarm.Cache.AlarmTTL == 0 {
		c.Alarm.Cache.AlarmTTL = 30
	}
	if c.Alarm.Cache.StateKeyPrefix == "" {
		c.Alarm.Cache.StateKeyPrefix = "alarm:state:"
	}
	if c.Alarm.PollInterval == 0 {
		c.Alarm.PollInterval = 10
	}
	if c.Alarm.Evaluation.BatchSize == 0 {
		c.Alarm.Evaluation.BatchSize = 10
	}
	if c.Alarm.IoTStream.Monitor == "" {
		c.Alarm.IoTStream.Monitor = "iot:monitor:stream"
	}
	if c.Alarm.IoTStream.Stat == "" {
		c.Alarm.IoTStream.Stat = "iot:stat:stream"
	}
	if c.Alarm.IoTStream.Event == "" {
		c.Alarm.IoTStream.Event = "iot:event:stream"
	}
	if c.Alarm.IoTStream.Alarm == "" {
		c.Alarm.IoTStream.Alarm = "iot:alarm:stream"
	}
	if c.Alarm.IoTStream.ConsumerGroup == "" {
		c.Alarm.IoTStream.ConsumerGroup = "wisefido-alarm-events"
	}
	if c.Alarm.IoTStream.ConsumerName == "" {
		c.Alarm.IoTStream.ConsumerName = "alarm-consumer-1"
	}
	if c.Alarm.IoTStream.BatchSize == 0 {
		c.Alarm.IoTStream.BatchSize = 10
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "json"
	}
	c.setRoomEngineDefaults()
	c.setAIPublishDefaults()
}

func (c *Config) setAIPublishDefaults() {
	a := &c.AIPublish
	// env 覆盖优先（部署/调试场景）
	if v := os.Getenv("AI_PUBLISH_MODE"); v != "" {
		a.Mode = v
	}
	if v := os.Getenv("AI_NODE_ID"); v != "" {
		a.NodeID = v
	}
	// 默认 log&publish；非法值回退到默认并 warn
	if a.Mode == "" {
		a.Mode = PublishModeLogPublish
	} else if a.Mode != PublishModeLog && a.Mode != PublishModeLogPublish {
		log.Printf("(warn) invalid ai_publish.mode=%q, falling back to %q (valid: %q | %q)",
			a.Mode, PublishModeLogPublish, PublishModeLog, PublishModeLogPublish)
		a.Mode = PublishModeLogPublish
	}
	// node_id 默认 "01"；多 AI 实例必须显式指定
	if a.NodeID == "" {
		a.NodeID = "01"
	}
}

func (c *Config) setRoomEngineDefaults() {
	r := &c.RoomEngine
	if r.Decay.ImmediateSec == 0 {
		r.Decay.ImmediateSec = 15 * 60
	}
	if r.Decay.WalkSec == 0 {
		r.Decay.WalkSec = 3 * 24 * 3600
	}
	if r.Decay.SitSec == 0 {
		r.Decay.SitSec = 24 * 3600
	}
	if r.Decay.LieSec == 0 {
		r.Decay.LieSec = 7 * 24 * 3600
	}
	if r.Decay.EventSec == 0 {
		r.Decay.EventSec = 7 * 24 * 3600
	}
	if r.Learn.WalkActiveX10 == 0 {
		r.Learn.WalkActiveX10 = 20
	}
	if r.Learn.WalkTraverse == 0 {
		r.Learn.WalkTraverse = 5
	}
	if r.Learn.SitActiveX10 == 0 {
		r.Learn.SitActiveX10 = 150
	}
	if r.Learn.LieAnomalyX10 == 0 {
		r.Learn.LieAnomalyX10 = 50
	}
	if r.Learn.BedToleranceCm == 0 {
		r.Learn.BedToleranceCm = 30
	}
	if r.Learn.ToiletToleranceCm == 0 {
		r.Learn.ToiletToleranceCm = 20
	}
	if r.Learn.ShowerToleranceCm == 0 {
		r.Learn.ShowerToleranceCm = 20
	}
	if r.Learn.MoveSpeedCms == 0 {
		r.Learn.MoveSpeedCms = 20
	}
	if r.Learn.NearTraverseWalk == 0 {
		r.Learn.NearTraverseWalk = 5
	}
	if r.Learn.NearTraverseDeny == 0 {
		r.Learn.NearTraverseDeny = 20
	}
	if r.Belief.DecayIntervalSec == 0 {
		r.Belief.DecayIntervalSec = 3600
	}
	if r.Belief.ScanIntervalSec == 0 {
		r.Belief.ScanIntervalSec = 300
	}
	if r.Belief.WinnerEvalSec == 0 {
		r.Belief.WinnerEvalSec = 86400
	}
	if r.Belief.ConfidenceFloor == 0 {
		r.Belief.ConfidenceFloor = 60
	}
	if r.Belief.ConfidenceFull == 0 {
		r.Belief.ConfidenceFull = 95
	}
	// Cumulative 是 promoteCell 的核心改进（同 type 累积取 max → 多日推向 100），
	// 没有"想关掉"的合理场景；不暴露为可配置开关，engine 内部硬编码启用。
	// （Belief.Cumulative 字段保留是为了 yaml 兼容；engine 不读取它。）
	if len(r.ParamSets) == 0 {
		r.ParamSets = []struct {
			Name   string  `yaml:"name"`
			Alpha  float64 `yaml:"alpha"`
			Beta   float64 `yaml:"beta"`
			FlipTh int     `yaml:"flip_th"`
		}{
			{Name: "conservative", Alpha: 0.01, Beta: 0.2, FlipTh: 10},
			{Name: "balanced", Alpha: 0.02, Beta: 0.5, FlipTh: 20},
			{Name: "aggressive", Alpha: 0.05, Beta: 1.0, FlipTh: 30},
		}
	}
	if r.Persist.SnapshotIntervalSec == 0 {
		r.Persist.SnapshotIntervalSec = 300
	}
	if r.Persist.Storage == "" {
		r.Persist.Storage = "postgres"
	}
	if r.Persist.Table == "" {
		r.Persist.Table = "roomengine_grid_snapshot"
	}
	// Persist.Enabled 不在此处自动开启：dev/test 场景需要能不持久化，由 yaml 显式控制。

	// RiskTime 默认 23:30 - 07:30
	if r.RiskTime.NightStartH == 0 && r.RiskTime.NightStartM == 0 &&
		r.RiskTime.NightEndH == 0 && r.RiskTime.NightEndM == 0 {
		r.RiskTime.NightStartH = 23
		r.RiskTime.NightStartM = 30
		r.RiskTime.NightEndH = 7
		r.RiskTime.NightEndM = 30
	}
	// BedsideFall 默认 180s 开窗 / 100cm 床边 / 900s 静止超时
	if r.BedsideFall.WindowSec == 0 {
		r.BedsideFall.WindowSec = 180
	}
	if r.BedsideFall.BedsideMarginCm == 0 {
		r.BedsideFall.BedsideMarginCm = 100
	}
	if r.BedsideFall.StillTimeoutSec == 0 {
		r.BedsideFall.StillTimeoutSec = 900
	}
}

func LoadFromEnv() (*Config, error) {
	cfg := &Config{}
	cfg.Database.Host = getEnv("DB_HOST", "localhost")
	if portStr := getEnv("DB_PORT", ""); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 {
			cfg.Database.Port = port
		} else {
			cfg.Database.Port = 5432
		}
	} else {
		cfg.Database.Port = 5432
	}
	cfg.Database.User = getEnv("DB_USER", "postgres")
	cfg.Database.Password = getEnv("DB_PASSWORD", "")
	cfg.Database.Database = getEnv("DB_NAME", "owlrd")
	cfg.Database.SSLMode = getEnv("DB_SSLMODE", "disable")
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = 0
	cfg.Alarm.Cache.RealtimeKeyPrefix = getEnv("CACHE_REALTIME_PREFIX", "vital-focus:card:")
	cfg.Alarm.Cache.RealtimeSuffix = ":realtime"
	cfg.Alarm.Cache.AlarmKeyPrefix = getEnv("CACHE_ALARM_PREFIX", "vital-focus:card:")
	cfg.Alarm.Cache.AlarmSuffix = ":alarms"
	cfg.Alarm.Cache.AlarmTTL = 30
	cfg.Alarm.Cache.StateKeyPrefix = getEnv("CACHE_STATE_PREFIX", "alarm:state:")
	cfg.Alarm.PollInterval = 10
	cfg.Alarm.Evaluation.BatchSize = 10
	cfg.Alarm.IoTStream.Enabled = getEnv("AI_IOT_STREAM_ENABLED", "true") == "true"
	cfg.Alarm.IoTStream.Monitor = getEnv("AI_STREAM_MONITOR", "iot:monitor:stream")
	cfg.Alarm.IoTStream.Stat = getEnv("AI_STREAM_STAT", "iot:stat:stream")
	cfg.Alarm.IoTStream.Event = getEnv("AI_STREAM_EVENT", "iot:event:stream")
	cfg.Alarm.IoTStream.Alarm = getEnv("AI_STREAM_ALARM", "iot:alarm:stream")
	cfg.Alarm.IoTStream.ConsumerGroup = getEnv("AI_IOT_CONSUMER_GROUP", "wisefido-alarm-events")
	cfg.Alarm.IoTStream.ConsumerName = getEnv("AI_IOT_CONSUMER_NAME", "alarm-consumer-1")
	batchSizeStr := getEnv("AI_IOT_BATCH_SIZE", "10")
	if v, err := strconv.ParseInt(batchSizeStr, 10, 64); err == nil && v > 0 {
		cfg.Alarm.IoTStream.BatchSize = v
	} else {
		cfg.Alarm.IoTStream.BatchSize = 10
	}
	cfg.Log.Level = getEnv("LOG_LEVEL", "info")
	cfg.Log.Format = getEnv("LOG_FORMAT", "json")
	cfg.setAIPublishDefaults()
	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
