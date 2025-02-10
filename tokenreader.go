// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"errors"
	"os"
)

// DefaultEnvTokenReader is an env token reader configured to read the token
// from the GITHUB_TOKEN env variable.
var DefaultEnvTokenReader = EnvTokenReader{
	VarName: "GITHUB_TOKEN",
}

// TokenReader is an interface that abstracts a mechanism to read a token securely.
type TokenReader interface {
	ReadToken() (string, error)
}

// EnvTokenReader is a token reader that extracts an API token from an environment
// variable
type EnvTokenReader struct {
	VarName string
}

// ReadToken reads the token string from the convigured `VarName` environment variable.
func (etr *EnvTokenReader) ReadToken() (string, error) {
	if etr.VarName == "" {
		return "", errors.New("environment variable name to read token not set")
	}
	return os.Getenv(etr.VarName), nil
}
