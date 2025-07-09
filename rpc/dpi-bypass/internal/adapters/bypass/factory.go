package bypass

import (
	"fmt"

	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/ports"
	"go.uber.org/zap"
)

// AdapterFactory фабрика для создания адаптеров обфускации
type AdapterFactory struct {
	logger *zap.Logger
}

// NewAdapterFactory создает новую фабрику адаптеров
func NewAdapterFactory(logger *zap.Logger) *AdapterFactory {
	return &AdapterFactory{
		logger: logger,
	}
}

// CreateAdapter создает адаптер для указанного метода обфускации
func (f *AdapterFactory) CreateAdapter(method domain.BypassMethod) (ports.BypassAdapter, error) {
	switch method {
	case domain.BypassMethodHTTPHeader:
		return NewCustomAdapter(f.logger), nil
	case domain.BypassMethodTLSHandshake:
		return NewCustomAdapter(f.logger), nil
	case domain.BypassMethodTCPFragment:
		return NewCustomAdapter(f.logger), nil
	case domain.BypassMethodUDPFragment:
		return NewCustomAdapter(f.logger), nil
	case domain.BypassMethodProxyChain:
		return NewCustomAdapter(f.logger), nil
	case domain.BypassMethodShadowsocks:
		return NewShadowsocksAdapter(f.logger), nil
	case domain.BypassMethodV2Ray:
		return NewV2RayAdapter(f.logger), nil
	case domain.BypassMethodObfs4:
		return NewObfs4Adapter(f.logger), nil
	case domain.BypassMethodCustom:
		return NewCustomAdapter(f.logger), nil
	default:
		return nil, fmt.Errorf("unsupported bypass method: %s", method)
	}
}

// CreateMultiAdapter создает мульти-адаптер, который может управлять несколькими методами
func (f *AdapterFactory) CreateMultiAdapter() *MultiBypassAdapter {
	return NewMultiBypassAdapter(f.logger)
}

// MultiBypassAdapter адаптер для управления несколькими методами обфускации
type MultiBypassAdapter struct {
	adapters map[domain.BypassMethod]ports.BypassAdapter
	logger   *zap.Logger
}

// NewMultiBypassAdapter создает новый мульти-адаптер
func NewMultiBypassAdapter(logger *zap.Logger) *MultiBypassAdapter {
	return &MultiBypassAdapter{
		adapters: make(map[domain.BypassMethod]ports.BypassAdapter),
		logger:   logger,
	}
}

// Start запускает bypass соединение с автоматическим выбором адаптера
func (m *MultiBypassAdapter) Start(config *domain.BypassConfig) error {
	// Получаем или создаем адаптер для данного метода
	adapter, exists := m.adapters[config.Method]
	if !exists {
		factory := NewAdapterFactory(m.logger)
		var err error
		adapter, err = factory.CreateAdapter(config.Method)
		if err != nil {
			return err
		}
		m.adapters[config.Method] = adapter
	}

	return adapter.Start(config)
}

// Stop останавливает bypass соединение
func (m *MultiBypassAdapter) Stop(id string) error {
	// Находим адаптер, который управляет данным соединением
	for _, adapter := range m.adapters {
		if adapter.IsRunning(id) {
			return adapter.Stop(id)
		}
	}

	return fmt.Errorf("bypass connection not found: %s", id)
}

// GetStats возвращает статистику bypass соединения
func (m *MultiBypassAdapter) GetStats(id string) (*domain.BypassStats, error) {
	// Находим адаптер, который управляет данным соединением
	for _, adapter := range m.adapters {
		if adapter.IsRunning(id) {
			return adapter.GetStats(id)
		}
	}

	return nil, nil
}

// IsRunning проверяет, запущено ли bypass соединение
func (m *MultiBypassAdapter) IsRunning(id string) bool {
	// Проверяем все адаптеры
	for _, adapter := range m.adapters {
		if adapter.IsRunning(id) {
			return true
		}
	}

	return false
}
