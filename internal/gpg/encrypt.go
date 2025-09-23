package gpg

import (
	"fmt"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
)

// EncryptWithGPGKey encrypts the content with the given GPG key id using gopenpgp and returns the armored ciphertext.
func EncryptWithGPGKey(plaintext []byte, keyID string) ([]byte, error) {
	armored, err := LoadArmoredPublicKey(keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to load armored public key: %w", err)
	}
	keyObj, err := crypto.NewKeyFromArmored(armored)
	if err != nil {
		return nil, fmt.Errorf("failed to parse armored public key: %w", err)
	}
	keyRing, err := crypto.NewKeyRing(keyObj)
	if err != nil {
		return nil, fmt.Errorf("failed to create keyring: %w", err)
	}
	message := crypto.NewPlainMessage(plaintext)
	encrypted, err := keyRing.Encrypt(message, nil)
	if err != nil {
		return nil, fmt.Errorf("gopenpgp encryption failed: %w", err)
	}
	armored, err = encrypted.GetArmored()
	if err != nil {
		return nil, fmt.Errorf("failed to get armored encrypted message: %w", err)
	}
	return []byte(armored), nil
}
