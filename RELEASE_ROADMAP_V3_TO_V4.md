# Release Roadmap: v3.60.0 → v3.65.0 → v4.0.0

**Status:** Approved Plan
**Last Updated:** October 2, 2025
**Current Version:** v3.60.0

## Overview

This roadmap outlines the path to align with the upstream Globus Python SDK through version 3.65.0, and prepare for the upcoming v4.0.0 architecture.

## Key Findings

### "Incomplete Work" Reality Check ✅

After thorough analysis of the Python SDK:

| "Missing" Method | Reality | Action |
|-----------------|---------|--------|
| `RecursiveTransfer()` | ✅ **Already implemented** as `SubmitRecursiveTransfer()` | Fix tests to use existing method |
| `SubmitTextFileCreation()` | ❌ **Doesn't exist** in Python SDK | Remove test references |
| `GetActivationRequirements()` | 🚫 **Deprecated** in Python SDK v3.61.0 | Remove tests, document deprecation |
| `ActivateEndpoint()` | 🚫 **Deprecated** in Python SDK v3.61.0 | Remove tests, document deprecation |

**Conclusion:** Our SDK is well-implemented. The "incomplete work" is actually test artifacts that need cleanup.

## Release Timeline

```
October 2025   November 2025   December 2025   January 2026    February 2026
     │               │               │               │               │
     v               v               v               v               v
  v3.61.0        v3.65.0        v4.0.0-alpha.1   v4.0.0-beta.1   v4.0.0
 (Test Fixes)   (Features)     (New Arch)       (Testing)       (Stable)
```

## Release 1: v3.61.0 - Test Alignment & Cleanup

**Target Date:** October 15, 2025 (2 weeks)
**Focus:** Fix test artifacts, no new features

### Scope

#### Phase 1A: Transfer Test Fixes
- [ ] Fix `TestIntegration_RecursiveTransfer` to use `SubmitRecursiveTransfer()`
- [ ] Remove 6 references to `SubmitTextFileCreation()` in resumable tests
- [ ] Update tests to use proper file creation (mkdir + write)
- [ ] Verify all transfer tests pass

#### Phase 1B: Activation Method Cleanup
- [ ] Remove test references to `GetActivationRequirements()` (2 locations)
- [ ] Remove test references to `ActivateEndpoint()` (1 location)
- [ ] Remove test references to `ActivationProfile` field (1 location)
- [ ] Document auto-activation in test comments
- [ ] Update client.go:290-293 deprecation note

#### Phase 1C: CLI Decision & Implementation
- [ ] Decision: Implement, remove, or defer CLI transfer commands
- [ ] If implement: Connect to existing SDK methods
- [ ] If remove: Clean up CLI package structure
- [ ] If defer: Add "Coming in v4.0" notices

#### Phase 1D: Documentation Updates
- [ ] Update `INCOMPLETE_WORK.md` with findings
- [ ] Add `PYTHON_SDK_ALIGNMENT_PLAN.md` (already created)
- [ ] Update `CHANGELOG.md` with v3.61.0 changes
- [ ] Create `doc/DEPRECATED_METHODS.md`

### Success Criteria
- ✅ All integration tests pass without skips
- ✅ CI/CD pipeline fully green
- ✅ Zero references to non-existent methods
- ✅ Documentation reflects reality

### Risks
- **Low** - Only test changes, no API modifications
- **Medium** - CLI decision may require stakeholder input

## Release 2: v3.65.0 - Feature Parity

**Target Date:** November 15, 2025 (4 weeks after v3.61.0)
**Focus:** Add Python SDK v3.61-v3.65 features

### Scope

#### Phase 2A: Groups Service Enhancements

**New Methods:**
```go
// v3.65.0: GetMyGroups with status filtering
func (c *Client) GetMyGroups(ctx context.Context, options *GetMyGroupsOptions) (*GroupList, error)

// v3.65.0: Batch role changes
func (c *Client) ChangeRoles(ctx context.Context, operations []RoleChange) (*BatchResult, error)

// v3.62.0/v3.63.0: Subscription management
func (c *Client) SetSubscriptionAdminVerified(ctx context.Context, groupID, subscriptionID string) error
```

**Files to Create/Modify:**
- [ ] `pkg/services/groups/client.go` - Add GetMyGroups with statuses
- [ ] `pkg/services/groups/batch.go` - New file for batch operations
- [ ] `pkg/services/groups/models.go` - Add new option types
- [ ] `pkg/services/groups/subscription.go` - New file for subscription methods
- [ ] `pkg/services/groups/client_test.go` - Unit tests
- [ ] `pkg/services/groups/integration_test.go` - Integration tests

**Estimated Time:** 1 week

#### Phase 2B: Search Service Enhancement

**New Method:**
```go
// v3.64.0: Update index metadata
func (c *Client) UpdateIndex(ctx context.Context, indexID string, update *IndexUpdate) (*Index, error)
```

**Files to Create/Modify:**
- [ ] `pkg/services/search/client.go` - Add UpdateIndex
- [ ] `pkg/services/search/models.go` - Add IndexUpdate type
- [ ] `pkg/services/search/client_test.go` - Unit tests
- [ ] `pkg/services/search/integration_test.go` - Integration tests

**Estimated Time:** 2 days

#### Phase 2C: Flows Service Enhancement

**New Type:**
```go
// v3.65.0: FlowTimer payload class
type FlowTimer struct {
    Name        string
    FlowID      string
    FlowInput   map[string]interface{}
    Schedule    TimerSchedule
    CallbackURL string
}
```

**Files to Create/Modify:**
- [ ] `pkg/services/flows/timer.go` - New file for FlowTimer
- [ ] `pkg/services/flows/models.go` - Update with timer types
- [ ] `pkg/services/flows/client.go` - Add timer helper methods
- [ ] `pkg/services/flows/timer_test.go` - Unit tests
- [ ] `examples/flow-timer/main.go` - Example usage

**Estimated Time:** 3 days

#### Phase 2D: Deprecation Markers

**Add to Transfer Service:**
```go
// Deprecated: SymlinkTarget is deprecated as of v3.64.0
SymlinkTarget string `json:"symlink_target,omitempty"`

// Deprecated: RecursiveSymlinks is deprecated as of v3.64.0
RecursiveSymlinks bool `json:"recursive_symlinks,omitempty"`

// Deprecated: SkipActivationCheck is deprecated as of v3.64.0
SkipActivationCheck bool `json:"skip_activation_check,omitempty"`
```

**Add to Compute Service:**
```go
// Deprecated: ComputeClient alias is deprecated as of v3.61.0.
// Use ComputeClientV2 or ComputeClientV3 explicitly.
type ComputeClient = Client
```

**Files to Modify:**
- [ ] `pkg/services/transfer/models.go` - Add deprecation markers
- [ ] `pkg/services/compute/client.go` - Add alias deprecation
- [ ] `doc/DEPRECATED_METHODS.md` - Document all deprecations

**Estimated Time:** 1 day

#### Phase 2E: Documentation & Testing

- [ ] Update `README.md` with v3.65.0 features
- [ ] Update `CHANGELOG.md` with complete feature list
- [ ] Create migration guide from v3.60.0 to v3.65.0
- [ ] Update API reference docs
- [ ] Add examples for all new features
- [ ] Comprehensive integration testing
- [ ] Update GoDoc documentation

**Estimated Time:** 1 week

### Success Criteria
- ✅ All Python SDK v3.61-v3.65 features implemented
- ✅ Complete test coverage for new features
- ✅ Documentation reflects all changes
- ✅ Examples demonstrate new capabilities
- ✅ Deprecation warnings are clear

### Risks
- **Medium** - New features may reveal edge cases
- **Low** - API stability well-documented upstream

## Release 3: v4.0.0-alpha.1 - New Architecture

**Target Date:** December 15, 2025 (4 weeks after v3.65.0)
**Focus:** Implement v4.x breaking changes

### Scope

#### Phase 3A: Architecture Decision

**Option A: Parallel Package (Recommended)**
```
github.com/scttfrdmn/globus-go-sdk/v3/
├── pkg/services/     # v3.x (stable)
└── pkg/v4/           # v4.x (alpha)
    └── services/
```

**Benefits:**
- Gradual migration path
- Both versions available simultaneously
- Clear import distinction

**Option B: Major Version Bump**
```
github.com/scttfrdmn/globus-go-sdk/v4/
└── pkg/services/     # v4.x only
```

**Benefits:**
- Clean break
- Semantic versioning alignment
- Go module best practices

**Decision Point:** Gather community feedback

#### Phase 3B: Core v4 Features

**1. GlobusApp Pattern**
```go
// New: Centralized app configuration
type GlobusApp struct {
    clientID     string
    clientSecret string
    scopes       []string
}

// New: App-based client initialization
client := transfer.NewClient(
    transfer.WithGlobusApp(app),
)
```

**Files to Create:**
- [ ] `pkg/v4/globusapp.go` - Core app type
- [ ] `pkg/v4/globusapp_test.go` - Unit tests
- [ ] `pkg/v4/services/*/client.go` - Update all clients

**2. Explicit Scope Handling**
```go
// Old (v3): Automatic scope handling
authClient.GetAuthorizationURL(state)

// New (v4): Explicit scopes required
authClient.GetAuthorizationURL(state, []string{
    transfer.Scope,
    groups.Scope,
})
```

**Files to Modify:**
- [ ] All `pkg/v4/services/*/client.go` - Remove default scopes
- [ ] All auth flow methods - Add scope parameters

**3. Base Path Changes**
```go
// Old (v3): Base URL includes version
c.doRequest("endpoint")  // → /v0.10/endpoint

// New (v4): Full path required
c.doRequest("/v0.10/endpoint")
```

**Files to Modify:**
- [ ] `pkg/v4/core/client.go` - Remove base path prepending
- [ ] All service clients - Use full paths

**4. Error Code Handling**
```go
// Old (v3): Default "Error"
if err.Code == "Error" { }

// New (v4): Default nil
if err.Code == nil { }
```

**Files to Modify:**
- [ ] `pkg/v4/core/errors.go` - Change default to nil
- [ ] Update all error handling code

#### Phase 3C: Migration Tooling

**Create Migration Tools:**
- [ ] Static analysis tool to detect v3 patterns
- [ ] Automated refactoring suggestions
- [ ] Migration test suite

**Documentation:**
- [ ] `doc/V4_MIGRATION_GUIDE.md` - Complete guide
- [ ] `doc/V4_BREAKING_CHANGES.md` - All changes listed
- [ ] `examples/v3-to-v4-migration/` - Before/after examples

**Estimated Time:** 6 weeks

### Success Criteria
- ✅ v4 architecture fully implemented
- ✅ All breaking changes documented
- ✅ Migration tooling available
- ✅ Alpha testing reveals no major issues

### Risks
- **High** - Breaking changes may impact users significantly
- **Medium** - Community feedback may require adjustments

## Release 4: v4.0.0-beta.1 - Testing & Refinement

**Target Date:** January 15, 2026 (4 weeks after alpha.1)
**Focus:** Beta testing and feedback integration

### Scope

- [ ] Address alpha feedback
- [ ] Refine migration tooling
- [ ] Update documentation based on user experience
- [ ] Add more examples
- [ ] Performance testing
- [ ] Security audit

**Estimated Time:** 4 weeks

## Release 5: v4.0.0 - Stable Release

**Target Date:** February 15, 2026 (4 weeks after beta.1)
**Focus:** Production-ready v4.0.0

### Scope

- [ ] Final bug fixes
- [ ] Complete documentation
- [ ] Release notes
- [ ] Migration support plan
- [ ] LTS plan for v3.x

## Maintenance Strategy

### v3.x LTS (Long-Term Support)

**Timeline:** Through December 2026 (12 months after v4.0.0)

**Support Level:**
- ✅ Security fixes
- ✅ Critical bug fixes
- ❌ New features
- ❌ Non-critical enhancements

**Branch:** `v3-lts`

### v4.x Active Development

**Timeline:** February 2026 onwards

**Focus:**
- New features
- Performance improvements
- Community-requested enhancements
- Alignment with Python SDK v4.x stable

## Implementation Checklist

### v3.61.0 (October 2025)
- [ ] Fix recursive transfer tests
- [ ] Remove activation method tests
- [ ] Fix resumable transfer test artifacts
- [ ] Update INCOMPLETE_WORK.md
- [ ] CLI decision
- [ ] Release notes

### v3.65.0 (November 2025)
- [ ] Groups enhancements
- [ ] Search UpdateIndex
- [ ] FlowTimer support
- [ ] Subscription management
- [ ] Deprecation markers
- [ ] Complete documentation
- [ ] Release notes

### v4.0.0-alpha.1 (December 2025)
- [ ] Architecture decision
- [ ] GlobusApp implementation
- [ ] Explicit scope handling
- [ ] Base path changes
- [ ] Error code updates
- [ ] Migration tooling
- [ ] Release notes

### v4.0.0-beta.1 (January 2026)
- [ ] Alpha feedback integration
- [ ] Refined migration tools
- [ ] Updated documentation
- [ ] Performance testing
- [ ] Security audit
- [ ] Release notes

### v4.0.0 (February 2026)
- [ ] Final bug fixes
- [ ] Complete documentation
- [ ] LTS plan for v3.x
- [ ] Community support plan
- [ ] Release announcement

## Questions & Decisions Needed

### Immediate (for v3.61.0)
1. **CLI Commands:** Implement, remove, or defer?
2. **Test Coverage:** What level for fixed tests?

### Short-term (for v3.65.0)
1. **Feature Priority:** Which v3.61-v3.65 features are most valuable?
2. **Breaking Changes:** Any v3.x deprecations to add?

### Long-term (for v4.0.0)
1. **Architecture:** Parallel package or major version bump?
2. **Migration Support:** How much tooling to provide?
3. **v3 LTS:** 12 months enough? 18 months?

## Success Metrics

### Technical Metrics
- CI/CD pipeline success rate: >99%
- Test coverage: >85%
- Integration test pass rate: 100%
- Documentation coverage: 100%

### Community Metrics
- Migration guide clarity (user survey)
- Feature adoption rate (telemetry)
- Issue resolution time (<7 days)
- Community contributions (increase)

## Resources & Dependencies

### Team
- 1 Lead Developer (full-time)
- 1 Documentation Specialist (part-time)
- Community contributors

### Dependencies
- Python SDK stability (v3.65.0 is stable)
- Python SDK v4.0.0 release date (currently beta)
- Community feedback cycles

### Infrastructure
- CI/CD pipeline (GitHub Actions)
- Documentation hosting (GitHub Pages)
- Package distribution (Go modules)

## Conclusion

This roadmap provides a clear path from v3.60.0 to v4.0.0:

1. **v3.61.0** - Quick wins, test alignment
2. **v3.65.0** - Feature parity with upstream
3. **v4.0.0** - Modern architecture, breaking changes

The phased approach allows for:
- Stable v3.x for current users
- Gradual feature additions
- Planned breaking changes in v4.x
- Long-term maintenance strategy

**Next Action:** Review and approve this roadmap, then begin v3.61.0 implementation.
