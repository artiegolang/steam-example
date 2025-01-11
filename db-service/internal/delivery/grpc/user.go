package grpc

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"petProject/db-service/internal/usecase/interfaces"
	"petProject/db-service/pkg/logger"
	"petProject/proto/user"
)

type UserService struct {
	user.UnimplementedUserServiceServer
	useCase interfaces.UserUseCase
}

func NewUserService(useCase interfaces.UserUseCase) *UserService {
	return &UserService{
		useCase: useCase,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	logger.Debug("Handling CreateUser request", zap.String("username", req.Username), zap.String("email", req.Email))

	if err := s.useCase.Register(ctx, req); err != nil {
		// В случае ошибки логируем её и возвращаем соответствующий gRPC статус
		logger.Error("Failed to create user", zap.Error(err))
		return &user.CreateUserResponse{
			Success: false,
			Message: "Failed to create user",
		}, status.Error(codes.Internal, "failed to create user")
	}

	return &user.CreateUserResponse{
		Success: true,
		Message: "User created",
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	logger.Debug("Handling GetUser request", zap.Int64("user_id", req.UserId))

	userInfo, err := s.useCase.GetUserProfile(ctx, req.UserId)
	if err != nil {
		logger.Error("failed to get user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	return &user.GetUserResponse{
		User: userInfo,
	}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	logger.Debug("Handling UpdateUser request", zap.Int64("user_id", req.UserId))

	if err := s.useCase.UpdateUserProfile(ctx, req); err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return &user.UpdateUserResponse{
			Success: false,
			Message: "Failed to update user",
		}, status.Error(codes.Internal, "failed to update user")
	}

	return &user.UpdateUserResponse{
		Success: true,
		Message: "User updated successfully",
	}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	logger.Debug("Handling DeleteUser request", zap.Int64("user_id", req.UserId))

	if err := s.useCase.DeleteUserProfile(ctx, req.UserId); err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return &user.DeleteUserResponse{
			Success: false,
			Message: "Failed to delete user",
		}, status.Error(codes.Internal, "failed to delete user")
	}

	return &user.DeleteUserResponse{
		Success: true,
		Message: "User deleted successfully",
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	logger.Debug("Handling Login request", zap.String("username", req.Username))

	response, err := s.useCase.Login(ctx, req)
	if err != nil {
		logger.Error("Failed to login", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to login")
	}

	return response, nil
}

func (s *UserService) GetBalance(ctx context.Context, req *user.GetBalanceRequest) (*user.GetBalanceResponse, error) {
	logger.Debug("Handling GetBalance request", zap.Int64("user_id", req.UserId))

	balance, err := s.useCase.GetUserBalance(ctx, req.UserId)
	if err != nil {
		logger.Error("Failed to get balance", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get balance")
	}

	return &user.GetBalanceResponse{
		Balance: balance,
	}, nil
}

// UpdateBalance обновляет баланс пользователя
func (s *UserService) UpdateBalance(ctx context.Context, req *user.UpdateBalanceRequest) (*user.UpdateBalanceResponse, error) {
	logger.Debug("Handling UpdateBalance request",
		zap.Int64("user_id", req.UserId),
		zap.Int64("new_balance", req.NewBalance))

	addBalanceReq := &user.AddBalanceRequest{
		UserId: req.UserId,
		Amount: req.NewBalance,
	}

	if err := s.useCase.AddBalance(ctx, addBalanceReq); err != nil {
		logger.Error("Failed to update balance", zap.Error(err))
		return &user.UpdateBalanceResponse{
			Success: false,
			Message: "Failed to update balance",
		}, status.Error(codes.Internal, "failed to update balance")
	}

	return &user.UpdateBalanceResponse{
		Success: true,
		Message: "Balance updated successfully",
	}, nil
}

// AddBalance добавляет средства к балансу пользователя
func (s *UserService) AddBalance(ctx context.Context, req *user.AddBalanceRequest) (*user.UpdateBalanceResponse, error) {
	logger.Debug("Handling AddBalance request",
		zap.Int64("user_id", req.UserId),
		zap.Int64("amount", req.Amount))

	if err := s.useCase.AddBalance(ctx, req); err != nil {
		logger.Error("Failed to add balance", zap.Error(err))
		return &user.UpdateBalanceResponse{
			Success: false,
			Message: "Failed to add balance",
		}, status.Error(codes.Internal, "failed to add balance")
	}

	return &user.UpdateBalanceResponse{
		Success: true,
		Message: "Balance added successfully",
	}, nil
}

// WithdrawBalance снимает средства с баланса пользователя
func (s *UserService) WithdrawBalance(ctx context.Context, req *user.WithdrawBalanceRequest) (*user.UpdateBalanceResponse, error) {
	logger.Debug("Handling WithdrawBalance request",
		zap.Int64("user_id", req.UserId),
		zap.Int64("amount", req.Amount))

	if err := s.useCase.WithdrawBalance(ctx, req); err != nil {
		logger.Error("Failed to withdraw balance", zap.Error(err))
		return &user.UpdateBalanceResponse{
			Success: false,
			Message: "Failed to withdraw balance",
		}, status.Error(codes.Internal, "failed to withdraw balance")
	}

	return &user.UpdateBalanceResponse{
		Success: true,
		Message: "Balance withdrawn successfully",
	}, nil
}

// UpdatePassword обновляет пароль пользователя
func (s *UserService) UpdatePassword(ctx context.Context, req *user.UpdatePasswordRequest) (*user.UpdatePasswordResponse, error) {
	logger.Debug("Handling UpdatePassword request", zap.Int64("user_id", req.UserId))

	if err := s.useCase.UpdatePassword(ctx, req); err != nil {
		logger.Error("Failed to update password", zap.Error(err))
		return &user.UpdatePasswordResponse{
			Success: false,
			Message: "Failed to update password",
		}, status.Error(codes.Internal, "failed to update password")
	}

	return &user.UpdatePasswordResponse{
		Success: true,
		Message: "Password updated successfully",
	}, nil
}
