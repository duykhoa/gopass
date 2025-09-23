package service

const (
	TemplateFreeForm         = "Free Form"
	TemplateEmailAndPassword = "email and password"
)

type Template struct {
	Name   string
	Fields []string
}

var Templates = []Template{
	{
		Name:   TemplateFreeForm,
		Fields: []string{"content"},
	},
	{
		Name:   TemplateEmailAndPassword,
		Fields: []string{"domain", "email", "password", "extra"},
	},
}

func GetTemplateByName(name string) *Template {
	for _, t := range Templates {
		if t.Name == name {
			return &t
		}
	}
	return nil
}
