package summary

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/notes"
	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/sashabaranov/go-openai"

	"github.com/spf13/cobra"
)

const (
	// System prompt for summarization
	summaryPrompt = `You are an expert software developer and project manager reviewing the collaboration between a developer and AI coding agent. Create a concise, actionable three-paragraph summary:

1. Main activities and progress: [2-3 key technical achievements or significant changes]
2. Issues and challenges: [Only list critical blockers or important technical challenges]
3. Next steps: [2-3 specific, actionable technical tasks or improvements]

Be direct and technical. Omit obvious or minor details. Focus on what matters for project progress.`

	// Default values
	defaultAPICallDelay = 2000
	defaultMaxRetries   = 3
	defaultRetryDelay   = 1000
)

// Config holds the configuration for the summary command
type Config struct {
	APICallDelay int
	MaxRetries   int
	RetryDelay   int
}

// Command returns the summary command
func Command() *cobra.Command {
	var cfg Config

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show a summary of project progress",
		RunE:  runSummary,
	}

	// Add flags for configuration
	cmd.Flags().IntVar(&cfg.APICallDelay, "api-delay", defaultAPICallDelay, "Delay between API calls in milliseconds")
	cmd.Flags().IntVar(&cfg.MaxRetries, "max-retries", defaultMaxRetries, "Maximum number of retries for API calls")
	cmd.Flags().IntVar(&cfg.RetryDelay, "retry-delay", defaultRetryDelay, "Delay between retries in milliseconds")
	cmd.Flags().StringP("date", "d", "", "Date to show summary for (YYYY-MM-DD)")
	cmd.Flags().StringP("project", "p", "", "Project name to show summary for")

	return cmd
}

// generateSummaryWithRetry generates a summary for all notes with retry logic
func generateSummaryWithRetry(client *openai.Client, notes []*notes.ProjectProgressNote, cfg Config) (string, error) {
	var lastErr error
	for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(cfg.RetryDelay) * time.Millisecond)
		}

		summary, err := generateSummary(client, notes)
		if err == nil {
			return summary, nil
		}
		lastErr = err

		// Check if error is retryable
		if strings.Contains(err.Error(), "rate limit") ||
			strings.Contains(err.Error(), "timeout") ||
			strings.Contains(err.Error(), "connection") {
			continue
		}
		// For non-retryable errors, return immediately
		return "", err
	}
	return "", fmt.Errorf("failed after %d retries: %w", cfg.MaxRetries, lastErr)
}

// generateSummary generates a summary for all notes
func generateSummary(client *openai.Client, notes []*notes.ProjectProgressNote) (string, error) {
	var prompt strings.Builder
	prompt.WriteString("Summarize these progress notes concisely:\n\n")

	// Sort notes by timestamp (most recent first)
	sort.Slice(notes, func(i, j int) bool {
		return notes[i].Timestamp.After(notes[j].Timestamp)
	})

	for _, note := range notes {
		prompt.WriteString(fmt.Sprintf("%s: %s\n", note.Timestamp.Format("15:04"), note.Title))
		prompt.WriteString(fmt.Sprintf("%s\n", note.Description))
		if len(note.Changes.FilesModified) > 0 {
			prompt.WriteString(fmt.Sprintf("Files modified: %d\n", len(note.Changes.FilesModified)))
		}
		prompt.WriteString("---\n")
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: summaryPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt.String(),
				},
			},
			MaxTokens: 1000,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func runSummary(cmd *cobra.Command, args []string) error {
	// Get configuration from flags
	cfg := Config{
		APICallDelay: defaultAPICallDelay,
		MaxRetries:   defaultMaxRetries,
		RetryDelay:   defaultRetryDelay,
	}

	// Override defaults with flag values if provided
	if cmd.Flags().Changed("api-delay") {
		cfg.APICallDelay, _ = cmd.Flags().GetInt("api-delay")
	}
	if cmd.Flags().Changed("max-retries") {
		cfg.MaxRetries, _ = cmd.Flags().GetInt("max-retries")
	}
	if cmd.Flags().Changed("retry-delay") {
		cfg.RetryDelay, _ = cmd.Flags().GetInt("retry-delay")
	}

	// Validate configuration
	if cfg.APICallDelay < 0 {
		return fmt.Errorf("API call delay cannot be negative")
	}
	if cfg.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}
	if cfg.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}

	dateStr, _ := cmd.Flags().GetString("date")
	projectName, _ := cmd.Flags().GetString("project")

	// If no project name provided, use current directory name
	if projectName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		projectName = filepath.Base(cwd)
	}

	var targetDate time.Time
	var err error
	if dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format: %w", err)
		}
	} else {
		targetDate = time.Now()
	}

	// Get progress notes
	notesManager, err := notes.NewNotesManager()
	if err != nil {
		return fmt.Errorf("failed to initialize notes manager: %w", err)
	}
	progressNotes, err := notesManager.GetProgressNotes(projectName)
	if err != nil {
		return fmt.Errorf("failed to get progress notes: %w", err)
	}

	// Filter notes for target date
	var targetNotes []*notes.ProjectProgressNote
	for _, note := range progressNotes {
		if note.Timestamp.Year() == targetDate.Year() &&
			note.Timestamp.Month() == targetDate.Month() &&
			note.Timestamp.Day() == targetDate.Day() {
			targetNotes = append(targetNotes, note)
		}
	}

	if len(targetNotes) == 0 {
		fmt.Printf("No progress notes found for project %s on %s\n", projectName, targetDate.Format("2006-01-02"))
		return nil
	}

	// Load config to get API key
	config, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create OpenAI client with config key
	client := openai.NewClient(config.OpenAIKey)

	// Generate summary
	fmt.Println("Generating summary...")
	summary, err := generateSummaryWithRetry(client, targetNotes, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	// Print the summary
	fmt.Printf("\nProgress Summary for %s - %s\n", projectName, targetDate.Format("2006-01-02"))
	fmt.Println("------------------------")
	fmt.Println(summary)

	return nil
}
