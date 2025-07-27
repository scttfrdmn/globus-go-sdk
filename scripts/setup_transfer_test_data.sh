#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
#
# This script sets up test data for the Globus Go SDK transfer integration tests.
# It creates sample files in the specified directory for testing transfers.

set -e

TEST_DIR="/Users/scttfrdmn/globus-test"
TIMESTAMP=$(date +%Y%m%d%H%M%S)

echo "Setting up transfer test data in $TEST_DIR..."

# Create test directory if it doesn't exist
mkdir -p "$TEST_DIR"
mkdir -p "$TEST_DIR/nested"
mkdir -p "$TEST_DIR/nested/subfolder"

# Create a small text file
echo "This is a small test file for Globus transfer tests. Created at $TIMESTAMP" > "$TEST_DIR/small_file.txt"

# Create a medium text file
dd if=/dev/urandom bs=1024 count=100 | base64 > "$TEST_DIR/medium_file.dat"

# Create a file in the nested directory
echo "This is a file in a nested directory. Created at $TIMESTAMP" > "$TEST_DIR/nested/nested_file.txt"

# Create a file in the nested subfolder
echo "This is a file in a nested subfolder. Created at $TIMESTAMP" > "$TEST_DIR/nested/subfolder/deep_file.txt"

# Create sample JSON data
cat > "$TEST_DIR/sample.json" << EOF
{
  "name": "Globus Test Data",
  "created": "$TIMESTAMP",
  "system": "terror",
  "system_id": "20b46e7f-230d-11f0-9913-0affeb91e4e5",
  "files": [
    {"name": "small_file.txt", "type": "text"},
    {"name": "medium_file.dat", "type": "binary"},
    {"name": "nested/nested_file.txt", "type": "text"},
    {"name": "nested/subfolder/deep_file.txt", "type": "text"}
  ]
}
EOF

# Set permissions
chmod -R 755 "$TEST_DIR"

echo "Test data setup complete."
echo "Created files:"
find "$TEST_DIR" -type f | sort

echo ""
echo "You can now run the transfer integration tests with:"
echo "go test -tags=integration ./pkg/services/transfer/..."