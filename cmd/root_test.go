package cmd

import (
	"bytes"
	"io"
	"os"
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
