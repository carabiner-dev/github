// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNCRequestWithContext(t *testing.T) {
	t.Parallel()
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

			res, err := nc.RequestWithContext(t.Context(), http.MethodGet, tc.url, nil)
			require.NoError(t, err)
			defer res.Body.Close() //nolint:errcheck
			require.NotNil(t, res)
			require.Equal(t, 200, res.StatusCode)
		})
	}
}

// testCaller points a caller at a TLS test server, trusting its certificate,
// with the given per-attempt timeout and retries.
func testCaller(t *testing.T, srv *httptest.Server, timeout time.Duration, retries uint) *NativeHTTPCaller {
	t.Helper()
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	client := srv.Client()
	client.Timeout = timeout
	return &NativeHTTPCaller{
		client:    client,
		plain:     client,
		Hostname:  u.Hostname(),
		retries:   retries,
		retryWait: time.Millisecond,
	}
}

// Pointing the caller at a host:port means its Hostname carries the port.
func testCallerWithPort(t *testing.T, srv *httptest.Server, timeout time.Duration, retries uint) *NativeHTTPCaller {
	t.Helper()
	nc := testCaller(t, srv, timeout, retries)
	u, err := url.Parse(srv.URL)
	require.NoError(t, err)
	nc.Hostname = u.Host
	return nc
}

func TestNCRetries(t *testing.T) {
	t.Parallel()
	t.Run("server errors are retried until success", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if calls.Add(1) < 3 {
				w.WriteHeader(http.StatusBadGateway)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		nc := testCallerWithPort(t, srv, time.Second, 3)
		res, err := nc.RequestWithContext(t.Context(), http.MethodGet, "x", nil)
		require.NoError(t, err)
		defer res.Body.Close() //nolint:errcheck
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, int32(3), calls.Load())
	})
	t.Run("a stalled attempt times out and is retried", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if calls.Add(1) < 2 {
				time.Sleep(300 * time.Millisecond)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		nc := testCallerWithPort(t, srv, 50*time.Millisecond, 3)
		res, err := nc.RequestWithContext(t.Context(), http.MethodGet, "x", nil)
		require.NoError(t, err)
		defer res.Body.Close() //nolint:errcheck
		require.Equal(t, http.StatusOK, res.StatusCode)
		require.Equal(t, int32(2), calls.Load())
	})
	t.Run("retries run out", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer srv.Close()
		nc := testCallerWithPort(t, srv, time.Second, 2)
		res, err := nc.RequestWithContext(t.Context(), http.MethodGet, "x", nil)
		require.Error(t, err)
		if res != nil {
			res.Body.Close() //nolint:errcheck,gosec
		}
		require.Equal(t, int32(3), calls.Load(), "one attempt plus two retries")
	})
	t.Run("client errors are not retried", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()
		nc := testCallerWithPort(t, srv, time.Second, 3)
		res, err := nc.RequestWithContext(t.Context(), http.MethodGet, "x", nil)
		require.Error(t, err)
		if res != nil {
			res.Body.Close() //nolint:errcheck,gosec
		}
		require.Equal(t, int32(1), calls.Load())
	})
	t.Run("a request with a body is not retried", func(t *testing.T) {
		t.Parallel()
		var calls atomic.Int32
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls.Add(1)
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer srv.Close()
		nc := testCallerWithPort(t, srv, time.Second, 3)
		res, err := nc.RequestWithContext(t.Context(), http.MethodPost, "x", strings.NewReader("{}"))
		require.Error(t, err)
		if res != nil {
			res.Body.Close() //nolint:errcheck,gosec
		}
		require.Equal(t, int32(1), calls.Load())
	})
	t.Run("a cancelled context stops the retries", func(t *testing.T) {
		t.Parallel()
		srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadGateway)
		}))
		defer srv.Close()
		nc := testCallerWithPort(t, srv, time.Second, 5)
		nc.retryWait = time.Hour
		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()
		res, err := nc.RequestWithContext(ctx, http.MethodGet, "x", nil)
		if res != nil {
			res.Body.Close() //nolint:errcheck,gosec
		}
		require.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestNCDefaults(t *testing.T) {
	t.Parallel()
	nc, err := NewNativeHTTPCaller(&Options{})
	require.NoError(t, err)
	require.Equal(t, DefaultTimeout, nc.client.Timeout)
	require.Equal(t, DefaultTimeout, nc.plain.Timeout)
	require.Equal(t, DefaultRetryWait, nc.retryWait)
	nc, err = NewNativeHTTPCaller(&Options{Token: "t", Timeout: time.Second, Retries: 7, RetryWait: time.Millisecond})
	require.NoError(t, err)
	require.Equal(t, time.Second, nc.client.Timeout, "the authenticated client is bounded too")
	require.Equal(t, uint(7), nc.retries)
	require.Equal(t, time.Millisecond, nc.retryWait)
}
