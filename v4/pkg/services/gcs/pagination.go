// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025-2026 Scott Friedman and Project Contributors
package gcs

import (
	"context"

	"github.com/scttfrdmn/globus-go-sdk/v4/pkg/paging"
)

// NewCollectionPager returns a marker Paginator over all collections matching
// options. GCS uses top-level has_next_page + marker pagination (not JSON:API
// links). Pass nil for default options.
func (c *CollectionClient) NewCollectionPager(options *ListCollectionsOptions) paging.Paginator[Collection] {
	pageSize := 0
	if options != nil && options.PageSize > 0 {
		pageSize = options.PageSize
	}
	return paging.NewMarkerPaginator(
		func(ctx context.Context, limit int, marker string) ([]Collection, bool, string, error) {
			o := &ListCollectionsOptions{Marker: marker, PageSize: limit}
			if options != nil {
				o.MappedCollectionID = options.MappedCollectionID
				o.Filter = options.Filter
				o.Include = options.Include
			}
			result, err := c.ListCollections(ctx, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.Data, result.HasNextPage, result.Marker, nil
		},
		pageSize,
	)
}

// NewStorageGatewayPager returns a marker Paginator over all storage gateways
// matching options. Pass nil for default options.
func (c *CollectionClient) NewStorageGatewayPager(options *StorageGatewayListOptions) paging.Paginator[StorageGateway] {
	pageSize := 0
	if options != nil && options.PageSize > 0 {
		pageSize = options.PageSize
	}
	return paging.NewMarkerPaginator(
		func(ctx context.Context, limit int, marker string) ([]StorageGateway, bool, string, error) {
			o := &StorageGatewayListOptions{Marker: marker, PageSize: limit}
			if options != nil {
				o.Include = options.Include
			}
			result, err := c.GetStorageGatewayList(ctx, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.Data, result.HasNextPage, result.Marker, nil
		},
		pageSize,
	)
}

// NewRolePager returns a marker Paginator over all roles matching options. Pass
// nil for default options.
func (c *CollectionClient) NewRolePager(options *RoleListOptions) paging.Paginator[GCSRole] {
	pageSize := 0
	if options != nil && options.PageSize > 0 {
		pageSize = options.PageSize
	}
	return paging.NewMarkerPaginator(
		func(ctx context.Context, limit int, marker string) ([]GCSRole, bool, string, error) {
			o := &RoleListOptions{Marker: marker, PageSize: limit}
			if options != nil {
				o.CollectionID = options.CollectionID
				o.Include = options.Include
			}
			result, err := c.GetRoleList(ctx, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.Data, result.HasNextPage, result.Marker, nil
		},
		pageSize,
	)
}

// NewUserCredentialPager returns a marker Paginator over all user credentials
// matching options. Pass nil for default options.
func (c *CollectionClient) NewUserCredentialPager(options *UserCredentialListOptions) paging.Paginator[UserCredential] {
	pageSize := 0
	if options != nil && options.PageSize > 0 {
		pageSize = options.PageSize
	}
	return paging.NewMarkerPaginator(
		func(ctx context.Context, limit int, marker string) ([]UserCredential, bool, string, error) {
			o := &UserCredentialListOptions{Marker: marker, PageSize: limit}
			if options != nil {
				o.StorageGateway = options.StorageGateway
			}
			result, err := c.GetUserCredentialList(ctx, o)
			if err != nil {
				return nil, false, "", err
			}
			return result.Data, result.HasNextPage, result.Marker, nil
		},
		pageSize,
	)
}
