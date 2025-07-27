<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK Versioning Strategy

## Overview

Starting with release v3.60.0, the Globus Go SDK adopts a **synchronized versioning strategy** with the official Globus Python SDK while maintaining its own evolution tracking.

## Versioning Scheme

The Go SDK uses a **hybrid versioning format** that tracks both the Python SDK compatibility and Go SDK evolution:

```
[PYTHON_SDK_VERSION]-[GO_SDK_BUILD]
```

### Examples

- `v3.60.0-1` - First Go SDK release synchronized with Python SDK v3.60.0
- `v3.60.0-2` - Second Go SDK release (patch/enhancement) still compatible with Python SDK v3.60.0
- `v3.61.0-1` - First Go SDK release synchronized with Python SDK v3.61.0

### Legacy Format (Pre-v3.60.0)

Prior releases used independent semantic versioning:
- `v0.9.17` - Last independent release
- `v0.9.18` - (Never released, transitioned to synchronized versioning)

## Rationale

### Benefits of Synchronized Versioning

1. **Clear Python SDK Compatibility** - Users immediately know which Python SDK features are supported
2. **Ecosystem Alignment** - Easier migration between Python and Go implementations
3. **Documentation Consistency** - Can reference Python SDK documentation with confidence
4. **API Parity Tracking** - Clear visibility into feature gaps or differences

### Go SDK Build Metadata

The build number after the hyphen tracks Go-specific evolution:
- **Bug fixes** in Go-specific code
- **Performance improvements** 
- **Go-specific features** (better type safety, Go idioms, etc.)
- **Documentation updates**
- **Test improvements**

## Release Process

### Major Python SDK Updates

When a new Python SDK version is released:

1. **Feature Analysis** - Compare Python SDK changes with Go SDK capabilities
2. **Gap Assessment** - Identify missing features or API differences  
3. **Implementation** - Add missing functionality to achieve parity
4. **Testing** - Verify compatibility and feature parity
5. **Release** - Tag as `vX.Y.Z-1` where X.Y.Z matches Python SDK version

### Go SDK Maintenance Releases

For Go-specific improvements without Python SDK changes:

1. **Increment Build Number** - e.g., `v3.60.0-1` → `v3.60.0-2`
2. **Maintain Python Compatibility** - No breaking changes to API parity
3. **Test Thoroughly** - Ensure continued Python SDK compatibility

## Version Selection Guidelines

### For New Projects

- Use the latest `vX.Y.Z-N` version for maximum features and fixes
- Check Python SDK documentation for feature references

### For Existing Projects

- **Patch updates** (`-1` → `-2`) are generally safe
- **Minor Python SDK updates** (`v3.60.0-x` → `v3.61.0-1`) may require testing
- **Major Python SDK updates** (`v3.x.x` → `v4.0.0`) require careful migration

## API Stability Commitment

### Python SDK Compatibility

- **STABLE** APIs match Python SDK stable APIs
- **Changes** only occur when Python SDK changes
- **Breaking changes** follow Python SDK breaking change schedule

### Go-Specific Enhancements

- Go SDK build increments (`-1` → `-2`) maintain API compatibility
- New Go-specific features use **BETA** or **ALPHA** stability levels
- Deprecation follows standard Go practices with migration periods

## Migration from Legacy Versioning

### Transition Timeline

- **v0.9.17** - Last independent release (December 2024)
- **v3.60.0-1** - First synchronized release (January 2025)

### Breaking Changes

The transition from v0.9.x to v3.60.0-x includes:

1. **Module Path** - Updated to `github.com/scttfrdmn/globus-go-sdk/v3`
2. **API Changes** - Aligned with Python SDK v3.60.0 patterns
3. **Error Handling** - Standardized error types and patterns
4. **Client Initialization** - Unified client creation patterns

### Migration Guide

See [MIGRATION_v3.md](MIGRATION_v3.md) for detailed upgrade instructions.

## Future Considerations

### Semantic Versioning Compliance

The synchronized versioning maintains semantic versioning principles:

- **Major version** (3.x.x) tracks Python SDK major version
- **Minor version** (x.60.x) tracks Python SDK minor version  
- **Patch version** (x.x.0) tracks Python SDK patch version
- **Build metadata** (-1, -2) tracks Go SDK increments

### Long-term Strategy

This approach provides:
- **Predictable upgrades** aligned with Python SDK releases
- **Clear compatibility** guarantees across language implementations
- **Flexible Go-specific** improvements without disrupting sync
- **Future-proof** versioning that can adapt to Python SDK changes

## References

- [Semantic Versioning 2.0.0](https://semver.org/)
- [Globus Python SDK Releases](https://github.com/globus/globus-sdk-python/releases)
- [Go Modules Version Numbers](https://go.dev/ref/mod#versions)