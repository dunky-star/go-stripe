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

// Encrypt seals plaintext with AES-GCM (authenticated). Output is base64url(nonce || ciphertext).
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

	out := aead.Seal(nonce, nonce, plaintext, nil)
	return base64.URLEncoding.EncodeToString(out), nil
}

// Decrypt reverses Encrypt: base64url decode, then GCM open.
func (e *Encryption) Decrypt(cryptoText string) (string, error) {
	raw, err := base64.URLEncoding.DecodeString(cryptoText)
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

	ns := aead.NonceSize()
	if len(raw) < ns {
		return "", err
	}

	nonce, payload := raw[:ns], raw[ns:]
	plain, err := aead.Open(nil, nonce, payload, nil)
	if err != nil {
		return "", err
	}

	return string(plain), nil
}
