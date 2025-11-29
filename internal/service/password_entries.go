package service

import (
	"os"
	"path/filepath"
	"strings"
)

func ListPasswordEntries(dir string) ([]string, error) {
	var entries []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".gpg") {
			rel, _ := filepath.Rel(dir, path)
			entries = append(entries, strings.TrimSuffix(rel, ".gpg"))
		}
		return nil
	})
	return entries, err
}
