package services_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	services "github.com/par1ram/silence/rpc/vpn-core/internal/services"
	. "github.com/par1ram/silence/rpc/vpn-core/internal/services/mocks"
	"go.uber.org/zap"
)

//go:generate mockgen -destination=mock_tunnel.go -package=services_test github.com/par1ram/silence/rpc/vpn-core/internal/ports KeyGenerator,WireGuardManager

var _ = Describe("TunnelService", func() {
	var tunnelService ports.TunnelManager
	var ctx context.Context
	var logger *zap.Logger
	var mockKeyGen *MockKeyGenerator
	var mockWG *MockWireGuardManager
	var ctrl *gomock.Controller

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockKeyGen = NewMockKeyGenerator(ctrl)
		mockWG = NewMockWireGuardManager(ctrl)
		logger = zap.NewNop()
		tunnelService = services.NewTunnelService(mockKeyGen, mockWG, logger)
		ctx = context.Background()
	})

	Describe("CreateTunnel", func() {
		It("should create tunnel and return tunnel info", func() {
			request := &domain.CreateTunnelRequest{
				Name:         "test-tunnel",
				ListenPort:   51820,
				MTU:          1420,
				AutoRecovery: true,
			}
			mockKeyGen.EXPECT().GenerateKeyPair().Return("pub", "priv", nil)
			tunnel, err := tunnelService.CreateTunnel(ctx, request)
			Expect(err).To(BeNil())
			Expect(tunnel).NotTo(BeNil())
			Expect(tunnel.Name).To(Equal(request.Name))
			Expect(tunnel.Status).To(Equal(domain.TunnelStatusInactive))
			Expect(tunnel.CreatedAt).NotTo(BeZero())
			Expect(tunnel.ID).NotTo(BeEmpty())
		})

		It("should generate unique IDs for multiple tunnels", func() {
			mockKeyGen.EXPECT().GenerateKeyPair().Return("pub1", "priv1", nil).Times(2)

			request1 := &domain.CreateTunnelRequest{Name: "tunnel-unique-1"}
			tunnel1, err1 := tunnelService.CreateTunnel(ctx, request1)
			Expect(err1).To(BeNil())
			Expect(tunnel1).NotTo(BeNil())
			Expect(tunnel1.ID).NotTo(BeEmpty())

			request2 := &domain.CreateTunnelRequest{Name: "tunnel-unique-2"}
			tunnel2, err2 := tunnelService.CreateTunnel(ctx, request2)
			Expect(err2).To(BeNil())
			Expect(tunnel2).NotTo(BeNil())
			Expect(tunnel2.ID).NotTo(BeEmpty())

			Expect(tunnel1.ID).NotTo(Equal(tunnel2.ID))
		})
	})

	Describe("GetTunnel", func() {
		It("should return tunnel", func() {
			request := &domain.CreateTunnelRequest{Name: "test-tunnel"}
			mockKeyGen.EXPECT().GenerateKeyPair().Return("pub", "priv", nil)
			createdTunnel, _ := tunnelService.CreateTunnel(ctx, request)
			tunnel, err := tunnelService.GetTunnel(ctx, createdTunnel.ID)
			Expect(err).To(BeNil())
			Expect(tunnel).NotTo(BeNil())
			Expect(tunnel.ID).To(Equal(createdTunnel.ID))
			Expect(tunnel.Name).To(Equal(createdTunnel.Name))
		})
		It("should return error for nonexistent tunnel", func() {
			tunnelID := "nonexistent-tunnel-id"
			tunnel, err := tunnelService.GetTunnel(ctx, tunnelID)
			Expect(err).NotTo(BeNil())
			Expect(tunnel).To(BeNil())
		})
	})

	Describe("ListTunnels", func() {
		It("should return list of tunnels", func() {
			request1 := &domain.CreateTunnelRequest{Name: "tunnel-1"}
			request2 := &domain.CreateTunnelRequest{Name: "tunnel-2"}
			mockKeyGen.EXPECT().GenerateKeyPair().Return("pub", "priv", nil).AnyTimes()
			_, err1 := tunnelService.CreateTunnel(ctx, request1)
			Expect(err1).To(BeNil())
			_, err2 := tunnelService.CreateTunnel(ctx, request2)
			Expect(err2).To(BeNil())
			tunnels, err := tunnelService.ListTunnels(ctx)
			Expect(err).To(BeNil())
			Expect(tunnels).NotTo(BeNil())
			Expect(len(tunnels)).To(BeNumerically(">=", 2))
		})
		It("should return empty list if no tunnels", func() {
			tunnels, err := tunnelService.ListTunnels(ctx)
			Expect(err).To(BeNil())
			Expect(tunnels).NotTo(BeNil())
			Expect(len(tunnels)).To(Equal(0))
		})
	})
})
