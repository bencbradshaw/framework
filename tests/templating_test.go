package tests_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	// We need to import the original package to test it.
	// However, to modify package-level variables like `templatesDir`,
	// we might need to use build tags or a more complex setup.
	// For now, let's try to test it by ensuring the "templates" directory
	// is in the expected place relative to the test execution.
	// The `go test` command runs tests from the package directory by default.
	// So, if `templating.HtmlRender` expects "./templates", we need to manage that.

	// A common pattern is to have test utility functions or to copy
	// the target package's code and modify it for tests if it's hard to test.

	// Let's try a direct approach first. The tests will be in `tests` package,
	// and `templating` package is in `../templating`.
	// `templating.HtmlRender` looks for "./templates".
	// We will create a temporary `templates` directory in the root for the duration of the test.
	"github.com/bencbradshaw/framework/templating"
)

// setupTeardownTestTemplates manages the creation and cleanup of a temporary 'templates'
// directory at the project root, copying test templates into it.
// This is to satisfy the hardcoded "./templates" path in templating.HtmlRender.
func setupTeardownTestTemplates(t *testing.T, testTemplatesSourceDir string) func() {
	originalCwd, _ := os.Getwd()
	// Assumes tests package is in 'tests' dir, one level down from project root.
	// So, to get to project root from 'tests' directory where `go test ./tests` runs:
	rootDir := filepath.Join(originalCwd, "..")

	// If running tests from project root (`go test github.com/bencbradshaw/framework/tests`)
	// originalCwd is already the project root.
	// Let's detect if we are in the 'tests' directory or project root.
	if filepath.Base(originalCwd) == "tests" {
		// We are in 'tests', so root is one level up
		rootDir = filepath.Join(originalCwd, "..")
	} else {
		// Assume we are in project root
		rootDir = originalCwd
	}


	tempTemplatesDir := filepath.Join(rootDir, "templates")
	// testTemplatesSourceDir is relative to the `tests` package dir, e.g., "./templates"
	// We need its absolute path or path relative to where this setup function is called (which is the tests dir)
	absTestTemplatesSourceDir, err := filepath.Abs(testTemplatesSourceDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path for testTemplatesSourceDir %s: %v", testTemplatesSourceDir, err)
	}


	// 1. Check if original './templates' exists at project root, if so, back it up
	backupDir := ""
	if _, err := os.Stat(tempTemplatesDir); !os.IsNotExist(err) {
		backupDir = tempTemplatesDir + "_backup_" + strings.ReplaceAll(t.Name(), string(filepath.Separator), "_") // make backup name more robust
		t.Logf("Backing up existing templates directory from %s to %s", tempTemplatesDir, backupDir)
		if err := os.Rename(tempTemplatesDir, backupDir); err != nil {
			// If rename fails (e.g. different devices), try copy and remove
			t.Logf("Rename failed, attempting copy and remove for backup: %v", err)
			// Simple copy directory function needed here, or use external lib.
			// For now, we'll skip robust backup on rename failure to keep it simple.
			// This might leave the original templates dir if it exists and rename fails.
			// Consider adding a proper directory copy utility if this becomes an issue.
			t.Fatalf("Failed to backup existing templates directory (rename failed): %v", err)
		}
	}

	// 2. Create the temporary './templates' dir by copying from our test source
	t.Logf("Creating temporary project root templates directory at %s from %s", tempTemplatesDir, absTestTemplatesSourceDir)
	if err := os.MkdirAll(tempTemplatesDir, 0755); err != nil {
		// Attempt to clean up backup if directory creation fails
		if backupDir != "" {
			os.Rename(backupDir, tempTemplatesDir)
		}
		t.Fatalf("Failed to create temp project root templates dir: %v", err)
	}

	// Copy files from absTestTemplatesSourceDir to tempTemplatesDir
	files, err := os.ReadDir(absTestTemplatesSourceDir)
	if err != nil {
		// Attempt to clean up
		os.RemoveAll(tempTemplatesDir)
		if backupDir != "" {
			os.Rename(backupDir, tempTemplatesDir)
		}
		t.Fatalf("Failed to read test templates source directory %s: %v", absTestTemplatesSourceDir, err)
	}
	for _, file := range files {
		sourceFile := filepath.Join(absTestTemplatesSourceDir, file.Name())
		destFile := filepath.Join(tempTemplatesDir, file.Name())
		data, err := os.ReadFile(sourceFile)
		if err != nil {
			t.Fatalf("Failed to read source template file %s: %v", sourceFile, err)
		}
		if err := os.WriteFile(destFile, data, 0644); err != nil {
			t.Fatalf("Failed to write to destination template file %s: %v", destFile, err)
		}
		t.Logf("Copied %s to %s", sourceFile, destFile)
	}


	// Teardown function
	return func() {
		t.Logf("Cleaning up temporary project root templates directory at %s", tempTemplatesDir)
		if err := os.RemoveAll(tempTemplatesDir); err != nil {
			t.Logf("Warning: failed to remove temporary project root templates directory %s: %v", tempTemplatesDir, err)
		}

		if backupDir != "" {
			t.Logf("Restoring original templates directory from %s to %s", backupDir, tempTemplatesDir)
			if err := os.Rename(backupDir, tempTemplatesDir); err != nil {
				t.Logf("Warning: failed to restore original templates directory from backup: %v", err)
			}
		}
	}
}


func TestHtmlRender_Success(t *testing.T) {
	// The path to our test templates, relative to the 'tests' directory itself
	testTemplatesSource := filepath.Join(".", "templates")
	teardown := setupTeardownTestTemplates(t, testTemplatesSource)
	defer teardown()

	data := map[string]any{
		"title": "Test Page Title",
		"name":  "Tester",
	}

	// Now call HtmlRender. It should look for "./templates/test_page.html" etc.
	// from the project root.
	renderedHTML, err := templating.HtmlRender("test_page.html", data)
	if err != nil {
		t.Fatalf("HtmlRender failed: %v", err)
	}

	// Check for base template content
	if !strings.Contains(renderedHTML, "<title>Base Test - Test Page Title</title>") {
		t.Errorf("Rendered HTML does not contain correct base title. Got:\n%s", renderedHTML)
	}
	if !strings.Contains(renderedHTML, "Base Header") {
		t.Errorf("Rendered HTML does not contain 'Base Header'. Got:\n%s", renderedHTML)
	}
	if !strings.Contains(renderedHTML, "Base Footer") {
		t.Errorf("Rendered HTML does not contain 'Base Footer'. Got:\n%s", renderedHTML)
	}

	// Check for entry template content
	if !strings.Contains(renderedHTML, "<div class=\"entry-point\">") {
		t.Errorf("Rendered HTML does not contain entry point div. Got:\n%s", renderedHTML)
	}
	if !strings.Contains(renderedHTML, "Entry Point Content.") {
		t.Errorf("Rendered HTML does not contain 'Entry Point Content.'. Got:\n%s", renderedHTML)
	}

	// Check for specific page content
	if !strings.Contains(renderedHTML, "<h1>Hello Tester</h1>") {
		t.Errorf("Rendered HTML does not contain '<h1>Hello Tester</h1>'. Got:\n%s", renderedHTML)
	}
	if !strings.Contains(renderedHTML, "This is a test page.") {
		t.Errorf("Rendered HTML does not contain 'This is a test page.'. Got:\n%s", renderedHTML)
	}
}

func TestHtmlRender_TemplateNotFound(t *testing.T) {
	testTemplatesSource := filepath.Join(".", "templates")
	teardown := setupTeardownTestTemplates(t, testTemplatesSource)
	defer teardown()

	data := map[string]any{"title": "Not Found Test"}
	_, err := templating.HtmlRender("non_existent_page.html", data)
	if err == nil {
		t.Errorf("Expected an error when rendering a non-existent template, but got nil")
	} else {
		t.Logf("Got expected error for non-existent template: %v", err)
		if !strings.Contains(err.Error(), "non_existent_page.html") && !strings.Contains(err.Error(),"no such file or directory") {
			// Check for either the filename or the specific OS error part
			t.Errorf("Error message does not contain the name of the missing file or 'no such file or directory'. Got: %s", err.Error())
		}
	}
}

func TestHtmlRender_MalformedTemplate(t *testing.T) {
	testTemplatesSource := filepath.Join(".", "templates")
	teardown := setupTeardownTestTemplates(t, testTemplatesSource)
	defer teardown()

	data := map[string]any{"title": "Malformed Test", "name": "User", "show": true}
	// HtmlRender will try to parse base.html, entry.html, and malformed_template.html
	_, err := templating.HtmlRender("malformed_template.html", data)
	if err == nil {
		t.Errorf("Expected an error when rendering a malformed template, but got nil")
	} else {
		t.Logf("Got expected error for malformed template: %v", err)
		// A more robust check for malformed template errors
		if !strings.Contains(err.Error(), "template: ") || (!strings.Contains(err.Error(), "malformed_template.html") && !strings.Contains(err.Error(), "base.html") && !strings.Contains(err.Error(), "entry.html")) {
		// The error could be in any of the parsed files if it's a definition error,
		// or specifically in malformed_template.html if it's during its specific parsing.
		// The error usually starts with "template: ".
			t.Errorf("Error message does not seem to indicate a template parsing error related to the involved files. Got: %s", err.Error())
		}
	}
}
