#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
# Script to sync reference documentation to the Hugo site

# Make sure we're in the docs-site directory
cd "$(dirname "$0")"

# Create directories if they don't exist
mkdir -p content/docs/reference/{auth,compute,flows,groups,search,timers,tokens,transfer}

# Copy reference documentation
echo "Syncing reference documentation..."
cp -r ../doc/reference/auth/*.md content/docs/reference/auth/
cp -r ../doc/reference/compute/*.md content/docs/reference/compute/
cp -r ../doc/reference/flows/*.md content/docs/reference/flows/
cp -r ../doc/reference/groups/*.md content/docs/reference/groups/
cp -r ../doc/reference/search/*.md content/docs/reference/search/
cp -r ../doc/reference/timers/*.md content/docs/reference/timers/
cp -r ../doc/reference/tokens/*.md content/docs/reference/tokens/
cp -r ../doc/reference/transfer/*.md content/docs/reference/transfer/

# Copy the Globus logo if it exists
if [ -f "../doc/images/globus-go-sdk-logo.png" ]; then
  mkdir -p static
  cp ../doc/images/globus-go-sdk-logo.png static/logo.png
  echo "Copied logo to static/logo.png"
fi

# Add front matter to files that don't have it
echo "Adding front matter to files..."
find content -name "*.md" -type f | while read -r file; do
  # Skip files that already have front matter
  if grep -q "^---" "$file"; then
    continue
  fi
  
  # Extract title from the first heading
  title=$(head -n1 "$file" | sed "s/^# //")
  
  # Create temporary file
  temp_file=$(mktemp)
  
  # Write front matter and original content to temp file
  echo "---" > "$temp_file"
  echo "title: \"$title\"" >> "$temp_file"
  echo "---" >> "$temp_file"
  cat "$file" >> "$temp_file"
  
  # Replace original file with temp file
  mv "$temp_file" "$file"
  
  echo "Added front matter to $file"
done

echo "Sync complete!"