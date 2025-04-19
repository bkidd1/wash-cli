package chat

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/brinleekidd/wash-cli/internal/chatmonitor"
	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/spf13/cobra"
)

// Command creates the chat command with start and stop subcommands
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chat",
		Short: "Monitor and analyze chat interactions",
		Long:  "Monitor and analyze chat interactions for insights and improvements",
	}

	cmd.AddCommand(startCmd())
	cmd.AddCommand(stopCmd())

	return cmd
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start chat monitoring",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			monitor, err := chatmonitor.NewChatMonitor(cfg)
			if err != nil {
				return fmt.Errorf("failed to create chat monitor: %w", err)
			}

			// Handle graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				monitor.Stop()
			}()

			if err := monitor.Start(); err != nil {
				return fmt.Errorf("failed to start chat monitor: %w", err)
			}

			return nil
		},
	}
}

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop chat monitoring",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			monitor, err := chatmonitor.NewChatMonitor(cfg)
			if err != nil {
				return fmt.Errorf("failed to create chat monitor: %w", err)
			}

			if err := monitor.Stop(); err != nil {
				return fmt.Errorf("failed to stop chat monitor: %w", err)
			}

			return nil
		},
	}
}
