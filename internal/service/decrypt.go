package service

import (
	"os"
	"path/filepath"
	"time"

	"github.com/duykhoa/gopass/internal/gpg"
)

type DecryptRequest struct {
	StoreDir   string
	Entry      string
	Passphrase string
	Cache      bool
	CachePath  string
}

type DecryptResult struct {
	Plaintext string
	Err       error
}

// DecryptAndMaybeCache handles passphrase caching and decryption.
func DecryptAndMaybeCache(req DecryptRequest) DecryptResult {
	if req.Cache && req.Passphrase != "" && req.CachePath != "" {
		err := gpg.EncryptAndCachePassphrase(req.Passphrase, req.CachePath, 30*time.Minute)
		if err != nil {
			return DecryptResult{"", err}
		}
	}
	gpgFile := filepath.Join(req.StoreDir, req.Entry+".gpg")
	plaintext, err := gpg.DecryptGPGFileWithKey(gpgFile, req.Passphrase)
	return DecryptResult{plaintext, err}
}

// GetDefaultCachePath returns the default cache path for the passphrase.
func GetDefaultCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gopass", "passphrase.cache")
}
