package template

import (
	"embed"
	"html/template"
	"strings"
)

//go:embed *.gohtml
var fs embed.FS

var Templates = template.Must(template.ParseFS(fs, "*"))

func Render(name string, data any) (string, error) {
	var buf strings.Builder
	err := Templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
