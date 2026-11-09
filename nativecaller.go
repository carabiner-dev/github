// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	apiVersion      = "2022-11-28"
	defaultHostname = "api.github.com"
)

type errorBody struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	StatusCode       string `json:"status"`
}

// NewNativeHTTPCaller returns a caller that talks to the GitHub API with
// Go's HTTP client, authenticated with the options' token when there is
// one. Every attempt is bounded by the options' timeout and a failed
// attempt is retried with backoff (see Options).
func NewNativeHTTPCaller(opts *Options) (*NativeHTTPCaller, error) {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	retryWait := opts.RetryWait
	if retryWait <= 0 {
		retryWait = DefaultRetryWait
	}
	plain := &http.Client{Timeout: timeout}
	client := plain
	if opts.Token != "" {
		// oauth2 builds its transport over the client it finds in the context.
		ctx := context.WithValue(context.Background(), oauth2.HTTPClient, plain)
		client = oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: opts.Token},
		))
		client.Timeout = timeout
	}
	host := defaultHostname
	if opts.Host != "" {
		host = opts.Host
	}
	return &NativeHTTPCaller{
		client:    client,
		plain:     plain,
		Hostname:  host,
		retries:   opts.Retries,
		retryWait: retryWait,
	}, nil
}

// NativeHTTPCaller sends requests to the GitHub API with Go's HTTP client.
type NativeHTTPCaller struct {
	// client sends authenticated requests to Hostname; plain sends the
	// unauthenticated ones a redirect takes elsewhere (release assets).
	client    *http.Client
	plain     *http.Client
	Hostname  string
	retries   uint
	retryWait time.Duration
}

// maxRetryWait caps the backoff between attempts.
const maxRetryWait = time.Minute

// RequestWithContext sends a request and returns its response. An attempt
// that fails on the transport, times out, or is answered 429 or 5xx (other
// than 501) is retried with backoff, as long as the request carries no
// body, since a body cannot be replayed. A 302 is followed, unauthenticated
// when it leaves the API host, as GitHub does for release assets.
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
		if parsed.Hostname() != nc.Hostname {
			sendWithoutAuth = true
			hostname = parsed.Host
		}
		endpoint = strings.TrimPrefix(endpoint, parsed.Scheme+"://"+parsed.Host)
	}
	urlString := fmt.Sprintf("https://%s/%s", hostname, strings.TrimPrefix(endpoint, "/"))
	client := nc.client
	if sendWithoutAuth {
		client = nc.plain
	}

	retries := nc.retries
	if r != nil {
		retries = 0
	}
	wait := nc.retryWait
	var resp *http.Response
	for attempt := uint(0); ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, urlString, r)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}
		req.Header.Add("Accept", "application/vnd.github+json")
		req.Header.Add("X-Github-Api-Version", apiVersion)
		resp, err = client.Do(req)
		if err == nil && !retryableStatus(resp.StatusCode) {
			break
		}
		if attempt >= retries {
			if err != nil {
				return nil, err
			}
			break
		}
		if resp != nil {
			resp.Body.Close() //nolint:errcheck,gosec // closing before a retry, nothing to do with the error
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
		wait = min(wait*2, maxRetryWait)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 399 {
		res, err := io.ReadAll(resp.Body)
		if err == nil {
			eb := errorBody{}
			if err = json.Unmarshal(res, &eb); err == nil {
				return resp, fmt.Errorf("HTTP Error %d sending request: %s", resp.StatusCode, eb.Message)
			}
		}
		return resp, fmt.Errorf("HTTP Error %d sending request", resp.StatusCode)
	}

	if resp.StatusCode == http.StatusFound {
		location := resp.Header.Get("Location")
		if location == "" {
			return nil, fmt.Errorf("got a 302 redirect but no location URL")
		}
		return nc.RequestWithContext(ctx, method, location, r)
	}
	return resp, nil
}

// retryableStatus reports whether a response status is worth another
// attempt: rate limiting and server errors, except Not Implemented.
func retryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || (code >= 500 && code != http.StatusNotImplemented)
}
