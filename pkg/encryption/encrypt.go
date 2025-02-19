package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
)

type KeyEncryptor struct {
	secretKey []byte
}

func NewKeyEncryptor(secretKey string) *KeyEncryptor {
	key, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return &KeyEncryptor{secretKey: key}
}

func (k *KeyEncryptor) Encrypt(apiKey string) (string, error) {
	block, err := aes.NewCipher(k.secretKey)
	if err != nil {
		return "", err
	}

	// Create GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(apiKey), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (k *KeyEncryptor) Decrypt(encryptedKey string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(k.secretKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
