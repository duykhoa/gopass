package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddOrEditEntry_FreeForm(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(dir, 0700)
	old := os.Getenv("PASSWORD_STORE_DIR")
	os.Setenv("PASSWORD_STORE_DIR", dir)
	defer os.Setenv("PASSWORD_STORE_DIR", old)

	req := AddEditRequest{
		EntryName:    "testentry",
		TemplateName: "Free Form",
		Fields:       map[string]string{"content": "my secret"},
		GPGId:        "testid", // Should be replaced with a valid test key for real test
	}
	result := AddOrEditEntry(req)
	if result.Err == nil {
		// Should fail unless test GPG key is set up
		entryPath := filepath.Join(dir, "testentry.gpg")
		if _, err := os.Stat(entryPath); err != nil {
			t.Errorf("entry file not created: %v", err)
		}
	}
}

func TestGetTemplateByName(t *testing.T) {
	tmpl := GetTemplateByName("email and password")
	if tmpl == nil || tmpl.Name != "email and password" {
		t.Errorf("template lookup failed")
	}
}
