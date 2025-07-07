package services_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/api/gateway/internal/services"
	"go.uber.org/zap"
)

var _ = Describe("ServerManagerProxy", func() {
	var (
		serverManagerProxy *services.ServerManagerProxy
		mockServerManager *httptest.Server
		logger           *zap.Logger
	)

	BeforeEach(func() {
		logger, _ = zap.NewDevelopment()
		mockServerManager = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/api/v1/server-manager/servers"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
		}))

		serverManagerProxy = services.NewServerManagerProxy(mockServerManager.URL, logger, mockServerManager.Client())
	})

	AfterEach(func() {
		mockServerManager.Close()
	})

	Describe("Proxy", func() {
		It("should proxy requests to the Server Manager service", func() {
			req := httptest.NewRequest("GET", "/api/v1/server-manager/servers", nil)
			rr := httptest.NewRecorder()

			serverManagerProxy.Proxy(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(Equal(`[]`))
		})
	})
})
