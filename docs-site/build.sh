#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

set -e

echo "Building documentation site..."

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

# Build the site
hugo --minify

echo "Documentation built successfully in public/ directory"
echo "Run 'hugo serve' to view locally"