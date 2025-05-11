# API Deprecation System

This document describes the deprecation system implemented as part of Phase 2 of the API Stability Implementation Plan for the Globus Go SDK.

## Overview

The deprecation system provides a standardized way to:

1. Mark features as deprecated
2. Log warnings when deprecated features are used
3. Document when features were deprecated and when they will be removed
4. Track all deprecated features in the codebase
5. Generate reports for release planning

## Components

The deprecation system consists of several components:

### 1. Runtime Deprecation Warnings

The `pkg/core/deprecation` package provides functions for logging deprecation warnings when deprecated code is used:

```go
// Log a deprecation warning
deprecation.LogWarning(
    logger,
    "DoSomething",
    "v0.9.0",
    "v1.0.0",
    "Use DoSomethingBetter instead.",
)
```

Key features:
- Configurable warning behavior (can be disabled or set to warn only once per feature)
- Integration with the SDK's logging system
- Consistent formatting of deprecation messages
- Support for tracking which features have been deprecated and when they will be removed

### 2. Deprecation Report Tool

The `cmd/depreport` tool scans the codebase for deprecated features and generates a report:

```bash
go run cmd/depreport/main.go -o DEPRECATED_FEATURES.md
```

The tool identifies deprecated features in two ways:
1. Functions, types, variables, and constants with a `// Deprecated: ...` comment
2. Code that calls the `deprecation.LogWarning` or `deprecation.LogFeatureWarning` functions

The generated report includes:
- All deprecated features grouped by planned removal version
- File locations and line numbers
- Deprecation and removal versions
- Migration guidance
- Summary statistics

### 3. Documentation Guidelines

The following guidelines should be followed when deprecating features:

#### When Deprecating a Public API

1. Add a doc comment with the `Deprecated:` prefix that includes:
   - When the feature was deprecated (version)
   - When the feature will be removed (version)
   - What to use instead (migration path)

2. Add a runtime warning using the `deprecation.LogWarning` function at the beginning of the implementation.

3. Update the package stability level if necessary.

4. Update the CHANGELOG.md file to document the deprecation.

#### Example

```go
// DoSomething does something.
// 
// Deprecated: This function was deprecated in v0.9.0 and will be removed in v1.0.0.
// Use DoSomethingBetter instead.
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

## Deprecation Lifecycle

Features go through the following lifecycle:

1. **Active**: The feature is fully supported.

2. **Deprecated**: 
   - The feature is marked as deprecated in doc comments
   - Runtime warnings are issued when the feature is used
   - The feature is listed in the deprecation report
   - The feature continues to work as expected

3. **Removed**:
   - The feature is removed in the version specified in the deprecation notice
   - Typically, features should be deprecated for at least one minor version before removal
   - Removal should align with semantic versioning (removals are breaking changes and require a major version bump)

## Integration with Release Process

The deprecation system integrates with the release process as follows:

1. Before a release, run the deprecation report tool to identify features scheduled for removal in the upcoming version.

2. For minor and patch releases, verify that no deprecated features are being removed prematurely.

3. For major releases, identify all features that were scheduled for removal and ensure they are properly removed.

4. Update the deprecation report after each release to track the current state of deprecated features.

## Future Enhancements

Planned enhancements to the deprecation system include:

1. Integration with the API comparison tool to ensure that removing deprecated features is properly flagged as a breaking change.

2. Automated verification during CI that features aren't removed before their planned removal version.

3. Enhanced documentation generation that automatically includes deprecation notices in the GoDoc and website documentation.

4. Support for marking entire packages as deprecated.

## Conclusion

The deprecation system provides a comprehensive approach to managing the lifecycle of API features, ensuring a smooth transition for users when features need to be changed or removed. By following a consistent deprecation process, we can maintain API stability while still evolving the SDK to meet new requirements.