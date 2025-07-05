package kubernetes

import (
	"context"
	"fmt"

	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// ScaleServer масштабирует сервер в Kubernetes
func (k *KubernetesAdapter) ScaleServer(ctx context.Context, serverID string, replicas int32) error {
	deployment, err := k.clientset.AppsV1().Deployments(k.namespace).Get(ctx, serverID, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.Spec.Replicas = ptr.To(replicas)
	_, err = k.clientset.AppsV1().Deployments(k.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	k.logger.Info("kubernetes server scaled",
		zap.String("id", serverID),
		zap.Int32("replicas", replicas))

	return nil
}

// ListServers получает список серверов из Kubernetes
func (k *KubernetesAdapter) ListServers(ctx context.Context) ([]*domain.Server, error) {
	deployments, err := k.clientset.AppsV1().Deployments(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "managed=server-manager",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	var servers []*domain.Server
	for _, deployment := range deployments.Items {
		server := &domain.Server{
			ID:        deployment.Name,
			Name:      deployment.Name,
			Type:      domain.ServerType(deployment.Labels["type"]),
			Region:    deployment.Labels["region"],
			Status:    domain.ServerStatus("unknown"),
			CreatedAt: deployment.CreationTimestamp.Time,
			UpdatedAt: deployment.CreationTimestamp.Time,
		}

		// Определяем статус
		if deployment.Status.ReadyReplicas > 0 {
			server.Status = domain.ServerStatus("running")
		} else if deployment.Status.Replicas == 0 {
			server.Status = domain.ServerStatus("stopped")
		} else {
			server.Status = domain.ServerStatus("starting")
		}

		servers = append(servers, server)
	}

	return servers, nil
}
