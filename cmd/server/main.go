package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sinhnguyen1411/stock-trading-be/cmd/server/config"
	grpcadapter "github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/grpc_server"
	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/http_gateway"
	"github.com/urfave/cli/v2"
)

var StartServerCmd = &cli.Command{
	Name:   "server",
	Usage:  "run http server",
	Action: StartServerAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "config",
			Aliases:     []string{"c"},
			Usage:       "Load configuration from file path`",
			DefaultText: "./cmd/server/config/local.yaml",
			Value:       "./cmd/server/config/local.yaml",
			Required:    false,
		},
	},
}

func StartServerAction(cmdCLI *cli.Context) error {
	cfgPath := cmdCLI.String("config")
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config from path\"%s\": %w", cfgPath, err)
	}
	return StartHTTPServer(cfg)
}

func StartHTTPServer(cfg *config.Config) error {
	if cfg.Env == "local" {
		shadow := *cfg
		shadow.Auth.AccessTokenSecret = "***"
		shadow.Auth.RefreshTokenSecret = "***"
		bs, err := json.Marshal(shadow)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}
		slog.Info("SERVER START CONFIG", "config", string(bs))
	}

	accessTokens, refreshTokens, err := buildTokenManagers(cfg.Auth)
	if err != nil {
		return fmt.Errorf("failed to init token managers: %w", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	cerr := make(chan error)
	go func() {
		for _err := range cerr {
			slog.Error("SERVER ERROR", "error", _err)
			stop <- syscall.SIGTERM
		}
	}()

	infra, err := InitInfrastructure(cfg)
	if err != nil {
		return fmt.Errorf("failed to init infrastructure: %w", err)
	}

	adapters, err := NewAdapters(infra)
	if err != nil {
		return fmt.Errorf("failed to new adapters: %w", err)
	}

	grpcServices, err := NewGrpcServices(*cfg, accessTokens, refreshTokens, infra, adapters)
	if err != nil {
		return fmt.Errorf("failed to new grpc services: %w", err)
	}

	grpcStop, cgrpcerr := grpcadapter.StartServer(cfg.GRPC, accessTokens, grpcServices...)
	go func() {
		for gerr := range cgrpcerr {
			cerr <- fmt.Errorf("grpc server error: %w", gerr)
		}
	}()
	defer grpcStop()

	httpgwServices, err := NewHTTPGatewayServices(*cfg, infra)
	if err != nil {
		return fmt.Errorf("failed to new http gateway services: %w", err)
	}

	httpStop, cherr := http_gateway.StartServer(cfg.HTTP, httpgwServices, nil)
	go func() {
		for herr := range cherr {
			cerr <- fmt.Errorf("http server error: %w", herr)
		}
	}()
	defer httpStop()

	slog.Info("SERVER STARTED")
	<-stop
	slog.Info("SERVER STOPPING")
	return nil
}
