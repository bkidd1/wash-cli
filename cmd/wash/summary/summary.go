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
	"github.com/sashabaranov/go-openai"

	"github.com/spf13/cobra"
)

const (
	// Maximum number of notes to process in a single batch
	maxBatchSize = 1
	// Delay between API calls in milliseconds
	apiCallDelay = 2000
	// System prompt for the initial summarization
	batchSystemPrompt = `Summarize these progress notes concisely:
1. Key activities/achievements
2. Challenges/solutions
3. Important decisions
Be brief and factual.`
	// System prompt for combining summaries
	combineSummaryPrompt = `Combine these summaries into a daily summary for %s.
Focus on:
1. Key achievements
2. Challenges/resolutions
3. Important decisions
Be concise.`
)

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show a summary of project progress",
		RunE:  runSummary,
	}

	cmd.Flags().StringP("date", "d", "", "Date to show summary for (YYYY-MM-DD)")
	cmd.Flags().StringP("project", "p", "", "Project name to show summary for")

	return cmd
}

// processBatch generates a summary for a batch of notes
func processBatch(client *openai.Client, notes []*notes.ProjectProgressNote) (string, error) {
	var prompt strings.Builder
	prompt.WriteString("Summarize these notes:\n\n")

	for _, note := range notes {
		prompt.WriteString(fmt.Sprintf("%s: %s\n", note.Timestamp.Format("15:04"), note.Title))
		prompt.WriteString(fmt.Sprintf("%s\n", note.Description))
		if len(note.Changes.FilesModified) > 0 {
			prompt.WriteString("Files: ")
			for i, file := range note.Changes.FilesModified {
				if i > 0 {
					prompt.WriteString(", ")
				}
				prompt.WriteString(filepath.Base(file))
			}
			prompt.WriteString("\n")
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
	dateStr, _ := cmd.Flags().GetString("date")
	projectName, _ := cmd.Flags().GetString("project")

	// If no project name provided, use current directory name
	if projectName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %v", err)
		}
		projectName = filepath.Base(cwd)
	}

	var targetDate time.Time
	var err error
	if dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format: %v", err)
		}
	} else {
		targetDate = time.Now()
	}

	// Get progress notes
	notesManager, err := notes.NewNotesManager()
	if err != nil {
		return fmt.Errorf("failed to initialize notes manager: %v", err)
	}
	progressNotes, err := notesManager.GetProgressNotes(projectName)
	if err != nil {
		return fmt.Errorf("failed to get progress notes: %v", err)
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

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Create OpenAI client
	client := openai.NewClient(apiKey)

	// Process notes in batches
	var batchSummaries []string
	for i := 0; i < len(targetNotes); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(targetNotes) {
			end = len(targetNotes)
		}

		fmt.Printf("Processing notes %d-%d of %d...\n", i+1, end, len(targetNotes))
		summary, err := processBatch(client, targetNotes[i:end])
		if err != nil {
			return fmt.Errorf("failed to process batch: %w", err)
		}
		batchSummaries = append(batchSummaries, summary)

		// Add delay between API calls
		time.Sleep(apiCallDelay * time.Millisecond)
	}

	// Add delay before final summary
	time.Sleep(apiCallDelay * time.Millisecond)

	// Combine all summaries
	fmt.Println("Generating final summary...")
	finalSummary, err := combineSummaries(client, batchSummaries, targetDate)
	if err != nil {
		return fmt.Errorf("failed to generate final summary: %w", err)
	}

	// Print the summary
	fmt.Printf("\nProgress Summary for %s - %s\n", projectName, targetDate.Format("2006-01-02"))
	fmt.Println("------------------------")
	fmt.Println(finalSummary)

	return nil
}
