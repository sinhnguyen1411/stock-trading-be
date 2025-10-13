package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/database"
	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/grpc_server"
	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/http_gateway"
	"github.com/spf13/viper"
)

type Config struct {
    Env          string              `json:"env" mapstructure:"env"`
    GRPC         grpc_server.Config  `json:"grpc" mapstructure:"grpc"`
    HTTP         http_gateway.Config `json:"http" mapstructure:"http"`
    DB           database.Config     `json:"db" mapstructure:"db"`
    Auth         AuthConfig          `json:"auth" mapstructure:"auth"`
    Notification NotificationConfig  `json:"notification" mapstructure:"notification"`
    Verification VerificationConfig  `json:"verification" mapstructure:"verification"`
}

type AuthConfig struct {
	AccessTokenSecret      string `json:"access_token_secret" mapstructure:"access_token_secret" yaml:"access_token_secret"`
	AccessTokenTTLMinutes  int    `json:"access_token_ttl_minutes" mapstructure:"access_token_ttl_minutes" yaml:"access_token_ttl_minutes"`
	RefreshTokenSecret     string `json:"refresh_token_secret" mapstructure:"refresh_token_secret" yaml:"refresh_token_secret"`
	RefreshTokenTTLMinutes int    `json:"refresh_token_ttl_minutes" mapstructure:"refresh_token_ttl_minutes" yaml:"refresh_token_ttl_minutes"`
	Issuer                 string `json:"issuer" mapstructure:"issuer" yaml:"issuer"`
	Audience               string `json:"audience" mapstructure:"audience" yaml:"audience"`
}

type NotificationConfig struct {
    Kafka KafkaConfig `json:"kafka" mapstructure:"kafka"`
    Email EmailConfig `json:"email" mapstructure:"email"`
}

type KafkaConfig struct {
	Brokers []string `json:"brokers" mapstructure:"brokers"`
	Topic   string   `json:"topic" mapstructure:"topic"`
	GroupID string   `json:"group_id" mapstructure:"group_id"`
}

type EmailConfig struct {
    Provider string     `json:"provider" mapstructure:"provider"`
    SMTP     SMTPConfig `json:"smtp" mapstructure:"smtp"`
    // VerificationURLBase is the base URL used to compose the verification link.
    // Example: http://127.0.0.1:18080/users/verify?token=
    VerificationURLBase string `json:"verification_url_base" mapstructure:"verification_url_base" yaml:"verification_url_base"`
}

type SMTPConfig struct {
    Host     string `json:"host" mapstructure:"host"`
    Port     int    `json:"port" mapstructure:"port"`
    Username string `json:"username" mapstructure:"username"`
    Password string `json:"password" mapstructure:"password"`
    From     string `json:"from" mapstructure:"from"`
    UseTLS   bool   `json:"use_tls" mapstructure:"use_tls"`
}

// VerificationConfig groups the verification token TTL and resend cooldown settings.
type VerificationConfig struct {
    // TokenTTLHours is the time-to-live for verification tokens, in hours.
    TokenTTLHours int `json:"token_ttl_hours" mapstructure:"token_ttl_hours" yaml:"token_ttl_hours"`
    // ResendCooldownSeconds is the minimum delay between resend requests, in seconds.
    ResendCooldownSeconds int `json:"resend_cooldown_seconds" mapstructure:"resend_cooldown_seconds" yaml:"resend_cooldown_seconds"`
}

func loadDefaultConfig() *Config {
    return &Config{
        Env: "local",
        GRPC: grpc_server.Config{
            Host: "0.0.0.0",
            Port: 9090,
        },
        HTTP: http_gateway.Config{
            Host: "0.0.0.0",
            Port: 8080,
        },
        DB: database.Config{
            Host:     "127.0.0.1",
            Port:     3306,
            User:     "stock_user",
            Password: "ps123456",
            Name:     "stock",
        },
        Auth: AuthConfig{
            AccessTokenSecret:      "change-me-in-production-please",
            AccessTokenTTLMinutes:  15,
            RefreshTokenSecret:     "change-me-in-production-too",
            RefreshTokenTTLMinutes: 60 * 24 * 3,
            Issuer:                 "stock-trading-be",
            Audience:               "stock-trading-clients",
        },
        Verification: VerificationConfig{
            TokenTTLHours:          24,
            ResendCooldownSeconds:  60,
        },
        Notification: NotificationConfig{
            Kafka: KafkaConfig{
                Brokers: []string{"localhost:29092"},
                Topic:   "stock.stock.user_outbox_events",
                GroupID: "email-service",
            },
            Email: EmailConfig{
                Provider: "smtp",
                SMTP: SMTPConfig{
                    Host:   "localhost",
                    Port:   1025,
                    From:   "no-reply@example.com",
                    UseTLS: false,
                },
                VerificationURLBase: "http://127.0.0.1:18080/users/verify?token=",
            },
        },
    }
}

func LoadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv()
	/**
	  |-------------------------------------------------------------------------
	  | You should set default config value here
	  | 1. Populate the default value in (Source code)
	  | 2. Then merge from config (YAML) and OS environment
	  |-----------------------------------------------------------------------*/
	// Load default config
	c := loadDefaultConfig()
	configBuffer, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default config: %w", err)
	}

	//1. Populate the default value in (Source code)
	if err := viper.ReadConfig(bytes.NewBuffer(configBuffer)); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	//2. Then merge from config (YAML) and OS environment
	if err := viper.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("failed to merge in config: %w", err)
	}
	// Populate all config again
	err = viper.Unmarshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return c, err
}
