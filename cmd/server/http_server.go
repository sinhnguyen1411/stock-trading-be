package server

import (
	"fmt"
	"github.com/bqdanh/stock-trading-be/cmd/server/config"

	"github.com/bqdanh/stock-trading-be/internal/adapters/server/http_gateway"
	usersgw "github.com/bqdanh/stock-trading-be/internal/adapters/server/http_gateway/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewHTTPGatewayServices(cfg config.Config, _ *InfrastructureDependencies) ([]http_gateway.GrpcGatewayServices, error) {
	grpcServerAddr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)
	grpcServerConn, err := grpc.NewClient(grpcServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(),
	)
	if err != nil {
		return nil, fmt.Errorf("fail to dial gRPC server(%s): %w", grpcServerAddr, err)
	}

	// new http gateway services
	userHttpGwService := usersgw.NewUserGatewayService(grpcServerConn)

	return []http_gateway.GrpcGatewayServices{
		userHttpGwService,
	}, nil
}
