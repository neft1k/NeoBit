package config

import (
	"os"
	"strings"
)

type LoggerConfig struct {
	Development      bool
	Level            string
	Encoding         string
	OutputPaths      []string
	ErrorOutputPaths []string
}

func GetLoggerConfig() LoggerConfig {
	return LoggerConfig{
		Development:      getEnvBool("LOG_DEVELOPMENT", false),
		Level:            getEnv("LOG_LEVEL", "info"),
		Encoding:         getEnv("LOG_ENCODING", "json"),
		OutputPaths:      splitEnv("LOG_OUTPUT_PATHS", "stdout"),
		ErrorOutputPaths: splitEnv("LOG_ERROR_OUTPUT_PATHS", "stderr"),
	}
}

func splitEnv(key string, def string) []string {
	val := os.Getenv(key)
	if val == "" {
		val = def
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func getEnv(key string, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvBool(key string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return def
	}
	switch v {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}
