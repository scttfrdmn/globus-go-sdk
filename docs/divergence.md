<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025-2026 Scott Friedman and Project Contributors -->

# Divergences from upstream Python globus-sdk

This SDK is a community Go port of the Python
[`globus/globus-sdk-python`](https://github.com/globus/globus-sdk-python),
which is the source of truth for API surface. Where the Go SDK deliberately
deviates from the upstream structure, the deviation is recorded here so the
choice is discoverable and intentional rather than accidental drift.

Wire-visible behavior (HTTP methods, URL paths, request/response JSON) always
matches upstream. Divergences are limited to how that surface is *organized*
into Go types and clients.

Sections below prefixed **v3:** apply to the frozen v3 module (tracking Python
globus-sdk 3.65.0); all other sections apply to the active v4 module.

## v3 Flows: base host, body envelope, removed action providers (Phase 2 audit vs 3.65.0)

The v3 flows client diverged from Python globus-sdk 3.65.0 on base URL, the run
body key, verbs, and a phantom action-provider surface. Corrected:

- **Base URL** `https://flows.automate.globus.org/` (was `flows.globus.org/v1/`).
- **RunFlow** posts to `POST /flows/{flow_id}/run` (was `POST /runs`) with the
  input under `body` (was `input`); `RunRequest.FlowID` is used in the URL, not
  the body. Added `ValidateRun` (`POST /flows/{id}/validate_run`).
- **Verbs:** `UpdateFlow`/`UpdateRun` use `PUT` (UpdateRun was `PATCH`).
  `DeleteRun` is `POST /runs/{id}/release`.
- **Removed the action-provider surface** (`ListActionProviders`,
  `GetActionProvider`, `ListActionRoles`, `GetActionRole`, their iterators, the
  `BatchGetActionRoles` batch, and the `ActionProvider*`/`ActionRole*` types) —
  no such routes exist in the globus-sdk flows service at 3.65.0.
- **Timestamp keys:** `RunResponse.started_at`→`start_time`,
  `completed_at`→`completion_time`; `RunLogEntry.created_at`→`time` (and its
  `run_id` field, which is not returned, was removed).
- **Added:** `ValidateFlow` (`POST /flows/validate`), `GetRunDefinition`,
  `DeleteRun`, `ResumeRun`.

## v3 Search: index-scoped routes, gmeta/index_list keys (Phase 2 audit vs 3.65.0)

The v3 search client posted to non-index-scoped routes and decoded the wrong
response keys. Corrected to match Python globus-sdk 3.65.0:

- **Index-scoped routes:** `IngestDocuments` → `POST /index/{id}/ingest`,
  `Search`/`StructuredSearch` → `POST /index/{id}/search`, `DeleteDocuments` →
  `POST /index/{id}/batch_delete_by_subject` (were `/ingest`, `/search`,
  `/delete`). `IndexID` moved out of the delete body into the path.
- **Response keys:** `IndexList` decodes `index_list` (was `indexes`);
  `SearchResponse` decodes results from `gmeta` and paginates via `has_next_page`
  + offset (was `results`/`page_token`); `DeleteDocumentsResponse` reads the
  top-level `task_id`; `TaskStatusResponse` gained `index_id`. The search
  iterator advances by offset.
- **Added methods:** `GetSearch` (GET), `Scroll`, `DeleteByQuery`, `GetSubject`/
  `DeleteSubject` and `GetEntry`/`DeleteEntry` (subject/entry_id as query params),
  `GetTaskList` (`GET /task_list/{index_id}`), and role ops (`CreateRole`,
  `GetRoleList`, `DeleteRole`).

## v3 Timers: realigned to the 3.65.0 wire (base URL, /jobs/, document shape)

The v3 timers client diverged from Python globus-sdk 3.65.0 on base URL, paths,
and document shape (same issues later seen in v4). Corrected:

- **Base URL** `https://timer.automate.globus.org/` (was `.../api/v1/`). Classic
  routes are under `/jobs/` and creation is `POST /v2/timer`.
- **Paths/verbs:** list `GET /jobs/`, get `GET /jobs/{id}`, create
  `POST /v2/timer` (body wrapped `{"timer": ...}`), legacy create `POST /jobs/`,
  update `PATCH /jobs/{id}` (was `PATCH /timers/{id}`), delete `DELETE /jobs/{id}`,
  pause/resume `POST /jobs/{id}/{pause,resume}` (resume takes optional
  `{"update_credentials": bool}`).
- **Removed phantom methods** with no upstream route: `RunTimer`, `ListRuns`,
  `GetRun`, `GetCurrentUser`, `CreateCronTimer`/`CreateFlowTimerCron` (no cron
  schedules at 3.65.0), and the `TimerRun`/`TimerRunList`/`RunResult`/`RunError`/
  `CurrentUserInfo`/`Callback`/`CreateTimerRequest`/`UpdateTimerRequest` types.
- **Create document** reshaped to `timer_type`/`name`/`schedule`/`body`
  (+`flow_id`); `Schedule` serializes the once/recurring shapes
  (`interval_seconds` int, structured `end`). Added `NewOnceSchedule`/
  `NewRecurringSchedule`/`NewTransferTimer`/`NewFlowTimer` builders, `CreateJob`,
  and reworked `CreateFlowTimer`/`CreateTransferTimer` helpers.
- `list_jobs` is not paginated upstream; `ListTimers` does a single fetch and
  `ListTimersOptions` is a query-param passthrough.

## v3 Groups: removed fabricated members/roles surface (Phase 2 audit vs 3.65.0)

The v3 groups client had the same fabrications later inherited by v4. Corrected
to match Python globus-sdk 3.65.0:

- **No members sub-resource / no roles resource.** Removed `ListMembers`/
  `AddMember`/`RemoveMember`/`UpdateMemberRole` (+ LowLevel variants), the roles
  CRUD (`ListRoles`/`GetRole`/`CreateRole`/`UpdateRole`/`DeleteRole`), the
  `ChangeRole`/`ChangeRoles`/`BatchMembershipActions` builder in batch.go, and
  the `Role`/`RoleCreate`/`RoleUpdate`/`RoleList`/`GroupSubscription` types.
  Membership data is read from the group via `GetGroup(..., include=memberships)`
  (`Group.Memberships`); mutations go through the new `BatchMembershipAction`
  (`POST /groups/{id}`). Role is a string (`member`/`manager`/`admin`).
- **No group-list route / no pagination.** Removed `ListGroups`/`ListGroupsV2`
  and the limit/offset/marker options. `GetMyGroups` now hits
  `GET /groups/my_groups`, returns a top-level `[]Group`, and sends `statuses`
  as one comma-joined param.
- **Corrected routes:** `GetGroupBySubscriptionID` → `GET /subscription_info/{id}`
  (was a `GET /groups?subscription_id=` query); preferences →
  `GET`/`PUT /preferences` (no group/identity path, passthrough map);
  `SetSubscriptionAdminVerified` → `PUT /groups/{id}/subscription_admin_verified`
  with `{"subscription_admin_verified_id": ...}` (nil → null); `UpdateGroup` uses
  `PUT`; `GetGroup` gained an `include` option.
- **`GroupPolicies`** reshaped to the real keys (`is_high_assurance`,
  `group_visibility`, `group_members_visibility`, `join_requests`,
  `signup_fields`, `authentication_assurance_timeout`); membership-fields use
  passthrough maps.

## Transfer: Streams/Tunnels folded into `transfer.Client`

- **Upstream:** the Streams/Tunnels API (Python SDK v4.3.0–v4.4.0) lives on the
  experimental `TransferClientV2`.
- **Here:** the methods (`CreateTunnel`, `GetTunnel`, `UpdateTunnel`,
  `DeleteTunnel`, `ListTunnels`, `GetStreamAccessPoint`, `ListStreamAccessPoints`,
  `GetTunnelEvents`) are added directly to the existing `transfer.Client`.
- **Why:** the Go v4 module has no `TransferClientV2`. A second client type would
  duplicate construction, config, and auth wiring for no user benefit.

## Transfer: bookmark CRUD folded into `transfer.Client`

- **Upstream:** bookmark management (Python SDK v4.6.0, amended v4.8.0) lives on
  the experimental `TransferClientV2` and speaks JSON:API under `/v2/bookmarks`.
- **Here:** `CreateBookmark`, `GetBookmark`, `ListBookmarks`, `UpdateBookmark`,
  and `DeleteBookmark` are added to the existing `transfer.Client`, matching the
  Streams/Tunnels approach above. The JSON:API envelope
  (`data.type`/`attributes`/`relationships`) is built and flattened internally;
  callers work with the flat `Bookmark`, `BookmarkCreate`, and `BookmarkUpdate`
  types.
- **Wire fidelity:** paths, HTTP verbs, and the JSON:API document shape match
  upstream exactly. Per upstream v4.8.0 the `pinned` field was removed, so it is
  absent from the Go models. Per upstream v4.8.1 `list_bookmarks` is **not**
  paginated (it returns the full set in a single response), so `ListBookmarks`
  does a single-page fetch and there is no bookmarks pager.
