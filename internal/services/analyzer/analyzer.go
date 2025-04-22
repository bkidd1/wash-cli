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

The wash notes provide additional context about work and decisions. Use these to inform your analysis, but prioritize the reminders.

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
	Client         *openai.Client
	cfg            *config.Config
	projectGoal    string
	rememberNotes  []string
	sessionManager *notes.SessionManager
}

// NewTerminalAnalyzer creates a new terminal analyzer
func NewTerminalAnalyzer(apiKey string, projectGoal string, rememberNotes []string) *TerminalAnalyzer {
	client := openai.NewClient(apiKey)

	// Create wash directory if it doesn't exist
	washDir := filepath.Join(os.Getenv("HOME"), ".wash")
	if err := os.MkdirAll(washDir, 0755); err != nil {
		fmt.Printf("Warning: Could not create wash directory: %v\n", err)
	}

	sessionManager, err := notes.NewSessionManager(washDir)
	if err != nil {
		fmt.Printf("Warning: Could not create session manager: %v\n", err)
		sessionManager = nil
	}

	return &TerminalAnalyzer{
		Client: client,
		cfg: &config.Config{
			OpenAIKey: apiKey,
		},
		projectGoal:    projectGoal,
		rememberNotes:  rememberNotes,
		sessionManager: sessionManager,
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

	// Add wash notes context from the most recent session (SECOND PRIORITY)
	if session := a.sessionManager.GetCurrentSession(); session != nil {
		fmt.Printf("Current Session ID: %s\n", session.ID)
		fmt.Printf("Session Project Name: %s\n", session.ProjectName)
		fmt.Printf("Session Project Goal: %s\n", session.ProjectGoal)

		recentRecords := a.sessionManager.GetRecentRecords(session.ID, 5*time.Minute)
		fmt.Printf("Number of Recent Records: %d\n", len(recentRecords))

		if len(recentRecords) > 0 {
			context.WriteString("RECENT WASH NOTES (USE THESE TO INFORM YOUR ANALYSIS):\n")
			for _, record := range recentRecords {
				switch r := record.(type) {
				case *notes.MonitorNote:
					context.WriteString(fmt.Sprintf("- %s: %s\n", r.Timestamp.Format("2006-01-02 15:04:05"), r.Analysis.CurrentApproach))
					if len(r.Analysis.Issues) > 0 {
						context.WriteString(fmt.Sprintf("  Issues: %s\n", strings.Join(r.Analysis.Issues, ", ")))
					}
					if len(r.Analysis.Solutions) > 0 {
						context.WriteString(fmt.Sprintf("  Solutions: %s\n", strings.Join(r.Analysis.Solutions, ", ")))
					}
				case *notes.FileNote:
					context.WriteString(fmt.Sprintf("- %s: %s\n", r.Timestamp.Format("2006-01-02 15:04:05"), r.Analysis))
					if len(r.Issues) > 0 {
						context.WriteString(fmt.Sprintf("  Issues: %s\n", strings.Join(r.Issues, ", ")))
					}
				case *notes.ProjectNote:
					context.WriteString(fmt.Sprintf("- %s: Project Analysis\n", r.Timestamp.Format("2006-01-02 15:04:05")))
					if len(r.Structure.Issues) > 0 {
						context.WriteString(fmt.Sprintf("  Structure Issues: %s\n", strings.Join(r.Structure.Issues, ", ")))
					}
				case *notes.BugNote:
					context.WriteString(fmt.Sprintf("- %s: Bug Report\n", r.Timestamp.Format("2006-01-02 15:04:05")))
					context.WriteString(fmt.Sprintf("  Description: %s\n", r.Description))
					if r.Solution != "" {
						context.WriteString(fmt.Sprintf("  Solution: %s\n", r.Solution))
					}
				}
			}
			context.WriteString("\n")
		}
	} else {
		fmt.Println("No current session found")
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
