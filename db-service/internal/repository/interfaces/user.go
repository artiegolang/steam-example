package interfaces

import (
	"context"
	"petProject/proto/user"
)

// CoreUserRepository отвечает за базовые операции с пользователем
type CoreUserRepository interface {
	CreateUser(ctx context.Context, req *user.CreateUserRequest) error
	GetUser(ctx context.Context, id int64) (*user.User, error)
	UpdateUser(ctx context.Context, req *user.UpdateUserRequest) error
	DeleteUser(ctx context.Context, id int64) error
}

// UserBalanceRepository отвечает за операции с балансом
type UserBalanceRepository interface {
	GetBalance(ctx context.Context, userID int64) (int64, error)
	UpdateBalance(ctx context.Context, req *user.UpdateBalanceRequest) error
}

// UserAuthRepository отвечает за аутентификацию и безопасность
type UserAuthRepository interface {
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	UpdatePassword(ctx context.Context, req *user.UpdatePasswordRequest) error
}
