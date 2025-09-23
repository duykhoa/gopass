//go:build integration
// +build integration

package service

import (
	"os"
	"testing"
)

func TestIntegration_AddEditAndDecrypt(t *testing.T) {
	// This test requires a valid GPG key and config
	dir := t.TempDir()
	os.MkdirAll(dir, 0700)
	old := os.Getenv("PASSWORD_STORE_DIR")
	os.Setenv("PASSWORD_STORE_DIR", dir)
	defer os.Setenv("PASSWORD_STORE_DIR", old)

	req := AddEditRequest{
		EntryName:    "integration",
		TemplateName: "Free Form",
		Fields:       map[string]string{"content": "integration secret"},
		GPGId:        "testid", // Replace with a valid test key for real integration
	}
	result := AddOrEditEntry(req)
	if result.Err != nil {
		t.Fatalf("add entry failed: %v", result.Err)
	}
	// Decrypt would go here if test GPG key is set up
}
