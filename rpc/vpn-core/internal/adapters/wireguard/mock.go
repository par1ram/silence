package wireguard

import (
	"net"
	"time"

	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"go.uber.org/zap"
)

// MockWGAdapter mock адаптер для тестирования WireGuard
type MockWGAdapter struct {
	logger *zap.Logger
}

// NewMockWGAdapter создает новый mock WireGuard адаптер
func NewMockWGAdapter(logger *zap.Logger) *MockWGAdapter {
	return &MockWGAdapter{
		logger: logger,
	}
}

// CreateInterface создает mock WireGuard интерфейс
func (m *MockWGAdapter) CreateInterface(name, privateKey string, listenPort, mtu int) error {
	m.logger.Info("mock: creating wireguard interface",
		zap.String("name", name),
		zap.Int("port", listenPort),
		zap.Int("mtu", mtu))
	return nil
}

// DeleteInterface удаляет mock WireGuard интерфейс
func (m *MockWGAdapter) DeleteInterface(name string) error {
	m.logger.Info("mock: deleting wireguard interface", zap.String("name", name))
	return nil
}

// AddPeer добавляет пира к mock интерфейсу
func (m *MockWGAdapter) AddPeer(deviceName, publicKey string, allowedIPs []net.IPNet, endpoint *net.UDPAddr, keepalive int) error {
	m.logger.Info("mock: adding peer to wireguard interface",
		zap.String("device", deviceName),
		zap.String("public_key", publicKey))
	return nil
}

// RemovePeer удаляет пира из mock интерфейса
func (m *MockWGAdapter) RemovePeer(deviceName, publicKey string) error {
	m.logger.Info("mock: removing peer from wireguard interface",
		zap.String("device", deviceName),
		zap.String("public_key", publicKey))
	return nil
}

// GetDeviceStats получает mock статистику устройства
func (m *MockWGAdapter) GetDeviceStats(deviceName string) (interface{}, error) {
	m.logger.Info("mock: getting device stats", zap.String("device", deviceName))
	return map[string]interface{}{
		"device": deviceName,
		"peers":  0,
		"rx":     0,
		"tx":     0,
	}, nil
}

// GetInterfaceStats получает mock статистику интерфейса
func (m *MockWGAdapter) GetInterfaceStats(interfaceName string) (*ports.InterfaceStats, error) {
	m.logger.Info("mock: getting interface stats", zap.String("interface", interfaceName))
	return &ports.InterfaceStats{
		InterfaceName: interfaceName,
		BytesRx:       1024 * 1024, // 1MB
		BytesTx:       512 * 1024,  // 512KB
		PeersCount:    2,
		LastUpdated:   time.Now().Unix(),
	}, nil
}

// GetPeerStats получает mock статистику пира
func (m *MockWGAdapter) GetPeerStats(interfaceName, publicKey string) (*ports.PeerStats, error) {
	m.logger.Info("mock: getting peer stats",
		zap.String("interface", interfaceName),
		zap.String("public_key", publicKey))
	return &ports.PeerStats{
		PublicKey:           publicKey,
		LastHandshake:       time.Now().Unix(),
		TransferRx:          1024 * 1024, // 1MB
		TransferTx:          512 * 1024,  // 512KB
		PersistentKeepalive: 25,
	}, nil
}

// Close закрывает mock клиент
func (m *MockWGAdapter) Close() error {
	m.logger.Info("mock: closing wireguard adapter")
	return nil
}
