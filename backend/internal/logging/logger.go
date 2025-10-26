package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

type Format int

const (
	JSONFormat Format = iota
	TextFormat
)

type SamplingConfig struct {
	Initial    int
	Thereafter int
}

type Config struct {
	Level    Level
	Output   io.Writer
	Format   Format
	Sampling *SamplingConfig
}

type Logger struct {
	config  Config
	fields  []Field
	mutex   sync.Mutex
	sampler *sampler
}

type sampler struct {
	initial    int
	thereafter int
	count      int
}

type Field struct {
	Key   string
	Value interface{}
}

func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Lazy(key string, fn func() interface{}) Field {
	return Field{Key: key, Value: fn}
}

func NewLogger(config *Config) *Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}

	logger := &Logger{config: *config}

	if config.Sampling != nil {
		logger.sampler = &sampler{
			initial:    config.Sampling.Initial,
			thereafter: config.Sampling.Thereafter,
		}
	}

	return logger
}

func (l *Logger) Debug(msg string, fields ...Field) {
	l.log(DebugLevel, msg, fields...)
}

func (l *Logger) Info(msg string, fields ...Field) {
	l.log(InfoLevel, msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	l.log(WarnLevel, msg, fields...)
}

func (l *Logger) Error(msg string, fields ...Field) {
	l.log(ErrorLevel, msg, fields...)
}

func (l *Logger) WithFields(fields ...Field) *Logger {
	newLogger := &Logger{
		config: l.config,
		fields: append(l.fields, fields...),
	}
	return newLogger
}

func (l *Logger) log(level Level, msg string, fields ...Field) {
	if level < l.config.Level {
		return
	}

	// Check sampling
	if l.sampler != nil && !l.shouldSample() {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     l.levelString(level),
		"message":   msg,
	}

	// Add context fields first
	for _, field := range l.fields {
		entry[field.Key] = l.evaluateField(field)
	}

	// Add log-specific fields
	for _, field := range fields {
		entry[field.Key] = l.evaluateField(field)
	}

	if l.config.Format == JSONFormat {
		json.NewEncoder(l.config.Output).Encode(entry)
	} else {
		fmt.Fprintf(l.config.Output, "%s [%s] %s\n",
			entry["timestamp"], entry["level"], msg)
	}
}

func (l *Logger) shouldSample() bool {
	if l.sampler == nil {
		return true
	}

	l.sampler.count++

	if l.sampler.count <= l.sampler.initial {
		return true
	}

	return (l.sampler.count-l.sampler.initial)%l.sampler.thereafter == 0
}

func (l *Logger) evaluateField(field Field) interface{} {
	if fn, ok := field.Value.(func() interface{}); ok {
		return fn()
	}
	return field.Value
}

func (l *Logger) levelString(level Level) string {
	switch level {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}
