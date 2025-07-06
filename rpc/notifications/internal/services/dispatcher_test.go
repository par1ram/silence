package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/notifications/internal/domain"
	"github.com/par1ram/silence/rpc/notifications/internal/services"
	. "github.com/par1ram/silence/rpc/notifications/internal/services/mocks"
)

//go:generate mockgen -destination=mock_dispatcher.go -package=services_test github.com/par1ram/silence/rpc/notifications/internal/services DeliveryAdapter

func TestDispatcherService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DispatcherService Suite")
}

var _ = Describe("DispatcherService", func() {
	var dispatcherService *services.DispatcherService
	var ctx context.Context
	var mockEmail *MockDeliveryAdapter
	var mockSMS *MockDeliveryAdapter
	var mockTelegram *MockDeliveryAdapter
	var mockPush *MockDeliveryAdapter
	var mockWebhook *MockDeliveryAdapter
	var mockSlack *MockDeliveryAdapter
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockEmail = NewMockDeliveryAdapter(ctrl)
		mockSMS = NewMockDeliveryAdapter(ctrl)
		mockTelegram = NewMockDeliveryAdapter(ctrl)
		mockPush = NewMockDeliveryAdapter(ctrl)
		mockWebhook = NewMockDeliveryAdapter(ctrl)
		mockSlack = NewMockDeliveryAdapter(ctrl)
		dispatcherService = services.NewDispatcherService(
			mockEmail,
			mockSMS,
			mockTelegram,
			mockPush,
			mockWebhook,
			mockSlack,
			nil, // analytics integration
		)
		ctx = context.Background()
	})

	Describe("Dispatch", func() {
		It("should dispatch email notification", func() {
			notification := &domain.Notification{
				ID:         "notification-123",
				Type:       domain.NotificationTypeSystemAlert,
				Priority:   domain.NotificationPriorityNormal,
				Title:      "Test notification",
				Message:    "This is a test notification",
				Channels:   []domain.NotificationChannel{domain.NotificationChannelEmail},
				Recipients: []string{"user@example.com"},
				Source:     "test",
				SourceID:   "test-123",
				Status:     domain.NotificationStatusPending,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			mockEmail.EXPECT().Send(ctx, notification).Return(nil)

			err := dispatcherService.Dispatch(ctx, notification)

			Expect(err).To(BeNil())
		})

		It("should handle email delivery error", func() {
			notification := &domain.Notification{
				ID:         "notification-123",
				Type:       domain.NotificationTypeSystemAlert,
				Priority:   domain.NotificationPriorityNormal,
				Title:      "Test notification",
				Message:    "This is a test notification",
				Channels:   []domain.NotificationChannel{domain.NotificationChannelEmail},
				Recipients: []string{"user@example.com"},
				Source:     "test",
				SourceID:   "test-123",
				Status:     domain.NotificationStatusPending,
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
			}

			expectedError := context.DeadlineExceeded
			mockEmail.EXPECT().Send(ctx, notification).Return(expectedError)

			err := dispatcherService.Dispatch(ctx, notification)

			Expect(err).To(BeNil()) // Dispatch doesn't return errors, just logs them
		})
	})
})
