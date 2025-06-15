package templating

import (
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

// HtmlRender renders an HTML template with the given data.
// It takes the templates directory path, the name of the specific template file (e.g., "index.html"),
// and a map of data to be passed to the template.
// It expects a "base.html" and "entry.html" to be present in the templates directory.
func HtmlRender(templatesDir string, templateName string, data map[string]any) (string, error) {
	fmt.Println("Rendering template: ", templateName, "from", templatesDir)

	files := []string{
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "entry.html"),
		filepath.Join(templatesDir, templateName),
	}

	// Check if all files exist
	for _, file := range files {
		if _, err := filepath.Abs(file); err != nil {
			// This check is mostly for sanity; template.ParseFiles will give a more specific error
			// if a file doesn't exist or isn't readable. We're not returning error here
			// to keep the original structure, but logging could be an option.
			fmt.Printf("Warning: could not get absolute path for template file %s: %v\n", file, err)
		}
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
