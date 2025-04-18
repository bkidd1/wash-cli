package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/brinleekidd/wash-cli/internal/analyzer"
	"github.com/brinleekidd/wash-cli/internal/monitor"
	"github.com/brinleekidd/wash-cli/internal/notes"
	"github.com/brinleekidd/wash-cli/internal/screenshot"
	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

var (
	cfg *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "wash",
	Short: "Wash is a CLI tool for code analysis and optimization",
	Long: `Wash is a command-line tool that helps analyze and optimize your code.
It provides various commands to analyze files, project structure, and more.`,
}

var analyzeFileCmd = &cobra.Command{
	Use:   "analyze-file [file]",
	Short: "Analyze a specific file for optimizations",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		client := openai.NewClient(cfg.OpenAIKey)
		analyzer := analyzer.NewAnalyzer(client)

		result, err := analyzer.AnalyzeFile(context.Background(), filePath)
		if err != nil {
			fmt.Printf("Error analyzing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result)
	},
}

var analyzeProjectCmd = &cobra.Command{
	Use:   "analyze-project [directory]",
	Short: "Analyze project structure and organization",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dirPath := args[0]
		client := openai.NewClient(cfg.OpenAIKey)
		analyzer := analyzer.NewAnalyzer(client)

		result, err := analyzer.AnalyzeProjectStructure(context.Background(), dirPath)
		if err != nil {
			fmt.Printf("Error analyzing project: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result)
	},
}

var analyzeChatCmd = &cobra.Command{
	Use:   "analyze-chat [file]",
	Short: "Analyze chat history and provide insights",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading chat file: %v\n", err)
			os.Exit(1)
		}

		client := openai.NewClient(cfg.OpenAIKey)
		analyzer := analyzer.NewAnalyzer(client)

		result, err := analyzer.AnalyzeChat(context.Background(), string(content))
		if err != nil {
			fmt.Printf("Error analyzing chat: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result)
	},
}

var analyzeChatSummaryCmd = &cobra.Command{
	Use:   "analyze-chat-summary [file]",
	Short: "Analyze chat history summary and provide insights",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading summary file: %v\n", err)
			os.Exit(1)
		}

		client := openai.NewClient(cfg.OpenAIKey)
		analyzer := analyzer.NewAnalyzer(client)

		result, err := analyzer.AnalyzeChatSummary(context.Background(), string(content))
		if err != nil {
			fmt.Printf("Error analyzing chat summary: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result)
	},
}

var monitorCmd = &cobra.Command{
	Use:   "monitor [paths...]",
	Short: "Monitor files and directories for changes",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		monitor, err := monitor.NewMonitor(args)
		if err != nil {
			fmt.Printf("Error creating monitor: %v\n", err)
			os.Exit(1)
		}

		if err := monitor.Start(); err != nil {
			fmt.Printf("Error starting monitor: %v\n", err)
			os.Exit(1)
		}
		defer monitor.Stop()

		// Handle graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		fmt.Println("Monitoring files... Press Ctrl+C to stop")
		for {
			select {
			case event := <-monitor.Events():
				fmt.Printf("[%s] %s: %s\n", event.Timestamp.Format("2006-01-02 15:04:05"), event.Type, event.Path)
			case <-sigChan:
				return
			}
		}
	},
}

var screenshotCmd = &cobra.Command{
	Use:   "screenshot [display]",
	Short: "Capture a screenshot of the specified display",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		displayIndex := 0
		if _, err := fmt.Sscanf(args[0], "%d", &displayIndex); err != nil {
			fmt.Printf("Invalid display index: %v\n", err)
			os.Exit(1)
		}

		screenshot, err := screenshot.Capture(displayIndex)
		if err != nil {
			fmt.Printf("Error capturing screenshot: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Screenshot saved to: %s\n", screenshot.Path)
	},
}

var noteCmd = &cobra.Command{
	Use:   "note [content]",
	Short: "Create a new note",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		note, err := notes.NewNote(args[0])
		if err != nil {
			fmt.Printf("Error creating note: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Note created at: %s\n", note.Path)
	},
}

var listNotesCmd = &cobra.Command{
	Use:   "list-notes",
	Short: "List all notes",
	Run: func(cmd *cobra.Command, args []string) {
		notePaths, err := notes.ListNotes()
		if err != nil {
			fmt.Printf("Error listing notes: %v\n", err)
			os.Exit(1)
		}

		if len(notePaths) == 0 {
			fmt.Println("No notes found")
			return
		}

		fmt.Println("Notes:")
		for _, path := range notePaths {
			fmt.Printf("- %s\n", path)
		}
	},
}

func init() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(analyzeFileCmd)
	rootCmd.AddCommand(analyzeProjectCmd)
	rootCmd.AddCommand(analyzeChatCmd)
	rootCmd.AddCommand(analyzeChatSummaryCmd)
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(screenshotCmd)
	rootCmd.AddCommand(noteCmd)
	rootCmd.AddCommand(listNotesCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
