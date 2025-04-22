package chatmonitor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/notes"
	"github.com/sashabaranov/go-openai"
)

type Analyzer struct {
	client         *openai.Client
	sessionManager *notes.SessionManager
}

func NewAnalyzer(client *openai.Client, sessionManager *notes.SessionManager) *Analyzer {
	return &Analyzer{
		client:         client,
		sessionManager: sessionManager,
	}
}

// formatContextForAI formats recent records into a context string for the AI
func formatContextForAI(records []interface{}) string {
	if len(records) == 0 {
		return "No recent context available."
	}

	var context strings.Builder
	context.WriteString("Recent context from the session:\n\n")

	for _, record := range records {
		switch r := record.(type) {
		case *notes.Interaction:
			context.WriteString(fmt.Sprintf("Interaction at %s:\n", r.Timestamp.Format("2006-01-02 15:04:05")))
			context.WriteString(fmt.Sprintf("Context: %s\n", r.Context.CurrentState))
			if len(r.Context.FilesChanged) > 0 {
				context.WriteString(fmt.Sprintf("Files Changed: %s\n", strings.Join(r.Context.FilesChanged, ", ")))
			}
			context.WriteString(fmt.Sprintf("Analysis: %s\n", r.Analysis.CurrentApproach))
			if len(r.Analysis.Issues) > 0 {
				context.WriteString(fmt.Sprintf("Issues: %s\n", strings.Join(r.Analysis.Issues, ", ")))
			}
			if len(r.Analysis.Solutions) > 0 {
				context.WriteString(fmt.Sprintf("Solutions: %s\n", strings.Join(r.Analysis.Solutions, ", ")))
			}
			context.WriteString("\n")
		case *notes.CodeChange:
			context.WriteString(fmt.Sprintf("Code Change at %s:\n", r.Timestamp.Format("2006-01-02 15:04:05")))
			context.WriteString(fmt.Sprintf("File: %s\n", r.File))
			context.WriteString(fmt.Sprintf("Description: %s\n", r.Description))
			if len(r.PotentialIssues) > 0 {
				context.WriteString(fmt.Sprintf("Potential Issues: %s\n", strings.Join(r.PotentialIssues, ", ")))
			}
			context.WriteString("\n")
		}
	}

	return context.String()
}

func (a *Analyzer) AnalyzeWithContext(screenshotPath string) (*notes.Interaction, error) {
	// Get current session
	session := a.sessionManager.GetCurrentSession()
	if session == nil {
		return nil, fmt.Errorf("no active session")
	}

	// Get recent records from the last 5 minutes
	recentRecords := a.sessionManager.GetRecentRecords(session.ID, 5*time.Minute)
	contextStr := formatContextForAI(recentRecords)

	// Read the screenshot
	imageBytes, err := os.ReadFile(screenshotPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read screenshot: %v", err)
	}

	// Create base64 encoded image
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)

	// Create the analysis prompt with context
	prompt := fmt.Sprintf(`You are an expert software architect and intermediary between a human developer and their AI coding agent. 
Your role is to analyze the chat interactions in the provided window screenshots and do two things:
1. Identify potential issues and improvements, and record better solutions. Especially issues that have been caused by human error/bias misguiding the AI via poor prompts/communication.
2. Document best practices they use and the solutions to how they fix bugs.

Recent context from the session:
%s

Based on this context and the current screenshot, please analyze the interaction and provide:
1. Current approach being taken
2. Any potential issues or improvements
3. Better solutions or approaches
4. Best practices observed

Format your response as a JSON object with the following structure:
{
    "current_approach": "string",
    "issues": ["string"],
    "solutions": ["string"],
    "best_practices": ["string"]
}`, contextStr)

	// Create the chat completion request
	resp, err := a.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4-vision-preview",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "user",
					Content: prompt,
					MultiContent: []openai.ChatMessagePart{
						{
							Type: "text",
							Text: prompt,
						},
						{
							Type: "image_url",
							ImageURL: &openai.ChatMessageImageURL{
								URL: fmt.Sprintf("data:image/png;base64,%s", base64Image),
							},
						},
					},
				},
			},
			MaxTokens: 1000,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %v", err)
	}

	// Parse the response into an Interaction struct
	var analysis struct {
		CurrentApproach string   `json:"current_approach"`
		Issues          []string `json:"issues"`
		Solutions       []string `json:"solutions"`
		BestPractices   []string `json:"best_practices"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis response: %v", err)
	}

	// Create and return the interaction
	interaction := &notes.Interaction{
		Timestamp:   time.Now(),
		ProjectName: session.ProjectName,
		ProjectGoal: session.ProjectGoal,
		Context: struct {
			CurrentState string   `json:"current_state"`
			FilesChanged []string `json:"files_changed,omitempty"`
		}{
			CurrentState: "Analyzing chat interaction",
		},
		Analysis: struct {
			CurrentApproach string   `json:"current_approach"`
			Issues          []string `json:"issues,omitempty"`
			Solutions       []string `json:"solutions,omitempty"`
			BestPractices   []string `json:"best_practices,omitempty"`
		}{
			CurrentApproach: analysis.CurrentApproach,
			Issues:          analysis.Issues,
			Solutions:       analysis.Solutions,
			BestPractices:   analysis.BestPractices,
		},
	}

	// Add the interaction to the session
	if err := a.sessionManager.AddRecord(interaction); err != nil {
		return nil, fmt.Errorf("failed to add interaction to session: %v", err)
	}

	return interaction, nil
}
