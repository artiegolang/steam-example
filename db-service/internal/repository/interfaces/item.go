package interfaces

import (
	"context"
	"petProject/proto/inventory"
)

// ItemRepository отвечает за операции с предметами
type ItemRepository interface {
	// Базовые CRUD операции с предметами
	CreateItem(ctx context.Context, req *inventory.CreateItemRequest) error
	GetItemByID(ctx context.Context, itemID int64) (*inventory.Item, error)
	GetItemByName(ctx context.Context, itemName string) (*inventory.Item, error)
	UpdateItemPrice(ctx context.Context, req *inventory.UpdatePriceRequest) error
	DeleteItem(ctx context.Context, req *inventory.DeleteItemRequest) error

	// Специальные операции поиска предметов
	GetItemsByRarity(ctx context.Context, req *inventory.GetItemByRarityRequest) ([]*inventory.Item, error)
	GetItemsByPriceRange(ctx context.Context, req *inventory.GetItemsByPriceRequest) ([]*inventory.Item, error)
}

// InventoryRepository отвечает за операции с инвентарем пользователей
type InventoryRepository interface {
	// Операции с инвентарем конкретного пользователя
	GetUserInventory(ctx context.Context, req *inventory.GetAllInventoryInfoRequest) (*inventory.InventoryInfo, error)
	UpdateItemCount(ctx context.Context, req *inventory.UpdateCountOfItemRequest) error
	GetAllUserItems(ctx context.Context, req *inventory.GetAllItemsRequest) ([]*inventory.Item, int32, error)
}
