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

## Corrected in the v4 parity audit (2026-07): core URL and OAuth2 encoding bugs

The Phase 2 parity audit (against Python globus-sdk tag 4.8.1) found two
cross-cutting `pkg/core` bugs that broke wire fidelity for *every* service. They
are recorded here because "4.5.0 parity" was previously claimed while these were
present:

- **Base path was discarded.** `core.Client.buildURL` did `u.Path = endpoint`,
  overwriting the base URL's path. Every service base URL carries a version
  prefix (auth `/v2`, transfer `/v0.10`, timers `/api/v1`, search `/v1`, …), so
  the prefix was silently dropped in production. Tests masked this by using a
  path-less `httptest` base URL. Fixed to join the endpoint onto the base path
  (`u.JoinPath`). Service call sites that had hardcoded the full prefix to work
  around this (auth's `/v2/api/projects`) were made relative to avoid
  double-prefixing.
- **OAuth2 bodies were JSON, not form-encoded.** `core.Client.DoRequest` always
  `json.Marshal`ed the body and sent `application/json`. The OAuth2
  token/introspect/revoke endpoints require `application/x-www-form-urlencoded`.
  A `url.Values` body (as auth passes) serialized to
  `{"grant_type":["authorization_code"]}` — wrong media type and array-valued
  fields. `DoRequest` now sends a `url.Values` body as flat
  `application/x-www-form-urlencoded`; all other bodies remain JSON.

## GCS: result envelope unpacking, marker pagination, full manager surface

The Phase 2 audit found the GCS client could not parse real responses (it decoded
flat objects and JSON:API pages; GCS wraps everything in a `result#1.0.0`
envelope with a `data` array). The client was rebuilt:

- **Envelope unpacking.** A generic `GCSResponse` type parses the envelope;
  single-object GETs unpack `data[0]` by matching the DATA_TYPE name (before
  `#`), and lists read the whole `data` array. Nearly every method uses this.
- **Marker pagination.** GCS returns top-level `has_next_page` + `marker` (not
  JSON:API `links.next`). The `CollectionPager` was rebuilt on
  `paging.MarkerPaginator`; the `JSONAPILinks`/`JSONAPIMeta`/`CollectionPage`
  types and `listCollectionsAbsolute` were removed. Added storage-gateway, role,
  and user-credential pagers.
- **`ListCollections` params fixed.** The fabricated `filter_owned`/`limit`/
  `offset` were removed; upstream uses `mapped_collection_id`, `filter`
  (comma-joined), `include` (comma-joined), `page_size`, `marker`.
- **Unauthenticated `/info`.** `GetGCSInfo` calls `GET /info` with no
  Authorization header, via the new `core.Client.DoRequestNoAuth`.
- **Full manager surface added:** endpoint (`GetEndpoint`/`UpdateEndpoint`),
  collection create, storage gateways CRUD, roles CRUD, and user credentials
  CRUD, each with its typed request `*Document` builder and response type.
  `UpdateStorageGateway` and the `Delete*` methods return the raw envelope (no
  data unpacking), matching upstream.
- **Known deferrals (recorded, not blockers):** connector-specific policy
  builder classes (POSIX/S3/…) are represented as `json.RawMessage` `policies`
  fields rather than typed helpers; the DATA_TYPE version auto-deduction that the
  Python SDK performs is left to the caller (DataType is omitempty, defaulting to
  the base version server-side). The Go-only `Downloader` has no upstream
  equivalent and is retained as an experimental convenience.

## Compute: passthrough documents, V2+V3 folded, phantom routes removed

Upstream Globus Compute (at 4.8.1) is only `__init__.py`, `client.py`, and
`errors.py` — it defines **no** request/response models and **no** pagination.
The Phase 2 audit realigned the Go client to that reality:

- **Passthrough everywhere.** Request bodies are `map[string]interface{}` and
  object responses are `map[string]interface{}`. The previously invented typed
  models (`Endpoint`, `EndpointList`, `FunctionRun`, `FunctionList`,
  `TaskStatus`, `TaskList`, `FunctionDefinition`, `FunctionRegistration`,
  `BatchTask*`, etc.) were removed — none were verifiable against the wire.
  Two endpoints do not return JSON objects, so their methods return their real
  shape: `GetEndpoints` (`GET /v2/endpoints`) returns a top-level array
  (`[]map[string]interface{}`), and `GetVersion` (`GET /v2/version`) returns an
  untyped value (a bare string with no `service`, or an object with one).
- **Base URL → host root.** `https://compute.api.globus.org` (was `.../v2`), and
  every endpoint carries its own `/v2` or `/v3` prefix, so both API surfaces are
  reachable through the path-joining buildURL.
- **V2 and V3 folded into one client.** Upstream ships `ComputeClientV2` and
  `ComputeClientV3` as separate classes; Go keeps one `compute.Client` and
  suffixes the v3 methods (`RegisterEndpointV3`, `UpdateEndpointV3`,
  `LockEndpointV3`, `GetEndpointAllowlistV3`, `RegisterFunctionV3`, `SubmitV3`).
- **Removed phantom methods** with no upstream route: `SubmitFunction`
  (POST /endpoints/{id}/functions), `CancelFunction`, `ListFunctions`,
  `UpdateFunction`, `ListTasks`, `CancelTask`, `RunBatch`, and both pagers.
  Task submission is `Submit` (`POST /v2/submit`) or `SubmitV3`; task listing by
  group is `GetTaskGroup`.
- **`GetEndpoints`** (was `ListEndpoints`) sends only the `role` param upstream
  supports — the fabricated `limit`/`offset` were removed. `GetTaskStatus` →
  `GetTask`, `GetBatchStatus` → `GetTaskBatch`. Added `GetVersion`,
  `GetResultAMQPURL`, `RegisterEndpoint`, `GetEndpointStatus`, `DeleteEndpoint`,
  `LockEndpoint`, `GetTaskGroup`, `RegisterFunction` (passthrough).

## Flows: SpecificFlowClient folded in; removed action providers; wire fixes

The Phase 2 audit corrected several Flows wire divergences and removed a phantom
surface:

- **One client.** Upstream splits `FlowsClient` and `SpecificFlowClient` (built
  with a flow_id). Go keeps one `flows.Client`: `RunFlow`/`ValidateRun` take a
  `flowID` arg; `ResumeRun` takes a `runID` (upstream `resume_run` posts to
  `/runs/{run_id}/resume`).
- **Removed action-provider methods** (`ListActionProviders`,
  `GetActionProvider`, `ListActionRoles`, `GetActionRole`) and their types — the
  globus-sdk flows service has no `/action_providers` routes at 4.8.1 (those live
  in the separate globus-automate-client).
- **`RunFlow`/`ValidateRun` body.** The flow input is sent under `body` (was
  `input`); added `label`/`tags`/`run_monitors`/`run_managers`/
  `activity_notification_policy` to `FlowInput`.
- **Verbs.** `UpdateFlow` and `UpdateRun` use `PUT` (were `PATCH`). `DeleteRun`
  is `POST /runs/{id}/release` (not an HTTP DELETE).
- **Removed `FlowAuthenticationPolicy`.** Upstream create/update flow accept only
  `authentication_policy_id` (a string), never an authentication-policy object.
  `FlowCreate`/`FlowUpdate` gained the full upstream field set (subtitle,
  flow_viewers/starters/administrators, run_managers/monitors, keywords,
  subscription_id, authentication_policy_id). `CreateFlow` now requires
  input_schema.
- **Pagination.** `list_flows`, `list_runs`, and `get_run_logs` are
  marker-paginated (keys `marker`/`has_next_page`), not limit/offset. Options
  reworked: flows use `filter_roles` (comma-joined), `filter_fulltext`,
  `orderby` (repeated params), `marker`; runs use `filter_flow_id`/`filter_roles`
  (comma-joined) + `marker`; run logs use `limit`, `reverse_order`, `marker`.
  `NewFlowsPager`/`NewRunsPager` switched to marker pagination and
  `NewRunLogsPager` added. `ListRegisteredAPIsOptions.OrderBy` is now `[]string`
  (repeated params).
- **New methods:** `ValidateFlow` (`POST /flows/validate`), `GetRunDefinition`,
  `DeleteRun`, `ResumeRun`, `ValidateRun`; `GetRun` gained a
  `*GetRunOptions{IncludeFlowDescription}` arg.

## Groups: removed fabricated members/roles surface and pagination

The Phase 2 audit found much of the Groups client hit routes that do not exist
in Globus Groups at 4.8.1. Corrected:

- **No members sub-resource.** The Groups API has no `/groups/{id}/members`
  routes. Removed `ListMembers`, `AddMember`, `RemoveMember`, `UpdateMemberRole`
  and the `MemberList`/`ListMembersOptions` types. Membership data is now read
  from the group document via `GetGroup(..., &GetGroupOptions{Include:
  []string{"memberships"}})` (exposed as `Group.Memberships`), and all
  membership mutations go through the new `BatchMembershipAction`
  (`POST /groups/{id}`) with a `BatchMembershipActions` document.
- **No roles resource.** Removed `ListRoles`/`GetRole`/`CreateRole`/
  `UpdateRole`/`DeleteRole` and the `Role`/`RoleCreate`/`RoleUpdate`/`RoleList`
  types. A membership's role is a string (`member`/`manager`/`admin`), not a
  role ID resource.
- **No group list / no pagination.** There is no `GET /groups` list route and
  the service registers no paginator. Removed `ListGroups`,
  `ListGroupsOptions`, and both `NewGroupsPager`/`NewMembersPager`. The only
  listing route is `GET /groups/my_groups` (`GetMyGroups`), which returns a
  top-level JSON array with `statuses` as a single comma-joined param; the
  fabricated `GroupList.has_next_page`/`next_page_token` fields were removed.
- **Account-level preferences / membership fields.** `GetIdentityPreferences`
  and `SetIdentityPreferences` now hit `GET`/`PUT /preferences` (no group or
  identity path params) with passthrough maps. `GetMembershipFields`/
  `SetMembershipFields` use passthrough maps instead of the fabricated
  `MembershipFields` wrapper.
- **`GroupPolicies`** reshaped to the real keys (`is_high_assurance`,
  `group_visibility`, `group_members_visibility`, `join_requests`,
  `signup_fields`, `authentication_assurance_timeout`); the fabricated
  `group_id`/`policies`/`is_high_risk_group`/`last_updated` keys were removed.
- Added `GetGroupBySubscriptionID` (`GET /subscription_info/{id}`) and
  `SetSubscriptionAdminVerified` (`PUT /groups/{id}/subscription_admin_verified`,
  nil ID sends JSON null).

## Auth: single client folds upstream's split clients; wire fixes

The Phase 2 audit filled a large missing surface and fixed wire bugs in the auth
client. Notable structure choices vs Python globus-sdk 4.8.1:

- **One client, many upstream clients.** Python splits auth across `AuthClient`
  (identities, projects, policies, clients, scopes, credentials, consents) and
  the login clients (`AuthLoginClient` / `ConfidentialAppAuthClient` /
  `NativeAppAuthClient`) for oauth2 grants, introspect/revoke, dependent tokens,
  and child/native-app client creation. The Go SDK folds all of these onto one
  `auth.Client`, so management and grant methods coexist.
- **Go-only extensions kept.** `StartDeviceAuthorization` /
  `PollDeviceAuthorization` / `WaitForDeviceAuthorization` (RFC 8628 device flow,
  `POST /oauth2/device/code`) and `GetAuthorizationURL` have no method on the
  4.8.1 Python auth client, but hit real Globus Auth wire endpoints; they are
  retained as a Go superset and are not phantom routes.
- **Form encoding.** `IntrospectToken` and `RevokeToken` now send
  `application/x-www-form-urlencoded` (they previously sent JSON, the wrong media
  type). `IntrospectToken` gained an `include` option; grant helpers
  (`ClientCredentialsTokens`, `GetDependentTokens`) are form-encoded.
- **Response envelopes.** Single objects are unwrapped from their key
  (`project`, `policy`, `client`, `scope`, `credential`) and collections from
  the plural key (`projects`, `identities`, `identity_providers`, `policies`,
  `clients`, `credentials`, `scopes`, `consents`). `GetProjects`/`GetProject`
  previously ignored the envelope and returned empty data.
- **`ProjectCreate` trimmed** to the four fields upstream accepts
  (`display_name`, `contact_email`, `admin_ids`, `admin_group_ids`); the create
  body is wrapped under `project`. The fabricated `public_contact_info` /
  `metadata` create fields were removed.
- **Host-root endpoints.** `GetOpenIDConfiguration` and `GetJWK` target the Auth
  host root (`/.well-known/openid-configuration`, then the advertised
  `jwks_uri`), outside the `/v2` base. They use the new
  `core.Client.DoRequestURL` to bypass the base-path join. The Python `as_pem`
  JWK decoding is client-internal crypto and out of wire scope.
- **`AuthClientInfo`** names the `/api/clients` resource type to avoid colliding
  with the `auth.Client` service type. `Consent.ID` and `dependency_path` are
  integers, not UUID strings.

## Search: subject/entry query params, ingest envelope, corrected verbs

The Phase 2 audit found several Search routes and bodies diverged from 4.8.1:

- **Subject/entry are query params.** `GetEntry`/`DeleteEntry` now hit
  `/index/{id}/entry?subject=&entry_id=` (entry_id optional), and new
  `GetSubject`/`DeleteSubject` hit `/index/{id}/subject?subject=`. The old
  `/index/{id}/subject/{subject}` path segments were not upstream routes.
- **`UpdateIndex`** uses `PATCH` (was `PUT`). `CreateIndex`/`IndexUpdate` bodies
  trimmed to `display_name`/`description` (upstream sends only those).
- **Ingest.** Upstream has one `ingest` method taking a
  `{ingest_type, ingest_data}` document. The Go `IngestEntry`/`IngestBatch`
  methods sent bodies without that envelope (broken wire) — replaced by a single
  `Ingest(indexID, data)` plus `NewGMetaEntryIngest`/`NewGMetaListIngest`
  builders. Added `DeleteByQuery` and `BatchDeleteBySubject`.
- **Tasks.** `GetTaskStatus` (mis-routed to `/index/{id}/task/{id}`) replaced by
  `GetTask` (`GET /task/{task_id}`, index-independent) and `GetTaskList`
  (`GET /task_list/{index_id}`). The task status field is `state`, not `status`.
- **`AddRole`** sends `{role_name, principal}` (was the fabricated
  `{principal, role_id}`); `GetRole` removed (no single-role GET route).
- **`SearchQuery`** now emits the required `@version: "query#1.0.0"`, `facets`
  is a list of objects (was `[]string`), and `post_facet_filters`/`boosts` were
  added; the non-upstream `bypass_visible_to` field was dropped. Added the GET
  `SearchGet` variant and `Scroll` (marker-paginated).
- **Pagination.** `index_list` is not paginated upstream — `NewIndexesPager` and
  the `limit`/`offset` params were removed; `filter_roles` is a single
  comma-joined value (was repeated params). Added `NewSearchPager`
  (offset-advancing, capped at 10000) and `NewScrollPager` (marker) matching the
  upstream paginators over the `gmeta` key.

## Timers: realigned to the 4.8.1 wire (paths, base URL, document shape)

The Phase 2 audit found the timers client diverged from upstream on nearly every
wire detail. Corrected to match Python globus-sdk 4.8.1's `TimersClient`:

- **Base URL** is `https://timer.automate.globus.org` (was `.../api/v1`). Upstream
  paths are absolute from the host root (`/jobs/`, `/v2/timer`).
- **Paths/verbs:** list `GET /jobs/`, get `GET /jobs/{id}`, create `POST /v2/timer`
  (body wrapped as `{"timer": ...}`), legacy create `POST /jobs/`, update
  `PATCH /jobs/{id}` (was `PUT /timers/{id}`), delete `DELETE /jobs/{id}`, pause
  `POST /jobs/{id}/pause`, resume `POST /jobs/{id}/resume` (optional
  `{"update_credentials": bool}` body).
- **Create document** reshaped to `timer_type` + `name` + `schedule` + `body`
  (+`flow_id` for flow timers), with `Schedule` serializing to the upstream
  once/recurring shapes (`interval_seconds` int, structured `end`). The old
  `Callback`/`Timer.Schedule.Interval`-as-string shape was fabricated.
- **Removed phantom methods** with no upstream classic-client route: `RunTimer`,
  `ListRuns`, `GetRun` (the classic `TimersClient` has no run-inspection surface),
  and the `CreateOnceTimer`/`CreateRecurringTimer` helpers (superseded by the
  `NewOnceSchedule`/`NewRecurringSchedule` builders).
- **`list_jobs`** is not paginated upstream; `ListTimers` does a single fetch and
  `ListTimersOptions` is a generic query-param passthrough.

## Transfer: bare-host base URL + full classic surface (Phase 2 audit)

The Phase 2 audit found the transfer client implemented ~15 methods (several
against non-existent routes) out of ~62 upstream. Corrected:

- **Base URL → bare host** `https://transfer.api.globus.org` (was `.../v0.10`).
  Classic routes now carry their own `/v0.10` prefix and the Beta tunnel/stream
  routes carry `/v2`, so both surfaces are reachable through the path-joining
  buildURL. (Before this, the `/v2/...` tunnel and bookmark literals were
  mis-routed to `.../v0.10/v2/...`.)
- **Removed phantom routes:** `ListEndpoints` hit `/endpoint_list` (no such
  route) — replaced by `EndpointSearch` (`GET /v0.10/endpoint_search`, offset
  paginated, capped at 1000 via `NewEndpointSearchPager`). The flat
  `/tunnel_list`/`/stream_access_point_list` list paths and `NewTunnelsPager`
  were replaced by the real JSON:API `/v2/tunnels` and `/v2/stream_access_points`
  (neither is marker-paginated upstream).
- **Submit auto-fetches `submission_id`:** `SubmitTransfer`/`SubmitDelete` now
  fetch a submission ID when the caller supplies none, matching upstream; added
  `submission_id`, `source_local_user`/`destination_local_user`, `local_user`,
  and `filter_rules` to the transfer/delete documents.
- **Added the missing classic surface:** endpoint update/delete/subscription,
  `operation_stat`, `local_user` on ls/mkdir/rename, `orderby` on ls/task_list,
  task `event_list`/`pause_info`/`successful_transfers`/`skipped_errors`,
  `update_task`, shared-endpoint family, endpoint role/ACL/server families, and
  the entire `endpoint_manager` family (monitored endpoints, task inspection,
  admin cancel/pause/resume, pause-rule CRUD). Open-ended admin documents use
  passthrough `map[string]interface{}` (`GenericResponse`).
- **`filter_status`/`orderby`/`filter_task_id` are comma-joined** into single
  params (were repeated); integer-boolean params (`show_hidden`,
  `filter_non_functional`, `filter_is_error`) serialize as `1`/`0`.

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
  with `map[string]interface{}` bodies and object results; `models.go` is a
  comment-only file. The previously-defined `ComputeEndpoint`,
  `ComputeEndpointList`, container/environment/dependency/batch model types and
  their `ListEndpoints`/`ListEndpointsOptions` surface were fabricated and have
  been removed. Two endpoints do not return JSON objects, so their methods return
  their real shape: `GetEndpoints` (`GET /v2/endpoints`) returns a top-level
  array (`[]map[string]interface{}`), and `GetVersion` (`GET /v2/version`)
  returns an untyped value (a bare string with no `service`, or an object).
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
  experimental `TransferClientV2`, speaking JSON:API under `/v2/`.
- **Here:** the methods (`CreateTunnel`, `GetTunnel`, `UpdateTunnel`,
  `DeleteTunnel`, `ListTunnels`, `GetStreamAccessPoint`, `ListStreamAccessPoints`,
  `GetTunnelEvents`) are on the existing `transfer.Client`. The Phase 2 audit
  corrected their wire shape: requests build the JSON:API document
  (`data.{type,relationships,attributes}` — e.g. tunnel create references the
  listener/initiator `StreamAccessPoint` ids under relationships), and responses
  are flattened from the JSON:API envelope into the flat `Tunnel`/
  `StreamAccessPoint` types. `UpdateTunnel` uses `PATCH`. Phase 1's flat
  request/response models were invented and have been replaced.
- **Why folded:** the Go v4 module has no `TransferClientV2`. A second client
  type would duplicate construction, config, and auth wiring for no user benefit.

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
