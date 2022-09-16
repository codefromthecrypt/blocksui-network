package contracts

import (
	"fmt"
	"math/big"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/contract"
)

func toHexString(bytes []byte) string {
	hexStr := "0x"
	for _, num := range bytes {
		hexStr += fmt.Sprintf("%x", num)
	}

	return hexStr
}

func StakingCost() (*big.Int, error) {
	ctr := contracts["BUINodeStaking"]

	res, err := ctr.Call("stakingCost", ethgo.Latest)
	if err != nil {
		return nil, err
	}

	return res["0"].(*big.Int), nil
}

func StakeBalance(address ethgo.Address) (*big.Int, error) {
	ctr := contracts["BUINodeStaking"]

	res, err := ctr.Call("balance", ethgo.Latest, address)
	if err != nil {
		return nil, err
	}

	return res["0"].(*big.Int), nil
}

func Verify(address ethgo.Address) (bool, error) {
	ctr := contracts["BUINodeStaking"]

	res, err := ctr.Call("verify", ethgo.Latest, address)
	if err != nil {
		return false, err
	}

	return res["0"].(bool), nil
}

func CalcStake(address ethgo.Address) (*big.Int, error) {
	cost, err := StakingCost()
	if err != nil {
		return nil, err
	}

	balance, err := StakeBalance(address)
	if err != nil {
		return nil, err
	}

	return cost.Sub(cost, balance), nil
}

func Register(sender contract.ContractOption, ip []byte, stake *big.Int) bool {
	ctr := ContractForSender("BUINodeStaking", sender)
	ipHash := toHexString(ip)

	txn, err := ctr.Txn("register", ipHash)
	if err != nil {
		fmt.Println(err)
		return false
	}

	txn.WithOpts(&contract.TxnOpts{
		Value: stake,
	})

	if err := txn.Do(); err != nil {
		fmt.Println(err)
		return false
	}

	receipt, err := txn.Wait()
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Printf("Successfully staked: %s\n", stake)
	fmt.Printf("Transaction Hash: %s\n", receipt.TransactionHash)

	return true
}

func Unregister(sender contract.ContractOption) bool {
	ctr := ContractForSender("BUINodeStaking", sender)

	txn, err := ctr.Txn("unregister")
	if err != nil {
		fmt.Println(err)
		return false
	}

	if err := txn.Do(); err != nil {
		fmt.Println(err)
		return false
	}

	receipt, err := txn.Wait()
	if err != nil {
		fmt.Println(err)
		return false
	}

	fmt.Println("Successfully unstaked")
	fmt.Printf("Transaction Hash: %s\n", receipt.TransactionHash)

	return true
}
