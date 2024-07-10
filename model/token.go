package model

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	"github.com/xlzd/gotp"
	"golang.org/x/crypto/pbkdf2"
)

type Token struct {
	AccountType             string      `json:"account_type"`
	Digits                  int         `json:"digits"`
	EncryptedSeed           string      `json:"encrypted_seed"`
	Issuer                  interface{} `json:"issuer"`
	KeyDerivationIterations int         `json:"key_derivation_iterations"`
	Logo                    interface{} `json:"logo"`
	Name                    string      `json:"name"`
	OriginalName            string      `json:"original_name"`
	PasswordTimestamp       int         `json:"password_timestamp"`
	Salt                    string      `json:"salt"`
	UniqueId                string      `json:"unique_id"`
}

// decrypt the EncryptedSeed and return it
func (t Token) Decrypt(passphrase []byte) ([]byte, error) {
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	seed, err := base64.StdEncoding.DecodeString(t.EncryptedSeed)
	if err != nil {
		return nil, err
	}

	iterations := 1000
	if t.KeyDerivationIterations != 0 {
		iterations = t.KeyDerivationIterations
	}

	key := pbkdf2.Key(bytes.TrimSpace(passphrase), []byte(t.Salt), iterations, 32, sha1.New)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(seed, seed)

	paddingByte := seed[len(seed)-1]
	paddingLength := int(paddingByte)

	if int(paddingLength) > block.BlockSize() {
		// bad pkcs5 padding means invalid password
		return nil, fmt.Errorf("incorrect password")
	}
	for _, b := range seed[len(seed)-int(paddingLength):] {
		// if any of the padding bytes dont match its an invalid password
		if b != paddingByte {
			return nil, fmt.Errorf("incorrect password")
		}
	}

	// decode and return it
	return seed[:len(seed)-int(paddingLength)], nil
}

func (t Token) TOTP(passphrase []byte) (*gotp.TOTP, error) {
	seed, err := t.Decrypt(passphrase)
	if err != nil {
		return nil, err
	}
	return gotp.NewTOTP(string(seed), t.Digits, 30, nil), nil
}
