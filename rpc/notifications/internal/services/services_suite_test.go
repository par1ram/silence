package services_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/notifications/internal/domain"
	"github.com/par1ram/silence/rpc/notifications/internal/services"
	. "github.com/par1ram/silence/rpc/notifications/internal/services/mocks"
)

//go:generate mockgen -destination=mock_dispatcher.go -package=services_test github.com/par1ram/silence/rpc/notifications/internal/services DeliveryAdapter

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}

var _ = Describe("Services", func() {
	var (
		dispatcherService   *services.DispatcherService
		integration         *services.AnalyticsIntegration
		mockAnalyticsServer *httptest.Server
		ctx                 context.Context
		mockEmail           *MockDeliveryAdapter
		mockSMS             *MockDeliveryAdapter
		mockTelegram        *MockDeliveryAdapter
		mockPush            *MockDeliveryAdapter
		mockWebhook         *MockDeliveryAdapter
		mockSlack           *MockDeliveryAdapter
		ctrl                *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockEmail = NewMockDeliveryAdapter(ctrl)
		mockSMS = NewMockDeliveryAdapter(ctrl)
		mockTelegram = NewMockDeliveryAdapter(ctrl)
		mockPush = NewMockDeliveryAdapter(ctrl)
		mockWebhook = NewMockDeliveryAdapter(ctrl)
		mockSlack = NewMockDeliveryAdapter(ctrl)

		mockAnalyticsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/metrics/errors"))
			var metric services.NotificationDeliveryMetric
			Expect(json.NewDecoder(r.Body).Decode(&metric)).To(Succeed())
			Expect(metric.Name).To(Or(Equal("notification_delivery_success"), Equal("notification_delivery_error")))
			w.WriteHeader(http.StatusCreated)
		}))
		integration = services.NewAnalyticsIntegration(mockAnalyticsServer.URL)

		dispatcherService = services.NewDispatcherService(
			mockEmail,
			mockSMS,
			mockTelegram,
			mockPush,
			mockWebhook,
			mockSlack,
			integration,
		)
		ctx = context.Background()
	})

	AfterEach(func() {
		mockAnalyticsServer.Close()
	})

	Describe("DispatcherService", func() {
		It("should dispatch email notification", func() {
			notification := &domain.Notification{
				Recipients: []string{"test@test.com"},
				Channels:   []domain.NotificationChannel{domain.NotificationChannelEmail},
			}

			mockEmail.EXPECT().Send(ctx, notification).Return(nil)

			err := dispatcherService.Dispatch(ctx, notification)

			Expect(err).To(BeNil())
		})

		It("should handle email delivery error", func() {
			notification := &domain.Notification{
				Recipients: []string{"test@test.com"},
				Channels:   []domain.NotificationChannel{domain.NotificationChannelEmail},
			}

			expectedError := context.DeadlineExceeded
			mockEmail.EXPECT().Send(ctx, notification).Return(expectedError)

			err := dispatcherService.Dispatch(ctx, notification)

			Expect(err).To(BeNil()) // Dispatch doesn't return errors, just logs them
		})

		It("should dispatch sms notification", func() {
			notification := &domain.Notification{
				Recipients: []string{"test@test.com"},
				Channels:   []domain.NotificationChannel{domain.NotificationChannelSMS},
			}

			mockSMS.EXPECT().Send(ctx, notification).Return(nil)

			err := dispatcherService.Dispatch(ctx, notification)

			Expect(err).To(BeNil())
		})

		It("should dispatch telegram notification", func() {
			notification := &domain.Notification{
				Recipients: []string{"test@test.com"},
				Channels:   []domain.NotificationChannel{domain.NotificationChannelTelegram},
			}

			mockTelegram.EXPECT().Send(ctx, notification).Return(nil)

			err := dispatcherService.Dispatch(ctx, notification)

			Expect(err).To(BeNil())
		})
	})

	Describe("AnalyticsIntegration", func() {
		It("should send a success metric", func() {
			notification := &domain.Notification{
				Recipients: []string{"test@test.com"},
			}
			err := integration.RecordNotificationDelivery(ctx, notification, domain.NotificationChannelEmail)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should send an error metric", func() {
			notification := &domain.Notification{
				Recipients: []string{"test@test.com"},
			}
			err := integration.RecordNotificationError(ctx, notification, domain.NotificationChannelEmail, context.DeadlineExceeded)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
