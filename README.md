# github.com/carabiner-dev/github

## A super simple client to talk to the GitHub API.

## Description

This library implements a very very simple client to talk to the GitHub 
(and someday possibly other) APIs. The goals of this client are:

1. Lightweight: Low dependency count.
2. Easy to use in tests, easy to mock.
3. Pluggable credential (token) providers.

Non-goals include:

1. Response parsing
2. Anything else beyond handling https connections.

## Usage

To use the client, simply export a token in `GITHUB_TOKEN` and run something
like:

```golang
package main

import  "github.com/carabiner-dev/github"

func main() {
    // Creat the client
    client, err := github.NewClient()

    // Call an API endpoint:
    resp, err := client.Call(ctx, http.MethodGet, "users/carabiner-dev", nil)
    if err != nil {
        fmt.Fprintln(os.Stderr, err.Error())
    }

    // Output the response to stdout:
    io.Copy(os.Stdout, resp.Body)
    resp.Close()
}
```

## Install

To pull the package simply use go get:

```
github.com/carabiner-dev/github
```

## License

This package is released under the Apache 2.0 license by Carabiner Systems, Inc. 
Patches issues and contributions are welcome!
