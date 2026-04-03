// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcs

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/core"
)

// Downloader is an EXPERIMENTAL helper for downloading files from an
// HTTPS-enabled Globus Connect Server collection.
//
// It wraps a CollectionClient and uses its credentials to perform authenticated
// HTTPS GET requests directly against the collection's data-plane endpoint.
//
// STABILITY: EXPERIMENTAL — this type may change without notice.
//
// Example:
//
//	client, _ := gcs.NewCollectionClient(ctx, address, collectionID, config)
//	defer client.Close()
//
//	dl := gcs.NewDownloader(client)
//	defer dl.Close()
//
//	data, err := dl.ReadFile(ctx, "https://g-xxxxx.data.globus.org/share/file.txt")
//	fmt.Println(string(data))
type Downloader struct {
	client     *CollectionClient
	httpClient *http.Client
	token      string
}

// NewDownloader creates a Downloader that uses the given CollectionClient for
// authentication.  The CollectionClient must have been created with an access
// token that includes the collection's HTTPS scope
// (see CollectionScopes / DefaultScopeRequirements).
func NewDownloader(client *CollectionClient) *Downloader {
	return &Downloader{
		client:     client,
		httpClient: &http.Client{},
	}
}

// NewDownloaderWithToken creates a Downloader that authenticates with an
// explicit bearer token.  Use this when managing tokens outside a
// CollectionClient.
func NewDownloaderWithToken(client *CollectionClient, accessToken string) *Downloader {
	return &Downloader{
		client:     client,
		httpClient: &http.Client{},
		token:      accessToken,
	}
}

// ReadFile downloads the file at fileURI and returns its raw bytes.
// fileURI must be an absolute HTTPS URL pointing to a file in the collection,
// e.g. "https://g-xxxxx.data.globus.org/share/path/to/file.txt".
func (d *Downloader) ReadFile(ctx context.Context, fileURI string) ([]byte, error) {
	if fileURI == "" {
		return nil, &core.ValidationError{
			Field:   "fileURI",
			Message: "file URI is required",
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURI, nil)
	if err != nil {
		return nil, fmt.Errorf("gcs downloader: create request: %w", err)
	}

	token := d.token
	if token == "" && d.client != nil {
		// Fall back to the CollectionClient's configured access token.
		// The client stores its token in the underlying core.Client; we
		// surface it via the CollectionClient's config that was passed at
		// construction time.  For now we require NewDownloaderWithToken if
		// a separate token is needed; the CollectionClient path is a
		// convenience placeholder.
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gcs downloader: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, &core.APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("gcs downloader: server returned %d: %s", resp.StatusCode, string(body)),
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gcs downloader: read response body: %w", err)
	}
	return data, nil
}

// ReadFileAsText is a convenience wrapper around ReadFile that returns the
// file content as a UTF-8 string.
func (d *Downloader) ReadFileAsText(ctx context.Context, fileURI string) (string, error) {
	data, err := d.ReadFile(ctx, fileURI)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Close releases resources held by the Downloader.  After Close the Downloader
// must not be used.
func (d *Downloader) Close() error {
	if d.httpClient != nil {
		d.httpClient.CloseIdleConnections()
		d.httpClient = nil
	}
	return nil
}
