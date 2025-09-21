package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
)

// SyncWithRemote pulls and pushes the password store directory with its remote.
func SyncWithRemote(storeDir string) error {
	repo, err := git.PlainOpen(storeDir)
	if err != nil {
		return fmt.Errorf("failed to open git repo: %w", err)
	}
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	// Pull from remote
	err = w.Pull(&git.PullOptions{RemoteName: "origin", Force: true})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("git pull failed: %w", err)
	}
	// Add all changes
	_ = w.AddWithOptions(&git.AddOptions{All: true})
	// Commit (if any changes)
	_, err = w.Commit("gopass sync", &git.CommitOptions{AllowEmptyCommits: true})
	if err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	// Push to remote
	err = repo.Push(&git.PushOptions{RemoteName: "origin"})
	if err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}
	return nil
}
