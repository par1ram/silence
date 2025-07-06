package services

import (
	"testing"

	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"github.com/stretchr/testify/assert"
)

func TestHealthService(t *testing.T) {
	t.Run("создание health сервиса", func(t *testing.T) {
		serviceName := "vpn-core"
		version := "1.0.0"

		healthService := NewHealthService(serviceName, version)

		assert.NotNil(t, healthService)
		assert.IsType(t, &HealthService{}, healthService)

		// Проверяем, что сервис реализует интерфейс
		var _ ports.HealthService = healthService
	})

	t.Run("получение статуса здоровья", func(t *testing.T) {
		serviceName := "vpn-core"
		version := "1.0.0"

		healthService := NewHealthService(serviceName, version)
		health := healthService.GetHealth()

		assert.NotNil(t, health)
		assert.Equal(t, serviceName, health.Service)
		assert.Equal(t, version, health.Version)
		assert.Equal(t, "ok", health.Status)
		assert.NotZero(t, health.Timestamp)
	})

	t.Run("создание сервиса с пустыми параметрами", func(t *testing.T) {
		healthService := NewHealthService("", "")

		assert.NotNil(t, healthService)
		health := healthService.GetHealth()

		assert.NotNil(t, health)
		assert.Equal(t, "", health.Service)
		assert.Equal(t, "", health.Version)
		assert.Equal(t, "ok", health.Status)
	})

	t.Run("создание сервиса с длинными параметрами", func(t *testing.T) {
		longServiceName := "very-long-service-name-for-testing-purposes"
		longVersion := "1.2.3.4.5.6.7.8.9.10"

		healthService := NewHealthService(longServiceName, longVersion)
		health := healthService.GetHealth()

		assert.NotNil(t, health)
		assert.Equal(t, longServiceName, health.Service)
		assert.Equal(t, longVersion, health.Version)
	})
}
