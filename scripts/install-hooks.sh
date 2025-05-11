#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
# Script to install git hooks

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

echo "Installing git hooks..."

# Check if .git/hooks directory exists
if [ ! -d "$REPO_ROOT/.git/hooks" ]; then
    echo "Error: .git/hooks directory not found!"
    exit 1
fi

# Install pre-commit hook
cp "$REPO_ROOT/.git/hooks/pre-commit" "$REPO_ROOT/.git/hooks/pre-commit.backup" 2>/dev/null || true
cp "$REPO_ROOT/.git/hooks/pre-commit" "$REPO_ROOT/.git/hooks/pre-commit.$(date +%Y%m%d%H%M%S).backup" 2>/dev/null || true

cat > "$REPO_ROOT/.git/hooks/pre-commit" << 'EOF'
#!/bin/bash
# Pre-commit hook to run essential checks before committing code

echo "Running pre-commit checks..."

# Store the exit status
EXIT_STATUS=0

# Run license header checks
echo "Checking license headers..."
./scripts/check-license-headers.sh
if [ $? -ne 0 ]; then
    echo "Error: License header check failed!"
    EXIT_STATUS=1
fi

# Run go fmt
echo "Running go fmt..."
go fmt ./...
if [ $? -ne 0 ]; then
    echo "Error: go fmt failed!"
    EXIT_STATUS=1
fi

# Run staticcheck if installed
GOBIN="$(go env GOPATH)/bin"
if [ -x "$GOBIN/staticcheck" ]; then
    echo "Running staticcheck..."
    "$GOBIN/staticcheck" ./...
    if [ $? -ne 0 ]; then
        echo "Warning: staticcheck found issues"
        # Don't fail the commit for linting issues
    fi
elif command -v staticcheck &> /dev/null; then
    echo "Running staticcheck..."
    staticcheck ./...
    if [ $? -ne 0 ]; then
        echo "Warning: staticcheck found issues"
        # Don't fail the commit for linting issues
    fi
else
    echo "staticcheck not found. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"
fi

# Run go vet
echo "Running go vet..."
go vet ./...
if [ $? -ne 0 ]; then
    echo "Error: go vet failed!"
    EXIT_STATUS=1
fi

# Run unit tests (short mode)
echo "Running unit tests (short mode)..."
go test ./pkg/... -short
if [ $? -ne 0 ]; then
    echo "Error: Unit tests failed!"
    EXIT_STATUS=1
fi

if [ $EXIT_STATUS -eq 0 ]; then
    echo "All pre-commit checks passed!"
fi

exit $EXIT_STATUS
EOF

chmod +x "$REPO_ROOT/.git/hooks/pre-commit"

echo "Git hooks installed successfully!"
