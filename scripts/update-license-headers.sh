#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# Update license headers in Go files

set -e

LICENSE_HEADER="// SPDX-License-Identifier: Apache-2.0"
COPYRIGHT="// Copyright (c) 2025 Scott Friedman and Project Contributors"

# Find all Go files in the project
files=$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*")

# Function to add license header if missing
add_license_header() {
  file=$1
  
  # Check if file already has license header
  if grep -q "$LICENSE_HEADER" "$file"; then
    return 0
  fi
  
  # Check if file already has copyright notice
  has_copyright=0
  if grep -q "$COPYRIGHT" "$file"; then
    has_copyright=1
  fi
  
  echo "Adding license header to $file"
  
  # Create a temporary file
  tmp_file=$(mktemp)
  
  # Add license header
  echo "$LICENSE_HEADER" > "$tmp_file"
  
  # Add copyright notice if missing
  if [ $has_copyright -eq 0 ]; then
    echo "$COPYRIGHT" >> "$tmp_file"
  fi
  
  # Add blank line and then the original content
  echo "" >> "$tmp_file"
  cat "$file" >> "$tmp_file"
  
  # Replace the original file
  mv "$tmp_file" "$file"
}

# Function to add copyright notice if missing
add_copyright_notice() {
  file=$1
  
  # Check if file already has copyright notice
  if grep -q "$COPYRIGHT" "$file"; then
    return 0
  fi
  
  # Check if file has license header
  has_license=0
  if grep -q "$LICENSE_HEADER" "$file"; then
    has_license=1
  fi
  
  echo "Adding copyright notice to $file"
  
  # Create a temporary file
  tmp_file=$(mktemp)
  
  # Add license header if present
  if [ $has_license -eq 1 ]; then
    echo "$LICENSE_HEADER" > "$tmp_file"
    echo "$COPYRIGHT" >> "$tmp_file"
    
    # Get the rest of the file without the license header
    tail -n +2 "$file" >> "$tmp_file"
  else
    # Add copyright notice at the beginning
    echo "$COPYRIGHT" > "$tmp_file"
    cat "$file" >> "$tmp_file"
  fi
  
  # Replace the original file
  mv "$tmp_file" "$file"
}

# Update each file
for file in $files; do
  add_license_header "$file"
  add_copyright_notice "$file"
done

echo "✅ License headers updated in all Go files"