package logging_test

import (
	"bytes"
	"os"
	"testing"

	"ledabeer/backend/internal/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobal_DefaultLogger(t *testing.T) {
	// Should provide default logger
	logger := logging.Default()
	assert.NotNil(t, logger)
}

func TestGlobal_SetDefault(t *testing.T) {
	// Should allow setting default logger
	buf := &bytes.Buffer{}
	customLogger := logging.NewLogger(&logging.Config{
		Level:  logging.DebugLevel,
		Output: buf,
	})

	logging.SetDefault(customLogger)

	logging.Info("test message")
	assert.Contains(t, buf.String(), "test message")
}

func TestConfig_FromEnv(t *testing.T) {
	// Should load configuration from environment
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_OUTPUT", "stdout")
	defer os.Clearenv()

	config, err := logging.ConfigFromEnv()
	require.NoError(t, err)

	assert.Equal(t, logging.DebugLevel, config.Level)
	assert.Equal(t, logging.JSONFormat, config.Format)
}

func TestConfig_FromFile(t *testing.T) {
	// Should load configuration from file
	configFile := `
level: info
format: json
output: /var/log/app.log
`
	tmpFile := createTempFile(t, configFile)
	defer os.Remove(tmpFile)

	config, err := logging.ConfigFromFile(tmpFile)
	require.NoError(t, err)

	assert.Equal(t, logging.InfoLevel, config.Level)
}

func createTempFile(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}
