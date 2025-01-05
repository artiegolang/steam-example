package impl

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"petProject/db-service/internal/repository/interfaces"
	"petProject/db-service/pkg/logger"
	"petProject/proto/user"
)

type UserUseCase struct {
	// Репозитории, доступные через интерфейсы
	coreRepo    interfaces.CoreUserRepository
	balanceRepo interfaces.UserBalanceRepository
	authRepo    interfaces.UserAuthRepository
}

// NewUserUseCase создает новый экземпляр UserUseCase
func NewUserUseCase(
	coreRepo interfaces.CoreUserRepository,
	balanceRepo interfaces.UserBalanceRepository,
	authRepo interfaces.UserAuthRepository,
) *UserUseCase {
	return &UserUseCase{
		coreRepo:    coreRepo,
		balanceRepo: balanceRepo,
		authRepo:    authRepo,
	}
}

func (u *UserUseCase) Register(ctx context.Context, req *user.CreateUserRequest) error {
	// Проверяем, что пользователь с таким username не существует
	logger.Debug("Checking if username already exists", zap.String("username", req.Username))
	existingUser, err := u.authRepo.GetByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		logger.Error("username already exists", zap.String("username", req.Username))
		return fmt.Errorf("username already exists: %s", req.Username)
	}

	// Хешируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash password", zap.Error(err))
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Обновляем пароль на хешированный
	req.Password = string(hashedPassword)

	// Создаем пользователя
	if err := u.coreRepo.CreateUser(ctx, req); err != nil {
		logger.Error("failed to create user", zap.Error(err))
		return fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info("User successfully registered", zap.String("username", req.Username))
	return nil
}

func (u *UserUseCase) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	logger.Debug("Attempting login", zap.String("username", req.Username))

	// Получаем пользователя по username
	existingUser, err := u.authRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		logger.Error("failed to find user", zap.String("username", req.Username), zap.Error(err))
		return nil, fmt.Errorf("invalid username or password")
	}

	// Проверяем пароль
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password))
	if err != nil {
		logger.Error("invalid password", zap.String("username", req.Username))
		return nil, fmt.Errorf("invalid username or password")
	}

	// Формируем ответ
	response := &user.LoginResponse{
		Success:  true,
		Message:  "login successful",
		UserId:   existingUser.UserId,
		Username: existingUser.Username,
		Role:     existingUser.Role,
	}

	logger.Info("User successfully logged in", zap.String("username", req.Username))
	return response, nil
}

func (u *UserUseCase) GetUserProfile(ctx context.Context, userID int64) (*user.User, error) {
	logger.Debug("Fetching user profile", zap.Int64("user_id", userID))

	// Получаем пользователя по ID
	existingUser, err := u.coreRepo.GetUser(ctx, userID)
	if err != nil {
		logger.Error("failed to get user", zap.Int64("user_id", userID), zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	response := &user.User{
		UserId:   existingUser.UserId,
		Username: existingUser.Username,
		Email:    existingUser.Email,
		Role:     existingUser.Role,
	}

	logger.Info("User profile fetched", zap.Int64("user_id", userID))
	return response, nil
}

func (u *UserUseCase) UpdateUserProfile(ctx context.Context, req *user.UpdateUserRequest) error {
	logger.Debug("Updating user profile", zap.Int64("user_id", req.UserId))

	// Проверяем существование пользователя
	existingUser, err := u.coreRepo.GetUser(ctx, req.UserId)
	if err != nil {
		logger.Error("user not found", zap.Int64("user_id", req.UserId), zap.Error(err))
		return fmt.Errorf("user not found: %w", err)
	}

	// Если меняется username, проверяем его уникальность
	if req.Username != "" && req.Username != existingUser.Username {
		if user, _ := u.authRepo.GetByUsername(ctx, req.Username); user != nil {
			logger.Error("username already taken", zap.String("username", req.Username))
			return fmt.Errorf("username already taken")
		}
	}

	// Обновляем профиль пользователя
	if err := u.coreRepo.UpdateUser(ctx, req); err != nil {
		logger.Error("failed to update user", zap.Int64("user_id", req.UserId), zap.Error(err))
		return fmt.Errorf("failed to update user: %w", err)
	}

	logger.Info("User profile updated", zap.Int64("user_id", req.UserId))
	return nil
}

func (u *UserUseCase) DeleteUserProfile(ctx context.Context, userID int64) error {
	logger.Debug("Deleting user profile", zap.Int64("user_id", userID))

	// Проверяем существование пользователя и его роль
	existingUser, err := u.coreRepo.GetUser(ctx, userID)
	if err != nil {
		logger.Error("user not found", zap.Int64("user_id", userID), zap.Error(err))
		return fmt.Errorf("user not found: %w", err)
	}

	// Проверяем роль пользователя
	if existingUser.Role == user.User_ADMIN { // Используем константу из proto файла
		logger.Error("cannot delete admin user", zap.Int64("user_id", userID))
		return fmt.Errorf("cannot delete admin user")
	}

	// Проверяем баланс пользователя
	balance, err := u.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		logger.Error("failed to check user balance", zap.Int64("user_id", userID), zap.Error(err))
		return fmt.Errorf("failed to check user balance: %w", err)
	}
	if balance > 0 {
		logger.Error("cannot delete user with positive balance",
			zap.Int64("user_id", userID),
			zap.Int64("balance", balance))
		return fmt.Errorf("cannot delete user with positive balance: %d", balance)
	}

	// Удаляем пользователя
	if err := u.coreRepo.DeleteUser(ctx, userID); err != nil {
		logger.Error("failed to delete user", zap.Int64("user_id", userID), zap.Error(err))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	logger.Info("User profile deleted", zap.Int64("user_id", userID))
	return nil
}

func (u *UserUseCase) GetUserBalance(ctx context.Context, userID int64) (int64, error) {
	logger.Debug("Fetching user balance", zap.Int64("user_id", userID))

	// Получаем баланс пользователя
	balance, err := u.balanceRepo.GetBalance(ctx, userID)
	if err != nil {
		logger.Error("failed to get user balance", zap.Int64("user_id", userID), zap.Error(err))
		return 0, fmt.Errorf("failed to get user balance: %w", err)
	}

	logger.Info("User balance fetched", zap.Int64("user_id", userID), zap.Int64("balance", balance))
	return balance, nil
}

func (u *UserUseCase) AddBalance(ctx context.Context, req *user.AddBalanceRequest) error {
	logger.Debug("Adding balance", zap.Int64("user_id", req.UserId), zap.Int64("amount", req.Amount))

	// Проверяем существование пользователя
	_, err := u.coreRepo.GetUser(ctx, req.UserId)
	if err != nil {
		logger.Error("user not found", zap.Int64("user_id", req.UserId), zap.Error(err))
		return fmt.Errorf("user not found: %w", err)
	}

	// Проверяем положительную сумму
	if req.Amount <= 0 {
		logger.Error("invalid amount", zap.Int64("amount", req.Amount))
		return fmt.Errorf("invalid amount: %d", req.Amount)
	}

	// Защита от переполнения баланса
	if req.Amount > 1_000_000_000 {
		logger.Error("amount too large", zap.Int64("amount", req.Amount))
		return fmt.Errorf("amount too large: %d", req.Amount)
	}

	// Преобразуем AddBalanceRequest в UpdateBalanceRequest
	updateReq := &user.UpdateBalanceRequest{
		UserId:     req.UserId,
		NewBalance: req.Amount,
	}

	// Добавляем баланс
	if err := u.balanceRepo.UpdateBalance(ctx, updateReq); err != nil {
		logger.Error("failed to add balance", zap.Int64("user_id", req.UserId), zap.Error(err))
		return fmt.Errorf("failed to add balance: %w", err)
	}

	logger.Info("Balance added successfully",
		zap.Int64("user_id", req.UserId),
		zap.Int64("amount", req.Amount))

	return nil
}

func (u *UserUseCase) WithdrawBalance(ctx context.Context, req *user.WithdrawBalanceRequest) error {
	logger.Debug("Withdrawing balance", zap.Int64("user_id", req.UserId), zap.Int64("amount", req.Amount))

	// Проверяем существование пользователя
	_, err := u.coreRepo.GetUser(ctx, req.UserId)
	if err != nil {
		logger.Error("user not found", zap.Int64("user_id", req.UserId), zap.Error(err))
		return fmt.Errorf("user not found: %w", err)
	}

	// Проверяем положительную сумму
	if req.Amount <= 0 {
		logger.Error("invalid amount", zap.Int64("amount", req.Amount))
		return fmt.Errorf("invalid amount: %d", req.Amount)
	}

	// Получаем текущий баланс
	balance, err := u.balanceRepo.GetBalance(ctx, req.UserId)
	if err != nil {
		logger.Error("failed to get user balance", zap.Int64("user_id", req.UserId), zap.Error(err))
		return fmt.Errorf("failed to get user balance: %w", err)
	}

	// Проверяем, что баланс не отрицательный
	if balance < req.Amount {
		logger.Error("insufficient balance", zap.Int64("user_id", req.UserId), zap.Int64("balance", balance))
		return fmt.Errorf("insufficient balance: %d", balance)
	}

	// Преобразуем WithdrawBalanceRequest в UpdateBalanceRequest
	updateReq := &user.UpdateBalanceRequest{
		UserId:     req.UserId,
		NewBalance: -req.Amount,
	}

	// Снимаем баланс
	if err := u.balanceRepo.UpdateBalance(ctx, updateReq); err != nil {
		logger.Error("failed to withdraw balance", zap.Int64("user_id", req.UserId), zap.Error(err))
		return fmt.Errorf("failed to withdraw balance: %w", err)
	}

	logger.Info("Balance withdrawn successfully",
		zap.Int64("user_id", req.UserId),
		zap.Int64("amount", req.Amount))

	return nil
}

func (u *UserUseCase) UpdatePassword(ctx context.Context, req *user.UpdatePasswordRequest) error {
	logger.Debug("Updating password", zap.Int64("user_id", req.UserId))

	// Получаем пользователя по ID
	existingUser, err := u.coreRepo.GetUser(ctx, req.UserId)
	if err != nil {
		logger.Error("user not found", zap.Int64("user_id", req.UserId), zap.Error(err))
		return fmt.Errorf("user not found: %w", err)
	}

	// Проверяем старый пароль
	err = bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.OldPassword))
	if err != nil {
		logger.Error("invalid old password", zap.Int64("user_id", req.UserId))
		return fmt.Errorf("invalid old password")
	}

	// Валидация нового пароля
	if len(req.NewPassword) < 8 {
		logger.Error("password too short", zap.Int64("user_id", req.UserId))
		return fmt.Errorf("new password must be at least 8 characters long")
	}

	// Проверяем, что новый пароль отличается от старого
	if req.NewPassword == req.OldPassword {
		logger.Error("new password same as old", zap.Int64("user_id", req.UserId))
		return fmt.Errorf("new password must be different from old password")
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash password", zap.Error(err))
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Обновляем пароль на хешированный
	req.NewPassword = string(hashedPassword)

	// Обновляем пароль
	if err := u.authRepo.UpdatePassword(ctx, req); err != nil {
		logger.Error("failed to update password", zap.Int64("user_id", req.UserId), zap.Error(err))
		return fmt.Errorf("failed to update password: %w", err)
	}

	logger.Info("Password updated successfully", zap.Int64("user_id", req.UserId))
	return nil
}
