#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
# Script to install git pre-push hook

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(git rev-parse --show-toplevel)"

echo "Installing git pre-push hook..."

# Check if .git/hooks directory exists
if [ ! -d "$REPO_ROOT/.git/hooks" ]; then
    echo "Error: .git/hooks directory not found!"
    exit 1
fi

# Install pre-push hook
cp "$REPO_ROOT/.git/hooks/pre-push" "$REPO_ROOT/.git/hooks/pre-push.backup" 2>/dev/null || true
cp "$REPO_ROOT/.git/hooks/pre-push" "$REPO_ROOT/.git/hooks/pre-push.$(date +%Y%m%d%H%M%S).backup" 2>/dev/null || true

cat > "$REPO_ROOT/.git/hooks/pre-push" << 'EOF'
#!/bin/bash
# Pre-push hook to run more comprehensive checks before pushing code

echo "Running pre-push checks..."

# Store the exit status
EXIT_STATUS=0

# Run all tests (including integration tests if possible)
echo "Running all tests..."
go test ./pkg/... 
if [ $? -ne 0 ]; then
    echo "Error: Tests failed!"
    echo "Do you want to continue pushing anyway? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo "Push aborted."
        exit 1
    fi
fi

# Check that documentation is up to date
if [ -d "./docs-site" ]; then
    echo "Checking documentation..."
    # Add documentation checks here if needed
fi

# Run security scan if available
if [ -f "./scripts/run_security_scan.sh" ]; then
    echo "Running security scan..."
    ./scripts/run_security_scan.sh
    if [ $? -ne 0 ]; then
        echo "Warning: Security scan found issues"
        echo "Do you want to continue pushing anyway? (y/N)"
        read -r response
        if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
            echo "Push aborted."
            exit 1
        fi
    fi
fi

echo "All pre-push checks completed!"

exit 0
EOF

chmod +x "$REPO_ROOT/.git/hooks/pre-push"

echo "Git pre-push hook installed successfully!"