package esbuild

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanw/esbuild/pkg/api"
)

// HtmlPlugin is an exported variable for the HTML plugin
var HtmlPlugin = api.Plugin{
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
			err := os.WriteFile(filepath.Join("templates", "entry.html"), []byte(wrappedContent), 0644)
			if err != nil {
				return api.OnEndResult{}, err
			}
			fmt.Println("wrote entry.twig in ./templates")
			return api.OnEndResult{}, nil
		})
	},
}
