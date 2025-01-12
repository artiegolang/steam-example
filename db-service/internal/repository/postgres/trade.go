package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"petProject/db-service/pkg/logger"
	"petProject/proto/trading"
	"time"
)

type TradeRepository struct {
	pool *pgxpool.Pool
}

func NewTradeRepository(pool *pgxpool.Pool) *TradeRepository {
	return &TradeRepository{
		pool: pool,
	}
}

func (r *TradeRepository) CreateTrade(ctx context.Context, req *trading.CreateTradeRequest) error {
	if req.TotalPrice <= 0 {
		logger.Error("invalid trade price",
			zap.Int64("price", req.TotalPrice))
		return fmt.Errorf("trade price cannot be negative or zero")
	}
	if req.TradeCreatorId == req.TradeReceiverId {
		logger.Error("invalid trade participants",
			zap.Int64("creator_id", req.TradeCreatorId),
			zap.Int64("receiver_id", req.TradeReceiverId))
		return fmt.Errorf("cannot create trade with yourself")
	}

	query := `
        INSERT INTO trades (
            trade_description, 
            trade_creator_id, 
            trade_receiver_id, 
            trade_status, 
            total_price, 
            trade_date, 
            updated_at
        )
        VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING trade_id`

	logger.Debug("creating new trade",
		zap.Int64("creator_id", req.TradeCreatorId),
		zap.Int64("receiver_id", req.TradeReceiverId),
		zap.String("description", req.TradeDescription))

	var tradeID int64
	err := r.pool.QueryRow(ctx, query,
		req.TradeDescription,
		req.TradeCreatorId,
		req.TradeReceiverId,
		trading.Trade_PENDING,
		req.TotalPrice,
	).Scan(&tradeID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // Unique violation
				logger.Error("trade already exists",
					zap.Int64("creator_id", req.TradeCreatorId),
					zap.Int64("receiver_id", req.TradeReceiverId))
				return fmt.Errorf("trade already exists between these users")
			}
		}
		logger.Error("failed to create trade",
			zap.Error(err))
		return fmt.Errorf("failed to create trade: %w", err)
	}

	logger.Info("trade successfully created",
		zap.Int64("trade_id", tradeID),
		zap.Int64("creator_id", req.TradeCreatorId),
		zap.Int64("receiver_id", req.TradeReceiverId))
	return nil
}

func (r *TradeRepository) GetTrade(ctx context.Context, tradeID int64) (*trading.Trade, error) {
	query := `
       SELECT 
           trade_id,
           trade_description,
           trade_creator_id,
           trade_receiver_id,
           trade_status,
           total_price,
           trade_date,
           updated_at
       FROM trades 
       WHERE trade_id = $1`

	logger.Debug("fetching trade by ID", zap.Int64("trade_id", tradeID))

	var trade trading.Trade
	var tradeDate, updatedAt time.Time

	err := r.pool.QueryRow(ctx, query, tradeID).Scan(
		&trade.TradeId,
		&trade.TradeDescription,
		&trade.TradeCreatorId,
		&trade.TradeReceiverId,
		&trade.TradeStatus,
		&trade.TotalPrice,
		&tradeDate,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error("trade not found", zap.Int64("trade_id", tradeID))
			return nil, fmt.Errorf("trade not found with id: %d", tradeID)
		}
		logger.Error("failed to fetch trade",
			zap.Int64("trade_id", tradeID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch trade: %w", err)
	}

	logger.Info("trade successfully fetched",
		zap.Int64("trade_id", tradeID),
		zap.Int64("creator_id", trade.TradeCreatorId),
		zap.Int64("receiver_id", trade.TradeReceiverId))

	return &trade, nil
}

func (r *TradeRepository) DeleteTrade(ctx context.Context, tradeID int64) error {
	query := `
        DELETE FROM trades 
        WHERE trade_id = $1 
        AND trade_status = 'PENDING'
        RETURNING trade_id`

	logger.Debug("deleting trade", zap.Int64("trade_id", tradeID))

	var deletedID int64
	err := r.pool.QueryRow(ctx, query, tradeID).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error("trade not found or not in PENDING status",
				zap.Int64("trade_id", tradeID))
			return fmt.Errorf("trade not found or cannot be deleted: %d", tradeID)
		}
		logger.Error("failed to delete trade",
			zap.Int64("trade_id", tradeID),
			zap.Error(err))
		return fmt.Errorf("failed to delete trade: %w", err)
	}

	logger.Info("trade successfully deleted", zap.Int64("trade_id", deletedID))
	return nil
}

func (r *TradeRepository) AcceptTrade(ctx context.Context, req *trading.AcceptTradeRequest) error {
	query := `
        UPDATE trades 
        SET 
            trade_status = 'ACCEPTED',
            total_price = $1,
            updated_at = CURRENT_TIMESTAMP
        WHERE trade_id = $2 
        AND trade_receiver_id = $3
        AND trade_status = 'PENDING'
        RETURNING trade_id`

	logger.Debug("accepting trade",
		zap.Int64("trade_id", req.TradeId),
		zap.Int64("receiver_id", req.TradeReceiverId),
		zap.Int64("confirmed_price", req.ConfirmedTotalPrice))

	var updatedID int64
	err := r.pool.QueryRow(ctx, query,
		req.ConfirmedTotalPrice,
		req.TradeId,
		req.TradeReceiverId,
	).Scan(&updatedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error("trade not found or cannot be accepted",
				zap.Int64("trade_id", req.TradeId),
				zap.Int64("receiver_id", req.TradeReceiverId))
			return fmt.Errorf("trade not found or cannot be accepted: %d", req.TradeId)
		}
		logger.Error("failed to accept trade",
			zap.Int64("trade_id", req.TradeId),
			zap.Error(err))
		return fmt.Errorf("failed to accept trade: %w", err)
	}

	logger.Info("trade successfully accepted",
		zap.Int64("trade_id", updatedID),
		zap.Int64("receiver_id", req.TradeReceiverId))
	return nil
}

func (r *TradeRepository) DeclineTrade(ctx context.Context, req *trading.DeclineTradeRequest) error {
	query := `
       UPDATE trades 
       SET 
           trade_status = 'DECLINED',
           updated_at = CURRENT_TIMESTAMP
       WHERE trade_id = $1 
       AND trade_receiver_id = $2
       AND trade_status = 'PENDING'
       RETURNING trade_id`

	logger.Debug("declining trade",
		zap.Int64("trade_id", req.TradeId),
		zap.Int64("receiver_id", req.TradeReceiverId))

	var updatedID int64
	err := r.pool.QueryRow(ctx, query,
		req.TradeId,
		req.TradeReceiverId,
	).Scan(&updatedID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error("trade not found or cannot be declined",
				zap.Int64("trade_id", req.TradeId),
				zap.Int64("receiver_id", req.TradeReceiverId))
			return fmt.Errorf("trade not found or cannot be declined: %d", req.TradeId)
		}
		logger.Error("failed to decline trade",
			zap.Int64("trade_id", req.TradeId),
			zap.Error(err))
		return fmt.Errorf("failed to decline trade: %w", err)
	}

	logger.Info("trade successfully declined",
		zap.Int64("trade_id", updatedID),
		zap.Int64("receiver_id", req.TradeReceiverId))
	return nil
}

func (r *TradeRepository) GetAllTrades(ctx context.Context, req *trading.GetAllTradesRequest) ([]*trading.Trade, int32, error) {
	// First get total count for pagination
	countQuery := `
       SELECT COUNT(*) 
       FROM trades
       WHERE trade_creator_id = $1 OR trade_receiver_id = $1`

	var totalCount int32
	err := r.pool.QueryRow(ctx, countQuery, req.UserId).Scan(&totalCount)
	if err != nil {
		logger.Error("failed to get total trades count",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get total trades count: %w", err)
	}

	// Then get paginated data
	query := `
       SELECT 
           trade_id,
           trade_description,
           trade_creator_id,
           trade_receiver_id,
           trade_status,
           total_price,
           trade_date,
           updated_at
       FROM trades 
       WHERE trade_creator_id = $1 OR trade_receiver_id = $1
       ORDER BY trade_date DESC
       LIMIT $2 OFFSET $3`

	offset := (req.PageNumber - 1) * req.PageSize

	logger.Debug("fetching all trades",
		zap.Int64("user_id", req.UserId),
		zap.Int32("page_size", req.PageSize),
		zap.Int32("page_number", req.PageNumber))

	rows, err := r.pool.Query(ctx, query,
		req.UserId,
		req.PageSize,
		offset)
	if err != nil {
		logger.Error("failed to fetch trades",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to fetch trades: %w", err)
	}
	defer rows.Close()

	var trades []*trading.Trade
	for rows.Next() {
		var trade trading.Trade
		var tradeDate, updatedAt time.Time

		err := rows.Scan(
			&trade.TradeId,
			&trade.TradeDescription,
			&trade.TradeCreatorId,
			&trade.TradeReceiverId,
			&trade.TradeStatus,
			&trade.TotalPrice,
			&tradeDate,
			&updatedAt,
		)
		if err != nil {
			logger.Error("failed to scan trade", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, &trade)
	}

	if err = rows.Err(); err != nil {
		logger.Error("error iterating over trades", zap.Error(err))
		return nil, 0, fmt.Errorf("error iterating over trades: %w", err)
	}

	logger.Info("trades successfully fetched",
		zap.Int64("user_id", req.UserId),
		zap.Int("trades_count", len(trades)),
		zap.Int32("total_count", totalCount))

	return trades, totalCount, nil
}

func (r *TradeRepository) GetHistoryOfTrades(ctx context.Context, req *trading.GetHistoryOfTradesRequest) ([]*trading.Trade, int32, error) {
	// Get total count for completed trades
	countQuery := `
       SELECT COUNT(*) 
       FROM trades
       WHERE (trade_creator_id = $1 OR trade_receiver_id = $1)
       AND trade_status IN ('COMPLETED', 'CANCELLED', 'DECLINED')`

	var totalCount int32
	err := r.pool.QueryRow(ctx, countQuery, req.UserId).Scan(&totalCount)
	if err != nil {
		logger.Error("failed to get total trades history count",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get total trades history count: %w", err)
	}

	// Get paginated history data
	query := `
       SELECT 
           trade_id,
           trade_description,
           trade_creator_id,
           trade_receiver_id,
           trade_status,
           total_price,
           trade_date,
           updated_at
       FROM trades 
       WHERE (trade_creator_id = $1 OR trade_receiver_id = $1)
       AND trade_status IN ('COMPLETED', 'CANCELLED', 'DECLINED')
       ORDER BY trade_date DESC
       LIMIT $2 OFFSET $3`

	offset := (req.PageNumber - 1) * req.PageSize

	logger.Debug("fetching trade history",
		zap.Int64("user_id", req.UserId),
		zap.Int32("page_size", req.PageSize),
		zap.Int32("page_number", req.PageNumber))

	rows, err := r.pool.Query(ctx, query,
		req.UserId,
		req.PageSize,
		offset)
	if err != nil {
		logger.Error("failed to fetch trade history",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to fetch trade history: %w", err)
	}
	defer rows.Close()

	var trades []*trading.Trade
	for rows.Next() {
		var trade trading.Trade
		var tradeDate, updatedAt time.Time

		err := rows.Scan(
			&trade.TradeId,
			&trade.TradeDescription,
			&trade.TradeCreatorId,
			&trade.TradeReceiverId,
			&trade.TradeStatus,
			&trade.TotalPrice,
			&tradeDate,
			&updatedAt,
		)
		if err != nil {
			logger.Error("failed to scan trade history", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan trade history: %w", err)
		}
		trades = append(trades, &trade)
	}

	if err = rows.Err(); err != nil {
		logger.Error("error iterating over trade history", zap.Error(err))
		return nil, 0, fmt.Errorf("error iterating over trade history: %w", err)
	}

	logger.Info("trade history successfully fetched",
		zap.Int64("user_id", req.UserId),
		zap.Int("trades_count", len(trades)),
		zap.Int32("total_count", totalCount))

	return trades, totalCount, nil
}

func (r *TradeRepository) GetTradesByStatus(ctx context.Context, req *trading.GetTradesByStatusRequest) ([]*trading.Trade, int32, error) {
	countQuery := `
       SELECT COUNT(*) 
       FROM trades
       WHERE trade_status = $1`

	var totalCount int32
	err := r.pool.QueryRow(ctx, countQuery, req.Status).Scan(&totalCount)
	if err != nil {
		logger.Error("failed to get total trades count by status",
			zap.String("status", req.Status.String()),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get total trades count by status: %w", err)
	}

	query := `
       SELECT 
           trade_id,
           trade_description,
           trade_creator_id,
           trade_receiver_id,
           trade_status,
           total_price,
           trade_date,
           updated_at
       FROM trades 
       WHERE trade_status = $1
       ORDER BY trade_date DESC
       LIMIT $2 OFFSET $3`

	offset := (req.PageNumber - 1) * req.PageSize

	logger.Debug("fetching trades by status",
		zap.String("status", req.Status.String()),
		zap.Int32("page_size", req.PageSize),
		zap.Int32("page_number", req.PageNumber))

	rows, err := r.pool.Query(ctx, query,
		req.Status,
		req.PageSize,
		offset)
	if err != nil {
		logger.Error("failed to fetch trades by status",
			zap.String("status", req.Status.String()),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to fetch trades by status: %w", err)
	}
	defer rows.Close()

	var trades []*trading.Trade
	for rows.Next() {
		var trade trading.Trade
		var tradeDate, updatedAt time.Time

		err := rows.Scan(
			&trade.TradeId,
			&trade.TradeDescription,
			&trade.TradeCreatorId,
			&trade.TradeReceiverId,
			&trade.TradeStatus,
			&trade.TotalPrice,
			&tradeDate,
			&updatedAt,
		)
		if err != nil {
			logger.Error("failed to scan trade", zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, &trade)
	}

	if err = rows.Err(); err != nil {
		logger.Error("error iterating over trades", zap.Error(err))
		return nil, 0, fmt.Errorf("error iterating over trades: %w", err)
	}

	logger.Info("trades by status successfully fetched",
		zap.String("status", req.Status.String()),
		zap.Int("trades_count", len(trades)),
		zap.Int32("total_count", totalCount))

	return trades, totalCount, nil
}
