# Changelog Template

This template ensures we follow "Keep a Changelog" practices consistently.

## Release Process

### 1. Before Making Changes
- Add entries to the `[Unreleased]` section as you make changes
- Use present tense, imperative mood ("Add feature" not "Added feature")
- Reference issues and pull requests where applicable
- Group similar changes together

### 2. When Creating a Release
- Move content from `[Unreleased]` to a new version section
- Add the release date in YYYY-MM-DD format
- Update the comparison links at the bottom
- Follow semantic versioning (major.minor.patch)

### 3. Section Guidelines

#### Added
For new features and capabilities:
- New API endpoints
- New client methods
- New configuration options
- New examples or documentation

#### Changed
For changes in existing functionality:
- API modifications that maintain compatibility
- Improved performance or behavior
- Documentation updates
- **Use "BREAKING:" prefix for breaking changes**

#### Deprecated
For soon-to-be removed features:
- Mark features that will be removed in future versions
- Always include when they will be removed
- Provide migration guidance

#### Removed
For removed features:
- Removed API endpoints
- Removed configuration options
- Removed deprecated features

#### Fixed
For bug fixes:
- Bug fixes
- Security patches
- Error handling improvements

#### Security
For security-related changes:
- Vulnerability fixes
- Security improvements
- Authentication enhancements

## Template for New Release

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New feature description

### Changed
- **BREAKING**: Breaking change description
- Non-breaking change description

### Deprecated
- Deprecated feature (will be removed in vN.0.0)

### Removed
- Removed feature description

### Fixed
- Bug fix description

### Security
- Security improvement description
```

## Version Synchronization with Python SDK

Since we're now synchronizing with the Python SDK:

1. **Check Python SDK version**: Always check the current Python SDK version before releasing
2. **Match major.minor**: Our releases should match Python SDK major.minor versions
3. **Independent patches**: Patch versions can be independent for Go-specific fixes
4. **Document alignment**: Always mention Python SDK version alignment in changelog

## Example Entries

### Good Examples:
```markdown
### Added
- **Transfer Service**: Added support for guest collections (#123)
- **Auth Service**: Added MFA support for WebAuthn devices
- **Documentation**: Added comprehensive API reference for all services

### Changed
- **BREAKING**: Unified error handling across all services - all services now return `GlobusError` type
- **Performance**: Improved connection pooling performance by 40%
- Updated examples to use new client initialization patterns

### Fixed
- Fixed memory leak in recursive transfer operations (#456)
- Fixed race condition in token refresh mechanism
- Corrected API endpoint URLs for Timers service
```

### Bad Examples (avoid these):
```markdown
### Added
- Stuff was added
- Various improvements
- Fixed things

### Changed
- Made changes to the code
- Updated various files
- Some breaking changes
```

## Markdown Formatting Guidelines

- Use `**Service Name**:` to prefix service-specific changes
- Use `**BREAKING**:` to clearly mark breaking changes
- Use backticks for code, API names, and technical terms
- Use numbered issue/PR references: `(#123)`
- Use present tense: "Add" not "Added"
- Be specific and descriptive