package store

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	gitcfg "github.com/go-git/go-git/v5/config"
)

// InitPasswordStore creates a new password store compatible with pass
func InitPasswordStore(baseDir, keyID, remoteURL string) error {
	if baseDir == "" {
		return fmt.Errorf("baseDir cannot be empty")
	}

	if keyID == "" {
		return fmt.Errorf("keyID cannot be empty")
	}

	if remoteURL == "" {
		// Right now we don't allow empty remoteURL, but we might in the future
		return fmt.Errorf("remoteURL cannot be empty")
	}

	if _, err := os.Stat(baseDir); err == nil {
		return fmt.Errorf("password store already exists at %s", baseDir)
	}
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return fmt.Errorf("failed to create store directory: %w", err)
	}
	gpgIdPath := filepath.Join(baseDir, ".gpg-id")
	if err := os.WriteFile(gpgIdPath, []byte(keyID+"\n"), 0600); err != nil {
		return fmt.Errorf("failed to write .gpg-id: %w", err)
	}
	gitRepo, err := git.PlainInit(baseDir, false)
	if err != nil {
		return fmt.Errorf("failed to init git repo: %w", err)
	}
	if remoteURL != "" {
		_, err = gitRepo.CreateRemote(&gitcfg.RemoteConfig{
			Name: "origin",
			URLs: []string{remoteURL},
		})
		if err != nil {
			return fmt.Errorf("failed to set remote: %w", err)
		}
	}
	gitattributesPath := filepath.Join(baseDir, ".gitattributes")
	if err := os.WriteFile(gitattributesPath, []byte("*.gpg diff=gpg\n"), 0600); err != nil {
		return fmt.Errorf("failed to write .gitattributes: %w", err)
	}
	return nil
}
