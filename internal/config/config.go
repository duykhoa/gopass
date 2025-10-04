package config

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	passwordStoreDir     string
	passwordStoreDirName string
	gpgId                string
	passphraseKey        []byte
	initOnce             sync.Once
)

func PasswordStoreDir() string {
	initOnce.Do(loadConfig)
	return passwordStoreDir
}

func PasswordStoreDirName() string {
	initOnce.Do(loadConfig)
	return passwordStoreDirName
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
	passwordStoreDirName = ".password-store"
	passwordStoreDir = filepath.Join(home, passwordStoreDirName)
	gpgIdPath := filepath.Join(passwordStoreDir, ".gpg-id")
	gpgIdBytes, err := os.ReadFile(gpgIdPath)
	if err == nil {
		gpgId = strings.TrimSpace(string(gpgIdBytes))
	}

	// This is a hardcoded passphrase encryption key, it is 
	// probably a good idea to generate a random passphrase 
	// in the first time running the app
	passphraseKey = []byte("gopass-demo-static-key-32bytes!!")
}
