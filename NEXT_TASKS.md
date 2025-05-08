<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Next Tasks for Globus Go SDK

## API Stability Implementation Progress

Based on the API Stability Implementation Plan, we are now working on the following tasks:

### Phase 1: Foundation (In Progress)

| Task | Description | Status |
|------|-------------|--------|
| Package stability indicators | Add stability annotations to all packages | In Progress |
| Release checklist | Create and implement standardized release process | Not Started |
| CHANGELOG enhancement | Restructure to track API changes more explicitly | Completed |
| CLAUDE.md update | Add API stability guidance for AI assistance | Completed |
| Code coverage targets | Define per-package coverage requirements | Not Started |

#### Package Stability Indicators Progress
- ✅ Created doc.go file for pkg (root package)
- ✅ Created doc.go file for pkg/core with BETA stability
- ✅ Created doc.go file for pkg/services/auth with STABLE stability
- ✅ Created doc.go file for pkg/services/tokens with BETA stability
- ✅ Created doc.go file for pkg/services/transfer with MIXED stability
- ✅ Created doc.go file for pkg/services/groups with STABLE stability
- ⏳ Create doc.go files for pkg/services/compute
- ⏳ Create doc.go files for pkg/services/flows
- ⏳ Create doc.go files for pkg/services/search
- ⏳ Create doc.go files for pkg/services/timers

### Phase 2: Tools & Infrastructure (Upcoming)

| Task | Description | Status |
|------|-------------|--------|
| API compatibility tool | Create tool to verify API compatibility between versions | Not Started |
| Deprecation system | Implement runtime deprecation warnings | Not Started |
| Contract testing | Add formal contract tests for core interfaces | Not Started |
| CI coverage integration | Add coverage tracking to CI pipeline | Not Started |
| Compatibility testing | Set up tests to verify backward compatibility | Not Started |

## Remaining Technical Debt

### 1. Fix Transfer Package Tests

There are several disabled test files that need to be fixed and re-enabled:

- ✅ `pkg/services/transfer/resumable_test.go.disabled` - Fixed
- ✅ `pkg/services/transfer/resumable_integration_test.go.disabled` - Fixed
- ✅ `pkg/services/transfer/streaming_iterator_test.go.disabled` - Fixed
- ✅ `pkg/services/transfer/memory_optimized_test.go.disabled` - Fixed

### 2. Fix Import Cycle Issues - ✅ Completed

The main import cycle issues have been resolved by:

1. **Core Interface Dependencies**
   - ✅ Created `pkg/core/interfaces` package with interface definitions
   - ✅ Added adapter implementations for interface verification
   - ✅ Updated code to use interfaces rather than concrete implementations

2. **Service Dependencies**
   - ✅ Updated auth service integration tests to use the new pattern
   - ✅ Fixed all import cycles by using proper interfaces and dependency inversion

### 3. Fix Duplication Issues

1. Resolve duplicate type definitions and method implementations:
   - DeleteItem in test_helpers.go vs models.go
   - Ensure consistent naming and field structure for all types

2. Fix test helpers to use consistent naming and avoid conflicts:
   - Prefix all test-specific types with "Test..." to avoid collision
   - Move test helpers to a separate testutils package if needed

### 4. Integration Testing - In Progress

1. ✅ Updated auth integration tests to work with the new interface pattern
2. ⏳ Update transfer integration tests to work with the new interface pattern
3. ⏳ Implement the credential verification checks in the integration test script
4. ⏳ Create proper environment setup for running integration tests

## Implementation Strategy

1. Complete Phase 1 of the API Stability Implementation Plan
2. Prepare for the 0.9.16 release with stability indicators
3. Begin Phase 2 with API compatibility tools
4. Continue addressing technical debt in parallel

This approach allows us to make progress on both API stability and technical debt, with a clear focus on improving the user experience and ensuring backward compatibility.