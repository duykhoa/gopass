package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func expandHome(path string) string {
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

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
	entries, err := listPasswordEntries(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestListPasswordEntries_WithEntries(t *testing.T) {
	dir := t.TempDir()
	files := []string{"entry1.gpg", "entry2.gpg", filepath.Join("subdir", "entry3.gpg")}
	for _, f := range files {
		path := filepath.Join(dir, f)
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte("dummy"), 0644)
	}

	entries, err := listPasswordEntries(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"entry1", "entry2", filepath.Join("subdir", "entry3")}
	if !equalUnordered(entries, want) {
		t.Errorf("got %v, want %v", entries, want)
	}
}

func equalUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	amap := make(map[string]int)
	bmap := make(map[string]int)
	for _, v := range a {
		amap[v]++
	}
	for _, v := range b {
		bmap[v]++
	}
	for k, v := range amap {
		if bmap[k] != v {
			return false
		}
	}
	return true
}
