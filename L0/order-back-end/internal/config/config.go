package config

import (
	"fmt"
	kfk "order-back-end/internal/kafka/config"
	"order-back-end/internal/postgres"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// httpConfig структура которая содержит порт для подключения по HTTP
type httpConfig struct {
	Port string `yaml:"port" envconfig:"PORT" default:"8081"`
}

// Config структура содержащая основные параменты в конфиге
type Config struct {
	HTTP     httpConfig      `yaml:"http" envconfig:"HTTP"`
	Postgres postgres.Config `yaml:"postgres" envconfig:"POSTGRES"`
	Kafka    kfk.Config      `yaml:"kafka" envconfig:"KAFKA"`
}

// NewConfig создает Config
func NewConfig() (*Config, error) {
	var cfg Config

	// Ищем config.yaml в нескольких местах для совместимости с тестами и Docker
	configPaths := []string{
		"./configs/config.yaml",
		"configs/config.yaml",
		"../configs/config.yaml",
		"../../configs/config.yaml",
	}

	var configPath string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configPath = path
			break
		}
	}

	if configPath == "" {
		return &Config{}, fmt.Errorf("config file not found")
	}

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return &Config{}, fmt.Errorf("error reading config: %w", err)
	}
	return &cfg, nil
}
