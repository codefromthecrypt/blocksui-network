package config

import (
	"os"
)

type Config struct {
	BlockContractCID   string
	BlocksCID          string
	Env                string
	HomeDir            string
	Port               string
	ProviderURL        string
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
		BlocksCID:          os.Getenv("BLOCKS_CID"),
		Env:                env,
		HomeDir:            hd,
		ProviderURL:        os.Getenv("PROVIDER_URL"),
		StakingContractCID: os.Getenv("STAKING_CID"),
		Web3Token:          os.Getenv("WEB3STORAGE_TOKEN"),
	}
}
