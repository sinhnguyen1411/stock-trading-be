package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sinhnguyen1411/stock-trading-be/cmd/server/config"
	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/notification"
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

	notificationShutdown := func() {}
	if shouldStartNotification(cfg.Notification) && adapters.OutboxRepository != nil {
		emailSender, err := buildEmailSender(cfg.Notification.Email)
		if err != nil {
			return fmt.Errorf("failed to init email sender: %w", err)
		}

		notifService, err := notification.NewService(notification.Options{
			Brokers: cfg.Notification.Kafka.Brokers,
			Topic:   cfg.Notification.Kafka.Topic,
			GroupID: cfg.Notification.Kafka.GroupID,
		}, adapters.OutboxRepository, emailSender)
		if err != nil {
			return fmt.Errorf("failed to init notification service: %w", err)
		}

		notifyCtx, notifyCancel := context.WithCancel(context.Background())
		go func() {
			if err := notifService.Start(notifyCtx); err != nil && !errors.Is(err, context.Canceled) {
				cerr <- fmt.Errorf("notification service error: %w", err)
			}
		}()

		notificationShutdown = func() {
			notifyCancel()
			if err := notifService.Close(); err != nil {
				slog.Warn("notification service close error", "error", err)
			}
		}
	} else {
		slog.Info("NOTIFICATION SERVICE DISABLED", "reason", "missing configuration")
	}
	defer notificationShutdown()

	slog.Info("SERVER STARTED")
	<-stop
	slog.Info("SERVER STOPPING")
	return nil
}

func shouldStartNotification(cfg config.NotificationConfig) bool {
	return len(cfg.Kafka.Brokers) > 0 && cfg.Kafka.Topic != ""
}

func buildEmailSender(cfg config.EmailConfig) (notification.EmailSender, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		provider = "smtp"
	}
    switch provider {
    case "smtp":
        smtpSender, err := notification.NewSMTPSender(notification.SMTPSenderConfig{
            Host:                cfg.SMTP.Host,
            Port:                cfg.SMTP.Port,
            Username:            cfg.SMTP.Username,
            Password:            cfg.SMTP.Password,
            From:                cfg.SMTP.From,
            UseTLS:              cfg.SMTP.UseTLS,
            VerificationURLBase: strings.TrimSpace(strings.ReplaceAll(cfg.VerificationURLBase, "\"", "")),
        })
        if err != nil {
            return nil, err
        }
        return smtpSender, nil
	case "noop":
		return notification.NewNoopSender(), nil
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", cfg.Provider)
	}
}
