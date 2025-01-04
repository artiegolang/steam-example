package main

import (
	"context"
	"go.uber.org/zap"
	"petProject/db-service/config"
	"petProject/db-service/pkg/logger/zaplog"
	"petProject/db-service/pkg/postgres"
)

func main() {
	cfg := &config.Config{
		Log: config.LogConfig{
			Development: true,
			Level:       "debug",
		},
		Postgres: config.PostgresConfig{
			Host:     "localhost",
			Port:     "5433",
			User:     "postgres",
			Password: "postgres",
			DBName:   "trading_platform",
			MaxConns: 10,
			MinConns: 2,
		},
	}

	// Инициализируем логгер
	zaplog.InitLogger(cfg.Log)

	log := zaplog.GetLogger()
	log.Info("Starting DB service")

	// Создаем контекст
	ctx := context.Background()

	// Создаем подключение к базе данных
	postgres, err := postgres.NewPostgreSQL(ctx, &cfg.Postgres, log)
	if err != nil {
		log.Error("Failed to initialize database", zap.Error(err))
		return
	}
	defer postgres.Close()
}
