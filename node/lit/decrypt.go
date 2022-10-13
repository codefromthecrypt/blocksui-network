package lit

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func PKCS7UnPadding(plaintext []byte) []byte {
	length := len(plaintext)
	unpadding := int(plaintext[length-1])
	return plaintext[:(length - unpadding)]
}

func AesDecrypt(key []byte, ciphertext []byte) (plaintext []byte) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	iv := ciphertext[:aes.BlockSize]
	plaintext = make([]byte, len(ciphertext)-aes.BlockSize)

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext[aes.BlockSize:])

	return PKCS7UnPadding(plaintext)
}

type DecryptionShareResponse struct {
	DecryptionShare string `json:"decryptionShare"`
	ErrorCode       string `json:"errorCode"`
	Message         string `json:"message"`
	Result          string `json:"result"`
	ShareIndex      uint8  `json:"shareIndex"`
	Status          string `json:"status"`
}

type DecryptResMsg struct {
	Share *DecryptionShareResponse
	Err   error
}

func closeWithError(msg string, ch chan DecryptResMsg) {
	ch <- DecryptResMsg{nil, fmt.Errorf(msg)}
	close(ch)
}

func GetDecryptionShare(url string, params EncryptedKeyParams, c *Client, ch chan DecryptResMsg) {
	reqBody, err := json.Marshal(params)
	if err != nil {
		closeWithError("LitClient:Key: failed to marshal req body.", ch)
		return
	}

	resp, err := c.NodeRequest(url+"/web/encryption/retrieve", reqBody)
	if err != nil {
		closeWithError("LitClient:Key: Request to nodes failed.", ch)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		closeWithError("LitClient:Key: Failed to read response.", ch)
		return
	}

	share := &DecryptionShareResponse{}
	if err := json.Unmarshal(body, share); err != nil {
		closeWithError("LitClient:Key: Failed unmarshal the response.", ch)
		return
	}

	ch <- DecryptResMsg{share, nil}
}

func ThresholdDecrypt(shares []DecryptionShareResponse, ciphertext, netPubKeySet string) ([]byte, error) {
	wasm, err := NewWasmInstance(context.Background())
	if err != nil {
		fmt.Println("GetEncryptionKey: failed to get wasm")
		return nil, err
	}
	defer wasm.Close()

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

	pkSetBytes, err := hex.DecodeString(netPubKeySet)
	if err != nil {
		return nil, err
	}

	for idx, b := range pkSetBytes {
		if _, err := wasm.Call("set_mc_byte", uint64(idx), uint64(b)); err != nil {
			fmt.Println("GetEncryptionKey: set_mc_byte failed")
			return nil, err
		}
	}

	ctBytes, err := hex.DecodeString(ciphertext)
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

	si := int(size.(uint64))
	result := make([]byte, 0, si)

	for i := 0; i < si; i++ {
		b, err := wasm.Call("get_msg_byte", uint64(i))
		if err != nil {
			fmt.Println("GetEncryptionKey: get_msg_byte failed")
			return nil, err
		}

		result = append(result, byte(b.(uint64)))
	}

	return result, nil
}
