// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNCRequestWithContext(t *testing.T) {
	for _, tc := range []struct {
		name    string
		url     string
		mustErr bool
	}{
		{"normal", "repos/carabiner-dev/github/releases", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			nc, err := NewNativeHTTPCaller(&Options{
				Host:        "api.github.com",
				Token:       "",
				EnsureToken: false,
			})
			require.NoError(t, err)

			res, err := nc.RequestWithContext(context.Background(), http.MethodGet, tc.url, nil)
			require.NoError(t, err)
			defer res.Body.Close()
			require.NotNil(t, res)
			require.Equal(t, 200, res.StatusCode)
		})
	}
}
