package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/brinleekidd/wash-cli/internal/analyzer"
	"github.com/brinleekidd/wash-cli/internal/chat"
	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

var (
	cfg         *config.Config
	chatMonitor *chat.ChatMonitor
)

var rootCmd = &cobra.Command{
	Use:   "wash",
	Short: "Wash is a CLI tool for monitoring and analyzing Cursor chat interactions",
	Long: `Wash is a command-line tool that helps monitor and analyze your interactions with Cursor's chat console.
It can identify potential mistakes or misguidance in your chat history to help you understand where you might have gone wrong.`,
}

var fileCmd = &cobra.Command{
	Use:   "file [file]",
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

var structureCmd = &cobra.Command{
	Use:   "structure [directory]",
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

var chatSummaryCmd = &cobra.Command{
	Use:   "chat summary",
	Short: "Analyze chat history and provide insights",
	Run: func(cmd *cobra.Command, args []string) {
		client := openai.NewClient(cfg.OpenAIKey)
		analyzer := analyzer.NewAnalyzer(client)

		// Read the chat analysis file
		notesDir := filepath.Join(os.Getenv("HOME"), ".wash", "notes")
		analysisPath := filepath.Join(notesDir, "chat_analysis.txt")

		content, err := os.ReadFile(analysisPath)
		if err != nil {
			fmt.Printf("Error reading chat analysis file: %v\n", err)
			os.Exit(1)
		}

		// Get a summary of the analysis
		result, err := analyzer.AnalyzeChatSummary(context.Background(), string(content))
		if err != nil {
			fmt.Printf("Error analyzing chat summary: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(result)
	},
}

var startChatCmd = &cobra.Command{
	Use:   "start chat",
	Short: "Start monitoring Cursor chat console",
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		chatMonitor, err = chat.NewChatMonitor(cfg)
		if err != nil {
			fmt.Printf("Error creating chat monitor: %v\n", err)
			os.Exit(1)
		}

		// Set up signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Start the chat monitor in a goroutine
		go func() {
			if err := chatMonitor.Start(); err != nil {
				fmt.Printf("Error starting chat monitor: %v\n", err)
				os.Exit(1)
			}
		}()

		fmt.Println("Chat monitoring started. Press Ctrl+C to stop.")

		// Wait for interrupt signal
		<-sigChan
		fmt.Println("\nReceived interrupt signal. Stopping chat monitor...")

		if err := chatMonitor.Stop(); err != nil {
			fmt.Printf("Error stopping chat monitor: %v\n", err)
			os.Exit(1)
		}
	},
}

var stopChatCmd = &cobra.Command{
	Use:   "stop chat",
	Short: "Stop monitoring Cursor chat console",
	Run: func(cmd *cobra.Command, args []string) {
		if chatMonitor == nil {
			fmt.Println("Chat monitor is not running")
			return
		}

		if err := chatMonitor.Stop(); err != nil {
			fmt.Printf("Error stopping chat monitor: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Chat monitoring stopped")
	},
}

func init() {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(fileCmd)
	rootCmd.AddCommand(structureCmd)
	rootCmd.AddCommand(chatSummaryCmd)
	rootCmd.AddCommand(startChatCmd)
	rootCmd.AddCommand(stopChatCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
