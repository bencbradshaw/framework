package templating

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

var templatesDir = "./templates"

func HtmlRender(templateName string, data map[string]any) (string, error) {
	fmt.Println("Rendering template: ", templateName)

	files := []string{
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "entry.html"),
		filepath.Join(templatesDir, templateName),
	}

	tmpl, err := template.ParseFiles(files...)
	if err != nil {
		return "", err
	}

	var output strings.Builder
	// Execute the template named "base"; this will pull in the "content"
	// definition from the specific file (e.g., about.html, index.html etc.)
	if err := tmpl.ExecuteTemplate(&output, "base", data); err != nil {
		return "", err
	}

	return output.String(), nil
}
