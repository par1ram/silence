package domain

import (
	"time"
)

// NotificationType тип уведомления
type NotificationType string

const (
	// Системные уведомления
	NotificationTypeSystemAlert   NotificationType = "system_alert"
	NotificationTypeServerDown    NotificationType = "server_down"
	NotificationTypeServerUp      NotificationType = "server_up"
	NotificationTypeHighLoad      NotificationType = "high_load"
	NotificationTypeLowDiskSpace  NotificationType = "low_disk_space"
	NotificationTypeBackupFailed  NotificationType = "backup_failed"
	NotificationTypeBackupSuccess NotificationType = "backup_success"
	NotificationTypeUpdateFailed  NotificationType = "update_failed"
	NotificationTypeUpdateSuccess NotificationType = "update_success"

	// Пользовательские уведомления
	NotificationTypeUserLogin           NotificationType = "user_login"
	NotificationTypeUserLogout          NotificationType = "user_logout"
	NotificationTypeUserRegistered      NotificationType = "user_registered"
	NotificationTypeUserBlocked         NotificationType = "user_blocked"
	NotificationTypeUserUnblocked       NotificationType = "user_unblocked"
	NotificationTypePasswordReset       NotificationType = "password_reset"
	NotificationTypeSubscriptionExpired NotificationType = "subscription_expired"
	NotificationTypeSubscriptionRenewed NotificationType = "subscription_renewed"

	// VPN уведомления
	NotificationTypeVPNConnected    NotificationType = "vpn_connected"
	NotificationTypeVPNDisconnected NotificationType = "vpn_disconnected"
	NotificationTypeVPNError        NotificationType = "vpn_error"
	NotificationTypeBypassBlocked   NotificationType = "bypass_blocked"
	NotificationTypeBypassSuccess   NotificationType = "bypass_success"

	// Аналитика уведомления
	NotificationTypeMetricsAlert      NotificationType = "metrics_alert"
	NotificationTypeAnomalyDetected   NotificationType = "anomaly_detected"
	NotificationTypeThresholdExceeded NotificationType = "threshold_exceeded"
)

// NotificationPriority приоритет уведомления
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// NotificationChannel канал доставки уведомления
type NotificationChannel string

const (
	NotificationChannelEmail    NotificationChannel = "email"
	NotificationChannelSMS      NotificationChannel = "sms"
	NotificationChannelPush     NotificationChannel = "push"
	NotificationChannelTelegram NotificationChannel = "telegram"
	NotificationChannelWebhook  NotificationChannel = "webhook"
	NotificationChannelSlack    NotificationChannel = "slack"
)

// NotificationStatus статус уведомления
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSending   NotificationStatus = "sending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
	NotificationStatusCancelled NotificationStatus = "cancelled"
)

// Notification основная модель уведомления
type Notification struct {
	ID          string                 `json:"id"`
	Type        NotificationType       `json:"type"`
	Priority    NotificationPriority   `json:"priority"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Channels    []NotificationChannel  `json:"channels"`
	Recipients  []string               `json:"recipients"`
	Source      string                 `json:"source"`    // Сервис-источник
	SourceID    string                 `json:"source_id"` // ID объекта в источнике
	Status      NotificationStatus     `json:"status"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	SentAt      *time.Time             `json:"sent_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// NotificationTemplate шаблон уведомления
type NotificationTemplate struct {
	ID        string                `json:"id"`
	Type      NotificationType      `json:"type"`
	Priority  NotificationPriority  `json:"priority"`
	Title     string                `json:"title"`
	Message   string                `json:"message"`
	Channels  []NotificationChannel `json:"channels"`
	Enabled   bool                  `json:"enabled"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// NotificationPreference предпочтения пользователя по уведомлениям
type NotificationPreference struct {
	UserID    string                `json:"user_id"`
	Type      NotificationType      `json:"type"`
	Channels  []NotificationChannel `json:"channels"`
	Enabled   bool                  `json:"enabled"`
	Schedule  *NotificationSchedule `json:"schedule,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// NotificationSchedule расписание уведомлений
type NotificationSchedule struct {
	StartTime string `json:"start_time"` // HH:MM
	EndTime   string `json:"end_time"`   // HH:MM
	Timezone  string `json:"timezone"`
	Days      []int  `json:"days"` // 0=Sunday, 1=Monday, etc.
}

// NotificationEvent событие для RabbitMQ
type NotificationEvent struct {
	ID          string                 `json:"id"`
	Type        NotificationType       `json:"type"`
	Priority    NotificationPriority   `json:"priority"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Channels    []NotificationChannel  `json:"channels"`
	Recipients  []string               `json:"recipients"`
	Source      string                 `json:"source"`
	SourceID    string                 `json:"source_id"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NotificationStats статистика уведомлений
type NotificationStats struct {
	TotalSent    int64                          `json:"total_sent"`
	TotalFailed  int64                          `json:"total_failed"`
	TotalPending int64                          `json:"total_pending"`
	ByType       map[NotificationType]int64     `json:"by_type"`
	ByChannel    map[NotificationChannel]int64  `json:"by_channel"`
	ByPriority   map[NotificationPriority]int64 `json:"by_priority"`
	SuccessRate  float64                        `json:"success_rate"`
	AverageDelay float64                        `json:"average_delay"` // в секундах
}

// EmailNotification настройки email уведомления
type EmailNotification struct {
	To          []string `json:"to"`
	Cc          []string `json:"cc,omitempty"`
	Bcc         []string `json:"bcc,omitempty"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	HTMLBody    string   `json:"html_body,omitempty"`
	Attachments []string `json:"attachments,omitempty"`
}

// SMSNotification настройки SMS уведомления
type SMSNotification struct {
	To      string `json:"to"`
	Message string `json:"message"`
	Sender  string `json:"sender,omitempty"`
}

// PushNotification настройки push уведомления
type PushNotification struct {
	Token    string            `json:"token"`
	Title    string            `json:"title"`
	Body     string            `json:"body"`
	Data     map[string]string `json:"data,omitempty"`
	Badge    int               `json:"badge,omitempty"`
	Sound    string            `json:"sound,omitempty"`
	Category string            `json:"category,omitempty"`
	Priority string            `json:"priority,omitempty"`
}

// TelegramNotification настройки Telegram уведомления
type TelegramNotification struct {
	ChatID      string      `json:"chat_id"`
	Message     string      `json:"message"`
	ParseMode   string      `json:"parse_mode,omitempty"` // HTML, Markdown
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

// WebhookNotification настройки webhook уведомления
type WebhookNotification struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body"`
	Timeout int               `json:"timeout,omitempty"`
}
