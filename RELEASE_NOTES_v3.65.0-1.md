<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Globus Go SDK v3.65.0-1

**Release Date:** October 25, 2025
**Python SDK Parity:** v3.65.0

This release brings the Globus Go SDK to full parity with the upstream Globus Python SDK v3.65.0, adding FlowTimer helpers for convenient flow-based timer creation and Groups status filtering.

## ✨ New Features

### FlowTimer Helper

Simplified creation of timers that run Globus Flows, matching Python SDK v3.65.0 functionality:

```go
// Define a flow to run
flowTimer := &timers.FlowTimer{
    FlowID:    "my-flow-id",
    FlowScope: "https://auth.globus.org/scopes/my-flow-id/flow_run",
    FlowInput: map[string]interface{}{
        "source": "/path/to/source",
        "dest":   "/path/to/dest",
    },
    FlowLabel: "My Flow Run",
}

// Create a one-time flow timer
timer, err := timersClient.CreateFlowTimerOnce(
    ctx,
    "Daily Backup",
    time.Now().Add(1*time.Hour),
    flowTimer,
    nil,
)

// Create a recurring flow timer (every 24 hours)
timer, err := timersClient.CreateFlowTimerRecurring(
    ctx,
    "Daily Backup",
    time.Now(),
    "P1D", // ISO 8601: 1 day
    nil,   // No end time
    flowTimer,
    nil,
)

// Create a cron-scheduled flow timer (every Monday at 9 AM)
timer, err := timersClient.CreateFlowTimerCron(
    ctx,
    "Weekly Report",
    "0 9 * * 1",        // Every Monday at 9:00 AM
    "America/New_York", // Timezone
    nil,                // No end time
    flowTimer,
    nil,
)
```

**New Components:**
- `FlowTimer` struct for flow configuration
- `CreateFlowTimerOnce()` - One-time flow execution
- `CreateFlowTimerRecurring()` - Recurring flow execution with ISO 8601 intervals
- `CreateFlowTimerCron()` - Cron-scheduled flow execution

**Example Application:** See `cmd/examples/timers-flow/main.go` for a complete working example.

### Groups Statuses Filter

Filter groups by status in `ListGroups()` operations:

```go
options := &groups.ListGroupsOptions{
    Statuses: []string{"active", "pending"},
}
groups, err := client.ListGroups(ctx, options)
```

This matches the Python SDK v3.65.0 `get_my_groups(statuses=...)` parameter and works with both `ListGroups()` and `ListGroupsV2()` methods.

## 🔄 Python SDK Synchronization Status

| Version | Status | Notes |
|---------|--------|-------|
| v3.63.0 | ✅ Synchronized | Previous release |
| v3.64.0 | ✅ Synchronized | SearchClient.UpdateIndex already present |
| **v3.65.0** | ✅ **Synchronized** | **This release** |

## ✅ Compatibility

- **Fully backward compatible** with v3.63.0-1
- **No breaking changes** - all existing code continues to work
- **Production ready** - comprehensive testing with 100% pass rate

## 📦 What's Included

### New Files
- `pkg/services/timers/flow_timer.go` - FlowTimer implementation
- `pkg/services/timers/flow_timer_test.go` - Comprehensive test suite
- `cmd/examples/timers-flow/main.go` - Complete example application

### Modified Files
- `pkg/services/groups/models.go` - Added Statuses field
- `pkg/services/groups/client.go` - Added statuses query handling
- `pkg/services/groups/client_test.go` - Added status filter tests
- `pkg/services/timers/doc.go` - Added FlowTimer documentation
- `pkg/core/version.go` - Version bump to 3.65.0
- `pkg/globus.go` - Version bump to 3.65.0
- `CHANGELOG.md` - Detailed release notes

## 🧪 Testing

All tests passing:
- **FlowTimer:** 5 test cases covering one-time, recurring, and cron timers
- **Groups Statuses:** 3 test cases covering single, multiple, and combined filters
- **Existing Tests:** All 100+ existing test suites continue to pass

## 📚 Documentation

- Updated package documentation with FlowTimer examples
- Complete working example application
- Detailed CHANGELOG entries
- Inline code documentation

## 🔗 Links

- **Documentation:** https://pkg.go.dev/github.com/scttfrdmn/globus-go-sdk/v3
- **Python SDK:** https://github.com/globus/globus-sdk-python
- **Globus Platform:** https://www.globus.org

## 📥 Installation

```bash
go get github.com/scttfrdmn/globus-go-sdk/v3@v3.65.0-1
```

## 🙏 Acknowledgments

This release maintains synchronization with the excellent work of the Globus Python SDK team.

---

**Full Changelog:** v3.63.0-1...v3.65.0-1
