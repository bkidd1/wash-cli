package main

import (
	"context"
	"fmt"
	"os"

	"github.com/brinleekidd/wash-cli/internal/analyzer"
	"github.com/brinleekidd/wash-cli/internal/monitor"
	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	analyzerInstance *analyzer.Analyzer
	monitorInstance  *monitor.Monitor
	cfg              *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "wash",
	Short: "Wash is a CLI tool for exploring alternative coding pathways",
	Long: `Wash is a CLI tool that helps developers explore alternative coding pathways
when they're stuck. It uses AI to analyze your code and suggest different approaches
you could have taken.`,
}

var analyzeCmd = &cobra.Command{
	Use:   "analyze [file]",
	Short: "Analyze a single file for optimization opportunities",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result, err := analyzerInstance.AnalyzePathway(context.Background(), args[0])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}

var exploreCmd = &cobra.Command{
	Use:   "explore [path]",
	Short: "Explore and analyze project structure",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := "."
		if len(args) > 0 {
			path = args[0]
		}
		result, err := analyzerInstance.ExploreProject(context.Background(), path)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start chat monitoring",
	Run: func(cmd *cobra.Command, args []string) {
		if err := monitorInstance.StartMonitoring(context.Background()); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Chat monitoring started")
	},
}

var stopMonitorCmd = &cobra.Command{
	Use:   "stop-monitor",
	Short: "Stop chat monitoring",
	Run: func(cmd *cobra.Command, args []string) {
		if err := monitorInstance.StopMonitoring(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Chat monitoring stopped")
	},
}

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Generate a summary of the chat analysis",
	Run: func(cmd *cobra.Command, args []string) {
		result, err := monitorInstance.GenerateSummary(context.Background())
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Wash settings",
}

var configSetCmd = &cobra.Command{
	Use:   "set-openai-key [key]",
	Short: "Set the OpenAI API key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg.OpenAIAPIKey = args[0]
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("OpenAI API key saved successfully")
	},
}

func init() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	analyzerInstance = analyzer.NewAnalyzer()
	monitorInstance, err = monitor.NewMonitor()
	if err != nil {
		fmt.Printf("Error initializing monitor: %v\n", err)
		os.Exit(1)
	}

	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(analyzeCmd)
	rootCmd.AddCommand(exploreCmd)
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(stopMonitorCmd)
	rootCmd.AddCommand(summaryCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
