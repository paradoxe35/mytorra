package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// Encryptor interface for encrypting and decrypting data
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AESEncryptor implements AES-256-GCM encryption
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor creates a new AES encryptor with the given key
func NewAESEncryptor(key string) Encryptor {
	// Use SHA-256 to ensure the key is exactly 32 bytes for AES-256
	hash := sha256.Sum256([]byte(key))
	return &AESEncryptor{
		key: hash[:],
	}
}

// Encrypt encrypts the plaintext using AES-256-GCM
func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create a nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to create nonce: %w", err)
	}

	// Encrypt the data
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return base64 encoded for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts the ciphertext using AES-256-GCM
func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Extract the nonce from the beginning
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}
