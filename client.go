// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"io"
	"net/http"
	"strings"
)

const DefaultAPIHostname = "api.github.com"

// Replaceable caller interface
type Caller interface {
	RequestWithContext(context.Context, string, string, io.Reader) (*http.Response, error)
}

// NewClient creates a new client. It can take any number of optional functions
func NewClient(optFns ...fnOpt) (*Client, error) {
	opts := defaultOptions
	for _, fn := range optFns {
		fn(&opts)
	}
	return NewClientWithOptions(opts)
}

// NewClientWithOptions creates a new client, taking a full options set.
func NewClientWithOptions(opts Options) (*Client, error) {
	// Ensure the client has a token to connect
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	// If we didn't get a caller in the options, default
	// to the stock github rest client.
	var rclient Caller
	var err error
	if opts.Caller == nil {
		rclient, err = NewNativeHTTPCaller(&opts)
		if err != nil {
			return nil, err
		}
		opts.Caller = rclient
	}

	return &Client{
		Options: opts,
		caller:  opts.Caller,
	}, nil
}

type Client struct {
	Options Options
	caller  Caller
}

// TokenScopes returns the scopes of the token as reflected by the GitHub API.
func (c *Client) TokenScopes() ([]string, error) {
	resp, err := c.caller.RequestWithContext(context.Background(), http.MethodGet, "/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //nolint:errcheck

	scopesString := resp.Header.Get("X-Oauth-Scopes")
	if scopesString == "" {
		return []string{}, nil
	}
	return strings.Split(scopesString, ", "), nil
}

func (c *Client) Call(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	return c.caller.RequestWithContext(ctx, method, path, body)
}
