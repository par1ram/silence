package domain

import (
	"time"
)

// MetricType определяет тип метрики
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric представляет базовую метрику
type Metric struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// ConnectionMetric метрики подключений
type ConnectionMetric struct {
	Metric
	UserID     string `json:"user_id"`
	ServerID   string `json:"server_id"`
	Protocol   string `json:"protocol"`
	BypassType string `json:"bypass_type"`
	Region     string `json:"region"`
	Duration   int64  `json:"duration_ms"`
	BytesIn    int64  `json:"bytes_in"`
	BytesOut   int64  `json:"bytes_out"`
}

// BypassEffectivenessMetric эффективность обхода DPI
type BypassEffectivenessMetric struct {
	Metric
	BypassType    string  `json:"bypass_type"`
	SuccessRate   float64 `json:"success_rate"`
	Latency       int64   `json:"latency_ms"`
	Throughput    float64 `json:"throughput_mbps"`
	BlockedCount  int64   `json:"blocked_count"`
	TotalAttempts int64   `json:"total_attempts"`
}

// UserActivityMetric активность пользователей
type UserActivityMetric struct {
	Metric
	UserID       string `json:"user_id"`
	SessionCount int64  `json:"session_count"`
	TotalTime    int64  `json:"total_time_minutes"`
	DataUsage    int64  `json:"data_usage_mb"`
	LoginCount   int64  `json:"login_count"`
}

// ServerLoadMetric нагрузка на серверы
type ServerLoadMetric struct {
	Metric
	ServerID    string  `json:"server_id"`
	Region      string  `json:"region"`
	CPUUsage    float64 `json:"cpu_usage_percent"`
	MemoryUsage float64 `json:"memory_usage_percent"`
	NetworkIn   float64 `json:"network_in_mbps"`
	NetworkOut  float64 `json:"network_out_mbps"`
	Connections int64   `json:"active_connections"`
}

// ErrorMetric метрики ошибок
type ErrorMetric struct {
	Metric
	ErrorType   string `json:"error_type"`
	Service     string `json:"service"`
	UserID      string `json:"user_id,omitempty"`
	ServerID    string `json:"server_id,omitempty"`
	StatusCode  int    `json:"status_code,omitempty"`
	Description string `json:"description"`
}

// TimeRange временной диапазон для запросов
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// AggregationType тип агрегации
type AggregationType string

const (
	AggregationSum     AggregationType = "sum"
	AggregationAvg     AggregationType = "avg"
	AggregationMin     AggregationType = "min"
	AggregationMax     AggregationType = "max"
	AggregationCount   AggregationType = "count"
	AggregationPercent AggregationType = "percent"
)

// QueryOptions опции для запросов метрик
type QueryOptions struct {
	TimeRange   TimeRange         `json:"time_range"`
	Aggregation AggregationType   `json:"aggregation"`
	GroupBy     []string          `json:"group_by,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
	Limit       int               `json:"limit,omitempty"`
	Offset      int               `json:"offset,omitempty"`
	Interval    string            `json:"interval,omitempty"` // для временных серий
}

// MetricResponse ответ с метриками
type MetricResponse struct {
	Metrics []Metric `json:"metrics"`
	Total   int64    `json:"total"`
	HasMore bool     `json:"has_more"`
}

// DashboardConfig конфигурация дашборда
type DashboardConfig struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Widgets     []DashboardWidget      `json:"widgets"`
	Layout      map[string]interface{} `json:"layout"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// DashboardWidget виджет дашборда
type DashboardWidget struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // chart, metric, table, etc.
	Title    string                 `json:"title"`
	Query    QueryOptions           `json:"query"`
	Config   map[string]interface{} `json:"config"`
	Position map[string]interface{} `json:"position"`
}
