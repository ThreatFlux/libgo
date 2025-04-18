package logger

import "go.uber.org/zap/zapcore"

// Logger defines the interface for logging
type Logger interface {
	// Debug logs a message at debug level
	Debug(msg string, fields ...Field)

	// Info logs a message at info level
	Info(msg string, fields ...Field)

	// Warn logs a message at warning level
	Warn(msg string, fields ...Field)

	// Error logs a message at error level
	Error(msg string, fields ...Field)

	// Fatal logs a message at fatal level then calls os.Exit(1)
	Fatal(msg string, fields ...Field)

	// WithFields returns a new Logger with the given fields added
	WithFields(fields ...Field) Logger

	// WithError returns a new Logger with the given error attached
	WithError(err error) Logger

	// Sync flushes any buffered log entries
	Sync() error
}

// Field represents a structured log field
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string Field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer Field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 Field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Uint64 creates a uint64 Field
func Uint64(key string, value uint64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 Field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool Field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error Field
func Error(err error) Field {
	return Field{Key: "error", Value: err}
}

// Any creates a field with any value
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field
func Duration(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Time creates a time field
func Time(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Binary creates a field for binary data
func Binary(key string, value []byte) Field {
	return Field{Key: key, Value: value}
}

// Level represents a logging level
type Level zapcore.Level

// Log levels
const (
	DebugLevel Level = Level(zapcore.DebugLevel)
	InfoLevel  Level = Level(zapcore.InfoLevel)
	WarnLevel  Level = Level(zapcore.WarnLevel)
	ErrorLevel Level = Level(zapcore.ErrorLevel)
	FatalLevel Level = Level(zapcore.FatalLevel)
)

// String returns the string representation of the Level
func (l Level) String() string {
	return zapcore.Level(l).String()
}
