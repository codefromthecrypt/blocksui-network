package lit

import (
	"blocksui-node/account"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

type EncryptedKeyParams struct {
	AuthSig               *account.AuthSig       `json:"authSig"`
	Chain                 string                 `json:"chain"`
	EvmContractConditions []EvmContractCondition `json:"evmContractConditions"`
	ToDecrypt             string                 `json:"toDecrypt"`
}

func (c *Client) GetEncryptionKey(
	params EncryptedKeyParams,
) ([]byte, error) {
	if !c.Ready {
		return nil, fmt.Errorf("LitClient: not ready")
	}

	ch := make(chan DecryptResMsg)

	for url := range c.ConnectedNodes {
		go GetDecryptionShare(url, params, c, ch)
	}

	shares := make([]DecryptionShareResponse, 0)
	count := 0
	for resp := range ch {
		if resp.Err != nil || resp.Share.ErrorCode != "" {
			if resp.Err != nil {
				fmt.Println(resp.Err)
			} else if resp.Share.Message != "" {
				fmt.Println(resp.Share.Message)
			}
		} else if resp.Share.Status == "fulfilled" || resp.Share.Result == "success" {
			shares = append(shares, *resp.Share)
		}
		count++

		if count >= len(c.ConnectedNodes) {
			break
		}
	}

	if len(shares) < int(c.MinimumNodeCount) {
		return nil, fmt.Errorf("LitClient: failed to retrieve enough shares")
	}

	sort.SliceStable(shares, func(i, j int) bool {
		return shares[i].ShareIndex < shares[j].ShareIndex
	})

	return ThresholdDecrypt(shares, params.ToDecrypt, c.NetworkPubKeySet)
}

func (c *Client) SaveEncryptionKey(
	symmetricKey []byte,
	authSig account.AuthSig,
	authConditions []EvmContractCondition,
	chain string,
) (string, error) {
	subPubKey, err := hex.DecodeString(c.SubnetPubKey)
	if err != nil {
		return "", err
	}

	key, err := ThresholdEncrypt(subPubKey, symmetricKey)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	hash.Write(key)
	hashStr := hex.EncodeToString(hash.Sum(nil))

	condJson, err := json.Marshal(authConditions)
	if err != nil {
		return "", err
	}

	cHash := sha256.New()
	cHash.Write(condJson)
	cHashStr := hex.EncodeToString(cHash.Sum(nil))

	ch := make(chan SaveCondMsg)

	for url := range c.ConnectedNodes {
		go StoreEncryptionConditionWithNode(
			url,
			SaveCondParams{
				Key:       hashStr,
				Val:       cHashStr,
				AuthSig:   authSig,
				Chain:     chain,
				Permanent: 1,
			},
			c,
			ch,
		)
	}

	count := 0
	var e error
	for msg := range ch {
		if msg.Err != nil || msg.Response == nil {
			fmt.Printf("Failed to store condition %v\n", msg.Err)
			e = msg.Err
		}
		count++

		if count >= len(c.ConnectedNodes) {
			break
		}
	}

	if e != nil {
		return "", e
	}

	return hex.EncodeToString(key), nil
}
