package config

type Config struct {
	Mode string
	Port string
}

func New(mode string, port string) *Config {
	return &Config{mode, port}
}
