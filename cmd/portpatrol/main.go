package main

import (
	"context"
	"fmt"
	"os"

	"github.com/containeroo/portpatrol/internal/app"
)

const version string = "v0.5.11"

func main() {
	// Create a root context
	ctx := context.Background()

	if err := app.Run(ctx, version, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
