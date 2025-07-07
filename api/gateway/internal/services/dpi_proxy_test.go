package services_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/api/gateway/internal/services"
	"go.uber.org/zap"
)

var _ = Describe("DPIProxy", func() {
	var (
		dpiProxy       *services.DPIProxy
		mockDPIServer *httptest.Server
		logger        *zap.Logger
	)

	BeforeEach(func() {
		logger, _ = zap.NewDevelopment()
		mockDPIServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/bypass"))
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":"456"}`))
		}))

		dpiProxy = services.NewDPIProxy(mockDPIServer.URL, logger, mockDPIServer.Client())
	})

	AfterEach(func() {
		mockDPIServer.Close()
	})

	Describe("Proxy", func() {
		It("should proxy requests to the DPI Bypass service", func() {
			req := httptest.NewRequest("POST", "/api/v1/dpi-bypass/bypass", nil)
			rr := httptest.NewRecorder()

			dpiProxy.Proxy(rr, req)

			Expect(rr.Code).To(Equal(http.StatusCreated))
			Expect(rr.Body.String()).To(Equal(`{"id":"456"}`))
		})
	})
})
