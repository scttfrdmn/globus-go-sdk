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
