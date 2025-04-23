package logger

import (
	"os"

	"image-analyzer-go/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger 全局日志对象
	Logger *zap.Logger
)

// Config 日志配置
type Config struct {
	// LogDir 日志目录
	LogDir string
	// LogFile 日志文件名
	LogFile string
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		LogDir:  "logs",
		LogFile: "image-analyzer.log",
	}
}

// GetLogPath 获取完整的日志文件路径
func (c *Config) GetLogPath() string {
	return c.LogDir + "/" + c.LogFile
}

// Init 初始化日志系统
func Init(cfg *config.Config) error {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	// 创建日志文件
	if err := os.MkdirAll(cfg.GetLogDir(), os.ModePerm); err != nil {
		return err
	}
	file, err := os.OpenFile(cfg.GetLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心
	core := zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(file),
			zapcore.InfoLevel,
		),
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		),
	)

	// 创建日志对象
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// Debug 输出调试日志
func Debug(msg string, fields ...zapcore.Field) {
	Logger.Debug(msg, fields...)
}

// Info 输出信息日志
func Info(msg string, fields ...zapcore.Field) {
	Logger.Info(msg, fields...)
}

// Warn 输出警告日志
func Warn(msg string, fields ...zapcore.Field) {
	Logger.Warn(msg, fields...)
}

// Error 输出错误日志
func Error(msg string, fields ...zapcore.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 输出致命错误日志并退出
func Fatal(msg string, fields ...zapcore.Field) {
	Logger.Fatal(msg, fields...)
}

// WithError 创建一个包含错误信息的字段
func WithError(err error) zapcore.Field {
	return zap.Error(err)
}

// WithString 创建一个字符串字段
func WithString(key, value string) zapcore.Field {
	return zap.String(key, value)
}

// WithInt 创建一个整数字段
func WithInt(key string, value int) zapcore.Field {
	return zap.Int(key, value)
}

// WithBool 创建一个布尔字段
func WithBool(key string, value bool) zapcore.Field {
	return zap.Bool(key, value)
}

// WithAny 创建一个任意字段
func WithAny(key string, value interface{}) zapcore.Field {
	return zap.Any(key, value)
}
