package grpc

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"petProject/db-service/internal/usecase/interfaces"
	"petProject/db-service/pkg/logger"
	"petProject/proto/inventory"
)

type ItemService struct {
	inventory.UnimplementedInventoryServer
	useCaseItem      interfaces.ItemUseCase
	useCaseInventory interfaces.InventoryUseCase
}

func NewItemService(useCaseItem interfaces.ItemUseCase, useCaseInventory interfaces.InventoryUseCase) *ItemService {
	return &ItemService{
		useCaseItem:      useCaseItem,
		useCaseInventory: useCaseInventory,
	}
}

// CreateItem — создание нового предмета
func (s *ItemService) CreateItem(ctx context.Context, req *inventory.CreateItemRequest) (*inventory.CreateItemResponse, error) {
	err := s.useCaseItem.CreateItem(ctx, req)
	if err != nil {
		logger.Error("error while creating item", zap.Int64("item_id", req.ItemId), zap.Error(err))
		return nil, fmt.Errorf("error while creating item: %w", err)
	}
	return &inventory.CreateItemResponse{
		Success: true,
		Message: "Item created",
	}, nil
}

// UpdatePrice — обновление цены предмета
func (s *ItemService) UpdatePrice(ctx context.Context, req *inventory.UpdatePriceRequest) (*inventory.UpdatePriceResponse, error) {
	err := s.useCaseItem.UpdateItemPrice(ctx, req)
	if err != nil {
		logger.Error("error while updating item price", zap.Int64("item_id", req.ItemId), zap.Error(err))
		return nil, fmt.Errorf("error while updating item price: %w", err)
	}
	return &inventory.UpdatePriceResponse{}, nil
}

// DeleteItem — удаление предмета
func (s *ItemService) DeleteItem(ctx context.Context, req *inventory.DeleteItemRequest) (*inventory.DeleteItemResponse, error) {
	err := s.useCaseItem.DeleteItem(ctx, req)
	if err != nil {
		logger.Error("error while deleting item", zap.Int64("item_id", req.ItemId), zap.Error(err))
		return nil, fmt.Errorf("error while deleting item: %w", err)
	}
	return &inventory.DeleteItemResponse{}, nil
}

// GetItemPrice — получить цену предмета по ID или имени
func (s *ItemService) GetItemPrice(ctx context.Context, req *inventory.GetItemPriceRequest) (*inventory.GetItemPriceResponse, error) {
	if req.ItemName == "" {
		logger.Error("invalid request: item_name is empty")
		return nil, fmt.Errorf("invalid request: item_name cannot be empty")
	}

	item, err := s.useCaseItem.GetItemByName(ctx, req.ItemName)
	if err != nil {
		logger.Error("error while getting item by name", zap.String("item_name", req.ItemName), zap.Error(err))
		return nil, fmt.Errorf("error while getting item by name: %w", err)
	}

	return &inventory.GetItemPriceResponse{
		ItemPrice: item.ItemPrice,
	}, nil
}

// GetItemByRarity — получить все предметы по конкретной редкости
func (s *ItemService) GetItemByRarity(ctx context.Context, req *inventory.GetItemByRarityRequest) (*inventory.GetItemByRarityResponse, error) {
	items, err := s.useCaseItem.GetItemsByRarity(ctx, req)
	if err != nil {
		logger.Error("error while getting items by rarity", zap.String("rarity", req.Rarity), zap.Error(err))
		return nil, fmt.Errorf("error while getting items by rarity: %w", err)
	}
	return &inventory.GetItemByRarityResponse{
		Items: items,
	}, nil
}

// GetItemsByPrice — получить предметы в заданном диапазоне цен
func (s *ItemService) GetItemsByPrice(ctx context.Context, req *inventory.GetItemsByPriceRequest) (*inventory.GetItemsByPriceResponse, error) {
	items, err := s.useCaseItem.GetItemsByPriceRange(ctx, req)
	if err != nil {
		logger.Error("error while getting items by price", zap.Int64("start_price", req.StartingPrice), zap.Int64("max_price", req.HighestPrice), zap.Error(err))
		return nil, fmt.Errorf("error while getting items by price: %w", err)
	}
	return &inventory.GetItemsByPriceResponse{
		Items: items,
	}, nil
}

// GetAllInventoryInfo — получить полную информацию об инвентаре пользователя
func (s *ItemService) GetAllInventoryInfo(ctx context.Context, req *inventory.GetAllInventoryInfoRequest) (*inventory.GetAllInventoryInfoResponse, error) {
	inv, err := s.useCaseInventory.GetUserInventory(ctx, req)
	if err != nil {
		logger.Error("error while getting user inventory", zap.Int64("user_id", req.UserId), zap.Error(err))
		return nil, fmt.Errorf("error while getting user inventory: %w", err)
	}
	return &inventory.GetAllInventoryInfoResponse{
		UserId:       inv.UserId,
		Username:     inv.Username,
		Items:        inv.Items,
		OverallPrice: inv.OverallPrice,
	}, nil
}

// GetAllItems — получить список всех предметов пользователя (с пагинацией)
func (s *ItemService) GetAllItems(ctx context.Context, req *inventory.GetAllItemsRequest) (*inventory.GetAllItemsResponse, error) {
	items, total, err := s.useCaseInventory.GetAllUserItems(ctx, req)
	if err != nil {
		logger.Error("error while getting all user items", zap.Int64("user_id", req.UserId), zap.Error(err))
		return nil, fmt.Errorf("error while getting all user items: %w", err)
	}
	return &inventory.GetAllItemsResponse{
		Items:      items,
		TotalItems: total,
	}, nil
}

// UpdateCountOfItem — изменить количество конкретного предмета в инвентаре
func (s *ItemService) UpdateCountOfItem(ctx context.Context, req *inventory.UpdateCountOfItemRequest) (*inventory.UpdateCountOfItemResponse, error) {
	err := s.useCaseInventory.UpdateItemCount(ctx, req)
	if err != nil {
		logger.Error("error while updating item count", zap.Int64("item_id", req.ItemId), zap.Error(err))
		return nil, fmt.Errorf("error while updating item count: %w", err)
	}
	return &inventory.UpdateCountOfItemResponse{}, nil
}
