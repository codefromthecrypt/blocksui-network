package contracts

import (
	"blocksui-node/config"
	"blocksui-node/ipfs"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/contract"
	"github.com/umbracle/ethgo/jsonrpc"
)

type ContractConfig struct {
	Address      ethgo.Address `json:"address"`
	Abi          *abi.ABI
	ContractName string `json:"contractName"`
}

type Contract struct {
	Address  ethgo.Address
	Abi      *abi.ABI
	Provider *contract.Contract
}

func (c *Contract) Txn(method string, args ...interface{}) (contract.Txn, error) {
	return c.Provider.Txn(method, args...)
}

func (c *Contract) Call(method string, block ethgo.BlockNumber, args ...interface{}) (map[string]interface{}, error) {
	return c.Provider.Call(method, block, args...)
}

type Contracts map[string]Contract

var client *jsonrpc.Client
var contracts Contracts

func LoadContracts(c *config.Config) error {
	if contracts != nil {
		return fmt.Errorf("Already initialized")
	}

	if client == nil {
		newClient, err := jsonrpc.NewClient(c.ProviderURL)
		if err != nil {
			return err
		}

		client = newClient
	}

	res, err := ipfs.Web3Get(c.StakingContractCID, c.Web3Token)
	if err != nil {
		fmt.Println("Web3 Error")
		return err
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("Failed to fetch the ABIs")
	}

	_, fsys, err := res.Files()
	if err != nil {
		return err
	}

	contracts = make(Contracts)

	fs.WalkDir(fsys, "/", func(path string, d fs.DirEntry, err error) error {
		info, _ := d.Info()
		if !info.IsDir() {
			file, err := fsys.Open(path)
			if err != nil {
				return err
			}

			data, err := io.ReadAll(file)
			if err != nil {
				return err
			}

			var cnf ContractConfig
			if err := json.Unmarshal(data, &cnf); err != nil {
				return err
			}

			// fmt.Printf("%s Address: %s\n", cnf.ContractName, cnf.Address)

			contracts[cnf.ContractName] = Contract{
				Address: cnf.Address,
				Abi:     cnf.Abi,
				Provider: contract.NewContract(
					cnf.Address,
					cnf.Abi,
					contract.WithJsonRPC(client.Eth()),
				),
			}
		}

		return err
	})

	return nil
}

func ContractForSender(name string, withSender contract.ContractOption) *Contract {
	c := contracts[name]
	opts := []contract.ContractOption{
		contract.WithJsonRPC(client.Eth()),
		withSender,
	}
	return &Contract{
		Address:  c.Address,
		Abi:      c.Abi,
		Provider: contract.NewContract(c.Address, c.Abi, opts...),
	}
}
