package main

import (
	"fmt"
	"os"

	"github.com/KevG1t/SpecAI/internal/app"
)

var version = "dev"

func main() {
	app.Version = app.ResolveVersion(version)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
