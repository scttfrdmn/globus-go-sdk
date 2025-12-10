// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package core

import (
	"net/http"
	"testing"
)

// TestClientClose tests the Close method
func TestClientClose(t *testing.T) {
	tests := []struct {
		name              string
		provideHTTPClient bool
		wantErr           bool
	}{
		{
			name:              "Close with internally created HTTP client",
			provideHTTPClient: false,
			wantErr:           false,
		},
		{
			name:              "Close with user-provided HTTP client",
			provideHTTPClient: true,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				AccessToken: "test-token",
				Scopes:      []string{"test-scope"},
			}

			if tt.provideHTTPClient {
				config.HTTPClient = &http.Client{}
			}

			client, err := NewClient(config)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			// Close should not error
			err = client.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Calling Close multiple times should be safe
			err = client.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() second call error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestClientCloseNilConfig tests Close with nil config edge case
func TestClientCloseNilConfig(t *testing.T) {
	// This test ensures Close handles edge cases gracefully
	client := &Client{
		config:            &Config{},
		httpClientCreated: true,
	}

	err := client.Close()
	if err != nil {
		t.Errorf("Close() with nil HTTPClient error = %v, want nil", err)
	}
}
