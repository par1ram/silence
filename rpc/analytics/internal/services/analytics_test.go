package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/analytics/internal/domain"
	"github.com/par1ram/silence/rpc/analytics/internal/services"
	"github.com/par1ram/silence/rpc/analytics/internal/services/mocks"
	"github.com/par1ram/silence/rpc/analytics/internal/telemetry"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mocks/mock_analytics.go -package=mocks github.com/par1ram/silence/rpc/analytics/internal/ports MetricsRepository,DashboardRepository,MetricsCollector,AlertService

func TestAnalyticsService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AnalyticsService Suite")
}

var _ = Describe("AnalyticsService", func() {
	var analyticsService *services.AnalyticsServiceImpl
	var ctx context.Context
	var logger *zap.Logger
	var mockMetricsRepo *mocks.MockMetricsRepository
	var mockDashboardRepo *mocks.MockDashboardRepository
	var mockCollector *mocks.MockMetricsCollector
	var mockAlertService *mocks.MockAlertService
	var ctrl *gomock.Controller
	var metricsCollector *telemetry.MetricsCollector
	var tracingManager *telemetry.TracingManager

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockMetricsRepo = mocks.NewMockMetricsRepository(ctrl)
		mockDashboardRepo = mocks.NewMockDashboardRepository(ctrl)
		mockCollector = mocks.NewMockMetricsCollector(ctrl)
		mockAlertService = mocks.NewMockAlertService(ctrl)
		logger = zap.NewNop()

		// Create mock telemetry components
		mockMeter := otel.GetMeterProvider().Meter("test")
		mockTracer := otel.GetTracerProvider().Tracer("test")

		var err error
		metricsCollector, err = telemetry.NewMetricsCollector(mockMeter, logger)
		Expect(err).ToNot(HaveOccurred())

		tracingManager = telemetry.NewTracingManager(mockTracer, logger)

		analyticsService = services.NewAnalyticsService(
			mockMetricsRepo,
			mockDashboardRepo,
			mockCollector,
			mockAlertService,
			logger,
			metricsCollector,
			tracingManager,
		).(*services.AnalyticsServiceImpl)
		ctx = context.Background()
	})

	Describe("GetServerLoadMetrics", func() {
		It("should return server load metrics", func() {
			startTime := time.Now().Add(-24 * time.Hour)
			endTime := time.Now()

			mockMetricsRepo.EXPECT().GetServerLoadMetrics(gomock.Any(), gomock.Any()).Return(&domain.MetricResponse{
				Metrics: []domain.Metric{},
				Total:   0,
				HasMore: false,
			}, nil)

			metrics, err := analyticsService.GetServerLoadMetrics(ctx, startTime, endTime)

			Expect(err).To(BeNil())
			Expect(metrics).To(BeEmpty()) // Should return empty slice
		})

		It("should return empty list when no metrics exist", func() {
			startTime := time.Now().Add(-24 * time.Hour)
			endTime := time.Now()

			mockMetricsRepo.EXPECT().GetServerLoadMetrics(gomock.Any(), gomock.Any()).Return(&domain.MetricResponse{
				Metrics: []domain.Metric{},
				Total:   0,
				HasMore: false,
			}, nil)

			metrics, err := analyticsService.GetServerLoadMetrics(ctx, startTime, endTime)

			Expect(err).To(BeNil())
			Expect(metrics).To(BeEmpty()) // Should return empty slice
		})
	})
})
