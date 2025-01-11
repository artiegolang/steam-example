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
	"petProject/proto/inventory"
	"time"
)

type ItemRepository struct {
	pool *pgxpool.Pool
}

func NewItemRepository(pool *pgxpool.Pool) *ItemRepository {
	return &ItemRepository{
		pool: pool,
	}
}

func (r *ItemRepository) CreateItem(ctx context.Context, req *inventory.CreateItemRequest) error {
	if req.ItemName == "" {
		logger.Error("empty item name")
		return fmt.Errorf("item name cannot be empty")
	}
	if req.ItemPrice < 0 {
		logger.Error("negative item price", zap.Int64("price", req.ItemPrice))
		return fmt.Errorf("item price cannot be negative")
	}

	query := `
        INSERT INTO items (item_id, item_name, item_price, created_at, updated_at)
        VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING item_id`

	logger.Debug("Creating new item",
		zap.String("item_name", req.ItemName),
		zap.Int64("item_id", req.ItemId),
		zap.Int64("price", req.ItemPrice))

	var itemID int64
	err := r.pool.QueryRow(ctx, query,
		req.ItemId,
		req.ItemName,
		req.ItemPrice,
	).Scan(&itemID)

	if err != nil {
		// Используем pgconn.PgError вместо pgx.PgError
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // Unique violation
				logger.Error("item with this ID already exists",
					zap.Int64("item_id", req.ItemId))
				return fmt.Errorf("item with ID %d already exists", req.ItemId)
			}
		}
		logger.Error("failed to create item",
			zap.Error(err),
			zap.String("item_name", req.ItemName))
		return fmt.Errorf("failed to create item: %w", err)
	}

	logger.Info("Item created",
		zap.Int64("item_id", itemID),
		zap.String("item_name", req.ItemName))
	return nil
}

func (r *ItemRepository) GetItemByID(ctx context.Context, itemID int64) (*inventory.Item, error) {
	query := `
        SELECT 
            item_id, 
            item_name, 
            item_price, 
            rarity, 
            created_at, 
            updated_at 
        FROM items 
        WHERE item_id = $1`

	logger.Debug("Fetching item by ID", zap.Int64("item_id", itemID))

	var item inventory.Item
	// Временные метки хранятся в БД, но не передаются в proto-сообщении
	var createdAt, updatedAt time.Time

	err := r.pool.QueryRow(ctx, query, itemID).Scan(
		&item.ItemId,
		&item.ItemName,
		&item.ItemPrice,
		&item.Rarity,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		// Специальная обработка случая, когда предмет не найден
		if err == pgx.ErrNoRows {
			logger.Error("item not found", zap.Int64("item_id", itemID))
			return nil, fmt.Errorf("item not found with id: %d", itemID)
		}
		logger.Error("failed to fetch item",
			zap.Int64("item_id", itemID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch item: %w", err)
	}

	logger.Info("Item fetched successfully",
		zap.Int64("item_id", itemID),
		zap.String("item_name", item.ItemName))
	return &item, nil
}

func (r *ItemRepository) GetItemByName(ctx context.Context, itemName string) (*inventory.Item, error) {
	query := `
        SELECT 
            item_id, 
            item_name, 
            item_price, 
            rarity, 
            created_at, 
            updated_at 
        FROM items 
        WHERE LOWER(item_name) = LOWER($1)`

	logger.Debug("Fetching item by name", zap.String("item_name", itemName))

	var item inventory.Item
	var createdAt, updatedAt time.Time

	err := r.pool.QueryRow(ctx, query, itemName).Scan(
		&item.ItemId,
		&item.ItemName,
		&item.ItemPrice,
		&item.Rarity,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Error("item not found", zap.String("item_name", itemName))
			return nil, fmt.Errorf("item not found with name: %s", itemName)
		}
		logger.Error("failed to fetch item",
			zap.String("item_name", itemName),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch item: %w", err)
	}

	logger.Info("Item fetched successfully", zap.String("item_name", itemName))
	return &item, nil
}

func (r *ItemRepository) UpdateItemPrice(ctx context.Context, req *inventory.UpdatePriceRequest) error {
	if req.Price < 0 {
		logger.Error("negative item price", zap.Int64("price", req.Price))
		return fmt.Errorf("item price cannot be negative")
	}

	query := `
		UPDATE items 
		SET item_price = $1, updated_at = CURRENT_TIMESTAMP
		WHERE item_id = $2
		RETURNING item_id`

	logger.Debug("Updating item price", zap.Int64("item_id", req.ItemId), zap.Int64("price", req.Price))
	var itemID int64
	err := r.pool.QueryRow(ctx, query, req.Price, req.ItemId).Scan(&itemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Error("item not found", zap.Int64("item_id", req.ItemId))
			return fmt.Errorf("item not found with id: %d", req.ItemId)
		}
		logger.Error("failed to update item price",
			zap.Int64("item_id", req.ItemId),
			zap.Error(err))
		return fmt.Errorf("failed to update item price: %w", err)
	}

	logger.Info("Item price updated", zap.Int64("item_id", itemID), zap.Int64("price", req.Price))
	return nil
}

func (r *ItemRepository) DeleteItem(ctx context.Context, req *inventory.DeleteItemRequest) error {
	checkQuery := `
        SELECT EXISTS(SELECT 1 FROM inventory WHERE item_id = $1)
    `
	var exists bool
	if err := r.pool.QueryRow(ctx, checkQuery, req.ItemId).Scan(&exists); err != nil {
		logger.Error("failed to check item usage",
			zap.Int64("item_id", req.ItemId),
			zap.Error(err))
		return fmt.Errorf("failed to check item usage: %w", err)
	}

	if exists {
		logger.Error("cannot delete item that exists in inventory",
			zap.Int64("item_id", req.ItemId))
		return fmt.Errorf("cannot delete item that exists in inventory: %d", req.ItemId)
	}

	query := `
		DELETE FROM items
		WHERE item_id = $1`

	logger.Debug("Deleting item", zap.Int64("item_id", req.ItemId))

	result, err := r.pool.Exec(ctx, query, req.ItemId)
	if err != nil {
		logger.Error("failed to delete item",
			zap.Int64("item_id", req.ItemId),
			zap.Error(err))
		return fmt.Errorf("failed to delete item: %w", err)
	}

	if result.RowsAffected() == 0 {
		logger.Error("item not found", zap.Int64("item_id", req.ItemId))
		return fmt.Errorf("item not found with id: %d", req.ItemId)
	}

	logger.Info("Item deleted", zap.Int64("item_id", req.ItemId))
	return nil
}

func (r *ItemRepository) GetItemsByRarity(ctx context.Context, req *inventory.GetItemByRarityRequest) ([]*inventory.Item, error) {
	// Формируем SQL запрос для выборки предметов по редкости
	// Выбираем все необходимые поля из таблицы items
	// WHERE rarity = $1 фильтрует предметы по заданной редкости
	query := `
       SELECT 
          item_id, 
          item_name, 
          item_price, 
          rarity, 
          created_at, 
          updated_at 
       FROM items 
       WHERE rarity = $1`

	logger.Debug("Fetching items by rarity", zap.String("rarity", req.Rarity))

	// Выполняем запрос к базе данных
	// Query используется вместо QueryRow, так как мы ожидаем multiple rows
	// Параметр req.Rarity передается в запрос вместо $1
	rows, err := r.pool.Query(ctx, query, req.Rarity)
	if err != nil {
		// Обрабатываем случай, когда предметы не найдены
		if err == pgx.ErrNoRows {
			logger.Error("items not found", zap.String("rarity", req.Rarity))
			return nil, fmt.Errorf("items not found with rarity: %s", req.Rarity)
		}
		// Обрабатываем другие ошибки базы данных
		logger.Error("failed to fetch items",
			zap.String("rarity", req.Rarity),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch item: %w", err)
	}

	// Создаем слайс для хранения результатов
	// []*inventory.Item означает, что мы храним указатели на предметы
	var items []*inventory.Item

	// Итерируемся по результатам запроса
	// rows.Next() перемещает курсор к следующей строке результата
	for rows.Next() {
		// Создаем временные переменные для каждой строки результата
		var item inventory.Item
		var createdAt, updatedAt time.Time

		// Сканируем значения из текущей строки в переменные
		// Важно соблюдать порядок полей, соответствующий SQL запросу
		if err := rows.Scan(
			&item.ItemId,
			&item.ItemName,
			&item.ItemPrice,
			&item.Rarity,
			&createdAt,
			&updatedAt,
		); err != nil {
			// Если возникла ошибка при сканировании, логируем и возвращаем ошибку
			logger.Error("failed to scan item", zap.Error(err))
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		// Добавляем указатель на предмет в итоговый слайс
		items = append(items, &item)
	}

	// Логируем успешное завершение операции
	logger.Info("Items fetched successfully", zap.String("rarity", req.Rarity))

	// Возвращаем слайс предметов и nil в качестве ошибки
	return items, nil
}

func (r *ItemRepository) GetItemsByPriceRange(ctx context.Context, req *inventory.GetItemsByPriceRequest) ([]*inventory.Item, error) {
	if req.StartingPrice < 0 || req.HighestPrice < 0 {
		logger.Error("negative price range",
			zap.Int64("starting_price", req.StartingPrice),
			zap.Int64("highest_price", req.HighestPrice))
		return nil, fmt.Errorf("price range cannot be negative")
	}

	if req.StartingPrice > req.HighestPrice {
		logger.Error("invalid price range",
			zap.Int64("starting_price", req.StartingPrice),
			zap.Int64("highest_price", req.HighestPrice))
		return nil, fmt.Errorf("starting price cannot be greater than highest price")
	}

	query := `
        SELECT 
            item_id, 
            item_name, 
            item_price, 
            rarity, 
            created_at, 
            updated_at 
        FROM items 
        WHERE item_price >= $1 AND item_price <= $2
        ORDER BY item_price
        LIMIT $3 OFFSET $4`

	offset := (req.PageNumber - 1) * req.PageSize

	logger.Debug("Fetching items by price range",
		zap.Int64("starting_price", req.StartingPrice),
		zap.Int64("highest_price", req.HighestPrice),
		zap.Int32("page_size", req.PageSize),
		zap.Int32("page_number", req.PageNumber))

	rows, err := r.pool.Query(ctx, query,
		req.StartingPrice,
		req.HighestPrice,
		req.PageSize,
		offset)
	if err != nil {
		logger.Error("failed to fetch items",
			zap.Int64("starting_price", req.StartingPrice),
			zap.Int64("highest_price", req.HighestPrice),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch items: %w", err)
	}
	defer rows.Close()

	var items []*inventory.Item

	for rows.Next() {
		var item inventory.Item
		var createdAt, updatedAt time.Time

		if err := rows.Scan(
			&item.ItemId,
			&item.ItemName,
			&item.ItemPrice,
			&item.Rarity,
			&createdAt,
			&updatedAt,
		); err != nil {
			logger.Error("failed to scan item", zap.Error(err))
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	if err = rows.Err(); err != nil {
		logger.Error("error iterating over rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	logger.Info("Items fetched successfully",
		zap.Int64("starting_price", req.StartingPrice),
		zap.Int64("highest_price", req.HighestPrice),
		zap.Int("items_found", len(items)))
	return items, nil
}

type InventoryRepository struct {
	pool *pgxpool.Pool
}

func NewInventoryRepository(pool *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{
		pool: pool,
	}
}

// Методы для работы с инвентарем
func (r *InventoryRepository) GetUserInventory(ctx context.Context, req *inventory.GetAllInventoryInfoRequest) (*inventory.InventoryInfo, error) {
	// Сначала получаем информацию о пользователе
	userQuery := `
        SELECT username 
        FROM users 
        WHERE user_id = $1`

	var inventoryInfo inventory.InventoryInfo
	err := r.pool.QueryRow(ctx, userQuery, req.UserId).Scan(&inventoryInfo.Username)
	if err != nil {
		if err == pgx.ErrNoRows {
			logger.Error("user not found", zap.Int64("user_id", req.UserId))
			return nil, fmt.Errorf("user not found: %d", req.UserId)
		}
		logger.Error("failed to fetch user",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Теперь получаем все предметы пользователя с их информацией
	itemsQuery := `
        SELECT i.item_id, i.item_name, i.item_price, i.rarity, inv.item_count
        FROM inventory inv
        JOIN items i ON inv.item_id = i.item_id
        WHERE inv.user_id = $1`

	rows, err := r.pool.Query(ctx, itemsQuery, req.UserId)
	if err != nil {
		logger.Error("failed to fetch inventory items",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch inventory items: %w", err)
	}
	defer rows.Close()

	// Инициализируем необходимые поля
	inventoryInfo.UserId = req.UserId
	inventoryInfo.Items = make([]*inventory.Item, 0)
	var overallPrice int64 = 0

	// Обрабатываем каждый предмет
	for rows.Next() {
		var item inventory.Item
		if err := rows.Scan(
			&item.ItemId,
			&item.ItemName,
			&item.ItemPrice,
			&item.Rarity,
			&item.ItemCount,
		); err != nil {
			logger.Error("failed to scan inventory item", zap.Error(err))
			return nil, fmt.Errorf("failed to scan inventory item: %w", err)
		}

		// Считаем общую стоимость
		overallPrice += item.ItemPrice * item.ItemCount
		inventoryInfo.Items = append(inventoryInfo.Items, &item)
	}

	if err = rows.Err(); err != nil {
		logger.Error("error iterating over inventory items", zap.Error(err))
		return nil, fmt.Errorf("error iterating over inventory items: %w", err)
	}

	// Устанавливаем общую стоимость
	inventoryInfo.OverallPrice = overallPrice

	logger.Info("Inventory fetched successfully",
		zap.Int64("user_id", req.UserId),
		zap.Int("items_count", len(inventoryInfo.Items)),
		zap.Int64("overall_price", overallPrice))

	return &inventoryInfo, nil
}

func (r *InventoryRepository) UpdateItemCount(ctx context.Context, req *inventory.UpdateCountOfItemRequest) error {
	// Проверяем входные данные на корректность
	if req.NewCount < 0 {
		logger.Error("negative item count",
			zap.Int64("new_count", req.NewCount),
			zap.Int64("item_id", req.ItemId))
		return fmt.Errorf("item count cannot be negative")
	}

	// Если количество равно 0, удаляем предмет из инвентаря
	if req.NewCount == 0 {
		query := `
            DELETE FROM inventory 
            WHERE user_id = $1 AND item_id = $2
            RETURNING user_id`

		logger.Debug("Removing item from inventory",
			zap.Int64("user_id", req.UserId),
			zap.Int64("item_id", req.ItemId))

		var userID int64
		err := r.pool.QueryRow(ctx, query, req.UserId, req.ItemId).Scan(&userID)
		if err != nil {
			if err == pgx.ErrNoRows {
				logger.Error("item not found in inventory",
					zap.Int64("user_id", req.UserId),
					zap.Int64("item_id", req.ItemId))
				return fmt.Errorf("item not found in inventory")
			}
			logger.Error("failed to remove item from inventory",
				zap.Error(err))
			return fmt.Errorf("failed to remove item from inventory: %w", err)
		}

		logger.Info("Item removed from inventory",
			zap.Int64("user_id", req.UserId),
			zap.Int64("item_id", req.ItemId))
		return nil
	}

	// Обновляем количество предметов
	// UPSERT (INSERT ... ON CONFLICT UPDATE) позволяет создать запись,
	// если её нет, или обновить существующую
	query := `
        INSERT INTO inventory (user_id, item_id, item_count)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id, item_id) 
        DO UPDATE SET item_count = EXCLUDED.item_count
        RETURNING user_id`

	logger.Debug("Updating item count in inventory",
		zap.Int64("user_id", req.UserId),
		zap.Int64("item_id", req.ItemId),
		zap.Int64("new_count", req.NewCount))

	var userID int64
	err := r.pool.QueryRow(ctx, query,
		req.UserId,
		req.ItemId,
		req.NewCount).Scan(&userID)

	if err != nil {
		logger.Error("failed to update item count",
			zap.Error(err))
		return fmt.Errorf("failed to update item count: %w", err)
	}

	logger.Info("Item count updated successfully",
		zap.Int64("user_id", req.UserId),
		zap.Int64("item_id", req.ItemId),
		zap.Int64("new_count", req.NewCount))
	return nil
}

func (r *InventoryRepository) GetAllUserItems(ctx context.Context, req *inventory.GetAllItemsRequest) ([]*inventory.Item, int32, error) {
	// Сначала получаем общее количество предметов для пагинации
	countQuery := `
        SELECT COUNT(*) 
        FROM inventory inv
        JOIN items i ON inv.item_id = i.item_id
        WHERE inv.user_id = $1`

	var totalItems int32
	err := r.pool.QueryRow(ctx, countQuery, req.UserId).Scan(&totalItems)
	if err != nil {
		logger.Error("failed to get total items count",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to get total items count: %w", err)
	}

	// Формируем запрос с пагинацией для получения предметов
	// Используем JOIN для получения полной информации о предметах
	query := `
        SELECT 
            i.item_id,
            i.item_name,
            i.item_price,
            i.rarity,
            inv.item_count
        FROM inventory inv
        JOIN items i ON inv.item_id = i.item_id
        WHERE inv.user_id = $1
        ORDER BY i.item_name
        LIMIT $2 OFFSET $3`

	// Вычисляем смещение для пагинации
	offset := (req.PageNumber - 1) * req.PageSize

	logger.Debug("Fetching user items",
		zap.Int64("user_id", req.UserId),
		zap.Int32("page_size", req.PageSize),
		zap.Int32("page_number", req.PageNumber))

	// Выполняем запрос с параметрами пагинации
	rows, err := r.pool.Query(ctx, query,
		req.UserId,
		req.PageSize,
		offset)
	if err != nil {
		logger.Error("failed to fetch user items",
			zap.Error(err))
		return nil, 0, fmt.Errorf("failed to fetch user items: %w", err)
	}
	defer rows.Close()

	// Создаем слайс для хранения результатов
	var items []*inventory.Item

	// Обрабатываем каждую строку результата
	for rows.Next() {
		var item inventory.Item
		if err := rows.Scan(
			&item.ItemId,
			&item.ItemName,
			&item.ItemPrice,
			&item.Rarity,
			&item.ItemCount,
		); err != nil {
			logger.Error("failed to scan item",
				zap.Error(err))
			return nil, 0, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, &item)
	}

	// Проверяем ошибки после цикла
	if err = rows.Err(); err != nil {
		logger.Error("error iterating over items",
			zap.Error(err))
		return nil, 0, fmt.Errorf("error iterating over items: %w", err)
	}

	logger.Info("User items fetched successfully",
		zap.Int64("user_id", req.UserId),
		zap.Int("items_found", len(items)),
		zap.Int32("total_items", totalItems))

	return items, totalItems, nil
}
