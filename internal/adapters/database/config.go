package database

// Config holds the database connection configuration.
type Config struct {
	Host     string `json:"host" mapstructure:"host"`
	Port     int    `json:"port" mapstructure:"port"`
	User     string `json:"user" mapstructure:"user"`
	Password string `json:"password" mapstructure:"password"`
	Name     string `json:"name" mapstructure:"name"`
}
