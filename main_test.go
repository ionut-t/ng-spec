package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainIntegration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	tempDir, err := os.MkdirTemp("", "ng-spec-integration")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	binaryPath := filepath.Join(tempDir, "ng-spec")

	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	buildCmd.Dir = projectRoot // Set working directory to project root

	var buildOutput bytes.Buffer
	buildCmd.Stdout = &buildOutput
	buildCmd.Stderr = &buildOutput

	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v\nBuild output: %s", err, buildOutput.String())
	}

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("Binary was not created at %s", binaryPath)
	}

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		name            string
		args            []string
		expectedFile    string
		expectedContent string
	}{
		{
			name:            "Default command",
			args:            []string{},
			expectedFile:    filepath.Base(tempDir) + ".component.spec.ts",
			expectedContent: filepath.Base(tempDir) + "Component",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Run the command
			cmd := exec.Command(binaryPath, tc.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()

			if err != nil {
				t.Fatalf("Command failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
			}

			expectedFilePath := filepath.Join(tempDir, tc.expectedFile)
			if _, err := os.Stat(expectedFilePath); os.IsNotExist(err) {
				t.Fatalf("Expected file was not created: %s", expectedFilePath)
			}

			content, err := os.ReadFile(expectedFilePath)
			if err != nil {
				t.Fatalf("Failed to read expected file: %v", err)
			}

			if contentStr := strings.ToLower(string(content)); !strings.Contains(contentStr, strings.ToLower(tc.expectedContent)) {
				t.Errorf("Expected content %q not found in file content: %s",
					tc.expectedContent, string(content))
			}
		})
	}
}

func findProjectRoot() (string, error) {
	// Start with the current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

func TestHelpCommand(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir, err := os.MkdirTemp("", "ng-spec-help-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	binaryPath := filepath.Join(tempDir, "ng-spec")
	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	cmd := exec.Command(binaryPath, "--help")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Help command failed: %v", err)
	}

	output := stdout.String()
	expectedPhrases := []string{
		"ng-spec",
		"Angular components",
		"Usage:",
		"Flags:",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Help output doesn't contain expected phrase: %s", phrase)
		}
	}
}

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}
