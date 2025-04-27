# Contributing to Globus Go SDK

Thank you for your interest in contributing to the Globus Go SDK! This document provides guidelines and instructions for contributing.

## Code of Conduct

Please be respectful and considerate of others when contributing to this project. We aim to foster an inclusive and welcoming community.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally: `git clone https://github.com/scttfrdmn/globus-go-sdk.git`
3. Add the upstream repository: `git remote add upstream https://github.com/scttfrdmn/globus-go-sdk.git`
4. Create a branch for your work: `git checkout -b your-feature-branch`

## Development Environment Setup

1. Ensure you have Go 1.20 or newer installed
2. Install development tools:
   ```bash
   # Install golangci-lint
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   
   # Install pre-commit
   pip install pre-commit
   pre-commit install
   ```

## Development Process

1. Before starting work, make sure your main branch is up to date:
   ```bash
   git checkout main
   git pull upstream main
   ```

2. Create a branch for your work:
   ```bash
   git checkout -b your-feature-branch
   ```

3. Make your changes and follow these guidelines:
   - Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
   - Include appropriate license headers in new files
   - Write tests for new functionality
   - Keep commit messages clear and descriptive

4. Run tests locally:
   ```bash
   # Run unit tests
   go test ./...
   
   # Run integration tests (if credentials are set up)
   ./scripts/run_integration_tests.sh
   ```

5. Run linters:
   ```bash
   golangci-lint run
   ```

6. Commit your changes with a clear message describing what you've done

7. Push to your fork:
   ```bash
   git push origin your-feature-branch
   ```

8. Create a pull request from your branch to the upstream main branch

## Pull Request Process

1. Ensure all tests pass and linters report no issues
2. Update documentation as needed
3. Add your changes to the CHANGELOG if appropriate
4. Fill out the pull request template completely
5. Request review from maintainers
6. Address any feedback from reviewers

## Testing

### Unit Tests

Unit tests should be added for all new functionality. Run unit tests with:

```bash
go test ./...
```

### Integration Tests

Integration tests interact with real Globus services and require credentials. See [INTEGRATION_TESTING.md](INTEGRATION_TESTING.md) for details.

## Coding Standards

- Follow standard Go conventions and [Effective Go](https://golang.org/doc/effective_go)
- Use meaningful variable and function names
- Add comments for exported functions, types, and constants
- Keep functions focused and concise
- Use proper error handling

## License Headers

All source files must include the proper license header:

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
```

You can use `./scripts/update-license-headers.sh` to add headers to new files.

## Documentation

- Document all exported functions, types, and constants
- Keep README and other docs updated with new features
- Use godoc-compatible comments for API documentation

## Release Process

1. Updates to the CHANGELOG
2. Version number increment following [semantic versioning](https://semver.org/)
3. Tag the new version

## Questions?

If you have questions about contributing, please open an issue or contact the maintainers.

Thank you for contributing to the Globus Go SDK!