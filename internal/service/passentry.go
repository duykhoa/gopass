package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AddOrEditPassEntry creates or updates a password entry with the given template and values.
// The entry is saved as a file under the password store directory, with template metadata appended.
func AddOrEditPassEntry(storeDir, entryName, templateName string, values map[string]string) error {
	tmpl := GetTemplateByName(templateName)
	if tmpl == nil {
		return fmt.Errorf("template not found: %s", templateName)
	}
	var sb strings.Builder
	for _, field := range tmpl.Fields {
		sb.WriteString(fmt.Sprintf("%s: %s\n", field, values[field]))
	}
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("template: %s\n", templateName))
	entryPath := filepath.Join(storeDir, entryName+".txt")
	return os.WriteFile(entryPath, []byte(sb.String()), 0600)
}
