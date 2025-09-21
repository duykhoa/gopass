package gpg

import (
	"os"
	"testing"
	"time"
)

func TestEncryptAndCachePassphraseAndDecrypt(t *testing.T) {
	cachePath := "/tmp/gopass_test_passphrase.cache"
	defer os.Remove(cachePath)
	pass := "testpassphrase123"
	duration := 2 * time.Second

	err := EncryptAndCachePassphrase(pass, cachePath, duration)
	if err != nil {
		t.Fatalf("EncryptAndCachePassphrase failed: %v", err)
	}

	got, valid, err := DecryptCachedPassphrase(cachePath)
	if err != nil {
		t.Fatalf("DecryptCachedPassphrase failed: %v", err)
	}
	if !valid {
		t.Fatalf("Expected valid cache, got expired or invalid")
	}
	if got != pass {
		t.Errorf("Expected passphrase %q, got %q", pass, got)
	}

	// Wait for expiry
	time.Sleep(duration + 1*time.Second)
	_, valid, err = DecryptCachedPassphrase(cachePath)
	if err != nil {
		t.Fatalf("DecryptCachedPassphrase after expiry failed: %v", err)
	}
	if valid {
		t.Errorf("Expected cache to be expired, but got valid")
	}
}
