package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCmdHelp(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the root command with --help flag
	oldArgs := os.Args
	os.Args = []string{"ng-spec", "--help"}

	defer func() {
		if r := recover(); r != nil {
			// This is expected since help may call os.Exit
		}
		// Restore args and stdout
		os.Args = oldArgs
		w.Close()
		os.Stdout = old
	}()

	rootCmd.Help()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedPhrases := []string{
		"ng-spec",
		"Angular components",
		"Testing Library",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Help output doesn't contain expected phrase: %s", phrase)
		}
	}
}

func TestRootCmdExecute(t *testing.T) {
	originalRun := rootCmd.Run
	defer func() { rootCmd.Run = originalRun }()

	executed := false
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		executed = true
	}

	err := rootCmd.Execute()

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !executed {
		t.Error("Run function was not executed")
	}
}

func TestRootCmdRunFunction(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ng-spec-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.Run(rootCmd, []string{})

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "generated successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	baseName := fmt.Sprintf("%s.component.spec.ts", filepath.Base(tempDir))
	_, err = os.Stat(baseName)
	if os.IsNotExist(err) {
		t.Errorf("Expected file %s to be created", baseName)
	}
}
