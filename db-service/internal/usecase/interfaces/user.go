package interfaces

import (
	"context"
	"petProject/proto/user"
)

// UserUseCase основные операции с пользователем
type UserUseCase interface {
	// Регистрация и аутентификация
	Register(ctx context.Context, req *user.CreateUserRequest) error
	Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error)

	// Управление профилем
	GetUserProfile(ctx context.Context, userID int64) (*user.User, error)
	UpdateUserProfile(ctx context.Context, req *user.UpdateUserRequest) error
	DeleteUserProfile(ctx context.Context, userID int64) error
}

// UserBalanceUseCase операции с балансом
type UserBalanceUseCase interface {
	GetUserBalance(ctx context.Context, userID int64) (int64, error)
	AddBalance(ctx context.Context, req *user.AddBalanceRequest) error
	WithdrawBalance(ctx context.Context, req *user.WithdrawBalanceRequest) error
}

// UserAuthUseCase операции с безопасностью
type UserAuthUseCase interface {
	// Безопасность
	UpdatePassword(ctx context.Context, req *user.UpdatePasswordRequest) error
}
