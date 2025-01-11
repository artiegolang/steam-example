package interfaces

import (
	"context"
	"petProject/proto/inventory"
)

type ItemUseCase interface {
	// Основные операции с предметами
	CreateItem(ctx context.Context, req *inventory.CreateItemRequest) error
	GetItemByID(ctx context.Context, itemID int64) (*inventory.Item, error)
	GetItemByName(ctx context.Context, itemName string) (*inventory.Item, error)
	UpdateItemPrice(ctx context.Context, req *inventory.UpdatePriceRequest) error
	DeleteItem(ctx context.Context, req *inventory.DeleteItemRequest) error

	// Операции поиска и фильтрации предметов
	GetItemsByRarity(ctx context.Context, req *inventory.GetItemByRarityRequest) ([]*inventory.Item, error)
	GetItemsByPriceRange(ctx context.Context, req *inventory.GetItemsByPriceRequest) ([]*inventory.Item, error)
}

// InventoryUseCase определяет бизнес-логику для работы с инвентарем пользователей
type InventoryUseCase interface {
	// Управление инвентарем пользователя
	GetUserInventory(ctx context.Context, req *inventory.GetAllInventoryInfoRequest) (*inventory.InventoryInfo, error)

	// Операции с предметами в инвентаре
	AddItemToInventory(ctx context.Context, req *inventory.UpdateCountOfItemRequest) error
	RemoveItemFromInventory(ctx context.Context, req *inventory.UpdateCountOfItemRequest) error
	UpdateItemCount(ctx context.Context, req *inventory.UpdateCountOfItemRequest) error

	// Получение списка предметов с пагинацией
	GetAllUserItems(ctx context.Context, req *inventory.GetAllItemsRequest) ([]*inventory.Item, int32, error)
}
