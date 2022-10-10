package lit

import (
	"blocksui-node/account"
	"context"
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

	wasm, err := NewWasmInstance(context.Background())
	if err != nil {
		fmt.Println("GetEncryptionKey: failed to get wasm")
		return nil, err
	}

	for i, share := range shares {
		if _, err := wasm.Call("set_share_indexes", uint64(i), uint64(share.ShareIndex)); err != nil {
			fmt.Println("GetEncryptionKey: set_share_indexes failed")
			return nil, err
		}

		shareBytes, err := hex.DecodeString(share.DecryptionShare)
		if err != nil {
			return nil, err
		}

		for idx, b := range shareBytes {
			if _, err := wasm.Call("set_decryption_shares_byte", uint64(idx), uint64(i), uint64(b)); err != nil {
				fmt.Println("GetEncryptionKey: set_decryption_shares_byte failed")
				return nil, err
			}
		}
	}

	pkSetBytes, err := hex.DecodeString(c.NetworkPubKeySet)
	if err != nil {
		return nil, err
	}

	for idx, b := range pkSetBytes {
		if _, err := wasm.Call("set_mc_byte", uint64(idx), uint64(b)); err != nil {
			fmt.Println("GetEncryptionKey: set_mc_byte failed")
			return nil, err
		}
	}

	ctBytes, err := hex.DecodeString(params.ToDecrypt)
	if err != nil {
		return nil, err
	}

	for idx, b := range ctBytes {
		if _, err := wasm.Call("set_ct_byte", uint64(idx), uint64(b)); err != nil {
			fmt.Println("GetEncryptionKey: set_ct_byte failed")
			return nil, err
		}
	}

	size, err := wasm.Call("combine_decryption_shares", uint64(len(shares)), uint64(len(pkSetBytes)), uint64(len(ctBytes)))
	if err != nil {
		fmt.Println("GetEncryptionKey: combine_decryption_shares failed")
		return nil, err
	}

	result := make([]byte, 0)

	for i := 0; i < int(size.(uint64)); i++ {
		b, err := wasm.Call("get_msg_byte", uint64(i))
		if err != nil {
			fmt.Println("GetEncryptionKey: get_msg_byte failed")
			return nil, err
		}

		result = append(result, byte(b.(uint64)))
	}

	wasm.Close()

	return result, nil
}

func (c *Client) SaveEncryptionKey(
	symmetricKey []byte,
	authSig account.AuthSig,
	authConditions []EvmContractCondition,
	chain string,
) (string, error) {
	// fmt.Printf("SubnetKey: %s\n", c.SubnetPubKey)
	subPubKey, err := hex.DecodeString(c.SubnetPubKey)
	if err != nil {
		return "", err
	}

	key, err := ThresholdEncrypt(subPubKey, symmetricKey)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(key)
	hashStr := fmt.Sprintf("%x", hash)

	condJson, err := json.Marshal(authConditions)
	if err != nil {
		return "", err
	}

	cHash := sha256.Sum256(condJson)
	cHashStr := fmt.Sprintf("%x", cHash)

	ch := make(chan SaveCondMsg)

	for url := range c.ConnectedNodes {
		go StoreEncryptionConditionWithNode(
			url,
			SaveCondParams{
				Key:     hashStr,
				Val:     cHashStr,
				AuthSig: authSig,
				Chain:   chain,
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
		return "", err
	}

	return hashStr, nil
}
