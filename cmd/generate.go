package cmd

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type userConfirmationInput interface {
	getConfirmation(prompt string) (bool, error)
	addACs() (string, string, error)
}

type userInput struct{}

func (ui userInput) getConfirmation(prompt string) (bool, error) {
	fmt.Print(prompt)
	var response string
	_, err := fmt.Scanln(&response)

	if err != nil {
		if err.Error() == "unexpected newline" {
			return false, nil
		}

		return false, err
	}

	return response == "y" || response == "Y", nil
}

func (ui userInput) addACs() (string, string, error) {
	acsLinkInput := huh.NewInput().Key("acsLink").Placeholder("ACs Link")
	acsDescriptionInput := huh.NewText().Key("acsDescription").Placeholder("Add ACs here")
	acsDescriptionInput.WithHeight(10)
	acsDescriptionInput.CharLimit(math.MaxInt32)
	acsDescriptionInput.ShowLineNumbers(true)

	form := huh.NewForm(
		huh.NewGroup(
			acsLinkInput,
			acsDescriptionInput,
		),
	)

	form.WithKeyMap(&huh.KeyMap{
		Input: huh.InputKeyMap{
			Next: key.NewBinding(
				key.WithKeys("tab", "enter"),
				key.WithHelp("enter / tab", "next"),
			),
		},
		Text: huh.TextKeyMap{
			Prev: key.NewBinding(
				key.WithKeys("shift+tab"),
				key.WithHelp("shift+tab", "back"),
			),
			Submit: key.NewBinding(
				key.WithKeys("alt+enter", "ctrl+s"),
				key.WithHelp("alt+enter / ctrl+s", "submit"),
			),
			NewLine: key.NewBinding(
				key.WithKeys("enter", "ctrl+j"),
				key.WithHelp("enter / ctrl+j", "new line"),
			),
			Editor: key.NewBinding(
				key.WithKeys("ctrl+e"),
				key.WithHelp("ctrl+e", "open editor"),
			),
		},
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("ctrl+c/esc", "Quit"),
		),
	})

	err := form.Run()

	if err != nil {
		return "", "", err
	}

	return form.GetString("acsLink"), form.GetString("acsDescription"), nil
}

func generateComponentTest(path string) {
	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		printError(err)
		return
	}

	componentPath := transformBasePath(path)

	if strings.HasPrefix(path, "/") {
		baseName := filepath.Base(path)

		if !strings.Contains(baseName, ".") && componentPath == "" {
			componentPath = baseName
		}
	}

	if componentPath == "" {
		componentPath = filepath.Base(currentWorkingDirectory)
	}

	filePath, err := createFilePath(path, componentPath, currentWorkingDirectory)
	if err != nil {
		printError(err)
		return
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		printError(err)
		return
	}

	input := userInput{}

	template := createTemplate(componentPath)

	useAcs, err := input.getConfirmation("\033[36m Generate the boilerplate based on ACs? (y/N): \033[0m")
	if err != nil {
		printError(err)
		return
	}

	if useAcs {
		acsLink, acsText, err := input.addACs()
		if err != nil {
			if err.Error() == "user aborted" {
				return
			}

			printError(err)
			return
		}

		if strings.TrimSpace(acsText) != "" {
			acsBlocks := parseAcs(acsText)
			template = integrateAcsWithTemplate(template, acsLink, acsBlocks)
		}
	}

	err = writeTestFile(filePath, template, input)
	if err != nil {
		if err.Error() == "operation cancelled" {
			return
		} else {
			printError(err)
		}

		return
	}

	fmt.Println("\033[32m Test file generated successfully at", filePath, "\033[0m")
}

func transformBasePath(path string) string {
	if len(path) == 0 {
		return ""
	}

	path = strings.TrimSuffix(path, "Component")

	var result strings.Builder
	result.WriteRune(unicode.ToLower(rune(path[0])))

	for i := 1; i < len(path); i++ {
		if unicode.IsUpper(rune(path[i])) {
			result.WriteRune('-')
			result.WriteRune(unicode.ToLower(rune(path[i])))
		} else {
			result.WriteRune(rune(path[i]))
		}
	}

	basePath := result.String()
	basePath = strings.Split(filepath.Base(basePath), ".")[0]

	return basePath
}

func createFilePath(basePath, componentName, currentWorkingDirectory string) (string, error) {
	fileName := componentName + ".component.spec.ts"

	if basePath == "" {
		return filepath.Join(currentWorkingDirectory, fileName), nil
	}

	if filepath.IsAbs(basePath) {
		return filepath.Join(currentWorkingDirectory, basePath, fileName), nil
	}

	return filepath.Join(currentWorkingDirectory, fileName), nil
}

func writeTestFile(filePath, content string, input userConfirmationInput) error {
	if _, err := os.Stat(filePath); err == nil {
		prompt := fmt.Sprintf("\033[33m Warning: %s already exists. Overwrite? (y/N): \033[0m", filePath)
		confirmed, err := input.getConfirmation(prompt)
		if err != nil {
			return err
		}
		if !confirmed {
			return fmt.Errorf("operation cancelled")
		}
	}

	newFile, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer newFile.Close()

	_, err = newFile.WriteString(content)
	return err
}

func printError(err error) {
	fmt.Printf("\033[31m Error generating test file: %v \033[0m\n", err)
}

func createTemplate(componentPath string) string {
	importName := strings.ToLower(componentPath)
	caser := cases.Title(language.English)
	componentName := caser.String(componentPath)
	componentName = strings.ReplaceAll(componentName, "-", "")

	template := fmt.Sprintf(`
import { TestbedHarnessEnvironment } from '@angular/cdk/testing/testbed';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideMockStore } from '@ngrx/store/testing';
import { render } from '@testing-library/angular';

import { %sComponent } from './%s.component';

/**
* ACs from:
*  - TODO: Link ACs tickets here
*/
describe('%sComponent', () => {
	const mount = async () => {
		const view = await render(%sComponent, {
			providers: [
				provideHttpClient(),
				provideHttpClientTesting(),
				provideMockStore(),
			],
		});

		const httpTestingController = TestBed.inject(HttpTestingController);
		const loader = TestbedHarnessEnvironment.loader(view.fixture);

		return { view, httpTestingController, loader };
	};

	it('should create', async () => {
		const { view } = await mount();
		expect(view.fixture.componentInstance).toBeTruthy();
	});
});
`,
		componentName,
		importName,
		componentName,
		componentName,
	)

	return strings.TrimPrefix(template, "\n")
}
