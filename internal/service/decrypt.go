package service

import (
	"os"
	"path/filepath"

	"github.com/duykhoa/gopass/internal/gpg"
)

type DecryptRequest struct {
	StoreDir   string
	Entry      string
	Passphrase string
	CachePath  string
}

type DecryptResult struct {
	Plaintext string
	Err       error
}

// Decrypt handles decryption.
func Decrypt(req DecryptRequest) DecryptResult {
	gpgFile := filepath.Join(req.StoreDir, req.Entry+".gpg")
	plaintext, err := gpg.DecryptGPGFileWithKey(gpgFile, req.Passphrase)

	return DecryptResult{plaintext, err}
}

// GetDefaultCachePath returns the default cache path for the passphrase.
func GetDefaultCachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gopass", "passphrase.cache")
}
