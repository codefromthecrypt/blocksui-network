package lit

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func PKCS7Padding(plaintext []byte) []byte {
	padding := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padtext...)
}

func Prng(length uint64) []byte {
	values := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, values); err != nil {
		panic(err)
	}

	return values
}

func AesEncrypt(key []byte, plaintext []byte) (ciphertext []byte) {
	padded := PKCS7Padding(plaintext)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}

	ciphertext = make([]byte, aes.BlockSize+len(padded))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], padded)

	return
}

func ThresholdEncrypt(subPubKey []byte, message []byte) ([]byte, error) {
	wasm, err := NewWasmInstance(context.Background())
	if err != nil {
		fmt.Println("Failed to get wasm")
		return nil, err
	}

	rngSize, err := wasm.Call("get_rng_values_size")
	if err != nil {
		fmt.Println("Failed to get rng size")
		return nil, err
	}

	rngValues := Prng(rngSize.(uint64))

	for i, value := range rngValues {
		if _, err := wasm.Call("set_rng_value", uint64(i), uint64(value)); err != nil {
			fmt.Println("Failed to set RNG value")
			return nil, err
		}
	}

	for i, b := range subPubKey {
		if _, err := wasm.Call("set_pk_byte", uint64(i), uint64(b)); err != nil {
			fmt.Printf("Failed to set pub key byte '%x' at %d\n", b, i)
			return nil, err
		}
	}

	for i, b := range message {
		if _, err := wasm.Call("set_msg_byte", uint64(i), uint64(b)); err != nil {
			fmt.Println("Failed to set message byte")
			return nil, err
		}
	}

	ctSize, err := wasm.Call("encrypt", uint64(len(message)))
	ciphertext := make([]byte, ctSize.(uint64))
	for i := uint64(0); i < ctSize.(uint64); i++ {
		b, err := wasm.Call("get_ct_byte", i)
		if err != nil {
			fmt.Println("Failed to get ciphertext byte")
			return nil, err
		}
		ciphertext = append(ciphertext, byte(b.(uint64)))
	}

	return ciphertext, nil
}
