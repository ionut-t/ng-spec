# ng-spec

A simple CLI tool for generating a minimal Angular component integration test boilerplate using [Angular Testing Library](https://testing-library.com/docs/angular-testing-library/intro/).

## Installation

```bash
# Using Go
go install github.com/ionut-t/ng-spec@latest
```

Or install the binary from the [Releases page](https://github.com/ionut-t/ng-spec/releases)

## Features

- Quickly generate boilerplate test files for Angular components
- Interactive AC (Acceptance Criteria) parsing to auto-generate test blocks
- Uses Angular Testing Library for modern, user-centric testing
- Configures common testing providers (HTTP, Store)
- Works with both relative and absolute component paths
- Runs with no arguments to generate a test for the current directory

## Usage

### Basic Usage

```bash
# Run with no arguments to generate a test based on the current directory name
ng-spec

# Generate a test for a specific component
ng-spec my-component

# Generate a test for a component in a specific path
ng-spec path/to/my-component

# Use an absolute path as the component name
ng-spec /dashboard
# This will create dashboard.component.spec.ts in the dashboard directory (under your current working directory)
```

### Using Acceptance Criteria

When running any of the commands above, you'll be prompted if you want to generate test blocks based on ACs:

```
Would you like to generate the boilerplate based on ACs? (y/N):
```

If you select "y", you'll be prompted to enter:

1. ACs link (e.g., JIRA ticket number)
2. ACs content in a structured format

#### AC Format Example

```
1. Feature/Category (creates a describe block)
  a. Test Case (creates an it block)
  b. Nested Feature (creates a nested describe block)
    i. Nested Test Case (creates an it block inside nested describe)
    ii. Another Nested Test Case
```

## Generated Test Structure

Each generated test includes:

- Proper imports for Angular Testing Library
- HTTP testing setup
- NgRx store mock setup
- TestbedHarnessEnvironment for component testing
- A basic "should create" test
- Generated test blocks based on your ACs (if provided)

### Example with ACs:

```typescript
import { TestbedHarnessEnvironment } from '@angular/cdk/testing/testbed';
import { provideHttpClient } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { TestBed } from '@angular/core/testing';
import { provideMockStore } from '@ngrx/store/testing';
import { render } from '@testing-library/angular';

import { DashboardComponent } from './dashboard.component';

/**
 * ACs from:
 *  - TODO: Link ACs tickets here
 */
describe('DashboardComponent', () => {
  const mount = async () => {
    const view = await render(DashboardComponent, {
      providers: [provideHttpClient(), provideHttpClientTesting(), provideMockStore()]
    });

    const httpTestingController = TestBed.inject(HttpTestingController);
    const loader = TestbedHarnessEnvironment.loader(view.fixture);

    return { view, httpTestingController, loader };
  };

  it('should create', async () => {
    const { view } = await mount();
    expect(view.fixture.componentInstance).toBeTruthy();
  });

  describe('User Management', () => {
    it('should create user', async () => {
      // TODO: Implement test
    });

    it('should delete user', async () => {
      // TODO: Implement test
    });
  });

  describe('Dashboard Features', () => {
    describe('Data Visualization', () => {
      it('should display charts', async () => {
        // TODO: Implement test
      });

      it('should refresh data', async () => {
        // TODO: Implement test
      });
    });
  });
});
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

