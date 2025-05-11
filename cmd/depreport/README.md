# Deprecation Report Tool

This tool scans the Globus Go SDK codebase for deprecated features and generates a report. It looks for:

1. Functions, types, variables, and constants marked with a `// Deprecated: ...` comment
2. Code that calls the `deprecation.LogWarning` or `deprecation.LogFeatureWarning` functions

## Usage

```
go run cmd/depreport/main.go [flags]
```

### Flags

- `-dir string`: Source directory to scan (default ".")
- `-o string`: Output file for the report (defaults to stdout)

## Report Format

The generated report is in Markdown format and includes:

- A list of all deprecated features grouped by planned removal version
- Features without a planned removal version
- The file and line number where each deprecated feature is defined
- When the feature was deprecated
- Migration guidance (if available)
- A summary of the total number of deprecated features

## Example

To generate a report for the entire SDK and save it to a file:

```bash
go run cmd/depreport/main.go -dir . -o DEPRECATED_FEATURES.md
```

## Integration with Release Process

This tool should be run as part of the release process to:

1. Identify features that should be removed in the upcoming release
2. Ensure proper documentation for deprecated features
3. Track the deprecation lifecycle of the SDK

## How to Mark Code as Deprecated

There are two recommended ways to mark code as deprecated:

### 1. Using Doc Comments

Add a comment beginning with `// Deprecated:` to the function, type, variable, or constant:

```go
// DoSomething does something.
// 
// Deprecated: This function was deprecated in v0.9.0 and will be removed in v1.0.0.
// Use DoSomethingBetter instead.
func DoSomething() {
    // ...
}
```

### 2. Using the Deprecation Package

Use the `deprecation.LogWarning` function to log a warning when the deprecated code is used:

```go
func DoSomething() {
    deprecation.LogWarning(
        logger,
        "DoSomething",
        "v0.9.0",
        "v1.0.0",
        "Use DoSomethingBetter instead.",
    )
    
    // Function implementation...
}
```