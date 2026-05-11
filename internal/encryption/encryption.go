package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

type Encryption struct {
	Key []byte
}

func (e *Encryption) Encrypt(text string) (string, error) {
	plaintext := []byte(text)

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aead.Seal(nonce, nonce, plaintext, nil)
	return base64.URLEncoding.EncodeToString(cipherText), nil
}

func (e *Encryption) Decrpyt(cryptoText string) (string, error) {
	cipherText, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aead.NonceSize()
	if len(cipherText) < nonceSize {
		return "", err
	}

	nonce, payload := cipherText[:nonceSize], cipherText[nonceSize:]
	plainText, err := aead.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}
