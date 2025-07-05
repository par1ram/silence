package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/par1ram/silence/rpc/notifications/internal/domain"
	"github.com/par1ram/silence/rpc/notifications/internal/ports"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQAdapter адаптер для работы с RabbitMQ
type RabbitMQAdapter struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	config       *RabbitMQConfig
	exchangeName string
	queueName    string
	routingKey   string
}

// RabbitMQConfig конфигурация RabbitMQ
type RabbitMQConfig struct {
	URL           string
	Exchange      string
	Queue         string
	RoutingKey    string
	ConsumerTag   string
	PrefetchCount int
}

// NewRabbitMQAdapter создает новый адаптер RabbitMQ
func NewRabbitMQAdapter(config *RabbitMQConfig) (*RabbitMQAdapter, error) {
	adapter := &RabbitMQAdapter{
		config:       config,
		exchangeName: config.Exchange,
		queueName:    config.Queue,
		routingKey:   config.RoutingKey,
	}

	if err := adapter.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	if err := adapter.setupExchangeAndQueue(); err != nil {
		return nil, fmt.Errorf("failed to setup exchange and queue: %w", err)
	}

	return adapter, nil
}

// connect устанавливает соединение с RabbitMQ
func (r *RabbitMQAdapter) connect() error {
	var err error
	r.conn, err = amqp.Dial(r.config.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	r.channel, err = r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Устанавливаем prefetch count
	err = r.channel.Qos(
		r.config.PrefetchCount, // prefetch count
		0,                      // prefetch size
		false,                  // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	return nil
}

// setupExchangeAndQueue настраивает exchange и очередь
func (r *RabbitMQAdapter) setupExchangeAndQueue() error {
	// Объявляем exchange
	err := r.channel.ExchangeDeclare(
		r.exchangeName, // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Объявляем очередь
	queue, err := r.channel.QueueDeclare(
		r.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Привязываем очередь к exchange
	err = r.channel.QueueBind(
		queue.Name,     // queue name
		r.routingKey,   // routing key
		r.exchangeName, // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	return nil
}

// PublishEvent публикует событие в RabbitMQ
func (r *RabbitMQAdapter) PublishEvent(ctx context.Context, event *domain.NotificationEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Устанавливаем заголовки
	headers := amqp.Table{
		"source":     event.Source,
		"type":       string(event.Type),
		"priority":   string(event.Priority),
		"timestamp":  event.Timestamp.Unix(),
		"message_id": event.ID,
	}

	// Публикуем сообщение
	err = r.channel.PublishWithContext(ctx,
		r.exchangeName, // exchange
		r.routingKey,   // routing key
		false,          // mandatory
		false,          // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			Headers:      headers,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Published notification event: %s, type: %s, source: %s",
		event.ID, event.Type, event.Source)
	return nil
}

// ConsumeEvents потребляет события из RabbitMQ
func (r *RabbitMQAdapter) ConsumeEvents(ctx context.Context, handler func(*domain.NotificationEvent) error) error {
	msgs, err := r.channel.Consume(
		r.queueName,          // queue
		r.config.ConsumerTag, // consumer
		false,                // auto-ack
		false,                // exclusive
		false,                // no-local
		false,                // no-wait
		nil,                  // args
	)
	if err != nil {
		return fmt.Errorf("failed to start consuming: %w", err)
	}

	log.Printf("Started consuming notifications from queue: %s", r.queueName)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping notification consumer")
				return
			case msg := <-msgs:
				if err := r.processMessage(msg, handler); err != nil {
					log.Printf("Error processing message: %v", err)
					// Отклоняем сообщение и помещаем в очередь повторов
					msg.Nack(false, true)
				} else {
					// Подтверждаем обработку
					msg.Ack(false)
				}
			}
		}
	}()

	return nil
}

// processMessage обрабатывает отдельное сообщение
func (r *RabbitMQAdapter) processMessage(msg amqp.Delivery, handler func(*domain.NotificationEvent) error) error {
	var event domain.NotificationEvent
	if err := json.Unmarshal(msg.Body, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Добавляем информацию из заголовков
	if source, ok := msg.Headers["source"].(string); ok {
		event.Source = source
	}
	if timestamp, ok := msg.Headers["timestamp"].(int64); ok {
		event.Timestamp = time.Unix(timestamp, 0)
	}

	log.Printf("Processing notification event: %s, type: %s, source: %s",
		event.ID, event.Type, event.Source)

	return handler(&event)
}

// GetQueueStats получает статистику очереди
func (r *RabbitMQAdapter) GetQueueStats(ctx context.Context) (ports.QueueStats, error) {
	queue, err := r.channel.QueueInspect(r.queueName)
	if err != nil {
		return ports.QueueStats{}, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return ports.QueueStats{
		QueueName:     queue.Name,
		MessageCount:  queue.Messages,
		ConsumerCount: queue.Consumers,
		ReadyCount:    queue.Messages,
		UnackedCount:  0, // В amqp091 нет прямого доступа к unacknowledged сообщениям
	}, nil
}

// Close закрывает соединение с RabbitMQ
func (r *RabbitMQAdapter) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			log.Printf("Error closing channel: %v", err)
		}
	}
	if r.conn != nil {
		if err := r.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}
	return nil
}

// Reconnect переподключается к RabbitMQ
func (r *RabbitMQAdapter) Reconnect() error {
	r.Close()
	return r.connect()
}

// IsConnected проверяет, активно ли соединение
func (r *RabbitMQAdapter) IsConnected() bool {
	return r.conn != nil && !r.conn.IsClosed()
}
