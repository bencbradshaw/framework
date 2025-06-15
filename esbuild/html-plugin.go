package esbuild

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// NewHtmlPlugin creates an esbuild plugin that generates an entry.html file
// in the specified templates directory.
func NewHtmlPlugin(templatesDir string) api.Plugin {
	return api.Plugin{
		Name: "html-plugin",
		Setup: func(build api.PluginBuild) {
			build.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
				var entryContent string
			for _, output := range result.OutputFiles {
				if strings.Contains(output.Path, "index") {
					scriptPath := filepath.Base(output.Path)
					if strings.HasSuffix(output.Path, ".js") {
						htmlScript := `<script type="module" src="/static/` + scriptPath + `"></script>`
						entryContent += htmlScript + "\n"
					}
					if strings.HasSuffix(output.Path, ".css") {
						htmlCss := `<link rel="stylesheet" href="/static/` + scriptPath + `">`
						entryContent += htmlCss + "\n"
					}
				}
			}
			wrappedContent := `{{define "entry"}}` + "\n" + entryContent + `{{end}}`
			// Ensure the templates directory exists
			if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
				if err := os.MkdirAll(templatesDir, 0755); err != nil {
					return api.OnEndResult{}, fmt.Errorf("failed to create templates directory %s: %w", templatesDir, err)
				}
			}
			err := os.WriteFile(filepath.Join(templatesDir, "entry.html"), []byte(wrappedContent), 0644)
			if err != nil {
				return api.OnEndResult{}, err
			}
			fmt.Printf("wrote entry.html in ./%s\n", templatesDir)
			return api.OnEndResult{}, nil
		})
	}, // Added comma here
}
} // Added missing function closing brace
