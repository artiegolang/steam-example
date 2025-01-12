package impl

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"go.uber.org/zap"
	"math"
	"petProject/db-service/internal/repository/interfaces"
	"petProject/db-service/pkg/logger"
	"petProject/proto/inventory"
)

type ItemUseCase struct {
	repo interfaces.ItemRepository
}

func NewItemUseCase(repo interfaces.ItemRepository) *ItemUseCase {
	return &ItemUseCase{repo: repo}
}

func (i *ItemUseCase) CreateItem(ctx context.Context, req *inventory.CreateItemRequest) error {
	if req.ItemName == "" {
		logger.Error("attempt to create item with empty name")
		return fmt.Errorf("item name cannot be empty")
	}
	if req.ItemPrice < 0 {
		logger.Error("attempt to create item with negative price",
			zap.Int64("price", req.ItemPrice))
		return fmt.Errorf("item price cannot be negative")
	}

	// Check if item already exists
	logger.Debug("checking if item exists",
		zap.String("item_name", req.ItemName))
	existingItem, err := i.repo.GetItemByName(ctx, req.ItemName)
	if err == nil && existingItem != nil {
		// If item is found - this is an error case
		logger.Error("item with this name already exists",
			zap.String("item_name", req.ItemName))
		return fmt.Errorf("item with name %s already exists", req.ItemName)
	}
	// If error is not "not found" - this is a real error
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		logger.Error("failed to check item existence",
			zap.Error(err))
		return fmt.Errorf("failed to check item existence: %w", err)
	}

	// Create the item
	if err := i.repo.CreateItem(ctx, req); err != nil {
		logger.Error("failed to create item", zap.Error(err))
		return fmt.Errorf("failed to create item: %w", err)
	}

	logger.Info("item successfully created",
		zap.String("item_name", req.ItemName),
		zap.Int64("price", req.ItemPrice))
	return nil
}

func (i *ItemUseCase) GetItemByID(ctx context.Context, itemID int64) (*inventory.Item, error) {
	logger.Debug("Check if item exists", zap.Int64("item_id", itemID))
	item, err := i.repo.GetItemByID(ctx, itemID)
	if err != nil {
		logger.Error("getting item by id", zap.Error(err))
		return nil, fmt.Errorf("getting item by id: %w", err)
	}

	logger.Info("Item found", zap.Int64("item_id", itemID))
	return item, nil
}

func (i *ItemUseCase) GetItemByName(ctx context.Context, itemName string) (*inventory.Item, error) {
	logger.Debug("Check if item exists", zap.String("item_name", itemName))
	item, err := i.repo.GetItemByName(ctx, itemName)
	if err != nil {
		logger.Error("getting item by name", zap.Error(err))
		return nil, fmt.Errorf("getting item by name: %w", err)
	}

	logger.Info("Item found", zap.String("item_name", itemName))
	return item, nil
}

func (i *ItemUseCase) UpdateItemPrice(ctx context.Context, req *inventory.UpdatePriceRequest) error {
	// First, let's validate the input
	if req.ItemId <= 0 {
		logger.Error("attempt to update price with invalid item ID",
			zap.Int64("item_id", req.ItemId))
		return fmt.Errorf("invalid item ID")
	}
	if req.Price < 0 {
		logger.Error("attempt to set negative price",
			zap.Int64("price", req.Price))
		return fmt.Errorf("item price cannot be negative")
	}

	// Get current item to check if it exists and compare prices
	logger.Debug("checking item existence and current price",
		zap.Int64("item_id", req.ItemId))
	currentItem, err := i.repo.GetItemByID(ctx, req.ItemId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error("item not found",
				zap.Int64("item_id", req.ItemId))
			return fmt.Errorf("item with ID %d not found", req.ItemId)
		}
		logger.Error("failed to get item",
			zap.Int64("item_id", req.ItemId),
			zap.Error(err))
		return fmt.Errorf("failed to get item: %w", err)
	}

	// Calculate price change percentage for validation
	priceChange := float64(req.Price-currentItem.ItemPrice) / float64(currentItem.ItemPrice)
	if math.Abs(priceChange) > 0.5 {
		logger.Error("price change exceeds allowed limit",
			zap.Int64("item_id", req.ItemId),
			zap.Int64("current_price", currentItem.ItemPrice),
			zap.Int64("new_price", req.Price),
			zap.Float64("change_percentage", priceChange*100))
		return fmt.Errorf("price change cannot exceed 50%% of current price")
	}

	// Update the price
	if err := i.repo.UpdateItemPrice(ctx, req); err != nil {
		logger.Error("failed to update item price",
			zap.Int64("item_id", req.ItemId),
			zap.Error(err))
		return fmt.Errorf("failed to update item price: %w", err)
	}

	logger.Info("item price successfully updated",
		zap.Int64("item_id", req.ItemId),
		zap.Int64("old_price", currentItem.ItemPrice),
		zap.Int64("new_price", req.Price))
	return nil
}

func (i *ItemUseCase) DeleteItem(ctx context.Context, req *inventory.DeleteItemRequest) error {
	logger.Debug("Check if item exists", zap.Int64("item_id", req.ItemId))
	_, err := i.repo.GetItemByID(ctx, req.ItemId)
	if err != nil {
		logger.Error("getting item by id", zap.Error(err))
		return fmt.Errorf("getting item by id: %w", err)
	}

	if err := i.repo.DeleteItem(ctx, req); err != nil {
		logger.Error("deleting item", zap.Error(err))
		return fmt.Errorf("deleting item: %w", err)
	}

	logger.Info("Item deleted", zap.Int64("item_id", req.ItemId))
	return nil
}

func (i *ItemUseCase) GetItemsByRarity(ctx context.Context, req *inventory.GetItemByRarityRequest) ([]*inventory.Item, error) {
	logger.Debug("Get items by rarity", zap.String("rarity", req.Rarity))
	items, err := i.repo.GetItemsByRarity(ctx, req)
	if err != nil {
		logger.Error("getting items by rarity", zap.Error(err))
		return nil, fmt.Errorf("getting items by rarity: %w", err)
	}

	logger.Info("Items found by rarity", zap.String("rarity", req.Rarity))
	return items, nil
}

func (i *ItemUseCase) GetItemsByPriceRange(ctx context.Context, req *inventory.GetItemsByPriceRequest) ([]*inventory.Item, error) {
	logger.Debug("Get items by price range", zap.Int64("min_price", req.StartingPrice), zap.Int64("max_price", req.HighestPrice))
	items, err := i.repo.GetItemsByPriceRange(ctx, req)
	if err != nil {
		logger.Error("getting items by price range", zap.Error(err))
		return nil, fmt.Errorf("getting items by price range: %w", err)
	}

	logger.Info("Items found by price range", zap.Int64("min_price", req.StartingPrice), zap.Int64("max_price", req.HighestPrice))
	return items, nil
}

type InventoryUseCase struct {
	repo interfaces.InventoryRepository
}

func NewInventoryUseCase(repo interfaces.InventoryRepository) *InventoryUseCase {
	return &InventoryUseCase{repo: repo}
}

func (i *InventoryUseCase) GetUserInventory(ctx context.Context, req *inventory.GetAllInventoryInfoRequest) (*inventory.InventoryInfo, error) {
	logger.Debug("Get user inventory", zap.Int64("user_id", req.UserId))
	inventoryInfo, err := i.repo.GetUserInventory(ctx, req)
	if err != nil {
		logger.Error("getting user inventory", zap.Error(err))
		return nil, fmt.Errorf("getting user inventory: %w", err)
	}

	logger.Info("User inventory found", zap.Int64("user_id", req.UserId))
	return inventoryInfo, nil
}

func (i *InventoryUseCase) UpdateItemCount(ctx context.Context, req *inventory.UpdateCountOfItemRequest) error {
	logger.Debug("Update item count", zap.Int64("item_id", req.ItemId), zap.Int64("new_count", req.NewCount))

	if err := i.repo.UpdateItemCount(ctx, req); err != nil {
		logger.Error("updating item count", zap.Error(err))
		return fmt.Errorf("updating item count: %w", err)
	}

	logger.Info("Item count updated", zap.Int64("item_id", req.ItemId), zap.Int64("new_count", req.NewCount))
	return nil
}

func (i *InventoryUseCase) GetAllUserItems(ctx context.Context, req *inventory.GetAllItemsRequest) ([]*inventory.Item, int32, error) {
	logger.Debug("Get all user items", zap.Int64("user_id", req.UserId))
	items, count, err := i.repo.GetAllUserItems(ctx, req)
	if err != nil {
		logger.Error("getting all user items", zap.Error(err))
		return nil, 0, fmt.Errorf("getting all user items: %w", err)
	}

	logger.Info("User items found", zap.Int64("user_id", req.UserId))
	return items, count, nil
}
