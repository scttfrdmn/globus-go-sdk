#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# This script updates license headers in documentation files to follow the standard format

# Loop through all .md files in the doc directory
find /Users/scttfrdmn/src/globus-go-sdk/doc -name "*.md" | while read -r file; do
  # Check if the file contains the old license header format
  if grep -q "<!-- SPDX-License-Identifier: Apache-2.0 -->" "$file"; then
    echo "Updating license header in $file"
    # Use sed to replace the old license header with the new one
    sed -i '' 's/<!-- SPDX-License-Identifier: Apache-2.0 -->\n<!-- .*Contributors -->/# SPDX-License-Identifier: Apache-2.0\n# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors/g' "$file"
  fi
done

echo "License header update completed"