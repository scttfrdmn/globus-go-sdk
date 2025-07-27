#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
# Script to install all git hooks

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Installing all git hooks..."

# Install pre-commit hook
"$SCRIPT_DIR/install-hooks.sh"

# Install pre-push hook
"$SCRIPT_DIR/install-pre-push-hook.sh"

echo "All git hooks installed successfully!"
echo "The following hooks are now active:"
echo "- pre-commit: Runs format checks, linting, and unit tests"
echo "- pre-push: Runs all tests including integration tests"
echo ""
echo "To bypass hooks temporarily, use git commit/push with --no-verify"