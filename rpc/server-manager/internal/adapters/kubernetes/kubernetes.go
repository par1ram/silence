package kubernetes

import (
	"context"
	"fmt"
	"time"

	"github.com/par1ram/silence/rpc/server-manager/internal/domain"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/utils/pointer"
)

// KubernetesAdapter адаптер для работы с Kubernetes
type KubernetesAdapter struct {
	clientset *kubernetes.Clientset
	namespace string
	logger    *zap.Logger
}

// NewKubernetesAdapter создает новый Kubernetes адаптер
func NewKubernetesAdapter(kubeconfig string, namespace string, logger *zap.Logger) (*KubernetesAdapter, error) {
	var config *rest.Config
	var err error

	// Пытаемся загрузить конфигурацию из файла или из кластера
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &KubernetesAdapter{
		clientset: clientset,
		namespace: namespace,
		logger:    logger,
	}, nil
}

// CreateServer создает сервер в Kubernetes (Deployment + Service)
func (k *KubernetesAdapter) CreateServer(ctx context.Context, server *domain.Server) error {
	// Создаем Deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":     server.Name,
				"type":    string(server.Type),
				"region":  server.Region,
				"managed": "server-manager",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": server.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": server.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  server.Name,
							Image: "silence/vpn-core:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "vpn",
									ContainerPort: 51820,
									Protocol:      corev1.ProtocolUDP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "SERVER_ID",
									Value: server.ID,
								},
								{
									Name:  "SERVER_TYPE",
									Value: string(server.Type),
								},
								{
									Name:  "REGION",
									Value: server.Region,
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("512Mi"),
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(8080),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       5,
							},
						},
					},
				},
			},
		},
	}

	// Создаем Deployment
	_, err := k.clientset.AppsV1().Deployments(k.namespace).Create(ctx, deployment, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create deployment: %w", err)
	}

	// Создаем Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":     server.Name,
				"managed": "server-manager",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": server.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(8080),
					Protocol:   corev1.ProtocolTCP,
				},
				{
					Name:       "vpn",
					Port:       51820,
					TargetPort: intstr.FromInt(51820),
					Protocol:   corev1.ProtocolUDP,
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

	// Создаем Service
	_, err = k.clientset.CoreV1().Services(k.namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	k.logger.Info("kubernetes server created",
		zap.String("name", server.Name),
		zap.String("namespace", k.namespace))

	return nil
}

// StartServer запускает сервер в Kubernetes
func (k *KubernetesAdapter) StartServer(ctx context.Context, serverID string) error {
	// Масштабируем Deployment до 1 реплики
	deployment, err := k.clientset.AppsV1().Deployments(k.namespace).Get(ctx, serverID, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.Spec.Replicas = pointer.Int32(1)
	_, err = k.clientset.AppsV1().Deployments(k.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	k.logger.Info("kubernetes server started", zap.String("id", serverID))
	return nil
}

// StopServer останавливает сервер в Kubernetes
func (k *KubernetesAdapter) StopServer(ctx context.Context, serverID string) error {
	// Масштабируем Deployment до 0 реплик
	deployment, err := k.clientset.AppsV1().Deployments(k.namespace).Get(ctx, serverID, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.Spec.Replicas = pointer.Int32(0)
	_, err = k.clientset.AppsV1().Deployments(k.namespace).Update(ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	k.logger.Info("kubernetes server stopped", zap.String("id", serverID))
	return nil
}

// DeleteServer удаляет сервер из Kubernetes
func (k *KubernetesAdapter) DeleteServer(ctx context.Context, serverID string) error {
	// Удаляем Deployment
	err := k.clientset.AppsV1().Deployments(k.namespace).Delete(ctx, serverID, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}

	// Удаляем Service
	err = k.clientset.CoreV1().Services(k.namespace).Delete(ctx, serverID, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	k.logger.Info("kubernetes server deleted", zap.String("id", serverID))
	return nil
}

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
			ServerID:    serverID,
			CPU:         0,
			Memory:      0,
			Disk:        0,
			Network:     0,
			Connections: 0,
			Timestamp:   time.Now(),
		}, nil
	}

	// Получаем метрики для первого Pod'а
	_ = pods.Items[0] // TODO: использовать pod для получения метрик
	stats := &domain.ServerStats{
		ServerID:    serverID,
		CPU:         0,
		Memory:      0,
		Disk:        0,
		Network:     0,
		Connections: 0,
		Timestamp:   time.Now(),
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
		ServerID:  serverID,
		Status:    "unknown",
		Message:   "",
		Timestamp: time.Now(),
	}

	// Определяем статус на основе состояния Deployment и Pod'ов
	if deployment.Status.ReadyReplicas > 0 {
		health.Status = "running"
		health.Message = "Server is running"
	} else if deployment.Status.Replicas == 0 {
		health.Status = "stopped"
		health.Message = "Server is stopped"
	} else {
		health.Status = "starting"
		health.Message = "Server is starting"
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

// ScaleServer масштабирует сервер в Kubernetes
func (k *KubernetesAdapter) ScaleServer(ctx context.Context, serverID string, replicas int32) error {
	deployment, err := k.clientset.AppsV1().Deployments(k.namespace).Get(ctx, serverID, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment: %w", err)
	}

	deployment.Spec.Replicas = pointer.Int32(replicas)
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
