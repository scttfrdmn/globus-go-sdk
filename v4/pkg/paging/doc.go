// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package paging provides generic pagination over Globus API list endpoints,
// mirroring the Python SDK's globus_sdk.paging module.
//
// # Available Paginators
//
//   - LimitOffsetPaginator: uses limit + offset query parameters (most services)
//   - MarkerPaginator: uses a cursor/marker string (Transfer tunnels)
//   - NextTokenPaginator: uses a next-page token (Groups)
//   - JSONAPIPaginator: follows JSON:API Links.Next URLs (GCS)
//
// # Basic usage
//
//	pager := client.NewFlowsPager(nil)
//	for pager.HasNext() {
//	    flows, err := pager.NextPage(ctx)
//	    if err != nil {
//	        return err
//	    }
//	    for _, f := range flows {
//	        fmt.Println(f.Title)
//	    }
//	}
//
// # Stability: BETA
//
// The API of this package may have minor changes in minor releases.
package paging
