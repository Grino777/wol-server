// Package logging sets up and configures logging.
package logging

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Grino777/wol-server/configs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Available logger levels.
const (
	// LevelDebug logs are typically voluminous, and are usually disabled in
	// production.
	LevelDebug = "DEBUG"
	// LevelInfo logs general application information.
	LevelInfo = "INFO"
	// LevelWarning logs are more important than Info, but don't need individual
	// human review. This is the default logging priority.
	LevelWarning = "WARNING"
	// LevelError logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	LevelError = "ERROR"
	// LevelCritical logs are particularly important errors. In development the
	// logger panics after writing the message.
	LevelCritical = "CRITICAL"
	// LevelAlert logs a message, then panics.
	LevelAlert = "ALERT"
	// LevelEmergency logs a message, then calls os.Exit(1).
	LevelEmergency = "EMERGENCY"
)

// NewLogger creates a new logger with the given configuration.
func NewLogger(level string, development bool) *zap.SugaredLogger {
	if development {
		encoder := zapcore.NewJSONEncoder(developmentEncoderConfig())
		writer := &colorAwareWriter{zapcore.Lock(os.Stderr)}
		core := zapcore.NewCore(encoder, writer, levelToZapLevel(level))
		logger := zap.New(core, zap.Development(), zap.AddCaller())
		return logger.Sugar()
	}

	config := &zap.Config{
		Level:            zap.NewAtomicLevelAt(levelToZapLevel(level)),
		Encoding:         encodingJSON,
		EncoderConfig:    productionEncoderConfig(),
		OutputPaths:      outputStderr,
		ErrorOutputPaths: outputStderr,
	}

	logger, err := config.Build()
	if err != nil {
		logger = zap.NewNop()
	}

	return logger.Sugar()
}

// colorAwareWriter is a WriteSyncer wrapper that unescapes JSON-encoded ANSI
// color codes (\u001b → raw ESC byte) before writing to the underlying writer.
// This allows the terminal to interpret color codes in JSON-formatted output.
type colorAwareWriter struct {
	zapcore.WriteSyncer
}

func (w *colorAwareWriter) Write(p []byte) (int, error) {
	n := len(p)
	replaced := bytes.ReplaceAll(p, []byte(`\u001b`), []byte("\x1b"))
	if _, err := w.WriteSyncer.Write(replaced); err != nil {
		return 0, err
	}
	return n, nil
}

// NewLoggerFromEnv creates a new logger from the environment. It consumes
// LOG_LEVEL for determining the level and LOG_MODE for determining the output
// parameters.
func NewLoggerFromEnv() *zap.SugaredLogger {
	// level := os.Getenv("LOG_LEVEL")
	level := configs.LogLevel
	// development := strings.ToLower(strings.TrimSpace(os.Getenv("LOG_MODE"))) == "development"
	development := strings.ToLower(strings.TrimSpace(configs.LogMode)) == "development"
	return NewLogger(level, development)
}

var (
	// defaultLogger is the default logger. It is initialized once per package
	// include upon calling DefaultLogger.
	defaultLogger     *zap.SugaredLogger
	defaultLoggerOnce sync.Once
)

// DefaultLogger returns the default logger for the package.
func DefaultLogger() *zap.SugaredLogger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = NewLoggerFromEnv()
	})
	return defaultLogger
}

// contextKey is a private string type to prevent collisions in the context map.
type contextKey string

// loggerKey points to the value in the context where the logger is stored.
const loggerKey = contextKey("logger")

// WithLogger creates a new context with the provided logger attached.
func WithLogger(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext returns the logger stored in the context. If no such logger
// exists, a default logger is returned.
func FromContext(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(loggerKey).(*zap.SugaredLogger); ok {
		return logger
	}
	return DefaultLogger()
}

func Info(ctx context.Context, args ...any) {
	FromContext(ctx).Info(args...)
}

func Infof(ctx context.Context, format string, args ...any) {
	FromContext(ctx).Infof(format, args...)
}

func Infow(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).Infow(msg, keysAndValues...)
}

func Debug(ctx context.Context, args ...any) {
	FromContext(ctx).Debug(args...)
}

func Debugf(ctx context.Context, format string, args ...any) {
	FromContext(ctx).Debugf(format, args...)
}

func Debugw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).Debugw(msg, keysAndValues...)
}

func Warn(ctx context.Context, args ...any) {
	FromContext(ctx).Warn(args...)
}

func Warnf(ctx context.Context, format string, args ...any) {
	FromContext(ctx).Warnf(format, args...)
}

func Warnw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).Warnw(msg, keysAndValues...)
}

func Error(ctx context.Context, args ...any) {
	FromContext(ctx).Error(args...)
}

func Errorf(ctx context.Context, format string, args ...any) {
	FromContext(ctx).Errorf(format, args...)
}

func Errorw(ctx context.Context, msg string, keysAndValues ...any) {
	FromContext(ctx).Errorw(msg, keysAndValues...)
}

const (
	timestamp  = "timestamp"
	severity   = "severity"
	logger     = "logger"
	caller     = "caller"
	message    = "message"
	stacktrace = "stacktrace"

	encodingJSON    = "json"
	encodingConsole = "console"
)

var outputStderr = []string{"stderr"}

func productionEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:    timestamp,
		LevelKey:   severity,
		NameKey:    logger,
		CallerKey:  caller,
		MessageKey: message,
		//StacktraceKey:  stacktrace,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    levelEncoder(),
		EncodeTime:     timeEncoder(),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func developmentEncoderConfig() zapcore.EncoderConfig {
	// Like zapcore.ShortCallerEncoder but in brackets.
	// ShortCallerEncoder serializes a caller in package/file:line format, trimming
	// all but the final directory from the full path.
	encodeCaller := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("(" + caller.TrimmedPath() + ")")
	}

	// Like zapcore.FullNameEncoder but in brackers.
	encodeName := func(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + loggerName + "]")
	}

	return zapcore.EncoderConfig{
		TimeKey:        timestamp,
		LevelKey:       severity,
		NameKey:        logger,
		CallerKey:      caller,
		MessageKey:     message,
		StacktraceKey:  stacktrace,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    levelColorEncoder(),
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   encodeCaller,
		EncodeName:     encodeName,
	}
}

// levelToZapLevel converts the given string to the appropriate zap level
// value.
func levelToZapLevel(s string) zapcore.Level {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarning:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	case LevelCritical:
		return zapcore.DPanicLevel
	case LevelAlert:
		return zapcore.PanicLevel
	case LevelEmergency:
		return zapcore.FatalLevel
	}

	return zapcore.WarnLevel
}

// levelEncoder transforms a zap level to the associated stackdriver level.
func levelEncoder() zapcore.LevelEncoder {
	return func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.DebugLevel:
			enc.AppendString(LevelDebug)
		case zapcore.InfoLevel:
			enc.AppendString(LevelInfo)
		case zapcore.WarnLevel:
			enc.AppendString(LevelWarning)
		case zapcore.ErrorLevel:
			enc.AppendString(LevelError)
		case zapcore.DPanicLevel:
			enc.AppendString(LevelCritical)
		case zapcore.PanicLevel:
			enc.AppendString(LevelAlert)
		case zapcore.FatalLevel:
			enc.AppendString(LevelEmergency)
		}
	}
}

// levelColorEncoder transforms a zap level to the associated stackdriver level
// wraps the levels in different colors.
func levelColorEncoder() zapcore.LevelEncoder {
	// Foreground colors.
	const (
		cRed     = 31
		cYellow  = 33
		cBlue    = 34
		cMagenta = 35
	)

	withColor := func(c uint8, s string) string {
		return fmt.Sprintf("\x1b[%dm%s\x1b[0m", c, s)
	}

	return func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		switch l {
		case zapcore.DebugLevel:
			enc.AppendString(withColor(cMagenta, LevelDebug))
		case zapcore.InfoLevel:
			enc.AppendString(withColor(cBlue, LevelInfo))
		case zapcore.WarnLevel:
			enc.AppendString(withColor(cYellow, LevelWarning))
		case zapcore.ErrorLevel:
			enc.AppendString(withColor(cRed, LevelError))
		case zapcore.DPanicLevel:
			enc.AppendString(withColor(cRed, LevelCritical))
		case zapcore.PanicLevel:
			enc.AppendString(withColor(cRed, LevelAlert))
		case zapcore.FatalLevel:
			enc.AppendString(withColor(cRed, LevelEmergency))
		}
	}
}

// timeEncoder encodes the time as RFC3339 nano.
func timeEncoder() zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(time.RFC3339Nano))
	}
}
