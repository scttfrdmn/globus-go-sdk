# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors

# Testing Goals for Globus Go SDK v0.2.0 Release

_Last Updated: April 28, 2025_

This document outlines the testing goals and progress for the Globus Go SDK v0.2.0 release. It serves as a checklist to ensure thorough testing coverage before the final release.

## Overall Goals

- [ ] **80% Code Coverage**: Aim for at least 80% code coverage across the codebase
- [ ] **All Services Tested**: Ensure all services have both unit and integration tests
- [ ] **Error Paths Tested**: Verify that error paths are tested, not just happy paths
- [ ] **Cross-platform Testing**: Test on Linux, macOS, and Windows
- [ ] **API Export Verification**: Verify all required APIs are properly exported
- [ ] **Security Testing**: Complete security scan with no critical or high issues

## Unit Testing Progress

| Package | Coverage % | Status | Issues to Address |
|---------|------------|--------|-------------------|
| `pkg/core` | 78% | In Progress | Circuit breaker tests incomplete |
| `pkg/core/auth` | 85% | Complete | None |
| `pkg/core/authorizers` | 90% | Complete | None |
| `pkg/core/ratelimit` | 82% | Complete | None |
| `pkg/core/http` | 85% | Complete | Added API verification tests |
| `pkg/core/transport` | 75% | In Progress | Need tests for connection pooling |
| `pkg/services/auth` | 88% | In Progress | MFA challenge tests failing |
| `pkg/services/transfer` | 80% | In Progress | Memory optimization tests need fixes |
| `pkg/services/search` | 85% | Complete | None |
| `pkg/services/flows` | 72% | In Progress | Batch tests failing |
| `pkg/services/groups` | 80% | Complete | None |
| `pkg/services/compute` | 75% | In Progress | Missing coverage for error paths |
| `pkg/verify_credentials` | 0% | Not Started | Need to add tests |

## Integration Testing Progress

| Service | Basic Tests | Advanced Features | Issues to Address |
|---------|-------------|-------------------|-------------------|
| Auth | ✅ Complete | ✅ Complete | Skip MFA tests for automated runs |
| Transfer | ✅ Complete | ⚠️ Partial | Need to fix import cycles |
| Search | ✅ Complete | ⚠️ Partial | Advanced query tests incomplete |
| Flows | ✅ Complete | ❌ Failing | Fix batch errors |
| Groups | ✅ Complete | ⚠️ Partial | Need tests for member role changes |
| Compute | ✅ Complete | ⚠️ Partial | Need tests for error conditions |

## Feature-specific Testing

| Feature | Unit Tests | Integration Tests | Status | Issues |
|---------|------------|-------------------|--------|--------|
| **Authentication** |  |  |  |  |
| Client Credentials Flow | ✅ | ✅ | Complete | None |
| Authorization Code Flow | ✅ | ⚠️ | Partial | Requires manual testing |
| Token Refresh | ✅ | ✅ | Complete | None |
| MFA Support | ✅ | ❌ | Failing | MFA mocking issues |
| Token Storage | ✅ | ✅ | Complete | None |
| **Transfer** |  |  |  |  |
| Basic File Transfer | ✅ | ✅ | Complete | None |
| Recursive Transfer | ✅ | ✅ | Complete | None |
| Resumable Transfer | ✅ | ⚠️ | Partial | Need long-running tests |
| Memory Optimization | ✅ | ⚠️ | Partial | Import cycles |
| **Rate Limiting** |  |  |  |  |
| Backoff Strategy | ✅ | ✅ | Complete | None |
| Circuit Breaker | ✅ | ⚠️ | Partial | Need failure testing |
| Response Handler | ✅ | ✅ | Complete | None |
| **API Export Verification** |  |  |  |  |
| HTTP Pool API | ✅ | ✅ | Complete | None |
| Connection Pools | ✅ | ✅ | Complete | None |
| Package Exports Tool | ✅ | ✅ | Complete | None |
| Dependent Projects | ⚠️ | ⚠️ | Partial | Need more test cases |
| **Error Handling** |  |  |  |  |
| Error Types | ✅ | ✅ | Complete | None |
| Error Recovery | ✅ | ⚠️ | Partial | Some edge cases untested |
| API Error Mapping | ✅ | ⚠️ | Partial | Auth error test failing |

## Credential Verification Tool Testing

| Component | Unit Tests | Manual Testing | Status | Issues |
|-----------|------------|----------------|--------|--------|
| SDK Implementation | ❌ | ✅ | Partial | Need unit tests |
| Standalone Implementation | ❌ | ✅ | Partial | Need unit tests |
| Auth Service Testing | ✅ | ✅ | Complete | None |
| Transfer Service Testing | ✅ | ✅ | Complete | None |
| Groups Service Testing | ✅ | ✅ | Complete | None |
| Search Service Testing | ✅ | ✅ | Complete | None |

## Before Release Checklist

- [x] Fix integration test environment variable loading
- [x] Fix auth integration tests to use current interfaces
- [x] Fix import cycles in the transfer package
- [x] Add API export verification tests and tools
- [ ] Fix failing tests in the Flows package
- [ ] Add tests for verify-credentials tool
- [ ] Improve test coverage for core transport package
- [ ] Run full integration test suite with real credentials
- [ ] Verify that all examples work with current API
- [ ] Update test documentation with latest practices

## Testing Priorities

1. **Critical Path**: Focus on the fundamental authentication and transfer functionality
2. **API Compatibility**: Ensure compatibility with the current Globus API
3. **API Export Verification**: Verify that all required APIs are properly exported
4. **Reliability**: Test error handling and recovery mechanisms
5. **Performance**: Validate rate limiting and optimization features
6. **Security**: Verify proper handling of tokens and sensitive information

## Automated vs. Manual Testing

Some features require manual testing due to interactive authentication flows or long-running operations:

| Feature | Automated Testing | Manual Testing Required | Notes |
|---------|------------------|-------------------------|-------|
| Authorization Code Flow | Partial | Yes | Requires interactive browser login |
| MFA Challenge | Partial | Yes | Requires interactive MFA response |
| Resumable Transfers | Partial | Yes | Test interruption and resumption |
| Web Application Example | No | Yes | Requires manual browser interaction |
| Performance Benchmarks | Yes | No | Automated but needs review |

## Next Steps

1. Fix the failing tests in the Flows package
2. Add unit tests for the verify-credentials tool
3. Complete the integration tests for all services
4. Run a full test suite with real credentials
5. Address any remaining issues before release

## Resources

- [Testing Guide](testing.md): Comprehensive testing documentation
- [Integration Testing Setup](../INTEGRATION_TESTING.md): Instructions for setting up integration tests
- [Verify Credentials Tool](../../cmd/verify-credentials/README.md): Documentation for the verification tool