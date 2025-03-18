// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRepoFromString(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		sut     string
		expect  string
		mustErr bool
	}{
		{"reposlug", "sigstore/cosign", "https://github.com/sigstore/cosign", false},
		{"noscheme", "github.com/sigstore/cosign", "https://github.com/sigstore/cosign", false},
		{"norepo", "github.com/sigstore", "", true},
		{"locator", "git+https://github.com/sigstore/cosign@main", "https://github.com/sigstore/cosign", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
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
