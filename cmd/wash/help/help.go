package help

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brinleekidd/wash-cli/internal/analyzer"
	"github.com/brinleekidd/wash-cli/internal/tracker"
	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

// Command creates the help command
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [query]",
		Short: "Get help with errors and issues",
		Long: `Get help with errors and issues, view related decisions, and get alternative approaches.
This command is particularly useful when you're stuck on an error or want to improve your
last code change.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current working directory
			wd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			// Get user's query
			query := strings.Join(args, " ")

			// Create new help command
			helpCmd, err := NewHelpCommand(wd, query)
			if err != nil {
				return fmt.Errorf("failed to create help command: %w", err)
			}

			// Execute the command
			return helpCmd.Execute()
		},
	}

	return cmd
}

// HelpCommand represents the help command
type HelpCommand struct {
	projectPath string
	query       string
	analyzer    *analyzer.Analyzer
	state       *tracker.ProjectState
}

// NewHelpCommand creates a new help command
func NewHelpCommand(projectPath string, query string) (*HelpCommand, error) {
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

	return &HelpCommand{
		projectPath: projectPath,
		query:       query,
		analyzer:    analyzer,
		state:       state,
	}, nil
}

// Execute runs the help command
func (c *HelpCommand) Execute() error {
	// Get project context from notes
	notesDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", filepath.Base(c.projectPath), "notes")
	analysisPath := filepath.Join(notesDir, "chat_analysis.txt")

	// Read the analysis file if it exists
	var projectContext string
	if content, err := os.ReadFile(analysisPath); err == nil {
		projectContext = string(content)
	}

	// Create the system prompt for objective analysis
	systemPrompt := `You are an expert software architect and development assistant. Your role is to provide objective, sometimes blunt, guidance to developers. 

When analyzing a problem:
1. First understand what the user is trying to achieve
2. Consider if their approach is fundamentally flawed
3. Look for patterns in the project history that might indicate deeper issues
4. Don't be afraid to tell them they're doing it wrong if that's the case
5. Provide clear, actionable alternatives that address the root cause

Format your response as follows:

# Analysis: [Brief description of the core issue]

## Current Approach
[What the user is trying to do, even if it's wrong]

## Why This Might Not Work
[Objective reasons why their approach might fail or be suboptimal]

## Better Solutions
1. [Primary solution]
   - Why it's better: [Clear explanation]
   - Implementation steps: [Step-by-step guide]
   - Potential challenges: [What to watch out for]

2. [Alternative solution]
   - Why it's better: [Clear explanation]
   - Implementation steps: [Step-by-step guide]
   - Potential challenges: [What to watch out for]

## Project Context
[How this fits into the broader project history and architecture]

## Technical Considerations
- [Important technical detail 1]
- [Important technical detail 2]

## Best Practices
- [Relevant best practice 1]
- [Relevant best practice 2]`

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
		return fmt.Errorf("failed to get analysis: %w", err)
	}

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
