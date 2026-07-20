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

## v3 Transfer: DATA envelopes, wire-param realignment, added families (Phase 2 audit vs 3.65.0)

- **Uppercase `DATA` list envelopes.** `EndpointList`, `TaskList` and `FileList`
  decoded items under `json:"data"`, but Transfer returns them under uppercase
  `DATA` (upstream iterable.py `default_iter_key=DATA`) — every list previously
  deserialized empty. Tags corrected to `DATA`. `FileList` also uses `endpoint`
  (not `endpoint_id`) and gained `total`. `TaskList` gained `total`/`offset`/
  `limit`; `FileListItem.DATA_TYPE` corrected.
- **endpoint_search / task_list / operation_ls query params.** `ListEndpoints`
  no longer sends `page_size`/`page_token` (not endpoint_search wire params) and
  now supports `filter_non_functional` (encoded 1/0) and `filter_entity_type`.
  `ListTasks` uses the 3.65.0 combined `filter` (key:v1,v2/key2:v3) + `orderby`
  form; the individual `filter_*`/`page_size`/`page_token` option fields are
  retained as **divergent Go aliases that are not sent**. `ListFiles`/
  `ListDirectory` drop `excluded_types`/`continue_from`/`marker` (not operation_ls
  params) and add `offset` + `local_user`; `show_hidden` encodes 1/0.
- **preserve_timestamp.** `TransferTaskRequest.PreserveMtime` marshaled the
  invalid key `preserve_mtime`; upstream TransferData writes `preserve_timestamp`.
  JSON key corrected (Go field name kept for source compatibility).
- **Checksum fields.** `TransferItem.Checksum` (`json:"checksum"`) was a phantom;
  replaced with `ExternalChecksum` (`external_checksum`) + `ChecksumAlgorithm`
  (`checksum_algorithm`) per upstream `add_item`.
- **Delete / transfer top-level fields.** `DeleteTaskRequest` gained top-level
  `recursive`/`ignore_missing`/`interpret_globs`/`local_user` (recursive is
  top-level, not per delete_item). `TransferTaskRequest` gained
  `source_local_user`/`destination_local_user`/`filter_rules`.
- **Subscription methods split.** The old `SetSubscriptionAdminVerified(endpoint,
  subscriptionID)` PUT `endpoint/{id}/subscription` with a bogus
  `DATA_TYPE:subscription_id_update` body actually implemented upstream
  `set_subscription_id`. Split into `SetSubscriptionID(collectionID,
  subscriptionID)` (PUT `endpoint/{id}/subscription`, body `{subscription_id}`,
  no DATA_TYPE) and `SetSubscriptionAdminVerified(collectionID, verified bool)`
  (PUT `endpoint/{id}/subscription_admin_verified`, body
  `{subscription_admin_verified}`).
- **`Mkdir`/`Rename` signatures** gained a trailing `*MkdirOptions`/`*RenameOptions`
  for the optional `local_user` body field (pass `nil` for none).
- **Added families** (all `DATA`-enveloped unless noted): endpoint
  update/delete, `CreateSharedEndpoint`; bookmarks CRUD; ACL rule CRUD; role
  CRUD; server list/get; `OperationStat`; task `event_list`/`pause_info`/
  `successful_transfers`/`skipped_errors` (marker paged) and `UpdateTask`;
  `my_effective_pause_rule_list`/`my_shared_endpoint_list`/`shared_endpoint_list`
  (next_token paged, items under `shared_endpoints`); and the full
  `endpoint_manager` surface (monitored/hosted endpoints, task list/get/events/
  pause_info/successful_transfers/skipped_errors, admin cancel/pause/resume,
  pause-rule CRUD). Slice filters comma-joined; `filter_is_error` encodes 1/0.
- **Removed phantom Streams/tunnels.** `streams.go` (`CreateTunnel`/`GetTunnel`/
  `UpdateTunnel`/`DeleteTunnel`/`ListTunnels`/`GetStreamAccessPoint`/
  `GetTunnelEvents`) hit `tunnel*` routes that do not exist at 3.65.0 (Streams is
  a v4.3.0+ addition); removed along with the `streams-tunnels` example.
- **Intentionally omitted:** deprecated GCSv4 activation/server-mutation/symlink
  routes remain omitted (GCSv5 auto-activates); `GetSubmissionID` still reads both
  `value` and `submission_id` (upstream uses `value`; harmless).

## v3 Compute: passthrough documents, host-root base, folded V2/V3 (Phase 2 audit vs 3.65.0)

- **No request/response models, no pagination.** Upstream `ComputeClient` (at
  3.65.0) sends and returns open-ended JSON documents and defines no model
  classes or paginators for the compute web service. The Go client mirrors this
  with `map[string]interface{}` bodies and results everywhere; `models.go` is a
  comment-only file. The previously-defined `ComputeEndpoint`,
  `ComputeEndpointList`, container/environment/dependency/batch model types and
  their `ListEndpoints`/`ListEndpointsOptions` surface were fabricated and have
  been removed.
- **Base URL → host root** (`https://compute.api.globus.org/`); the `/v2` and
  `/v3` prefixes live in each endpoint path rather than the base.
- **V2 and V3 folded into one client.** V3 endpoint/function/submit routes are
  exposed as methods with a `V3` suffix (`RegisterEndpointV3`, `UpdateEndpointV3`,
  `LockEndpointV3`, `GetEndpointAllowlistV3`, `RegisterFunctionV3`, `SubmitV3`)
  alongside their V2 counterparts.
- **Removed phantom subsystems.** Container, environment, dependency and
  batch-builder helpers (and their examples) modeled a client-side task-packing
  surface that upstream does not expose over HTTP; task batches are submitted as
  passthrough documents to `POST /v2/submit` (or `POST /v3/endpoints/{id}/submit`).

## v3 Auth: removed phantom MFA REST routes; added resource surface (Phase 2 audit vs 3.65.0)

- **Removed phantom MFA routes.** `GetMFAChallenge` (`/oauth2/mfa/challenge/{id}`)
  and `RespondToMFAChallenge` (`/oauth2/mfa/response`) do not exist at 3.65.0. MFA
  is completed by resubmitting to `POST /v2/oauth2/token` with the MFA response
  fields; `ExchangeAuthorizationCodeWithMFA`/`RefreshTokenWithMFA` now do that via
  `resubmitWithMFA`, and `CheckForMFARequired` surfaces the challenge ID from the
  token-endpoint error rather than fetching challenge details.
- **`Identity.id`** — the primary key JSON tag is `id` (was the invalid
  `identity_id`).
- **Added the resource surface** (all responses use the upstream envelopes —
  single objects under `project`/`policy`/`client`/`scope`, collections under
  the plural key; mutually-exclusive query params comma-joined): `GetIdentities`,
  `GetIdentityProviders`; projects CRUD (`GetProjects`/`GetProject`/
  `CreateProject`/`UpdateProject`/`DeleteProject`); policies CRUD; clients CRUD
  (+`CreateChildClient`, `CreateNativeAppInstance`); client-credentials
  (`GetClientCredentials`/`CreateClientCredential`/`DeleteClientCredential`);
  scopes CRUD; `GetConsents`; `OAuth2GetDependentTokens` (top-level array
  response); `OAuth2ValidateToken`; and OIDC `GetOpenIDConfiguration`/`GetJWK`
  (host-root, outside the `/v2` base) plus `Userinfo`.

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
