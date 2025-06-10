package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/DimaJoyti/go-coffee/cmd/task-cli/commands"
	"github.com/DimaJoyti/go-coffee/pkg/logger"
)

var (
	version   = "2.0.0"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Initialize enhanced logger
	logConfig := logger.DevelopmentConfig()
	logConfig.Service = "go-coffee-cli"
	log := logger.NewLogger(logConfig)

	// Display startup banner
	displayBanner(log)

	// Set version information
	commands.SetVersionInfo(version, buildTime, gitCommit)

	// Execute the root command with enhanced error handling
	commands.Execute()
}

func init() {
	// Handle panics gracefully with enhanced error reporting
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "🚨 Fatal error occurred: %v\n", r)

			// Print stack trace in debug mode
			if os.Getenv("DEBUG") == "true" {
				buf := make([]byte, 1024)
				for {
					n := runtime.Stack(buf, false)
					if n < len(buf) {
						buf = buf[:n]
						break
					}
					buf = make([]byte, 2*len(buf))
				}
				fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", buf)
			}

			os.Exit(1)
		}
	}()
}

// displayBanner shows an enhanced startup banner
func displayBanner(log *logger.Logger) {
	banner := `
☕ ═══════════════════════════════════════════════════════════════════════════════ ☕
   ██████╗  ██████╗       ██████╗ ██████╗ ███████╗███████╗███████╗███████╗
  ██╔════╝ ██╔═══██╗     ██╔════╝██╔═══██╗██╔════╝██╔════╝██╔════╝██╔════╝
  ██║  ███╗██║   ██║     ██║     ██║   ██║█████╗  █████╗  █████╗  █████╗
  ██║   ██║██║   ██║     ██║     ██║   ██║██╔══╝  ██╔══╝  ██╔══╝  ██╔══╝
  ╚██████╔╝╚██████╔╝     ╚██████╗╚██████╔╝██║     ██║     ███████╗███████╗
   ╚═════╝  ╚═════╝       ╚═════╝ ╚═════╝ ╚═╝     ╚═╝     ╚══════╝╚══════╝

                    🌐 Web3 Coffee Ecosystem CLI Tool 🤖
☕ ═══════════════════════════════════════════════════════════════════════════════ ☕`

	fmt.Println(banner)

	log.WithFields(map[string]interface{}{
		"version":    version,
		"build_time": buildTime,
		"git_commit": gitCommit,
		"go_version": runtime.Version(),
	}).Info("🚀 Go Coffee CLI initialized")
}
