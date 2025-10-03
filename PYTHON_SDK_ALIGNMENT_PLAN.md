# Python SDK Alignment & Upgrade Plan

**Date:** October 2, 2025
**Current Go SDK Version:** v3.60.0
**Target Python SDK Version:** v3.65.0
**Future Target:** v4.0.0

## Executive Summary

After analyzing the upstream Globus Python SDK and our Go SDK implementation, the "incomplete work" documented in `INCOMPLETE_WORK.md` consists primarily of **test artifacts** that don't align with actual Python SDK patterns. The Python SDK handles these operations differently, and several referenced methods (like activation) are **deprecated**.

## Analysis of "Incomplete" Methods

### ❌ Methods That Don't Exist in Python SDK

#### 1. `SubmitTextFileCreation()`
- **Status:** Not a Python SDK method
- **Reality:** Tests reference a non-existent method
- **Action:** Remove test references, use proper file operations instead

#### 2. `RecursiveTransfer()` as a standalone method
- **Status:** Doesn't exist as a direct method
- **Reality:** We **already implemented** this correctly as `SubmitRecursiveTransfer()`
- **Python SDK Pattern:** Uses `TransferData.add_item(src, dest, recursive=True)`
- **Our Implementation:** `pkg/services/transfer/recursive.go:93` - `SubmitRecursiveTransfer()`
- **Action:** Update tests to use existing `SubmitRecursiveTransfer()` method

#### 3. `GetActivationRequirements()` & `ActivateEndpoint()`
- **Status:** **DEPRECATED** in Python SDK v3.61.0+
- **Reason:** Globus Connect Server v4 end-of-life
- **Python SDK Note:** "may no longer function with the Transfer API"
- **Our Implementation:** Already removed (line 290-293 in `client.go`)
- **Action:** Remove test references, document deprecation

#### 4. `ActivationProfile` field
- **Status:** Related to deprecated activation methods
- **Action:** Remove test references

### ✅ What We Already Have (Correctly Implemented)

Our Go SDK **already has** proper implementations:
- ✅ **Recursive transfers**: `SubmitRecursiveTransfer()` in `recursive.go`
- ✅ **Resumable transfers**: Complete implementation in `resumable.go`
- ✅ **Basic transfers**: `SubmitTransfer()` with options
- ✅ **Activation note**: Lines 290-293 explain modern auto-activation

## Required Changes for v3.65.0 Alignment

### Phase 1: Fix Test Artifacts (Immediate)

#### A. Update Integration Tests
1. **File:** `pkg/services/transfer/integration_test.go:676`
   - Change from: `t.Skip("RecursiveTransfer method not yet implemented")`
   - Change to: Use existing `SubmitRecursiveTransfer()` method
   - Implementation already exists!

2. **File:** `pkg/services/transfer/integration_test.go:911,933,953,995`
   - Remove activation test blocks
   - Add comment explaining deprecation
   - Document auto-activation with proper scopes

3. **File:** `pkg/services/transfer/resumable_integration_test.go` (6 locations)
   - Remove `SubmitTextFileCreation()` references
   - Use proper file creation methods (mkdir + transfer)
   - Align with actual API patterns

#### B. Update CLI Stubs
**File:** `cmd/globus-cli/transfer/transfer.go`

Options:
1. **Implement** - Connect to actual SDK methods
2. **Remove** - If CLI is not a v3.65.0 priority
3. **Document** - Mark as "Coming in v4.0"

### Phase 2: Add Python SDK v3.61-v3.65 Features

#### A. Groups Service Enhancements (v3.65.0)

**File:** `pkg/services/groups/client.go`

Add methods:
```go
// get_my_groups with statuses parameter (v3.65.0)
func (c *Client) GetMyGroups(ctx context.Context, options *GetMyGroupsOptions) (*GroupList, error)

type GetMyGroupsOptions struct {
    Statuses []string  // New in v3.65.0: filter by status
    Limit    int
    Offset   int
}
```

**File:** `pkg/services/groups/batch.go` (new file)

Add batch operations:
```go
// BatchMembershipActions.change_role() (v3.65.0)
type BatchMembershipActions struct {
    operations []MembershipOperation
}

func (b *BatchMembershipActions) ChangeRole(userID, groupID, role string) *BatchMembershipActions

// GroupsManager.change_roles() (v3.65.0)
func (c *Client) ChangeRoles(ctx context.Context, operations []RoleChange) (*BatchResult, error)
```

#### B. Search Service Enhancement (v3.64.0)

**File:** `pkg/services/search/client.go`

Add method:
```go
// update_index for modifying index names/descriptions (v3.64.0)
func (c *Client) UpdateIndex(ctx context.Context, indexID string, update *IndexUpdate) (*Index, error)

type IndexUpdate struct {
    DisplayName *string  `json:"display_name,omitempty"`
    Description *string  `json:"description,omitempty"`
}
```

#### C. Flows Service Enhancement (v3.65.0)

**File:** `pkg/services/flows/timer.go` (new file)

Add FlowTimer payload class:
```go
// FlowTimer payload class (v3.65.0)
type FlowTimer struct {
    Name           string                 `json:"name"`
    FlowID         string                 `json:"flow_id"`
    FlowInput      map[string]interface{} `json:"flow_input"`
    Schedule       TimerSchedule          `json:"schedule"`
    CallbackURL    string                 `json:"callback_url,omitempty"`
}

type TimerSchedule struct {
    Type     string `json:"type"`      // "cron" or "interval"
    Value    string `json:"value"`     // cron expression or interval
    Timezone string `json:"timezone,omitempty"`
}
```

#### D. Groups Subscription Management (v3.62.0-v3.63.0)

**File:** `pkg/services/groups/client.go`

Add method:
```go
// set_subscription_admin_verified (v3.63.0, renamed from v3.62.0)
func (c *Client) SetSubscriptionAdminVerified(ctx context.Context, groupID, subscriptionID string) error
```

### Phase 3: Deprecation Cleanup

#### A. Transfer Service Deprecations (v3.64.0)

**File:** `pkg/services/transfer/models.go`

Add deprecation markers:
```go
// TransferItem.SymlinkTarget - DEPRECATED in v3.64.0
// Deprecated: add_symlink_item is deprecated. Use standard file operations.
SymlinkTarget string `json:"symlink_target,omitempty"`

// TransferData.RecursiveSymlinks - DEPRECATED in v3.64.0
// Deprecated: recursive_symlinks parameter is deprecated.
RecursiveSymlinks bool `json:"recursive_symlinks,omitempty"`

// TransferData.SkipActivationCheck - DEPRECATED in v3.64.0
// Deprecated: skip_activation_check is deprecated. Modern endpoints auto-activate.
SkipActivationCheck bool `json:"skip_activation_check,omitempty"`
```

#### B. Compute Client Alias (v3.61.0)

**File:** `pkg/services/compute/client.go`

Add deprecation notice:
```go
// DEPRECATED: ComputeClient is deprecated in favor of explicit versioning.
// Use ComputeClientV2 or ComputeClientV3 based on your endpoint version.
// This alias will be removed in v4.0.0.
type ComputeClient = Client  // Points to ComputeClientV3 by default
```

## v4.0.0 Migration Plan

### Major Breaking Changes from Python SDK 4.0.0b2

#### 1. Client Initialization Changes

**Current (v3.x):**
```go
client, err := transfer.NewClient(
    transfer.WithAuthorizer(authorizer),
)
```

**Future (v4.x):**
```go
app := globus.NewGlobusApp(
    globus.WithClientID(clientID),
    globus.WithClientSecret(clientSecret),
)

client, err := transfer.NewClient(
    transfer.WithGlobusApp(app),
)
```

#### 2. Explicit Scope Handling

**Change:** No default scopes - must be explicit

**Current (v3.x):**
```go
// Scopes handled automatically
authClient.GetAuthorizationURL(state)
```

**Future (v4.x):**
```go
// Must specify scopes explicitly
authClient.GetAuthorizationURL(state, []string{
    transfer.TransferScope,
    groups.GroupsScope,
})
```

#### 3. Base Path Removal

**Change:** Full paths required in all methods

**Current (v3.x):**
```go
// BaseURL includes /v0.10/
c.doRequestLowLevel(ctx, "GET", "endpoint", nil, nil, &result)
// Results in: https://transfer.api.globus.org/v0.10/endpoint
```

**Future (v4.x):**
```go
// Must use full path
c.doRequestLowLevel(ctx, "GET", "/v0.10/endpoint", nil, nil, &result)
// Must include version in path
```

#### 4. Error Code Changes

**Current (v3.x):**
```go
// Default error code: "Error"
if err.Code == "Error" {
    // Handle generic error
}
```

**Future (v4.x):**
```go
// Default error code: nil
if err.Code == nil {
    // Handle generic error
}
```

### v4.0.0 Implementation Strategy

#### Option A: Parallel v4 Package (Recommended)

**Approach:**
```
pkg/
├── services/        # v3.x (current)
│   ├── auth/
│   ├── transfer/
│   └── ...
└── v4/              # v4.x (new)
    ├── services/
    │   ├── auth/
    │   ├── transfer/
    │   └── ...
    └── globusapp.go
```

**Benefits:**
- Users can migrate gradually
- No breaking changes to existing code
- Clear import paths: `v3/pkg/services/...` vs `v3/pkg/v4/services/...`

#### Option B: Major Version Bump

**Approach:**
```
github.com/scttfrdmn/globus-go-sdk/v4
```

**Benefits:**
- Clean break, clear semantic versioning
- Module system handles versions naturally
- Aligns with Go module best practices

**Drawbacks:**
- Users must update imports
- Can't use both v3 and v4 simultaneously (easily)

## Recommended Release Strategy

### Release 1: v3.61.0 (Immediate - 1-2 weeks)
**Focus:** Fix test artifacts, align with Python SDK patterns

- ✅ Fix `RecursiveTransfer` tests to use existing `SubmitRecursiveTransfer()`
- ✅ Remove deprecated activation method tests
- ✅ Fix `SubmitTextFileCreation` test references
- ✅ Update `INCOMPLETE_WORK.md` to reflect reality
- ✅ Add deprecation markers for activation methods
- ✅ Update CLI stubs (implement, remove, or document)

**No new features** - just alignment fixes

### Release 2: v3.65.0 (2-4 weeks)
**Focus:** Feature parity with Python SDK v3.65.0

- ✅ Groups: Add `GetMyGroups()` with `statuses` parameter
- ✅ Groups: Add batch membership operations
- ✅ Groups: Add subscription management
- ✅ Search: Add `UpdateIndex()` method
- ✅ Flows: Add `FlowTimer` payload class
- ✅ Add deprecation markers for v3.64.0 deprecations
- ✅ Compute: Add deprecation notice for `ComputeClient` alias
- ✅ Update documentation
- ✅ Update CHANGELOG with all v3.61-v3.65 features

### Release 3: v4.0.0-alpha.1 (4-8 weeks)
**Focus:** Implement v4.x architecture

- ✅ Create `pkg/v4/` package structure OR bump to `v4` module
- ✅ Implement `GlobusApp` pattern
- ✅ Remove default scopes
- ✅ Update base path handling
- ✅ Change error code defaults
- ✅ Create migration guide
- ✅ Parallel release with v3.65.x maintenance

### Release 4: v4.0.0 (8-12 weeks)
**Focus:** Stable v4 release

- ✅ Beta testing feedback incorporated
- ✅ Complete documentation
- ✅ Migration tooling
- ✅ Maintain v3.x LTS branch

## Implementation Priorities

### High Priority (v3.61.0)
1. Fix recursive transfer tests ← **Tests fail due to misunderstanding**
2. Remove activation method test references ← **Tests fail due to deprecated APIs**
3. Fix resumable transfer test artifacts ← **Tests fail due to non-existent method**
4. Update INCOMPLETE_WORK.md ← **Document reflects misunderstanding**

### Medium Priority (v3.65.0)
1. Groups enhancements (most requested)
2. Search.UpdateIndex() (commonly used)
3. FlowTimer support (growing use case)
4. Subscription management (enterprise feature)

### Low Priority (v4.0.0)
1. v4 architecture design
2. Migration tooling
3. Breaking change implementation
4. Beta testing phase

## Success Metrics

### v3.61.0
- ✅ All integration tests pass without skips
- ✅ CI/CD remains green
- ✅ No new features, just fixes

### v3.65.0
- ✅ Feature parity with Python SDK v3.65.0
- ✅ All new features documented
- ✅ Integration tests for new features
- ✅ CHANGELOG updated

### v4.0.0
- ✅ Breaking changes clearly documented
- ✅ Migration guide with examples
- ✅ Alpha/beta feedback cycle complete
- ✅ Community adoption plan

## Next Steps

1. **Review this plan** - Confirm strategy and priorities
2. **Start v3.61.0** - Fix test artifacts immediately
3. **Plan v3.65.0** - Gather feedback on desired features
4. **Design v4.0.0** - Architecture decisions and community input

## Questions for Consideration

1. **CLI Priority:** Should we implement, remove, or defer CLI commands?
2. **v4 Approach:** Parallel package or major version bump?
3. **Feature Selection:** Which v3.61-v3.65 features are most valuable?
4. **Timeline:** Aggressive (6 weeks) or conservative (12 weeks) for v3.65.0?
5. **v4 Timing:** Start now or wait for Python SDK 4.0.0 stable release?

## Conclusion

The "incomplete work" is largely a **testing misalignment** rather than missing functionality. Our core implementation is solid and aligns well with Python SDK patterns. The path forward is:

1. **Fix test artifacts** (v3.61.0) - Quick win, clean CI/CD
2. **Add v3.65 features** (v3.65.0) - Catch up to latest stable
3. **Plan v4 migration** (v4.0.0) - Future-proof architecture

This approach maintains stability while modernizing the SDK.
