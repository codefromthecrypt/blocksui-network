package config

import (
	"os"
)

type Config struct {
	BlockContractCID   string
	Env                string
	HomeDir            string
	Port               string
	PrimitivesCID      string
	ProviderURL        string
	RecoveryPhrase     string
	StakingContractCID string
	Web3Token          string
}

func (c *Config) WithPort(port string) *Config {
	c.Port = port
	return c
}

func New(env string) *Config {
	hd, err := os.UserHomeDir()
	if err != nil {
		hd = "/"
	}

	return &Config{
		BlockContractCID:   os.Getenv("BLOCK_NFT_CID"),
		Env:                env,
		HomeDir:            hd,
		PrimitivesCID:      os.Getenv("PRIMITIVES_CID"),
		ProviderURL:        os.Getenv("PROVIDER_URL"),
		RecoveryPhrase:     os.Getenv("RECOVERY_PHRASE"),
		StakingContractCID: os.Getenv("STAKING_CID"),
		Web3Token:          os.Getenv("WEB3STORAGE_TOKEN"),
	}
}
