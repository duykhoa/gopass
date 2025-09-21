package gpg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ExportArmoredPrivateKey exports the private key with the given keyID to the given output path in ASCII-armored format.
// If passphrase is not empty, it will be passed to gpg using --pinentry-mode loopback.
func ExportArmoredPrivateKey(keyID, outputPath, passphrase string) error {
	args := []string{"--export-secret-keys", "--armor", keyID}
	if passphrase != "" {
		args = append([]string{"--pinentry-mode", "loopback", "--passphrase", passphrase}, args...)
	}
	cmd := exec.Command("gpg", args...)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to export private key: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(outputPath, out, 0600)
}
