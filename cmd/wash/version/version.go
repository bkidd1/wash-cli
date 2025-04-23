package versioncmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Command returns the version command
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Long:  `Print the version information including the version number, commit hash, and build date.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd.Root().Version)
		},
	}
}
