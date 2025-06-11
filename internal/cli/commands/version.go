package commands

import (
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version command
func NewVersionCommand(version, commit, date string) *cobra.Command {
	var detailed bool

	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			if detailed {
				printDetailedVersion(version, commit, date)
			} else {
				printSimpleVersion(version)
			}
		},
	}

	cmd.Flags().BoolVarP(&detailed, "detailed", "d", false, "Show detailed version information")

	return cmd
}

func printSimpleVersion(version string) {
	fmt.Printf("gocoffee version %s\n", version)
}

func printDetailedVersion(version, commit, date string) {
	color.Cyan("Go Coffee CLI")
	fmt.Printf("Version:    %s\n", version)
	fmt.Printf("Commit:     %s\n", commit)
	fmt.Printf("Built:      %s\n", date)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
