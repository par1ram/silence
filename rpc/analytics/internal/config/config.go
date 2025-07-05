package config

import (
	"time"

	"github.com/par1ram/silence/shared/config"
)

type Config struct {
	HTTP     HTTPConfig     `mapstructure:"http"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	InfluxDB InfluxDBConfig `mapstructure:"influxdb"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Metrics  MetricsConfig  `mapstructure:"metrics"`
	Log      LogConfig      `mapstructure:"log"`
}

type HTTPConfig struct {
	Port         string        `mapstructure:"port" default:":8080"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" default:"30s"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" default:"30s"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" default:"60s"`
}

type GRPCConfig struct {
	Port string `mapstructure:"port" default:":9090"`
}

type InfluxDBConfig struct {
	URL      string `mapstructure:"url" default:"http://localhost:8086"`
	Token    string `mapstructure:"token"`
	Org      string `mapstructure:"org" default:"silence"`
	Bucket   string `mapstructure:"bucket" default:"analytics"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type RedisConfig struct {
	URL      string `mapstructure:"url" default:"localhost:6379"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" default:"0"`
}

type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled" default:"true"`
	Port    string `mapstructure:"port" default:":9091"`
}

type LogConfig struct {
	Level string `mapstructure:"level" default:"info"`
	JSON  bool   `mapstructure:"json" default:"false"`
}

func Load() (*Config, error) {
	cfg := &Config{}

	if err := config.Load("analytics", cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
