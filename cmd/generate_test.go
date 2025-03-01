package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractComponentName(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Empty path", "", ""},
		{"Simple component", "user", "user"},
		{"Component with extension", "user.component.ts", "user"},
		{"Component with Component suffix", "userComponent", "user"},
		{"Path with directory", "/path/to/user", "user"},
		{"Path with directory and extension", "/path/to/user.component.ts", "user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractComponentName(tt.path)
			if result != tt.expected {
				t.Errorf("extractComponentName(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestCreateFilePath(t *testing.T) {
	// Get current working directory for test
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name                  string
		basePath              string
		componentName         string
		currentWorkingDir     string
		expectedPathSuffix    string
		expectErrorContaining string
	}{
		{
			name:               "Empty path",
			basePath:           "",
			componentName:      "user",
			currentWorkingDir:  cwd,
			expectedPathSuffix: "user.component.spec.ts",
		},
		{
			name:               "Relative path",
			basePath:           "src/app",
			componentName:      "user",
			currentWorkingDir:  cwd,
			expectedPathSuffix: "user.component.spec.ts",
		},
		{
			name:               "Absolute path",
			basePath:           "/src/app",
			componentName:      "user",
			currentWorkingDir:  cwd,
			expectedPathSuffix: filepath.Join(cwd, "/src/app/user.component.spec.ts"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := createFilePath(tt.basePath, tt.componentName, tt.currentWorkingDir)

			if tt.expectErrorContaining != "" {
				if err == nil {
					t.Errorf("createFilePath(%q, %q, %q) expected error containing %q, got nil",
						tt.basePath, tt.componentName, tt.currentWorkingDir, tt.expectErrorContaining)
				} else if !strings.Contains(err.Error(), tt.expectErrorContaining) {
					t.Errorf("createFilePath(%q, %q, %q) error %q does not contain %q",
						tt.basePath, tt.componentName, tt.currentWorkingDir, err.Error(), tt.expectErrorContaining)
				}
				return
			}

			if err != nil {
				t.Errorf("createFilePath(%q, %q, %q) unexpected error: %v",
					tt.basePath, tt.componentName, tt.currentWorkingDir, err)
				return
			}

			if tt.basePath == "" || !filepath.IsAbs(tt.basePath) {
				if !strings.HasSuffix(result, tt.expectedPathSuffix) {
					t.Errorf("createFilePath(%q, %q, %q) = %q, want path ending with %q",
						tt.basePath, tt.componentName, tt.currentWorkingDir, result, tt.expectedPathSuffix)
				}
			} else if result != tt.expectedPathSuffix {
				t.Errorf("createFilePath(%q, %q, %q) = %q, want %q",
					tt.basePath, tt.componentName, tt.currentWorkingDir, result, tt.expectedPathSuffix)
			}
		})
	}
}

func TestCreateTemplate(t *testing.T) {
	tests := []struct {
		name           string
		componentName  string
		expectedPhrase string
	}{
		{"Simple component", "user", "UserComponent"},
		{"Lowercase component", "profile", "ProfileComponent"},
		{"Component with spaces", "user profile", "User ProfileComponent"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createTemplate(tt.componentName)
			if !strings.Contains(result, tt.expectedPhrase) {
				t.Errorf("createTemplate(%q) does not contain %q", tt.componentName, tt.expectedPhrase)
			}
		})
	}
}

func TestGenerateComponentTest(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "ng-spec-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for test
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name            string
		path            string
		expectedFile    string
		expectedContent string
	}{
		{
			name:            "Default with empty path",
			path:            "",
			expectedFile:    filepath.Base(tempDir) + ".component.spec.ts",
			expectedContent: filepath.Base(tempDir),
		},
		{
			name:            "With explicit component name",
			path:            "user",
			expectedFile:    "user.component.spec.ts",
			expectedContent: "UserComponent",
		},
		{
			name:            "With absolute path",
			path:            "/profile", // Using a simple absolute path
			expectedFile:    filepath.Join(tempDir, "profile", "profile.component.spec.ts"),
			expectedContent: "ProfileComponent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear any existing files
			files, err := os.ReadDir(tempDir)
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".component.spec.ts") {
					os.Remove(filepath.Join(tempDir, file.Name()))
				}
			}

			if strings.Contains(tt.expectedFile, filepath.Join(tempDir, "profile")) {
				dir := filepath.Dir(tt.expectedFile)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
			}

			// Capture stdout to avoid polluting test output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			generateComponentTest(tt.path)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Consume stdout to prevent hanging
			var buf bytes.Buffer
			io.Copy(&buf, r)

			if _, err := os.Stat(tt.expectedFile); os.IsNotExist(err) {
				t.Errorf("generateComponentTest(%q) did not create expected file at %s",
					tt.path, tt.expectedFile)
				return
			}

			content, err := os.ReadFile(tt.expectedFile)
			if err != nil {
				t.Fatal(err)
			}

			if !strings.Contains(string(content), tt.expectedContent) {
				t.Errorf("generateComponentTest(%q) file does not contain expected content %q",
					tt.path, tt.expectedContent)
			}
		})
	}
}

type mockUserInput struct {
	Response bool
}

func (m mockUserInput) getConfirmation(prompt string) (bool, error) {
	return m.Response, nil
}

type errorUserInput struct{}

func (e errorUserInput) getConfirmation(prompt string) (bool, error) {
	return false, fmt.Errorf("simulated error")
}

func TestWriteTestFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ng-spec-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Test case 1: Creating a new file (doesn't require confirmation)
	filePath := filepath.Join(tempDir, "new.component.spec.ts")

	mockInput := mockUserInput{Response: true}

	err = writeTestFile(filePath, "new", mockInput)
	if err != nil {
		t.Errorf("writeTestFile() error creating new file: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "NewComponent") {
		t.Errorf("writeTestFile() did not create file with expected content")
	}

	// Test case 2: Overwriting a file with confirmation (yes)
	yesInput := mockUserInput{Response: true}

	err = writeTestFile(filePath, "overwrite", yesInput)
	if err != nil {
		t.Errorf("writeTestFile() error with 'yes' confirmation: %v", err)
	}

	// Check that file was overwritten
	content, err = os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "OverwriteComponent") {
		t.Errorf("writeTestFile() did not overwrite file with new content")
	}

	// Test case 3: Attempting to overwrite a file but denying confirmation (no)
	noInput := mockUserInput{Response: false}

	err = writeTestFile(filePath, "denied", noInput)
	if err == nil {
		t.Errorf("writeTestFile() expected error when confirmation denied")
	}

	if err != nil && !strings.Contains(err.Error(), "cancelled") {
		t.Errorf("writeTestFile() error = %v, want 'operation cancelled'", err)
	}

	content, err = os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(content), "OverwriteComponent") {
		t.Errorf("writeTestFile() file content changed despite cancelled overwrite")
	}

	errorInput := errorUserInput{}

	err = writeTestFile(filePath, "error", errorInput)
	if err == nil {
		t.Errorf("writeTestFile() expected error when confirmation fails")
	}

	if err != nil && !strings.Contains(err.Error(), "simulated error") {
		t.Errorf("writeTestFile() error = %v, want 'simulated error'", err)
	}
}

func TestPrintError(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printError(fmt.Errorf("test error"))

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test error") {
		t.Errorf("printError() output = %q, want to contain 'test error'", output)
	}

	if !strings.Contains(output, "\033[31m") {
		t.Errorf("printError() output = %q, want to contain red color code", output)
	}
}
