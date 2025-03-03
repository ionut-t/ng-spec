package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type userConfirmationInput interface {
	getConfirmation(prompt string) (bool, error)
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

	if err := writeTestFile(filePath, componentPath, input); err != nil {
		if err.Error() == "operation cancelled" {
			fmt.Println("\033[33m Operation cancelled \033[0m")
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

func writeTestFile(filePath, componentPath string, input userConfirmationInput) error {
	if _, err := os.Stat(filePath); err == nil {
		prompt := fmt.Sprintf("\033[33mWarning: %s already exists. Overwrite? (y/N): \033[0m", filePath)
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

	_, err = newFile.WriteString(createTemplate(componentPath))
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
