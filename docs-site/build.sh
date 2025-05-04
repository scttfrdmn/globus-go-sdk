#!/bin/bash
# Script to build the Hugo site for deployment

# Make sure we're in the docs-site directory
cd "$(dirname "$0")"

# First sync the documentation
./sync-docs.sh

# Then build the site with minification
echo "Building site..."
hugo --minify

echo "Build complete! The site is in the 'public' directory."