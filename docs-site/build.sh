#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

set -e

VERSION="latest"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -v|--version)
      VERSION="$2"
      shift # past argument
      shift # past value
      ;;
    *)
      # Unknown option
      echo "Unknown option: $1"
      echo "Usage: $0 [-v|--version VERSION]"
      echo "  -v, --version VERSION   Specify the documentation version to build (e.g., v0.9.12, latest)"
      exit 1
      ;;
  esac
done

echo "Building documentation site for version: $VERSION"

# Make sure Hugo is installed
if ! command -v hugo &> /dev/null; then
    echo "Hugo is required but not installed. Please install it first:"
    echo "  brew install hugo (macOS)"
    echo "  apt-get install hugo (Ubuntu/Debian)"
    exit 1
fi

# Make sure book theme is present
if [ ! -d "themes/hugo-book" ]; then
    echo "Installing Hugo Book theme..."
    git clone https://github.com/alex-shpak/hugo-book.git themes/hugo-book
fi

# Run the sync-docs script to update documentation from the source
if [ -f "./sync-docs.sh" ]; then
    echo "Syncing documentation from source..."
    bash ./sync-docs.sh
else
    echo "Warning: sync-docs.sh not found. Documentation may be out of date."
fi

# Update config for the specified version
CONFIG_FILE="config/_default/config.toml"
if [ -f "$CONFIG_FILE" ]; then
    echo "Configuring for version: $VERSION"
    
    # Reset all selected flags to false
    sed -i.bak 's/selected = true/selected = false/g' "$CONFIG_FILE"
    
    # Set the appropriate version as selected
    if [ "$VERSION" = "latest" ]; then
        sed -i.bak 's/{ version = "main", path = "\/developer-tools\/go-sdk\/latest" }/{ version = "main", path = "\/developer-tools\/go-sdk\/latest", selected = true }/g' "$CONFIG_FILE"
    else
        sed -i.bak "s/{ version = \"$VERSION\", path = \"\/developer-tools\/go-sdk\/$VERSION\" }/{ version = \"$VERSION\", path = \"\/developer-tools\/go-sdk\/$VERSION\", selected = true }/g" "$CONFIG_FILE"
    fi
    
    # Clean up backup file
    rm -f "${CONFIG_FILE}.bak"
fi

# Generate API documentation if gomarkdoc is installed
if command -v gomarkdoc &> /dev/null; then
    echo "Generating API documentation..."
    mkdir -p content/docs/reference/api
    gomarkdoc --output content/docs/reference/api/{{.Dir}}.md ../pkg/...
else
    echo "gomarkdoc not installed. Skipping API documentation generation."
    echo "Install with: go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest"
fi

# Build the site
if [ "$VERSION" = "latest" ]; then
    # For latest version, use the default baseURL
    hugo --minify
else
    # For specific versions, customize the baseURL to include the version
    hugo --minify --baseURL "https://scttfrdmn.github.io/globus-go-sdk/$VERSION/"
fi

echo "Documentation built successfully in public/ directory"
echo "Run 'hugo serve' to view locally"