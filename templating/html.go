package templating

import (
	"html/template"
	"path/filepath"
	"strings"
)

// templatesDir is the directory where your template files are located
var templatesDir = "./templates"

// HtmlRender renders the specified template with the given data and returns the rendered HTML as a string.
func HtmlRender(templatePath string, data map[string]interface{}) (string, error) {
	// Parse base template along with the specific page template
	tmpl, err := template.ParseGlob(filepath.Join(templatesDir, "*.gohtml"))
	if err != nil {
		return "", err
	}

	var output strings.Builder
	// Execute the base template, passing the data
	if err := tmpl.ExecuteTemplate(&output, "base", data); err != nil {
		return "", err
	}

	return output.String(), nil
}
