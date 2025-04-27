<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Security Tooling

This document provides information about the security tools integrated into the Globus Go SDK development workflow, how to use them, and best practices for maintaining a secure codebase.

## Overview

The Globus Go SDK uses several security tools to detect and prevent security issues:

1. **gosec** - Static analysis security tool for Go code
2. **nancy** - Dependency vulnerability scanner
3. **gitleaks** - Secret scanner for preventing sensitive data in the repository
4. **shellcheck** - Static analysis tool for shell scripts

These tools are integrated at multiple levels in the development workflow:

- **Pre-commit hooks** - Run automatically before commits
- **CI/CD pipelines** - Run on pull requests and pushes to main
- **Manual scanning** - Available via Makefile targets and command-line tool
- **Scheduled scans** - Run weekly to check for new vulnerabilities

## Using the Security Tools

### Pre-commit Hooks

Security tools are integrated as pre-commit hooks to check your code before committing:

1. Install pre-commit:
   ```bash
   pip install pre-commit
   pre-commit install
   ```

2. The hooks will automatically run when you commit. To run manually:
   ```bash
   pre-commit run --all-files
   ```

### Makefile Targets

The Makefile includes security scan targets:

```bash
# Run all security scans
make security-check

# Run specific scans
make gosec
make nancy
make gitleaks
```

### Security Test Command-Line Tool

The SDK includes a dedicated security test tool (`cmd/security-test/main.go`):

```bash
# Build the tool
go build -o security-test ./cmd/security-test

# Run a self-test
./security-test -self

# Scan dependencies
./security-test -deps

# Analyze a token for security issues
./security-test -token "your_token" -client-id "your_client_id"
```

### GitHub Actions

The following GitHub Actions workflows run security scans:

1. **go.yml** - Includes gosec and nancy scans as part of the main CI pipeline
2. **security-scan.yml** - Comprehensive security scan with gosec, nancy, and gitleaks

You can manually trigger a security scan:
1. Go to the "Actions" tab in the GitHub repository
2. Select "Security Scan" workflow
3. Click "Run workflow"

## Understanding Results

### gosec Results

gosec reports are output in JSON and SARIF formats:

- **Severity**: Issues are rated as LOW, MEDIUM, or HIGH
- **Confidence**: How confident gosec is about the finding (LOW, MEDIUM, HIGH)
- **CWE**: Common Weakness Enumeration ID for the issue

Example issue:
```json
{
  "severity": "HIGH",
  "confidence": "HIGH",
  "cwe": {
    "id": "327",
    "url": "https://cwe.mitre.org/data/definitions/327.html"
  },
  "rule_id": "G402",
  "details": "TLS InsecureSkipVerify set true.",
  "file": "/path/to/file.go",
  "line": "42",
  "code": "tls.Config{InsecureSkipVerify: true}"
}
```

### nancy Results

nancy output includes:

- **Package**: The affected dependency
- **Version**: The vulnerable version
- **CVE**: Common Vulnerabilities and Exposures ID
- **Severity**: How severe the vulnerability is
- **Description**: Details about the vulnerability

### gitleaks Results

gitleaks identifies potential secrets in the codebase:

- **Description**: The type of secret found
- **File**: The file containing the potential secret
- **Line**: The line number where the secret was found
- **Secret**: A masked version of the identified secret

## Best Practices

1. **Run Pre-commit Hooks**: Let pre-commit hooks run before committing code

2. **Review Security Scan Results**: Always review and address security issues

3. **Understand False Positives**: Some security findings may be false positives - document these cases

4. **Do Not Disable Security Checks**: Avoid disabling security checks without good reason

5. **Update Dependencies**: Regularly update dependencies to fix security issues

6. **Secure API Token Handling**:
   - Never hardcode tokens or credentials
   - Use environment variables for sensitive configuration
   - Consider token rotation for long-lived services

7. **Code Review**: Always review security-sensitive code with extra care

## Common Security Issues

### 1. Insecure TLS Configuration

```go
// INSECURE
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true, // Don't do this!
        },
    },
}

// SECURE
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
    },
}
```

### 2. Hardcoded Credentials

```go
// INSECURE
const apiKey = "1234567890abcdef" // Don't do this!

// SECURE
apiKey := os.Getenv("API_KEY")
if apiKey == "" {
    return errors.New("API_KEY environment variable is required")
}
```

### 3. SQL Injection

```go
// INSECURE
query := fmt.Sprintf("SELECT * FROM users WHERE name = '%s'", name) // Don't do this!

// SECURE
query := "SELECT * FROM users WHERE name = ?"
rows, err := db.Query(query, name)
```

### 4. Command Injection

```go
// INSECURE
exec.Command("bash", "-c", "ls "+userInput) // Don't do this!

// SECURE
exec.Command("ls", userInput)
```

## Contact

For questions about security practices or to report security issues:

- Open an issue with the "security" label
- Email the security team (see SECURITY_GUIDELINES.md for contact details)