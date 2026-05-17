package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// utcMillisEncoder ts 编码为 UTC + 毫秒精度 + "Z" 后缀：2026-05-17T02:46:18.383Z
// 跨服务统一时间戳格式，便于 journalctl/grep + DB SQL JOIN 时间对齐。
// 服务内部全 UTC（per memory server_internal_utc_only）；TZ 转换只在 API 边界。
// ms 精度对 ops/审计/trace 关联够用；纳秒过细。
func utcMillisEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format("2006-01-02T15:04:05.000Z"))
}

// NewLogger 创建新的Logger实例
// level: "debug", "info", "warn", "error" (默认: "info")
// format: "json" 或 "console" (默认: "json")
// serviceName: 服务名称（用于SaaS多租户日志管理，如 "wisefido-data"）
func NewLogger(level string, format string, serviceName string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}
	
	var config zap.Config
	if format == "console" {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapLevel)
		config.DisableStacktrace = true
	} else {
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapLevel)
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
	}
	// 统一时间戳：UTC + ms + "Z" 后缀（无论 console / json 格式）
	config.EncoderConfig.TimeKey = "ts"
	config.EncoderConfig.EncodeTime = utcMillisEncoder
	
	// 构建基础logger
	baseLogger, err := config.Build()
	if err != nil {
		return nil, err
	}
	
	// 如果提供了服务名称，添加为全局字段（用于SaaS日志管理）
	if serviceName != "" {
		baseLogger = baseLogger.With(zap.String("service_name", serviceName))
	}
	
	// 添加主机名（可选，用于分布式系统）
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		baseLogger = baseLogger.With(zap.String("hostname", hostname))
	}
	
	return baseLogger, nil
}

// NewLoggerWithDefaults 使用默认配置创建Logger实例（向后兼容）
// 默认: level="info", format="json", serviceName=""
func NewLoggerWithDefaults() (*zap.Logger, error) {
	return NewLogger("info", "json", "")
}

// NewDevelopmentLogger 创建开发环境Logger（向后兼容）
func NewDevelopmentLogger() (*zap.Logger, error) {
	return zap.NewDevelopment()
}

