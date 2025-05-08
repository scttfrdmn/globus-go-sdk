<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Release Plan for v0.9.11

## Steps

1. Create and switch to release branch
   ```bash
   git checkout -b v0.9.11-release
   ```

2. Update version number in code (`pkg/core/version.go`)
   - Change version constant from "0.9.10" to "0.9.11"

3. Update CHANGELOG.md
   - Add v0.9.11 section with release date
   - Document all changes since v0.9.10 including:
     - Bug fixes
     - Improvements
     - New features
     - Documentation updates

4. Run tests to ensure everything works
   ```bash
   go test ./...
   go vet ./...
   ```

5. Check documentation for version references
   - Update any version-specific documentation
   - Ensure examples reference the correct version

6. Create tag and GitHub release
   - Tag: `git tag v0.9.11`
   - Push tag: `git push origin v0.9.11`
   - Create GitHub release with release notes from CHANGELOG.md

7. Merge release branch back to main after release
   ```bash
   git checkout main
   git merge v0.9.11-release
   git push
   ```

## Post-release tasks
1. Update version in main branch to next development version if applicable
2. Announce release to users