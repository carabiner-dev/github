// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"fmt"
)

type Options struct {
	Host        string
	Token       string
	EnsureToken bool
	TokenReader TokenReader
	Caller      Caller
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
	if o.EnsureToken {
		if err := o.ensureToken(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
