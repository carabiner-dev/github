// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import "github.com/cli/go-gh/v2/pkg/api"

const DefaultAPIHostname = "api.github.com"

func buildGithubRestClient(opts Options) (*api.RESTClient, error) {
	return api.NewRESTClient(api.ClientOptions{
		AuthToken: opts.Token,
		Host:      opts.Host,
	})
}
