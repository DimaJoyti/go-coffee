package main

import (
	"fmt"
	"os"

	"github.com/DimaJoyti/go-coffee/cmd/task-cli/commands"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Set version information
	commands.SetVersionInfo(version, buildTime, gitCommit)
	
	// Execute the root command
	commands.Execute()
}

func init() {
	// Handle panics gracefully
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Fatal error: %v\n", r)
			os.Exit(1)
		}
	}()
}
