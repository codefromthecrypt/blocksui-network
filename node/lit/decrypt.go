package lit

import (
	"crypto/aes"
	"crypto/cipher"
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
