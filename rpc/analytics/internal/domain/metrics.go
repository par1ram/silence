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
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Tags      map[string]string `json:"tags,omitempty"`
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

// MetricFilters фильтры для метрик
type MetricFilters struct {
	Name      string            `json:"name"`
	Tags      map[string]string `json:"tags,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
}

// MetricHistoryRequest запрос истории метрик
type MetricHistoryRequest struct {
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Interval  string    `json:"interval"`
}

// TimeSeriesPoint точка временного ряда
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// StatisticsRequest запрос статистики
type StatisticsRequest struct {
	Type      string    `json:"type"`
	Period    string    `json:"period"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// Statistics статистика
type Statistics struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"`
	Value        float64   `json:"value"`
	Unit         string    `json:"unit"`
	CalculatedAt time.Time `json:"calculated_at"`
	Period       string    `json:"period"`
}

// SystemStats системная статистика
type SystemStats struct {
	TotalUsers           int64     `json:"total_users"`
	ActiveUsers          int64     `json:"active_users"`
	TotalConnections     int64     `json:"total_connections"`
	ActiveConnections    int64     `json:"active_connections"`
	TotalDataTransferred int64     `json:"total_data_transferred"`
	ServersCount         int64     `json:"servers_count"`
	ActiveServers        int64     `json:"active_servers"`
	AvgConnectionTime    float64   `json:"avg_connection_time"`
	SystemLoad           float64   `json:"system_load"`
	LastUpdated          time.Time `json:"last_updated"`
}

// UserStatsRequest запрос статистики пользователя
type UserStatsRequest struct {
	UserID    string    `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// UserStats статистика пользователя
type UserStats struct {
	UserID               string    `json:"user_id"`
	TotalConnections     int64     `json:"total_connections"`
	TotalDataTransferred int64     `json:"total_data_transferred"`
	TotalSessionTime     int64     `json:"total_session_time"`
	FavoriteServersCount int       `json:"favorite_servers_count"`
	AvgConnectionTime    float64   `json:"avg_connection_time"`
	FirstConnection      time.Time `json:"first_connection"`
	LastConnection       time.Time `json:"last_connection"`
}

// DashboardData данные дашборда
type DashboardData struct {
	SystemStats          *SystemStats      `json:"system_stats"`
	ConnectionsOverTime  []TimeSeriesPoint `json:"connections_over_time"`
	DataTransferOverTime []TimeSeriesPoint `json:"data_transfer_over_time"`
	ServerUsage          []ServerUsage     `json:"server_usage"`
	RegionStats          []RegionStats     `json:"region_stats"`
	Alerts               []Alert           `json:"alerts"`
}

// ServerUsage использование сервера
type ServerUsage struct {
	ServerID          string  `json:"server_id"`
	ServerName        string  `json:"server_name"`
	ActiveConnections int64   `json:"active_connections"`
	CPUUsage          float64 `json:"cpu_usage"`
	MemoryUsage       float64 `json:"memory_usage"`
	NetworkUsage      float64 `json:"network_usage"`
}

// RegionStats статистика региона
type RegionStats struct {
	Region          string  `json:"region"`
	UserCount       int64   `json:"user_count"`
	ConnectionCount int64   `json:"connection_count"`
	DataTransferred int64   `json:"data_transferred"`
	AvgLatency      float64 `json:"avg_latency"`
}

// PredictionRequest запрос предсказания
type PredictionRequest struct {
	ServerID   string `json:"server_id"`
	HoursAhead int    `json:"hours_ahead"`
}

// TrendRequest запрос тренда
type TrendRequest struct {
	MetricName string `json:"metric_name"`
	DaysAhead  int    `json:"days_ahead"`
}

// PredictionPoint точка предсказания
type PredictionPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	PredictedValue float64   `json:"predicted_value"`
	Confidence     float64   `json:"confidence"`
}
