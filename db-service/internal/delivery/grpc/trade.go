package grpc

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"petProject/db-service/internal/usecase/interfaces"
	"petProject/db-service/pkg/logger"
	"petProject/proto/trading"
)

type TradingService struct {
	trading.UnimplementedTradingServer
	useCaseTrade interfaces.TradeUseCase
}

func NewTradingService(useCaseTrade interfaces.TradeUseCase) *TradingService {
	return &TradingService{
		useCaseTrade: useCaseTrade,
	}
}

func (s *TradingService) CreateTrade(ctx context.Context, req *trading.CreateTradeRequest) (*trading.CreateTradeResponse, error) {
	err := s.useCaseTrade.CreateTrade(ctx, req)
	if err != nil {
		logger.Error("failed to create trade",
			zap.Int64("creator_id", req.TradeCreatorId),
			zap.Int64("receiver_id", req.TradeReceiverId),
			zap.Error(err))
		return &trading.CreateTradeResponse{
			Success: false,
			Message: "Failed to create trade",
		}, fmt.Errorf("failed to create trade: %w", err)
	}

	logger.Info("trade created successfully",
		zap.Int64("creator_id", req.TradeCreatorId),
		zap.Int64("receiver_id", req.TradeReceiverId))
	return &trading.CreateTradeResponse{
		Success: true,
		Message: "Trade created successfully",
	}, nil
}

func (s *TradingService) AcceptTrade(ctx context.Context, req *trading.AcceptTradeRequest) (*trading.AcceptTradeResponse, error) {
	err := s.useCaseTrade.AcceptTrade(ctx, req)
	if err != nil {
		logger.Error("failed to accept trade",
			zap.Int64("trade_id", req.TradeId),
			zap.Int64("receiver_id", req.TradeReceiverId),
			zap.Error(err))
		return &trading.AcceptTradeResponse{
			Success: false,
			Message: "Failed to accept trade",
		}, fmt.Errorf("failed to accept trade: %w", err)
	}

	logger.Info("trade accepted successfully",
		zap.Int64("trade_id", req.TradeId),
		zap.Int64("receiver_id", req.TradeReceiverId))
	return &trading.AcceptTradeResponse{
		Success: true,
		Message: "Trade accepted successfully",
	}, nil
}

func (s *TradingService) DeclineTrade(ctx context.Context, req *trading.DeclineTradeRequest) (*trading.DeclineTradeResponse, error) {
	err := s.useCaseTrade.DeclineTrade(ctx, req)
	if err != nil {
		logger.Error("failed to decline trade",
			zap.Int64("trade_id", req.TradeId),
			zap.Int64("receiver_id", req.TradeReceiverId),
			zap.Error(err))
		return &trading.DeclineTradeResponse{
			Success: false,
			Message: "Failed to decline trade",
		}, fmt.Errorf("failed to decline trade: %w", err)
	}

	logger.Info("trade declined successfully",
		zap.Int64("trade_id", req.TradeId),
		zap.Int64("receiver_id", req.TradeReceiverId))
	return &trading.DeclineTradeResponse{
		Success: true,
		Message: "Trade declined successfully",
	}, nil
}

func (s *TradingService) GetTrade(ctx context.Context, req *trading.GetTradeRequest) (*trading.GetTradeResponse, error) {
	trade, err := s.useCaseTrade.GetTrade(ctx, req.TradeId)
	if err != nil {
		logger.Error("failed to get trade",
			zap.Int64("trade_id", req.TradeId),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get trade: %w", err)
	}

	logger.Info("trade retrieved successfully",
		zap.Int64("trade_id", req.TradeId))
	return &trading.GetTradeResponse{
		Trade: trade,
	}, nil
}

func (s *TradingService) GetAllTrades(ctx context.Context, req *trading.GetAllTradesRequest) (*trading.GetAllTradesResponse, error) {
	trades, total, err := s.useCaseTrade.GetAllTrades(ctx, req)
	if err != nil {
		logger.Error("failed to get all trades",
			zap.Int64("user_id", req.UserId),
			zap.Int32("page_size", req.PageSize),
			zap.Int32("page_number", req.PageNumber),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get all trades: %w", err)
	}

	logger.Info("trades retrieved successfully",
		zap.Int64("user_id", req.UserId),
		zap.Int("trades_count", len(trades)),
		zap.Int32("total_items", total))
	return &trading.GetAllTradesResponse{
		Trades:     trades,
		PageSize:   req.PageSize,
		PageNumber: req.PageNumber,
		TotalItems: total,
	}, nil
}

func (s *TradingService) GetHistoryOfTrades(ctx context.Context, req *trading.GetHistoryOfTradesRequest) (*trading.GetHistoryOfTradesResponse, error) {
	trades, total, err := s.useCaseTrade.GetHistoryOfTrades(ctx, req)
	if err != nil {
		logger.Error("failed to get trade history",
			zap.Int64("user_id", req.UserId),
			zap.Int32("page_size", req.PageSize),
			zap.Int32("page_number", req.PageNumber),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get trade history: %w", err)
	}

	logger.Info("trade history retrieved successfully",
		zap.Int64("user_id", req.UserId),
		zap.Int("trades_count", len(trades)),
		zap.Int32("total_items", total))
	return &trading.GetHistoryOfTradesResponse{
		Trades:     trades,
		PageSize:   req.PageSize,
		PageNumber: req.PageNumber,
		TotalItems: total,
	}, nil
}

func (s *TradingService) GetTradesByStatus(ctx context.Context, req *trading.GetTradesByStatusRequest) (*trading.GetTradesByStatusResponse, error) {
	trades, total, err := s.useCaseTrade.GetTradesByStatus(ctx, req)
	if err != nil {
		logger.Error("failed to get trades by status",
			zap.String("status", req.Status.String()),
			zap.Int32("page_size", req.PageSize),
			zap.Int32("page_number", req.PageNumber),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get trades by status: %w", err)
	}

	logger.Info("trades by status retrieved successfully",
		zap.String("status", req.Status.String()),
		zap.Int("trades_count", len(trades)),
		zap.Int32("total_items", total))
	return &trading.GetTradesByStatusResponse{
		Trades:     trades,
		PageSize:   req.PageSize,
		PageNumber: req.PageNumber,
		TotalItems: total,
	}, nil
}

func (s *TradingService) DeleteTrade(ctx context.Context, req *trading.DeleteTradeRequest) (*trading.DeleteTradeResponse, error) {
	err := s.useCaseTrade.DeleteTrade(ctx, req.TradeId)
	if err != nil {
		logger.Error("failed to delete trade",
			zap.Int64("trade_id", req.TradeId),
			zap.Error(err))
		return &trading.DeleteTradeResponse{
			Success: false,
			Message: "Failed to delete trade",
		}, fmt.Errorf("failed to delete trade: %w", err)
	}

	logger.Info("trade deleted successfully",
		zap.Int64("trade_id", req.TradeId))
	return &trading.DeleteTradeResponse{
		Success: true,
		Message: "Trade deleted successfully",
	}, nil
}
