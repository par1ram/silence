package services_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	svc "github.com/par1ram/silence/rpc/vpn-core/internal/services"
	"go.uber.org/zap"
)

var _ = Describe("Peer Monitor", func() {
	var (
		peerService *svc.PeerService
		ctx         context.Context
		logger      *zap.Logger
	)

	BeforeEach(func() {
		logger = zap.NewNop()
		peerService = svc.NewPeerService(logger).(*svc.PeerService)
		ctx = context.Background()
	})

	Describe("UpdatePeerStats", func() {
		Context("when peer exists", func() {
			It("should update peer statistics", func() {
				// Add a peer first
				addReq := &domain.AddPeerRequest{
					TunnelID:  "test-tunnel",
					PublicKey: "pubkey1",
				}
				createdPeer, err := peerService.AddPeer(ctx, addReq)
				Expect(err).NotTo(HaveOccurred())
				Expect(createdPeer).NotTo(BeNil())

				stats := &ports.PeerStats{
					TransferRx:    100,
					TransferTx:    200,
					LastHandshake: time.Now().Unix(),
				}

				err = peerService.UpdatePeerStats(ctx, addReq.TunnelID, createdPeer.ID, stats)
				Expect(err).NotTo(HaveOccurred())

				updatedPeer, err := peerService.GetPeer(ctx, addReq.TunnelID, createdPeer.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedPeer.TransferRx).To(Equal(stats.TransferRx))
				Expect(updatedPeer.TransferTx).To(Equal(stats.TransferTx))
				Expect(updatedPeer.LastHandshake.Unix()).To(BeNumerically("~", stats.LastHandshake, 1))
				Expect(updatedPeer.LastSeen).NotTo(BeZero())
				Expect(updatedPeer.UpdatedAt).NotTo(BeZero())
			})

			It("should update peer status based on last handshake", func() {
				addReq := &domain.AddPeerRequest{
					TunnelID:  "test-tunnel-status",
					PublicKey: "pubkey-status",
				}
				createdPeer, err := peerService.AddPeer(ctx, addReq)
				Expect(err).NotTo(HaveOccurred())

				// Active status
				statsActive := &ports.PeerStats{
					LastHandshake: time.Now().Unix() - 30, // 30 seconds ago
				}
				err = peerService.UpdatePeerStats(ctx, addReq.TunnelID, createdPeer.ID, statsActive)
				Expect(err).NotTo(HaveOccurred())
				peer, _ := peerService.GetPeer(ctx, addReq.TunnelID, createdPeer.ID)
				Expect(peer.Status).To(Equal(domain.PeerStatusActive))

				// Inactive status
				statsInactive := &ports.PeerStats{
					LastHandshake: time.Now().Unix() - int64(5*time.Minute.Seconds()) - 30, // 5 min 30 sec ago
				}
				err = peerService.UpdatePeerStats(ctx, addReq.TunnelID, createdPeer.ID, statsInactive)
				Expect(err).NotTo(HaveOccurred())
				peer, _ = peerService.GetPeer(ctx, addReq.TunnelID, createdPeer.ID)
				Expect(peer.Status).To(Equal(domain.PeerStatusInactive))

				// Offline status
				statsOffline := &ports.PeerStats{
					LastHandshake: time.Now().Unix() - int64(15*time.Minute.Seconds()), // 15 min ago
				}
				err = peerService.UpdatePeerStats(ctx, addReq.TunnelID, createdPeer.ID, statsOffline)
				Expect(err).NotTo(HaveOccurred())
				peer, _ = peerService.GetPeer(ctx, addReq.TunnelID, createdPeer.ID)
				Expect(peer.Status).To(Equal(domain.PeerStatusOffline))
			})
		})

		Context("when peer does not exist", func() {
			It("should return an error", func() {
				stats := &ports.PeerStats{}
				err := peerService.UpdatePeerStats(ctx, "non-existent-tunnel", "non-existent-peer", stats)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tunnel not found"))
			})
		})
	})

	Describe("GetPeerHealth", func() {
		Context("when peer exists", func() {
			It("should return peer health information", func() {
				addReq := &domain.AddPeerRequest{
					TunnelID:  "health-tunnel",
					PublicKey: "pubkey-health",
				}
				createdPeer, err := peerService.AddPeer(ctx, addReq)
				Expect(err).NotTo(HaveOccurred())

				// Manually set some peer properties for health calculation
				peerService.GetPeersMap()[addReq.TunnelID][createdPeer.ID].LastHandshake = time.Now().Add(-1 * time.Minute)
				peerService.GetPeersMap()[addReq.TunnelID][createdPeer.ID].Latency = 50 * time.Millisecond
				peerService.GetPeersMap()[addReq.TunnelID][createdPeer.ID].PacketLoss = 0.05
				peerService.GetPeersMap()[addReq.TunnelID][createdPeer.ID].Status = domain.PeerStatusActive

				health, err := peerService.GetPeerHealth(ctx, addReq.TunnelID, createdPeer.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(health).NotTo(BeNil())
				Expect(health.PeerID).To(Equal(createdPeer.ID))
				Expect(health.Status).To(Equal(domain.PeerStatusActive))
				Expect(health.LastHandshake.Unix()).To(BeNumerically("~", time.Now().Add(-1*time.Minute).Unix(), 1))
				Expect(health.Latency).To(Equal(50 * time.Millisecond))
				Expect(health.PacketLoss).To(Equal(0.05))
				Expect(health.ConnectionQuality).To(BeNumerically(">", 0))
			})
		})

		Context("when peer does not exist", func() {
			It("should return an error", func() {
				health, err := peerService.GetPeerHealth(ctx, "non-existent-tunnel", "non-existent-peer")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tunnel not found"))
				Expect(health).To(BeNil())
			})
		})
	})

	Describe("EnablePeer", func() {
		Context("when peer exists", func() {
			It("should enable the peer", func() {
				addReq := &domain.AddPeerRequest{
					TunnelID:  "enable-tunnel",
					PublicKey: "pubkey-enable",
				}
				createdPeer, err := peerService.AddPeer(ctx, addReq)
				Expect(err).NotTo(HaveOccurred())
				peerService.GetPeersMap()[addReq.TunnelID][createdPeer.ID].Status = domain.PeerStatusInactive // Set initial status

				err = peerService.EnablePeer(ctx, addReq.TunnelID, createdPeer.ID)
				Expect(err).NotTo(HaveOccurred())

				updatedPeer, err := peerService.GetPeer(ctx, addReq.TunnelID, createdPeer.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedPeer.Status).To(Equal(domain.PeerStatusActive))
				Expect(updatedPeer.UpdatedAt).NotTo(BeZero())
			})
		})

		Context("when peer does not exist", func() {
			It("should return an error", func() {
				err := peerService.EnablePeer(ctx, "non-existent-tunnel", "non-existent-peer")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tunnel not found"))
			})
		})
	})

	Describe("DisablePeer", func() {
		Context("when peer exists", func() {
			It("should disable the peer", func() {
				addReq := &domain.AddPeerRequest{
					TunnelID:  "disable-tunnel",
					PublicKey: "pubkey-disable",
				}
				createdPeer, err := peerService.AddPeer(ctx, addReq)
				Expect(err).NotTo(HaveOccurred())
				peerService.GetPeersMap()[addReq.TunnelID][createdPeer.ID].Status = domain.PeerStatusActive // Set initial status

				err = peerService.DisablePeer(ctx, addReq.TunnelID, createdPeer.ID)
				Expect(err).NotTo(HaveOccurred())

				updatedPeer, err := peerService.GetPeer(ctx, addReq.TunnelID, createdPeer.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(updatedPeer.Status).To(Equal(domain.PeerStatusInactive))
				Expect(updatedPeer.UpdatedAt).NotTo(BeZero())
			})
		})

		Context("when peer does not exist", func() {
			It("should return an error", func() {
				err := peerService.DisablePeer(ctx, "non-existent-tunnel", "non-existent-peer")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("tunnel not found"))
			})
		})
	})
})