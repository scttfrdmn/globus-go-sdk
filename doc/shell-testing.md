<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Shell Script Testing Guide

This guide explains how the shell script testing infrastructure works in the Globus Go SDK and how to use it for developing and testing shell scripts.

## Overview

The Globus Go SDK uses two main tools for ensuring shell script quality:

1. **ShellCheck**: A static analysis tool for shell scripts that provides warnings and suggestions for bash/sh shell scripts
2. **BATS (Bash Automated Testing System)**: A testing framework for Bash that provides a TAP-compliant testing experience

Together, these tools provide comprehensive quality assurance for shell scripts in the project.

## Setting Up

### Installing ShellCheck

ShellCheck is used to lint shell scripts. To install it:

- **macOS**: `brew install shellcheck`
- **Ubuntu/Debian**: `apt-get install shellcheck`
- **Other platforms**: See [ShellCheck's installation guide](https://github.com/koalaman/shellcheck#installing)

### Installing BATS

The project includes a script to install BATS and its dependencies:

```sh
# Install BATS
./scripts/install_bats.sh
```

This will install:
- `bats-core`: The main BATS testing framework
- `bats-support`: Helper functions for better output
- `bats-assert`: Assertion functions for easier testing
- `bats-file`: File-related assertion functions

## Using ShellCheck

### Running ShellCheck

The project includes a script to run ShellCheck on all shell scripts:

```sh
# Lint all shell scripts
./scripts/lint_shell_scripts.sh

# Or use the Makefile target
make lint-shell
```

### ShellCheck Configuration

ShellCheck is configured via `.shellcheckrc` in the project root. This file contains:

- Disabled checks that aren't relevant to our project
- Project-specific settings

## Using BATS

### Writing BATS Tests

BATS tests are written in `.bats` files located in the `tests/` directory. Here's a basic example:

```bash
#!/usr/bin/env bats
# Load helper libraries
load "bats/bats-support/load.bash"
load "bats/bats-assert/load.bash"
load "bats/bats-file/load.bash"

# Test function
@test "Example test" {
  run echo "Hello, world!"
  assert_success
  assert_output "Hello, world!"
}
```

### Running BATS Tests

The project includes a script to run all BATS tests:

```sh
# Run all BATS tests
./scripts/run_shell_tests.sh

# Or use the Makefile target
make test-shell
```

To run specific tests:

```sh
# Run a specific test file
./scripts/run_shell_tests.sh tests/test_specific_script.bats
```

## Best Practices

### Shell Script Development

1. **Always run ShellCheck**: Before committing any shell script changes, run ShellCheck to ensure quality
2. **Write tests**: Add BATS tests for any new shell script functionality
3. **Keep scripts modular**: Make functions testable by keeping them small and focused
4. **Use `set -e`**: Include `set -e` in scripts to fail fast on errors
5. **Document scripts**: Add comments explaining the purpose and usage of scripts and functions

### Testing Techniques

1. **Setup and teardown**: Use `setup()` and `teardown()` functions for test preparation and cleanup
2. **Mocking**: Create mock functions to simulate commands and isolate unit tests
3. **Temporary directories**: Use temporary directories for test file operations
4. **Helper functions**: Extract common test logic into helper functions
5. **Assertions**: Use the assertion libraries (bats-assert, bats-file) for clearer tests

## Continuous Integration

The project has a GitHub Actions workflow that runs on all shell script changes:

1. **shellcheck**: Runs ShellCheck on all shell scripts
2. **shell-tests**: Runs BATS tests for all shell scripts

This ensures that all shell scripts committed to the repository meet quality standards.

## Example: Testing a Script

Here's an example of testing a script that counts files in a directory:

```bash
# Script: scripts/count_files.sh
#!/bin/bash
# Count files in a directory
count_files() {
  local dir="$1"
  find "$dir" -type f | wc -l
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  count_files "$@"
fi
```

```bash
# Test: tests/test_count_files.bats
#!/usr/bin/env bats
load "bats/bats-support/load.bash"
load "bats/bats-assert/load.bash"

# Source the script to test
source "../scripts/count_files.sh"

setup() {
  # Create a temporary directory
  TEST_DIR="$(mktemp -d)"
  
  # Create some files
  touch "$TEST_DIR/file1.txt"
  touch "$TEST_DIR/file2.txt"
}

teardown() {
  # Clean up
  rm -rf "$TEST_DIR"
}

@test "count_files counts files correctly" {
  # Run the function
  run count_files "$TEST_DIR"
  
  # Assert the output
  assert_output "2"
}
```

## Resources

- [ShellCheck Documentation](https://github.com/koalaman/shellcheck)
- [BATS Documentation](https://github.com/bats-core/bats-core)
- [BATS Assert Library](https://github.com/bats-core/bats-assert)
- [BATS File Library](https://github.com/bats-core/bats-file)
- [Shell Script Best Practices](https://kvz.io/bash-best-practices.html)