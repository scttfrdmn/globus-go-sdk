<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Development Guide

This document provides detailed information for developers working on the Globus Go SDK.

## Development Environment Setup

### Prerequisites

- Go 1.21 or higher
- Git
- Make
- golangci-lint
- pre-commit (requires Python)

### Initial Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/scttfrdmn/globus-go-sdk.git
   cd globus-go-sdk
   ```

2. Install development tools:
   ```bash
   make setup
   ```

   This installs:
   - golangci-lint for linting
   - goimports for import formatting
   - pre-commit hooks
   - Go dependencies

## Development Workflow

### Code Style and Standards

This project enforces code standards through multiple tools:

1. **Pre-commit hooks**: Run automatically before each commit to check:
   - SPDX license headers
   - Code formatting
   - Import organization
   - Go vetting
   - YAML validation
   - Trailing whitespace and EOF issues

2. **golangci-lint**: A comprehensive Go linter aggregator:
   ```bash
   make lint
   ```

3. **go fmt**: Standard Go code formatter:
   ```bash
   make fmt
   ```

4. **go vet**: Static analyzer for common mistakes:
   ```bash
   make vet
   ```

### Testing

#### Running Tests

Run all tests:
```bash
make test
```

Run tests with coverage reports:
```bash
make test-coverage
```

Run integration tests (requires Globus credentials):
```bash
make test-integration
```

#### Test Structure

- Unit tests: `*_test.go` files in the same package as the code they test
- Integration tests: `*_integration_test.go` files with build tag `integration`
- Test fixtures: Located in `testdata/` directories

#### Test Coverage Goals

- Aim for >80% code coverage
- All exported functions must have tests
- Error paths should be tested
- Mock external dependencies for unit testing

### Continuous Integration

The project uses GitHub Actions for CI with several workflows:

1. **Main workflow** (`go.yml`):
   - Runs on each PR and push to main
   - Performs linting, testing, and building
   - Checks multiple Go versions

2. **CodeQL Analysis** (`codeql.yml`):
   - Runs security scanning
   - Identifies potential vulnerabilities

3. **Code Coverage** (`codecov.yml`):
   - Generates and uploads coverage reports
   - Provides visibility on test coverage

### Pre-commit Hooks

Pre-commit hooks run automatically before each commit to catch issues early:

- To install hooks:
  ```bash
  pre-commit install
  ```

- To run hooks manually:
  ```bash
  pre-commit run --all-files
  ```

## Project Structure and Architecture

### Package Organization

```
github.com/scttfrdmn/globus-go-sdk/
├── pkg/                         # Main SDK code
│   ├── core/                    # Core SDK functionality
│   │   ├── authorizers/         # Authentication mechanisms
│   │   ├── config/              # Configuration management
│   │   ├── transport/           # HTTP transport layer
│   ├── services/                # Service-specific clients
│   │   ├── auth/                # Authentication service
│   │   ├── groups/              # Groups service
│   │   └── ...                  # Other services
├── internal/                    # Private implementation details
├── cmd/                         # Example applications
└── doc/                         # Documentation
```

### Design Patterns

1. **Interface-based design**: Define interfaces for key components to enable mocking and testing
2. **Functional options**: Use functional options pattern for configuration
3. **Context propagation**: All API operations accept a context for cancellation
4. **Error wrapping**: Use error wrapping to provide context

## Documentation Standards

### Code Documentation

- All exported symbols (functions, types, variables) must have documentation comments
- Use examples where appropriate
- Follow GoDoc conventions

### Project Documentation

- `README.md`: Project overview, installation, basic usage
- `CONTRIBUTING.md`: Contribution guidelines
- `doc/ARCHITECTURE.md`: Design and architecture details
- `doc/DEVELOPMENT.md`: This file, development guide

## Working with Globus Auth

For integration tests and examples, you'll need Globus credentials:

1. Set up environment variables:
   ```bash
   export GLOBUS_CLIENT_ID=your-client-id
   export GLOBUS_CLIENT_SECRET=your-client-secret
   ```

2. To skip integration tests when credentials aren't available:
   ```bash
   export SKIP_INTEGRATION=true
   ```

## Release Process

1. Update version in `pkg/globus.go`
2. Update CHANGELOG.md
3. Tag the release in git
4. Push the tag
5. GitHub Actions will automatically build and release

## Common Issues and Debugging

### Debugging Test Failures

- Run failed tests with verbose output:
  ```bash
  go test -v -run TestSpecificTest ./path/to/package
  ```

- Use environment variable for debug logging:
  ```bash
  GLOBUS_SDK_LOG_LEVEL=debug go test ./...
  ```

### Working with Pre-commit Hooks

- If pre-commit hooks are causing issues, temporarily skip with:
  ```bash
  git commit --no-verify
  ```

- Update pre-commit hooks:
  ```bash
  pre-commit autoupdate
  ```

## Getting Help

- Create an issue in the GitHub repository
- Review existing documentation
- Check for similar issues that may have been resolved