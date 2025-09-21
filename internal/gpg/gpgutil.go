package gpg

import (
	"os/exec"
)

// CheckGPGAvailable returns true if the gpg command is available in the system.
func CheckGPGAvailable() bool {
	_, err := exec.LookPath("gpg")
	return err == nil
}
