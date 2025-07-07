package services_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/api/gateway/internal/services"
	"go.uber.org/zap"
)

var _ = Describe("AuthProxy", func() {
	var (
		authProxy     *services.AuthProxy
		mockAuthServer *httptest.Server
		logger        *zap.Logger
	)

	BeforeEach(func() {
		logger, _ = zap.NewDevelopment()
		mockAuthServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Expect(r.Header.Get("X-Internal-Token")).To(Equal("test-token"))
			Expect(r.URL.Path).To(Equal("/login"))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}))

		authProxy = services.NewAuthProxy(mockAuthServer.URL, "test-token", logger, mockAuthServer.Client())
	})

	AfterEach(func() {
		mockAuthServer.Close()
	})

	Describe("Proxy", func() {
		It("should proxy requests to the auth service", func() {
			req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
			rr := httptest.NewRecorder()

			authProxy.Proxy(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Body.String()).To(Equal("ok"))
		})
	})
})
