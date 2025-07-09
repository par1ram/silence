package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	svc "github.com/par1ram/silence/rpc/vpn-core/internal/services"
	mocks "github.com/par1ram/silence/rpc/vpn-core/internal/services/mocks"
	"go.uber.org/zap"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}

var _ = Describe("Services", func() {
	var (
		tunnelService *svc.TunnelService
		mockKeyGen    *mocks.MockKeyGenerator
		mockWgManager *mocks.MockWireGuardManager
		logger        *zap.Logger
		ctrl          *gomock.Controller
		ctx           context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockKeyGen = mocks.NewMockKeyGenerator(ctrl)
		mockWgManager = mocks.NewMockWireGuardManager(ctrl)
		logger = zap.NewNop()
		tunnelService = svc.NewTunnelService(mockKeyGen, mockWgManager, logger).(*svc.TunnelService)
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("TunnelService", func() {
		Describe("CreateTunnel", func() {
			It("should create a new tunnel", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				req := &domain.CreateTunnelRequest{
					Name:         "test-tunnel",
					ListenPort:   51820,
					MTU:          1420,
					AutoRecovery: true,
				}
				tunnel, err := tunnelService.CreateTunnel(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(tunnel).NotTo(BeNil())
				Expect(tunnel.Name).To(Equal("test-tunnel"))
				Expect(tunnel.Status).To(Equal(domain.TunnelStatusInactive))
				Expect(tunnel.PublicKey).To(Equal("pubkey"))
				Expect(tunnel.PrivateKey).To(Equal("privkey"))
			})

			It("should return an error if key generation fails", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("", "", errors.New("key gen error"))
				req := &domain.CreateTunnelRequest{
					Name: "test-tunnel",
				}
				tunnel, err := tunnelService.CreateTunnel(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(tunnel).To(BeNil())
			})
		})

		Describe("GetTunnel", func() {
			It("should return the tunnel if it exists", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, _ := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "test-tunnel"})
				tunnel, err := tunnelService.GetTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(tunnel).To(Equal(createdTunnel))
			})

			It("should return an error if the tunnel does not exist", func() {
				tunnel, err := tunnelService.GetTunnel(ctx, "non-existent-id")
				Expect(err).To(HaveOccurred())
				Expect(tunnel).To(BeNil())
			})
		})

		Describe("ListTunnels", func() {
			It("should return a list of all tunnels", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey1", "privkey1", nil).Times(2)
				_, _ = tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "tunnel1"})
				_, _ = tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "tunnel2"})
				tunnels, err := tunnelService.ListTunnels(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(tunnels)).To(Equal(2))
			})

			It("should return an empty list if no tunnels exist", func() {
				tunnels, err := tunnelService.ListTunnels(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(len(tunnels)).To(Equal(0))
			})
		})

		Describe("DeleteTunnel", func() {
			It("should delete the tunnel if it exists", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, _ := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "test-tunnel"})
				err := tunnelService.DeleteTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				_, err = tunnelService.GetTunnel(ctx, createdTunnel.ID)
				Expect(err).To(HaveOccurred())
			})

			It("should return an error if the tunnel does not exist", func() {
				err := tunnelService.DeleteTunnel(ctx, "non-existent-id")
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("StartTunnel", func() {
			It("should start the tunnel", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, _ := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "test-tunnel", ListenPort: 51820, MTU: 1420})
				mockWgManager.EXPECT().CreateInterface(createdTunnel.Interface, createdTunnel.PrivateKey, createdTunnel.ListenPort, createdTunnel.MTU).Return(nil)
				err := tunnelService.StartTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusActive))
			})

			It("should return an error if the tunnel does not exist", func() {
				err := tunnelService.StartTunnel(ctx, "non-existent-id")
				Expect(err).To(HaveOccurred())
			})

			It("should return an error if CreateInterface fails", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, _ := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "test-tunnel", ListenPort: 51820, MTU: 1420})
				mockWgManager.EXPECT().CreateInterface(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("interface error"))
				err := tunnelService.StartTunnel(ctx, createdTunnel.ID)
				Expect(err).To(HaveOccurred())
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusError))
			})
		})

		Describe("StopTunnel", func() {
			It("should stop the tunnel", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, _ := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "test-tunnel", ListenPort: 51820, MTU: 1420})
				mockWgManager.EXPECT().CreateInterface(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				_ = tunnelService.StartTunnel(ctx, createdTunnel.ID)
				mockWgManager.EXPECT().DeleteInterface(createdTunnel.Interface).Return(nil)
				err := tunnelService.StopTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusInactive))
			})

			It("should return an error if the tunnel does not exist", func() {
				err := tunnelService.StopTunnel(ctx, "non-existent-id")
				Expect(err).To(HaveOccurred())
			})

			It("should return an error if DeleteInterface fails", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, _ := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{Name: "test-tunnel", ListenPort: 51820, MTU: 1420})
				mockWgManager.EXPECT().CreateInterface(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				_ = tunnelService.StartTunnel(ctx, createdTunnel.ID)
				mockWgManager.EXPECT().DeleteInterface(gomock.Any()).Return(errors.New("delete error"))
				err := tunnelService.StopTunnel(ctx, createdTunnel.ID)
				Expect(err).To(HaveOccurred())
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusError))
			})
		})
	})

	Describe("HealthService", func() {
		var healthService *svc.HealthService

		BeforeEach(func() {
			healthService = svc.NewHealthService("test-service", "1.0.0").(*svc.HealthService)
		})

		Describe("GetHealth", func() {
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
})
