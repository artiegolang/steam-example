package config

type Config struct {
	Log      LogConfig
	Postgres PostgresConfig
}

type LogConfig struct {
	Development bool
	Level       string // "debug", "info", "error"
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	MaxConns int32
	MinConns int32
}
