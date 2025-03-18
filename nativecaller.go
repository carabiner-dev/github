// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

const (
	apiVersion      = "2022-11-28"
	defaultHostname = "api.github.com"
)

func NewNativeHTTPCaller(opts *Options) (*NativeHTTPCaller, error) {
	var client *http.Client = http.DefaultClient
	if opts.Token != "" {
		client = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: opts.Token},
		))
	}
	host := defaultHostname
	if opts.Host != "" {
		host = opts.Host
	}
	return &NativeHTTPCaller{
		client:   client,
		Hostname: host,
	}, nil
}

// NativeHTTPClient implements the caller interface with minimal dependencies
type NativeHTTPCaller struct {
	client   *http.Client
	Hostname string
}

// RequestWithContext makes a request to the server without auth
func (nc *NativeHTTPCaller) RequestWithContext(
	ctx context.Context, method string, endpoint string, r io.Reader) (*http.Response, error) {

	req, err := http.NewRequestWithContext(ctx, method, "https://"+nc.Hostname+"/"+strings.TrimPrefix(endpoint, "/"), r)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", apiVersion)

	return nc.client.Do(req)
}
