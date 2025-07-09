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

//go:generate mockgen -destination=mocks/mock_bypass_adapter.go -package=mocks github.com/par1ram/silence/rpc/dpi-bypass/internal/ports BypassAdapter

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
		bypassService = services.NewBypassService(mockAdapter, logger).(*services.BypassService)
		healthService = services.NewHealthService("test-service", "1.0.0")
		ctx = context.Background()
	})

	Describe("BypassService", func() {
		It("should create bypass and return bypass info", func() {
			request := &domain.CreateBypassConfigRequest{
				Name:        "test-bypass",
				Description: "test bypass configuration",
				Type:        domain.BypassTypeTunnelObfuscation,
				Method:      domain.BypassMethodObfs4,
				Parameters: map[string]string{
					"local_port":  "1080",
					"remote_host": "test-server.com",
					"remote_port": "443",
					"encryption":  "aes-256-gcm",
				},
			}

			config, err := bypassService.CreateBypassConfig(ctx, request)

			Expect(err).To(BeNil())
			Expect(config).NotTo(BeNil())
			Expect(config.Name).To(Equal(request.Name))
			Expect(config.Method).To(Equal(request.Method))
			Expect(config.Status).To(Equal(domain.BypassStatusInactive))
			Expect(config.CreatedAt).NotTo(BeZero())
		})
	})

	Describe("GetBypass", func() {
		It("should return bypass", func() {
			request := &domain.CreateBypassConfigRequest{
				Name:        "test-bypass",
				Description: "test bypass configuration",
				Type:        domain.BypassTypeTunnelObfuscation,
				Method:      domain.BypassMethodObfs4,
				Parameters: map[string]string{
					"local_port":  "1080",
					"remote_host": "test-server.com",
					"remote_port": "443",
					"encryption":  "aes-256-gcm",
				},
			}
			createdConfig, _ := bypassService.CreateBypassConfig(ctx, request)

			config, err := bypassService.GetBypassConfig(ctx, createdConfig.ID)

			Expect(err).To(BeNil())
			Expect(config).NotTo(BeNil())
			Expect(config.ID).To(Equal(createdConfig.ID))
		})

		It("should return error for non-existent bypass", func() {
			config, err := bypassService.GetBypassConfig(ctx, "non-existent-id")

			Expect(err).NotTo(BeNil())
			Expect(config).To(BeNil())
		})
	})

	Describe("ListBypasses", func() {
		It("should return list of bypasses", func() {
			request1 := &domain.CreateBypassConfigRequest{
				Name:        "test-bypass-1",
				Description: "test bypass configuration 1",
				Type:        domain.BypassTypeTunnelObfuscation,
				Method:      domain.BypassMethodObfs4,
				Parameters: map[string]string{
					"local_port":  "1080",
					"remote_host": "test-server.com",
					"remote_port": "443",
					"encryption":  "aes-256-gcm",
				},
			}
			request2 := &domain.CreateBypassConfigRequest{
				Name:        "test-bypass-2",
				Description: "test bypass configuration 2",
				Type:        domain.BypassTypeTunnelObfuscation,
				Method:      domain.BypassMethodShadowsocks,
				Parameters: map[string]string{
					"local_port":  "1081",
					"remote_host": "test-server2.com",
					"remote_port": "443",
					"encryption":  "aes-256-gcm",
				},
			}
			_, err1 := bypassService.CreateBypassConfig(ctx, request1)
			Expect(err1).To(BeNil())
			_, err2 := bypassService.CreateBypassConfig(ctx, request2)
			Expect(err2).To(BeNil())

			configs, total, err := bypassService.ListBypassConfigs(ctx, nil)

			Expect(err).To(BeNil())
			Expect(configs).NotTo(BeNil())
			Expect(len(configs)).To(Equal(2))
			Expect(total).To(Equal(2))
		})

		It("should return empty list when no bypasses exist", func() {
			configs, total, err := bypassService.ListBypassConfigs(ctx, nil)

			Expect(err).To(BeNil())
			Expect(configs).NotTo(BeNil())
			Expect(len(configs)).To(Equal(0))
			Expect(total).To(Equal(0))
		})
	})

	Describe("CreateBypass with Shadowsocks", func() {
		It("should create bypass and return bypass info", func() {
			request := &domain.CreateBypassConfigRequest{
				Name:        "test-shadowsocks-bypass",
				Description: "test shadowsocks bypass configuration",
				Type:        domain.BypassTypeTunnelObfuscation,
				Method:      domain.BypassMethodShadowsocks,
				Parameters: map[string]string{
					"local_port":  "1081",
					"remote_host": "test-server.com",
					"remote_port": "8388",
					"encryption":  "aes-256-gcm",
				},
			}

			config, err := bypassService.CreateBypassConfig(ctx, request)

			Expect(err).To(BeNil())
			Expect(config).NotTo(BeNil())
			Expect(config.Name).To(Equal(request.Name))
			Expect(config.Method).To(Equal(request.Method))
			Expect(config.Status).To(Equal(domain.BypassStatusInactive))
			Expect(config.CreatedAt).NotTo(BeZero())
		})
	})

	Describe("CreateBypass with V2Ray", func() {
		It("should create bypass and return bypass info", func() {
			request := &domain.CreateBypassConfigRequest{
				Name:        "test-v2ray-bypass",
				Description: "test v2ray bypass configuration",
				Type:        domain.BypassTypeTunnelObfuscation,
				Method:      domain.BypassMethodV2Ray,
				Parameters: map[string]string{
					"local_port":  "1082",
					"remote_host": "test-server.com",
					"remote_port": "10808",
					"encryption":  "vmess",
				},
			}

			config, err := bypassService.CreateBypassConfig(ctx, request)

			Expect(err).To(BeNil())
			Expect(config).NotTo(BeNil())
			Expect(config.Name).To(Equal(request.Name))
			Expect(config.Method).To(Equal(request.Method))
			Expect(config.Status).To(Equal(domain.BypassStatusInactive))
			Expect(config.CreatedAt).NotTo(BeZero())
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
