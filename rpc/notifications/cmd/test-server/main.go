package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	grpcserver "github.com/par1ram/silence/rpc/notifications/internal/adapters/grpc"
	"github.com/par1ram/silence/rpc/notifications/internal/services"
)

func main() {
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

	// Создаем интеграцию с analytics (заглушка)
	analytics := &services.AnalyticsIntegration{}

	// Создаем dispatcher с заглушками
	dispatcher := services.NewDispatcherService(
		&services.StubDeliveryAdapter{Name: "email"},
		&services.StubDeliveryAdapter{Name: "sms"},
		&services.StubDeliveryAdapter{Name: "telegram"},
		&services.StubDeliveryAdapter{Name: "push"},
		&services.StubDeliveryAdapter{Name: "webhook"},
		&services.StubDeliveryAdapter{Name: "slack"},
		analytics,
	)

	// gRPC сервер
	grpcSrv := grpcserver.NewServer(8080, dispatcher)
	go func() {
		if err := grpcSrv.Start(); err != nil {
			log.Fatalf("[main] grpc server error: %v", err)
		}
	}()

	log.Println("[main] standalone notifications service started on :8080")

	// Ждем завершения
	<-ctx.Done()
	log.Println("[main] notifications service stopped")
}
