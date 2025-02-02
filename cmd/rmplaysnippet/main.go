// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The rmplaysnippet binary removes a code snippet from play.golang.org given its URL
// or ID. It will always connect to the production datastore instance, ignoring any
// local value of DATASTORE_EMULATOR_HOST.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"golang.org/x/build/buildenv"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s {http(s)://play.golang.org/p/<id> | <id>}\n", os.Args[0])
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(2)
	}

	snippetID := os.Args[1]
	prefixes := []string{
		"https://play.golang.org/p/",
		"http://play.golang.org/p/",
		"https://go.dev/play/p/",
		"http://go.dev/play/p/",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(os.Args[1], p) {
			snippetID = strings.TrimPrefix(os.Args[1], p)
			break
		}
	}
	if snippetID == "" {
		usage()
		os.Exit(2)
	}
	if strings.Contains(snippetID, "/") {
		usage()
		fmt.Fprintf(os.Stderr, "Invalid Snippet ID %q (contains slash)\n", snippetID)
		os.Exit(2)
	}

	fmt.Printf("Really delete Snippet with ID %q? [y,N]: ", snippetID)
	var confirm string
	fmt.Scanln(&confirm)
	if !strings.HasPrefix(strings.ToLower(confirm), "y") {
		fmt.Printf("Aborting ...\n")
		os.Exit(0)
	}

	buildenv.CheckUserCredentials()

	// Don't attempt to connect to a locally-running datastore instance.
	if err := os.Setenv("DATASTORE_EMULATOR_HOST", ""); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to clear env var DATASTORE_EMULATOR_HOST: %v\n", err)
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	client, err := datastore.NewClient(ctx, "golang-org")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create Datastore client: %v\n", err)
		os.Exit(1)
	}
	k := datastore.NameKey("Snippet", snippetID, nil)
	if client.Get(ctx, k, new(struct{})) == datastore.ErrNoSuchEntity {
		fmt.Fprintf(os.Stderr, "Snippet with ID %q does not exist\n", snippetID)
		os.Exit(0)
	}
	fmt.Printf("Deleting snippet %q ...\n", snippetID)
	if err := client.Delete(ctx, k); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to delete Snippet with ID %q: %v\n", snippetID, err)
		fmt.Fprintf(os.Stderr, "rmplaysnippet requires Application Default Credentials.\n")
		fmt.Fprintf(os.Stderr, "Did you run `gcloud auth application-default login`?\n")
		os.Exit(1)
	}
	fmt.Printf("Snippet with ID %q deleted\n", snippetID)
}
