package services

import (
	"context"
	"log"

	"github.com/par1ram/silence/rpc/notifications/internal/domain"
)

// DispatcherService маршрутизирует события на delivery-адаптеры
type DispatcherService struct {
	Email     DeliveryAdapter
	SMS       DeliveryAdapter
	Telegram  DeliveryAdapter
	Push      DeliveryAdapter
	Webhook   DeliveryAdapter
	Slack     DeliveryAdapter
	Analytics *AnalyticsIntegration
}

// DeliveryAdapter интерфейс для отправки уведомлений
// (stub для MVP)
type DeliveryAdapter interface {
	Send(ctx context.Context, notification *domain.Notification) error
}

// NewDispatcherService конструктор
func NewDispatcherService(email, sms, telegram, push, webhook, slack DeliveryAdapter, analytics *AnalyticsIntegration) *DispatcherService {
	return &DispatcherService{
		Email:     email,
		SMS:       sms,
		Telegram:  telegram,
		Push:      push,
		Webhook:   webhook,
		Slack:     slack,
		Analytics: analytics,
	}
}

// Dispatch маршрутизирует уведомление по каналам
func (d *DispatcherService) Dispatch(ctx context.Context, notification *domain.Notification) error {
	for _, channel := range notification.Channels {
		var err error
		switch channel {
		case domain.NotificationChannelEmail:
			if d.Email != nil {
				err = d.Email.Send(ctx, notification)
				if err != nil {
					log.Printf("[dispatcher] email send error: %v", err)
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationError(ctx, notification, channel, err)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationError: %v", errAnalytics)
						}
					}
				} else {
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationDelivery(ctx, notification, channel)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationDelivery: %v", errAnalytics)
						}
					}
				}
			}
		case domain.NotificationChannelSMS:
			if d.SMS != nil {
				err = d.SMS.Send(ctx, notification)
				if err != nil {
					log.Printf("[dispatcher] sms send error: %v", err)
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationError(ctx, notification, channel, err)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationError: %v", errAnalytics)
						}
					}
				} else {
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationDelivery(ctx, notification, channel)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationDelivery: %v", errAnalytics)
						}
					}
				}
			}
		case domain.NotificationChannelTelegram:
			if d.Telegram != nil {
				err = d.Telegram.Send(ctx, notification)
				if err != nil {
					log.Printf("[dispatcher] telegram send error: %v", err)
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationError(ctx, notification, channel, err)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationError: %v", errAnalytics)
						}
					}
				} else {
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationDelivery(ctx, notification, channel)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationDelivery: %v", errAnalytics)
						}
					}
				}
			}
		case domain.NotificationChannelPush:
			if d.Push != nil {
				err = d.Push.Send(ctx, notification)
				if err != nil {
					log.Printf("[dispatcher] push send error: %v", err)
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationError(ctx, notification, channel, err)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationError: %v", errAnalytics)
						}
					}
				} else {
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationDelivery(ctx, notification, channel)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationDelivery: %v", errAnalytics)
						}
					}
				}
			}
		case domain.NotificationChannelWebhook:
			if d.Webhook != nil {
				err = d.Webhook.Send(ctx, notification)
				if err != nil {
					log.Printf("[dispatcher] webhook send error: %v", err)
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationError(ctx, notification, channel, err)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationError: %v", errAnalytics)
						}
					}
				} else {
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationDelivery(ctx, notification, channel)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationDelivery: %v", errAnalytics)
						}
					}
				}
			}
		case domain.NotificationChannelSlack:
			if d.Slack != nil {
				err = d.Slack.Send(ctx, notification)
				if err != nil {
					log.Printf("[dispatcher] slack send error: %v", err)
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationError(ctx, notification, channel, err)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationError: %v", errAnalytics)
						}
					}
				} else {
					if d.Analytics != nil {
						errAnalytics := d.Analytics.RecordNotificationDelivery(ctx, notification, channel)
						if errAnalytics != nil {
							log.Printf("[dispatcher] analytics RecordNotificationDelivery: %v", errAnalytics)
						}
					}
				}
			}
		default:
			log.Printf("[dispatcher] unknown channel: %s", channel)
		}
	}
	return nil
}

// StubDeliveryAdapter простая заглушка для MVP
type StubDeliveryAdapter struct {
	Name string
}

func (s *StubDeliveryAdapter) Send(ctx context.Context, notification *domain.Notification) error {
	log.Printf("[stub-%s] Отправка уведомления: type=%s, title=%s, recipients=%v", s.Name, notification.Type, notification.Title, notification.Recipients)
	return nil
}
