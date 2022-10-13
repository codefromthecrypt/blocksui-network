package account

import (
	"blocksui-node/config"
	"crypto/ecdsa"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/contract"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
)

type Account struct {
	Address ethgo.Address
	Client  *jsonrpc.Client
	IP      []byte
	Wallet  *wallet.Key
	AuthSig *AuthSig
}

func (a *Account) Sender() contract.ContractOption {
	return contract.WithSender(a.Wallet)
}

func getIpAddress() (*net.UDPAddr, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	ipAddress := conn.LocalAddr().(*net.UDPAddr)

	return ipAddress, nil
}

func GenerateAccount(homeDir string) (*Account, error) {
	fmt.Println("Generating the Node account...")
	privKey, err := crypto.GenerateKey()

	if err != nil {
		return nil, err
	}

	phrase, err := bip39.NewMnemonic(crypto.FromECDSA(privKey))

	if err != nil {
		return nil, err
	}

	keyfile, err := os.Create(filepath.Join(homeDir, ".bui", "keyfile"))
	if err != nil {
		fmt.Println("file error")
		return nil, err
	}
	keyfile.Close()

	if err := crypto.SaveECDSA(filepath.Join(homeDir, ".bui/keyfile"), privKey); err != nil {
		return nil, err
	}

	fmt.Println("")
	fmt.Println("Your node has generated a new Ethereum wallet that will be used for submitting your stake and receiving rewards. This is a self-custodial wallet meaning that you are responsible for backing up your recovery phrase in case your private key is deleted.\n")
	fmt.Println("Make sure you copy this recovery phrase, write it down on paper, and store it safely. If you lose this phrase and your private keys are deleted, you will not be able to recover any funds held in the wallet.\n")
	fmt.Println("Your recovery phrase is:\n")
	fmt.Println(phrase)
	fmt.Println("")

	ip, err := getIpAddress()
	if err != nil {
		return nil, err
	}

	wallet := wallet.NewKey(privKey)

	return &Account{
		Address: wallet.Address(),
		IP:      []byte(ip.IP),
		Wallet:  wallet,
	}, nil
}

func LoadAccount(c *config.Config) (*Account, error) {
	if _, err := os.Stat(filepath.Join(c.HomeDir, ".bui/keyfile")); err != nil {
		return nil, fmt.Errorf("Keyfile not found")
	}

	client, err := jsonrpc.NewClient(c.ProviderURL)
	if err != nil {
		return nil, err
	}

	privKey, err := crypto.LoadECDSA(filepath.Join(c.HomeDir, ".bui/keyfile"))

	if err != nil {
		return nil, err
	}

	ip, err := getIpAddress()
	if err != nil {
		return nil, err
	}

	wallet := wallet.NewKey(privKey)

	return &Account{
		Address: wallet.Address(),
		Client:  client,
		IP:      []byte(ip.IP),
		Wallet:  wallet,
	}, nil
}

func RecoverAccount(c *config.Config) (*Account, error) {
	if _, err := os.Stat(filepath.Join(c.HomeDir, ".bui/keyfile")); err == nil {
		return nil, fmt.Errorf("Keyfile found. Use LoadAccount instead.")
	}

	var privKey *ecdsa.PrivateKey
	var err error
	if c.PrivateKey != "" {
		privKey, err = crypto.HexToECDSA(c.PrivateKey)
		if err != nil {
			fmt.Println("Failed to parse priv key")
			return nil, err
		}
	} else if c.RecoveryPhrase != "" {
		seed, err := bip39.NewSeedWithErrorChecking(c.RecoveryPhrase, "")
		if err != nil {
			return nil, err
		}
		masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
		if err != nil {
			return nil, err
		}
		privKey, err = wallet.DefaultDerivationPath.Derive(masterKey)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("Private key or Recovery phrase missing")
	}

	keyfile, err := os.Create(filepath.Join(c.HomeDir, ".bui", "keyfile"))
	if err != nil {
		fmt.Println("file error")
		return nil, err
	}
	keyfile.Close()

	if err := crypto.SaveECDSA(filepath.Join(c.HomeDir, ".bui/keyfile"), privKey); err != nil {
		return nil, err
	}
	wallet := wallet.NewKey(privKey)

	client, err := jsonrpc.NewClient(c.ProviderURL)
	if err != nil {
		return nil, err
	}

	ip, err := getIpAddress()
	if err != nil {
		return nil, err
	}

	return &Account{
		Address: wallet.Address(),
		Client:  client,
		IP:      []byte(ip.IP),
		Wallet:  wallet,
	}, nil
}
