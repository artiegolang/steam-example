package impl

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"petProject/db-service/internal/repository/interfaces"
	"petProject/db-service/pkg/logger"
	"petProject/proto/trading"
)

type TradeUseCase struct {
	repo interfaces.TradeRepository
}

func NewTradeUseCase(repo interfaces.TradeRepository) *TradeUseCase {
	return &TradeUseCase{repo: repo}
}

func (i *TradeUseCase) CreateTrade(ctx context.Context, req *trading.CreateTradeRequest) error {
	// Validate basic params
	if req.TradeCreatorId <= 0 || req.TradeReceiverId <= 0 {
		logger.Error("invalid user IDs in trade request",
			zap.Int64("creator_id", req.TradeCreatorId),
			zap.Int64("receiver_id", req.TradeReceiverId))
		return fmt.Errorf("invalid user IDs")
	}

	if req.TotalPrice < 0 {
		logger.Error("negative price in trade request",
			zap.Int64("price", req.TotalPrice))
		return fmt.Errorf("trade price cannot be negative")
	}

	if len(req.Items) == 0 {
		logger.Error("attempt to create empty trade")
		return fmt.Errorf("trade must contain at least one item")
	}

	// Создаем сделку
	logger.Debug("creating new trade",
		zap.Int64("creator_id", req.TradeCreatorId),
		zap.Int64("receiver_id", req.TradeReceiverId))

	if err := i.repo.CreateTrade(ctx, req); err != nil {
		logger.Error("failed to create trade", zap.Error(err))
		return fmt.Errorf("failed to create trade: %w", err)
	}

	logger.Info("trade successfully created",
		zap.Int64("creator_id", req.TradeCreatorId),
		zap.Int64("receiver_id", req.TradeReceiverId))
	return nil
}

func (i *TradeUseCase) GetTrade(ctx context.Context, tradeID int64) (*trading.Trade, error) {
	if tradeID <= 0 {
		logger.Error("invalid trade ID", zap.Int64("trade_id", tradeID))
		return nil, fmt.Errorf("invalid trade ID")
	}

	logger.Debug("getting trade details", zap.Int64("trade_id", tradeID))

	trade, err := i.repo.GetTrade(ctx, tradeID)
	if err != nil {
		logger.Error("failed to get trade",
			zap.Int64("trade_id", tradeID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get trade: %w", err)
	}

	logger.Info("trade details retrieved",
		zap.Int64("trade_id", tradeID),
		zap.Int64("creator_id", trade.TradeCreatorId),
		zap.Int64("receiver_id", trade.TradeReceiverId))
	return trade, nil
}

func (i *TradeUseCase) DeleteTrade(ctx context.Context, tradeID int64) error {
	if tradeID <= 0 {
		logger.Error("invalid trade ID", zap.Int64("trade_id", tradeID))
		return fmt.Errorf("invalid trade ID")
	}

	// Проверяем существование и статус сделки перед удалением
	trade, err := i.repo.GetTrade(ctx, tradeID)
	if err != nil {
		logger.Error("failed to get trade for deletion",
			zap.Int64("trade_id", tradeID),
			zap.Error(err))
		return fmt.Errorf("failed to get trade: %w", err)
	}

	// Проверяем, можно ли удалить сделку
	if trade.TradeStatus != trading.Trade_PENDING {
		logger.Error("cannot delete non-pending trade",
			zap.Int64("trade_id", tradeID),
			zap.String("status", trade.TradeStatus.String()))
		return fmt.Errorf("only pending trades can be deleted")
	}

	logger.Debug("deleting trade", zap.Int64("trade_id", tradeID))

	if err := i.repo.DeleteTrade(ctx, tradeID); err != nil {
		logger.Error("failed to delete trade",
			zap.Int64("trade_id", tradeID),
			zap.Error(err))
		return fmt.Errorf("failed to delete trade: %w", err)
	}

	logger.Info("trade successfully deleted", zap.Int64("trade_id", tradeID))
	return nil
}

func (i *TradeUseCase) AcceptTrade(ctx context.Context, req *trading.AcceptTradeRequest) error {
	// Базовая валидация
	if req.TradeId <= 0 {
		logger.Error("invalid trade ID", zap.Int64("trade_id", req.TradeId))
		return fmt.Errorf("invalid trade ID")
	}
	if req.ConfirmedTotalPrice < 0 {
		logger.Error("negative price confirmation",
			zap.Int64("price", req.ConfirmedTotalPrice))
		return fmt.Errorf("confirmed price cannot be negative")
	}

	// Проверяем существование и статус сделки
	trade, err := i.repo.GetTrade(ctx, req.TradeId)
	if err != nil {
		logger.Error("failed to get trade for acceptance",
			zap.Int64("trade_id", req.TradeId),
			zap.Error(err))
		return fmt.Errorf("failed to get trade: %w", err)
	}

	// Проверяем, что подтверждает именно получатель
	if trade.TradeReceiverId != req.TradeReceiverId {
		logger.Error("unauthorized trade acceptance attempt",
			zap.Int64("trade_id", req.TradeId),
			zap.Int64("receiver_id", req.TradeReceiverId))
		return fmt.Errorf("only trade receiver can accept the trade")
	}

	// Проверяем статус сделки
	if trade.TradeStatus != trading.Trade_PENDING {
		logger.Error("cannot accept non-pending trade",
			zap.Int64("trade_id", req.TradeId),
			zap.String("status", trade.TradeStatus.String()))
		return fmt.Errorf("only pending trades can be accepted")
	}

	// Проверяем соответствие цены
	if trade.TotalPrice != req.ConfirmedTotalPrice {
		logger.Error("price mismatch in trade acceptance",
			zap.Int64("trade_id", req.TradeId),
			zap.Int64("original_price", trade.TotalPrice),
			zap.Int64("confirmed_price", req.ConfirmedTotalPrice))
		return fmt.Errorf("confirmed price doesn't match trade price")
	}

	logger.Debug("accepting trade",
		zap.Int64("trade_id", req.TradeId),
		zap.Int64("receiver_id", req.TradeReceiverId))

	if err := i.repo.AcceptTrade(ctx, req); err != nil {
		logger.Error("failed to accept trade",
			zap.Int64("trade_id", req.TradeId),
			zap.Error(err))
		return fmt.Errorf("failed to accept trade: %w", err)
	}

	logger.Info("trade successfully accepted",
		zap.Int64("trade_id", req.TradeId),
		zap.Int64("receiver_id", req.TradeReceiverId))
	return nil
}

func (i *TradeUseCase) DeclineTrade(ctx context.Context, req *trading.DeclineTradeRequest) error {
	// Базовая валидация
	if req.TradeId <= 0 {
		logger.Error("invalid trade ID", zap.Int64("trade_id", req.TradeId))
		return fmt.Errorf("invalid trade ID")
	}

	// Проверяем существование и статус сделки
	trade, err := i.repo.GetTrade(ctx, req.TradeId)
	if err != nil {
		logger.Error("failed to get trade for declining",
			zap.Int64("trade_id", req.TradeId),
			zap.Error(err))
		return fmt.Errorf("failed to get trade: %w", err)
	}

	// Проверяем, что отклоняет именно получатель
	if trade.TradeReceiverId != req.TradeReceiverId {
		logger.Error("unauthorized trade decline attempt",
			zap.Int64("trade_id", req.TradeId),
			zap.Int64("receiver_id", req.TradeReceiverId))
		return fmt.Errorf("only trade receiver can decline the trade")
	}

	// Проверяем статус сделки
	if trade.TradeStatus != trading.Trade_PENDING {
		logger.Error("cannot decline non-pending trade",
			zap.Int64("trade_id", req.TradeId),
			zap.String("status", trade.TradeStatus.String()))
		return fmt.Errorf("only pending trades can be declined")
	}

	logger.Debug("declining trade",
		zap.Int64("trade_id", req.TradeId),
		zap.Int64("receiver_id", req.TradeReceiverId))

	if err := i.repo.DeclineTrade(ctx, req); err != nil {
		logger.Error("failed to decline trade",
			zap.Int64("trade_id", req.TradeId),
			zap.Error(err))
		return fmt.Errorf("failed to decline trade: %w", err)
	}

	logger.Info("trade successfully declined",
		zap.Int64("trade_id", req.TradeId),
		zap.Int64("receiver_id", req.TradeReceiverId))
	return nil
}

func (i *TradeUseCase) GetAllTrades(ctx context.Context, req *trading.GetAllTradesRequest) ([]*trading.Trade, int32, error) {
	// Валидация параметров пагинации
	if req.PageSize <= 0 {
		req.PageSize = 10 // Default page size
	}
	if req.PageNumber <= 0 {
		req.PageNumber = 1 // Default page number
	}
	if req.UserId <= 0 {
		logger.Error("invalid user ID", zap.Int64("user_id", req.UserId))
		return nil, 0, fmt.Errorf("invalid user ID")
	}

	logger.Debug("getting all trades",
		zap.Int64("user_id", req.UserId),
		zap.Int32("page_size", req.PageSize),
		zap.Int32("page_number", req.PageNumber))

	trades, total, err := i.repo.GetAllTrades(ctx, req)
	if err != nil {
		logger.Error("failed to get trades",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get trades: %w", err)
	}

	logger.Info("trades successfully retrieved",
		zap.Int64("user_id", req.UserId),
		zap.Int("trades_count", len(trades)),
		zap.Int32("total_count", total))
	return trades, total, nil
}

func (i *TradeUseCase) GetHistoryOfTrades(ctx context.Context, req *trading.GetHistoryOfTradesRequest) ([]*trading.Trade, int32, error) {
	// Валидация параметров пагинации
	if req.PageSize <= 0 {
		req.PageSize = 10 // Default page size
	}
	if req.PageNumber <= 0 {
		req.PageNumber = 1 // Default page number
	}
	if req.UserId <= 0 {
		logger.Error("invalid user ID", zap.Int64("user_id", req.UserId))
		return nil, 0, fmt.Errorf("invalid user ID")
	}

	logger.Debug("getting trade history",
		zap.Int64("user_id", req.UserId),
		zap.Int32("page_size", req.PageSize),
		zap.Int32("page_number", req.PageNumber))

	trades, total, err := i.repo.GetHistoryOfTrades(ctx, req)
	if err != nil {
		logger.Error("failed to get trade history",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get trade history: %w", err)
	}

	logger.Info("trade history successfully retrieved",
		zap.Int64("user_id", req.UserId),
		zap.Int("trades_count", len(trades)),
		zap.Int32("total_count", total))
	return trades, total, nil
}

func (i *TradeUseCase) GetTradesByStatus(ctx context.Context, req *trading.GetTradesByStatusRequest) ([]*trading.Trade, int32, error) {
	// Валидация параметров пагинации
	if req.PageSize <= 0 {
		req.PageSize = 10 // Default page size
	}
	if req.PageNumber <= 0 {
		req.PageNumber = 1 // Default page number
	}

	// Валидация статуса
	if !isValidTradeStatus(req.Status) {
		logger.Error("invalid trade status", zap.String("status", req.Status.String()))
		return nil, 0, fmt.Errorf("invalid trade status")
	}

	logger.Debug("getting trades by status",
		zap.String("status", req.Status.String()),
		zap.Int32("page_size", req.PageSize),
		zap.Int32("page_number", req.PageNumber))

	trades, total, err := i.repo.GetTradesByStatus(ctx, req)
	if err != nil {
		logger.Error("failed to get trades by status",
			zap.String("status", req.Status.String()),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get trades by status: %w", err)
	}

	logger.Info("trades by status successfully retrieved",
		zap.String("status", req.Status.String()),
		zap.Int("trades_count", len(trades)),
		zap.Int32("total_count", total))
	return trades, total, nil
}

// Вспомогательная функция для валидации статуса
func isValidTradeStatus(status trading.Trade_TradeStatus) bool {
	switch status {
	case trading.Trade_PENDING,
		trading.Trade_ACCEPTED,
		trading.Trade_DECLINED,
		trading.Trade_CANCELLED,
		trading.Trade_COMPLETED:
		return true
	default:
		return false
	}
}
