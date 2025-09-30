package grpc_server

import (
	"errors"
	"fmt"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"log/slog"
	"net"
	"syscall"

	_ "github.com/prometheus/client_golang/prometheus"
	"github.com/sinhnguyen1411/stock-trading-be/internal/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Config struct {
	Host string `json:"host" mapstructure:"host" yaml:"host"`
	Port int    `json:"port" mapstructure:"port" yaml:"port"`
}

type Service interface {
	RegisterService(s grpc.ServiceRegistrar)
}

func StartServer(grpcCfg Config, tokenValidator security.AccessTokenManager, services ...Service) (gracefulStop func(), cerr chan error) {
	grpc_prometheus.EnableHandlingTimeHistogram()
	grpcService := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_prometheus.UnaryServerInterceptor,
			AuthUnaryServerInterceptor(tokenValidator),
			RequestValidationUnaryServerInterceptor(),
		),
		grpc.ChainStreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.MaxConcurrentStreams(1000),
		grpc.MaxRecvMsgSize(1024*1024*50), // 50MB
	)
	for _, service := range services {
		service.RegisterService(grpcService)
	}
	reflection.Register(grpcService)
	var cerrChan = make(chan error, 1)

	go func() {
		defer close(cerrChan)

		grpcListener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", grpcCfg.Host, grpcCfg.Port))
		if err != nil {
			if errors.Is(err, syscall.EADDRINUSE) {
				slog.Warn("GRPC PORT IN USE REBINDING", "port", grpcCfg.Port)
				grpcListener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", grpcCfg.Host, 0))
			}
			if err != nil {
				cerrChan <- fmt.Errorf("failed to listen: %v", err)
				return
			}
		}
		slog.Info("GRPC SERVER RUNNING", "addr", grpcListener.Addr().String())
		defer slog.Info("GRPC SERVER STOPPING")

		if err := grpcService.Serve(grpcListener); err != nil {
			cerrChan <- fmt.Errorf("failed to serve: %v", err)
			return
		}
	}()

	return grpcService.GracefulStop, cerrChan
}
