# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Contributing to Globus Go SDK

Thank you for your interest in contributing to the Globus Go SDK! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project follows the [Contributor Covenant](https://www.contributor-covenant.org/) Code of Conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR-USERNAME/globus-go-sdk.git
   cd globus-go-sdk
   ```
3. **Add the upstream repository**:
   ```bash
   git remote add upstream https://github.com/ORIGINAL-OWNER/globus-go-sdk.git
   ```
4. **Install dependencies**:
   ```bash
   go mod tidy
   ```

## Development Workflow

1. **Create a new branch** for your feature or bugfix:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Install Git hooks** for local testing:
   ```bash
   ./scripts/install-all-hooks.sh
   ```
   These hooks will run tests and checks locally before commits and pushes.

3. **Write code** following Go best practices:
   - Follow [Effective Go](https://golang.org/doc/effective_go)
   - Adhere to [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
   - Run `gofmt -s -w .` to format your code

4. **Write tests** for your code:
   - Unit tests are required for all new code
   - Aim for >80% code coverage
   - Use Go's standard testing package

5. **Document your code**:
   - Add comments for exported functions, types, and constants
   - Follow [godoc](https://blog.golang.org/godoc) conventions
   - Update project documentation if needed

6. **Commit your changes**:
   - Use descriptive commit messages
   - Reference issue numbers in commit messages if applicable
   - SPDX license identifier and copyright headers must be preserved
   - Git hooks will run tests before each commit

7. **Keep your branch updated**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

8. **Submit a pull request**:
   - Clearly describe what the PR does
   - Mention any related issues
   - Request reviews from maintainers

## Pull Request Process

1. Ensure all tests pass and the code builds
2. Update documentation if necessary
3. Follow the PR template
4. Address review comments in a timely manner
5. Once approved, a maintainer will merge your PR

## Coding Standards

### Go Version

This project targets Go 1.21 and above.

### Code Style

- Use `gofmt` (or `goimports`) to format your code
- Follow standard Go naming conventions:
  - MixedCaps or mixedCaps (not snake_case)
  - Short, descriptive names
  - Acronyms should be all uppercase (HTTP, URL, etc.)
- Keep functions focused and small
- Use meaningful variable names

### Package Structure

Maintain the established package structure:

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
```

### Testing

- Write both unit and integration tests
- Test files should be named `*_test.go`
- Use Go's built-in testing package
- Mock external services when appropriate
- Include examples in documentation

### Error Handling

- Use meaningful error messages
- Return errors rather than panicking
- Use custom error types when appropriate
- Wrap errors for better context

### Documentation

- Document all exported functions, types, and variables
- Include examples where appropriate
- Keep the README and other documentation up to date

## License

By contributing to this project, you agree that your contributions will be licensed under the project's [Apache 2.0 License](LICENSE).

Every file containing source code must include the SPDX license identifier and copyright notice:

```go
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
```

## Communication

- Use GitHub Issues for bug reports and feature requests
- Use GitHub Discussions for questions and general discussion
- Be respectful and considerate in all communications

Thank you for contributing to the Globus Go SDK!