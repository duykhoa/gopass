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

// DecryptAndMaybeCacheWithCache is a simplified API for main: handles cache and decryption for an entry.
func DecryptAndMaybeCacheWithCache(entry, passphrase string, cache bool) DecryptResult {
	req := DecryptRequest{
		StoreDir:   config.PasswordStoreDir(),
		Entry:      entry,
		Passphrase: passphrase,
		Cache:      cache,
		CachePath:  getCachePath(),
	}
	return DecryptAndMaybeCache(req)
}

func getCachePath() string {
	return GetDefaultCachePath()
}
