package error

import (
	"github.com/spf13/cobra"
)

// NewErrorCmd creates the error analysis command
func NewErrorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bruh",
		Short: "Analyze the last error or change",
		Long: `Takes a screenshot and analyzes the last change or error,
providing suggestions for resolution. This command is particularly
useful when you're stuck on an error or want to improve your
last code change.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement error analysis
			return nil
		},
	}

	return cmd
}
