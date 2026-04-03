// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors

// Package gcs provides an EXPERIMENTAL client for the Globus Connect Server (GCS)
// Collections API, along with a downloader utility for HTTPS-based file access
// and JSON:API pagination helpers.
//
// # STABILITY: EXPERIMENTAL
//
// APIs in this package may change or be removed without notice in any release.
// This package tracks Python SDK v4.5.0 experimental features:
//   - GCSCollectionClient — client for the GCS manager /api/collections endpoint
//   - GCSDownloader — helper for downloading files from HTTPS-enabled collections
//   - JSON:API pagination — page/pager types for JSON:API response iteration
//
// # Typical usage
//
//	client, err := gcs.NewCollectionClient(ctx, "https://g-xxxxx.data.globus.org", config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	collection, err := client.GetCollection(ctx, collectionID)
//
//	downloader := gcs.NewDownloader(client)
//	defer downloader.Close()
//
//	data, err := downloader.ReadFile(ctx, "https://g-xxxxx.data.globus.org/path/to/file.txt")
package gcs
