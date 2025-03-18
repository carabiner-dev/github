// SPDX-FileCopyrightText: Copyright 2025 Carabiner Systems, Inc
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"net/url"
	"strings"
)

// RepoFromString returns a repository URL from a string. It supports
// full URLs, repo slugs (org/repo), VCS locators and possibly more
// input types will be added in the future.
func RepoFromString(str string) (string, error) {
	// String is a github URL without scheme
	if strings.HasPrefix(str, "github.com") {
		str = "https://" + str
	}

	u, err := url.Parse(str)
	if err != nil {
		return "", fmt.Errorf("parsing url string: %w", err)
	}

	host := u.Hostname()
	if host == "" {
		host = "github.com"
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("URL path needs to have an org and repo")
	}

	// If it's a locator, trim the branch (if present)
	if strings.HasPrefix(u.Scheme, "git+") {
		f, _, _ := strings.Cut(parts[len(parts)-1], "@")
		parts[len(parts)-1] = f
	}

	scheme := strings.TrimPrefix(u.Scheme, "git+")
	if scheme == "" {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s/%s/%s", scheme, host, parts[0], parts[1]), nil
}
