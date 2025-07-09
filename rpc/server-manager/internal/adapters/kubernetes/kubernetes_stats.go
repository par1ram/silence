package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetServerStats получает статистику сервера из Kubernetes
func (k *KubernetesAdapter) GetServerStats(ctx context.Context, serverID string) (*domain.ServerStats, error) {
	// Получаем Pod'ы для сервера
	pods, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", serverID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return &domain.ServerStats{
			ServerID:     serverID,
			CPUUsage:     0.0,
			MemoryUsage:  0.0,
			StorageUsage: 0.0,
			NetworkIn:    0,
			NetworkOut:   0,
			Uptime:       0,
			RequestCount: 0,
			ResponseTime: 0.0,
			ErrorRate:    0.0,
			Timestamp:    time.Now(),
		}, nil
	}

	// Получаем метрики для первого Pod'а
	_ = pods.Items[0] // TODO: использовать pod для получения метрик
	stats := &domain.ServerStats{
		ServerID:     serverID,
		CPUUsage:     0.0,
		MemoryUsage:  0.0,
		StorageUsage: 0.0,
		NetworkIn:    0,
		NetworkOut:   0,
		Uptime:       0,
		RequestCount: 0,
		ResponseTime: 0.0,
		ErrorRate:    0.0,
		Timestamp:    time.Now(),
	}

	// TODO: Получить реальные метрики из Kubernetes Metrics API
	// Это требует дополнительной настройки metrics-server

	return stats, nil
}

// GetServerHealth получает здоровье сервера из Kubernetes
func (k *KubernetesAdapter) GetServerHealth(ctx context.Context, serverID string) (*domain.ServerHealth, error) {
	// Получаем Deployment
	deployment, err := k.clientset.AppsV1().Deployments(k.namespace).Get(ctx, serverID, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// Получаем Pod'ы
	pods, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", serverID),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	health := &domain.ServerHealth{
		ServerID:    serverID,
		Status:      domain.ServerStatusError,
		Message:     "",
		LastCheckAt: time.Now(),
		Checks:      []map[string]interface{}{},
	}

	// Определяем статус на основе состояния Deployment и Pod'ов
	if deployment.Status.ReadyReplicas > 0 {
		health.Status = domain.ServerStatusRunning
		health.Message = "Server is running"
	} else if deployment.Status.Replicas == 0 {
		health.Status = domain.ServerStatusStopped
		health.Message = "Server is stopped"
	} else {
		health.Status = domain.ServerStatusError
		health.Message = "Server has issues"
	}

	// Проверяем состояние Pod'ов
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodFailed {
			health.Status = "error"
			health.Message = "Pod failed"
			break
		}
	}

	return health, nil
}
