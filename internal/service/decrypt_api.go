package service

import (
	"log/slog"
	"time"

	"github.com/duykhoa/gopass/internal/config"
	"github.com/duykhoa/gopass/internal/gpg"
)

const cacheDuration = 30 * time.Minute

// GetCachedPassphrase returns the cached passphrase and whether it is valid.
func GetCachedPassphrase() (string, bool) {
	pass, valid, err := gpg.DecryptCachedPassphrase(getCachePath())
	slog.Info("Failed to decrypt cached passphrase", "error", err, "valid", valid)
	return pass, valid
}

// DecryptAndCacheIfOk attempts decryption, and if successful, caches the passphrase.
func DecryptAndCacheIfOk(entry, passphrase string) DecryptResult {
	req := DecryptRequest{
		StoreDir:   config.PasswordStoreDir(),
		Entry:      entry,
		Passphrase: passphrase,
		CachePath:  getCachePath(),
	}
	result := Decrypt(req)
	if result.Err == nil && passphrase != "" {
		CachePassphrase(passphrase)
	}
	return result
}

// CachePassphrase stores the passphrase in the cache file.
func CachePassphrase(passphrase string) error {
	return gpg.EncryptAndCachePassphrase(passphrase, getCachePath(), cacheDuration)
}

func getCachePath() string {
	return GetDefaultCachePath()
}
