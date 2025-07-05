package wireguard

import (
	"fmt"
	"net"
	"time"

	"github.com/par1ram/silence/rpc/vpn-core/internal/ports"
	"go.uber.org/zap"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// WGAdapter адаптер для управления WireGuard интерфейсами
type WGAdapter struct {
	client *wgctrl.Client
	logger *zap.Logger
}

// NewWGAdapter создает новый WireGuard адаптер
func NewWGAdapter(logger *zap.Logger) (*WGAdapter, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create wgctrl client: %w", err)
	}

	return &WGAdapter{
		client: client,
		logger: logger,
	}, nil
}

// CreateInterface создает WireGuard интерфейс
func (w *WGAdapter) CreateInterface(name, privateKey string, listenPort, mtu int) error {
	// Декодируем приватный ключ
	key, err := wgtypes.ParseKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Конфигурация устройства
	cfg := wgtypes.Config{
		PrivateKey: &key,
		ListenPort: &listenPort,
	}

	// Создаем устройство
	if err := w.client.ConfigureDevice(name, cfg); err != nil {
		return fmt.Errorf("failed to configure device %s: %w", name, err)
	}

	w.logger.Info("wireguard interface created",
		zap.String("name", name),
		zap.Int("port", listenPort))

	return nil
}

// DeleteInterface удаляет WireGuard интерфейс
func (w *WGAdapter) DeleteInterface(name string) error {
	// Останавливаем устройство (удаляем конфигурацию)
	cfg := wgtypes.Config{
		PrivateKey: &wgtypes.Key{},
		ListenPort: nil,
	}

	if err := w.client.ConfigureDevice(name, cfg); err != nil {
		return fmt.Errorf("failed to delete device %s: %w", name, err)
	}

	w.logger.Info("wireguard interface deleted", zap.String("name", name))
	return nil
}

// AddPeer добавляет пира к интерфейсу
func (w *WGAdapter) AddPeer(deviceName, publicKey string, allowedIPs []net.IPNet, endpoint *net.UDPAddr, keepalive int) error {
	// Декодируем публичный ключ
	key, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	// Получаем текущую конфигурацию
	device, err := w.client.Device(deviceName)
	if err != nil {
		return fmt.Errorf("failed to get device %s: %w", deviceName, err)
	}

	// Конвертируем keepalive в Duration
	keepaliveDuration := time.Duration(keepalive) * time.Second

	// Конвертируем существующих пиров в PeerConfig
	var peerConfigs []wgtypes.PeerConfig
	for _, peer := range device.Peers {
		peerConfigs = append(peerConfigs, wgtypes.PeerConfig{
			PublicKey:                   peer.PublicKey,
			AllowedIPs:                  peer.AllowedIPs,
			Endpoint:                    peer.Endpoint,
			PersistentKeepaliveInterval: &peer.PersistentKeepaliveInterval,
		})
	}

	// Добавляем нового пира
	peerConfigs = append(peerConfigs, wgtypes.PeerConfig{
		PublicKey:                   key,
		AllowedIPs:                  allowedIPs,
		Endpoint:                    endpoint,
		PersistentKeepaliveInterval: &keepaliveDuration,
	})

	// Обновляем конфигурацию
	cfg := wgtypes.Config{
		Peers: peerConfigs,
	}

	if err := w.client.ConfigureDevice(deviceName, cfg); err != nil {
		return fmt.Errorf("failed to add peer to device %s: %w", deviceName, err)
	}

	w.logger.Info("peer added to wireguard interface",
		zap.String("device", deviceName),
		zap.String("public_key", publicKey))

	return nil
}

// RemovePeer удаляет пира из интерфейса
func (w *WGAdapter) RemovePeer(deviceName, publicKey string) error {
	// Декодируем публичный ключ
	key, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	// Получаем текущую конфигурацию
	device, err := w.client.Device(deviceName)
	if err != nil {
		return fmt.Errorf("failed to get device %s: %w", deviceName, err)
	}

	// Удаляем пира и конвертируем в PeerConfig
	var peerConfigs []wgtypes.PeerConfig
	for _, peer := range device.Peers {
		if peer.PublicKey != key {
			peerConfigs = append(peerConfigs, wgtypes.PeerConfig{
				PublicKey:                   peer.PublicKey,
				AllowedIPs:                  peer.AllowedIPs,
				Endpoint:                    peer.Endpoint,
				PersistentKeepaliveInterval: &peer.PersistentKeepaliveInterval,
			})
		}
	}

	// Обновляем конфигурацию
	cfg := wgtypes.Config{
		Peers: peerConfigs,
	}

	if err := w.client.ConfigureDevice(deviceName, cfg); err != nil {
		return fmt.Errorf("failed to remove peer from device %s: %w", deviceName, err)
	}

	w.logger.Info("peer removed from wireguard interface",
		zap.String("device", deviceName),
		zap.String("public_key", publicKey))

	return nil
}

// GetDeviceStats получает статистику устройства
func (w *WGAdapter) GetDeviceStats(deviceName string) (interface{}, error) {
	device, err := w.client.Device(deviceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get device %s: %w", deviceName, err)
	}

	return device, nil
}

// GetInterfaceStats получает статистику интерфейса
func (w *WGAdapter) GetInterfaceStats(interfaceName string) (*ports.InterfaceStats, error) {
	device, err := w.client.Device(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get device %s: %w", interfaceName, err)
	}

	// Вычисляем общую статистику по всем пирам
	var totalRx, totalTx int64
	for _, peer := range device.Peers {
		totalRx += peer.ReceiveBytes
		totalTx += peer.TransmitBytes
	}

	stats := &ports.InterfaceStats{
		InterfaceName: interfaceName,
		BytesRx:       totalRx,
		BytesTx:       totalTx,
		PeersCount:    len(device.Peers),
		LastUpdated:   time.Now().Unix(),
	}

	return stats, nil
}

// GetPeerStats получает статистику конкретного пира
func (w *WGAdapter) GetPeerStats(interfaceName, publicKey string) (*ports.PeerStats, error) {
	device, err := w.client.Device(interfaceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get device %s: %w", interfaceName, err)
	}

	// Декодируем публичный ключ
	key, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Ищем пира
	for _, peer := range device.Peers {
		if peer.PublicKey == key {
			var lastHandshake int64
			if !peer.LastHandshakeTime.IsZero() {
				lastHandshake = peer.LastHandshakeTime.Unix()
			}

			stats := &ports.PeerStats{
				PublicKey:           publicKey,
				LastHandshake:       lastHandshake,
				TransferRx:          peer.ReceiveBytes,
				TransferTx:          peer.TransmitBytes,
				PersistentKeepalive: int(peer.PersistentKeepaliveInterval.Seconds()),
			}
			return stats, nil
		}
	}

	return nil, fmt.Errorf("peer not found: %s", publicKey)
}

// Close закрывает клиент
func (w *WGAdapter) Close() error {
	return w.client.Close()
}
