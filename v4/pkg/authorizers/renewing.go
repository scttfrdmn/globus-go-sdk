// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package authorizers

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// renewingAuthorizer is an unexported base for authorizers that automatically
// refresh tokens before they expire. Concrete types embed this and implement
// fetchNewTokens to perform the actual credential refresh.
type renewingAuthorizer struct {
	mu          sync.Mutex
	accessToken string
	expiresAt   time.Time

	// fetchNewTokens is called when the token is close to expiry.
	// Implementors must update accessToken and expiresAt under the mutex.
	fetchNewTokens func(ctx context.Context) error
}

// refreshThreshold is how far before expiry we proactively refresh.
const refreshThreshold = 60 * time.Second

// ensureValidToken checks whether the token needs refreshing and fetches new
// credentials if necessary. It is safe to call concurrently.
func (r *renewingAuthorizer) ensureValidToken(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Refresh if the token is missing or within the refresh threshold.
	if r.accessToken == "" || time.Now().Add(refreshThreshold).After(r.expiresAt) {
		return r.fetchNewTokens(ctx)
	}
	return nil
}

// getAuthorizationHeader returns a Bearer header after ensuring token validity.
func (r *renewingAuthorizer) getAuthorizationHeader(ctx context.Context) (string, error) {
	if err := r.ensureValidToken(ctx); err != nil {
		return "", err
	}
	r.mu.Lock()
	token := r.accessToken
	r.mu.Unlock()
	return fmt.Sprintf("Bearer %s", token), nil
}

// setToken updates the stored token and expiry. Must be called with r.mu held.
func (r *renewingAuthorizer) setToken(accessToken string, expiresAt time.Time) {
	r.accessToken = accessToken
	r.expiresAt = expiresAt
}
