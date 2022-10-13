package config

import (
	"os"
)

type Config struct {
	ChainName       string
	ContractsCID    string
	Env             string
	HomeDir         string
	LitVersion      string
	MinLitNodeCount uint8
	NetworkName     string
	Port            string
	PrimitivesCID   string
	PrivateKey      string
	ProviderURL     string
	RecoveryPhrase  string
	Web3Token       string
}

func (c *Config) Chain() string {
	if c.NetworkName == "mainnet" {
		return c.ChainName
	} else {
		return c.NetworkName
	}
}

func New(env string) *Config {
	hd, err := os.UserHomeDir()
	if err != nil {
		hd = "/"
	}

	return &Config{
		ChainName:       os.Getenv("CHAIN_NAME"),
		ContractsCID:    os.Getenv("CONTRACTS_CID"),
		Env:             env,
		HomeDir:         hd,
		LitVersion:      os.Getenv("LIT_VERSION"),
		MinLitNodeCount: 6,
		NetworkName:     os.Getenv("NETWORK_NAME"),
		PrimitivesCID:   os.Getenv("PRIMITIVES_CID"),
		PrivateKey:      os.Getenv("PRIVATE_KEY"),
		ProviderURL:     os.Getenv("PROVIDER_URL"),
		RecoveryPhrase:  os.Getenv("RECOVERY_PHRASE"),
		Web3Token:       os.Getenv("WEB3STORAGE_TOKEN"),
	}
}
