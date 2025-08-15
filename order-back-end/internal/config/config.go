package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	kfk "order-back-end/internal/kafka/config"
	"order-back-end/internal/postgres"
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
func NewConfig() (Config, error) {
	var config Config
	if err := cleanenv.ReadConfig("./config/config.yaml", &config); err != nil {
		if err = cleanenv.ReadEnv(&config); err != nil {
			return Config{}, err
		}
	}
	return config, nil
}
