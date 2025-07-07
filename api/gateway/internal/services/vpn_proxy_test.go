package services_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/api/gateway/internal/services"
	"go.uber.org/zap"
)

var _ = Describe("VPNProxy", func() {
	var (
		vpnProxy       *services.VPNProxy
		mockVPNServer *httptest.Server
		logger        *zap.Logger
	)

	BeforeEach(func() {
		logger, _ = zap.NewDevelopment()
		mockVPNServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.URL.Path).To(Equal("/tunnels"))
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"id":"123"}`))
		}))

		vpnProxy = services.NewVPNProxy(mockVPNServer.URL, logger, mockVPNServer.Client())
	})

	AfterEach(func() {
		mockVPNServer.Close()
	})

	Describe("Proxy", func() {
		It("should proxy requests to the VPN Core service", func() {
			req := httptest.NewRequest("POST", "/api/v1/vpn/tunnels", nil)
			rr := httptest.NewRecorder()

			vpnProxy.Proxy(rr, req)

			Expect(rr.Code).To(Equal(http.StatusCreated))
			Expect(rr.Body.String()).To(Equal(`{"id":"123"}`))
		})
	})
})
