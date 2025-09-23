package gpg

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ExportArmoredPublicKey exports the public key with the given keyID to ~/.gopass/<keyID>.public.asc in ASCII-armored format.
func ExportArmoredPublicKey(keyID, outputPath string) error {
	cmd := exec.Command("gpg", "--export", "--armor", keyID)
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to export public key: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return os.WriteFile(outputPath, out, 0600)
}
