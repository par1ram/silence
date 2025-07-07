package services_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/domain"
	"github.com/par1ram/silence/rpc/dpi-bypass/internal/services"
	. "github.com/par1ram/silence/rpc/dpi-bypass/internal/services/mocks"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mock_bypass.go -package=services_test github.com/par1ram/silence/rpc/dpi-bypass/internal/ports BypassAdapter

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}

var _ = Describe("Services", func() {
	var bypassService *services.BypassService
	var healthService *services.HealthService
	var ctx context.Context
	var logger *zap.Logger
	var mockAdapter *MockBypassAdapter
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockAdapter = NewMockBypassAdapter(ctrl)
		logger = zap.NewNop()
		bypassService = services.NewBypassService(mockAdapter, logger)
		healthService = services.NewHealthService("test-service", "1.0.0")
		ctx = context.Background()
	})

	Describe("BypassService", func() {
		It("should create bypass and return bypass info", func() {
			request := &domain.CreateBypassRequest{
				Name:       "test-bypass",
				Method:     domain.BypassMethodObfs4,
				LocalPort:  1080,
				RemoteHost: "test-server.com",
				RemotePort: 443,
				Encryption: "aes-256-gcm",
			}

			bypass, err := bypassService.CreateBypass(ctx, request)

			Expect(err).To(BeNil())
			Expect(bypass).NotTo(BeNil())
			Expect(bypass.Name).To(Equal(request.Name))
			Expect(bypass.Method).To(Equal(request.Method))
			Expect(bypass.Status).To(Equal("inactive"))
			Expect(bypass.CreatedAt).NotTo(BeZero())
		})
	})

	Describe("GetBypass", func() {
		It("should return bypass", func() {
			request := &domain.CreateBypassRequest{
				Name:       "test-bypass",
				Method:     domain.BypassMethodObfs4,
				LocalPort:  1080,
				RemoteHost: "test-server.com",
				RemotePort: 443,
				Encryption: "aes-256-gcm",
			}
			createdBypass, _ := bypassService.CreateBypass(ctx, request)

			bypass, err := bypassService.GetBypass(ctx, createdBypass.ID)

			Expect(err).To(BeNil())
			Expect(bypass).NotTo(BeNil())
			Expect(bypass.ID).To(Equal(createdBypass.ID))
			Expect(bypass.Name).To(Equal(createdBypass.Name))
		})

		It("should return error for nonexistent bypass", func() {
			bypassID := "nonexistent-bypass-id"

			bypass, err := bypassService.GetBypass(ctx, bypassID)

			Expect(err).NotTo(BeNil())
			Expect(bypass).To(BeNil())
		})
	})

	Describe("ListBypasses", func() {
		It("should return list of bypasses", func() {
			request1 := &domain.CreateBypassRequest{
				Name:       "bypass-1",
				Method:     domain.BypassMethodObfs4,
				LocalPort:  1080,
				RemoteHost: "server1.com",
				RemotePort: 443,
				Encryption: "aes-256-gcm",
			}
			request2 := &domain.CreateBypassRequest{
				Name:       "bypass-2",
				Method:     domain.BypassMethodShadowsocks,
				LocalPort:  1081,
				RemoteHost: "server2.com",
				RemotePort: 8388,
				Encryption: "aes-256-gcm",
			}
			_, err1 := bypassService.CreateBypass(ctx, request1)
			Expect(err1).To(BeNil())
			_, err2 := bypassService.CreateBypass(ctx, request2)
			Expect(err2).To(BeNil())

			bypasses, err := bypassService.ListBypasses(ctx)

			Expect(err).To(BeNil())
			Expect(bypasses).NotTo(BeNil())
			Expect(len(bypasses)).To(BeNumerically(">=", 2))
		})

		It("should return empty list if no bypasses", func() {
			bypasses, err := bypassService.ListBypasses(ctx)

			Expect(err).To(BeNil())
			Expect(bypasses).NotTo(BeNil())
			Expect(len(bypasses)).To(Equal(0))
		})
	})

	Describe("CreateBypass with Shadowsocks", func() {
		It("should create bypass and return bypass info", func() {
			request := &domain.CreateBypassRequest{
				Name:       "test-shadowsocks-bypass",
				Method:     domain.BypassMethodShadowsocks,
				LocalPort:  1081,
				RemoteHost: "test-server.com",
				RemotePort: 8388,
				Encryption: "aes-256-gcm",
			}

			bypass, err := bypassService.CreateBypass(ctx, request)

			Expect(err).To(BeNil())
			Expect(bypass).NotTo(BeNil())
			Expect(bypass.Name).To(Equal(request.Name))
			Expect(bypass.Method).To(Equal(request.Method))
			Expect(bypass.Status).To(Equal("inactive"))
			Expect(bypass.CreatedAt).NotTo(BeZero())
		})
	})

	Describe("CreateBypass with V2Ray", func() {
		It("should create bypass and return bypass info", func() {
			request := &domain.CreateBypassRequest{
				Name:       "test-v2ray-bypass",
				Method:     domain.BypassMethodV2Ray,
				LocalPort:  1082,
				RemoteHost: "test-server.com",
				RemotePort: 10086,
				Encryption: "none",
			}

			bypass, err := bypassService.CreateBypass(ctx, request)

			Expect(err).To(BeNil())
			Expect(bypass).NotTo(BeNil())
			Expect(bypass.Name).To(Equal(request.Name))
			Expect(bypass.Method).To(Equal(request.Method))
			Expect(bypass.Status).To(Equal("inactive"))
			Expect(bypass.CreatedAt).NotTo(BeZero())
		})
	})

	Describe("HealthService", func() {
		It("should return health status", func() {
			status := healthService.GetHealth()

			Expect(status).NotTo(BeNil())
			Expect(status.Service).To(Equal("test-service"))
			Expect(status.Version).To(Equal("1.0.0"))
			Expect(status.Status).To(Equal("ok"))
			Expect(status.Timestamp).NotTo(BeZero())
		})
	})
})
