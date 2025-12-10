<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Python SDK v4.2.0 Tracking

**Python SDK Release Date:** December 3, 2025
**Go SDK Current Version:** v4.1.0-2 (Python SDK v4.1.0 parity)
**Target Go SDK Version:** v4.2.0-1 (Python SDK v4.2.0 parity)

## Python SDK v4.2.0 Changes

### Python Version Support (Not Applicable to Go)
- ✅ Added Python 3.14 support
- ✅ Dropped Python 3.8 support

*Note: These changes are Python-specific and do not require Go SDK changes.*

### New Features Requiring Implementation

#### 1. Context Manager Support (Close() Method)
**Priority:** High
**Effort:** Medium

**Python Implementation:**
```python
with GlobusApp("myapp") as app:
    client = app.get_transfer_client()
    # Resources automatically cleaned up on exit
```

**Go SDK Implementation Required:**
- Add `Close()` method to `core.Client`
- Add `Close()` method to all service clients (auth, groups, transfer, search, flows, timers, compute)
- Implement proper HTTP transport cleanup
- Add defer-based usage examples
- Update documentation with Go resource management patterns

**Example Go Pattern:**
```go
client, err := transfer.NewClient(ctx, config)
if err != nil {
    return err
}
defer client.Close()
```

#### 2. TimersClient.AddAppFlowUserScope
**Priority:** Medium
**Effort:** Low

**Python Implementation:**
```python
timers_client = app.get_timers_client()
timers_client.add_app_flow_user_scope(flow_id)
```

**Go SDK Implementation Required:**
- Add `AddAppFlowUserScope(ctx context.Context, flowID string) error` method to `v4/pkg/services/timers/client.go`
- Method should register required scope dependencies for flow timers
- Add integration with GlobusApp/Config scope management
- Add unit tests and documentation

#### 3. GARE Auto-Retry
**Priority:** Medium
**Effort:** High

**Python Feature:**
- Automatic Globus Auth Requirements Error (GARE) redriving
- Disabled by default
- Enabled via `auto_retry_gares=True` configuration

**Go SDK Implementation Required:**
- Add GARE detection in error handling (`core/errors.go`)
- Implement automatic consent flow retry logic
- Add `AutoRetryGAREs` boolean to `core.Config`
- Implement consent URL handling and token refresh
- Add comprehensive testing for GARE scenarios
- Document GARE handling patterns for Go developers

#### 4. Resource Leak Fix
**Priority:** High
**Effort:** Low

**Python Issue Fixed:**
- `GlobusApp` created internal client objects without closing transports

**Go SDK Status:**
- ✅ Review current v4 implementation for similar issues
- ✅ Ensure all HTTP clients and transports are properly closed
- ✅ Add resource cleanup tests
- ⚠️  **Current v4 implementation needs audit** - minimal HTTP client management exists

## Implementation Status

### v3 SDK Status
- **Current Version:** v3.65.0-1
- **Python SDK Parity:** ✅ v3.65.0 (latest)
- **Status:** Production-ready and fully synchronized

### v4 SDK Status
- **Current Version:** v4.1.0-2
- **Python SDK Parity:** ⚠️ v4.1.0 (behind by one minor version)
- **Implementation Completeness:** ⚠️ Minimal/Prototype

**v4 Current State:**
- Skeleton implementation with basic CRUD operations
- Missing advanced features like connection pooling, rate limiting, comprehensive error handling
- No tests for v4 code
- Minimal documentation

## Recommendations

### Short Term (Immediate)
1. ✅ Update upstream tracking to v4.2.0
2. 📝 Create GitHub issue tracking v4.2.0 implementation
3. 📝 Document v4.2.0 requirements (this document)

### Medium Term (Next Sprint)
1. **For v3 SDK:**
   - Continue monitoring Python SDK v3.x for any new releases
   - v3 is stable and production-ready

2. **For v4 SDK:**
   - Complete v4 core infrastructure before adding v4.2.0 features:
     - Comprehensive HTTP client with connection pooling
     - Rate limiting and circuit breaker patterns
     - Full error handling and retry logic
     - Token storage and refresh mechanisms
     - Comprehensive test coverage

3. **Then implement v4.2.0 features:**
   - Add Close() methods with proper resource cleanup
   - Implement AddAppFlowUserScope for TimersClient
   - Add GARE auto-retry with configuration options
   - Comprehensive testing and documentation

### Long Term (Future Releases)
1. Achieve full v4 SDK feature parity with v3 SDK
2. Migrate v3 production users to v4 with clear migration guides
3. Establish automated Python SDK tracking and synchronization

## Risk Assessment

### v4.2.0 Implementation Risks
- **High Risk:** Implementing v4.2.0 features on incomplete v4 infrastructure
- **Medium Risk:** Resource management patterns may differ significantly between Python and Go
- **Low Risk:** v4.2.0 changes are backward-compatible (minor version bump)

### Mitigation Strategies
1. **Complete v4 core first:** Don't add v4.2.0 features until v4 infrastructure is solid
2. **Learn from v3:** Use proven patterns from v3 SDK implementation
3. **Incremental approach:** Implement Close() methods first (highest priority, lowest risk)
4. **Testing focus:** Comprehensive tests for resource cleanup and GARE handling

## Timeline Estimate

### Minimal v4.2.0 Implementation (Close() only)
- Effort: 2-3 days
- Includes: Basic Close() methods, simple resource cleanup, basic tests

### Complete v4.2.0 Implementation
- Effort: 1-2 weeks
- Includes: All features, comprehensive tests, full documentation, GARE handling

### Full v4 Infrastructure + v4.2.0
- Effort: 4-6 weeks
- Includes: Complete v4 redesign, all v4.2.0 features, production-ready quality

## Next Steps

1. **User Decision Required:**
   - Option A: Implement minimal v4.2.0 (Close() methods only) on current v4 skeleton
   - Option B: Complete v4 infrastructure first, then add v4.2.0 features
   - Option C: Focus on v3 stability, defer v4.2.0 until v4 infrastructure is ready

2. **Recommended Approach (Option C):**
   - v3 SDK is production-ready and current
   - v4 SDK needs foundational work before adding new features
   - Create tracking issue for v4.2.0 implementation
   - Revisit when v4 infrastructure is complete

## References

- [Python SDK v4.2.0 Release](https://github.com/globus/globus-sdk-python/releases/tag/4.2.0)
- [Python SDK v4.1.0 Release](https://github.com/globus/globus-sdk-python/releases/tag/4.1.0)
- [Go SDK v4 Directory](./v4/)
- [Go SDK v3 (Production)](./pkg/)

---

**Document Status:** Initial draft (Dec 9, 2025)
**Last Updated:** December 9, 2025
**Owner:** Go SDK Maintainers
