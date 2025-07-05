package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"go.uber.org/zap"
)

// RedisDashboardRepository реализация репозитория дашбордов с Redis
type RedisDashboardRepository struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisDashboardRepository создает новый репозиторий дашбордов
func NewRedisDashboardRepository(addr, password string, db int, logger *zap.Logger) (*RedisDashboardRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &RedisDashboardRepository{
		client: client,
		logger: logger,
	}, nil
}

// CreateDashboard создает дашборд
func (r *RedisDashboardRepository) CreateDashboard(ctx context.Context, dashboard domain.DashboardConfig) error {
	data, err := json.Marshal(dashboard)
	if err != nil {
		return fmt.Errorf("failed to marshal dashboard: %w", err)
	}

	key := fmt.Sprintf("dashboard:%s", dashboard.ID)
	if err := r.client.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("failed to save dashboard: %w", err)
	}

	// Добавляем в список всех дашбордов
	if err := r.client.SAdd(ctx, "dashboards", dashboard.ID).Err(); err != nil {
		return fmt.Errorf("failed to add dashboard to list: %w", err)
	}

	return nil
}

// GetDashboard получает дашборд
func (r *RedisDashboardRepository) GetDashboard(ctx context.Context, id string) (*domain.DashboardConfig, error) {
	key := fmt.Sprintf("dashboard:%s", id)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("dashboard not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	var dashboard domain.DashboardConfig
	if err := json.Unmarshal(data, &dashboard); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dashboard: %w", err)
	}

	return &dashboard, nil
}

// UpdateDashboard обновляет дашборд
func (r *RedisDashboardRepository) UpdateDashboard(ctx context.Context, dashboard domain.DashboardConfig) error {
	return r.CreateDashboard(ctx, dashboard)
}

// DeleteDashboard удаляет дашборд
func (r *RedisDashboardRepository) DeleteDashboard(ctx context.Context, id string) error {
	key := fmt.Sprintf("dashboard:%s", id)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	// Удаляем из списка всех дашбордов
	if err := r.client.SRem(ctx, "dashboards", id).Err(); err != nil {
		return fmt.Errorf("failed to remove dashboard from list: %w", err)
	}

	return nil
}

// ListDashboards получает список дашбордов
func (r *RedisDashboardRepository) ListDashboards(ctx context.Context) ([]domain.DashboardConfig, error) {
	ids, err := r.client.SMembers(ctx, "dashboards").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard list: %w", err)
	}

	var dashboards []domain.DashboardConfig
	for _, id := range ids {
		dashboard, err := r.GetDashboard(ctx, id)
		if err != nil {
			r.logger.Warn("Failed to get dashboard", zap.String("id", id), zap.String("error", err.Error()))
			continue
		}
		dashboards = append(dashboards, *dashboard)
	}

	return dashboards, nil
}

// Close закрывает соединение с Redis
func (r *RedisDashboardRepository) Close() {
	r.client.Close()
}
