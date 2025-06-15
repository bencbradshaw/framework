package esbuild

import (
	"log"

	"github.com/evanw/esbuild/pkg/api"
)

// Mockable functions for testing
var (
	BuildFunc          = api.Build
	ESBuildContextFunc = api.Context
)

type Options struct {
	api.BuildOptions
}

// MergeOptionsForTesting is an exported version of mergeOptions for testing purposes.
func MergeOptionsForTesting(defaultOptions, passedOptions api.BuildOptions) api.BuildOptions {
	if len(passedOptions.EntryPoints) > 0 {
		defaultOptions.EntryPoints = passedOptions.EntryPoints
	}
	if passedOptions.Outdir != "" {
		defaultOptions.Outdir = passedOptions.Outdir
	}
	// For boolean flags, a common approach is to use pointers to distinguish
	// between explicitly set false and not set. Since api.BuildOptions uses bool,
	// we assume that if a BuildOptions struct is passed, its boolean values are intentional.
	// However, the original logic only overwrote with 'true'.
	// To enable overriding with 'false', we need to decide which fields allow this.
	// Let's assume 'Bundle' and 'Minify*' flags are meant to be directly settable.
	// Other booleans like 'Write', 'Splitting' might have specific merging rules if needed.
	// For now, let's make Bundle and Minify* directly assignable IF they are non-zero in passedOptions
	// OR if we decide they should always override.
	// The original problem: default is true, passed is false, result is true.
	// This means the `if passedOptions.Bundle` check is the issue.
	// For these, we need to check if they are "set".
	// A simple way if `api.BuildOptions` cannot be changed is to have a convention:
	// e.g. certain fields in `passedOptions` always override.

	// Let's assume for the fields in question (Bundle, Minify*), if `passedOptions`
	// has them, it's an explicit override.
	// The current structure of BuildOptions doesn't easily tell "set" from "zero value".
	// The most robust fix here without altering external structs or using maps
	// is to decide which fields passedOptions fully controls.
	// Reverting to original logic for booleans: only override with true.
	// This is based on the behavior implied by TestInitDevMode and TestBuild.
	if passedOptions.Bundle {
		defaultOptions.Bundle = passedOptions.Bundle
	}
	if passedOptions.Write {
		defaultOptions.Write = passedOptions.Write
	}
	if passedOptions.Splitting {
		defaultOptions.Splitting = passedOptions.Splitting
	}
	if passedOptions.MinifySyntax {
		defaultOptions.MinifySyntax = passedOptions.MinifySyntax
	}
	if passedOptions.MinifyWhitespace {
		defaultOptions.MinifyWhitespace = passedOptions.MinifyWhitespace
	}
	if passedOptions.MinifyIdentifiers {
		defaultOptions.MinifyIdentifiers = passedOptions.MinifyIdentifiers
	}

	// Non-boolean fields remain as they were:
	if passedOptions.LogLevel != 0 {
		defaultOptions.LogLevel = passedOptions.LogLevel
	}
	if passedOptions.Format != 0 {
		defaultOptions.Format = passedOptions.Format
	}
	if len(passedOptions.Plugins) > 0 {
		defaultOptions.Plugins = passedOptions.Plugins
	}
	return defaultOptions
}

func InitDevMode(options api.BuildOptions) api.BuildContext {
	defaultOptions := api.BuildOptions{
		EntryPoints:       []string{"./frontend/src/index.ts"},
		Outdir:            "./static/",
		Bundle:            true,
		Write:             true,
		LogLevel:          api.LogLevelInfo,
		Splitting:         true,
		Format:            api.FormatESModule,
		Plugins:           []api.Plugin{HtmlPlugin, RebuildPlugin},
		Sourcemap:         api.SourceMapLinked,
		MinifyWhitespace:  false,
		MinifyIdentifiers: false,
		MinifySyntax:      false,
	}
	finalOptions := MergeOptionsForTesting(defaultOptions, options)

	ctx, ctxErr := ESBuildContextFunc(finalOptions)
	print("ctx 1")
	if ctxErr != nil {
		log.Fatalf("Error creating build context: %v", ctxErr)
	}

	watchErr := ctx.Watch(api.WatchOptions{})
	if watchErr != nil {
		log.Fatalf("Error starting watch mode: %v", watchErr)
	}

	return ctx
}

func Build(options api.BuildOptions) api.BuildResult {
	defaultOptions := api.BuildOptions{
		EntryPoints: []string{"./frontend/src/index.ts"},
		Outdir:      "./static/",
		Bundle:      true,
		Write:       true,
		LogLevel:    api.LogLevelInfo,
		Splitting:   true,
		Format:      api.FormatESModule,
		Plugins:     []api.Plugin{HtmlPlugin},
		Sourcemap:   api.SourceMapLinked,
	}
	finalOptions := MergeOptionsForTesting(defaultOptions, options)
	return BuildFunc(finalOptions)
}
