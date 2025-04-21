package version

import (
	"fmt"

	"github.com/bkidd1/wash-cli/pkg/version"
	"github.com/spf13/cobra"
)

// Command creates the version command
func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Get())
		},
	}
}
