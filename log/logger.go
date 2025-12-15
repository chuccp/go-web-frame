package log

import (
	"log"
	"os"
	"path/filepath"
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

func (l *logger) info(msg string, fields ...zap.Field) {
	l.zap.Info(msg, fields...)
}
func (l *logger) error(msg string, fields ...zap.Field) {
	l.zap.Error(msg, fields...)
}
func (l *logger) debug(msg string, fields ...zap.Field) {
	l.zap.Debug(msg, fields...)
}
func (l *logger) warn(msg string, fields ...zap.Field) {
	l.zap.Warn(msg, fields...)
}
func (l *logger) fatal(msg string, fields ...zap.Field) {
	l.zap.Fatal(msg, fields...)
}
func (l *logger) panic(msg string, fields ...zap.Field) {
	l.zap.Panic(msg, fields...)
}

func (l *logger) sync() error {
	return l.zap.Sync()
}

var lock *sync.RWMutex = new(sync.RWMutex)

func Info(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.info(msg, fields...)
}
func Error(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.error(msg, fields...)
}
func Errors(msg string, errs ...error) {
	lock.RLock()
	defer lock.RUnlock()
	fields := make([]zap.Field, len(errs))
	for i, e := range errs {
		fields[i] = zap.Error(e)
	}
	defaultLogger.error(msg, fields...)
	log.Println(errs)
}
func Debug(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.debug(msg, fields...)
}
func Warn(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.warn(msg, fields...)
}
func Fatal(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.fatal(msg, fields...)
}
func Panic(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.panic(msg, fields...)
}
func PanicErrors(msg string, errs ...error) {
	lock.RLock()
	defer lock.RUnlock()
	fields := make([]zap.Field, len(errs))
	for i, e := range errs {
		fields[i] = zap.Error(e)
	}
	defaultLogger.panic(msg, fields...)
	log.Println(errs)
}

func Sync() error {
	lock.RLock()
	defer lock.RUnlock()
	return defaultLogger.sync()
}

func getDefaultLogger() *logger {
	cores := zapcore.NewTee(getStdoutLogWriter())
	log := zap.New(cores, zap.AddCaller(), zap.AddCallerSkip(2))
	return &logger{
		zap: log,
	}
}

func InitLogger(logPath string) {
	if len(logPath) > 0 {
		abs, err := filepath.Abs(logPath)
		if err != nil {
			Panic("日志文件路径错误", zap.Error(err))
			return
		}
		logPath = abs
	}
	Info("日志保存路径", zap.String("logPath", logPath))
	lock.Lock()
	defer lock.Unlock()
	cores := zapcore.NewTee(getFileLogWriter(logPath), getStdoutLogWriter())
	log := zap.New(cores, zap.AddCaller(), zap.AddCallerSkip(2))
	defaultLogger = &logger{
		zap: log,
	}
}
