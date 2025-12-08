package log

import (
	"os"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/util"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(util.TimestampFormat)
	return zapcore.NewJSONEncoder(encoderConfig)
}
func getFileLogWriter(path string) zapcore.Core {
	logger := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     30,   //days
		Compress:   true, // disabled by default
	}
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, zapcore.AddSync(logger), zapcore.InfoLevel)
	return core
}
func getStdoutLogWriter() zapcore.Core {
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, os.Stdout, zapcore.DebugLevel)
	return core
}

type Logger struct {
	zap    *zap.Logger
	config *config2.Config
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

func InitLogger(Config *config2.Config) *Logger {
	path := Config.GetString("log.path")
	cores := zapcore.NewTee(getStdoutLogWriter())
	if len(path) > 0 {
		writeFileCore := getFileLogWriter(path)
		cores = zapcore.NewTee(writeFileCore, getStdoutLogWriter())
	}
	logger := zap.New(cores, zap.AddCaller(), zap.AddCallerSkip(1))
	return &Logger{
		zap:    logger,
		config: Config,
	}
}
