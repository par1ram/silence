package services_test

import (
	"context"
	"errors"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	svc "github.com/par1ram/silence/rpc/vpn-core/internal/services"
	mocks "github.com/par1ram/silence/rpc/vpn-core/internal/services/mocks"
	"go.uber.org/zap"
)

var _ = Describe("Tunnel Monitor", func() {
	var (
		tunnelService *svc.TunnelService
		mockKeyGen    *mocks.MockKeyGenerator
		mockWgManager *mocks.MockWireGuardManager
		ctrl          *gomock.Controller
		ctx           context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockKeyGen = mocks.NewMockKeyGenerator(ctrl)
		mockWgManager = mocks.NewMockWireGuardManager(ctrl)
		logger := zap.NewNop()
		tunnelService = svc.NewTunnelService(mockKeyGen, mockWgManager, logger).(*svc.TunnelService)
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetTunnelStats", func() {
		Context("when tunnel exists", func() {
			It("should return tunnel stats", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					ListenPort: 51820,
					MTU:        1420,
				})
				Expect(err).NotTo(HaveOccurred())

				// Start the tunnel to set the start time for uptime calculation
				mockWgManager.EXPECT().CreateInterface(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				err = tunnelService.StartTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())

				mockWgManager.EXPECT().GetInterfaceStats(createdTunnel.Interface).Return(&ports.InterfaceStats{
					BytesRx: 100,
					BytesTx: 200,
				}, nil)

				stats, err := tunnelService.GetTunnelStats(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(stats).NotTo(BeNil())
				Expect(stats.TunnelID).To(Equal(createdTunnel.ID))
				Expect(stats.BytesRx).To(Equal(int64(100)))
				Expect(stats.BytesTx).To(Equal(int64(200)))
				Expect(stats.PeersCount).To(Equal(0))
				Expect(stats.ActivePeers).To(Equal(0))
				Expect(stats.Uptime).To(BeNumerically(">", 0))
				Expect(stats.ErrorCount).To(Equal(0))
				Expect(stats.RecoveryCount).To(Equal(0))
			})

			It("should handle GetInterfaceStats error gracefully", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					ListenPort: 51820,
					MTU:        1420,
				})
				Expect(err).NotTo(HaveOccurred())

				mockWgManager.EXPECT().GetInterfaceStats(createdTunnel.Interface).Return(nil, errors.New("stats error"))

				stats, err := tunnelService.GetTunnelStats(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(stats.BytesRx).To(Equal(int64(0)))
				Expect(stats.BytesTx).To(Equal(int64(0)))
			})
		})

		Context("when tunnel does not exist", func() {
			It("should return an error", func() {
				stats, err := tunnelService.GetTunnelStats(ctx, "non-existent-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tunnel not found"))
				Expect(stats).To(BeNil())
			})
		})
	})

	Describe("HealthCheck", func() {
		Context("when tunnel exists", func() {
			It("should return healthy status for active tunnel", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					ListenPort: 51820,
					MTU:        1420,
				})
				Expect(err).NotTo(HaveOccurred())

				mockWgManager.EXPECT().CreateInterface(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
				err = tunnelService.StartTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())

				req := &domain.HealthCheckRequest{TunnelID: createdTunnel.ID}
				resp, err := tunnelService.HealthCheck(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.TunnelID).To(Equal(createdTunnel.ID))
				Expect(resp.Status).To(Equal("healthy"))
				Expect(resp.PeersHealth).To(BeEmpty())
				Expect(resp.Uptime).To(BeNumerically(">", 0))
				Expect(resp.ErrorCount).To(Equal(0))
			})

			It("should return unhealthy status for inactive tunnel", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					ListenPort: 51820,
					MTU:        1420,
				})
				Expect(err).NotTo(HaveOccurred())

				req := &domain.HealthCheckRequest{TunnelID: createdTunnel.ID}
				resp, err := tunnelService.HealthCheck(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.Status).To(Equal("unhealthy"))
			})

			It("should include peer health information", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					ListenPort: 51820,
					MTU:        1420,
				})
				Expect(err).NotTo(HaveOccurred())

				// Manually add a peer to the tunnel's internal map for testing purposes
				tunnelService.GetPeers()[createdTunnel.ID] = []*domain.Peer{
					{
						ID:                "peer1",
						Status:            domain.PeerStatusActive,
						LastHandshake:     time.Now().Add(-5 * time.Second),
						Latency:           10 * time.Millisecond,
						PacketLoss:        0.1,
						ConnectionQuality: 0.9,
					},
				}

				req := &domain.HealthCheckRequest{TunnelID: createdTunnel.ID}
				resp, err := tunnelService.HealthCheck(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.PeersHealth).To(HaveLen(1))
				Expect(resp.PeersHealth[0].PeerID).To(Equal("peer1"))
				Expect(resp.PeersHealth[0].Status).To(Equal(domain.PeerStatusActive))
			})
		})

		Context("when tunnel does not exist", func() {
			It("should return an error", func() {
				req := &domain.HealthCheckRequest{TunnelID: "non-existent-id"}
				resp, err := tunnelService.HealthCheck(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tunnel not found"))
				Expect(resp).To(BeNil())
			})
		})
	})

	Describe("EnableAutoRecovery", func() {
		Context("when tunnel exists", func() {
			It("should enable auto recovery", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					AutoRecovery: false,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.AutoRecovery).To(BeFalse())

				err = tunnelService.EnableAutoRecovery(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.AutoRecovery).To(BeTrue())
			})
		})

		Context("when tunnel does not exist", func() {
			It("should return an error", func() {
				err := tunnelService.EnableAutoRecovery(ctx, "non-existent-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tunnel not found"))
			})
		})
	})

	Describe("DisableAutoRecovery", func() {
		Context("when tunnel exists", func() {
			It("should disable auto recovery", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					AutoRecovery: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.AutoRecovery).To(BeTrue())

				err = tunnelService.DisableAutoRecovery(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.AutoRecovery).To(BeFalse())
			})
		})

		Context("when tunnel does not exist", func() {
			It("should return an error", func() {
				err := tunnelService.DisableAutoRecovery(ctx, "non-existent-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tunnel not found"))
			})
		})
	})

	Describe("RecoverTunnel", func() {
		Context("when tunnel exists", func() {
			It("should recover an inactive tunnel", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					ListenPort: 51820,
					MTU:        1420,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusInactive))

				mockWgManager.EXPECT().CreateInterface(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

				err = tunnelService.RecoverTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusActive))
				// Verify recovery count by checking the internal map directly
				Expect(tunnelService.GetRecoveryCounts()[createdTunnel.ID]).To(Equal(1))
			})

			It("should recover an active tunnel by deleting and recreating interface", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					ListenPort: 51820,
					MTU:        1420,
				})
				Expect(err).NotTo(HaveOccurred())

				mockWgManager.EXPECT().CreateInterface(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
				mockWgManager.EXPECT().DeleteInterface(gomock.Any()).Return(nil)

				err = tunnelService.StartTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusActive))

				err = tunnelService.RecoverTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusActive))
				// Verify recovery count by checking the internal map directly
				Expect(tunnelService.GetRecoveryCounts()[createdTunnel.ID]).To(Equal(1))
			})

			It("should handle DeleteInterface error during recovery", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					ListenPort: 51820,
					MTU:        1420,
				})
				Expect(err).NotTo(HaveOccurred())

				mockWgManager.EXPECT().CreateInterface(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(2)
				mockWgManager.EXPECT().DeleteInterface(gomock.Any()).Return(errors.New("delete error"))

				err = tunnelService.StartTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusActive))

				err = tunnelService.RecoverTunnel(ctx, createdTunnel.ID)
				Expect(err).NotTo(HaveOccurred()) // Should not return error, just log
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusActive))
			})

			It("should return an error if CreateInterface fails during recovery", func() {
				mockKeyGen.EXPECT().GenerateKeyPair().Return("pubkey", "privkey", nil)
				createdTunnel, err := tunnelService.CreateTunnel(ctx, &domain.CreateTunnelRequest{
					Name:       "test-tunnel",
					ListenPort: 51820,
					MTU:        1420,
				})
				Expect(err).NotTo(HaveOccurred())

				mockWgManager.EXPECT().CreateInterface(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("create error"))

				err = tunnelService.RecoverTunnel(ctx, createdTunnel.ID)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to recreate wireguard interface"))
				Expect(createdTunnel.Status).To(Equal(domain.TunnelStatusError))
				// Verify error count by checking the internal map directly
				Expect(tunnelService.GetErrorCounts()[createdTunnel.ID]).To(Equal(1))
			})
		})

		Context("when tunnel does not exist", func() {
			It("should return an error", func() {
				err := tunnelService.RecoverTunnel(ctx, "non-existent-id")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tunnel not found"))
			})
		})
	})
})
