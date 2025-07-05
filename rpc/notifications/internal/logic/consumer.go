package logic

import (
	"context"
	"encoding/json"
	"log"

	"github.com/par1ram/silence/rpc/notifications/internal/domain"
	"github.com/par1ram/silence/rpc/notifications/internal/services"
	amqp "github.com/rabbitmq/amqp091-go"
)

// NotificationConsumer слушает очередь и вызывает dispatcher
type NotificationConsumer struct {
	Dispatcher *services.DispatcherService
}

func NewNotificationConsumer(dispatcher *services.DispatcherService) *NotificationConsumer {
	return &NotificationConsumer{Dispatcher: dispatcher}
}

// HandleMessage обрабатывает одно сообщение из RabbitMQ
func (c *NotificationConsumer) HandleMessage(ctx context.Context, msg amqp.Delivery) error {
	var event domain.NotificationEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		log.Printf("[consumer] failed to unmarshal event: %v", err)
		return err
	}

	notification := &domain.Notification{
		ID:          event.ID,
		Type:        event.Type,
		Priority:    event.Priority,
		Title:       event.Title,
		Message:     event.Message,
		Data:        event.Data,
		Channels:    event.Channels,
		Recipients:  event.Recipients,
		Source:      event.Source,
		SourceID:    event.SourceID,
		ScheduledAt: event.ScheduledAt,
		CreatedAt:   event.Timestamp,
	}

	if err := c.Dispatcher.Dispatch(ctx, notification); err != nil {
		log.Printf("[consumer] dispatch error: %v", err)
		return err
	}

	log.Printf("[consumer] notification processed: id=%s, type=%s, channels=%v", notification.ID, notification.Type, notification.Channels)
	return nil
}
