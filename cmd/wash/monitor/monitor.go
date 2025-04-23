package monitor

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/monitor/chatmonitor"
	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	projectName string
)

// Command creates the monitor command with start and stop subcommands
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor and analyze development interactions",
		Long: `Monitor and analyze your development workflow to provide insights and improvements.
The monitor tracks:
- Code changes
- Development patterns
- Interaction patterns
- Time spent on tasks
- Project progress

Use the start and stop subcommands to control monitoring.

Examples:
  # Start monitoring current project
  wash monitor start

  # Start monitoring specific project
  wash monitor start --project my-project

  # Stop monitoring
  wash monitor stop`,
	}

	// Add global flags
	cmd.PersistentFlags().StringVarP(&projectName, "project", "p", "", "Project name (defaults to current directory name)")

	// Add subcommands
	cmd.AddCommand(startCmd())
	cmd.AddCommand(stopCmd())

	return cmd
}

func startCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start monitoring development workflow",
		Long: `Start monitoring your development workflow to track progress and provide insights.
The monitor will:
1. Track code changes and interactions
2. Analyze development patterns
3. Generate progress reports
4. Provide optimization suggestions`,
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

			fmt.Printf("Monitoring started for %s\n", projectName)

			// Start timer display
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			// Wait for monitor to complete
			for {
				select {
				case <-sigChan:
					m.Stop()
					return nil
				case <-ticker.C:
					elapsed := time.Since(m.StartTime())
					hours := int(elapsed.Hours())
					minutes := int(elapsed.Minutes()) % 60
					seconds := int(elapsed.Seconds()) % 60
					fmt.Printf("\rRunning for: %02d:%02d:%02d", hours, minutes, seconds)
				}
			}
		},
	}

	return cmd
}

func stopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop monitoring development workflow",
		Long: `Stop the development workflow monitor.
This will:
1. Stop tracking new changes
2. Save current progress
3. Generate final report`,
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if err := m.Stop(); err != nil {
				return fmt.Errorf("failed to stop monitor: %w", err)
			}

			fmt.Println("Monitoring stopped")
			return nil
		},
	}

	return cmd
}
