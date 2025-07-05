package ports

import (
	"context"
	"time"

	"github.com/par1ram/silence/rpc/notifications/internal/domain"
)

// NotificationService интерфейс для работы с уведомлениями
type NotificationService interface {
	// Создание и отправка уведомлений
	CreateNotification(ctx context.Context, notification *domain.Notification) error
	SendNotification(ctx context.Context, notification *domain.Notification) error
	SendBulkNotifications(ctx context.Context, notifications []*domain.Notification) error

	// Получение уведомлений
	GetNotification(ctx context.Context, id string) (*domain.Notification, error)
	GetNotifications(ctx context.Context, filter NotificationFilter) ([]*domain.Notification, error)
	GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]*domain.Notification, error)

	// Обновление статуса
	UpdateNotificationStatus(ctx context.Context, id string, status domain.NotificationStatus) error
	RetryFailedNotification(ctx context.Context, id string) error
	CancelNotification(ctx context.Context, id string) error

	// Статистика
	GetNotificationStats(ctx context.Context, filter NotificationStatsFilter) (*domain.NotificationStats, error)

	// Обработка событий из RabbitMQ
	ProcessNotificationEvent(ctx context.Context, event *domain.NotificationEvent) error
}

// NotificationTemplateService интерфейс для работы с шаблонами
type NotificationTemplateService interface {
	CreateTemplate(ctx context.Context, template *domain.NotificationTemplate) error
	GetTemplate(ctx context.Context, id string) (*domain.NotificationTemplate, error)
	GetTemplates(ctx context.Context) ([]*domain.NotificationTemplate, error)
	UpdateTemplate(ctx context.Context, template *domain.NotificationTemplate) error
	DeleteTemplate(ctx context.Context, id string) error
	GetTemplateByType(ctx context.Context, notificationType domain.NotificationType) (*domain.NotificationTemplate, error)
}

// NotificationPreferenceService интерфейс для работы с предпочтениями
type NotificationPreferenceService interface {
	SetUserPreference(ctx context.Context, preference *domain.NotificationPreference) error
	GetUserPreference(ctx context.Context, userID string, notificationType domain.NotificationType) (*domain.NotificationPreference, error)
	GetUserPreferences(ctx context.Context, userID string) ([]*domain.NotificationPreference, error)
	DeleteUserPreference(ctx context.Context, userID string, notificationType domain.NotificationType) error
	SetDefaultPreferences(ctx context.Context, userID string) error
}

// NotificationDeliveryService интерфейс для доставки уведомлений
type NotificationDeliveryService interface {
	SendEmail(ctx context.Context, email *domain.EmailNotification) error
	SendSMS(ctx context.Context, sms *domain.SMSNotification) error
	SendPush(ctx context.Context, push *domain.PushNotification) error
	SendTelegram(ctx context.Context, telegram *domain.TelegramNotification) error
	SendWebhook(ctx context.Context, webhook *domain.WebhookNotification) error
	SendSlack(ctx context.Context, message string, channel string) error
}

// NotificationQueueService интерфейс для работы с очередью
type NotificationQueueService interface {
	PublishEvent(ctx context.Context, event *domain.NotificationEvent) error
	ConsumeEvents(ctx context.Context, handler func(*domain.NotificationEvent) error) error
	GetQueueStats(ctx context.Context) (QueueStats, error)
}

// NotificationFilter фильтр для получения уведомлений
type NotificationFilter struct {
	UserID   string                      `json:"user_id,omitempty"`
	Type     domain.NotificationType     `json:"type,omitempty"`
	Status   domain.NotificationStatus   `json:"status,omitempty"`
	Priority domain.NotificationPriority `json:"priority,omitempty"`
	Source   string                      `json:"source,omitempty"`
	Channel  domain.NotificationChannel  `json:"channel,omitempty"`
	FromDate *time.Time                  `json:"from_date,omitempty"`
	ToDate   *time.Time                  `json:"to_date,omitempty"`
	Limit    int                         `json:"limit,omitempty"`
	Offset   int                         `json:"offset,omitempty"`
}

// NotificationStatsFilter фильтр для статистики
type NotificationStatsFilter struct {
	FromDate *time.Time              `json:"from_date,omitempty"`
	ToDate   *time.Time              `json:"to_date,omitempty"`
	Source   string                  `json:"source,omitempty"`
	Type     domain.NotificationType `json:"type,omitempty"`
}

// QueueStats статистика очереди
type QueueStats struct {
	QueueName     string `json:"queue_name"`
	MessageCount  int    `json:"message_count"`
	ConsumerCount int    `json:"consumer_count"`
	ReadyCount    int    `json:"ready_count"`
	UnackedCount  int    `json:"unacked_count"`
}

// NotificationRepository интерфейс для работы с базой данных
type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	GetByID(ctx context.Context, id string) (*domain.Notification, error)
	GetByFilter(ctx context.Context, filter NotificationFilter) ([]*domain.Notification, error)
	Update(ctx context.Context, notification *domain.Notification) error
	Delete(ctx context.Context, id string) error
	GetPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error)
	GetFailedNotifications(ctx context.Context, limit int) ([]*domain.Notification, error)
	GetStats(ctx context.Context, filter NotificationStatsFilter) (*domain.NotificationStats, error)
}

// NotificationTemplateRepository интерфейс для работы с шаблонами в БД
type NotificationTemplateRepository interface {
	Create(ctx context.Context, template *domain.NotificationTemplate) error
	GetByID(ctx context.Context, id string) (*domain.NotificationTemplate, error)
	GetByType(ctx context.Context, notificationType domain.NotificationType) (*domain.NotificationTemplate, error)
	GetAll(ctx context.Context) ([]*domain.NotificationTemplate, error)
	Update(ctx context.Context, template *domain.NotificationTemplate) error
	Delete(ctx context.Context, id string) error
}

// NotificationPreferenceRepository интерфейс для работы с предпочтениями в БД
type NotificationPreferenceRepository interface {
	Set(ctx context.Context, preference *domain.NotificationPreference) error
	Get(ctx context.Context, userID string, notificationType domain.NotificationType) (*domain.NotificationPreference, error)
	GetAll(ctx context.Context, userID string) ([]*domain.NotificationPreference, error)
	Delete(ctx context.Context, userID string, notificationType domain.NotificationType) error
	SetDefaults(ctx context.Context, userID string) error
}
