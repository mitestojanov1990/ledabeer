package logging

import (
	"os"
	"sync"
)

var (
	defaultLogger *Logger
	defaultMutex  sync.RWMutex
)

func init() {
	defaultLogger = NewLogger(&Config{
		Level:  InfoLevel,
		Output: os.Stdout,
		Format: TextFormat,
	})
}

func Default() *Logger {
	defaultMutex.RLock()
	defer defaultMutex.RUnlock()
	return defaultLogger
}

func SetDefault(logger *Logger) {
	defaultMutex.Lock()
	defer defaultMutex.Unlock()
	defaultLogger = logger
}

func Debug(msg string, fields ...Field) {
	Default().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	Default().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	Default().Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	Default().Error(msg, fields...)
}
