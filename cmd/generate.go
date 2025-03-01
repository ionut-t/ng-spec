package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type userInput interface {
	getConfirmation(prompt string) (bool, error)
}

type realUserInput struct{}

func (ui realUserInput) getConfirmation(prompt string) (bool, error) {
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
	println("generateComponentTest", path)

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		printError(err)
		return
	}

	componentName := extractComponentName(path)

	if strings.HasPrefix(path, "/") {
		baseName := filepath.Base(path)

		if !strings.Contains(baseName, ".") && componentName == "" {
			componentName = baseName
		}
	}

	if componentName == "" {
		componentName = filepath.Base(currentWorkingDirectory)
	}

	filePath, err := createFilePath(path, componentName, currentWorkingDirectory)
	if err != nil {
		printError(err)
		return
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		printError(err)
		return
	}

	input := realUserInput{}

	if err := writeTestFile(filePath, componentName, input); err != nil {
		if err.Error() == "operation cancelled" {
			fmt.Println("\033[33m Operation cancelled \033[0m")
		} else {
			printError(err)
		}

		return
	}

	fmt.Println("\033[32m Test file generated successfully at", filePath, "\033[0m")
}

func extractComponentName(path string) string {
	componentName := strings.Split(filepath.Base(path), ".")[0]
	return strings.TrimSuffix(componentName, "Component")
}

func createFilePath(basePath, componentName, currentWorkingDirectory string) (string, error) {
	fileName := componentName + ".component.spec.ts"

	println("basePath", basePath)

	if basePath == "" {
		return filepath.Join(currentWorkingDirectory, fileName), nil
	}

	println(filepath.IsAbs(basePath))

	// If path is absolute, join with current working directory
	// so it can be used as: ng-spec /path/to/component
	if filepath.IsAbs(basePath) {
		return filepath.Join(currentWorkingDirectory, basePath, fileName), nil
	}

	return filepath.Join(currentWorkingDirectory, fileName), nil
}

func writeTestFile(filePath, componentName string, input userInput) error {
	println(filePath)

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

	_, err = newFile.WriteString(createTemplate(componentName))
	return err
}

func printError(err error) {
	fmt.Printf("\033[31m Error generating test file: %v \033[0m\n", err)
}

func createTemplate(componentName string) string {
	caser := cases.Title(language.English)
	componentName = caser.String(componentName)
	importName := strings.ToLower(componentName)

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
