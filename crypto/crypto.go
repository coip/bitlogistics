package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

const Lennysecret = "(° ͜ʖ °) ~ > (͡~ ͜ʖ ͡°)"

var c cipher.Block

func init() {
	var key = []byte(os.Getenv("KEY"))
	if len(key) < 1 {
		key = []byte(lennysecret)
	}
	var err error
	c, err = aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
}

//Decrypt uses "crypto/aes" to gcm.Open() && return the result in a new bytes.Buffer pointer
func Decrypt(input []byte) *bytes.Buffer {
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		fmt.Println(err)
	}

	nonceSize := gcm.NonceSize()
	if len(input) < nonceSize {
		fmt.Println(err)
	}

	nonce, input := input[:nonceSize], input[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, input, nil)
	if err != nil {
		fmt.Println(err)
	}
	return bytes.NewBuffer(plaintext)
}

//Encrypt uses "crypto/aes" to gcm.Seal() && return the result in a new bytes.Buffer pointer
func Encrypt(input []byte) *bytes.Buffer {

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		fmt.Println(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println(err)
	}

	return bytes.NewBuffer(gcm.Seal(nonce, nonce, input, nil))
}
