package interfaces

import (
	"context"
	"petProject/proto/trading"
)

type TradeUseCase interface {
	CreateTrade(ctx context.Context, req *trading.CreateTradeRequest) error
	GetTrade(ctx context.Context, tradeID int64) (*trading.Trade, error)
	DeleteTrade(ctx context.Context, tradeID int64) error

	// Операции со статусом сделки
	AcceptTrade(ctx context.Context, req *trading.AcceptTradeRequest) error
	DeclineTrade(ctx context.Context, req *trading.DeclineTradeRequest) error

	// Операции получения списка сделок
	GetAllTrades(ctx context.Context, req *trading.GetAllTradesRequest) ([]*trading.Trade, int32, error)
	GetHistoryOfTrades(ctx context.Context, req *trading.GetHistoryOfTradesRequest) ([]*trading.Trade, int32, error)
	GetTradesByStatus(ctx context.Context, req *trading.GetTradesByStatusRequest) ([]*trading.Trade, int32, error)
}
