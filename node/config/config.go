package config

import "os"

type Config struct {
	BlockContractCID   string
	Env                string
	HomeDir            string
	Port               string
	ProviderURL        string
	StakingContractCID string
	Web3Token          string
}

func New(
	env string,
	homeDir string,
	port string,
) *Config {
	return &Config{
		os.Getenv("BLOCK_NFT_CID"),
		env,
		homeDir,
		port,
		os.Getenv("PROVIDER_URL"),
		os.Getenv("STAKING_CID"),
		os.Getenv("WEB3STORAGE_TOKEN"),
	}
}
