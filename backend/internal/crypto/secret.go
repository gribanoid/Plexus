package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const prefix = "enc:v1:"

// KeyFromString derives a 32-byte AES key from an arbitrary secret string.
func KeyFromString(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

// EncryptString encrypts plaintext with AES-GCM and returns a versioned ciphertext.
func EncryptString(key []byte, plaintext string) (string, error) {
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return prefix + base64.RawStdEncoding.EncodeToString(sealed), nil
}

// DecryptString decrypts a value produced by EncryptString.
// Plaintext values without the enc:v1: prefix are returned unchanged (legacy rows).
func DecryptString(key []byte, value string) (string, error) {
	if !strings.HasPrefix(value, prefix) {
		return value, nil
	}
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes")
	}
	raw, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(value, prefix))
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := raw[:nonceSize], raw[nonceSize:]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
