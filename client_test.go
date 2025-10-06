// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenScopes(t *testing.T) {
	t.Parallel()
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip()
	}

	c, err := NewClient()
	require.NoError(t, err)
	scopes, err := c.TokenScopes()
	require.NoError(t, err)
	fmt.Printf("%+v\n", scopes)
	require.NotNil(t, scopes)
	t.Error()
}

func TestCall(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name     string
		data     []byte
		synthErr error
	}{
		{"normal", []byte("Hi\n"), nil},
		{"normal", nil, errors.New("Superbad HTTP Error!")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Initialize the caller with the file caller to test
			c, err := NewClient(
				WithCaller(&FileCaller{
					SourcePath: "testdata/file.txt",
					Error:      tc.synthErr,
				}),
			)
			require.NoError(t, err)

			res, err := c.Call(t.Context(), http.MethodGet, "http://example.com/", nil)
			if tc.synthErr != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			defer res.Body.Close() //nolint:errcheck

			data, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, tc.data, data)
		})
	}
}
