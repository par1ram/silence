package domain

import "time"

// BypassMethod тип метода обхода DPI
type BypassMethod string

const (
	BypassMethodShadowsocks BypassMethod = "shadowsocks"
	BypassMethodV2Ray       BypassMethod = "v2ray"
	BypassMethodObfs4       BypassMethod = "obfs4"
	BypassMethodCustom      BypassMethod = "custom"
)

// BypassConfig конфигурация для обхода DPI
type BypassConfig struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Method     BypassMethod `json:"method"`
	LocalPort  int          `json:"local_port"`
	RemoteHost string       `json:"remote_host"`
	RemotePort int          `json:"remote_port"`
	Password   string       `json:"password,omitempty"`
	Encryption string       `json:"encryption"`
	Status     string       `json:"status"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
}

// CreateBypassRequest запрос на создание bypass конфигурации
type CreateBypassRequest struct {
	Name       string       `json:"name"`
	Method     BypassMethod `json:"method"`
	LocalPort  int          `json:"local_port"`
	RemoteHost string       `json:"remote_host"`
	RemotePort int          `json:"remote_port"`
	Password   string       `json:"password,omitempty"`
	Encryption string       `json:"encryption"`
}

// BypassStats статистика bypass соединения
type BypassStats struct {
	ID           string    `json:"id"`
	BytesRx      int64     `json:"bytes_rx"`
	BytesTx      int64     `json:"bytes_tx"`
	Connections  int       `json:"connections"`
	LastActivity time.Time `json:"last_activity"`
	ErrorCount   int       `json:"error_count"`
}
