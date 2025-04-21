package monitor

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	chatmonitor "github.com/bkidd1/wash-cli/internal/monitor/chatmonitor"
	"github.com/bkidd1/wash-cli/pkg/config"
	"github.com/spf13/cobra"
)

// Command creates the monitor command with start and stop subcommands
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor and analyze interactions",
		Long:  "Monitor and analyze interactions for insights and improvements",
	}

	cmd.AddCommand(startCmd())
	cmd.AddCommand(stopCmd())

	return cmd
}

func startCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start monitoring",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			m, err := chatmonitor.NewMonitor(cfg)
			if err != nil {
				return fmt.Errorf("failed to create monitor: %w", err)
			}

			// Handle graceful shutdown
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				m.Stop()
			}()

			if err := m.Start(); err != nil {
				return fmt.Errorf("failed to start monitor: %w", err)
			}

			return nil
		},
	}
}

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop monitoring",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			m, err := chatmonitor.NewMonitor(cfg)
			if err != nil {
				return fmt.Errorf("failed to create monitor: %w", err)
			}

			if err := m.Stop(); err != nil {
				return fmt.Errorf("failed to stop monitor: %w", err)
			}

			return nil
		},
	}
}
