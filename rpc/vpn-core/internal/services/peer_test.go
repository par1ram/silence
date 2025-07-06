package services_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	services "github.com/par1ram/silence/rpc/vpn-core/internal/services"
	"go.uber.org/zap"
)

var _ = Describe("PeerService", func() {
	var peerService ports.PeerManager
	var ctx context.Context
	var logger *zap.Logger

	BeforeEach(func() {
		logger = zap.NewNop()
		peerService = services.NewPeerService(logger)
		ctx = context.Background()
	})

	Describe("AddPeer", func() {
		It("should add peer to tunnel", func() {
			request := &domain.AddPeerRequest{
				TunnelID:            "tunnel-123",
				Name:                "test-peer",
				PublicKey:           "public-key-123",
				AllowedIPs:          []string{"10.0.0.2/32"},
				Endpoint:            "192.168.1.100:51820",
				PersistentKeepalive: 25,
			}

			peer, err := peerService.AddPeer(ctx, request)

			Expect(err).To(BeNil())
			Expect(peer).NotTo(BeNil())
			Expect(peer.TunnelID).To(Equal(request.TunnelID))
			Expect(peer.Name).To(Equal(request.Name))
			Expect(peer.PublicKey).To(Equal(request.PublicKey))
			Expect(peer.AllowedIPs).To(Equal(request.AllowedIPs))
			Expect(peer.Endpoint).To(Equal(request.Endpoint))
			Expect(peer.PersistentKeepalive).To(Equal(request.PersistentKeepalive))
			Expect(peer.Status).To(Equal(domain.PeerStatusInactive))
			Expect(peer.CreatedAt).NotTo(BeZero())
		})

		It("should add multiple peers to same tunnel", func() {
			request1 := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "peer-1",
				PublicKey:  "public-key-1",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			request2 := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "peer-2",
				PublicKey:  "public-key-2",
				AllowedIPs: []string{"10.0.0.3/32"},
			}

			peer1, err1 := peerService.AddPeer(ctx, request1)
			Expect(err1).To(BeNil())
			Expect(peer1).NotTo(BeNil())

			peer2, err2 := peerService.AddPeer(ctx, request2)
			Expect(err2).To(BeNil())
			Expect(peer2).NotTo(BeNil())

			Expect(peer1.ID).NotTo(Equal(peer2.ID))
		})
	})

	Describe("GetPeer", func() {
		It("should return peer by ID", func() {
			request := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "test-peer",
				PublicKey:  "public-key-123",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			createdPeer, _ := peerService.AddPeer(ctx, request)

			peer, err := peerService.GetPeer(ctx, request.TunnelID, createdPeer.ID)

			Expect(err).To(BeNil())
			Expect(peer).NotTo(BeNil())
			Expect(peer.ID).To(Equal(createdPeer.ID))
			Expect(peer.Name).To(Equal(createdPeer.Name))
		})

		It("should return error for nonexistent tunnel", func() {
			peer, err := peerService.GetPeer(ctx, "nonexistent-tunnel", "peer-123")

			Expect(err).NotTo(BeNil())
			Expect(peer).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("tunnel not found"))
		})

		It("should return error for nonexistent peer", func() {
			request := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "test-peer",
				PublicKey:  "public-key-123",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			peerService.AddPeer(ctx, request)

			peer, err := peerService.GetPeer(ctx, "tunnel-123", "nonexistent-peer")

			Expect(err).NotTo(BeNil())
			Expect(peer).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("peer not found"))
		})
	})

	Describe("ListPeers", func() {
		It("should return list of peers for tunnel", func() {
			request1 := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "peer-1",
				PublicKey:  "public-key-1",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			request2 := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "peer-2",
				PublicKey:  "public-key-2",
				AllowedIPs: []string{"10.0.0.3/32"},
			}
			peerService.AddPeer(ctx, request1)
			peerService.AddPeer(ctx, request2)

			peers, err := peerService.ListPeers(ctx, "tunnel-123")

			Expect(err).To(BeNil())
			Expect(peers).NotTo(BeNil())
			Expect(len(peers)).To(Equal(2))
		})

		It("should return empty list for tunnel without peers", func() {
			peers, err := peerService.ListPeers(ctx, "empty-tunnel")

			Expect(err).NotTo(BeNil())
			Expect(peers).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("tunnel not found"))
		})

		It("should return error for nonexistent tunnel", func() {
			peers, err := peerService.ListPeers(ctx, "nonexistent-tunnel")

			Expect(err).NotTo(BeNil())
			Expect(peers).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("tunnel not found"))
		})
	})

	Describe("RemovePeer", func() {
		It("should remove peer from tunnel", func() {
			request := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "test-peer",
				PublicKey:  "public-key-123",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			createdPeer, _ := peerService.AddPeer(ctx, request)

			err := peerService.RemovePeer(ctx, request.TunnelID, createdPeer.ID)

			Expect(err).To(BeNil())

			// Verify peer is removed
			peer, err := peerService.GetPeer(ctx, request.TunnelID, createdPeer.ID)
			Expect(err).NotTo(BeNil())
			Expect(peer).To(BeNil())
		})

		It("should return error for nonexistent tunnel", func() {
			err := peerService.RemovePeer(ctx, "nonexistent-tunnel", "peer-123")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("tunnel not found"))
		})

		It("should return error for nonexistent peer", func() {
			request := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "test-peer",
				PublicKey:  "public-key-123",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			peerService.AddPeer(ctx, request)

			err := peerService.RemovePeer(ctx, "tunnel-123", "nonexistent-peer")

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("peer not found"))
		})
	})

	Describe("UpdatePeerStats", func() {
		It("should update peer statistics", func() {
			request := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "test-peer",
				PublicKey:  "public-key-123",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			createdPeer, _ := peerService.AddPeer(ctx, request)

			stats := &ports.PeerStats{
				TransferRx:    1024,
				TransferTx:    2048,
				LastHandshake: 1234567890,
			}

			err := peerService.UpdatePeerStats(ctx, request.TunnelID, createdPeer.ID, stats)

			Expect(err).To(BeNil())

			// Verify stats are updated
			peer, err := peerService.GetPeer(ctx, request.TunnelID, createdPeer.ID)
			Expect(err).To(BeNil())
			Expect(peer.TransferRx).To(Equal(int64(1024)))
			Expect(peer.TransferTx).To(Equal(int64(2048)))
			Expect(peer.LastHandshake).NotTo(BeZero())
		})

		It("should update peer status based on last handshake", func() {
			request := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "test-peer",
				PublicKey:  "public-key-123",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			createdPeer, _ := peerService.AddPeer(ctx, request)

			// Recent handshake - should be active
			stats := &ports.PeerStats{
				TransferRx:    1024,
				TransferTx:    2048,
				LastHandshake: time.Now().Unix(),
			}

			err := peerService.UpdatePeerStats(ctx, request.TunnelID, createdPeer.ID, stats)
			Expect(err).To(BeNil())

			peer, err := peerService.GetPeer(ctx, request.TunnelID, createdPeer.ID)
			Expect(err).To(BeNil())
			Expect(peer.Status).To(Equal(domain.PeerStatusActive))
		})
	})

	Describe("GetPeerHealth", func() {
		It("should return peer health information", func() {
			request := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "test-peer",
				PublicKey:  "public-key-123",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			createdPeer, _ := peerService.AddPeer(ctx, request)

			health, err := peerService.GetPeerHealth(ctx, request.TunnelID, createdPeer.ID)

			Expect(err).To(BeNil())
			Expect(health).NotTo(BeNil())
			Expect(health.PeerID).To(Equal(createdPeer.ID))
			Expect(health.Status).To(Equal(domain.PeerStatusInactive))
		})
	})

	Describe("EnablePeer", func() {
		It("should enable peer", func() {
			request := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "test-peer",
				PublicKey:  "public-key-123",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			createdPeer, _ := peerService.AddPeer(ctx, request)

			err := peerService.EnablePeer(ctx, request.TunnelID, createdPeer.ID)

			Expect(err).To(BeNil())

			peer, err := peerService.GetPeer(ctx, request.TunnelID, createdPeer.ID)
			Expect(err).To(BeNil())
			Expect(peer.Status).To(Equal(domain.PeerStatusActive))
		})
	})

	Describe("DisablePeer", func() {
		It("should disable peer", func() {
			request := &domain.AddPeerRequest{
				TunnelID:   "tunnel-123",
				Name:       "test-peer",
				PublicKey:  "public-key-123",
				AllowedIPs: []string{"10.0.0.2/32"},
			}
			createdPeer, _ := peerService.AddPeer(ctx, request)

			// First enable the peer
			peerService.EnablePeer(ctx, request.TunnelID, createdPeer.ID)

			// Then disable it
			err := peerService.DisablePeer(ctx, request.TunnelID, createdPeer.ID)

			Expect(err).To(BeNil())

			peer, err := peerService.GetPeer(ctx, request.TunnelID, createdPeer.ID)
			Expect(err).To(BeNil())
			Expect(peer.Status).To(Equal(domain.PeerStatusInactive))
		})
	})
})
