package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
)

// GetScalingPolicies получает политики масштабирования
func (s *ServerService) GetScalingPolicies(ctx context.Context) ([]*domain.ScalingPolicy, error) {
	if s.scalingRepo == nil {
		return []*domain.ScalingPolicy{}, nil
	}
	return s.scalingRepo.ListPolicies(ctx)
}

// CreateScalingPolicy создает политику масштабирования
func (s *ServerService) CreateScalingPolicy(ctx context.Context, policy *domain.ScalingPolicy) error {
	if s.scalingRepo == nil {
		return fmt.Errorf("scaling repository not initialized")
	}
	policy.ID = uuid.New().String()
	return s.scalingRepo.SavePolicy(ctx, policy)
}

// UpdateScalingPolicy обновляет политику масштабирования
func (s *ServerService) UpdateScalingPolicy(ctx context.Context, id string, policy *domain.ScalingPolicy) error {
	if s.scalingRepo == nil {
		return fmt.Errorf("scaling repository not initialized")
	}
	policy.ID = id
	return s.scalingRepo.UpdatePolicy(ctx, policy)
}

// DeleteScalingPolicy удаляет политику масштабирования
func (s *ServerService) DeleteScalingPolicy(ctx context.Context, id string) error {
	if s.scalingRepo == nil {
		return fmt.Errorf("scaling repository not initialized")
	}
	return s.scalingRepo.DeletePolicy(ctx, id)
}

// EvaluateScaling оценивает необходимость масштабирования
func (s *ServerService) EvaluateScaling(ctx context.Context) error {
	// TODO: реализовать логику масштабирования
	s.logger.Info("evaluating scaling policies")
	return nil
}
