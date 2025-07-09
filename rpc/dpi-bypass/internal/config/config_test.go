package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Test case 1: Default values
	cfg := Load()
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "1.0.0", cfg.Version)
	assert.Equal(t, ":9091", cfg.GRPC.Address)

	// Test case 2: Environment variables
	logLevel := "debug"
	version := "2.0.0"
	grpcAddress := ":9092"

	os.Setenv("LOG_LEVEL", logLevel)
	os.Setenv("VERSION", version)
	os.Setenv("GRPC_ADDRESS", grpcAddress)

	cfg = Load()
	assert.Equal(t, logLevel, cfg.LogLevel)
	assert.Equal(t, version, cfg.Version)
	assert.Equal(t, grpcAddress, cfg.GRPC.Address)

	// Clean up environment variables
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("VERSION")
	os.Unsetenv("GRPC_ADDRESS")
}
