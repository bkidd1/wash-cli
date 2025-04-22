package monitor

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/bkidd1/wash-cli/internal/services/monitor/chatmonitor"
	"github.com/bkidd1/wash-cli/internal/utils/config"
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
	var projectName string

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start monitoring",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If project name not provided, use current directory name
			if projectName == "" {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get current directory: %w", err)
				}
				projectName = filepath.Base(cwd)
			}

			// Load configuration
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create monitor
			m, err := chatmonitor.NewMonitor(cfg, projectName)
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

			// Start monitoring
			if err := m.Start(); err != nil {
				return fmt.Errorf("failed to start monitor: %w", err)
			}

			fmt.Printf("Monitoring started for project: %s\n", projectName)
			fmt.Println("Press Ctrl+C to stop monitoring...")

			// Wait for monitor to complete
			select {
			case <-sigChan:
				m.Stop()
				return nil
			}
		},
	}

	cmd.Flags().StringVarP(&projectName, "project", "p", "", "Project name (defaults to current directory name)")
	return cmd
}

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop monitoring",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load configuration
			cfg, err := config.LoadConfig()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create monitor
			m, err := chatmonitor.NewMonitor(cfg, "")
			if err != nil {
				return fmt.Errorf("failed to create monitor: %w", err)
			}

			if err := m.Stop(); err != nil {
				return fmt.Errorf("failed to stop monitor: %w", err)
			}

			fmt.Println("Monitoring stopped")
			return nil
		},
	}
}
