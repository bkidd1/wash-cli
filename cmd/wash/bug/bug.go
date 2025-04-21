package bug

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bkidd1/wash-cli/internal/analyzer"
	"github.com/bkidd1/wash-cli/internal/tracker"
	"github.com/bkidd1/wash-cli/pkg/config"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

// Command creates the bug command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bug [query]",
		Short: "Get help with bugs and issues",
		Long: `Get help with bugs and issues, view related decisions, and get alternative approaches.
This command analyzes your query along with the project's context to provide targeted solutions.
It uses the history of decisions and errors to give more relevant advice.

Examples:
  wash bug "PID file not cleaning up"
  wash bug "process already finished error"
  wash bug "how to handle concurrent access"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current working directory
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			// Get user's query
			query := strings.Join(args, " ")

			// Create new bug command
			bugCmd, err := NewBugCommand(wd, query)
			if err != nil {
				return fmt.Errorf("failed to create bug command: %w", err)
			}

			// Execute the command
			return bugCmd.Execute()
		},
	}

	return cmd
}

// BugCommand represents the bug command
type BugCommand struct {
	projectPath string
	query       string
	analyzer    *analyzer.Analyzer
	state       *tracker.ProjectState
}

// NewBugCommand creates a new bug command
func NewBugCommand(projectPath string, query string) (*BugCommand, error) {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize analyzer
	analyzer := analyzer.NewAnalyzer(cfg.OpenAIKey)

	// Load project state
	state, err := tracker.NewProjectState(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize project state: %w", err)
	}

	return &BugCommand{
		projectPath: projectPath,
		query:       query,
		analyzer:    analyzer,
		state:       state,
	}, nil
}

// loadingAnimation shows a simple loading animation
func loadingAnimation(done chan bool) {
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Printf("\r") // Clear the line
			return
		default:
			fmt.Printf("\rWashing bug... %s", spinner[i])
			i = (i + 1) % len(spinner)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Execute runs the bug command
func (c *BugCommand) Execute() error {
	// Get project context from notes
	notesDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", filepath.Base(c.projectPath), "notes")
	analysisPath := filepath.Join(notesDir, "chat_analysis.txt")

	// Read the analysis file if it exists
	var projectContext string
	if content, err := os.ReadFile(analysisPath); err == nil {
		// Limit context to last 2000 characters to avoid token limits
		if len(content) > 2000 {
			projectContext = "... (earlier context omitted) ...\n" + string(content[len(content)-2000:])
		} else {
			projectContext = string(content)
		}
	}

	// Create a channel to signal when analysis is done
	done := make(chan bool)
	go loadingAnimation(done)

	// Create the system prompt for objective analysis
	systemPrompt := `You are an expert software architect and intermediary between a human developer and their AI coding agent. 
	Your role is to analyze the bug they provided to then return potential fixes/improvements. 
	Consider how issues may have been caused by human error/bias that has been misguiding the AI coding agent via poor prompts/communication. 
	Provide your objective, sometimes blunt, guidance. 

When analyzing a problem:
1. First understand what the user is trying to achieve
2. Consider if their approach is fundamentally flawed
3. Look for patterns in the project history that might indicate deeper issues. Refenerence the project's wash notes as context.
4. Provide clear, actionable alternatives that address the root cause

Format your response as follows:

## Wash Solutions
1. [Primary solution]
   - [Clear and concise explanation of why it's better than the current approach]

2. [Alternative solution]
   - [Clear and concise explanation of why it's better than the current approach]
   
 ##Wash Notes
   - [Relevant notes from the project's wash notes]`

	// Get analysis from OpenAI
	resp, err := c.analyzer.Client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
				},
				{
					Role: openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(`Project Context:
%s

User's Query:
%s

Please provide objective guidance, even if it means telling me I'm doing it wrong.`, projectContext, c.query),
				},
			},
		},
	)
	if err != nil {
		done <- true
		return fmt.Errorf("failed to get analysis: %w", err)
	}

	// Signal that analysis is complete
	done <- true

	// Print the analysis
	fmt.Println(resp.Choices[0].Message.Content)

	// Track this interaction
	decision := tracker.Decision{
		ID:             fmt.Sprintf("decision-%d", time.Now().UnixNano()),
		Timestamp:      time.Now(),
		OriginalAsk:    c.query,
		Implementation: resp.Choices[0].Message.Content,
	}
	if err := c.state.TrackDecision(decision); err != nil {
		return fmt.Errorf("failed to track decision: %w", err)
	}

	return nil
}
