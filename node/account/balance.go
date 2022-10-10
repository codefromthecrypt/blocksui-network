package account

import (
	"blocksui-node/contracts"
	"fmt"
	"math/big"

	"github.com/umbracle/ethgo"
)

func (a *Account) Balance() (*big.Int, error) {
	return a.Client.Eth().GetBalance(a.Address, ethgo.Latest)
}

func (a *Account) StakeBalance() (*big.Int, error) {
	return contracts.StakeBalance(a.Address)
}

func (a *Account) VerifyStake() bool {
	cost, err := contracts.StakingCost()
	if err != nil {
		fmt.Printf("[contracts]\t%v\n", err)
		return false
	}

	balance, err := contracts.StakeBalance(a.Address)
	if err != nil {
		fmt.Printf("[contracts]\t%v\n", err)
		return false
	}

	return balance.Cmp(cost) != -1
}
