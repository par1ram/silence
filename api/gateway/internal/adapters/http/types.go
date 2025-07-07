package http

import "time"

// ===== VPN Connection Types =====

type VPNConnectRequest struct {
	Name         string `json:"name"`
	ListenPort   int    `json:"listen_port"`
	MTU          int    `json:"mtu"`
	AutoRecovery bool   `json:"auto_recovery"`
	ServerID     string `json:"server_id,omitempty"`
	Region       string `json:"region,omitempty"`
}

type VPNConnectResponse struct {
	TunnelID   string    `json:"tunnel_id"`
	TunnelName string    `json:"tunnel_name"`
	PublicKey  string    `json:"public_key"`
	Endpoint   string    `json:"endpoint"`
	Status     string    `json:"status"`
	Config     VPNConfig `json:"config"`
	CreatedAt  time.Time `json:"created_at"`
}

type VPNConfig struct {
	Interface  string   `json:"interface"`
	PrivateKey string   `json:"private_key"`
	Address    string   `json:"address"`
	DNS        []string `json:"dns"`
	Peer       VPNPeer  `json:"peer"`
}

type VPNPeer struct {
	PublicKey  string   `json:"public_key"`
	AllowedIPs []string `json:"allowed_ips"`
	Endpoint   string   `json:"endpoint"`
}

// ===== DPI Connection Types =====

type DPIConnectRequest struct {
	Method     string `json:"method"`
	Name       string `json:"name"`
	RemoteHost string `json:"remote_host"`
	RemotePort int    `json:"remote_port"`
	LocalPort  int    `json:"local_port,omitempty"`
	Password   string `json:"password,omitempty"`
	Encryption string `json:"encryption,omitempty"`
	ServerID   string `json:"server_id,omitempty"`
}

type DPIConnectResponse struct {
	BypassID  string                 `json:"bypass_id"`
	Method    string                 `json:"method"`
	LocalPort int                    `json:"local_port"`
	Status    string                 `json:"status"`
	Config    map[string]interface{} `json:"config"`
	CreatedAt time.Time              `json:"created_at"`
}

// ===== Shadowsocks Types =====

type ShadowsocksConnectRequest struct {
	Name       string `json:"name"`
	ServerHost string `json:"server_host"`
	ServerPort int    `json:"server_port"`
	Password   string `json:"password"`
	Encryption string `json:"encryption"`
	LocalPort  int    `json:"local_port,omitempty"`
	Timeout    int    `json:"timeout,omitempty"`
}

type ShadowsocksConnectResponse struct {
	ConnectionID string            `json:"connection_id"`
	LocalPort    int               `json:"local_port"`
	Status       string            `json:"status"`
	Config       ShadowsocksConfig `json:"config"`
	CreatedAt    time.Time         `json:"created_at"`
}

type ShadowsocksConfig struct {
	Server     string `json:"server"`
	ServerPort int    `json:"server_port"`
	LocalPort  int    `json:"local_port"`
	Password   string `json:"password"`
	Method     string `json:"method"`
}

// ===== V2Ray Types =====

type V2RayConnectRequest struct {
	Name       string `json:"name"`
	ServerHost string `json:"server_host"`
	ServerPort int    `json:"server_port"`
	UUID       string `json:"uuid"`
	AlterID    int    `json:"alter_id"`
	Security   string `json:"security,omitempty"`
	Network    string `json:"network,omitempty"`
	LocalPort  int    `json:"local_port,omitempty"`
}

type V2RayConnectResponse struct {
	ConnectionID string      `json:"connection_id"`
	LocalPort    int         `json:"local_port"`
	Status       string      `json:"status"`
	Config       V2RayConfig `json:"config"`
	CreatedAt    time.Time   `json:"created_at"`
}

type V2RayConfig struct {
	Inbounds  []interface{} `json:"inbounds"`
	Outbounds []interface{} `json:"outbounds"`
	Routing   interface{}   `json:"routing"`
}

// ===== Obfs4 Types =====

type Obfs4ConnectRequest struct {
	Name      string `json:"name"`
	Bridge    string `json:"bridge"`
	Cert      string `json:"cert"`
	IATMode   string `json:"iat_mode,omitempty"`
	LocalPort int    `json:"local_port,omitempty"`
}

type Obfs4ConnectResponse struct {
	ConnectionID string      `json:"connection_id"`
	LocalPort    int         `json:"local_port"`
	Status       string      `json:"status"`
	Config       Obfs4Config `json:"config"`
	CreatedAt    time.Time   `json:"created_at"`
}

type Obfs4Config struct {
	Bridge    string `json:"bridge"`
	Transport string `json:"transport"`
	LocalPort int    `json:"local_port"`
}

// ===== Connection Management Types =====

type DisconnectRequest struct {
	ConnectionID string `json:"connection_id,omitempty"`
	TunnelID     string `json:"tunnel_id,omitempty"`
	BypassID     string `json:"bypass_id,omitempty"`
	All          bool   `json:"all,omitempty"`
}

type DisconnectResponse struct {
	Disconnected []string `json:"disconnected"`
	Status       string   `json:"status"`
	Errors       []string `json:"errors,omitempty"`
}

// ===== Status Types =====

type ConnectionStatusResponse struct {
	VPNTunnels           []VPNTunnelStatus `json:"vpn_tunnels"`
	DPIBypasses          []DPIBypassStatus `json:"dpi_bypasses"`
	ActiveConnections    int               `json:"active_connections"`
	TotalDataTransferred int64             `json:"total_data_transferred"`
	Uptime               int64             `json:"uptime"`
}

type VPNTunnelStatus struct {
	TunnelID    string    `json:"tunnel_id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	ConnectedAt time.Time `json:"connected_at"`
	PeersCount  int       `json:"peers_count"`
	BytesRx     int64     `json:"bytes_rx"`
	BytesTx     int64     `json:"bytes_tx"`
}

type DPIBypassStatus struct {
	BypassID    string    `json:"bypass_id"`
	Name        string    `json:"name"`
	Method      string    `json:"method"`
	Status      string    `json:"status"`
	LocalPort   int       `json:"local_port"`
	ConnectedAt time.Time `json:"connected_at"`
	BytesRx     int64     `json:"bytes_rx"`
	BytesTx     int64     `json:"bytes_tx"`
}
