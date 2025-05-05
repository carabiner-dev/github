// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

// FileCaller implements the Caller but returns the data as read from a file
// specified in SourcePath. This is intended mostly for testing.
type FileCaller struct {
	SourcePath string
	Error      error
}

// RequestWithContext simulates the request to the API by returning a response from
// a file. If Error is set, it will return nil and the synthetic error.
func (fc *FileCaller) RequestWithContext(_ context.Context, _, _ string, _ io.Reader) (*http.Response, error) {
	if fc.Error != nil {
		return nil, fc.Error
	}
	f, err := os.Open(fc.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("opening source file: %w", err)
	}
	return &http.Response{
		Status:           "OK",
		StatusCode:       http.StatusOK,
		Body:             f,
		ContentLength:    999,
		TransferEncoding: []string{},
	}, nil
}
