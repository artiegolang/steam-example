package domain

import "time"

type Trade struct {
	TradeID          int64       `json:"trade_id"`
	TradeDescription string      `json:"trade_description"`
	CreatorID        int64       `json:"creator_id"`
	ReceiverID       int64       `json:"receiver_id"`
	Status           TradeStatus `json:"status"`
	TotalPrice       int64       `json:"total_price"`
	TradeDate        time.Time   `json:"trade_date"`
	UpdatedAt        time.Time   `json:"updated_at"`
	Items            []Item      `json:"items"`
}

type TradeStatus string

const (
	TradeStatusPending   TradeStatus = "PENDING"
	TradeStatusAccepted  TradeStatus = "ACCEPTED"
	TradeStatusDeclined  TradeStatus = "DECLINED"
	TradeStatusCancelled TradeStatus = "CANCELLED"
	TradeStatusCompleted TradeStatus = "COMPLETED"
)
