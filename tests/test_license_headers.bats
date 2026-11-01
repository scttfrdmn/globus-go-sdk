#!/usr/bin/env bats
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

# Tests for scripts/check-license-headers.sh. The script takes no arguments and
# scans the current directory: .go and .sh files without an SPDX line are
# errors (exit 1); .md files only warn.

load "bats/bats-support/load.bash"
load "bats/bats-assert/load.bash"

CHECK_SCRIPT="${BATS_TEST_DIRNAME}/../scripts/check-license-headers.sh"
SPDX="// SPDX-License-Identifier: Apache-2.0"

setup() {
    TEST_DIR="$(mktemp -d)"
    mkdir -p "$TEST_DIR/pkg/core" "$TEST_DIR/scripts"
    cd "$TEST_DIR"
}

teardown() {
    rm -rf "$TEST_DIR"
}

@test "passes when every Go and shell file has an SPDX header" {
    printf '%s\npackage core\n' "$SPDX" > pkg/core/ok.go
    printf '#!/bin/bash\n# SPDX-License-Identifier: Apache-2.0\n' > scripts/ok.sh
    run bash "$CHECK_SCRIPT"
    assert_success
    assert_output --partial "All files have proper SPDX license headers"
}

@test "fails and names a Go file missing its header" {
    printf '%s\npackage core\n' "$SPDX" > pkg/core/ok.go
    printf 'package core\n' > pkg/core/missing.go
    run bash "$CHECK_SCRIPT"
    assert_failure
    assert_output --partial "Missing SPDX header: ./pkg/core/missing.go"
    refute_output --partial "ok.go"
    assert_output --partial "Found 1 files missing SPDX license headers"
}

@test "fails and names a shell script missing its header" {
    printf '#!/bin/bash\necho hi\n' > scripts/missing.sh
    run bash "$CHECK_SCRIPT"
    assert_failure
    assert_output --partial "Missing SPDX header: ./scripts/missing.sh"
}

@test "counts every missing file" {
    printf 'package core\n' > pkg/core/a.go
    printf 'package core\n' > pkg/core/b.go
    printf '#!/bin/bash\n' > scripts/c.sh
    run bash "$CHECK_SCRIPT"
    assert_failure
    assert_output --partial "Found 3 files missing SPDX license headers"
}

@test "markdown without a header only warns" {
    printf '%s\npackage core\n' "$SPDX" > pkg/core/ok.go
    printf '# Title\n' > README.md
    run bash "$CHECK_SCRIPT"
    assert_success
    assert_output --partial "Missing SPDX header: ./README.md"
}

@test "vendored Go files are ignored" {
    mkdir -p vendor/x
    printf 'package x\n' > vendor/x/x.go
    run bash "$CHECK_SCRIPT"
    assert_success
}
