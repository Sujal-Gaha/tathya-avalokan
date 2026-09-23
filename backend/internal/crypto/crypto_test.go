package crypto

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func TestEncryptionRoundtrip(t *testing.T) {
	ResetCipherKeyForTest()
	secret := map[string]string{
		"password":       "ProductionSuperSecretPassword#2026",
		"connection_url": "postgresql://usr:pass@host/db",
	}

	cipherText, err := EncryptCredentials(secret)
	if err != nil {
		t.Fatalf("Failed to encrypt credentials: %v", err)
	}

	if cipherText == "" {
		t.Fatal("Ciphertext should not be empty")
	}

	decrypted, err := DecryptCredentials(cipherText)
	if err != nil {
		t.Fatalf("Failed to decrypt credentials: %v", err)
	}

	if decrypted["password"] != secret["password"] {
		t.Errorf("Expected password %q, got %q", secret["password"], decrypted["password"])
	}
	if decrypted["connection_url"] != secret["connection_url"] {
		t.Errorf("Expected connection_url %q, got %q", secret["connection_url"], decrypted["connection_url"])
	}
}

func TestMaskConnectionURI(t *testing.T) {
	rawURI := "postgresql://dbuser:MyClearPassword123@prod-cluster.internal:5432/analytics"
	masked := MaskConnectionURI(rawURI)

	if strings.Contains(masked, "MyClearPassword123") {
		t.Errorf("Password was not masked: %s", masked)
	}
	if !strings.Contains(masked, "dbuser:********@prod-cluster.internal") {
		t.Errorf("Expected masked format not found: %s", masked)
	}

	nonAuthURI := "sqlite:///path/to/db.sqlite"
	if MaskConnectionURI(nonAuthURI) != nonAuthURI {
		t.Errorf("Non-auth URI was modified: %s", MaskConnectionURI(nonAuthURI))
	}
}

func TestEncryptionKeyWarningLoggedWhenUnset(t *testing.T) {
	ResetCipherKeyForTest()
	_ = os.Unsetenv("TATHYA_ENCRYPTION_KEY")
	_ = os.Unsetenv("APP_SECRET_KEY")

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	_, err := EncryptCredentials(map[string]string{"password": "test"})
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	logOutput := buf.String()
	if !strings.Contains(logOutput, "TATHYA_ENCRYPTION_KEY is not set") {
		t.Errorf("Expected warning log not emitted: %s", logOutput)
	}
	if !IsFallbackKey() {
		t.Errorf("Expected fallback key to be active")
	}
}
