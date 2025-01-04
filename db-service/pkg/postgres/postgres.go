package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"petProject/db-service/config"
	"petProject/db-service/pkg/logger"
)

type PostgreSQL struct {
	log  logger.Logger
	pool *pgxpool.Pool
}

func NewPostgreSQL(ctx context.Context, cfg *config.PostgresConfig, log logger.Logger) (*PostgreSQL, error) {
	// Формируем строку подключения
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	// Создаем пул соединений
	pool, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to db: %w", err)
	}

	// Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("error pinging db: %w", err)
	}

	log.Info("Successfully connected to database",
		zap.String("host", cfg.Host),
		zap.String("database", cfg.DBName))

	return &PostgreSQL{
		pool: pool,
		log:  log,
	}, nil
}

func (p *PostgreSQL) Close() {
	if p.pool != nil {
		p.pool.Close()
		p.log.Info("Database connection closed")
	}
}
