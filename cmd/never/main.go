package main

import (
	"context"
	"fmt"
	"os"

	"github.com/containeroo/never/internal/app"
)

const Version string = "dev"

func main() {
	// Create a root context
	ctx := context.Background()

	if err := app.Run(ctx, Version, os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
