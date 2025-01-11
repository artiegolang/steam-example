package interfaces

import (
	"context"
	"petProject/proto/user"
)

// UserUseCase основные операции с пользователем
type UserUseCase interface {
	// Основные операции
	Register(ctx context.Context, req *user.CreateUserRequest) error
	GetUserProfile(ctx context.Context, userID int64) (*user.User, error)
	UpdateUserProfile(ctx context.Context, req *user.UpdateUserRequest) error
	DeleteUserProfile(ctx context.Context, userID int64) error

	// Операции с балансом
	GetUserBalance(ctx context.Context, userID int64) (int64, error)
	AddBalance(ctx context.Context, req *user.AddBalanceRequest) error
	WithdrawBalance(ctx context.Context, req *user.WithdrawBalanceRequest) error

	// Операции с аутентификацией
	Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error)
	UpdatePassword(ctx context.Context, req *user.UpdatePasswordRequest) error
}
