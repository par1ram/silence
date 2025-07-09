package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	grpcserver "github.com/par1ram/silence/rpc/notifications/internal/adapters/grpc"
	"github.com/par1ram/silence/rpc/notifications/internal/adapters/rabbitmq"
	"github.com/par1ram/silence/rpc/notifications/internal/config"
	"github.com/par1ram/silence/rpc/notifications/internal/domain"
	"github.com/par1ram/silence/rpc/notifications/internal/logic"
	"github.com/par1ram/silence/rpc/notifications/internal/services"
)

func main() {
	cfg := config.Load()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("[main] shutdown signal received")
		cancel()
	}()

	rmqCfg := &rabbitmq.RabbitMQConfig{
		URL:           cfg.RabbitMQ.URL,
		Exchange:      cfg.RabbitMQ.Exchange,
		Queue:         cfg.RabbitMQ.Queue,
		RoutingKey:    cfg.RabbitMQ.RoutingKey,
		ConsumerTag:   cfg.RabbitMQ.ConsumerTag,
		PrefetchCount: cfg.RabbitMQ.PrefetchCount,
	}

	rmq, err := rabbitmq.NewRabbitMQAdapter(rmqCfg)
	if err != nil {
		log.Fatalf("[main] failed to init RabbitMQ: %v", err)
	}
	defer rmq.Close()

	// Создаем интеграцию с analytics
	analytics := services.NewAnalyticsIntegration(cfg.Analytics.URL)

	dispatcher := services.NewDispatcherService(
		&services.StubDeliveryAdapter{Name: "email"},
		&services.StubDeliveryAdapter{Name: "sms"},
		&services.StubDeliveryAdapter{Name: "telegram"},
		&services.StubDeliveryAdapter{Name: "push"},
		&services.StubDeliveryAdapter{Name: "webhook"},
		&services.StubDeliveryAdapter{Name: "slack"},
		analytics,
	)
	consumer := logic.NewNotificationConsumer(dispatcher)

	// gRPC сервер
	grpcSrv := grpcserver.NewServer(cfg.Server.Port, dispatcher)
	go func() {
		if err := grpcSrv.Start(); err != nil {
			log.Fatalf("[main] grpc server error: %v", err)
		}
	}()

	log.Println("[main] notifications service started, waiting for events...")

	err = rmq.ConsumeEvents(ctx, func(event *domain.NotificationEvent) error {
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
		return consumer.Dispatcher.Dispatch(ctx, notification)
	})
	if err != nil {
		log.Fatalf("[main] failed to start consumer: %v", err)
	}

	// Ждем завершения
	<-ctx.Done()
	log.Println("[main] notifications service stopped")
}
