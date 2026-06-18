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

// Config holds the logger configuration settings.
type Config struct {
	Level      string // Log level: debug, info, warn, error
	Path       string // Log file path
	Write      bool   // Whether to enable file-based logging
	MaxSize    int    // Max size per log file (MB), default 500
	MaxBackups int    // Max number of old log files retained, default 3
	MaxAge     int    // Max days to retain old log files, default 30
	Compress   bool   // Whether to compress old log files, default true
	LocalTime  bool   // Whether to use local time, default false (UTC)
}

// Key returns the configuration key for this logger config.
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

// TimestampFormat is the time format used in log output.
var TimestampFormat = "2006-01-02 15:04:05"
var defaultLogger = getDefaultLogger()

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(TimestampFormat)
	return zapcore.NewJSONEncoder(encoderConfig)
}
func getFileLogWriter(cfg *Config) zapcore.Core {
		// Apply default values
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

// Info logs an informational message.
func Info(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.info(msg, fields...)
}
// Error logs an error message.
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
// Errors logs an error message with one or more error values.
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
// Debug logs a debug-level message.
func Debug(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.debug(msg, fields...)
}
// Warn logs a warning message.
func Warn(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.warn(msg, fields...)
}
// Fatal logs a fatal message and calls os.Exit(1).
func Fatal(msg string, fields ...zap.Field) {
	lock.RLock()
	defer lock.RUnlock()
	defaultLogger.fatal(msg, fields...)
}
// Panic logs a panic message and then panics.
// Error fields are also printed via log.Printf for stack trace visibility.
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
// PanicErrors logs a panic message with error values and prints them via log.Printf.
// Unlike zap.Panic, this only logs the message without actually panicking.
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
// PrintPanic prints error stack traces via log.Printf without panicking.
func PrintPanic(errs ...error) {
	for _, err := range errs {
		log.Printf("%+v\n", err)
	}
}

// Sync flushes any buffered log entries.
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
// InitLogger initializes or reconfigures the global logger with the given config.
func InitLogger(logConfig *Config) {

	level, err := zapcore.ParseLevel(logConfig.Level)
	if err != nil {
		level = zapcore.InfoLevel
		Error("log level", zap.Error(err), zap.String("level", level.String()))
	}
	Info("log level", zap.String("level", level.String()), zap.Bool("write", logConfig.Write))
	if logConfig.Write {
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
