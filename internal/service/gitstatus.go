package service

import (
	"github.com/duykhoa/gopass/internal/config"
	"github.com/go-git/go-git/v5"
)

// GetGitStatus returns a string describing if there are uncommitted changes or if local branch is ahead.
func GetGitStatus() string {
	repo, err := git.PlainOpen(config.PasswordStoreDir())
	if err != nil {
		return ""
	}
	w, err := repo.Worktree()
	if err != nil {
		return ""
	}
	status, err := w.Status()
	if err != nil {
		return ""
	}
	if !status.IsClean() {
		return "New entry or changes detected, please sync."
	}
	// Check if branch is ahead
	head, err := repo.Head()
	if err != nil {
		return ""
	}
	branch, err := repo.Branch(head.Name().Short())
	if err != nil {
		return ""
	}
	if branch.Merge != "" {
		return "Local branch is ahead, please sync."
	}
	return ""
}
