# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Running Tests via Git Hooks

This document provides instructions for Claude and other developers on how to run tests via the Git hooks in the Globus Go SDK.

## Available Git Hooks

The repository includes two primary Git hooks:

1. **Pre-commit Hook**: Runs before each commit to validate code quality and basic tests
2. **Pre-push Hook**: Runs before pushing to validate more comprehensive tests

## Setting Up Git Hooks

### Installing All Hooks

To install all available Git hooks, run:

```bash
./scripts/install-all-hooks.sh
```

### Installing Specific Hooks

To install only specific hooks:

```bash
# Install only pre-commit hook
./scripts/install-hooks.sh

# Install only pre-push hook
./scripts/install-pre-push-hook.sh
```

## What the Hooks Run

### Pre-commit Hook

The pre-commit hook runs the following checks:

1. License header verification via `./scripts/check-license-headers.sh`
2. Code formatting with `go fmt ./...`
3. Code linting with `staticcheck ./...` (if installed)
4. Static analysis with `go vet ./...`
5. Unit tests in short mode with `go test ./pkg/... -short`

If any of these checks fail, the commit will be blocked.

### Pre-push Hook

The pre-push hook runs more comprehensive checks:

1. All tests (including integration tests) with `go test ./pkg/...`
2. Documentation checks (if applicable)
3. Security scan (if available) via `./scripts/run_security_scan.sh`

If these checks fail, you'll be prompted whether to continue with the push or abort.

## Bypassing Hooks

In certain situations, you may need to bypass the hooks, for example, when committing work-in-progress code that doesn't pass all checks yet. To bypass hooks, add the `--no-verify` flag to your git command:

```bash
git commit --no-verify -m "WIP: ..."
git push --no-verify
```

**Note:** Bypassing hooks should be used sparingly. The hooks are designed to maintain code quality and prevent issues.

## Troubleshooting

### Hook Not Running

If hooks don't run when expected:

1. Ensure the hook files are executable: `chmod +x .git/hooks/pre-commit .git/hooks/pre-push`
2. Check if Git is configured to use hooks: `git config core.hooksPath`
3. Reinstall the hooks: `./scripts/install-all-hooks.sh`

### Hook Errors

Common errors:

- **License header errors**: Run `./scripts/standardize-spdx-headers.sh` to fix
- **Formatting issues**: Run `go fmt ./...` to fix
- **Linting issues**: Address issues reported by staticcheck
- **Test failures**: Examine the detailed error output and fix failing tests

## Running Tests Manually

You can also run the same tests manually without Git hooks:

```bash
# Run checks similar to pre-commit hook
./scripts/check-license-headers.sh
go fmt ./...
go vet ./...
go test ./pkg/... -short

# Run checks similar to pre-push hook
go test ./pkg/...
./scripts/run_security_scan.sh
```

## Integration with CI/CD

These Git hooks complement the CI/CD pipelines by catching issues earlier in the development process, before code reaches GitHub. The same tests will run in GitHub Actions workflows when you push your changes.