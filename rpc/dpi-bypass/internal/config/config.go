package config

import (
	"os"
)

type Config struct {
	LogLevel string
	Version  string
	GRPC     GRPCConfig
}

type GRPCConfig struct {
	Address string
}

func Load() *Config {
	return &Config{
		LogLevel: getEnv("LOG_LEVEL", "info"),
		Version:  getEnv("VERSION", "1.0.0"),
		GRPC: GRPCConfig{
			Address: getEnv("GRPC_ADDRESS", ":9091"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
