package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// CryptoService handles encryption and decryption of secrets.
type CryptoService interface {
	Encrypt(plaintext []byte) (string, error)
	Decrypt(ciphertext string) ([]byte, error)
}

// AESGCMEncryption implements CryptoService using AES-GCM.
// In a real production system, this would interact with a KMS.
type AESGCMEncryption struct {
	masterKey []byte
}

// NewAESGCMEncryption creates a new instance with a provided master key.
// masterKey must be 32 bytes for AES-256.
func NewAESGCMEncryption(masterKey []byte) (*AESGCMEncryption, error) {
	if len(masterKey) != 32 {
		return nil, errors.New("master key must be 32 bytes")
	}
	return &AESGCMEncryption{masterKey: masterKey}, nil
}

// Encrypt encrypts data using AES-GCM with a random nonce.
// It returns a base64 encoded string containing nonce+ciphertext.
func (a *AESGCMEncryption) Encrypt(plaintext []byte) (string, error) {
	block, err := aes.NewCipher(a.masterKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64 encoded string.
func (a *AESGCMEncryption) Decrypt(encodedCiphertext string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(a.masterKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
