package cmd

import (
	"context"
	"go.uber.org/zap"
	"vizurth/wildberries-test-l0/order-back-end/internal/config"
	"vizurth/wildberries-test-l0/order-back-end/internal/logger"
	"vizurth/wildberries-test-l0/order-back-end/internal/postgres"
)

func main() {
	ctx := context.Background()

	cfg, err := config.NewConfig()

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "config.New error", zap.Error(err))
	}

	db, err := postgres.New(ctx, cfg.Postgres)

	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "postgres.New error", zap.Error(err))
	}
	db.Ping(ctx)
	defer db.Close()

}
