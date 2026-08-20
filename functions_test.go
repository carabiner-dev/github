// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepoFromString(t *testing.T) {
	t.Parallel()
	const cosignRepoURL = "https://github.com/sigstore/cosign"
	for _, tc := range []struct {
		name    string
		sut     string
		expect  string
		mustErr bool
	}{
		{"reposlug", "sigstore/cosign", cosignRepoURL, false},
		{"noscheme", "github.com/sigstore/cosign", cosignRepoURL, false},
		{"norepo", "github.com/sigstore", "", true},
		{"locator", "git+https://github.com/sigstore/cosign@main", cosignRepoURL, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewClient()
			require.NoError(t, err)
			res, err := RepoFromString(tc.sut)
			if tc.mustErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expect, res)
		})
	}
}
