package service

import (
	"testing"
)

func TestGetTemplateByName(t *testing.T) {
	tmpl := GetTemplateByName("Email and password")
	if tmpl == nil || tmpl.Name != "Email and password" {
		t.Errorf("template lookup failed")
	}
}
