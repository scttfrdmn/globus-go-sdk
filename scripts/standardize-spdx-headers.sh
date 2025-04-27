#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025 Scott Friedman and Project Contributors

# This script standardizes all license headers in the codebase to the SPDX format

set -e  # Exit on error

# Print colorized output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Standardizing license headers to SPDX format...${NC}"

# Go source files
find . -name "*.go" -not -path "./vendor/*" | while read file; do
  # Skip if file already has SPDX header
  if grep -q "SPDX-License-Identifier: Apache-2.0" "$file"; then
    echo -e "${GREEN}✓ SPDX header already present in ${file}${NC}"
    continue
  fi

  # Check if file has existing copyright
  if grep -q "Copyright" "$file"; then
    echo -e "${YELLOW}⚠ Replacing existing copyright in ${file}${NC}"
    # Replace existing copyright line(s)
    sed -i.bak '/[Cc]opyright/d' "$file"
  else
    echo -e "${BLUE}Adding SPDX header to ${file}${NC}"
  fi

  # Get the first line (could be package statement or build tag)
  first_line=$(head -1 "$file")
  
  # Handle build tags specially
  if [[ "$first_line" == "//go:build"* ]]; then
    # Get second line too if it's a build constraint comment
    second_line=$(sed -n '2p' "$file")
    if [[ "$second_line" == "// +build"* ]]; then
      # Insert after the second line
      sed -i.bak "2a\\
// SPDX-License-Identifier: Apache-2.0\\
// Copyright (c) 2025 Scott Friedman and Project Contributors
" "$file"
    else
      # Insert after the first line
      sed -i.bak "1a\\
// SPDX-License-Identifier: Apache-2.0\\
// Copyright (c) 2025 Scott Friedman and Project Contributors
" "$file"
    fi
  else
    # Insert at the beginning of the file
    sed -i.bak "1i\\
// SPDX-License-Identifier: Apache-2.0\\
// Copyright (c) 2025 Scott Friedman and Project Contributors
" "$file"
  fi
  
  # Remove backup file
  rm "${file}.bak"
done

# Shell scripts
find . -name "*.sh" | while read file; do
  # Skip if file already has SPDX header
  if grep -q "SPDX-License-Identifier: Apache-2.0" "$file"; then
    echo -e "${GREEN}✓ SPDX header already present in ${file}${NC}"
    continue
  fi

  # Check if file has existing copyright
  if grep -q "Copyright" "$file"; then
    echo -e "${YELLOW}⚠ Replacing existing copyright in ${file}${NC}"
    # Replace existing copyright line(s)
    sed -i.bak '/[Cc]opyright/d' "$file"
  else
    echo -e "${BLUE}Adding SPDX header to ${file}${NC}"
  fi

  # Check for shebang line
  first_line=$(head -1 "$file")
  if [[ "$first_line" == "#!/"* ]]; then
    # Insert after shebang
    sed -i.bak "1a\\
# SPDX-License-Identifier: Apache-2.0\\
# Copyright (c) 2025 Scott Friedman and Project Contributors
" "$file"
  else
    # Insert at the beginning of the file
    sed -i.bak "1i\\
# SPDX-License-Identifier: Apache-2.0\\
# Copyright (c) 2025 Scott Friedman and Project Contributors
" "$file"
  fi
  
  # Remove backup file
  rm "${file}.bak"
done

# Markdown files
find . -name "*.md" | while read file; do
  # Skip if file already has SPDX header
  if grep -q "SPDX-License-Identifier: Apache-2.0" "$file"; then
    echo -e "${GREEN}✓ SPDX header already present in ${file}${NC}"
    continue
  fi

  # Check if file has existing copyright
  if grep -q "Copyright" "$file"; then
    echo -e "${YELLOW}⚠ Replacing existing copyright in ${file}${NC}"
    # Replace existing copyright line(s)
    sed -i.bak '/[Cc]opyright/d' "$file"
  else
    echo -e "${BLUE}Adding SPDX header to ${file}${NC}"
  fi

  # Insert at the beginning of the file
  sed -i.bak "1i\\
<!-- SPDX-License-Identifier: Apache-2.0 -->\\
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
" "$file"
  
  # Remove backup file
  rm "${file}.bak"
done

echo -e "${GREEN}✓ Standardized all license headers to SPDX format${NC}"