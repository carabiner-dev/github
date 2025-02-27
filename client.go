// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"io"
	"net/http"
)

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
	if opts.Caller == nil {
		rclient, err := buildGithubRestClient(opts)
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

func (c *Client) Call(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	return c.caller.RequestWithContext(ctx, method, path, body)
}
