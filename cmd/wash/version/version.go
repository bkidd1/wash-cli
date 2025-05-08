package versioncmd

import (
	"fmt"

	"github.com/bkidd1/wash-cli/pkg/version"
	"github.com/spf13/cobra"
)

// Command returns the version command
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version information including the version number, commit hash, build date, Go version, and platform.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Get())
		},
	}
}
