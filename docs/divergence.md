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
