package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

func WebhookEncryptionKey() ([]byte, error) {
	configuredKey := os.Getenv("WEBHOOK_ENCRYPTION_KEY")
	if len(configuredKey) == KeyBytes {
		return []byte(configuredKey), nil
	}

	decodedKey, err := base64.RawStdEncoding.DecodeString(configuredKey)
	if err != nil {
		decodedKey, err = base64.StdEncoding.DecodeString(configuredKey)
	}
	if err != nil || len(decodedKey) != KeyBytes {
		return nil, fmt.Errorf("WEBHOOK_ENCRYPTION_KEY must be a 32-byte key or base64-encoded 32-byte key")
	}

	return decodedKey, nil
}

const (
	SecretBytes = 32
	KeyBytes    = 32
	NonceSize   = 12
)

// GenerateSecret generates a cryptographically secure random
// webhook secret and returns it as a base64 string.
func GenerateSecret() (string, error) {
	secret := make([]byte, SecretBytes)

	if _, err := io.ReadFull(rand.Reader, secret); err != nil {
		return "", fmt.Errorf("generate secret: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(secret), nil
}

// Encrypt encrypts plaintext using AES-256-GCM.
//
// key must be exactly 32 bytes.
func Encrypt(plaintext string, key []byte) (string, error) {
	if len(key) != KeyBytes {
		return "", fmt.Errorf("encryption key must be %d bytes", KeyBytes)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	// Every encryption needs a fresh random nonce.
	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}

	// nonce is stored together with the ciphertext.
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	// nonce + ciphertext are encoded together for database storage.
	result := append(nonce, ciphertext...)

	return base64.RawStdEncoding.EncodeToString(result), nil
}

// Decrypt decrypts an AES-256-GCM encrypted string.
func Decrypt(encrypted string, key []byte) (string, error) {
	if len(key) != KeyBytes {
		return "", fmt.Errorf("encryption key must be %d bytes", KeyBytes)
	}

	data, err := base64.RawStdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("decode encrypted data: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()

	if len(data) < nonceSize {
		return "", fmt.Errorf("encrypted data is too short")
	}

	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt secret: %w", err)
	}

	return string(plaintext), nil
}

// Sign creates an HMAC-SHA256 signature for a webhook payload.
// The payload should be the exact raw bytes that will be sent
// in the HTTP request body.
func Sign(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)

	signature := mac.Sum(nil)

	return "sha256=" + base64.RawURLEncoding.EncodeToString(signature)
}

// Verify checks whether a webhook signature is valid.
func Verify(payload []byte, secret string, expectedSignature string) bool {
	actualSignature := Sign(payload, secret)

	return hmac.Equal(
		[]byte(actualSignature),
		[]byte(expectedSignature),
	)
}

// GenerateEncryptionKey generates a random 32-byte AES-256 key.
//
// This key should normally be generated ONCE and stored in an
// environment variable or secret manager. Do NOT generate a new
// key every time the application starts.
func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, KeyBytes)

	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("generate encryption key: %w", err)
	}

	return key, nil
}
