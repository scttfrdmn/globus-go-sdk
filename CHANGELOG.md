<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed — Auth (v3 module, Phase 2 parity audit vs 3.65.0)

**Breaking** within the v3 line:

- Removed phantom MFA REST methods `GetMFAChallenge` and
  `RespondToMFAChallenge` (no `/oauth2/mfa/*` route at 3.65.0); MFA is completed
  by resubmitting to the token endpoint. `Identity.IdentityID`→`Identity.ID`
  (JSON `id`).

Added: `GetIdentities`, `GetIdentityProviders`, projects CRUD, policies CRUD,
clients CRUD (+`CreateChildClient`/`CreateNativeAppInstance`), client-credentials
CRUD, scopes CRUD, `GetConsents`, `OAuth2GetDependentTokens`,
`OAuth2ValidateToken`, and OIDC `GetOpenIDConfiguration`/`GetJWK`/`Userinfo`.

### Changed — Flows (v3 module, Phase 2 parity audit vs 3.65.0)

**Breaking** within the v3 line:

- Base URL corrected to `https://flows.automate.globus.org/`. `RunFlow` posts to
  `/flows/{id}/run` with input under `body` (was `/runs` + `input`).
  `UpdateFlow`/`UpdateRun` use PUT; `DeleteRun` is `POST /runs/{id}/release`.
- Removed the action-provider surface (`ListActionProviders`,
  `GetActionProvider`, `ListActionRoles`, `GetActionRole`, their iterators/batch,
  and `ActionProvider*`/`ActionRole*` types) — no upstream route at 3.65.0.
- `RunResponse` timestamp keys → `start_time`/`completion_time`; `RunLogEntry`
  timestamp → `time` (removed `run_id`).
- Added `ValidateFlow`, `ValidateRun`, `GetRunDefinition`, `DeleteRun`,
  `ResumeRun`.

### Changed — Search (v3 module, Phase 2 parity audit vs 3.65.0)

**Breaking** within the v3 line:

- Ingest/search/delete now use index-scoped routes (`/index/{id}/ingest`,
  `/index/{id}/search`, `/index/{id}/batch_delete_by_subject`).
- `IndexList` decodes `index_list`; `SearchResponse` reads `gmeta` +
  `has_next_page` (offset-paginated iterator); `DeleteDocumentsResponse` reads
  top-level `task_id`; `TaskStatusResponse` gained `index_id` and dropped
  unverified counters.
- Added `GetSearch`, `Scroll`, `DeleteByQuery`, `GetSubject`/`DeleteSubject`,
  `GetEntry`/`DeleteEntry`, `GetTaskList`, and role ops (`CreateRole`,
  `GetRoleList`, `DeleteRole`).

### Changed — Timers (v3 module, Phase 2 parity audit vs 3.65.0)

**Breaking** within the v3 line:

- Base URL corrected to `https://timer.automate.globus.org/`; paths moved to
  `/jobs/` and `POST /v2/timer` (wrapped `{"timer": ...}`); update is
  `PATCH /jobs/{id}`.
- Create document reshaped (`timer_type`/`name`/`schedule`/`body`, `flow_id` for
  flow timers); `Schedule` uses once/recurring shapes (`interval_seconds`,
  structured `end`). New `NewOnceSchedule`/`NewRecurringSchedule`/
  `NewTransferTimer`/`NewFlowTimer` builders and `CreateJob`.
- `CreateTimer`/`UpdateTimer` take an `interface{}` document; `ResumeTimer` takes
  an optional `*bool`; `CreateFlowTimer`/`CreateTransferTimer` reworked.
- Removed phantom methods/types: `RunTimer`, `ListRuns`, `GetRun`,
  `GetCurrentUser`, cron timers, and `TimerRun`/`RunResult`/`Callback`/
  `CreateTimerRequest`/`UpdateTimerRequest`/`CurrentUserInfo`.

### Changed — Groups (v3 module — `github.com/scttfrdmn/globus-go-sdk/v3`, Phase 2 parity audit vs 3.65.0)

**Breaking** within the v3 line — removed methods that hit nonexistent routes:

- Removed the members sub-resource and roles resource (`ListMembers`/`AddMember`/
  `RemoveMember`/`UpdateMemberRole` + LowLevel variants; `ListRoles`/`GetRole`/
  `CreateRole`/`UpdateRole`/`DeleteRole`; the `ChangeRole(s)`/batch builder;
  `GetGroupSubscription`; and the `Role*`/`GroupSubscription` types). Use
  `BatchMembershipAction` (`POST /groups/{id}`) and read memberships via
  `GetGroup` with `include=memberships`.
- Removed `ListGroups`/`ListGroupsV2` (no list route/pagination upstream) — use
  `GetMyGroups` (`GET /groups/my_groups`, `[]Group`, comma-joined statuses).
- `GetGroup` takes `(ctx, id, *GetGroupOptions)`; `UpdateGroup` uses PUT;
  preferences hit `/preferences`; `GetGroupBySubscriptionID` hits
  `/subscription_info/{id}`; `GroupPolicies` reshaped to real keys.

### Added (v4 module — `github.com/scttfrdmn/globus-go-sdk/v4`)

Closes the wire-visible gap to upstream Python globus-sdk **v4.8.1**.

- **Flows registered APIs** (Python SDK v4.6.0; `per_page` added v4.7.0):
  - `GetRegisteredAPI(ctx, id)` — `GET /registered_apis/{id}`
  - `ListRegisteredAPIs(ctx, options)` — `GET /registered_apis`, marker
    pagination, items under the `registered_apis` key, query params
    `filter_roles` (comma-joined), `orderby`, `marker`, `per_page`
  - `NewRegisteredAPIsPager(options)` — marker-based paginator
  - New types: `RegisteredAPI`, `RegisteredAPIRoles`, `RegisteredAPIList`,
    `ListRegisteredAPIsOptions`
- **Transfer bookmark management** (Python SDK v4.6.0, amended v4.8.0):
  - `CreateBookmark`, `GetBookmark`, `ListBookmarks`, `UpdateBookmark`,
    `DeleteBookmark` — JSON:API under `/v2/bookmarks`
  - New types: `Bookmark`, `BookmarkCreate`, `BookmarkUpdate`, `BookmarkList`,
    `ListBookmarksOptions`. No `pinned` field (removed upstream in v4.8.0).
  - **Divergence:** upstream places these on the experimental
    `TransferClientV2`; the Go v4 module folds them into `transfer.Client`,
    matching the Streams/Tunnels approach. Per upstream v4.8.1 `list_bookmarks`
    is not paginated, so `ListBookmarks` does a single-page fetch. See
    [docs/divergence.md](docs/divergence.md).

### Changed
- **Upstream parity tracking overhauled.** `.github/upstream-versions.json` now
  separates `ported` (the upstream Python SDK version this module actually
  implements) from `seen` (the latest release CI has observed). The
  `check-upstream-releases` workflow now runs daily, enumerates *every* release
  in the gap (not just the newest), files one issue per release, and only bumps
  `seen` — `ported` is bumped by humans when features land. This surfaces parity
  gaps honestly instead of hiding them.
- **v4 module `ported` parity bumped to `v4.8.1`.** `core.Version` (v4) is now
  `4.8.1`. The v4.8.0/v4.8.1 upstream changes not reflected here are
  Python-internal (orjson support, representation providers,
  `get_current_transport`) and have no Go equivalent.
- **Pre-commit hook now checks both Go modules.** `scripts/install-hooks.sh`
  previously ran `go fmt`/`go vet`/`go test` only from the repo root (the v3
  module); it now iterates over both the root (v3) and `v4/` modules.

### Fixed
- **Version constants now report the true parity point.** The `core.Version`
  constant (and `pkg.Version` in the v3 module) previously read `4.4.0` in both
  modules. Corrected to `3.65.0` (v3 module, final v3 line) and, with this
  release, `4.8.1` (v4 module). `UserAgent()` now reports the correct version.

## [4.5.0-2] - 2026-04-03

### Added (v4 module — `github.com/scttfrdmn/globus-go-sdk/v4`)

Implements the Python SDK's *application framework* layer — the pieces that manage
credentials, refresh tokens, paginate results, and handle interactive login without
writing plumbing code. Mirrors `globus_sdk.authorizers`, `globus_sdk.token_storage`,
`globus_sdk.paging`, `globus_sdk.login_flows`, and `globus_sdk.globus_app`.

- **`v4/pkg/core`** — `Authorizer` interface added to `core.Config`:
  - New `Authorizer` interface: `GetAuthorizationHeader(ctx) (string, error)` + `HandleMissingAuthorization(ctx) bool`
  - `Config.Authorizer` field: when set, `DoRequest` calls it for the Authorization header instead of using the static `AccessToken`
  - `Config.Validate()` now accepts either `AccessToken` OR `Authorizer` (previously required `AccessToken`)

- **`v4/pkg/authorizers`** — BETA — mirrors `globus_sdk.authorizers`:
  - `AccessTokenAuthorizer` — static bearer token, never refreshes
  - `RefreshTokenAuthorizer` — auto-refreshes using a refresh token; options: `WithInitialAccessToken`, `WithOnRefresh`, `WithAuthBaseURL`, `WithHTTPClient`
  - `ClientCredentialsAuthorizer` — obtains tokens via OAuth2 client credentials grant; options: `WithClientCredentialsAuthBaseURL`, `WithClientCredentialsHTTPClient`
  - Internal `renewingAuthorizer` base: mutex-protected, 60-second proactive refresh threshold

- **`v4/pkg/tokenstorage`** — BETA — mirrors `globus_sdk.token_storage`:
  - `TokenData` struct with `IsExpired()` and `ExpiresIn()` helpers
  - `TokenStorage` interface: `Store`, `Get`, `Remove`, `GetAll`, `Close`
  - `MemoryTokenStorage` — thread-safe, in-process storage
  - `JSONTokenStorage` — atomic file-backed storage (write-then-rename); namespace-partitioned; format `{"version":"2.0","by_rs":{...}}`

- **`v4/pkg/paging`** — BETA — mirrors `globus_sdk.paging`:
  - Generic `Paginator[T any]` interface: `HasNext() bool` + `NextPage(ctx) ([]T, error)`
  - `LimitOffsetPaginator[T]` — auto-increments offset until `fetched >= total`
  - `MarkerPaginator[T]` — cursor-based, stops when server returns `hasMore = false`
  - `NextTokenPaginator[T]` — token-based, stops when `hasNextPage = false`
  - `JSONAPIPaginator[T]` — follows `Links.Next` absolute URLs (JSON:API spec)
  - `gcs.CollectionPager` migrated to use `JSONAPIPaginator` internally (API unchanged)
  - `NewPager()` factory methods added to service clients:
    - `transfer.Client`: `NewEndpointsPager`, `NewTasksPager`, `NewTunnelsPager`
    - `flows.Client`: `NewFlowsPager`, `NewRunsPager`
    - `search.Client`: `NewIndexesPager`
    - `groups.Client`: `NewGroupsPager`, `NewMembersPager`
    - `compute.Client`: `NewEndpointsPager`, `NewTasksPager`

- **`v4/pkg/login`** — BETA — mirrors `globus_sdk.login_flows`:
  - `LoginFlowManager` interface + `AuthParams` + `LoginResult`
  - `CommandLineLoginFlowManager` — prints authorization URL, reads auth code from stdin, exchanges for tokens; handles Globus Auth `other_tokens` extension for multi-resource-server responses; options: `WithCLIRedirectURI`, `WithCLIAuthBaseURL`, `WithCLIHTTPClient`

- **`v4/pkg/app`** — BETA — mirrors `globus_sdk.globus_app`:
  - `GlobusApp` interface: `Login`, `Logout`, `LoginRequired`, `GetAuthorizer`, `AddScopeRequirements`, `Close`
  - `AppConfig`: `TokenStorage`, `LoginFlowManager`, `RequestRefreshTokens`, `Environment`
  - `UserApp` — interactive browser login; `GetAuthorizer` returns `RefreshTokenAuthorizer` (if refresh token available) or `AccessTokenAuthorizer`; `LoginRequired()` checks all registered resource servers
  - `ClientApp` — machine-to-machine; `GetAuthorizer` returns `ClientCredentialsAuthorizer`; `Login`/`Logout` are no-ops

- **Version**: v4 module version bumped to `4.5.0-2`

## [4.5.0-1] - 2026-04-03

### Added (v4 module — `github.com/scttfrdmn/globus-go-sdk/v4`)

- **GCS package** (`v4/pkg/services/gcs`) — EXPERIMENTAL:
  - `CollectionClient` with `GetCollection`, `ListCollections`, `UpdateCollection`, `DeleteCollection`, `NewCollectionPager`
  - `Downloader` with `ReadFile`, `ReadFileAsText` for HTTPS file access without Transfer service
  - `CollectionPager` — JSON:API cursor-based paginator

- **Auth**: `GetAuthorizationURL`, `StartDeviceAuthorization`, `PollDeviceAuthorization`, `WaitForDeviceAuthorization` (RFC 8628 device flow, Python SDK v4.0.0)

- **Compute**: `RegisterFunction`, `UpdateFunction`, `DeleteFunction`, `GetTaskStatus`, `ListTasks`, `CancelTask`, `RunBatch`, `GetBatchStatus`; models: `FunctionDefinition`, `FunctionRegistration`, `FunctionUpdate`, `BatchTaskRequest`, `BatchTaskResponse`, `BatchTaskStatus`

- **Flows**: `CreateFlow`, `UpdateFlow`, `DeleteFlow`, `UpdateRun`, `GetRunLogs`, `WaitForRun`; action providers: `ListActionProviders`, `GetActionProvider`, `ListActionRoles`, `GetActionRole`; models: `ActionProvider`, `ActionProviderList`, `ActionRole`, `ActionRoleList`

- **Groups**: `UpdateMemberRole`, `GetGroupPolicies`, `SetGroupPolicies`, `GetMyGroups`, `GetRole`, `ListRoles`, `CreateRole`, `UpdateRole`, `DeleteRole`, `GetIdentityPreferences`, `SetIdentityPreferences`, `GetMembershipFields`, `SetMembershipFields`; models: `RoleCreate`, `RoleUpdate`, `RoleList`, `IdentityPreferences`, `MembershipFields`

- **Search**: `IndexList` (Python SDK v4.5.0), `GetTaskStatus`; models: `IndexList`, `ListIndexesOptions`, `IngestTaskStatus`

- **Timers**: `CreateTransferTimer`, `RunTimer`, `ListRuns`, `GetRun`; models: `TimerRun`, `TimerRunList`, `ListRunsOptions`

- **Transfer**: `GetSubmissionID`, `ListStreamAccessPoints`, `UpdateTunnel` nil-data validation

- **Test coverage**: `v4/pkg/testhelpers` shared helpers; `client_test.go` for all 7 service packages; `tunnel_test.go` covering all 8 BETA Streams/Tunnel methods

- **Version**: v4 module version bumped to `4.5.0`

### Changed (v3 module — `github.com/scttfrdmn/globus-go-sdk/v3`)

- **Version alignment**: v3 module version scheme now tracks the upstream Python SDK
  version exactly. Previous version `3.60.0-1` is superseded by `4.5.0-1` to
  reflect that this module implements the Python SDK v4.5.0 service client API
  in Go v3-module style (stable, production-ready). No functional changes.


## [4.4.0-2] - 2026-02-17

### Fixed
- **Groups `SetSubscriptionAdminVerified` endpoint** (v3 module): Regression from rebase during v4.4.0-1 reverted the correct API path
  - Endpoint corrected: `/groups/{id}/subscription_id` → `/groups/{id}/subscription`
  - `DATA_TYPE` corrected: `subscription_id_update` → `subscription_update`

### Added (v4 module — `github.com/scttfrdmn/globus-go-sdk/v4`)
- **v4 module synced from 4.2.0 to 4.4.0**
- **Search**: `ReopenIndex(ctx, indexID)` — reopen a previously deleted index (Python SDK v4.0.0b1)
- **Flows**: `FlowAuthenticationPolicy` struct; `FlowCreate` and `FlowUpdate` types with `AuthenticationPolicy` field (Python SDK v4.1.0)
- **Transfer**: Full Streams/Tunnel API (Python SDK v4.3.0–v4.4.0):
  - `CreateTunnel`, `GetTunnel`, `UpdateTunnel`, `DeleteTunnel`, `ListTunnels`
  - `GetStreamAccessPoint`
  - `GetTunnelEvents`
  - New types: `Tunnel`, `TunnelList`, `TunnelCreate`, `TunnelUpdate`, `StreamAccessPoint`, `TunnelEvent`, `TunnelEventList`, `ListTunnelsOptions`, `ListTunnelEventsOptions`
- **Version**: v4 module version constant bumped to `4.4.0`

## [4.4.0-1] - 2026-02-17

### Added
- **Python SDK v4.4.0 synchronization** (v3 module)
  - **Transfer**: `GetTunnelEvents(ctx, tunnelID, options)` - fetch events associated with a Globus Streams tunnel

## [4.3.0-1] - 2026-01-15

### Added
- **Python SDK v4.3.0 synchronization** (v3 module) — Comprehensive Globus Streams API support in `TransferClient` (`pkg/services/transfer/streams.go`):
  - `CreateTunnel(ctx, data)` - create a new Globus Streams tunnel
  - `GetTunnel(ctx, tunnelID)` - retrieve a tunnel by ID
  - `UpdateTunnel(ctx, tunnelID, data)` - update an existing tunnel
  - `DeleteTunnel(ctx, tunnelID)` - delete a tunnel
  - `ListTunnels(ctx, options)` - list tunnels owned by current user
  - `GetStreamAccessPoint(ctx, accessPointID)` - fetch a Stream Access Point
  - New types: `Tunnel`, `TunnelList`, `CreateTunnelData`, `UpdateTunnelData`, `StreamAccessPoint`, `TunnelEvent`, `TunnelEventList`

## [4.2.0-1] - 2025-12-10

### Added
- **Python SDK v4.2.0 synchronization** (v3 module)
  - **Timers**: `FlowUserScope(flowID string) string` — returns the scope string needed for a TimersClient to execute a specific flow (Python SDK equivalent: `add_app_flow_user_scope()` on GlobusApp)
  - **Timers**: `Close()` method on `Client` for releasing idle HTTP connections
- **v4 module created** (`github.com/scttfrdmn/globus-go-sdk/v4`) — clean-room v4 implementation with context-first API design, explicit scopes, and `Close()` on all service clients

## [4.1.0-1] - 2025-11-01

### Added
- **Python SDK v4.1.0 synchronization** (v3 module)
  - **Flows**: `FlowAuthenticationPolicy` struct for specifying authentication requirements on flows
  - **Flows**: `AuthenticationPolicy *FlowAuthenticationPolicy` field added to `FlowCreateRequest` and `FlowUpdateRequest`
  - Note: Service support for authentication policy may be pending as of this release

## [4.0.1-1] - 2025-10-20

### Fixed
- **Python SDK v4.0.1 synchronization** (v3 module)
  - **Transfer**: Added missing `SetSubscriptionAdminVerified(ctx, endpointID, subscriptionID)` with corrected route (Python SDK since v3.59.0, route-fixed in v4.0.1)

## [4.0.0-1] - 2025-10-15

### Added
- **Python SDK v4.0.0 synchronization** (v3 module)
  - **Search**: `ReopenIndex(ctx, indexID)` — reopen a previously deleted search index (Python SDK v4.0.0b1)
  - **Flows**: `ListRuns()` was already present in the Go SDK

### Deprecated
- The following methods remain in the codebase with deprecation warnings but are scheduled for removal in Go SDK v5.0.0:
  - `TransferClient.SetupGridFTPV4Server()` (deprecated v3.61.0)
  - `TransferClient.ConfigureGCSV4Endpoint()` (deprecated v3.61.0)
  - `TransferClient.GetGCSV4ServerList()` (deprecated v3.61.0)
  - `NewComputeClientV2()` in pkg root (deprecated v3.61.0)
  - `GroupsClient.SetSubscriptionAdminVerifiedID()` (deprecated v3.63.0)

### Technical Notes
- **Breaking changes from Python SDK v4.0.0 not yet applied to Go SDK v4.x**:
  - `GlobusAPIError.code` still defaults to `"Error"` (Python SDK changed default to `nil`)
  - Client base path handling unchanged (Python SDK removed automatic base_path prepending)
  - Scope system changes (Python SDK made Scope immutable, added ScopeParser) not yet ported
  - These breaking changes will be addressed in a future Go SDK v5.0.0 release

## [3.65.0-1] - 2025-10-02

### Added
- **Python SDK v3.65.0 synchronization**
  - **Groups**: `GetMyGroups()` now accepts `statuses []string` parameter for filtering groups by membership status (e.g., "active", "invited", "pending")
  - **Groups**: New batch role change operations:
    - `ChangeRole(ctx, groupID, identityID, roleID string) error` - single role change
    - `ChangeRoles(ctx, changes []RoleChange) (*BatchRoleChangeResult, error)` - batch role changes
    - `NewBatchMembershipActions() *BatchMembershipActions` - fluent batch builder with `ChangeRole()` and `Execute()` methods
  - **Flows**: New `FlowTimer` payload class for creating Globus Timers that execute flows:
    - `NewFlowTimer(name, flowID string, flowInput map[string]interface{}, schedule TimerSchedule) *FlowTimer`
    - Schedule types: `NewCronSchedule()`, `NewIntervalSchedule()`, `NewOnceSchedule()`
    - Builder methods: `WithCallbackURL()`, `WithFlowScope()`, `WithRunManagers()`, `WithRunMonitors()`
    - `Validate()` method for pre-submission validation

### Technical Details
- **Version**: Updated SDK version constant to 3.65.0
- **New Files**: `pkg/services/groups/batch.go`, `pkg/services/flows/timer.go`
- **Python SDK Parity**: Maintains synchronization with upstream Globus Python SDK v3.65.0
- **Backward Compatibility**: Full backward compatibility; `GetMyGroups()` `statuses` parameter is optional (nil = all statuses)

## [3.64.0-1] - 2025-09-25

### Added
- **Python SDK v3.64.0 synchronization**
  - **Search**: `UpdateIndex(ctx context.Context, indexID string, request *IndexUpdateRequest) (*Index, error)` method for updating search index metadata

### Technical Details
- **Version**: Updated SDK version constant to 3.64.0
- **Python SDK Parity**: Maintains synchronization with upstream Globus Python SDK v3.64.0

## [3.63.0-1] - 2025-09-18

### Changed
- **Python SDK v3.63.0 synchronization**
  - **Method Rename**: `SetSubscriptionAdminVerifiedID` renamed to `SetSubscriptionAdminVerified` in Groups client
  - Updated method naming to match upstream Python SDK v3.63.0 naming convention
  - Updated all tests to use new method name for consistency

### Deprecated
- **Groups Client Method**
  - `SetSubscriptionAdminVerifiedID()` method is now deprecated in favor of `SetSubscriptionAdminVerified()`
  - Deprecated method remains functional and delegates to the new method for backward compatibility
  - Will be removed in a future major version

### Technical Details
- **Version**: Updated SDK version constant to 3.63.0
- **Backward Compatibility**: Full backward compatibility maintained through deprecated method delegation
- **Testing**: All existing tests updated to use new method names while maintaining deprecated method testing
- **Python SDK Parity**: Maintains synchronization with upstream Globus Python SDK v3.63.0

This release maintains our commitment to tracking upstream Python SDK releases while ensuring backward compatibility for existing users.

## [3.62.0-3] - 2025-08-08

### Added
- **Comprehensive Testing Enhancement Infrastructure**
  - Complete Phase 1 & Phase 2 testing infrastructure following Python SDK patterns
  - **71 comprehensive tests** across unit, functional, and integration test suites
  - **Metadata-driven testing system** with JSON test scenarios for enhanced test organization
  - **Shared testing utilities** in `pkg/testhelpers/` for consistent test patterns across services
  - **Enhanced error scenario testing** with systematic HTTP error code coverage (4xx, 5xx responses)
  - **Workflow-based functional testing** for end-to-end user journey validation
  - **Python SDK parity method testing** covering all 9 parity methods with comprehensive validation

- **New Testing Infrastructure Files**
  - `pkg/testhelpers/fixtures.go` - Shared testing infrastructure and utilities
  - `pkg/services/groups/unit/*` - Comprehensive unit testing suite with 8 test files
  - `pkg/services/groups/functional/*` - Workflow-based functional tests
  - `pkg/services/groups/integration/*` - End-to-end integration testing
  - `TESTING_ENHANCEMENT_PLAN.md` - Complete testing strategy and implementation roadmap

- **Enhanced Test Coverage**
  - **Error Scenario Testing**: Systematic testing of all HTTP error conditions with JSON error response parsing
  - **Subscription Method Testing**: Complete test coverage for v3.62.0 subscription functionality
  - **Python SDK Parity Validation**: Integration tests covering all 9 Python SDK parity methods
  - **Metadata-Driven Test Scenarios**: 15+ structured test cases with variable substitution and templates
  - **Network Error Handling**: Timeout, connection failure, and network-level error scenario testing

### Changed
- **Improved Error Handling in Groups Client**
  - Enhanced `doRequestLowLevel` method to properly handle `core.Error` types
  - Fixed JSON error response parsing with proper `GlobusError` creation
  - Improved error propagation from core HTTP client to service-specific clients

- **Enhanced Test Organization**
  - Restructured groups service testing into unit, functional, and integration test suites
  - Implemented consistent test patterns following upstream Python SDK approaches
  - Added emoji-based test logging for better test output readability

### Fixed
- **Network Timeout Test Stability**
  - Fixed hanging network timeout tests by replacing infinite blocking with controlled timeouts
  - Improved test reliability and reduced CI/CD execution time

### Technical Details
- **Files Modified**: 4 files enhanced (422 lines added)
- **New Files**: 13 new test infrastructure files (4,000+ lines)
- **Test Suite Coverage**: Unit tests, functional workflows, integration scenarios, error handling, and model validation
- **Python SDK Parity**: Complete testing of all subscription management, policy configuration, identity preferences, and membership field methods
- **Infrastructure Improvements**: Mock server enhancements, variable substitution system, dependency resolution, and test case generation utilities

This release significantly enhances the SDK's testing infrastructure to ensure robust quality assurance and maintainability, following proven patterns from the upstream Python SDK while maintaining full backward compatibility.

## [3.62.0-2] - 2025-01-27

### Fixed
- **Version consistency fix**
  - Corrected Version constant in `pkg/core/version.go` from "3.60.0" to "3.62.0"
  - Ensures consistency with v3.62.0 release tags and numbering
  - Addresses oversight from v3.62.0-1 release process

## [3.62.0-1] - 2025-01-27

### Added
- **Python SDK v3.62.0 feature synchronization**
  - Maintained synchronized versioning with Python SDK v3.62.0
  - Groups service subscription_id support
  - SetSubscriptionAdminVerifiedID() method for setting group subscription IDs (admin-only)
  - GetGroupSubscription() method for retrieving group subscription information
  - GroupSubscription type for handling subscription data

### Changed
- **Version synchronization**
  - Updated SDK version to 3.62.0 to match Python SDK v3.62.0
  - All changes maintain backward compatibility with existing v3.61.x code

## [3.61.0-1] - 2025-01-27

### Added
- **Python SDK v3.61.0 feature synchronization**
  - Maintained synchronized versioning with Python SDK v3.61.0
  - Added comprehensive deprecation warnings for legacy functionality

### Deprecated
- **Globus Connect Server v4 support**
  - SetupGridFTPV4Server() method deprecated
  - ConfigureGCSV4Endpoint() method deprecated  
  - GetGCSV4ServerList() method deprecated
  - GCSV4Config, GCSV4ServerList, GCSV4Server types deprecated
  - All GCS v4 methods will emit deprecation warnings when used
- **ComputeClient alias deprecated**
  - ComputeClient type alias deprecated in favor of compute.Client
  - NewComputeClientV2() function deprecated in favor of compute.NewClient()
  - Users encouraged to use compute.Client directly

### Changed
- Updated SDK version to v3.61.0 to maintain Python SDK synchronization

## [3.60.0-1] - 2025-01-27

### Added
- **Version synchronization with Python SDK**
  - Updated versioning to hybrid format `[PYTHON_SDK_VERSION]-[GO_SDK_BUILD]` (v3.60.0-1)
  - Implemented synchronized versioning with Python SDK v3.60.0  
  - Added comprehensive versioning strategy documentation (VERSIONING_STRATEGY.md)
  - Updated module path to github.com/scttfrdmn/globus-go-sdk/v3
- **Globus Auth Requirements Error (GARE) support**
  - Added GlobusAuthRequirementsError type for handling dependent consent errors
  - Implemented recognition of `dependent_consent_required` errors from Auth API
  - Added support for authorization parameters containing dependent scopes
  - Added helper functions: IsGlobusAuthRequirementsError(), IsConsentRequired(), IsDependentConsentRequired()
- **Unified error handling system**
  - Standardized `GlobusError` type across all services
  - Added consistent error context and debugging information
  - Implemented service-specific error codes and messages
- **Consistent client initialization patterns**
  - Unified `NewClient()` functions across all services
  - Standardized configuration and options handling
  - Enhanced client lifecycle management
- **Standardized response and pagination patterns**
  - Unified `Response[T]` wrapper structures
  - Consistent `PaginatedResponse[T]` across all services
  - Enhanced metadata handling and request tracking
- **Updated API versions to match current Globus APIs**
  - Transfer API updated to latest v0.10+ endpoints
  - Auth API aligned with current OAuth2 specifications
  - Groups API updated to v2 endpoints
  - Search API updated to v1 with latest features
  - Flows API updated to v1 endpoints
  - Compute API updated to v2 endpoints
  - Timers API updated to v1 endpoints
- **Enhanced deprecation system matching Python SDK**
  - Added deprecation warnings and migration guidance
  - Implemented deprecation lifecycle management
  - Added deprecation reporting tools
- **Keep a Changelog compliance**
  - Improved changelog structure and consistency
  - Added semantic versioning compliance
  - Enhanced release documentation standards

### Changed
- **BREAKING**: Version updated from v0.9.15 to v3.60.0 to align with Python SDK
- **BREAKING**: Unified error handling - all services now use `GlobusError` type
- **BREAKING**: Standardized client initialization - all services use consistent `NewClient()` pattern
- **BREAKING**: Consistent response structures - all services use `Response[T]` wrapper
- **BREAKING**: Updated API endpoints to match current Globus APIs
- **BREAKING**: Reorganized package structure for better consistency
- Enhanced documentation structure and consistency
- Updated examples and documentation for v3.60.0

### Deprecated
- Legacy error handling patterns (will be removed in v4.0.0)
- Old client initialization methods (will be removed in v4.0.0)
- Inconsistent response structures (will be removed in v4.0.0)

### Removed
- Legacy debugging utilities (moved to proper package structure)
- Deprecated lint tools (replaced with modern alternatives)
- Inconsistent internal APIs (replaced with unified patterns)

### Fixed
- Fixed internal consistency issues across services
- Corrected API version mismatches
- Fixed package conflicts in debug files
- Resolved function redeclarations across the codebase
- Updated auth and transfer client usage patterns
- Replaced deprecated io/ioutil with io package functions
- Fixed variable naming to avoid conflicts (e.g., `err` → `tokenErr`)
- Improved error handling in contract tests
- Fixed missing imports in compute example files

### Security
- Enhanced token handling security
- Improved credential validation mechanisms  
- Updated security practices to match current standards

### Migration from v0.9.x

This release introduces **breaking changes** that require migration:

1. **Update import paths**:
   ```go
   // OLD
   import "github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
   
   // NEW  
   import "github.com/scttfrdmn/globus-go-sdk/v3/pkg/services/auth"
   ```

2. **Update go.mod**:
   ```bash
   go get github.com/scttfrdmn/globus-go-sdk/v3
   ```

3. **Version tracking**: The SDK now follows Python SDK versioning with format `[PYTHON_SDK_VERSION]-[GO_SDK_BUILD]`

4. **GARE support**: New error handling for dependent consent scenarios - see auth package documentation

For detailed migration guidance, see [VERSIONING_STRATEGY.md](VERSIONING_STRATEGY.md).

## [0.9.15] - 2025-05-08

### Fixed
- Properly tagged release for the connection pool functions fix (issue #13)
  - Ensured correct Git tag pointing to the fixed code
  - Verified build works with downstream dependencies
  - Fixed tagging issues from previous release attempts

## [0.9.15] - 2025-05-08

### Fixed
- Properly tagged release for the connection pool functions fix (issue #13)
  - Ensured correct Git tag pointing to the fixed code
  - Verified build works with downstream dependencies
  - Fixed tagging issues from previous release attempts

## [0.9.14] - 2025-05-07

### Fixed
- Verified and reinforced fix for missing connection pool functions (issue #13)
  - Added comprehensive test suite to validate connection pool functions
  - Added verification script to confirm proper implementation
  - Ensured all required functions are properly defined and exported

### Added
- Comprehensive test coverage for connection pool functions
  - Added unit tests in pkg/core/connection_pool_test.go
  - Added verification script in scripts/verify_connection_pool_functions.go
  - Added test harness to simulate downstream project usage

## [0.9.13] - 2025-05-07

### Fixed
- Restored missing connection pool functions that were referenced in transport_init.go
  - Added missing SetConnectionPoolManager and EnableDefaultConnectionPool functions
  - Added GetConnectionPool and GetHTTPClientForService helper functions
  - Fixed breaking changes introduced in v0.9.11 that caused downstream projects to fail compilation

## [0.9.11] - 2025-05-07

### Fixed
- Fixed string formatting issues in example files
- Added missing ExpiresAt() method to TokenResponse in auth package
- Fixed client initialization patterns with proper error handling
- Fixed GitHub Actions workflow for API compatibility testing
- Updated API compatibility workflow to properly handle GitHub token authentication
- Fixed type references in integration tests

## [0.9.10] - 2025-05-07

### Fixed
- Fixed build error with undefined `httppool.NewHttpConnectionPoolManager` function
- Updated connection pool initialization to use the global pool manager

## [0.9.9] - 2025-05-07

### Added
- Comprehensive API compatibility testing suite
- Interface implementation verification tests
- Dependent project build test script
- Compiler-enforced API contracts using interfaces
- GitHub Actions workflow for API compatibility checks

### Changed
- Updated version to 0.9.9

## [0.9.8] - 2025-05-07

### Fixed
- Added GetVersionCheck() and SetVersionCheck() methods to Config in pkg/core/config/config.go
- Updated api_version.go to use GetVersionCheck() and SetVersionCheck() instead of direct field access
- Added SyncChecksum alias for SyncLevelChecksum in transfer package for backward compatibility
- Updated version to 0.9.8

## [0.9.7] - 2025-05-07

### Fixed
- Fixed mfaErr variable detection in auth/mfa.go
- Ensured VersionCheck field in Config struct is properly exported

## [0.9.6] - 2025-05-07

### Fixed
- Fixed duplicate tokenRequest method in auth/mfa.go
- Fixed type naming consistency with ClientConfig in transfer package
- Fixed incorrect DeleteItem structure in test and debug files
- Removed redundant Recursive field from DeleteItem that's unsupported by the API
- Fixed JSON marshaling issues with function fields in ResumableTransferOptions
- Added proper DataType setting for TransferItems in resumable transfers
- Fixed duplicate setupMockServer functions in transfer tests

## [0.9.5] - 2025-05-07

### Fixed
- Resolved import cycle issues between packages
- Restructured connection pool management to use interfaces
- Added additional pool configuration capabilities
- Created improved pool manager implementation

## [0.9.4] - 2025-05-07

### Fixed
- Added missing ClientInterface methods to Client type
- Fixed unused imports in client_with_pool.go
- Resolved interface implementation issues causing compilation errors in consuming applications

## [0.9.3] - 2025-05-07

### Fixed
- Added missing logging.go file in transport package that caused compilation errors
- Fixed "undefined: logRequest and logResponse" errors when using the SDK

## [0.9.2] - 2025-05-07

### Added
- Versioned documentation with Hugo-book theme
- GitHub Pages deployment workflows for documentation
- Comprehensive documentation for all API surfaces
- Enhanced GitHub Actions workflows with better CI/CD integration

### Fixed
- Documentation deployment issues
- Version compatibility checking in service clients
- GitHub Pages configuration
- Minor documentation formatting issues

## [0.9.1] - 2025-05-02

### Fixed
- Added missing interfaces package required by SDK consumers
- Fixed dependency issues when importing the SDK
- Added interface definitions for authorization, client operations, connection pools, and transport

## [0.9.0] - 2025-05-02

### Added
- Enhanced Compute service with workflow and task group capabilities
- Workflow management (creation, execution, status tracking)
- Dependency graph support for complex compute workflows
- Task group functionality for parallel execution
- Expanded container management capabilities
- Environment and secret management
- Improved API version compatibility checking
- Enhanced HTTP debugging with detailed request/response logging
- New example for Compute workflows and task groups

### Fixed
- Improved error handling in transport layer
- Enhanced connection pool management for better stability
- Fixed integration tests for all service clients
- Standardized error reporting formats across services
- Improved thread safety in concurrent operations

## [0.8.0] - 2025-03-15

### Added
- Compute service implementation
  - Batch job support
  - Container management
  - Dependency handling
  - Environment configuration
- Enhanced Auth package with options pattern
- Added Transport layer interfaces

### Changed
- Updated client implementation with connection pooling
- Improved error handling
- Enhanced logging with context-based logging

### Fixed
- Token refresh handling
- Race conditions in transport layer
- Authentication error handling

## [0.7.0] - 2025-01-30

### Added
- Flows service implementation
  - Flow management
  - Execution control
  - Status monitoring
- Search service implementation
  - Advanced query capabilities
  - Indexing operations
  - Result pagination
- Timers service implementation

### Changed
- Refactored Transfer service for better performance
- Improved error types and handling
- Enhanced documentation

### Fixed
- Memory leaks in Transfer operations
- Authentication token handling bugs

## [0.6.0] - 2024-12-05

### Added
- Groups service implementation
  - Group management (create, list, update, delete)
  - Membership management (add, remove, update roles)
  - Role management operations
- Transfer service implementation
  - File and directory operations
  - Task management
  - Status monitoring
- Auth service implementation
  - OAuth flow implementations
  - Token management

### Changed
- Improved SDK configuration options
- Enhanced error handling

### Fixed
- Connection handling in HTTP client
- Error propagation issues

## [0.5.0] - 2024-10-15

### Added
- Initial SDK framework
- Core client implementation
- Configuration management
- HTTP transport layer
- Basic authorization mechanisms

[Unreleased]: https://github.com/scttfrdmn/globus-go-sdk/compare/v3.60.0...HEAD
[3.60.0]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.15...v3.60.0
[0.9.15]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.14...v0.9.15
[0.9.14]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.13...v0.9.14
[0.9.13]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.12...v0.9.13
[0.9.12]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.11...v0.9.12
[0.9.11]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.10...v0.9.11
[0.9.10]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.9...v0.9.10
[0.9.9]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.8...v0.9.9
[0.9.8]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.7...v0.9.8
[0.9.7]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.6...v0.9.7
[0.9.6]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.5...v0.9.6
[0.9.5]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.4...v0.9.5
[0.9.4]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.3...v0.9.4
[0.9.3]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.2...v0.9.3
[0.9.2]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.1...v0.9.2
[0.9.1]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.9.0...v0.9.1
[0.9.0]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.8.0...v0.9.0
[0.8.0]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.7.0...v0.8.0
[0.7.0]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/scttfrdmn/globus-go-sdk/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/scttfrdmn/globus-go-sdk/releases/tag/v0.5.0