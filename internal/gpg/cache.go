package gpg

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"time"
)

type passphraseCache struct {
	Passphrase string    `json:"passphrase"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// getCacheKey returns a static key for AES (for demo: use a fixed key, but in production use a better secret)
func getCacheKey() []byte {
	// Use SHA-256 to ensure a 32-byte key
	sum := sha256.Sum256([]byte("gopass-demo-static-key"))
	return sum[:]
}

func EncryptAndCachePassphrase(passphrase, cachePath string, duration time.Duration) error {
	cache := passphraseCache{
		Passphrase: passphrase,
		ExpiresAt:  time.Now().Add(duration),
	}
	plaintext, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(getCacheKey())
	if err != nil {
		return err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return err
	}
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	out := base64.StdEncoding.EncodeToString(append(nonce, ciphertext...))
	return os.WriteFile(cachePath, []byte(out), 0600)
}

func DecryptCachedPassphrase(cachePath string) (string, bool, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return "", false, err
	}
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", false, err
	}
	block, err := aes.NewCipher(getCacheKey())
	if err != nil {
		return "", false, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", false, err
	}
	nonceSize := aead.NonceSize()
	if len(decoded) < nonceSize {
		return "", false, errors.New("invalid cache data")
	}
	nonce := decoded[:nonceSize]
	ciphertext := decoded[nonceSize:]
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", false, err
	}
	var cache passphraseCache
	if err := json.Unmarshal(plaintext, &cache); err != nil {
		return "", false, err
	}

	slog.Info("Decrypted cached passphrase", "expires_at", cache.ExpiresAt, "time now", time.Now())
	
	if time.Now().After(cache.ExpiresAt) {
		return "", false, nil
	}
	return cache.Passphrase, true, nil
}
