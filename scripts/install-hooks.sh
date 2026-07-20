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

GOBIN="$(go env GOPATH)/bin"
REPO_ROOT="$(git rev-parse --show-toplevel)"

# This repo hosts two independent Go modules: the v3 module at the repo root
# and the v4 module under v4/. Each needs its own fmt/vet/test pass, since a
# root-level "./..." only sees the module it is run from.
MODULES=("$REPO_ROOT" "$REPO_ROOT/v4")

for MOD in "${MODULES[@]}"; do
    if [ ! -f "$MOD/go.mod" ]; then
        continue
    fi
    echo "--- Checking module: $MOD ---"
    cd "$MOD" || { echo "Error: cannot cd to $MOD"; EXIT_STATUS=1; continue; }

    # Run go fmt
    echo "Running go fmt..."
    go fmt ./...
    if [ $? -ne 0 ]; then
        echo "Error: go fmt failed in $MOD!"
        EXIT_STATUS=1
    fi

    # Run staticcheck if installed
    if [ -x "$GOBIN/staticcheck" ]; then
        echo "Running staticcheck..."
        "$GOBIN/staticcheck" ./... || echo "Warning: staticcheck found issues"
    elif command -v staticcheck &> /dev/null; then
        echo "Running staticcheck..."
        staticcheck ./... || echo "Warning: staticcheck found issues"
    else
        echo "staticcheck not found. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"
    fi

    # Run go vet
    echo "Running go vet..."
    go vet ./...
    if [ $? -ne 0 ]; then
        echo "Error: go vet failed in $MOD!"
        EXIT_STATUS=1
    fi

    # Run unit tests (short mode)
    echo "Running unit tests (short mode)..."
    go test ./pkg/... -short
    if [ $? -ne 0 ]; then
        echo "Error: Unit tests failed in $MOD!"
        EXIT_STATUS=1
    fi
done

cd "$REPO_ROOT" || true

if [ $EXIT_STATUS -eq 0 ]; then
    echo "All pre-commit checks passed!"
fi

exit $EXIT_STATUS
EOF

chmod +x "$REPO_ROOT/.git/hooks/pre-commit"

echo "Git hooks installed successfully!"
