package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Load загружает конфигурацию из файла
func Load(configName string, config interface{}) error {
	// Ищем конфигурационный файл в нескольких местах
	configPaths := []string{
		"./etc",
		"../etc",
		"../../etc",
		"../../../etc",
	}

	var configFile string
	for _, path := range configPaths {
		potentialFile := filepath.Join(path, configName+".yaml")
		if _, err := os.Stat(potentialFile); err == nil {
			configFile = potentialFile
			break
		}
	}

	if configFile == "" {
		return fmt.Errorf("config file %s.yaml not found in any of the expected paths", configName)
	}

	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	// Читаем переменные окружения
	viper.AutomaticEnv()

	// Читаем конфигурационный файл
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Десериализуем в структуру
	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// LoadFromEnv загружает конфигурацию из переменных окружения
func LoadFromEnv(config interface{}) error {
	viper.AutomaticEnv()

	if err := viper.Unmarshal(config); err != nil {
		return fmt.Errorf("failed to unmarshal config from env: %w", err)
	}

	return nil
}

// NormalizePort добавляет двоеточие к порту, если его нет
func NormalizePort(port string) string {
	if port == "" {
		return ":8080"
	}
	if !strings.HasPrefix(port, ":") {
		return ":" + port
	}
	return port
}
