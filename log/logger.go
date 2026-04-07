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

type Config struct {
	Level      string // 日志级别: debug, info, warn, error
	Path       string // 日志文件路径
	Write      bool   // 是否为后台写入模式
	MaxSize    int    // 单个日志文件最大大小 (MB)，默认 500
	MaxBackups int    // 保留的旧日志文件最大数量，默认 3
	MaxAge     int    // 保留旧日志文件的最大天数，默认 30
	Compress   bool   // 是否压缩旧日志文件，默认 true
	LocalTime  bool   // 是否使用本地时间，默认 false (使用 UTC)
}

func (c *Config) Key() string {
	return "web.log"
}

func defaultConfig() *Config {
	return &Config{
		Level:      "info",
		Path:       "",
		Write:      false,
		MaxSize:    500, // 500 MB
		MaxBackups: 3,
		MaxAge:     30, // 30 days
		Compress:   true,
		LocalTime:  false,
	}
}

var TimestampFormat = "2006-01-02 15:04:05"
var defaultLogger = getDefaultLogger()

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(TimestampFormat)
	return zapcore.NewJSONEncoder(encoderConfig)
}
func getFileLogWriter(cfg *Config) zapcore.Core {
	// 应用默认值
	maxSize := cfg.MaxSize
	if maxSize <= 0 {
		maxSize = 500
	}
	maxBackups := cfg.MaxBackups
	if maxBackups <= 0 {
		maxBackups = 3
	}
	maxAge := cfg.MaxAge
	if maxAge <= 0 {
		maxAge = 30
	}

	logger := &lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    maxSize,    // megabytes
		MaxBackups: maxBackups, // number of backups
		MaxAge:     maxAge,     // days
		Compress:   cfg.Compress,
		LocalTime:  cfg.LocalTime,
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
	zap       *zap.Logger
	logConfig *Config
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
	//for _, field := range fields {
	//	if field.Type == zapcore.ErrorType {
	//		log.Printf("%+v\n", field.Interface)
	//	}
	//}
	defaultLogger.error(msg, fields...)
}
func Errors(msg string, errs ...error) {
	lock.RLock()
	defer lock.RUnlock()
	fields := make([]zap.Field, len(errs))
	for i, e := range errs {
		fields[i] = zap.Error(e)
	}
	//for _, err := range errs {
	//	log.Printf("%+v\n", err)
	//}
	defaultLogger.error(msg, fields...)
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
	for _, field := range fields {
		if field.Type == zapcore.ErrorType {
			log.Printf("%+v\n", field.Interface)
		}
	}
	defaultLogger.panic(msg, fields...)
}
func PanicErrors(msg string, errs ...error) {
	lock.RLock()
	defer lock.RUnlock()
	fields := make([]zap.Field, len(errs))
	for i, e := range errs {
		fields[i] = zap.Error(e)
	}
	for _, err := range errs {
		log.Printf("%+v\n", err)
	}
	defaultLogger.panic(msg, fields...)

}
func PrintPanic(errs ...error) {
	for _, err := range errs {
		log.Printf("%+v\n", err)
	}
}

func Sync() error {
	lock.RLock()
	defer lock.RUnlock()
	return defaultLogger.sync()
}

func getDefaultLogger() *logger {
	cores := zapcore.NewTee(getStdoutLogWriter())
	l := zap.New(cores, zap.AddCaller(), zap.AddCallerSkip(2))
	return &logger{
		zap: l,
	}
}
func InitLogger(logConfig *Config) {
	mode := logConfig.Write
	level, err := zapcore.ParseLevel(logConfig.Level)
	if err != nil {
		level = zapcore.InfoLevel
		Error("log level", zap.Error(err), zap.String("level", level.String()))
	}
	Info("Running Mode", zap.String("level", logConfig.Level), zap.Bool("run in the background", mode))
	if !mode {
		if len(logConfig.Path) > 0 {
			abs, err := filepath.Abs(logConfig.Path)
			if err == nil {
				logConfig.Path = abs
				Info(" log save path", zap.String("logPath", logConfig.Path))
				cores := zapcore.NewTee(getFileLogWriter(logConfig), getStdoutLogWriter())
				l := zap.New(cores, zap.AddCaller(), zap.AddCallerSkip(2), zap.IncreaseLevel(level))
				lock.Lock()
				defer lock.Unlock()
				defaultLogger = &logger{
					zap: l,
				}
				return
			}
			Error("log file path", zap.Error(err))
		} else {
			Info(" log save path has not been set")
		}
	}
	lock.Lock()
	defer lock.Unlock()
	defaultLogger = &logger{
		zap: zap.New(getStdoutLogWriter(), zap.AddCaller(), zap.AddCallerSkip(2), zap.IncreaseLevel(level)),
	}
}
