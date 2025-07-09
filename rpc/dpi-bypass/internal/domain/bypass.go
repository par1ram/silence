package domain

import "time"

// BypassType тип обхода DPI
type BypassType string

const (
	BypassTypeDomainFronting      BypassType = "domain_fronting"
	BypassTypeSNIMasking          BypassType = "sni_masking"
	BypassTypePacketFragmentation BypassType = "packet_fragmentation"
	BypassTypeProtocolObfuscation BypassType = "protocol_obfuscation"
	BypassTypeTunnelObfuscation   BypassType = "tunnel_obfuscation"
)

// BypassMethod тип метода обхода DPI
type BypassMethod string

const (
	BypassMethodHTTPHeader   BypassMethod = "http_header"
	BypassMethodTLSHandshake BypassMethod = "tls_handshake"
	BypassMethodTCPFragment  BypassMethod = "tcp_fragment"
	BypassMethodUDPFragment  BypassMethod = "udp_fragment"
	BypassMethodProxyChain   BypassMethod = "proxy_chain"
	BypassMethodShadowsocks  BypassMethod = "shadowsocks"
	BypassMethodV2Ray        BypassMethod = "v2ray"
	BypassMethodObfs4        BypassMethod = "obfs4"
	BypassMethodCustom       BypassMethod = "custom"
)

// BypassStatus статус обхода
type BypassStatus string

const (
	BypassStatusInactive BypassStatus = "inactive"
	BypassStatusActive   BypassStatus = "active"
	BypassStatusError    BypassStatus = "error"
	BypassStatusTesting  BypassStatus = "testing"
)

// RuleType тип правила обхода
type RuleType string

const (
	RuleTypeDomain   RuleType = "domain"
	RuleTypeIP       RuleType = "ip"
	RuleTypePort     RuleType = "port"
	RuleTypeProtocol RuleType = "protocol"
	RuleTypeRegex    RuleType = "regex"
)

// RuleAction действие правила
type RuleAction string

const (
	RuleActionAllow     RuleAction = "allow"
	RuleActionBlock     RuleAction = "block"
	RuleActionBypass    RuleAction = "bypass"
	RuleActionFragment  RuleAction = "fragment"
	RuleActionObfuscate RuleAction = "obfuscate"
)

// BypassConfig конфигурация для обхода DPI
type BypassConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        BypassType        `json:"type"`
	Method      BypassMethod      `json:"method"`
	Status      BypassStatus      `json:"status"`
	Parameters  map[string]string `json:"parameters"`
	Rules       []*BypassRule     `json:"rules"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// CreateBypassConfigRequest запрос на создание bypass конфигурации
type CreateBypassConfigRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        BypassType        `json:"type"`
	Method      BypassMethod      `json:"method"`
	Parameters  map[string]string `json:"parameters"`
}

// UpdateBypassConfigRequest запрос на обновление bypass конфигурации
type UpdateBypassConfigRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        BypassType        `json:"type"`
	Method      BypassMethod      `json:"method"`
	Parameters  map[string]string `json:"parameters"`
}

// BypassConfigFilters фильтры для конфигураций обхода
type BypassConfigFilters struct {
	Type   BypassType   `json:"type"`
	Status BypassStatus `json:"status"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

// StartBypassRequest запрос на запуск обхода
type StartBypassRequest struct {
	ConfigID   string            `json:"config_id"`
	TargetHost string            `json:"target_host"`
	TargetPort int               `json:"target_port"`
	Options    map[string]string `json:"options"`
}

// BypassSession сессия обхода
type BypassSession struct {
	ID         string       `json:"id"`
	ConfigID   string       `json:"config_id"`
	TargetHost string       `json:"target_host"`
	TargetPort int          `json:"target_port"`
	Status     BypassStatus `json:"status"`
	StartedAt  time.Time    `json:"started_at"`
	Message    string       `json:"message"`
}

// BypassSessionStatus статус сессии обхода
type BypassSessionStatus struct {
	SessionID       string       `json:"session_id"`
	ConfigID        string       `json:"config_id"`
	Status          BypassStatus `json:"status"`
	TargetHost      string       `json:"target_host"`
	TargetPort      int          `json:"target_port"`
	StartedAt       time.Time    `json:"started_at"`
	DurationSeconds int64        `json:"duration_seconds"`
	Message         string       `json:"message"`
}

// BypassStats статистика bypass соединения
type BypassStats struct {
	ID                     string    `json:"id"`
	ConfigID               string    `json:"config_id"`
	SessionID              string    `json:"session_id"`
	BytesSent              int64     `json:"bytes_sent"`
	BytesReceived          int64     `json:"bytes_received"`
	PacketsSent            int64     `json:"packets_sent"`
	PacketsReceived        int64     `json:"packets_received"`
	ConnectionsEstablished int64     `json:"connections_established"`
	ConnectionsFailed      int64     `json:"connections_failed"`
	SuccessRate            float64   `json:"success_rate"`
	AverageLatency         float64   `json:"average_latency"`
	StartTime              time.Time `json:"start_time"`
	EndTime                time.Time `json:"end_time"`
}

// BypassHistoryRequest запрос истории обхода
type BypassHistoryRequest struct {
	ConfigID  string    `json:"config_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// BypassHistoryEntry запись истории обхода
type BypassHistoryEntry struct {
	ID               string       `json:"id"`
	ConfigID         string       `json:"config_id"`
	SessionID        string       `json:"session_id"`
	TargetHost       string       `json:"target_host"`
	TargetPort       int          `json:"target_port"`
	Status           BypassStatus `json:"status"`
	StartedAt        time.Time    `json:"started_at"`
	EndedAt          time.Time    `json:"ended_at"`
	DurationSeconds  int64        `json:"duration_seconds"`
	BytesTransferred int64        `json:"bytes_transferred"`
	ErrorMessage     string       `json:"error_message"`
}

// BypassRule правило обхода
type BypassRule struct {
	ID         string            `json:"id"`
	ConfigID   string            `json:"config_id"`
	Name       string            `json:"name"`
	Type       RuleType          `json:"type"`
	Action     RuleAction        `json:"action"`
	Pattern    string            `json:"pattern"`
	Parameters map[string]string `json:"parameters"`
	Priority   int               `json:"priority"`
	Enabled    bool              `json:"enabled"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// AddBypassRuleRequest запрос на добавление правила обхода
type AddBypassRuleRequest struct {
	ConfigID   string            `json:"config_id"`
	Name       string            `json:"name"`
	Type       RuleType          `json:"type"`
	Action     RuleAction        `json:"action"`
	Pattern    string            `json:"pattern"`
	Parameters map[string]string `json:"parameters"`
	Priority   int               `json:"priority"`
}

// UpdateBypassRuleRequest запрос на обновление правила обхода
type UpdateBypassRuleRequest struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       RuleType          `json:"type"`
	Action     RuleAction        `json:"action"`
	Pattern    string            `json:"pattern"`
	Parameters map[string]string `json:"parameters"`
	Priority   int               `json:"priority"`
	Enabled    bool              `json:"enabled"`
}

// BypassRuleFilters фильтры для правил обхода
type BypassRuleFilters struct {
	ConfigID string   `json:"config_id"`
	Type     RuleType `json:"type"`
	Enabled  bool     `json:"enabled"`
	Limit    int      `json:"limit"`
	Offset   int      `json:"offset"`
}
