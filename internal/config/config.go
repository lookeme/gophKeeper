package config

import "time"

// Config - Output struct of configuration, used to validate.
type Config struct {
	ListenAddress string `mapstructure:"listen_address"`
	Logging       struct {
		Format string   `mapstructure:"format"`
		Level  string   `mapstructure:"level"`
		Output []string `mapstructure:"output"`
	}

	Database struct {
		Postgres *struct {
			ConnectionString      string        `mapstructure:"connection_string"`
			MaxOpenConnections    int32         `mapstructure:"max_open_connections"`
			MaxIdleConnections    int32         `mapstructure:"max_idle_connections"`
			ConnectionMaxLifetime time.Duration `mapstructure:"connection_max_lifetime"`
		}
	}
}
