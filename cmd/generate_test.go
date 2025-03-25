package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Mock implementation for userInput.addACs
type mockUserInput struct {
	confirmationResponse bool
	acsLink              string
	acsText              string
}

func (m mockUserInput) getConfirmation(prompt string) (bool, error) {
	return m.confirmationResponse, nil
}

func (m mockUserInput) addACs(prompt string) (string, string, error) {
	return m.acsLink, m.acsText, nil
}

func TestTransformBasePath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"Empty path", "", ""},
		{"Simple component", "user", "user"},
		{"Component with extension", "user.component.ts", "user"},
		{"Component with Component suffix", "userComponent", "user"},
		{"Component with name", "UserProfile", "user-profile"},
		{"Path with directory", "/path/to/user", "user"},
		{"Path with directory and extension", "/path/to/user.component.ts", "user"},
		{"Component with multiple capital letters", "UserProfileSettings", "user-profile-settings"},
		{"Component with numbers", "User2Factor", "user2-factor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transformBasePath(tt.path)
			if result != tt.expected {
				t.Errorf("transformBasePath(%q) = %q, want %q", tt.path, result, tt.expected)
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
		name               string
		basePath           string
		componentName      string
		currentWorkingDir  string
		expectedPathSuffix string
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
		{
			name:               "With dashes in component name",
			basePath:           "",
			componentName:      "user-profile",
			currentWorkingDir:  cwd,
			expectedPathSuffix: "user-profile.component.spec.ts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := createFilePath(tt.basePath, tt.componentName, tt.currentWorkingDir)

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
		name            string
		componentName   string
		expectedPhrases []string
	}{
		{
			name:          "Simple component",
			componentName: "user",
			expectedPhrases: []string{
				"UserComponent",
				"import { UserComponent } from './user.component';",
				"const mount = async () => {",
				"it('should create', async () => {",
				"ACs from:",
			},
		},
		{
			name:          "Component with dashes",
			componentName: "user-profile",
			expectedPhrases: []string{
				"UserProfileComponent",
				"import { UserProfileComponent } from './user-profile.component';",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createTemplate(tt.componentName)

			for _, phrase := range tt.expectedPhrases {
				if !strings.Contains(result, phrase) {
					t.Errorf("createTemplate(%q) does not contain expected phrase: %q", tt.componentName, phrase)
				}
			}
		})
	}
}

func TestWriteTestFile(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "ng-spec-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	filePath := filepath.Join(tempDir, "test.component.spec.ts")
	content := "describe('Test', () => {});"

	// Test case 1: Write to a new file
	mockInput := mockUserInput{confirmationResponse: true}
	err = writeTestFile(filePath, content, mockInput)

	if err != nil {
		t.Errorf("writeTestFile() failed to write new file: %v", err)
	}

	// Check if file was created with the right content
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(fileContent) != content {
		t.Errorf("writeTestFile() wrote incorrect content: got %q, want %q", string(fileContent), content)
	}

	// Test case 2: Overwrite existing file with confirmation
	newContent := "describe('Updated', () => {});"
	mockConfirmInput := mockUserInput{confirmationResponse: true}

	err = writeTestFile(filePath, newContent, mockConfirmInput)
	if err != nil {
		t.Errorf("writeTestFile() failed to overwrite file with confirmation: %v", err)
	}

	fileContent, err = os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(fileContent) != newContent {
		t.Errorf("writeTestFile() did not properly overwrite file: got %q, want %q", string(fileContent), newContent)
	}

	// Test case 3: Attempt to overwrite without confirmation
	finalContent := "describe('Final', () => {});"
	mockDenyInput := mockUserInput{confirmationResponse: false}

	err = writeTestFile(filePath, finalContent, mockDenyInput)
	if err == nil || err.Error() != "operation cancelled" {
		t.Errorf("writeTestFile() should have been cancelled, got error: %v", err)
	}

	// Check file still has previous content
	fileContent, err = os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if string(fileContent) != newContent {
		t.Errorf("writeTestFile() should not have changed file content when confirmation denied: got %q, want %q", string(fileContent), newContent)
	}
}

func TestGenerateComponentTestWithACs(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "ng-spec-acs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Save current working directory and change to temp directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	// Set up test cases
	tests := []struct {
		name              string
		componentName     string
		useACs            bool
		acsLink           string
		acsText           string
		expectedFile      string
		expectedContent   []string
		unexpectedContent []string
	}{
		{
			name:          "Generate with ACs",
			componentName: "user",
			useACs:        true,
			acsLink:       "JIRA-123",
			acsText:       "1. Create user\na. Enter valid data\nb. Submit form",
			expectedFile:  "user.component.spec.ts",
			expectedContent: []string{
				"UserComponent",
				"JIRA-123",
				"describe('Create user'",
				"it('should enter valid data'",
				"it('should submit form'",
			},
		},
		{
			name:          "Generate without ACs",
			componentName: "profile",
			useACs:        false,
			acsLink:       "", // Not used
			acsText:       "", // Not used
			expectedFile:  "profile.component.spec.ts",
			expectedContent: []string{
				"ProfileComponent",
				"TODO: Link ACs tickets here", // Default placeholder remains
			},
			unexpectedContent: []string{
				"describe('Create user'",
			},
		},
		{
			name:          "Generate with ACs link but empty ACs content",
			componentName: "empty",
			useACs:        true,
			acsLink:       "JIRA-456",
			acsText:       "", // Empty ACs
			expectedFile:  "empty.component.spec.ts",
			expectedContent: []string{
				"EmptyComponent",
				"JIRA-456", // Link should still be updated
			},
			unexpectedContent: []string{
				"describe('Create", // No ACs blocks should be generated
			},
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up any existing files
			files, _ := filepath.Glob(filepath.Join(tempDir, "*.component.spec.ts"))
			for _, f := range files {
				os.Remove(f)
			}

			// Create a mock for the userConfirmationInput interface
			mockInput := mockUserInput{
				confirmationResponse: tc.useACs,
				acsLink:              tc.acsLink,
				acsText:              tc.acsText,
			}

			// Create the file path
			filePath := filepath.Join(tempDir, tc.expectedFile)

			// Generate the template based on inputs
			template := createTemplate(tc.componentName)

			if tc.useACs && tc.acsText != "" {
				acsBlocks := parseAcs(tc.acsText)
				template = integrateAcsWithTemplate(template, tc.acsLink, acsBlocks)
			} else if tc.useACs && tc.acsText == "" && tc.acsLink != "" {
				// If AC text is empty but link is provided, still update the link
				template = integrateAcsWithTemplate(template, tc.acsLink, "")
			}

			// Write to file
			err := writeTestFile(filePath, template, mockInput)
			if err != nil {
				t.Errorf("Failed to write test file: %v", err)
				return
			}

			// Verify file was created
			_, err = os.Stat(filePath)
			if os.IsNotExist(err) {
				t.Errorf("Expected file %s was not created", filePath)
				return
			}

			// Read file content
			content, err := os.ReadFile(filePath)
			if err != nil {
				t.Errorf("Failed to read file: %v", err)
				return
			}

			contentStr := string(content)

			// Check for expected content
			for _, expectedStr := range tc.expectedContent {
				if !strings.Contains(contentStr, expectedStr) {
					t.Errorf("Expected content %q not found in generated file", expectedStr)
				}
			}

			// Check that unwanted content is not there
			for _, unexpectedStr := range tc.unexpectedContent {
				if strings.Contains(contentStr, unexpectedStr) {
					t.Errorf("Unexpected content %q found in generated file", unexpectedStr)
				}
			}
		})
	}
}
