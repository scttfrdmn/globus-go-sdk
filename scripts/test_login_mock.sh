#!/bin/bash
# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2025-2026 Scott Friedman and Project Contributors

# Offline smoke test of the globus-cli login plumbing against the local mock
# auth server. No network and no real Globus credentials are used. Verifies that:
#   1. `globus-cli login` completes against the mock and stores a token per
#      resource server (via the mock's other_tokens payload), and
#   2. `globus-cli token export-env` emits a GLOBUS_TEST_<SVC>_TOKEN line for
#      each of transfer/groups/search/flows.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

MOCK_ADDR=":8099"
MOCK_BASE="http://localhost:8099/v2/"
CALLBACK="http://localhost:8080/callback"

TMPHOME="$(mktemp -d)"
MOCK_BIN="$(mktemp -u)"
CLI_BIN="$(mktemp -u)"
MOCK_PID=""
LOGIN_PID=""

cleanup() {
	if [ -n "$MOCK_PID" ]; then
		kill "$MOCK_PID" 2>/dev/null || true
		wait "$MOCK_PID" 2>/dev/null || true
	fi
	if [ -n "$LOGIN_PID" ]; then
		kill "$LOGIN_PID" 2>/dev/null || true
		wait "$LOGIN_PID" 2>/dev/null || true
	fi
	rm -rf "$TMPHOME" "$MOCK_BIN" "$CLI_BIN"
}
trap cleanup EXIT

echo "Building mock-auth-server and globus-cli..."
go build -o "$MOCK_BIN" ./cmd/mock-auth-server
go build -o "$CLI_BIN" ./cmd/globus-cli

echo "Starting mock auth server on $MOCK_ADDR..."
"$MOCK_BIN" -addr "$MOCK_ADDR" >/dev/null 2>&1 &
MOCK_PID=$!
sleep 1

echo "Running login against the mock (isolated HOME)..."
login_log="$(mktemp)"
HOME="$TMPHOME" GLOBUS_AUTH_BASE_URL="$MOCK_BASE" "$CLI_BIN" login >"$login_log" 2>&1 &
LOGIN_PID=$!
sleep 2

# The CLI printed the authorize URL and is waiting on the callback. Emulate the
# browser by hitting the mock authorize endpoint, which 302s to the callback
# with a code, completing the exchange.
state="$(grep -o 'state=[^ &]*' "$login_log" | head -1 | cut -d= -f2)"
curl -sL "${MOCK_BASE}oauth2/authorize?redirect_uri=${CALLBACK}&state=${state}" >/dev/null
wait "$LOGIN_PID" 2>/dev/null || true
LOGIN_PID=""

if ! grep -q "Login successful" "$login_log"; then
	echo "FAIL: login did not complete:" >&2
	cat "$login_log" >&2
	rm -f "$login_log"
	exit 1
fi
rm -f "$login_log"

echo "Checking token export-env output..."
export_out="$(HOME="$TMPHOME" GLOBUS_AUTH_BASE_URL="$MOCK_BASE" "$CLI_BIN" token export-env)"
echo "$export_out"

fail=0
for var in GLOBUS_TEST_TRANSFER_TOKEN GLOBUS_TEST_GROUPS_TOKEN GLOBUS_TEST_SEARCH_TOKEN GLOBUS_TEST_FLOWS_TOKEN; do
	if ! echo "$export_out" | grep -q "export ${var}="; then
		echo "FAIL: missing ${var} in export-env output" >&2
		fail=1
	fi
done

if [ "$fail" -ne 0 ]; then
	exit 1
fi

echo "PASS: login stored per-resource-server tokens and export-env emitted all four."
