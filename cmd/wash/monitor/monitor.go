package monitor

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/monitor/chatmonitor"
	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	projectName string
	pidFile     = filepath.Join(os.TempDir(), "wash-monitor.pid")
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

Use the stop subcommand to stop monitoring.

Examples:
  # Start monitoring current project
  wash monitor

  # Start monitoring specific project
  wash monitor --project my-project

  # Stop monitoring
  wash monitor stop`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if monitor is already running
			if _, err := os.Stat(pidFile); err == nil {
				// Read PID from file
				pidBytes, err := os.ReadFile(pidFile)
				if err != nil {
					// Clean up invalid PID file
					os.Remove(pidFile)
				} else {
					pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
					if err == nil {
						// Check if process exists and is running
						process, err := os.FindProcess(pid)
						if err == nil {
							// On Unix systems, FindProcess always succeeds, so we need to check if the process is actually running
							if err := process.Signal(syscall.Signal(0)); err == nil {
								return fmt.Errorf("monitor is already running. Use 'wash monitor stop' to stop it first")
							}
						}
					}
					// Clean up invalid or stale PID file
					os.Remove(pidFile)
				}
			}

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

			// Start monitoring
			if err := m.Start(); err != nil {
				return fmt.Errorf("failed to start monitor: %w", err)
			}

			// Write PID to file
			if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
				return fmt.Errorf("failed to write PID file: %w", err)
			}

			// Start time for elapsed time calculation
			startTime := time.Now()

			// Create a ticker for updating the timer display
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			// Create a channel for handling interrupts
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

			// Display timer in foreground
			fmt.Println("Monitoring started. Press Ctrl+C to stop.")
			for {
				select {
				case <-ticker.C:
					elapsed := time.Since(startTime)
					fmt.Printf("\rMonitoring for: %02d:%02d:%02d",
						int(elapsed.Hours()),
						int(elapsed.Minutes())%60,
						int(elapsed.Seconds())%60)
				case <-interrupt:
					fmt.Println("\nStopping monitor...")
					m.Stop()
					os.Remove(pidFile)
					return nil
				}
			}
		},
	}

	// Add global flags
	cmd.PersistentFlags().StringVarP(&projectName, "project", "p", "", "Project name (defaults to current directory name)")

	// Add stop command
	cmd.AddCommand(stopCmd())

	return cmd
}

func runMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "run-monitor",
		Short:  "Run the monitor process (internal use)",
		Hidden: true,
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

			// Start monitoring
			if err := m.Start(); err != nil {
				return fmt.Errorf("failed to start monitor: %w", err)
			}

			// Start time for elapsed time calculation
			startTime := time.Now()

			// Create a ticker for updating the timer display
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			// Create a channel for handling interrupts
			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)

			// Display timer in foreground
			fmt.Println("Monitoring started. Press Ctrl+C to stop.")
			for {
				select {
				case <-ticker.C:
					elapsed := time.Since(startTime)
					fmt.Printf("\rMonitoring for: %02d:%02d:%02d",
						int(elapsed.Hours()),
						int(elapsed.Minutes())%60,
						int(elapsed.Seconds())%60)
				case <-interrupt:
					fmt.Println("\nStopping monitor...")
					m.Stop()
					os.Remove(pidFile)
					return nil
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
			// Check if PID file exists
			if _, err := os.Stat(pidFile); os.IsNotExist(err) {
				fmt.Println("No monitor process is running")
				return nil
			}

			// Read PID from file
			pidBytes, err := os.ReadFile(pidFile)
			if err != nil {
				// Clean up invalid PID file
				os.Remove(pidFile)
				fmt.Println("No monitor process is running")
				return nil
			}

			pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
			if err != nil {
				// Clean up invalid PID file
				os.Remove(pidFile)
				fmt.Println("No monitor process is running")
				return nil
			}

			// Check if process exists and is running
			process, err := os.FindProcess(pid)
			if err != nil {
				// Clean up PID file for non-existent process
				os.Remove(pidFile)
				fmt.Println("No monitor process is running")
				return nil
			}

			// On Unix systems, FindProcess always succeeds, so we need to check if the process is actually running
			if err := process.Signal(syscall.Signal(0)); err != nil {
				// Process not running, clean up PID file
				os.Remove(pidFile)
				fmt.Println("No monitor process is running")
				return nil
			}

			// Send termination signal to the process group
			pgid, err := syscall.Getpgid(pid)
			if err != nil {
				// Clean up PID file if we can't get the process group
				os.Remove(pidFile)
				fmt.Println("No monitor process is running")
				return nil
			}

			if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
				return fmt.Errorf("failed to stop monitor: %w", err)
			}

			// Remove PID file
			os.Remove(pidFile)

			fmt.Println("Monitoring stopped")
			return nil
		},
	}

	return cmd
}
