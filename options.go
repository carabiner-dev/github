// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"fmt"
	"time"
)

// Defaults for the native caller's HTTP behavior. Each attempt is bounded
// by DefaultTimeout; an attempt that fails on the transport, times out, or
// gets a 429 or a 5xx is retried DefaultRetries times, waiting
// DefaultRetryWait before the first retry and doubling after each.
const (
	DefaultTimeout   = 15 * time.Second
	DefaultRetries   = uint(3)
	DefaultRetryWait = 2 * time.Second
)

type Options struct {
	Host        string
	Token       string
	EnsureToken bool
	TokenReader TokenReader
	Caller      Caller
	// Timeout bounds each HTTP attempt, headers and body. Zero uses
	// DefaultTimeout.
	Timeout time.Duration
	// Retries is how many times a failed attempt is retried: a transport
	// error, a timeout, a 429 or a 5xx other than 501. Requests with a body
	// are never retried, since the body has been consumed.
	Retries uint
	// RetryWait is the wait before the first retry; it doubles after each,
	// up to a minute. Zero uses DefaultRetryWait.
	RetryWait time.Duration
}

type fnOpt func(*Options)

func WithHost(h string) fnOpt {
	return func(opts *Options) {
		opts.Host = h
	}
}

func WithToken(t string) fnOpt {
	return func(opts *Options) {
		opts.Token = t
	}
}

func WithEnsureToken(e bool) fnOpt {
	return func(opts *Options) {
		opts.EnsureToken = e
	}
}

func WithTokenReader(r TokenReader) fnOpt {
	return func(opts *Options) {
		opts.TokenReader = r
	}
}

// WithTimeout bounds each HTTP attempt.
func WithTimeout(d time.Duration) fnOpt {
	return func(opts *Options) {
		opts.Timeout = d
	}
}

// WithRetries sets how many times a failed attempt is retried.
func WithRetries(n uint) fnOpt {
	return func(opts *Options) {
		opts.Retries = n
	}
}

// WithRetryWait sets the wait before the first retry.
func WithRetryWait(d time.Duration) fnOpt {
	return func(opts *Options) {
		opts.RetryWait = d
	}
}

func WithCaller(c Caller) fnOpt {
	return func(opts *Options) {
		opts.Caller = c
	}
}

var defaultOptions = Options{
	Host: DefaultAPIHostname,
	TokenReader: &EnvTokenReader{
		VarName: "GITHUB_TOKEN",
	},
	Timeout:   DefaultTimeout,
	Retries:   DefaultRetries,
	RetryWait: DefaultRetryWait,
}

// ensureToken makes sure we have a token. If there is none set, we
// read it using the token reader
func (o *Options) ensureToken() error {
	if o.Token != "" {
		return nil
	}
	if o.TokenReader == nil {
		return errors.New("no token set and no token reader configured")
	}

	token, err := o.TokenReader.ReadToken()
	if err != nil {
		return fmt.Errorf("reading token: %w", err)
	}

	if token == "" {
		return fmt.Errorf("unable to get a token from the token reader")
	}

	o.Token = token
	return nil
}

// Validate ensures the client options are sane
func (o *Options) Validate() error {
	errs := []error{}
	if err := o.ensureToken(); err != nil {
		if o.EnsureToken {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
