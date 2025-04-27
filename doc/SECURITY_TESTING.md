<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Security Testing Guide

This document provides information about security testing practices for the Globus Go SDK, including how to set up, run, and interpret different types of security tests.

## Overview

The Globus Go SDK implements several layers of security testing:

1. **Static Analysis** - Code scanning to find potential security issues
2. **Dependency Scanning** - Checking for vulnerabilities in dependencies
3. **Secret Detection** - Preventing accidental commit of secrets
4. **Token Analysis** - Analyzing tokens for security best practices
5. **Integration Testing** - Testing with real credentials and services

## Setting Up Security Testing Environment

### Prerequisites

- Go 1.18 or later
- Pre-commit (for pre-commit hooks)
- Docker (optional, for containerized tests)

### Installing Security Tools

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Install nancy
go install github.com/sonatype-nexus-community/nancy@latest

# Install gitleaks
go install github.com/zricethezav/gitleaks/v8@latest

# Install shellcheck
# macOS
brew install shellcheck
# Ubuntu
apt-get install shellcheck
```

### Configuring Pre-commit Hooks

```bash
# Install pre-commit
pip install pre-commit

# Install hooks
pre-commit install
```

## Running Security Tests

### Manual Testing

#### Static Analysis with gosec

```bash
# Run gosec on the entire codebase
gosec ./...

# Run gosec with JSON output for integration with other tools
gosec -fmt=json -out=gosec-results.json ./...

# Focus on high-severity issues only
gosec -severity=high ./...
```

#### Dependency Scanning with nancy

```bash
# Scan all dependencies
go list -json -m all | nancy sleuth

# Exclude development dependencies
go list -json -m all | nancy sleuth --exclude-dev

# Output as JSON
go list -json -m all | nancy sleuth --output json > nancy-results.json
```

#### Secret Detection with gitleaks

```bash
# Scan current directory
gitleaks detect

# Scan with custom configuration
gitleaks detect --config gitleaks.toml

# Output as JSON
gitleaks detect --report-format json --report-path gitleaks-report.json
```

#### Command-line Security Test Tool

The SDK provides a dedicated security test tool:

```bash
# Build the tool
go build -o security-test ./cmd/security-test

# Run security self-test
./security-test -self

# Analyze token for security issues
./security-test -token "your_token" -client-id "your_client_id"
```

### Automated Testing

#### GitHub Actions Workflows

The repository includes several GitHub Actions workflows for security testing:

1. **go.yml** - Includes gosec and nancy in main CI pipeline
2. **security-scan.yml** - Dedicated workflow for comprehensive security scanning
3. **shell-lint.yml** - Workflow for shell script linting with shellcheck

#### Makefile Targets

```bash
# Run all security checks
make security-check

# Run specific checks
make gosec
make nancy
make gitleaks
```

## Interpreting Results

### gosec Results

gosec categorizes findings by severity and rule ID:

- **G101** - Hardcoded credentials
- **G102** - Binding to all network interfaces
- **G103** - Unsafe use of unsafe.Pointer
- **G104** - Unhandled errors
- **G107** - URL provided to HTTP request as taint input
- **G108** - Profiling endpoint automatically exposed on /debug/pprof
- **G109** - Potential Integer overflow
- **G110** - Potential DoS vulnerability via decompression bomb
- **G201** - SQL query construction using string concatenation
- **G202** - SQL query construction using string format
- **G203** - Use of unescaped data in HTML templates
- **G204** - Subprocess launched with function call as argument or cmd arguments
- **G301** - Poor file permissions used when creating a directory
- **G302** - Poor file permissions used when creating a file
- **G303** - Creating tempfile using a predictable path
- **G304** - File path provided as taint input
- **G305** - File traversal when extracting zip archive
- **G306** - Poor file permissions used when writing to a file
- **G307** - Deferring a method which returns an error
- **G401** - Crypto weak block size
- **G402** - TLS InsecureSkipVerify set true
- **G403** - RSA keys should be at least 2048 bits
- **G404** - Weak random number generator (math/rand instead of crypto/rand)
- **G501** - Import blocklist: crypto/md5
- **G502** - Import blocklist: crypto/sha1
- **G503** - Import blocklist: crypto/sha256
- **G504** - Import blocklist: crypto/sha512
- **G505** - Import blocklist: crypto/des
- **G601** - Implicit memory aliasing in for loop

### nancy Results

nancy provides information about vulnerabilities in dependencies:

- **CVE ID** - The Common Vulnerabilities and Exposures identifier
- **CVSS Score** - The Common Vulnerability Scoring System score (0-10)
- **Affected Package** - The affected dependency
- **Vulnerable Versions** - The affected versions
- **Fixed Version** - The version where the vulnerability is fixed

### gitleaks Results

gitleaks detects potential secrets in the codebase:

- **Rule** - The rule that triggered the detection
- **Secret** - A masked version of the detected secret
- **File** - The file where the secret was found
- **Line** - The line number where the secret was found

## Addressing Security Issues

### Priority Levels

1. **Critical** - Must be fixed immediately
   - Authentication/authorization bypasses
   - Remote code execution
   - Token leakage
   - High CVSS (9.0-10.0) vulnerabilities

2. **High** - Must be fixed in the next release
   - Information disclosure
   - Medium-High CVSS (7.0-8.9) vulnerabilities
   - Sensitive data exposure

3. **Medium** - Should be fixed in a timely manner
   - Low-Medium CVSS (4.0-6.9) vulnerabilities
   - Insecure configurations

4. **Low** - Fix when convenient
   - Low CVSS (0.1-3.9) vulnerabilities
   - Code quality issues

### Remediation Process

1. **Triage**: Assess the severity and impact
2. **Document**: Create an issue with details about the vulnerability
3. **Test**: Create a test case that reproduces the issue
4. **Fix**: Implement a fix
5. **Verify**: Ensure the fix resolves the issue
6. **Release**: Include the fix in the next appropriate release

### False Positives

If you identify a false positive:

1. Document the finding and why it's a false positive
2. Add an appropriate comment to the code:
   ```go
   // gosec:ignore:G404 Using math/rand is acceptable for non-cryptographic purposes
   ```
3. Configure the tool to exclude the false positive in future scans

## Best Practices

1. **Regular Scanning**: Run security scans regularly

2. **Dependency Updates**: Keep dependencies up-to-date

3. **Secure Coding Training**: Ensure developers understand security best practices

4. **Code Review**: Include security considerations in code reviews

5. **Security Testing**: Include security tests in the test suite

6. **Secret Management**: Use a secure method for managing secrets

7. **Documentation**: Document security decisions and configurations

## Resources

- [OWASP Go Security Cheatsheet](https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Go_Security_Cheatsheet.md)
- [gosec Documentation](https://github.com/securego/gosec)
- [nancy Documentation](https://github.com/sonatype-nexus-community/nancy)
- [gitleaks Documentation](https://github.com/zricethezav/gitleaks)
- [Go Security Best Practices](https://blog.sqreen.com/go-security-best-practices/)