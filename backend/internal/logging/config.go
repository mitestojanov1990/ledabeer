package logging

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

func ConfigFromEnv() (*Config, error) {
	config := &Config{
		Level:  InfoLevel,
		Format: TextFormat,
		Output: os.Stdout,
	}

	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		level, err := parseLevel(levelStr)
		if err != nil {
			return nil, err
		}
		config.Level = level
	}

	if formatStr := os.Getenv("LOG_FORMAT"); formatStr != "" {
		if strings.ToLower(formatStr) == "json" {
			config.Format = JSONFormat
		}
	}

	return config, nil
}

func ConfigFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		Output string `yaml:"output"`
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	result := &Config{
		Level:  InfoLevel,
		Format: TextFormat,
		Output: os.Stdout,
	}

	if config.Level != "" {
		level, err := parseLevel(config.Level)
		if err != nil {
			return nil, err
		}
		result.Level = level
	}

	if config.Format != "" {
		if strings.ToLower(config.Format) == "json" {
			result.Format = JSONFormat
		}
	}

	return result, nil
}

func parseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn", "warning":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	default:
		return InfoLevel, fmt.Errorf("unknown log level: %s", s)
	}
}
