package gpg

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// LoadArmoredPublicKey loads the recipient's armored public key from ~/.gopass/<keyID>.public.asc
func LoadArmoredPublicKey(keyID string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user: %w", err)
	}
	gopassDir := filepath.Join(usr.HomeDir, ".gopass")
	armoredKeyPath := filepath.Join(gopassDir, keyID+".public.asc")
	keyData, err := os.ReadFile(armoredKeyPath)
	if err != nil {
		// Try to export the public key if not found
		if exportErr := ExportArmoredPublicKey(keyID, armoredKeyPath); exportErr != nil {
			return "", fmt.Errorf("failed to read armored public key and export failed: %w", exportErr)
		}
		keyData, err = os.ReadFile(armoredKeyPath)
		if err != nil {
			return "", fmt.Errorf("failed to read armored public key after export: %w", err)
		}
	}
	return string(keyData), nil
}
