package tests_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/bencbradshaw/framework/esbuild"
	"github.com/evanw/esbuild/pkg/api"
)

// Test mergeOptions thoroughly
func TestMergeOptions(t *testing.T) {
	defaultOpts := api.BuildOptions{
		EntryPoints:       []string{"default.js"},
		Outdir:            "dist_default",
		Bundle:            true,
		MinifySyntax:      true,
		LogLevel:          api.LogLevelInfo,
		Format:            api.FormatESModule,
	}

	passedOptsFull := api.BuildOptions{
		EntryPoints:       []string{"custom.js"},
		Outdir:            "dist_custom",
		Bundle:            false, // Override
		MinifySyntax:      false, // Override
		MinifyWhitespace:  true,  // New
		LogLevel:          api.LogLevelDebug, // Override
		Format:            api.FormatIIFE,    // Override
		Plugins:           []api.Plugin{{Name: "CustomPlugin"}}, // New
	}

	// Case 1: Pass full options, expect overrides and additions
	merged1 := esbuild.MergeOptionsForTesting(defaultOpts, passedOptsFull)

	if !reflect.DeepEqual(merged1.EntryPoints, passedOptsFull.EntryPoints) {
		t.Errorf("Case 1 EntryPoints: expected %v, got %v", passedOptsFull.EntryPoints, merged1.EntryPoints)
	}
	if merged1.Outdir != passedOptsFull.Outdir {
		t.Errorf("Case 1 Outdir: expected %s, got %s", passedOptsFull.Outdir, merged1.Outdir)
	}
	// With original merge logic: `if passed.Field { default.Field = true }`
	// If default.Bundle is true and passed.Bundle is false, merged.Bundle remains true.
	if merged1.Bundle != defaultOpts.Bundle { // Expect default true, as passedOptsFull.Bundle is false
		t.Errorf("Case 1 Bundle: expected %v (default), got %v. Passed value was %v.", defaultOpts.Bundle, merged1.Bundle, passedOptsFull.Bundle)
	}
	// If default.MinifySyntax is true and passed.MinifySyntax is false, merged.MinifySyntax remains true.
	if merged1.MinifySyntax != defaultOpts.MinifySyntax { // Expect default true, as passedOptsFull.MinifySyntax is false
		t.Errorf("Case 1 MinifySyntax: expected %v (default), got %v. Passed value was %v.", defaultOpts.MinifySyntax, merged1.MinifySyntax, passedOptsFull.MinifySyntax)
	}
	if !merged1.MinifyWhitespace { // From passedOptsFull (default is false, passed is true)
		t.Errorf("Case 1 MinifyWhitespace: expected true, got false")
	}
	if merged1.LogLevel != passedOptsFull.LogLevel {
		t.Errorf("Case 1 LogLevel: expected %v, got %v", passedOptsFull.LogLevel, merged1.LogLevel)
	}
	if merged1.Format != passedOptsFull.Format {
		t.Errorf("Case 1 Format: expected %v, got %v", passedOptsFull.Format, merged1.Format)
	}
	if len(merged1.Plugins) != 1 || merged1.Plugins[0].Name != "CustomPlugin" {
		t.Errorf("Case 1 Plugins: expected one plugin named 'CustomPlugin', got %v", merged1.Plugins)
	}


	// Case 2: Pass empty options, expect defaults
	// Create a fresh copy of defaultOpts for this test case, as MergeOptionsForTesting modifies its first argument.
	defaultOptsForCase2 := api.BuildOptions{
		EntryPoints:       []string{"default.js"},
		Outdir:            "dist_default",
		Bundle:            true,
		MinifySyntax:      true,
		LogLevel:          api.LogLevelInfo,
		Format:            api.FormatESModule,
	}
	passedOptsEmpty := api.BuildOptions{}
	merged2 := esbuild.MergeOptionsForTesting(defaultOptsForCase2, passedOptsEmpty)

	if !reflect.DeepEqual(merged2.EntryPoints, defaultOptsForCase2.EntryPoints) {
		t.Errorf("Case 2 EntryPoints: expected %v, got %v", defaultOptsForCase2.EntryPoints, merged2.EntryPoints)
	}
	if merged2.Outdir != defaultOptsForCase2.Outdir {
		t.Errorf("Case 2 Outdir: expected %s, got %s", defaultOptsForCase2.Outdir, merged2.Outdir)
	}
	if merged2.Bundle != defaultOptsForCase2.Bundle {
		t.Errorf("Case 2 Bundle: expected %v, got %v", defaultOptsForCase2.Bundle, merged2.Bundle)
	}
	// ... check other default fields ...

	// Case 3: Pass partial options
	defaultOptsForCase3 := api.BuildOptions{ // Fresh copy
		EntryPoints:       []string{"default.js"},
		Outdir:            "dist_default",
		Bundle:            true,
		MinifySyntax:      true,
		LogLevel:          api.LogLevelInfo,
		Format:            api.FormatESModule,
	}
	passedOptsPartial := api.BuildOptions{
		EntryPoints: []string{"partial.js"},
		LogLevel:    api.LogLevelWarning,
	}
	merged3 := esbuild.MergeOptionsForTesting(defaultOptsForCase3, passedOptsPartial)
	if !reflect.DeepEqual(merged3.EntryPoints, passedOptsPartial.EntryPoints) {
		t.Errorf("Case 3 EntryPoints: expected %v, got %v", passedOptsPartial.EntryPoints, merged3.EntryPoints)
	}
	if merged3.Outdir != defaultOptsForCase3.Outdir { // Should be default
		t.Errorf("Case 3 Outdir: expected %s, got %s", defaultOptsForCase3.Outdir, merged3.Outdir)
	}
	if merged3.LogLevel != passedOptsPartial.LogLevel { // Should be partial
		t.Errorf("Case 3 LogLevel: expected %v, got %v", passedOptsPartial.LogLevel, merged3.LogLevel)
	}
}

// Mock for api.BuildContext
type mockEsbuildContext struct {
	api.BuildContext // Embed to satisfy the interface
	watchCalled bool
	mu          sync.Mutex
}

func (m *mockEsbuildContext) Watch(options api.WatchOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.watchCalled = true
	return nil
}
// Ensure all methods of api.BuildContext are implemented
func (m *mockEsbuildContext) Serve(options api.ServeOptions) (api.ServeResult, error) { return api.ServeResult{}, nil }
func (m *mockEsbuildContext) Rebuild() api.BuildResult { return api.BuildResult{} } // Corrected based on previous framework_test.go
func (m *mockEsbuildContext) Dispose() {}


var (
	// For InitDevMode mock
	mockCtx                       *mockEsbuildContext
	mockESBuildContextFuncActual  func(options api.BuildOptions) (api.BuildContext, *api.ContextError) // Correct error type
	capturedInitDevModeOptions    api.BuildOptions
	initDevModeCaptureMutex       sync.Mutex

	// For Build mock
	mockBuildFuncActual           func(options api.BuildOptions) api.BuildResult
	buildFuncCalled               bool
	capturedBuildOptions          api.BuildOptions
	buildCaptureMutex             sync.Mutex
)


func TestInitDevMode(t *testing.T) {
	// Setup mock for ESBuildContextFunc
	originalESBuildContextFunc := esbuild.ESBuildContextFunc
	mockESBuildContextFuncActual = func(options api.BuildOptions) (api.BuildContext, *api.ContextError) { // Correct error type
		initDevModeCaptureMutex.Lock()
		capturedInitDevModeOptions = options
		initDevModeCaptureMutex.Unlock()
		mockCtx = &mockEsbuildContext{}
		return mockCtx, nil // Return nil for *api.ContextError
	}
	esbuild.ESBuildContextFunc = mockESBuildContextFuncActual
	defer func() { esbuild.ESBuildContextFunc = originalESBuildContextFunc }()

	// Reset capture variables
	initDevModeCaptureMutex.Lock()
	capturedInitDevModeOptions = api.BuildOptions{}
	initDevModeCaptureMutex.Unlock()
	mockCtx = nil


	userOptions := api.BuildOptions{
		EntryPoints: []string{"./frontend/custom/main.ts"},
		Outdir:      "./public/js",
	}

	ctx := esbuild.InitDevMode(userOptions)
	if ctx == nil {
		t.Fatal("InitDevMode returned nil context")
	}

	if mockCtx == nil { // Check if mockCtx was initialized by the mock function
		t.Fatal("mockCtx was not initialized by mockESBuildContextFuncActual")
	}
	mockCtx.mu.Lock()
	if !mockCtx.watchCalled {
		t.Errorf("BuildContext.Watch was not called")
	}
	mockCtx.mu.Unlock()

	initDevModeCaptureMutex.Lock()
	defer initDevModeCaptureMutex.Unlock() // Ensure mutex is unlocked

	if !reflect.DeepEqual(capturedInitDevModeOptions.EntryPoints, userOptions.EntryPoints) {
		t.Errorf("InitDevMode: EntryPoints not merged correctly. Expected %v, got %v", userOptions.EntryPoints, capturedInitDevModeOptions.EntryPoints)
	}
	if capturedInitDevModeOptions.Outdir != userOptions.Outdir {
		t.Errorf("InitDevMode: Outdir not merged correctly. Expected %s, got %s", userOptions.Outdir, capturedInitDevModeOptions.Outdir)
	}

	if capturedInitDevModeOptions.Bundle != true {
		t.Errorf("InitDevMode: Bundle expected true, got %v", capturedInitDevModeOptions.Bundle)
	}
	if capturedInitDevModeOptions.Sourcemap != api.SourceMapLinked {
		t.Errorf("InitDevMode: Sourcemap expected %v, got %v", api.SourceMapLinked, capturedInitDevModeOptions.Sourcemap)
	}
	if capturedInitDevModeOptions.MinifyWhitespace != false {
		t.Errorf("InitDevMode: MinifyWhitespace expected false, got %v", capturedInitDevModeOptions.MinifyWhitespace)
	}
	if capturedInitDevModeOptions.MinifyIdentifiers != false {
		t.Errorf("InitDevMode: MinifyIdentifiers expected false, got %v", capturedInitDevModeOptions.MinifyIdentifiers)
	}
	if capturedInitDevModeOptions.MinifySyntax != false {
		t.Errorf("InitDevMode: MinifySyntax expected false, got %v", capturedInitDevModeOptions.MinifySyntax)
	}

	pluginNames := []string{}
	for _, p := range capturedInitDevModeOptions.Plugins {
		pluginNames = append(pluginNames, p.Name)
	}
	expectedPlugins := []string{"html-plugin", "rebuild-plugin"} // Corrected case
	foundCount := 0
	for _, expectedName := range expectedPlugins {
		for _, actualName := range pluginNames {
			if actualName == expectedName {
				foundCount++
				break
			}
		}
	}
	if foundCount != len(expectedPlugins) {
		t.Errorf("InitDevMode: Did not find all expected plugins (%v). Got plugins: %v", expectedPlugins, pluginNames)
	}
}


func TestBuild(t *testing.T) {
	originalBuildFunc := esbuild.BuildFunc
	mockBuildFuncActual = func(options api.BuildOptions) api.BuildResult {
		buildCaptureMutex.Lock()
		capturedBuildOptions = options
		buildFuncCalled = true
		buildCaptureMutex.Unlock()
		return api.BuildResult{}
	}
	esbuild.BuildFunc = mockBuildFuncActual
	defer func() { esbuild.BuildFunc = originalBuildFunc }()

	buildCaptureMutex.Lock()
	capturedBuildOptions = api.BuildOptions{}
	buildFuncCalled = false
	buildCaptureMutex.Unlock()

	userOptions := api.BuildOptions{
		EntryPoints: []string{"./src/app.ts"},
		Outdir:      "./dist_prod",
		MinifySyntax: true,
		MinifyWhitespace: true,
		MinifyIdentifiers: true,
	}

	esbuild.Build(userOptions)

	buildCaptureMutex.Lock()
	defer buildCaptureMutex.Unlock() // Ensure mutex is unlocked

	if !buildFuncCalled {
		t.Fatal("esbuild.BuildFunc was not called by esbuild.Build")
	}

	if !reflect.DeepEqual(capturedBuildOptions.EntryPoints, userOptions.EntryPoints) {
		t.Errorf("Build: EntryPoints not merged correctly. Expected %v, got %v", userOptions.EntryPoints, capturedBuildOptions.EntryPoints)
	}
	if capturedBuildOptions.Outdir != userOptions.Outdir {
		t.Errorf("Build: Outdir not merged correctly. Expected %s, got %s", userOptions.Outdir, capturedBuildOptions.Outdir)
	}
	if capturedBuildOptions.MinifySyntax != userOptions.MinifySyntax {
		t.Errorf("Build: MinifySyntax not merged correctly. Expected %v, got %v", userOptions.MinifySyntax, capturedBuildOptions.MinifySyntax)
	}
    if capturedBuildOptions.MinifyWhitespace != userOptions.MinifyWhitespace {
        t.Errorf("Build: MinifyWhitespace not merged correctly. Expected %v, got %v", userOptions.MinifyWhitespace, capturedBuildOptions.MinifyWhitespace)
    }
    if capturedBuildOptions.MinifyIdentifiers != userOptions.MinifyIdentifiers {
        t.Errorf("Build: MinifyIdentifiers not merged correctly. Expected %v, got %v", userOptions.MinifyIdentifiers, capturedBuildOptions.MinifyIdentifiers)
    }

	if capturedBuildOptions.Bundle != true {
		t.Errorf("Build: Bundle expected true, got %v", capturedBuildOptions.Bundle)
	}
    if capturedBuildOptions.Sourcemap != api.SourceMapLinked {
        t.Errorf("Build: Sourcemap expected %v, got %v", api.SourceMapLinked, capturedBuildOptions.Sourcemap)
    }

	foundHtmlPlugin := false
	for _, p := range capturedBuildOptions.Plugins {
		if p.Name == "html-plugin" {  // Corrected case
			foundHtmlPlugin = true
			break
		}
	}
	if !foundHtmlPlugin {
		t.Errorf("Build: Did not find HtmlPlugin. Got plugins: %v", capturedBuildOptions.Plugins)
	}
}
