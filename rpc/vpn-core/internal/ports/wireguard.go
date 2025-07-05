package ports

import "net"

// WireGuardManager интерфейс для управления WireGuard интерфейсами
type WireGuardManager interface {
	CreateInterface(name, privateKey string, listenPort, mtu int) error
	DeleteInterface(name string) error
	AddPeer(deviceName, publicKey string, allowedIPs []net.IPNet, endpoint *net.UDPAddr, keepalive int) error
	RemovePeer(deviceName, publicKey string) error
	GetDeviceStats(deviceName string) (interface{}, error)
	// Новые методы для мониторинга
	GetInterfaceStats(interfaceName string) (*InterfaceStats, error)
	GetPeerStats(interfaceName, publicKey string) (*PeerStats, error)
	Close() error
}

// InterfaceStats статистика интерфейса
type InterfaceStats struct {
	InterfaceName string `json:"interface_name"`
	BytesRx       int64  `json:"bytes_rx"`
	BytesTx       int64  `json:"bytes_tx"`
	PeersCount    int    `json:"peers_count"`
	LastUpdated   int64  `json:"last_updated"`
}

// PeerStats статистика пира
type PeerStats struct {
	PublicKey           string `json:"public_key"`
	LastHandshake       int64  `json:"last_handshake"`
	TransferRx          int64  `json:"transfer_rx"`
	TransferTx          int64  `json:"transfer_tx"`
	PersistentKeepalive int    `json:"persistent_keepalive"`
}
