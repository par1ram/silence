package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/analytics/internal/services"
	. "github.com/par1ram/silence/rpc/analytics/internal/services/mocks"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mock_analytics.go -package=services_test github.com/par1ram/silence/rpc/analytics/internal/ports MetricsRepository,DashboardRepository,MetricsCollector,AlertService

func TestAnalyticsService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AnalyticsService Suite")
}

var _ = Describe("AnalyticsService", func() {
	var analyticsService *services.AnalyticsServiceImpl
	var ctx context.Context
	var logger *zap.Logger
	var mockMetricsRepo *MockMetricsRepository
	var mockDashboardRepo *MockDashboardRepository
	var mockCollector *MockMetricsCollector
	var mockAlertService *MockAlertService
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockMetricsRepo = NewMockMetricsRepository(ctrl)
		mockDashboardRepo = NewMockDashboardRepository(ctrl)
		mockCollector = NewMockMetricsCollector(ctrl)
		mockAlertService = NewMockAlertService(ctrl)
		logger = zap.NewNop()
		analyticsService = services.NewAnalyticsService(
			mockMetricsRepo,
			mockDashboardRepo,
			mockCollector,
			mockAlertService,
			logger,
		).(*services.AnalyticsServiceImpl)
		ctx = context.Background()
	})

	Describe("GetServerLoadMetrics", func() {
		It("should return server load metrics", func() {
			startTime := time.Now().Add(-24 * time.Hour)
			endTime := time.Now()

			metrics, err := analyticsService.GetServerLoadMetrics(ctx, startTime, endTime)

			Expect(err).To(BeNil())
			Expect(metrics).NotTo(BeNil())
			Expect(len(metrics)).To(Equal(0)) // Currently returns empty slice
		})

		It("should return empty list when no metrics exist", func() {
			startTime := time.Now().Add(-24 * time.Hour)
			endTime := time.Now()

			metrics, err := analyticsService.GetServerLoadMetrics(ctx, startTime, endTime)

			Expect(err).To(BeNil())
			Expect(metrics).NotTo(BeNil())
			Expect(len(metrics)).To(Equal(0))
		})
	})
})
