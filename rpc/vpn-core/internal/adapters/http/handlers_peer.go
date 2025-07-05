package http

import (
	"encoding/json"
	"net/http"

	"github.com/par1ram/silence/rpc/vpn-core/internal/domain"
	"go.uber.org/zap"
)

// AddPeerHandler добавляет пира к туннелю
func (h *Handlers) AddPeerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.AddPeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	peer, err := h.peerManager.AddPeer(r.Context(), &req)
	if err != nil {
		h.logger.Error("failed to add peer", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peer)
}

// GetPeerHandler получает пира по ID
func (h *Handlers) GetPeerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tunnelID := r.URL.Query().Get("tunnel_id")
	peerID := r.URL.Query().Get("peer_id")

	if tunnelID == "" || peerID == "" {
		http.Error(w, "Missing tunnel_id or peer_id", http.StatusBadRequest)
		return
	}

	peer, err := h.peerManager.GetPeer(r.Context(), tunnelID, peerID)
	if err != nil {
		h.logger.Error("failed to get peer", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(peer)
}

// ListPeersHandler возвращает список пиров туннеля
func (h *Handlers) ListPeersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tunnelID := r.URL.Query().Get("tunnel_id")
	if tunnelID == "" {
		http.Error(w, "Missing tunnel_id", http.StatusBadRequest)
		return
	}

	peers, err := h.peerManager.ListPeers(r.Context(), tunnelID)
	if err != nil {
		h.logger.Error("failed to list peers", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"peers": peers,
	})
}

// RemovePeerHandler удаляет пира из туннеля
func (h *Handlers) RemovePeerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tunnelID := r.URL.Query().Get("tunnel_id")
	peerID := r.URL.Query().Get("peer_id")

	if tunnelID == "" || peerID == "" {
		http.Error(w, "Missing tunnel_id or peer_id", http.StatusBadRequest)
		return
	}

	err := h.peerManager.RemovePeer(r.Context(), tunnelID, peerID)
	if err != nil {
		h.logger.Error("failed to remove peer", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
	})
}
