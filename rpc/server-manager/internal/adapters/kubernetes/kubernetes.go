package kubernetes

import (
	"context"
	"fmt"

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
	"k8s.io/utils/ptr"
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
			Replicas: ptr.To(int32(1)),
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

	deployment.Spec.Replicas = ptr.To(int32(1))
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

	deployment.Spec.Replicas = ptr.To(int32(0))
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
