package summary

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/notes"
	"github.com/sashabaranov/go-openai"

	"github.com/spf13/cobra"
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

	// Prepare the prompt for AI analysis
	var prompt string
	prompt = fmt.Sprintf("Please analyze these progress notes and provide a concise summary of the day's work. Focus on:\n\n")
	prompt += "1. Key achievements and progress made\n"
	prompt += "2. Any challenges or issues encountered\n"
	prompt += "3. Important decisions or changes\n\n"
	prompt += "Progress Notes:\n\n"

	for _, note := range targetNotes {
		prompt += fmt.Sprintf("Title: %s\n", note.Title)
		prompt += fmt.Sprintf("Description: %s\n", note.Description)
		if len(note.Changes.FilesModified) > 0 {
			prompt += "Files Modified:\n"
			for _, file := range note.Changes.FilesModified {
				prompt += fmt.Sprintf("- %s\n", file)
			}
		}
		prompt += fmt.Sprintf("Risk Level: %s\n", note.Impact.RiskLevel)
		prompt += "---\n"
	}

	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	// Create OpenAI client
	client := openai.NewClient(apiKey)

	// Generate summary using AI
	resp, err := client.CreateChatCompletion(
		cmd.Context(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant that summarizes project progress notes in a clear and concise manner.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %v", err)
	}

	// Print the summary
	fmt.Printf("\nProgress Summary for %s - %s\n", projectName, targetDate.Format("2006-01-02"))
	fmt.Println("------------------------")
	fmt.Println(resp.Choices[0].Message.Content)

	return nil
}
