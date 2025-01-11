package grpc

import (
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"petProject/db-service/internal/usecase/interfaces"
	"petProject/db-service/pkg/logger"
	"petProject/proto/inventory"
	"petProject/proto/trading"
	"petProject/proto/user"
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

	// Создаём новый gRPC сервер
	s := grpc.NewServer()

	return &Server{
		listener: listener,
		server:   s,
	}, nil
}

// RegisterServices регистрирует все сервисы в gRPC сервере
func (s *Server) RegisterServices(
	userUseCase interfaces.UserUseCase,
	inventoryUseCase interfaces.InventoryUseCase,
	itemUseCase interfaces.ItemUseCase,
	tradingUseCase interfaces.TradingUseCase, // Предполагается, что интерфейс называется TradingUseCase
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
