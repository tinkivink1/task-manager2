package apiserver

import "TaskManager/internal/storage/postgres"

type Config struct {
	BindAddr 	string 			`toml:"bind_addr"`
	LogLevel 	string 			`toml:"log_level"`
	JWTSecret 	string			`toml:"jwt_secret"`
	Storage    	*postgres.Config
}

func NewConfig() *Config {
	return &Config{
		BindAddr: ":8080",
		LogLevel: "debug",
		Storage: postgres.NewConfig(),	
	}
}