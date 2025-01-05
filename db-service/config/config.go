package config

type Config struct {
	Postgres PostgresConfig
	Log      LogConfig
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
