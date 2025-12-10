package log

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var TimestampFormat = "2006-01-02 15:04:05"
var defaultLogger = getDefaultLogger()

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(TimestampFormat)
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

type logger struct {
	zap *zap.Logger
}

func (l *logger) Info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}
func (l *logger) Error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}
func (l *logger) Debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}
func (l *logger) Warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}
func (l *logger) Fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}
func (l *logger) Panic(msg string, fields ...zap.Field) {
	l.zap.Panic(msg, fields...)
}

func (l *logger) Sync() error {
	return l.zap.Sync()
}

var lock *sync.RWMutex = new(sync.RWMutex)

func Info(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.Info(msg, fields...)
}
func Error(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.Error(msg, fields...)
}
func Debug(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.Debug(msg, fields...)
}
func Warn(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.Warn(msg, fields...)
}
func Fatal(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.Fatal(msg, fields...)
}
func Panic(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.Panic(msg, fields...)
}
func Sync() error {
	lock.RLock()
	defer lock.RUnlock()
	return defaultLogger.Sync()
}

func getDefaultLogger() *logger {
	cores := zapcore.NewTee(getStdoutLogWriter())
	log := zap.New(cores, zap.AddCaller(), zap.AddCallerSkip(2))
	return &logger{
		zap: log,
	}
}

func InitLogger(path string) {
	lock.Lock()
	defer lock.Unlock()
	cores := zapcore.NewTee(getStdoutLogWriter())
	if len(path) > 0 {
		writeFileCore := getFileLogWriter(path)
		cores = zapcore.NewTee(writeFileCore, getStdoutLogWriter())
	}
	log := zap.New(cores, zap.AddCaller(), zap.AddCallerSkip(2))
	defaultLogger = &logger{
		zap: log,
	}
}
