package tests_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bencbradshaw/framework"
	"github.com/bencbradshaw/framework/esbuild" // For mocking
	"github.com/evanw/esbuild/pkg/api"
)

// Helper to set up and tear down templates for framework tests
// Similar to the one in templating_test.go
func setupFrameworkTestTemplates(t *testing.T, testTemplatesSourceDir string) func() {
	originalCwd, _ := os.Getwd()
	// Tests run from 'tests' dir. Framework.go expects 'templates' at project root.
	rootDir := filepath.Join(originalCwd, "..")
	projectTemplatesDir := filepath.Join(rootDir, "templates")

	backupDir := ""
	if _, err := os.Stat(projectTemplatesDir); !os.IsNotExist(err) {
		backupDir = projectTemplatesDir + "_backup_fw_test_" + strings.ReplaceAll(t.Name(), string(filepath.Separator), "_")
		if err := os.Rename(projectTemplatesDir, backupDir); err != nil {
			t.Fatalf("Failed to backup existing project templates directory: %v", err)
		}
	}

	// Copy test templates to where Framework.go expects them (project_root/templates)
	// AND to where templating.HtmlRender expects them (project_root/templating/templates)
	sourceDirAbs, err := filepath.Abs(testTemplatesSourceDir) // Should be /app/tests/fw_templates
	if err != nil {
		t.Fatalf("Failed to get absolute path for sourceDir %s: %v", testTemplatesSourceDir, err)
	}

	// Target 1: project_root/templates (already defined as projectTemplatesDir -> /app/templates)
	// projectRootTemplatesDir := projectTemplatesDir // This line is removed. Use projectTemplatesDir.

	// Target 2: CWD_of_test/templates. CWD is originalCwd = /app/tests
	testPackageLocalTemplatesDir := filepath.Join(originalCwd, "templates") // -> /app/tests/templates

	// Backup for Target 2
	backupDirTestLocal := ""
	if _, err := os.Stat(testPackageLocalTemplatesDir); !os.IsNotExist(err) {
		backupDirTestLocal = testPackageLocalTemplatesDir + "_backup_fw_test_local_" + strings.ReplaceAll(t.Name(), string(filepath.Separator), "_")
		if err := os.Rename(testPackageLocalTemplatesDir, backupDirTestLocal); err != nil {
			t.Logf("Warning: Failed to backup existing test-local templates directory: %v.", err)
			backupDirTestLocal = ""
		}
	}

	// Create target directories
	if err := os.MkdirAll(projectTemplatesDir, 0755); err != nil { // Use projectTemplatesDir
		if backupDir != "" { os.Rename(backupDir, projectTemplatesDir) } // Restore main backup
		t.Fatalf("Failed to create project root templates dir for test: %v", err)
	}
	if err := os.MkdirAll(testPackageLocalTemplatesDir, 0755); err != nil {
		if backupDirTestLocal != "" { os.Rename(backupDirTestLocal, testPackageLocalTemplatesDir) }
		os.RemoveAll(projectTemplatesDir) // Clean up other created dir // Use projectTemplatesDir
		if backupDir != "" { os.Rename(backupDir, projectTemplatesDir) }
		t.Fatalf("Failed to create test-local templates dir for test: %v", err)
	}

	files, err := os.ReadDir(sourceDirAbs)
	if err != nil {
		// Cleanup both
		os.RemoveAll(projectTemplatesDir) // Use projectTemplatesDir
		if backupDir != "" { os.Rename(backupDir, projectTemplatesDir) }
		os.RemoveAll(testPackageLocalTemplatesDir)
		if backupDirTestLocal != "" { os.Rename(backupDirTestLocal, testPackageLocalTemplatesDir) }
		t.Fatalf("Failed to read test templates source directory %s: %v", sourceDirAbs, err)
	}

	for _, file := range files {
		sourceFile := filepath.Join(sourceDirAbs, file.Name())
		data, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("Failed to read source template file %s: %v", sourceFile, err)
		}

		// Write to project_root/templates
		destFileProjectRoot := filepath.Join(projectTemplatesDir, file.Name()) // Use projectTemplatesDir
		if err := os.WriteFile(destFileProjectRoot, data, 0644); err != nil {
			t.Fatalf("Failed to write to project_root/templates file %s: %v", destFileProjectRoot, err)
		}
		// Write to CWD_of_test/templates
		destFileTestLocal := filepath.Join(testPackageLocalTemplatesDir, file.Name())
		if err := os.WriteFile(destFileTestLocal, data, 0644); err != nil {
			t.Fatalf("Failed to write to test-local/templates file %s: %v", destFileTestLocal, err)
		}
	}

	return func() {
		// Teardown for project_root/templates
		if err := os.RemoveAll(projectTemplatesDir); err != nil { // Use projectTemplatesDir
			t.Logf("Warning: failed to remove project root templates directory %s: %v", projectTemplatesDir, err)
		}
		if backupDir != "" {
			if err := os.Rename(backupDir, projectTemplatesDir); err != nil {
				t.Logf("Warning: failed to restore original project root templates directory: %v", err)
			}
		}

		// Teardown for CWD_of_test/templates
		if err := os.RemoveAll(testPackageLocalTemplatesDir); err != nil {
			t.Logf("Warning: failed to remove test-local templates directory %s: %v", testPackageLocalTemplatesDir, err)
		}
		if backupDirTestLocal != "" {
			if err := os.Rename(backupDirTestLocal, testPackageLocalTemplatesDir); err != nil {
				t.Logf("Warning: failed to restore original test-local templates directory: %v", err)
			}
		}
	}
}


// Mock esbuild.BuildFunc
var (
	mockBuildFuncCalled           bool
	mockBuildFuncOptions          api.BuildOptions
	mockESBuildContextFuncCalled  bool
	mockESBuildContextFuncOptions api.BuildOptions
	mu                            sync.Mutex
)

func mockBuild(options api.BuildOptions) api.BuildResult {
	mu.Lock()
	defer mu.Unlock()
	mockBuildFuncCalled = true
	mockBuildFuncOptions = options
	// Return a successful build result, can be empty if not checked
	return api.BuildResult{Errors: []api.Message{}, Warnings: []api.Message{}}
}

// Mock for api.Context used by esbuild.InitDevMode
type mockBuildContext struct {
	api.BuildContext // Embed to satisfy the interface
	watchCalled bool
	disposeCalled bool
}
func (m *mockBuildContext) Watch(options api.WatchOptions) error {
	m.watchCalled = true
	return nil
}
func (m *mockBuildContext) Serve(options api.ServeOptions) (api.ServeResult, error) { return api.ServeResult{}, nil }
func (m *mockBuildContext) Rebuild() api.BuildResult { return api.BuildResult{} }
func (m *mockBuildContext) Dispose() {
	m.disposeCalled = true
}


func mockESBuildContext(options api.BuildOptions) (api.BuildContext, *api.ContextError) {
	mu.Lock()
	defer mu.Unlock()
	mockESBuildContextFuncCalled = true
	mockESBuildContextFuncOptions = options
	return &mockBuildContext{}, nil
}


func TestFramework_Run_DevMode(t *testing.T) {
	originalBuildFunc := esbuild.BuildFunc
	originalContextFunc := esbuild.ESBuildContextFunc
	esbuild.BuildFunc = mockBuild
	esbuild.ESBuildContextFunc = mockESBuildContext
	defer func() {
		esbuild.BuildFunc = originalBuildFunc
		esbuild.ESBuildContextFunc = originalContextFunc
	}()

	mu.Lock()
	mockESBuildContextFuncCalled = false // Reset flag
	mu.Unlock()

	params := &framework.InitParams{
		IsDevMode:                  true,
		AutoRegisterTemplateRoutes: false,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	var mux http.Handler
	// var server *httptest.Server // Removed as it's not used

	go func() {
		defer wg.Done()
		// framework.Run is blocking in dev mode.
		// It also starts an HTTP server internally if not provided one.
		// We don't have a direct way to shut down that server from here to stop Run.
		// The mockBuildContext's Watch could be made to exit on a signal for a cleaner shutdown.
		// For now, this test relies on checking the mock call and then letting the test finish.
		// The main goroutine won't wait for this goroutine past a timeout if it hangs.
		mux = framework.Run(params)
		if mux != nil {
			// If Run returns (e.g. if esbuild watch somehow exits), we can try to start a server
			// to ensure it's a valid mux, but this part might not be reached if it blocks.
			// server = httptest.NewServer(mux)
		}
	}()

	// Give Run some time to start up and call InitDevMode
	// Increased sleep to reduce flakiness if InitDevMode is called slightly later.
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	if !mockESBuildContextFuncCalled {
		t.Errorf("esbuild.InitDevMode (via ESBuildContextFunc) was not called in dev mode")
	}
	mu.Unlock()

	// Attempt to signal shutdown to esbuild.Watch if possible (e.g., by closing a channel it select{}s on).
	// This is not currently possible with the mockBuildContext as is.
	// Test will pass if mock called, but `Run` goroutine might linger until test process ends.
	// If `server` was started, `server.Close()` would be here.
	// `wg.Wait()` would hang.
}

func TestFramework_Run_ProdMode(t *testing.T) {
	originalBuildFunc := esbuild.BuildFunc
	originalContextFunc := esbuild.ESBuildContextFunc
	esbuild.BuildFunc = mockBuild
	esbuild.ESBuildContextFunc = mockESBuildContext
	defer func() {
		esbuild.BuildFunc = originalBuildFunc
		esbuild.ESBuildContextFunc = originalContextFunc
	}()

	mu.Lock()
	mockBuildFuncCalled = false
	mockESBuildContextFuncCalled = false
	mu.Unlock()

	params := &framework.InitParams{
		IsDevMode:                  false,
		AutoRegisterTemplateRoutes: false,
	}
	// framework.Run in production mode sets up routes and might start a server if not careful.
	// It should be non-blocking unless it explicitly starts a server and blocks on ListenAndServe.
	// The current Framework.go Run does not start a server itself, it returns a mux.
	mux := framework.Run(params)
	if mux == nil {
		t.Fatal("framework.Run returned nil mux")
	}

	mu.Lock()
	if mockESBuildContextFuncCalled {
		t.Errorf("esbuild.InitDevMode (via ESBuildContextFunc) was called in prod mode")
	}
	if mockBuildFuncCalled { // framework.Run itself should not call esbuild.Build
		t.Errorf("esbuild.Build (via BuildFunc) was called by Run in prod mode (should not happen)")
	}
	mu.Unlock()
}

func TestFramework_Run_AutoRegisterTemplateRoutes(t *testing.T) {
	teardownTemplates := setupFrameworkTestTemplates(t, "./fw_templates")
	defer teardownTemplates()

	params := &framework.InitParams{
		IsDevMode:                  false,
		AutoRegisterTemplateRoutes: true,
	}
	mux := framework.Run(params)
	if mux == nil {
		t.Fatal("framework.Run returned nil mux")
	}

	server := httptest.NewServer(mux)
	defer server.Close()

	tests := []struct {
		route              string
		expectedStatusCode int
		expectedContent    []string
	}{
		{"/", http.StatusOK, []string{"<h1>Index Page</h1>", "Base Header", "Entry Point"}},
		{"/index", http.StatusOK, []string{"<h1>Index Page</h1>", "Base Header"}},
		{"/about", http.StatusOK, []string{"<h1>About Page</h1>", "Framework v", "Base Header"}},
		{"/users/", http.StatusOK, []string{"<h1>Users Subroute</h1>", "Base Header"}}, // Corrected path
		{"/nonexistent", http.StatusOK, []string{"<h1>Index Page</h1>"}},
	}

	for _, tt := range tests {
		t.Run(tt.route, func(t *testing.T) {
			res, err := http.Get(server.URL + tt.route)
			if err != nil {
				t.Fatalf("HTTP GET request to %s failed: %v", tt.route, err)
			}
			defer res.Body.Close()

			bodyBytes, _ := io.ReadAll(res.Body)
			bodyString := string(bodyBytes)

			if res.StatusCode != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tt.expectedStatusCode, res.StatusCode, bodyString)
			}

			for _, content := range tt.expectedContent {
				if !strings.Contains(bodyString, content) {
					t.Errorf("Expected response for %s to contain '%s', but it didn't. Body:\n%s", tt.route, content, bodyString)
				}
			}
		})
	}
}


func TestFramework_Run_StaticAndEventsRoutes(t *testing.T) {
	params := &framework.InitParams{AutoRegisterTemplateRoutes: false}
	mux := framework.Run(params)
	server := httptest.NewServer(mux)
	defer server.Close()

	respEvents, errEvents := http.Get(server.URL + "/events")
	if errEvents != nil {
		t.Fatalf("Failed to connect to /events: %v", errEvents)
	}
	defer respEvents.Body.Close()
	if respEvents.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK for /events, got %d", respEvents.StatusCode)
	}
	if respEvents.Header.Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected Content-Type text/event-stream for /events, got %s", respEvents.Header.Get("Content-Type"))
	}

	projectRootDir, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("Could not get project root dir: %v", err)
	}
	staticDir := filepath.Join(projectRootDir, "static")
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		t.Fatalf("Could not create static dir: %v", err)
	}
	defer os.RemoveAll(staticDir)

	staticFilePath := filepath.Join(staticDir, "test.txt")
	if err := os.WriteFile(staticFilePath, []byte("static content"), 0644); err != nil {
		t.Fatalf("Could not write static file: %v", err)
	}

	respStatic, errStatic := http.Get(server.URL + "/static/test.txt")
	if errStatic != nil {
		t.Fatalf("Failed to get /static/test.txt: %v", errStatic)
	}
	defer respStatic.Body.Close()

	bodyBytes, _ := io.ReadAll(respStatic.Body) // Read body for checking and error reporting
	bodyString := string(bodyBytes)

	if respStatic.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK for /static/test.txt, got %d. Body: %s", respStatic.StatusCode, bodyString)
	}
	if bodyString != "static content" {
		t.Errorf("Expected 'static content' from /static/test.txt, got '%s'", bodyString)
	}
}


func TestFramework_RenderWithHtmlResponse_Success(t *testing.T) {
	teardownTemplates := setupFrameworkTestTemplates(t, "./fw_templates")
	defer teardownTemplates()

	rr := httptest.NewRecorder()
	data := map[string]interface{}{"title": "RenderTest", "message": "Hello Render"}

	framework.RenderWithHtmlResponse(rr, "index.html", data)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Expected Content-Type 'text/html; charset=utf-8', got '%s'", contentType)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "<h1>Index Page</h1>") {
		t.Errorf("Rendered HTML does not contain '<h1>Index Page</h1>'. Got:\n%s", body)
	}
	if !strings.Contains(body, "Hello Render") {
		t.Errorf("Rendered HTML does not contain 'Hello Render'. Got:\n%s", body)
	}
}

func TestFramework_RenderWithHtmlResponse_TemplateError(t *testing.T) {
	teardownTemplates := setupFrameworkTestTemplates(t, "./fw_templates")
	defer teardownTemplates()

	rr := httptest.NewRecorder()
	data := map[string]interface{}{"title": "ErrorTest"}

	framework.RenderWithHtmlResponse(rr, "nonexistent.html", data)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "Error rendering template") {
		t.Errorf("Expected error message in body. Got:\n%s", body)
	}
}


func TestFramework_Build(t *testing.T) {
	originalBuildFunc := esbuild.BuildFunc
	esbuild.BuildFunc = mockBuild
	defer func() { esbuild.BuildFunc = originalBuildFunc }()

	mu.Lock()
	mockBuildFuncCalled = false
	mockBuildFuncOptions = api.BuildOptions{}
	mu.Unlock()

	testEntry := []string{"./test/entry.js"}
	params := framework.InitParams{
		EsbuildOpts: api.BuildOptions{
			EntryPoints: testEntry,
			Outdir:      "out",
		},
	}
	framework.Build(params)

	mu.Lock()
	if !mockBuildFuncCalled {
		t.Errorf("esbuild.Build (via BuildFunc) was not called")
	}
	if len(mockBuildFuncOptions.EntryPoints) == 0 || mockBuildFuncOptions.EntryPoints[0] != testEntry[0] {
		t.Errorf("esbuild.Build called with incorrect EntryPoints: expected %v, got %v", testEntry, mockBuildFuncOptions.EntryPoints)
	}
	if mockBuildFuncOptions.Outdir != "out" {
		t.Errorf("esbuild.Build called with incorrect Outdir: expected 'out', got '%s'", mockBuildFuncOptions.Outdir)
	}
	mu.Unlock()
}
