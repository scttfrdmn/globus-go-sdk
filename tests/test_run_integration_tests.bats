#!/usr/bin/env bats
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

# Tests for scripts/run_integration_tests.sh. A stub `go` on PATH records its
# arguments, so no real tests or Globus calls run.

load "bats/bats-support/load.bash"
load "bats/bats-assert/load.bash"

SCRIPT="${BATS_TEST_DIRNAME}/../scripts/run_integration_tests.sh"

setup() {
    TEST_DIR="$(mktemp -d)"
    mkdir -p "$TEST_DIR/bin" "$TEST_DIR/repo/v4"
    GO_LOG="$TEST_DIR/go.log"
    cat > "$TEST_DIR/bin/go" <<STUB
#!/bin/bash
echo "\$PWD|\$*" >> "$GO_LOG"
STUB
    chmod +x "$TEST_DIR/bin/go"
    export PATH="$TEST_DIR/bin:$PATH"
    unset GLOBUS_TEST_CLIENT_ID GLOBUS_TEST_CLIENT_SECRET SKIP_INTEGRATION
    cd "$TEST_DIR/repo"
}

teardown() {
    rm -rf "$TEST_DIR"
}

with_creds() {
    export GLOBUS_TEST_CLIENT_ID="cid" GLOBUS_TEST_CLIENT_SECRET="secret"
}

@test "SKIP_INTEGRATION=true exits 0 without running go" {
    SKIP_INTEGRATION=true run bash "$SCRIPT"
    assert_success
    assert_output --partial "Skipping integration tests"
    [ ! -e "$GO_LOG" ]
}

@test "fails when required credentials are missing" {
    run bash "$SCRIPT"
    assert_failure
    assert_output --partial "GLOBUS_TEST_CLIENT_ID is not set"
    assert_output --partial "GLOBUS_TEST_CLIENT_SECRET is not set"
    [ ! -e "$GO_LOG" ]
}

@test "loads credentials from .env.test" {
    printf '# comment\n\nGLOBUS_TEST_CLIENT_ID=from-file\nGLOBUS_TEST_CLIENT_SECRET=s=with=equals\n' > .env.test
    run bash "$SCRIPT"
    assert_success
    assert_output --partial "Loaded GLOBUS_TEST_CLIENT_ID"
    assert_output --partial "GLOBUS_TEST_CLIENT_SECRET is set"
}

@test "no arguments runs both modules uncached" {
    with_creds
    run bash "$SCRIPT"
    assert_success
    run cat "$GO_LOG"
    assert_line --index 0 "$TEST_DIR/repo|test -v -tags=integration -count=1 ./..."
    assert_line --index 1 "$TEST_DIR/repo/v4|test -v -tags=integration -count=1 ./..."
}

@test "one argument runs that package tree" {
    with_creds
    run bash "$SCRIPT" pkg/services/auth
    assert_success
    run cat "$GO_LOG"
    assert_output "$TEST_DIR/repo|test -v -tags=integration ./pkg/services/auth/..."
}

@test "two arguments add a -run pattern" {
    with_creds
    run bash "$SCRIPT" pkg/services/transfer Transfer
    assert_success
    run cat "$GO_LOG"
    assert_output "$TEST_DIR/repo|test -v -tags=integration ./pkg/services/transfer/... -run Transfer"
}

@test "three arguments print usage and fail" {
    with_creds
    run bash "$SCRIPT" a b c
    assert_failure
    assert_output --partial "Usage:"
}
