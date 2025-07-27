#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
# Simple script to summarize test results

echo "# Globus Go SDK Test Summary"
echo "Generated: $(date)"
echo

echo "## Transfer Service Tests"
echo
echo "### Unit Tests"
echo "```"
go test ./pkg/services/transfer | grep -E "ok|FAIL"
echo "```"
echo
echo "### Integration Tests"
echo "```"
go test -tags=integration ./pkg/services/transfer | grep -E "ok|FAIL"
echo "```"
echo
echo "#### Integration Tests Details"
echo "```"
go test -tags=integration ./pkg/services/transfer -v | grep -E "--- PASS:|--- FAIL:|--- SKIP:"
echo "```"
echo
echo "## Auth Service Tests"
echo
echo "### Unit Tests"
echo "```"
go test ./pkg/services/auth | grep -E "ok|FAIL"
echo "```"
echo
echo "### Integration Tests"
echo "```"
go test -tags=integration ./pkg/services/auth | grep -E "ok|FAIL"
echo "```"
echo
echo "#### Integration Tests Details"
echo "```"
go test -tags=integration ./pkg/services/auth -v | grep -E "--- PASS:|--- FAIL:|--- SKIP:"
echo "```"