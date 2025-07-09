package notifications

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/par1ram/silence/api/gateway/api/proto/notifications"
)

// Client gRPC клиент для notifications сервиса
type Client struct {
	conn   *grpc.ClientConn
	client pb.NotificationsServiceClient
}

// NewClient создает новый gRPC клиент для notifications сервиса
func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(10*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to notifications service: %w", err)
	}

	client := pb.NewNotificationsServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close закрывает соединение
func (c *Client) Close() error {
	return c.conn.Close()
}

// Health проверяет здоровье сервиса
func (c *Client) Health(ctx context.Context) (*pb.HealthResponse, error) {
	return c.client.Health(ctx, &pb.HealthRequest{})
}

// DispatchNotification отправляет уведомление
func (c *Client) DispatchNotification(ctx context.Context, req *pb.DispatchNotificationRequest) (*pb.DispatchNotificationResponse, error) {
	return c.client.DispatchNotification(ctx, req)
}

// GetNotification получает уведомление по ID
func (c *Client) GetNotification(ctx context.Context, id string) (*pb.Notification, error) {
	req := &pb.GetNotificationRequest{
		Id: id,
	}
	return c.client.GetNotification(ctx, req)
}

// ListNotifications получает список уведомлений
func (c *Client) ListNotifications(ctx context.Context, recipient string, notType pb.NotificationType, status pb.NotificationStatus, source string, limit, offset int32) (*pb.ListNotificationsResponse, error) {
	req := &pb.ListNotificationsRequest{
		Recipient: recipient,
		Type:      notType,
		Status:    status,
		Source:    source,
		Limit:     limit,
		Offset:    offset,
	}
	return c.client.ListNotifications(ctx, req)
}

// UpdateNotificationStatus обновляет статус уведомления
func (c *Client) UpdateNotificationStatus(ctx context.Context, id string, status pb.NotificationStatus, errorMsg string) (*pb.Notification, error) {
	req := &pb.UpdateNotificationStatusRequest{
		Id:     id,
		Status: status,
		Error:  errorMsg,
	}
	return c.client.UpdateNotificationStatus(ctx, req)
}

// CreateTemplate создает шаблон уведомления
func (c *Client) CreateTemplate(ctx context.Context, notType pb.NotificationType, priority pb.NotificationPriority, title, message string, channels []pb.NotificationChannel, enabled bool) (*pb.NotificationTemplate, error) {
	req := &pb.CreateTemplateRequest{
		Type:     notType,
		Priority: priority,
		Title:    title,
		Message:  message,
		Channels: channels,
		Enabled:  enabled,
	}
	return c.client.CreateTemplate(ctx, req)
}

// GetTemplate получает шаблон по ID
func (c *Client) GetTemplate(ctx context.Context, id string) (*pb.NotificationTemplate, error) {
	req := &pb.GetTemplateRequest{
		Id: id,
	}
	return c.client.GetTemplate(ctx, req)
}

// ListTemplates получает список шаблонов
func (c *Client) ListTemplates(ctx context.Context, notType pb.NotificationType, enabled bool, limit, offset int32) (*pb.ListTemplatesResponse, error) {
	req := &pb.ListTemplatesRequest{
		Type:    notType,
		Enabled: enabled,
		Limit:   limit,
		Offset:  offset,
	}
	return c.client.ListTemplates(ctx, req)
}

// UpdateTemplate обновляет шаблон
func (c *Client) UpdateTemplate(ctx context.Context, id string, notType pb.NotificationType, priority pb.NotificationPriority, title, message string, channels []pb.NotificationChannel, enabled bool) (*pb.NotificationTemplate, error) {
	req := &pb.UpdateTemplateRequest{
		Id:       id,
		Type:     notType,
		Priority: priority,
		Title:    title,
		Message:  message,
		Channels: channels,
		Enabled:  enabled,
	}
	return c.client.UpdateTemplate(ctx, req)
}

// DeleteTemplate удаляет шаблон
func (c *Client) DeleteTemplate(ctx context.Context, id string) (*pb.DeleteTemplateResponse, error) {
	req := &pb.DeleteTemplateRequest{
		Id: id,
	}
	return c.client.DeleteTemplate(ctx, req)
}

// GetUserPreferences получает настройки пользователя
func (c *Client) GetUserPreferences(ctx context.Context, userID string, notType pb.NotificationType) (*pb.NotificationPreference, error) {
	req := &pb.GetUserPreferencesRequest{
		UserId: userID,
		Type:   notType,
	}
	return c.client.GetUserPreferences(ctx, req)
}

// UpdateUserPreferences обновляет настройки пользователя
func (c *Client) UpdateUserPreferences(ctx context.Context, userID string, notType pb.NotificationType, channels []pb.NotificationChannel, enabled bool, schedule *pb.NotificationSchedule) (*pb.NotificationPreference, error) {
	req := &pb.UpdateUserPreferencesRequest{
		UserId:   userID,
		Type:     notType,
		Channels: channels,
		Enabled:  enabled,
		Schedule: schedule,
	}
	return c.client.UpdateUserPreferences(ctx, req)
}

// GetNotificationStats получает статистику уведомлений
func (c *Client) GetNotificationStats(ctx context.Context, source string, notType pb.NotificationType) (*pb.NotificationStats, error) {
	req := &pb.GetNotificationStatsRequest{
		Source: source,
		Type:   notType,
	}
	return c.client.GetNotificationStats(ctx, req)
}
