package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	// Test case 1: Default values
	cfg := Load()
	assert.Equal(t, "8080", cfg.HTTPPort)
	assert.Equal(t, "9090", cfg.GRPCPort)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "1.0.0", cfg.Version)
	assert.Equal(t, "/etc/wireguard", cfg.WireGuardDir)
	assert.Equal(t, "wg0", cfg.Interface)
	assert.Equal(t, 51820, cfg.ListenPort)
	assert.Equal(t, 1420, cfg.MTU)

	// Test case 2: Environment variables
	httpPort := "8888"
	grpcPort := "9999"
	logLevel := "debug"
	version := "2.0.0"
	wgDir := "/tmp/wireguard"
	iface := "wg1"
	listenPort := 51821
	mtu := 1500

	os.Setenv("HTTP_PORT", httpPort)
	os.Setenv("GRPC_PORT", grpcPort)
	os.Setenv("LOG_LEVEL", logLevel)
	os.Setenv("VERSION", version)
	os.Setenv("WIREGUARD_DIR", wgDir)
	os.Setenv("WIREGUARD_INTERFACE", iface)
	os.Setenv("WIREGUARD_LISTEN_PORT", strconv.Itoa(listenPort))
	os.Setenv("WIREGUARD_MTU", strconv.Itoa(mtu))

	cfg = Load()
	assert.Equal(t, httpPort, cfg.HTTPPort)
	assert.Equal(t, grpcPort, cfg.GRPCPort)
	assert.Equal(t, logLevel, cfg.LogLevel)
	assert.Equal(t, version, cfg.Version)
	assert.Equal(t, wgDir, cfg.WireGuardDir)
	assert.Equal(t, iface, cfg.Interface)
	assert.Equal(t, listenPort, cfg.ListenPort)
	assert.Equal(t, mtu, cfg.MTU)

	// Clean up environment variables
	os.Unsetenv("HTTP_PORT")
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("VERSION")
	os.Unsetenv("WIREGUARD_DIR")
	os.Unsetenv("WIREGUARD_INTERFACE")
	os.Unsetenv("WIREGUARD_LISTEN_PORT")
	os.Unsetenv("WIREGUARD_MTU")
}
