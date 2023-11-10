package redis

type Config struct {
	addr     string `toml:"addr"`
	password string `toml:"password"`
}

func NewConfig() *Config {
	return &Config{}
}
