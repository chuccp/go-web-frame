package log

import (
	"log"
	"os"
	"path/filepath"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/term"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level string
	Path  string
	Write bool
}

func (c *Config) Key() string {
	return "web.log"
}
func defaultConfig() *Config {
	return &Config{
		Level: "info",
		Path:  "",
		Write: false,
	}
}

func IsBackgroundMode() bool {
	isStdoutTTY := term.IsTerminal(int(os.Stdout.Fd()))
	isStderrTTY := term.IsTerminal(int(os.Stderr.Fd()))
	_, hasNohup := os.LookupEnv("NOHUP")
	Info("运行模式", zap.Bool("isStdoutTTY", isStdoutTTY), zap.Bool("isStderrTTY", isStderrTTY), zap.Bool("hasNohup", hasNohup))
	return hasNohup && (isStdoutTTY || isStderrTTY)
}

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
	defaultLogger.error(msg, fields...)
}
func Errors(msg string, errs ...error) {
	lock.RLock()
	defer lock.RUnlock()
	fields := make([]zap.Field, len(errs))
	for i, e := range errs {
		fields[i] = zap.Error(e)
	}
	log.Println(errs)
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
	defaultLogger.panic(msg, fields...)
}
func PanicErrors(msg string, errs ...error) {
	lock.RLock()
	defer lock.RUnlock()
	fields := make([]zap.Field, len(errs))
	for i, e := range errs {
		fields[i] = zap.Error(e)
	}
	log.Println(errs)
	defaultLogger.panic(msg, fields...)

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
		Error("日志级别错误", zap.Error(err), zap.String("level", level.String()))
	}
	Info("运行模式", zap.String("level", logConfig.Level), zap.Bool("是否后台运行写日志", mode))
	if !mode {
		if len(logConfig.Path) > 0 {
			abs, err := filepath.Abs(logConfig.Path)
			if err == nil {
				logConfig.Path = abs
				Info("日志保存路径", zap.String("logPath", logConfig.Path))
				cores := zapcore.NewTee(getFileLogWriter(logConfig.Path), getStdoutLogWriter())
				l := zap.New(cores, zap.AddCaller(), zap.AddCallerSkip(2), zap.IncreaseLevel(level))
				lock.Lock()
				defer lock.Unlock()
				defaultLogger = &logger{
					zap: l,
				}
				return
			}
			Error("日志文件路径错误", zap.Error(err))
		} else {
			Info("日志保存路径没有设置")
		}
	}
	lock.Lock()
	defer lock.Unlock()
	defaultLogger = &logger{
		zap: zap.New(getStdoutLogWriter(), zap.AddCaller(), zap.AddCallerSkip(2), zap.IncreaseLevel(level)),
	}
}
