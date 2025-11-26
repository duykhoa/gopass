package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/duykhoa/gopass/internal/config"
	"github.com/duykhoa/gopass/internal/gpg"
)

type AddEditRequest struct {
	EntryName    string
	TemplateName string
	Fields       map[string]string
	GPGId        string
}

type AddEditResult struct {
	Err error
}

func AddOrEditEntry(req AddEditRequest) AddEditResult {
	if req.EntryName == "" {
		return AddEditResult{fmt.Errorf("entry name cannot be empty")}
	}

	var content strings.Builder
	if req.TemplateName == "Free Form" {
		content.WriteString(req.Fields["content"])
	} else {
		tmpl := GetTemplateByName(req.TemplateName)
		if tmpl == nil {
			return AddEditResult{fmt.Errorf("template not found: %s", req.TemplateName)}
		}
		for _, field := range tmpl.Fields {
			content.WriteString(fmt.Sprintf("%s: %s\n", field, req.Fields[field]))
		}
		if extra, ok := req.Fields["extra"]; ok && extra != "" {
			content.WriteString(fmt.Sprintf("extra: %s\n", extra))
		}
	}
	// Ensure separator is on its own line
	s := content.String()
	if !strings.HasSuffix(s, "\n") {
		content.WriteString("\n")
	}
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("template: %s\n", req.TemplateName))

	// Encrypt and store
	storeDir := config.PasswordStoreDir()
	entryPath := filepath.Join(storeDir, req.EntryName+".gpg")
	var gpgId string
	if gpgId == "" {
		gpgId = config.GPGId()
	}
	ciphertext, err := gpg.EncryptWithGPGKey([]byte(content.String()), gpgId)
	if err != nil {
		return AddEditResult{err}
	}
	if err := os.WriteFile(entryPath, ciphertext, 0600); err != nil {
		return AddEditResult{err}
	}
	return AddEditResult{nil}
}

// DeleteEntry removes the .gpg file for the given entry name from the password store.
func DeleteEntry(entryName string) error {
	storeDir := config.PasswordStoreDir()
	entryPath := filepath.Join(storeDir, entryName+".gpg")
	return os.Remove(entryPath)
}
