package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/notes"
	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/sashabaranov/go-openai"
)

const (
	terminalSystemPrompt = `You are an expert software architect and intermediary between a human developer and their AI coding agent. Your role is to analyze their code and interactions to identify potential issues and improvements. Especially issues that may have been caused by human error/bias misguiding the AI via poor prompts/communication.

CRITICAL: The reminders are the highest priority context. They may indicate how an issue was succesfully solved in the past.

The wash notes provide additional context about recent work and decisions. Use these to inform your analysis, but prioritize the reminders.

Focus your analysis on three priority levels:

1. Critical! Must Fix
   Security vulnerabilities
   Data corruption risks
   Performance bottlenecks
   Major architectural flaws
   Breaking changes
   Issues related to user reminders

2. Should Fix
   Code maintainability issues
   Common best practice violations
   Performance issues
   Potential future problems
   Suboptimal patterns

3. Could Fix
   Alternative tool/language recommendations
   Code style suggestions
   Documentation improvements
   Minor refactoring opportunities

Limit yourself to one "Could Fix" per response.

For each issue identified, provide a concise and clear description of the problem.

It may also be the case that the code is currently optimal and changing things would be unneeded. If no issues are found at a particular priority level, say "No issues found". Don't print any response for subcriteria if you find no issue.

DO NOT include any introductory text, summaries, conclusions or direct references to the provided context. Start directly with the priority levels and their issues.`
)

// TerminalAnalyzer represents a code analyzer that returns formatted terminal output
type TerminalAnalyzer struct {
	Client        *openai.Client
	cfg           *config.Config
	projectGoal   string
	rememberNotes []string
	notesManager  *notes.NotesManager
}

// NewTerminalAnalyzer creates a new terminal analyzer
func NewTerminalAnalyzer(apiKey string, projectGoal string, rememberNotes []string) *TerminalAnalyzer {
	client := openai.NewClient(apiKey)

	// Create wash directory if it doesn't exist
	washDir := filepath.Join(os.Getenv("HOME"), ".wash")
	if err := os.MkdirAll(washDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create wash directory: %v\n", err)
	}

	notesManager, err := notes.NewNotesManager()
	if err != nil {
		fmt.Printf("Warning: Could not create notes manager: %v\n", err)
		notesManager = nil
	}

	return &TerminalAnalyzer{
		Client: client,
		cfg: &config.Config{
			OpenAIKey: apiKey,
		},
		projectGoal:   projectGoal,
		rememberNotes: rememberNotes,
		notesManager:  notesManager,
	}
}

// UpdateProjectContext updates the project goal and remember notes
func (a *TerminalAnalyzer) UpdateProjectContext(projectGoal string, rememberNotes []string) {
	a.projectGoal = projectGoal
	a.rememberNotes = rememberNotes
}

// getContextualPrompt returns the system prompt with project context
func (a *TerminalAnalyzer) getContextualPrompt() string {
	var context strings.Builder

	fmt.Println("\n=== DEBUG: Context Data ===")
	fmt.Printf("Project Goal: %s\n", a.projectGoal)
	fmt.Printf("Remember Notes: %v\n", a.rememberNotes)

	// Add remember notes if they exist (TOP PRIORITY)
	if len(a.rememberNotes) > 0 {
		context.WriteString("CRITICAL USER REMINDERS (MUST CONSIDER THESE FIRST):\n")
		for i, note := range a.rememberNotes {
			context.WriteString(fmt.Sprintf("%d. %s\n", i+1, note))
		}
		context.WriteString("\n")
	}

	// Add recent monitor notes if available
	if a.notesManager != nil {
		// Get the current working directory name as project name
		cwd, err := os.Getwd()
		if err == nil {
			projectName := filepath.Base(cwd)
			interactions, err := a.notesManager.LoadInteractions(projectName)
			if err == nil && len(interactions) > 0 {
				context.WriteString("RECENT WASH NOTES (USE THESE TO INFORM YOUR ANALYSIS):\n")
				cutoff := time.Now().Add(-5 * time.Minute)
				for _, interaction := range interactions {
					if interaction.Timestamp.After(cutoff) {
						context.WriteString(fmt.Sprintf("- %s: %s\n", interaction.Timestamp.Format("2006-01-02 15:04:05"), interaction.Analysis.CurrentApproach))
						if len(interaction.Analysis.Issues) > 0 {
							context.WriteString(fmt.Sprintf("  Issues: %s\n", strings.Join(interaction.Analysis.Issues, ", ")))
						}
						if len(interaction.Analysis.Solutions) > 0 {
							context.WriteString(fmt.Sprintf("  Solutions: %s\n", strings.Join(interaction.Analysis.Solutions, ", ")))
						}
					}
				}
				context.WriteString("\n")
			}
		}
	}

	// Add project goal (LOWEST PRIORITY)
	context.WriteString(fmt.Sprintf("PROJECT GOAL:\n%s\n\n", a.projectGoal))

	fmt.Println("=== END DEBUG ===\n")

	return fmt.Sprintf("%s\n%s", context.String(), terminalSystemPrompt)
}

// AnalyzeFile analyzes a single file and returns formatted terminal output
func (a *TerminalAnalyzer) AnalyzeFile(ctx context.Context, filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	resp, err := a.Client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: a.getContextualPrompt(),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: string(content),
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error getting analysis: %w", err)
	}

	// Format the response with priority levels
	analysis := fmt.Sprintf(`# Code Analysis
*Generated on %s*

%s`, time.Now().Format(time.RFC3339), resp.Choices[0].Message.Content)

	return analysis, nil
}

// AnalyzeProjectStructure analyzes the project structure and returns formatted terminal output
func (a *TerminalAnalyzer) AnalyzeProjectStructure(ctx context.Context, dirPath string) (string, error) {
	var fileList strings.Builder
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip common directories
			if info.Name() == "node_modules" || info.Name() == ".git" {
				return filepath.SkipDir
			}
			fileList.WriteString(fmt.Sprintf("📁 %s\n", path))
		} else {
			relPath, _ := filepath.Rel(dirPath, path)
			fileList.WriteString(fmt.Sprintf("  📄 %s\n", relPath))
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error walking directory: %w", err)
	}

	resp, err := a.Client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: a.getContextualPrompt() + "\n\nFocus on project structure, organization, and architecture. DO NOT include any introductory text or summaries.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Project Structure:\n%s\n\nAnalyze this project structure and identify issues at each priority level. Start directly with the priority levels.", fileList.String()),
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error getting analysis: %w", err)
	}

	// Format the response with priority levels
	analysis := fmt.Sprintf(`# Project Structure Analysis
*Generated on %s*

%s`, time.Now().Format(time.RFC3339), resp.Choices[0].Message.Content)

	return analysis, nil
}

// AnalyzeChat analyzes chat history and returns formatted terminal output
func (a *TerminalAnalyzer) AnalyzeChat(ctx context.Context, chatHistory string) (string, error) {
	resp, err := a.Client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: a.getContextualPrompt() + "\n\nFocus on the interaction patterns and communication effectiveness.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: chatHistory,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error getting analysis: %w", err)
	}

	// Format the response with priority levels
	analysis := fmt.Sprintf(`# Chat Analysis
*Generated on %s*

%s`, time.Now().Format(time.RFC3339), resp.Choices[0].Message.Content)

	return analysis, nil
}

// GetErrorFix analyzes chat history for specific error patterns and returns formatted terminal output
func (a *TerminalAnalyzer) GetErrorFix(ctx context.Context, chatHistory string, errorType string) (string, error) {
	resp, err := a.Client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: a.getContextualPrompt() + fmt.Sprintf("\n\nFocus on fixing the specific error type: %s", errorType),
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: chatHistory,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error getting error fix: %w", err)
	}

	// Format the response with priority levels
	analysis := fmt.Sprintf(`# Error Fix Analysis: %s
*Generated on %s*

%s`, errorType, time.Now().Format(time.RFC3339), resp.Choices[0].Message.Content)

	return analysis, nil
}
