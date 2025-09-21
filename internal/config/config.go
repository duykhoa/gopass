package config

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	passwordStoreDir string
	gpgId            string
	passphraseKey    []byte
	initOnce         sync.Once
)

func PasswordStoreDir() string {
	initOnce.Do(loadConfig)
	return passwordStoreDir
}

func GPGId() string {
	initOnce.Do(loadConfig)
	return gpgId
}

func PassphraseKey() []byte {
	initOnce.Do(loadConfig)
	return passphraseKey
}

func loadConfig() {
	home, _ := os.UserHomeDir()
	passwordStoreDir = filepath.Join(home, ".password-store")
	gpgIdPath := filepath.Join(passwordStoreDir, ".gpg-id")
	gpgIdBytes, err := os.ReadFile(gpgIdPath)
	if err == nil {
		gpgId = string(gpgIdBytes)
	}
	// For demo, use a static key. In production, use a secure method.
	passphraseKey = []byte("gopass-demo-static-key-32bytes!!")
}
