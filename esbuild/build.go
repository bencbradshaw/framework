package esbuild

import (
	"log"

	"github.com/evanw/esbuild/pkg/api"
)

type Options struct {
	api.BuildOptions
}

func mergeOptions(defaultOptions, passedOptions api.BuildOptions) api.BuildOptions {
	if len(passedOptions.EntryPoints) > 0 {
		defaultOptions.EntryPoints = passedOptions.EntryPoints
	}
	if passedOptions.Outdir != "" {
		defaultOptions.Outdir = passedOptions.Outdir
	}
	if passedOptions.Bundle {
		defaultOptions.Bundle = passedOptions.Bundle
	}
	if passedOptions.Write {
		defaultOptions.Write = passedOptions.Write
	}
	if passedOptions.LogLevel != 0 {
		defaultOptions.LogLevel = passedOptions.LogLevel
	}
	if passedOptions.Splitting {
		defaultOptions.Splitting = passedOptions.Splitting
	}
	if passedOptions.Format != 0 {
		defaultOptions.Format = passedOptions.Format
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
	finalOptions := mergeOptions(defaultOptions, options)

	ctx, ctxErr := api.Context(finalOptions)
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
	finalOptions := mergeOptions(defaultOptions, options)
	return api.Build(finalOptions)
}
