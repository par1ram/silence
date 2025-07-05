package domain

import (
	"time"
)

// TunnelStatus статус туннеля
type TunnelStatus string

const (
	TunnelStatusActive     TunnelStatus = "active"
	TunnelStatusInactive   TunnelStatus = "inactive"
	TunnelStatusError      TunnelStatus = "error"
	TunnelStatusRecovering TunnelStatus = "recovering"
)

// PeerStatus статус пира
type PeerStatus string

const (
	PeerStatusActive   PeerStatus = "active"
	PeerStatusInactive PeerStatus = "inactive"
	PeerStatusError    PeerStatus = "error"
	PeerStatusOffline  PeerStatus = "offline"
)

// Tunnel VPN туннель
type Tunnel struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Interface  string       `json:"interface"`
	Status     TunnelStatus `json:"status"`
	PublicKey  string       `json:"public_key"`
	PrivateKey string       `json:"private_key,omitempty"`
	ListenPort int          `json:"listen_port"`
	MTU        int          `json:"mtu"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	// Новые поля для мониторинга
	LastHealthCheck  time.Time `json:"last_health_check,omitempty"`
	HealthStatus     string    `json:"health_status,omitempty"`
	AutoRecovery     bool      `json:"auto_recovery"`
	RecoveryAttempts int       `json:"recovery_attempts"`
}

// Peer пир в туннеле
type Peer struct {
	ID                  string     `json:"id"`
	TunnelID            string     `json:"tunnel_id"`
	Name                string     `json:"name,omitempty"`
	PublicKey           string     `json:"public_key"`
	AllowedIPs          []string   `json:"allowed_ips"`
	Endpoint            string     `json:"endpoint,omitempty"`
	PersistentKeepalive int        `json:"persistent_keepalive,omitempty"`
	Status              PeerStatus `json:"status"`
	LastHandshake       time.Time  `json:"last_handshake,omitempty"`
	TransferRx          int64      `json:"transfer_rx"`
	TransferTx          int64      `json:"transfer_tx"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	// Новые поля для мониторинга
	LastSeen          time.Time     `json:"last_seen,omitempty"`
	ConnectionQuality float64       `json:"connection_quality,omitempty"` // 0.0 - 1.0
	Latency           time.Duration `json:"latency,omitempty"`
	PacketLoss        float64       `json:"packet_loss,omitempty"` // 0.0 - 1.0
}

// TunnelStats статистика туннеля
type TunnelStats struct {
	TunnelID    string    `json:"tunnel_id"`
	BytesRx     int64     `json:"bytes_rx"`
	BytesTx     int64     `json:"bytes_tx"`
	PeersCount  int       `json:"peers_count"`
	ActivePeers int       `json:"active_peers"`
	LastUpdated time.Time `json:"last_updated"`
	// Новые поля для детальной статистики
	Uptime        time.Duration `json:"uptime"`
	ErrorCount    int           `json:"error_count"`
	RecoveryCount int           `json:"recovery_count"`
}

// CreateTunnelRequest запрос на создание туннеля
type CreateTunnelRequest struct {
	Name         string `json:"name"`
	ListenPort   int    `json:"listen_port"`
	MTU          int    `json:"mtu"`
	AutoRecovery bool   `json:"auto_recovery"`
}

// AddPeerRequest запрос на добавление пира
type AddPeerRequest struct {
	TunnelID            string   `json:"tunnel_id"`
	Name                string   `json:"name,omitempty"`
	PublicKey           string   `json:"public_key"`
	AllowedIPs          []string `json:"allowed_ips"`
	Endpoint            string   `json:"endpoint,omitempty"`
	PersistentKeepalive int      `json:"persistent_keepalive,omitempty"`
}

// HealthCheckRequest запрос на проверку здоровья
type HealthCheckRequest struct {
	TunnelID string `json:"tunnel_id"`
}

// HealthCheckResponse ответ на проверку здоровья
type HealthCheckResponse struct {
	TunnelID    string        `json:"tunnel_id"`
	Status      string        `json:"status"`
	LastCheck   time.Time     `json:"last_check"`
	PeersHealth []PeerHealth  `json:"peers_health"`
	Uptime      time.Duration `json:"uptime"`
	ErrorCount  int           `json:"error_count"`
}

// PeerHealth здоровье пира
type PeerHealth struct {
	PeerID            string        `json:"peer_id"`
	Status            PeerStatus    `json:"status"`
	LastHandshake     time.Time     `json:"last_handshake"`
	Latency           time.Duration `json:"latency"`
	PacketLoss        float64       `json:"packet_loss"`
	ConnectionQuality float64       `json:"connection_quality"`
}
