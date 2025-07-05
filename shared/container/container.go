package container

import (
	"sync"

	"go.uber.org/zap"
)

// Container DI контейнер для управления зависимостями
type Container struct {
	mu    sync.RWMutex
	items map[string]interface{}
}

// New создает новый DI контейнер
func New() *Container {
	return &Container{
		items: make(map[string]interface{}),
	}
}

// Register регистрирует зависимость в контейнере
func (c *Container) Register(name string, item interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[name] = item
}

// Get получает зависимость из контейнера
func (c *Container) Get(name string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, exists := c.items[name]
	return item, exists
}

// MustGet получает зависимость из контейнера или паникует если не найдена
func (c *Container) MustGet(name string) interface{} {
	item, exists := c.Get(name)
	if !exists {
		panic("dependency not found: " + name)
	}
	return item
}

// GetLogger получает логгер из контейнера
func (c *Container) GetLogger() *zap.Logger {
	return c.MustGet("logger").(*zap.Logger)
}

// RegisterLogger регистрирует логгер в контейнере
func (c *Container) RegisterLogger(logger *zap.Logger) {
	c.Register("logger", logger)
}
