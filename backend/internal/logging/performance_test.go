package logging_test

import (
	"bytes"
	"io"
	"testing"

	"ledabeer/backend/internal/logging"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Sampling(t *testing.T) {
	// Should sample high-frequency logs
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.DebugLevel,
		Output: buf,
		Sampling: &logging.SamplingConfig{
			Initial:    10,
			Thereafter: 100,
		},
	})

	// Log 1000 debug messages
	for i := 0; i < 1000; i++ {
		logger.Debug("high frequency log")
	}

	lines := bytes.Count(buf.Bytes(), []byte("\n"))
	// Should have ~10 initial + ~10 sampled = ~20 logs
	assert.Less(t, lines, 50)
	assert.Greater(t, lines, 10)
}

func TestLogger_LazyEvaluation(t *testing.T) {
	// Should not evaluate expensive operations for filtered logs
	buf := &bytes.Buffer{}
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.ErrorLevel,
		Output: buf,
	})

	evaluated := false
	logger.Debug("test", logging.Lazy("expensive", func() interface{} {
		evaluated = true
		return "expensive value"
	}))

	assert.False(t, evaluated, "Should not evaluate for filtered log")
}

func BenchmarkLogger_JSON(b *testing.B) {
	logger := logging.NewLogger(&logging.Config{
		Level:  logging.InfoLevel,
		Output: io.Discard,
		Format: logging.JSONFormat,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message",
			logging.String("key1", "value1"),
			logging.Int("key2", 42),
		)
	}
}
