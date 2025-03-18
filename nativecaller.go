// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	ctx context.Context, method string, endpoint string, r io.Reader,
) (*http.Response, error) {
	var sendWithoutAuth bool
	hostname := nc.Hostname
	if strings.HasPrefix(endpoint, "https://") || strings.HasPrefix(endpoint, "http://") {
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return nil, fmt.Errorf("parsing redirect URL: %w", err)
		}
		// If we're requesting now from a different host
		// (from a redirect), then we generate a new client
		// to not leak the token
		if parsed.Hostname() != nc.Hostname {
			sendWithoutAuth = true
			hostname = nc.Hostname
		}
		endpoint = strings.TrimPrefix(endpoint, parsed.Scheme+"://"+parsed.Host)
	}

	url := fmt.Sprintf("https://%s/%s", hostname, strings.TrimPrefix(endpoint, "/"))
	req, err := http.NewRequestWithContext(ctx, method, url, r)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", apiVersion)

	var resp *http.Response
	if sendWithoutAuth {
		resp, err = http.DefaultClient.Do(req)
	} else {
		resp, err = nc.client.Do(req)
	}
	if err != nil {
		return nil, err
	}

	//  If we got a redirect, then call again on the redirect target
	if resp.StatusCode == 302 {
		location := resp.Header.Get("location")
		if location == "" {
			return nil, fmt.Errorf("got a 302 redirect but no location URL")
		}
		return nc.RequestWithContext(ctx, method, location, r)
	}
	return resp, nil
}
