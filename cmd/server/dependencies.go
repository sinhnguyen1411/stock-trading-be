package server

import (
	"database/sql"
	"github.com/sinhnguyen1411/stock-trading-be/cmd/server/config"
)

type InfrastructureDependencies struct {
	db *sql.DB
}

func InitInfrastructure(cfg *config.Config) (*InfrastructureDependencies, error) {
	return &InfrastructureDependencies{}, nil
}

type Adapters struct {
}

func NewAdapters(cfg *config.Config, infra *InfrastructureDependencies) (*Adapters, error) {
	return &Adapters{}, nil
}
