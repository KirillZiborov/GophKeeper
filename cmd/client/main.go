package main

import (
	"fmt"

	"github.com/KirillZiborov/GophKeeper/cmd/client/cmd"
	"github.com/KirillZiborov/GophKeeper/internal/logging"
)

var (
	// Use go run -ldflags to set up build variables while compiling.
	buildVersion = "N/A" // Build version
	buildDate    = "N/A" // Build date
)

func main() {
	// Print build info.
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)

	// Initialize the logging system.
	err := logging.Initialize()
	if err != nil {
		logging.Sugar.Errorw("Internal logging error", "error", err)
	}
	cmd.Execute()
}
