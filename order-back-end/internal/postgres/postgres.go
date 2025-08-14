package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string `yaml:"host" env:"PG_HOST" envDefault:"localhost"`
	Port     string `yaml:"port" env:"PG_PORT" envDefault:"5432"`
	User     string `yaml:"user" env:"PG_USER" envDefault:"postgres"`
	Password string `yaml:"password" env:"PG_PASSWORD" envDefault:"postgres"`
	Database string `yaml:"database" env:"PG_DB" envDefault:"postgres"`
}

func New(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	connString := cfg.ConnString()

	conn, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (c *Config) ConnString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
}
