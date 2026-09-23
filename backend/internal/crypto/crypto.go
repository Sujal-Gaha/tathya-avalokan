package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"sync"
)

var (
	once       sync.Once
	derivedKey [32]byte
	isFallback bool
	initErr    error
)

// InitCipherKey derives a 32-byte key from TATHYA_ENCRYPTION_KEY or the fallback APP_SECRET_KEY.
func InitCipherKey(encryptionKey, fallbackKey string) {
	once.Do(func() {
		if encryptionKey == "" {
			encryptionKey = os.Getenv("TATHYA_ENCRYPTION_KEY")
		}
		if encryptionKey == "" {
			isFallback = true
			log.Println("⚠️  TATHYA_ENCRYPTION_KEY is not set. Falling back to a deterministic DEVELOPMENT key. This is INSECURE and must never be used in production.")
			if fallbackKey == "" {
				fallbackKey = os.Getenv("APP_SECRET_KEY")
			}
			if fallbackKey == "" {
				fallbackKey = "tathya-avalokan-default-dev-secret-key-32b"
			}
			derivedKey = sha256.Sum256([]byte(fallbackKey))
		} else {
			isFallback = false
			derivedKey = sha256.Sum256([]byte(encryptionKey))
		}
	})
}

// ResetCipherKeyForTest allows tests to reset the cipher key initialization.
func ResetCipherKeyForTest() {
	once = sync.Once{}
	isFallback = false
	initErr = nil
}

// IsFallbackKey returns true if the dev fallback key is in use.
func IsFallbackKey() bool {
	return isFallback
}

// EncryptCredentials encrypts a credential map into a base64-encoded AES-256-GCM ciphertext.
func EncryptCredentials(credentials map[string]string) (string, error) {
	InitCipherKey("", "")

	block, err := aes.NewCipher(derivedKey[:])
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

	plaintext, err := json.Marshal(credentials)
	if err != nil {
		return "", err
	}

	// Seal appends ciphertext + tag to nonce
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// DecryptCredentials decrypts a base64-encoded AES-256-GCM ciphertext back to a credential map.
func DecryptCredentials(encryptedText string) (map[string]string, error) {
	InitCipherKey("", "")

	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(derivedKey[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var credentials map[string]string
	if err := json.Unmarshal(plaintext, &credentials); err != nil {
		return nil, err
	}

	return credentials, nil
}

// HasConfiguredCredentials checks if encryptedText contains non-empty password or connection_url.
func HasConfiguredCredentials(encryptedText string) bool {
	if encryptedText == "" {
		return false
	}
	creds, err := DecryptCredentials(encryptedText)
	if err != nil {
		return false
	}
	return (creds["password"] != "" || creds["connection_url"] != "")
}
