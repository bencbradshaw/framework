package framework_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bencbradshaw/framework"
	"github.com/evanw/esbuild/pkg/api"
)

func TestConfigurableTemplatesDir(t *testing.T) {
	customTemplatesPath := "custom_test_templates"
	// Clean up any previous test runs
	_ = os.RemoveAll(customTemplatesPath)

	err := os.MkdirAll(customTemplatesPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create custom templates directory: %v", err)
	}
	defer func() {
		err := os.RemoveAll(customTemplatesPath)
		if err != nil {
			t.Logf("Failed to remove custom templates directory: %v", err)
		}
	}()

	baseHTMLContent := `{{define "base"}} {{template "content" .}} {{end}}`
	indexHTMLContent := `{{define "content"}}Custom Template Content{{end}}`
	entryHTMLContent := `{{define "entry"}}{{end}}` // Required by HtmlRender

	err = ioutil.WriteFile(filepath.Join(customTemplatesPath, "base.html"), []byte(baseHTMLContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write base.html: %v", err)
	}
	err = ioutil.WriteFile(filepath.Join(customTemplatesPath, "index.html"), []byte(indexHTMLContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write index.html: %v", err)
	}
	err = ioutil.WriteFile(filepath.Join(customTemplatesPath, "entry.html"), []byte(entryHTMLContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write entry.html: %v", err)
	}

	params := &framework.InitParams{
		IsDevMode:                  false, // Avoid esbuild for this test
		AutoRegisterTemplateRoutes: true,
		TemplatesDir:               customTemplatesPath,
		EsbuildOpts:                api.BuildOptions{
			// Provide minimal esbuild options if dev mode were true, not strictly needed here
		},
	}

	router := framework.Run(params)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
		t.Logf("Response body: %s", rr.Body.String())
	}

	expectedContent := "Custom Template Content"
	if !strings.Contains(rr.Body.String(), expectedContent) {
		t.Errorf("handler returned unexpected body: got %s want substring %s",
			rr.Body.String(), expectedContent)
	}
}
