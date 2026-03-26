package logx

import (
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// LambdaLogger provides zerolog-based logging for Lambda (same behavior as golambda.Logger).
type LambdaLogger struct {
	zeroLogger zerolog.Logger
}

// NewLambdaLogger returns a new LambdaLogger using LOG_LEVEL.
func NewLambdaLogger(logLevel string) *LambdaLogger {
	var zeroLogLevel zerolog.Level
	switch strings.ToLower(logLevel) {
	case "trace":
		zeroLogLevel = zerolog.TraceLevel
	case "debug":
		zeroLogLevel = zerolog.DebugLevel
	case "info":
		zeroLogLevel = zerolog.InfoLevel
	case "error":
		zeroLogLevel = zerolog.ErrorLevel
	default:
		zeroLogLevel = zerolog.InfoLevel
	}

	var writer io.Writer = zerolog.ConsoleWriter{Out: os.Stdout}
	if _, ok := os.LookupEnv("AWS_LAMBDA_FUNCTION_NAME"); ok {
		writer = os.Stdout
	}

	logger := zerolog.New(writer).Level(zeroLogLevel).With().Timestamp().Logger()
	return &LambdaLogger{
		zeroLogger: logger,
	}
}

// LogEntry is one record of logging.
type LogEntry struct {
	logger *LambdaLogger
	values map[string]interface{}
}

// Entry returns a new LogEntry.
func (x *LambdaLogger) Entry() *LogEntry {
	return &LogEntry{
		logger: x,
		values: make(map[string]interface{}),
	}
}

// Trace logs at trace level.
func (x *LambdaLogger) Trace(msg string) { x.Entry().Trace(msg) }

// Debug logs at debug level.
func (x *LambdaLogger) Debug(msg string) { x.Entry().Debug(msg) }

// Info logs at info level.
func (x *LambdaLogger) Info(msg string) { x.Entry().Info(msg) }

// Error logs at error level.
func (x *LambdaLogger) Error(msg string) { x.Entry().Error(msg) }

// Set adds a permanent field to the logger.
func (x *LambdaLogger) Set(key string, value interface{}) {
	x.zeroLogger = x.zeroLogger.With().Interface(key, value).Logger()
}

// With starts a log entry with one field.
func (x *LambdaLogger) With(key string, value interface{}) *LogEntry {
	entry := x.Entry()
	entry.values[key] = value
	return entry
}

// With adds a field to this entry.
func (x *LogEntry) With(key string, value interface{}) *LogEntry {
	x.values[key] = value
	return x
}

func (x *LogEntry) bind(ev *zerolog.Event) {
	for k, v := range x.values {
		ev.Interface(k, v)
	}
}

// Trace emits a trace-level message.
func (x *LogEntry) Trace(msg string) {
	ev := x.logger.zeroLogger.Trace()
	x.bind(ev)
	ev.Msg(msg)
}

// Debug emits a debug-level message.
func (x *LogEntry) Debug(msg string) {
	ev := x.logger.zeroLogger.Debug()
	x.bind(ev)
	ev.Msg(msg)
}

// Info emits an info-level message.
func (x *LogEntry) Info(msg string) {
	ev := x.logger.zeroLogger.Info()
	x.bind(ev)
	ev.Msg(msg)
}

// Error emits an error-level message.
func (x *LogEntry) Error(msg string) {
	ev := x.logger.zeroLogger.Error()
	x.bind(ev)
	ev.Msg(msg)
}

// Logger is the default application logger (configured from LOG_LEVEL).
var Logger *LambdaLogger

func init() {
	logLevel := os.Getenv("LOG_LEVEL")
	Logger = NewLambdaLogger(logLevel)
}
