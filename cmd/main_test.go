package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := expandHome("~/.password-store")
	want := filepath.Join(home, ".password-store")
	if got != want {
		t.Errorf("expandHome failed: got %q, want %q", got, want)
	}
}

func TestListPasswordEntries_Empty(t *testing.T) {
	dir := t.TempDir()
	old := passwordStoreDir
	defer func() { passwordStoreDir = old }()
	passwordStoreDir = dir
	entries, err := listPasswordEntries()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
