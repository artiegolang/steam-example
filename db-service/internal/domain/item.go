package domain

import (
	"time"
)

type Item struct {
	ItemName   string     `json:"item_name"`
	ItemID     int64      `json:"item_id"`
	ItemPrice  int64      `json:"item_price"`
	ItemCount  int64      `json:"item_count"`
	ItemRarity ItemRarity `json:"item_rarity"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// InventoryItem представляет связь между пользователем и предметом
type InventoryItem struct {
	UserID    int64 `json:"user_id"`
	ItemID    int64 `json:"item_id"`
	ItemCount int64 `json:"item_count"`
}

// Pagination для обработки постраничного вывода
type Pagination struct {
	PageSize   int32 `json:"page_size"`
	PageNumber int32 `json:"page_number"`
}

type Inventory struct {
	UserName     string `json:"user_name"`
	UserID       int64  `json:"user_id"`
	Items        []Item `json:"items"`
	OverAllPrice int64  `json:"over_all_price"`
}

type ItemRarity string

const (
	RarityCommon    ItemRarity = "COMMON"
	RarityUncommon  ItemRarity = "UNCOMMON"
	RarityRare      ItemRarity = "RARE"
	RarityEpic      ItemRarity = "EPIC"
	RarityLegendary ItemRarity = "LEGENDARY"
)
