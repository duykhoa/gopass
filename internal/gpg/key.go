package gpg

import (
	"os/exec"
	"strings"
)

// HasPublicKey returns true if the given keyID is present in the user's public keyring.
func HasPublicKey(keyID string) bool {
	cmd := exec.Command("gpg", "--list-keys", keyID)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), keyID)
}
