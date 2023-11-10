package apiserver

type Config struct {
	BindAddr  string `toml:"bind_addr"`
	LogLevel  string `toml:"log_level"`
	JWTSecret string `toml:"jwt_secret"`
	Caching   bool   `toml:"caching_responses"`
}

func NewConfig() *Config {
	return &Config{
		BindAddr: ":8080",
		LogLevel: "debug",
	}
}
