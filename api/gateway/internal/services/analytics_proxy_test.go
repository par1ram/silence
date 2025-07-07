package services_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/api/gateway/internal/services"
	"go.uber.org/zap"
)

var _ = Describe("AnalyticsProxy", func() {
	var (
		analyticsProxy *services.AnalyticsProxy
		mockAnalyticsServer *httptest.Server
		logger         *zap.Logger
	)

	BeforeEach(func() {
		logger, _ = zap.NewDevelopment()
		mockAnalyticsServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/metrics"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}))

		analyticsProxy = services.NewAnalyticsProxy(mockAnalyticsServer.URL, logger, mockAnalyticsServer.Client())
	})

	AfterEach(func() {
		mockAnalyticsServer.Close()
	})

	Describe("Proxy", func() {
		It("should proxy requests to the Analytics service", func() {
			req := httptest.NewRequest("GET", "/api/v1/analytics/metrics", nil)
			rr := httptest.NewRecorder()

			analyticsProxy.Proxy(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(Equal(`{"status":"ok"}`))
		})
	})
})
