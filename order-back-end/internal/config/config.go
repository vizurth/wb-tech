package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"vizurth/wildberries-test-l0/order-back-end/internal/postgres"
)

type Config struct {
	Port     string          `yaml:"port" envconfig:"PORT" default:"8080"`
	Postgres postgres.Config `yaml:"postgres" envconfig:"POSTGRES"`
}

func NewConfig() (Config, error) {
	var config Config
	if err := cleanenv.ReadConfig("./config/config.yaml", &config); err != nil {
		if err = cleanenv.ReadEnv(&config); err != nil {
			return Config{}, err
		}
	}
	``
	return config, nil
}
