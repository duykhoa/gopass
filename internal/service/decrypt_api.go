package service

import (
	"github.com/duykhoa/gopass/internal/config"
	"github.com/duykhoa/gopass/internal/gpg"
)

// GetCachedPassphrase returns the cached passphrase and whether it is valid.
func GetCachedPassphrase() (string, bool) {
	pass, valid, _ := gpg.DecryptCachedPassphrase(getCachePath())
	return pass, valid
}

// DecryptAndCacheIfOk attempts decryption, and if successful, caches the passphrase.
func DecryptAndCacheIfOk(entry, passphrase string) DecryptResult {
	req := DecryptRequest{
		StoreDir:   config.PasswordStoreDir(),
		Entry:      entry,
		Passphrase: passphrase,
		Cache:      false,
		CachePath:  getCachePath(),
	}
	result := DecryptAndMaybeCache(req)
	if result.Err == nil && passphrase != "" {
		CachePassphrase(passphrase)
	}
	return result
}

// CachePassphrase stores the passphrase in the cache file.
func CachePassphrase(passphrase string) error {
	return gpg.EncryptAndCachePassphrase(passphrase, getCachePath(), 30*60) // 30 min
}

func getCachePath() string {
	return GetDefaultCachePath()
}
