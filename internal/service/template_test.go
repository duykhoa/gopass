package service

import "testing"

func TestTemplatesDefined(t *testing.T) {
	if len(Templates) == 0 {
		t.Fatal("no templates defined")
	}
	for _, tmpl := range Templates {
		if tmpl.Name == "" {
			t.Errorf("template missing name")
		}
		if len(tmpl.Fields) == 0 {
			t.Errorf("template %s missing fields", tmpl.Name)
		}
	}
}
