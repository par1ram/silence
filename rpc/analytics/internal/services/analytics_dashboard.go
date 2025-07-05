package services

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"go.uber.org/zap"
)

// Методы для работы с дашбордами

func (s *AnalyticsServiceImpl) CreateDashboard(ctx context.Context, dashboard domain.DashboardConfig) error {
	if dashboard.CreatedAt.IsZero() {
		dashboard.CreatedAt = time.Now()
	}
	dashboard.UpdatedAt = time.Now()

	if err := s.dashboardRepo.CreateDashboard(ctx, dashboard); err != nil {
		s.logger.Error("Failed to create dashboard",
			zap.String("error", err.Error()),
			zap.String("dashboard_id", dashboard.ID),
		)
		return fmt.Errorf("failed to create dashboard: %w", err)
	}

	s.logger.Info("Dashboard created", zap.String("dashboard_id", dashboard.ID))
	return nil
}

func (s *AnalyticsServiceImpl) GetDashboard(ctx context.Context, id string) (*domain.DashboardConfig, error) {
	dashboard, err := s.dashboardRepo.GetDashboard(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get dashboard",
			zap.String("error", err.Error()),
			zap.String("dashboard_id", id),
		)
		return nil, fmt.Errorf("failed to get dashboard: %w", err)
	}

	return dashboard, nil
}

func (s *AnalyticsServiceImpl) UpdateDashboard(ctx context.Context, dashboard domain.DashboardConfig) error {
	dashboard.UpdatedAt = time.Now()

	if err := s.dashboardRepo.UpdateDashboard(ctx, dashboard); err != nil {
		s.logger.Error("Failed to update dashboard",
			zap.String("error", err.Error()),
			zap.String("dashboard_id", dashboard.ID),
		)
		return fmt.Errorf("failed to update dashboard: %w", err)
	}

	s.logger.Info("Dashboard updated", zap.String("dashboard_id", dashboard.ID))
	return nil
}

func (s *AnalyticsServiceImpl) DeleteDashboard(ctx context.Context, id string) error {
	if err := s.dashboardRepo.DeleteDashboard(ctx, id); err != nil {
		s.logger.Error("Failed to delete dashboard",
			zap.String("error", err.Error()),
			zap.String("dashboard_id", id),
		)
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	s.logger.Info("Dashboard deleted", zap.String("dashboard_id", id))
	return nil
}

func (s *AnalyticsServiceImpl) ListDashboards(ctx context.Context) ([]domain.DashboardConfig, error) {
	dashboards, err := s.dashboardRepo.ListDashboards(ctx)
	if err != nil {
		s.logger.Error("Failed to list dashboards", zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to list dashboards: %w", err)
	}

	return dashboards, nil
}
