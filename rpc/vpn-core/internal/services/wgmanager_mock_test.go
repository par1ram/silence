package services

import (
	"net"

	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
)

type mockWGManager struct {
	CreateErr             error
	DeleteErr             error
	Stats                 *ports.InterfaceStats
	StatsErr              error
	GetInterfaceStatsFunc func(interfaceName string) (*ports.InterfaceStats, error)
}

func (m *mockWGManager) CreateInterface(name, privateKey string, port, mtu int) error {
	return m.CreateErr
}
func (m *mockWGManager) DeleteInterface(name string) error {
	return m.DeleteErr
}
func (m *mockWGManager) AddPeer(deviceName, publicKey string, allowedIPs []net.IPNet, endpoint *net.UDPAddr, keepalive int) error {
	return nil
}
func (m *mockWGManager) RemovePeer(deviceName, publicKey string) error {
	return nil
}
func (m *mockWGManager) GetDeviceStats(deviceName string) (interface{}, error) {
	return nil, nil
}
func (m *mockWGManager) GetInterfaceStats(interfaceName string) (*ports.InterfaceStats, error) {
	if m.GetInterfaceStatsFunc != nil {
		return m.GetInterfaceStatsFunc(interfaceName)
	}
	if m.Stats != nil || m.StatsErr != nil {
		return m.Stats, m.StatsErr
	}
	return &ports.InterfaceStats{}, nil
}
func (m *mockWGManager) GetPeerStats(interfaceName, publicKey string) (*ports.PeerStats, error) {
	return nil, nil
}
func (m *mockWGManager) Close() error {
	return nil
}
