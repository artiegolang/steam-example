package grpc

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"petProject/db-service/internal/usecase/interfaces"
	"petProject/db-service/pkg/logger"
	"petProject/proto/inventory"
	"petProject/proto/trading"
	"petProject/proto/user"
	"time"
)

type Server struct {
	listener       net.Listener
	server         *grpc.Server
	userService    *UserService
	itemService    *ItemService
	tradingService *TradingService
}

// NewServer создаёт новый экземпляр gRPC сервера, слушающий на указанном порту
func NewServer(port string) (*Server, error) {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %s: %v", port, err)
	}

	// Создаём gRPC сервер с опциями
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(serverInterceptor()),
	}
	s := grpc.NewServer(opts...)

	return &Server{
		listener: listener,
		server:   s,
	}, nil
}

// Middleware для логирования и обработки ошибок
func serverInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		logger.Debug("Incoming request",
			zap.String("method", info.FullMethod),
			zap.Any("request", req))

		resp, err := handler(ctx, req)

		logger.Debug("Request completed",
			zap.String("method", info.FullMethod),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))

		return resp, err
	}
}

// RegisterServices регистрирует все сервисы в gRPC сервере
func (s *Server) RegisterServices(
	userUseCase interfaces.UserUseCase,
	inventoryUseCase interfaces.InventoryUseCase,
	itemUseCase interfaces.ItemUseCase,
	tradingUseCase interfaces.TradeUseCase,
) {
	// Создаём экземпляры сервисов
	s.userService = NewUserService(userUseCase)
	s.itemService = NewItemService(itemUseCase, inventoryUseCase) // Передаём оба use case
	s.tradingService = NewTradingService(tradingUseCase)

	// Регистрируем сервисы в gRPC сервере
	user.RegisterUserServiceServer(s.server, s.userService)
	inventory.RegisterInventoryServer(s.server, s.itemService)
	trading.RegisterTradingServer(s.server, s.tradingService)

	logger.Info("gRPC services registered")
}

// Start запускает gRPC сервер и начинает принимать запросы
func (s *Server) Start() error {
	logger.Info("Starting gRPC server", zap.String("port", s.listener.Addr().String()))

	if err := s.server.Serve(s.listener); err != nil {
		return fmt.Errorf("failed to serve gRPC server: %w", err)
	}

	return nil
}

// Stop останавливает gRPC сервер gracefully
func (s *Server) Stop() {
	logger.Info("Stopping gRPC server")
	s.server.GracefulStop()
}
