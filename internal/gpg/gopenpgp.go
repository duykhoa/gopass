package gpg

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// DecryptGPGFileWithKey decrypts a GPG file using the first private key found in the user's GPG keyring.
func DecryptGPGFileWithKey(gpgFile, passphrase string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	// Find .gpg-id in the password store directory
	storeDir := filepath.Dir(gpgFile)
	gpgIdPath := filepath.Join(storeDir, ".gpg-id")
	gpgIdData, err := os.ReadFile(gpgIdPath)
	if err != nil {
		return "", fmt.Errorf("failed to read .gpg-id: %w", err)
	}
	keyID := strings.TrimSpace(string(gpgIdData))
	gopassDir := filepath.Join(usr.HomeDir, ".gopass")
	armoredKeyPath := filepath.Join(gopassDir, keyID+".asc")
	// If armored key does not exist, export it
	if _, err := os.Stat(armoredKeyPath); os.IsNotExist(err) {
		if err := ExportArmoredPrivateKey(keyID, armoredKeyPath, passphrase); err != nil {
			return "", fmt.Errorf("failed to export armored private key: %w", err)
		}
	}
	keyData, err := os.ReadFile(armoredKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read armored private key: %w", err)
	}
	keyObj, err := crypto.NewKeyFromArmored(string(keyData))
	if err != nil {
		return "", fmt.Errorf("failed to parse armored private key: %w", err)
	}
	unlockedKey, err := keyObj.Unlock([]byte(passphrase))
	if err != nil {
		return "", fmt.Errorf("failed to unlock key: %w", err)
	}
	defer keyObj.ClearPrivateParams()

	ciphertext, err := os.ReadFile(gpgFile)
	if err != nil {
		return "", fmt.Errorf("failed to read gpg file: %w", err)
	}
	message := crypto.NewPGPMessage(ciphertext)
	keyring, err := crypto.NewKeyRing(unlockedKey)
	if err != nil {
		return "", fmt.Errorf("failed to create keyring: %w", err)
	}
	plain, err := keyring.Decrypt(message, nil, 0)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}
	return string(plain.GetBinary()), nil
}
