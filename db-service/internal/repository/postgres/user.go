package postgres

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"petProject/db-service/pkg/logger"
	"petProject/proto/user"
	"time"
)

type UserRepository struct {
	pool *pgxpool.Pool
	log  logger.Logger
}

func NewUserRepository(pool *pgxpool.Pool, log logger.Logger) *UserRepository {
	return &UserRepository{
		pool: pool,
		log:  log,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, req *user.CreateUserRequest) error {
	query := `
        INSERT INTO users (username, email, password, role, status, balance, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
        RETURNING user_id`

	r.log.Debug("Creating new user",
		zap.String("username", req.Username),
		zap.String("email", req.Email))

	var userID int64
	err := r.pool.QueryRow(ctx, query,
		req.Username,
		req.Email,
		req.Password,
		req.Role,
		"ACTIVE",
		0,
	).Scan(&userID)

	if err != nil {
		r.log.Error("failed to create user", zap.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}

	r.log.Info("User created",
		zap.Int64("user_id", userID),
		zap.String("username", req.Username))

	return nil
}

func (r *UserRepository) GetUser(ctx context.Context, id int64) (*user.User, error) {
	// Создаем SQL запрос для получения всех нужных полей пользователя
	// мы НЕ выбираем пароль - это важно для безопасности
	query := `
        SELECT 
            user_id,
            username,
            email,
            role,
            status,
            balance,
            created_at,
            updated_at
        FROM users
        WHERE user_id = $1 AND status != 'DELETED'`

	// Логируем начало операции получения пользователя
	r.log.Debug("Fetching user by ID", zap.Int64("user_id", id))

	// Создаем переменные для сканирования результата
	var user user.User
	var createdAt, updatedAt time.Time

	// Выполняем запрос и сканируем результат в переменные
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.UserId,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.Status,
		&user.Balance,
		&createdAt,
		&updatedAt,
	)

	// Обрабатываем возможные ошибки
	if err != nil {
		// Если записи не найдено, возвращаем специальную ошибку
		if err == pgx.ErrNoRows {
			r.log.Error("user not found", zap.Int64("user_id", id))
			return nil, fmt.Errorf("user not found with id: %d", id)
		}
		// Если произошла другая ошибка, логируем её
		r.log.Error("failed to fetch user",
			zap.Int64("user_id", id),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Логируем успешное получение пользователя
	r.log.Debug("Successfully fetched user",
		zap.Int64("user_id", id),
		zap.String("username", user.Username))

	// Возвращаем найденного пользователя
	return &user, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) error {
	// Создаем SQL запрос. Используем динамическое обновление полей,
	// обновляя только те поля, которые действительно изменились
	query := `
        UPDATE users 
        SET 
            username = COALESCE($1, username),
            email = COALESCE($2, email),
            role = COALESCE($3, role),
            updated_at = CURRENT_TIMESTAMP
        WHERE user_id = $4 AND status != 'DELETED'
        RETURNING user_id`

	// Логируем начало операции обновления
	r.log.Debug("Updating user",
		zap.Int64("user_id", req.UserId),
		zap.String("username", req.Username),
		zap.String("email", req.Email))

	// Выполняем запрос. Использование QueryRow вместо Exec позволяет нам
	// убедиться, что запись действительно существует и была обновлена
	var updatedID int64
	err := r.pool.QueryRow(ctx, query,
		req.Username,
		req.Email,
		req.Role,
		req.UserId,
	).Scan(&updatedID)

	// Обрабатываем возможные ошибки
	if err != nil {
		// Если записи не найдено
		if err == pgx.ErrNoRows {
			r.log.Error("user not found for update",
				zap.Int64("user_id", req.UserId))
			return fmt.Errorf("user not found for update: %d", req.UserId)
		}

		// Логируем другие ошибки
		r.log.Error("failed to update user",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Логируем успешное обновление
	r.log.Info("Successfully updated user",
		zap.Int64("user_id", req.UserId))

	return nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id int64) error {
	query := `
        UPDATE users
        SET status = 'DELETED'
        WHERE user_id = $1 AND status != 'DELETED'
        RETURNING user_id`

	r.log.Debug("Deleting user", zap.Int64("user_id", id))

	var deletedID int64
	err := r.pool.QueryRow(ctx, query, id).Scan(&deletedID)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.log.Error("user not found for delete", zap.Int64("user_id", id))
			return fmt.Errorf("user not found for delete: %d", id)
		}

		r.log.Error("failed to delete user",
			zap.Int64("user_id", id),
			zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	r.log.Info("Successfully deleted user", zap.Int64("user_id", id))
	return nil
}

func (r *UserRepository) GetBalance(ctx context.Context, userID int64) (int64, error) {
	query := `
        SELECT balance FROM users WHERE user_id = $1 AND status != 'DELETED'`

	r.log.Debug("Fetching user balance", zap.Int64("user_id", userID))

	var balance int64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&balance)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.log.Error("user not found", zap.Int64("user_id", userID))
			return 0, fmt.Errorf("user not found with id: %d", userID)
		}
		r.log.Error("failed to fetch user balance",
			zap.Int64("user_id", userID),
			zap.Error(err))
		return 0, fmt.Errorf("failed to fetch user balance: %w", err)
	}

	r.log.Debug("Successfully fetched user balance", zap.Int64("user_id", userID), zap.Int64("balance", balance))

	return balance, nil
}

func (r *UserRepository) UpdateBalance(ctx context.Context, req *user.UpdateBalanceRequest) error {
	query := `
		UPDATE balance SET balance = balance + $1 WHERE user_id = $2`

	r.log.Debug("Updating user balance", zap.Int64("user_id", req.UserId), zap.Int64("amount", req.NewBalance))

	var balance int64
	err := r.pool.QueryRow(ctx, query, req.NewBalance, req.UserId).Scan(&balance)

	if err != nil {
		if err == pgx.ErrNoRows {
			r.log.Error("user not found", zap.Int64("user_id", req.UserId))
			return fmt.Errorf("user not found with id: %d", req.UserId)
		}
		r.log.Error("failed to update user balance",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	r.log.Debug("Successfully updated user balance", zap.Int64("user_id", req.UserId), zap.Int64("balance", balance))
	return nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	query := `
        SELECT 
            user_id,
            username,
            email,
            role,
            status,
            balance,
            created_at,
            updated_at
        FROM users
        WHERE username = $1 AND status != 'DELETED'`

	r.log.Debug("Fetching user by username", zap.String("username", username))

	var user user.User
	var createdAt, updatedAt time.Time

	// Выполняем запрос и сканируем результат в переменные
	err := r.pool.QueryRow(ctx, query).Scan(
		&user.UserId,
		&user.Username,
		&user.Email,
		&user.Role,
		&user.Status,
		&user.Balance,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		// Если записи не найдено, возвращаем специальную ошибку
		if err == pgx.ErrNoRows {
			r.log.Error("user not found", zap.String("username", username))
			return nil, fmt.Errorf("user not found with name: %s", username)
		}
		// Если произошла другая ошибка, логируем её
		r.log.Error("failed to fetch user",
			zap.String("username", username),
			zap.Error(err))
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	// Логируем успешное получение пользователя
	r.log.Debug("Successfully fetched user",
		zap.String("username", user.Username))

	// Возвращаем найденного пользователя
	return &user, nil
	return nil, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, req *user.UpdatePasswordRequest) error {
	query := `
		UPDATE password SET password = $1 WHERE user_id = $2`

	r.log.Debug("Updating user password", zap.Int64("user_id", req.UserId))

	var password string
	err := r.pool.QueryRow(ctx, query, req.NewPassword, req.UserId).Scan(&password)
	if err != nil {
		if err == pgx.ErrNoRows {
			r.log.Error("user not found", zap.Int64("user_id", req.UserId))
			return fmt.Errorf("user not found with id: %d", req.UserId)
		}
		r.log.Error("failed to update user password",
			zap.Int64("user_id", req.UserId),
			zap.Error(err))
		return fmt.Errorf("failed to update user password: %w", err)
	}

	r.log.Debug("Successfully updated user password", zap.Int64("user_id", req.UserId))

	return nil
}
