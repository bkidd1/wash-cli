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
	// Default values
	defaultMaxBatchSize = 1
	defaultAPICallDelay = 2000
	defaultMaxRetries   = 3
	defaultRetryDelay   = 1000

	// System prompt for the initial summarization
	batchSystemPrompt = `Summarize these notes in 2-3 sentences max:
1. Main activities, key decisions, general progress.
2. Possible errors/mistakes - non-optimal decisions made in the chat.
3. Potential alternative approaches to the errors/mistakes.
Be extremely brief.`
	// System prompt for combining summaries
	combineSummaryPrompt = `Combine these summaries into a three paragraph summary for %s.
Structure your summary like this:
1. Descriptive summary of specific activities, key decisions, general progress.
2. The errors/mistakes - non-optimal decisions made in the chat (if any).
3. Alternative approaches to the errors/mistakes (if any).
In addition to the paragraphs, include a list of files modified (if specificallydocumented in the notes)
Be concise and specific.`
)

// Config holds the configuration for the summary command
type Config struct {
	MaxBatchSize int
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
	cmd.Flags().IntVar(&cfg.MaxBatchSize, "batch-size", defaultMaxBatchSize, "Maximum number of notes to process in a single batch")
	cmd.Flags().IntVar(&cfg.APICallDelay, "api-delay", defaultAPICallDelay, "Delay between API calls in milliseconds")
	cmd.Flags().IntVar(&cfg.MaxRetries, "max-retries", defaultMaxRetries, "Maximum number of retries for API calls")
	cmd.Flags().IntVar(&cfg.RetryDelay, "retry-delay", defaultRetryDelay, "Delay between retries in milliseconds")
	cmd.Flags().StringP("date", "d", "", "Date to show summary for (YYYY-MM-DD)")
	cmd.Flags().StringP("project", "p", "", "Project name to show summary for")

	return cmd
}

// processBatchWithRetry generates a summary for a batch of notes with retry logic
func processBatchWithRetry(client *openai.Client, notes []*notes.ProjectProgressNote, cfg Config) (string, error) {
	var lastErr error
	for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(cfg.RetryDelay) * time.Millisecond)
		}

		summary, err := processBatch(client, notes)
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

// processBatch generates a summary for a batch of notes
func processBatch(client *openai.Client, notes []*notes.ProjectProgressNote) (string, error) {
	var prompt strings.Builder
	prompt.WriteString("Summarize these notes concisely:\n\n")

	for _, note := range notes {
		// Only include essential information
		prompt.WriteString(fmt.Sprintf("%s: %s\n", note.Timestamp.Format("15:04"), note.Title))

		// Truncate description if too long
		desc := note.Description
		if len(desc) > 200 {
			desc = desc[:200] + "..."
		}
		prompt.WriteString(fmt.Sprintf("%s\n", desc))

		// Only include file count if there are changes
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
					Content: batchSystemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt.String(),
				},
			},
			MaxTokens: 500,
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate batch summary: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

// combineSummariesWithRetry combines multiple batch summaries into a final summary with retry logic
func combineSummariesWithRetry(client *openai.Client, summaries []string, date time.Time, cfg Config) (string, error) {
	var lastErr error
	for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(cfg.RetryDelay) * time.Millisecond)
		}

		summary, err := combineSummaries(client, summaries, date)
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

// combineSummaries combines multiple batch summaries into a final summary
func combineSummaries(client *openai.Client, summaries []string, date time.Time) (string, error) {
	var prompt strings.Builder
	prompt.WriteString(fmt.Sprintf("Combine these summaries for %s:\n\n", date.Format("2006-01-02")))

	for i, summary := range summaries {
		prompt.WriteString(fmt.Sprintf("Summary %d:\n%s\n---\n", i+1, summary))
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: fmt.Sprintf(combineSummaryPrompt, date.Format("2006-01-02")),
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
		return "", fmt.Errorf("failed to combine summaries: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

func runSummary(cmd *cobra.Command, args []string) error {
	// Get configuration from flags
	cfg := Config{
		MaxBatchSize: defaultMaxBatchSize,
		APICallDelay: defaultAPICallDelay,
		MaxRetries:   defaultMaxRetries,
		RetryDelay:   defaultRetryDelay,
	}

	// Override defaults with flag values if provided
	if cmd.Flags().Changed("batch-size") {
		cfg.MaxBatchSize, _ = cmd.Flags().GetInt("batch-size")
	}
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
	if cfg.MaxBatchSize < 1 {
		return fmt.Errorf("batch size must be at least 1")
	}
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

	// Filter notes for target date and sort by timestamp
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

	// Sort notes by timestamp
	sort.Slice(targetNotes, func(i, j int) bool {
		return targetNotes[i].Timestamp.Before(targetNotes[j].Timestamp)
	})

	// Load config to get API key
	config, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create OpenAI client with config key
	client := openai.NewClient(config.OpenAIKey)

	// Process notes in batches
	var batchSummaries []string
	for i := 0; i < len(targetNotes); i += cfg.MaxBatchSize {
		end := i + cfg.MaxBatchSize
		if end > len(targetNotes) {
			end = len(targetNotes)
		}

		fmt.Printf("Processing notes %d-%d of %d...\n", i+1, end, len(targetNotes))
		summary, err := processBatchWithRetry(client, targetNotes[i:end], cfg)
		if err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}
		batchSummaries = append(batchSummaries, summary)

		// Add delay between API calls
		time.Sleep(time.Duration(cfg.APICallDelay) * time.Millisecond)
	}

	// Add delay before final summary
	time.Sleep(time.Duration(cfg.APICallDelay) * time.Millisecond)

	// Combine all summaries
	fmt.Println("Generating final summary...")
	finalSummary, err := combineSummariesWithRetry(client, batchSummaries, targetDate, cfg)
	if err != nil {
		return fmt.Errorf("failed to generate final summary: %w", err)
	}

	// Print the summary
	fmt.Printf("\nProgress Summary for %s - %s\n", projectName, targetDate.Format("2006-01-02"))
	fmt.Println("------------------------")
	fmt.Println(finalSummary)

	return nil
}
