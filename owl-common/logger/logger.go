package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

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
		// 使用开发模式配置（控制台输出）
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapLevel)
	} else {
		// 使用生产模式配置（JSON输出）
		config = zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// 输出到标准输出（便于Docker和日志收集器捕获）
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}
	}
	
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

