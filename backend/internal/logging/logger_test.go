package logging_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"ledabeer/backend/internal/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_LogLevels(t *testing.T) {
	// Should support all log levels
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.DebugLevel,
		Output: buf,
		Format: logging.JSONFormat,
	})

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	// Verify all messages logged
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.Len(t, lines, 5) // 4 messages + empty line
}

func TestLogger_LevelFiltering(t *testing.T) {
	// Should filter messages below configured level
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.WarnLevel,
		Output: buf,
	})

	logger.Debug("debug message") // Should not appear
	logger.Info("info message")   // Should not appear
	logger.Warn("warn message")   // Should appear
	logger.Error("error message") // Should appear

	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	assert.Len(t, lines, 3) // 2 messages + empty line
}

func TestLogger_StructuredFields(t *testing.T) {
	// Should support structured field logging
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.InfoLevel,
		Output: buf,
		Format: logging.JSONFormat,
	})

	logger.Info("user action",
		logging.String("user_id", "user123"),
		logging.String("action", "login"),
		logging.Int("attempts", 3),
	)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "user action", logEntry["message"])
	assert.Equal(t, "user123", logEntry["user_id"])
	assert.Equal(t, "login", logEntry["action"])
	assert.Equal(t, float64(3), logEntry["attempts"])
}

func TestLogger_JSONFormat(t *testing.T) {
	// Should output valid JSON
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.InfoLevel,
		Output: buf,
		Format: logging.JSONFormat,
	})

	logger.Info("test message")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Contains(t, logEntry, "timestamp")
	assert.Contains(t, logEntry, "level")
	assert.Contains(t, logEntry, "message")
}

func TestLogger_TextFormat(t *testing.T) {
	// Should output human-readable text
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.InfoLevel,
		Output: buf,
		Format: logging.TextFormat,
	})

	logger.Info("test message")

	output := buf.String()
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "test message")
}
