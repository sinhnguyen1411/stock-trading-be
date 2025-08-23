package server

import (
	"database/sql"

	"github.com/sinhnguyen1411/stock-trading-be/cmd/server/config"
	"github.com/sinhnguyen1411/stock-trading-be/internal/adapters/database"
	"github.com/sinhnguyen1411/stock-trading-be/internal/ports"
)

type InfrastructureDependencies struct {
	DB *sql.DB
}

// InitInfrastructure establishes connections to external infrastructure such as
// the database. The connection is attempted synchronously so that subsequent
// components can depend on its availability. When the connection cannot be
// established, the returned struct will contain a nil DB pointer allowing
// callers to gracefully fall back to in-memory repositories.
func InitInfrastructure(cfg *config.Config) (*InfrastructureDependencies, error) {
	if err := database.ConnectDB(cfg.DB); err != nil {
		// Database connection failed; proceed with nil DB so callers can
		// decide to use an alternative implementation.
		return &InfrastructureDependencies{DB: nil}, nil
	}
	return &InfrastructureDependencies{DB: database.DB}, nil
}

// Adapters groups concrete implementations that satisfy the application's
// ports. These can be backed by real infrastructure or in-memory fallbacks.
type Adapters struct {
	UserRepository ports.UserRepository
}

// NewAdapters wires repositories based on available infrastructure
// dependencies. If no database connection is present, an in-memory
// implementation is used to keep the application functional in a degraded
// mode.
func NewAdapters(infra *InfrastructureDependencies) (*Adapters, error) {
	var repo ports.UserRepository
	if infra.DB != nil {
		repo = database.NewMysqlUserRepository()
	} else {
		repo = database.NewInMemoryUserRepository()
	}
	return &Adapters{UserRepository: repo}, nil
}
