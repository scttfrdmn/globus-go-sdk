<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->
# Next Tasks for Globus Go SDK

## API Stability Implementation Progress

Based on the API Stability Implementation Plan, we are now working on the following tasks:

### Phase 1: Foundation (✅ Completed)

| Task | Description | Status |
|------|-------------|--------|
| Package stability indicators | Add stability annotations to all packages | ✅ Completed |
| Release checklist | Create and implement standardized release process | ✅ Completed |
| CHANGELOG enhancement | Restructure to track API changes more explicitly | ✅ Completed |
| CLAUDE.md update | Add API stability guidance for AI assistance | ✅ Completed |
| Code coverage targets | Define per-package coverage requirements | ✅ Completed |

#### Package Stability Indicators Progress
- ✅ Created doc.go file for pkg (root package)
- ✅ Created doc.go file for pkg/core with BETA stability
- ✅ Created doc.go file for pkg/services/auth with STABLE stability
- ✅ Created doc.go file for pkg/services/tokens with BETA stability
- ✅ Created doc.go file for pkg/services/transfer with MIXED stability
- ✅ Created doc.go file for pkg/services/groups with STABLE stability
- ✅ Created doc.go file for pkg/services/compute with BETA stability
- ✅ Created doc.go file for pkg/services/flows with BETA stability
- ✅ Created doc.go file for pkg/services/search with BETA stability
- ✅ Created doc.go file for pkg/services/timers with BETA stability

### Phase 2: Tools & Infrastructure (In Progress)

| Task | Description | Status |
|------|-------------|--------|
| API compatibility tool | Create tool to verify API compatibility between versions | ✅ Completed |
| Deprecation system | Implement runtime deprecation warnings | ✅ Completed |
| Contract testing | Add formal contract tests for core interfaces | ✅ Completed |
| CI coverage integration | Add coverage tracking to CI pipeline | Not Started |
| Compatibility testing | Set up tests to verify backward compatibility | Not Started |

#### API Compatibility Tool Progress
- ✅ Implemented `cmd/apigen` tool to extract API signatures from Go code
- ✅ Implemented `cmd/apicompare` tool to compare API signatures between versions
- ✅ Created comprehensive documentation in `API_STABILITY_PHASE2_SUMMARY.md`

#### Deprecation System Progress
- ✅ Created `pkg/core/deprecation` package for runtime deprecation warnings
- ✅ Implemented configurable warning system with logging integration
- ✅ Added `cmd/depreport` tool to generate reports of deprecated features
- ✅ Created example implementation in `pkg/core/deprecated_example.go`
- ✅ Documented the deprecation system in `API_DEPRECATION_SYSTEM.md`

#### Contract Testing Progress
- ✅ Created `pkg/core/contracts` package for contract tests
- ✅ Implemented contract tests for all core interfaces (`ClientInterface`, `Transport`, etc.)
- ✅ Added mock implementations for use in testing
- ✅ Created examples showing how to use contract tests
- ✅ Documented contract testing system in `CONTRACT_TESTING.md`

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

1. ✅ Complete Phase 1 of the API Stability Implementation Plan
2. ✅ Prepare for the 0.9.16 release with stability indicators
3. ✅ Begin Phase 2 with API compatibility tools and deprecation system
4. ⏳ Continue Phase 2 with contract testing and CI integration
5. ⏳ Continue addressing technical debt in parallel
6. ⏳ Prepare for the 0.9.17 release with full API compatibility verification

### Next Actions

1. Integrate API stability tools into CI/CD pipeline:
   - Create GitHub Actions workflow for API comparison
   - Implement automated API verification between releases
   - Add contract testing to CI pipeline

2. Implement code coverage tracking:
   - Set up coverage reporting in CI
   - Track coverage trends over time
   - Set coverage targets for different packages

3. Prepare documentation updates:
   - Update CONTRIBUTING.md with deprecation guidelines
   - Add contract testing guidance for contributors
   - Create comprehensive user guide for API stability tools

This approach allows us to make significant progress on API stability while continuing to address technical debt, with a clear focus on improving the user experience and ensuring backward compatibility.