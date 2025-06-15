package tests_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bencbradshaw/framework/env"
)

// cleanEnvVars is a helper to unset environment variables set during tests
func cleanEnvVars(vars ...string) {
	for _, v := range vars {
		os.Unsetenv(v)
	}
}

func TestLoadEnvFile_Success(t *testing.T) {
	testEnvPath := filepath.Join(".", "test.env") // Assumes test.env is in the same dir as the test executable (which is 'tests')

	// Ensure vars are not set before test
	cleanEnvVars("TEST_VAR1", "TEST_VAR2", "SPACED_VAR")
	defer cleanEnvVars("TEST_VAR1", "TEST_VAR2", "SPACED_VAR")

	loadedVars, err := env.LoadEnvFile(testEnvPath)
	if err != nil {
		t.Fatalf("LoadEnvFile failed: %v", err)
	}

	if loadedVars == nil {
		t.Fatal("loadedVars map is nil")
	}

	expectedVars := map[string]string{
		"TEST_VAR1":  "Hello World",
		"TEST_VAR2":  "12345",
		"SPACED_VAR": "spaced value", // Values should be trimmed
	}

	for key, expectedValue := range expectedVars {
		// Check from returned map
		val, ok := loadedVars.Vars[key]
		if !ok {
			t.Errorf("Expected variable %s not found in loadedVars map", key)
		}
		if valStr, ok := val.(string); !ok || valStr != expectedValue {
			t.Errorf("Variable %s: expected map value '%s', got '%v'", key, expectedValue, val)
		}

		// Check from os.Getenv
		envVal := os.Getenv(key)
		if envVal != expectedValue {
			t.Errorf("Variable %s: expected os.Getenv value '%s', got '%s'", key, expectedValue, envVal)
		}
	}

	if _, ok := loadedVars.Vars["MALFORMED_LINE"]; ok {
		t.Error("MALFORMED_LINE should not have been loaded")
	}
	if os.Getenv("MALFORMED_LINE") != "" {
		t.Error("MALFORMED_LINE should not be set in environment")
	}

	if _, ok := loadedVars.Vars["COMMENT_VAR"]; ok {
		t.Error("COMMENT_VAR should not have been loaded")
	}
	if os.Getenv("COMMENT_VAR") != "" {
		t.Error("COMMENT_VAR should not be set in environment")
	}
}

func TestLoadEnvFile_NotFound(t *testing.T) {
	nonExistentPath := filepath.Join(".", "non_existent.env")
	_, err := env.LoadEnvFile(nonExistentPath)
	if err == nil {
		t.Errorf("Expected an error when loading a non-existent file, but got nil")
	}
	// In the current Framework.go, it logs "No .env file loaded" but doesn't return an error from main execution flow.
	// env.LoadEnvFile itself *does* return an error if the file doesn't exist. This test confirms that.
}

func TestLoadEnvFile_EmptyPath(t *testing.T) {
	_, err := env.LoadEnvFile("")
	if err == nil {
		t.Errorf("Expected an error when loading with an empty path, but got nil")
	}
}

func TestLoadEnvFile_EmptyFile(t *testing.T) {
	emptyEnvPath := filepath.Join(".", "empty.env")
	// Create the empty file
	file, createErr := os.Create(emptyEnvPath)
	if createErr != nil {
		t.Fatalf("Failed to create empty.env: %v", createErr)
	}
	file.Close()
	defer os.Remove(emptyEnvPath) // Clean up

	loadedVars, err := env.LoadEnvFile(emptyEnvPath)
	if err != nil {
		t.Fatalf("LoadEnvFile failed for empty file: %v", err)
	}
	if loadedVars == nil {
		t.Fatal("loadedVars map is nil for empty file")
	}
	if len(loadedVars.Vars) != 0 {
		t.Errorf("Expected loadedVars map to be empty, but got %d items", len(loadedVars.Vars))
	}
}

func TestLoadEnvFile_MalformedLineDoesNotStopLoading(t *testing.T) {
	testEnvPath := filepath.Join(".", "test.env") // Uses the existing test.env with a malformed line

	cleanEnvVars("TEST_VAR1", "SPACED_VAR") // Clean vars that might be set by other tests or this one
	defer cleanEnvVars("TEST_VAR1", "SPACED_VAR")

	_, err := env.LoadEnvFile(testEnvPath)
	if err != nil {
		t.Fatalf("LoadEnvFile failed: %v", err)
	}

	// Check if a variable defined AFTER the malformed line is still loaded
	// In our current test.env, SPACED_VAR is after MALFORMED_LINE.
	expectedValue := "spaced value"
	envVal := os.Getenv("SPACED_VAR")
	if envVal != expectedValue {
		t.Errorf("SPACED_VAR: expected os.Getenv value '%s', got '%s' (checking if loading continued past malformed line)", expectedValue, envVal)
	}
}
