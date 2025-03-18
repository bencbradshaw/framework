package esbuild

import (
	"fmt"

	"github.com/bencbradshaw/framework/events"

	"github.com/evanw/esbuild/pkg/api"
)

// HtmlPlugin is an exported variable for the HTML plugin
var RebuildPlugin = api.Plugin{
	Name: "rebuild-plugin",
	Setup: func(build api.PluginBuild) {
		build.OnEnd(func(result *api.BuildResult) (api.OnEndResult, error) {
			events.EmitEvent("esbuild", map[string]bool{"done": true})
			fmt.Println("emitted esbuild event")
			return api.OnEndResult{}, nil
		})
	},
}
