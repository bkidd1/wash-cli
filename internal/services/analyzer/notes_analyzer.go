package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/sashabaranov/go-openai"
)

const (
	notesSystemPrompt = `You are an expert software architect and intermediary between a human developer and their AI coding agent. Your role is to analyze their code and interactions to identify potential issues and improvements.

Focus on three levels:
1. Critical! Must Fix
2. Should Fix
3. Could Fix

For each issue identified, provide a concise (1-2 sentences) and clear description of the problem.

Do not write introduction or conclusion paragraphs. Simply return your analysis as a JSON object with the following structure:
{
    "critical_issues": ["string"],
    "should_fix": ["string"],
    "could_fix": ["string"]
}

If no issues are found at a particular priority level, return an empty array.`
)

// NotesAnalyzer represents a code analyzer that returns structured data
type NotesAnalyzer struct {
	Client        *openai.Client
	cfg           *config.Config
	projectGoal   string
	rememberNotes []string
}

// NewNotesAnalyzer creates a new notes analyzer
func NewNotesAnalyzer(apiKey string, projectGoal string, rememberNotes []string) *NotesAnalyzer {
	client := openai.NewClient(apiKey)
	return &NotesAnalyzer{
		Client: client,
		cfg: &config.Config{
			OpenAIKey: apiKey,
		},
		projectGoal:   projectGoal,
		rememberNotes: rememberNotes,
	}
}

// UpdateProjectContext updates the project goal and remember notes
func (a *NotesAnalyzer) UpdateProjectContext(projectGoal string, rememberNotes []string) {
	a.projectGoal = projectGoal
	a.rememberNotes = rememberNotes
}

// getContextualPrompt returns the system prompt with project context
func (a *NotesAnalyzer) getContextualPrompt() string {
	context := fmt.Sprintf("The user's end-goal is %s", a.projectGoal)
	if len(a.rememberNotes) > 0 {
		context += fmt.Sprintf(", and they want to remind you that:\n%s", strings.Join(a.rememberNotes, "\n"))
	}
	return fmt.Sprintf("%s\n\n%s", context, notesSystemPrompt)
}

// Analysis represents the structured analysis results
type Analysis struct {
	CriticalIssues []string `json:"critical_issues"`
	ShouldFix      []string `json:"should_fix"`
	CouldFix       []string `json:"could_fix"`
}

// AnalyzeFile analyzes a single file and returns structured analysis
func (a *NotesAnalyzer) AnalyzeFile(ctx context.Context, filePath string) (*Analysis, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
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
		return nil, fmt.Errorf("error getting analysis: %w", err)
	}

	var analysis Analysis
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
		return nil, fmt.Errorf("error parsing analysis: %w", err)
	}

	return &analysis, nil
}

// AnalyzeProjectStructure analyzes the project structure and returns structured analysis
func (a *NotesAnalyzer) AnalyzeProjectStructure(ctx context.Context, dirPath string) (*Analysis, error) {
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
		return nil, fmt.Errorf("error walking directory: %w", err)
	}

	resp, err := a.Client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: a.getContextualPrompt() + "\n\nFocus on project structure, organization, and architecture.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Project Structure:\n%s\n\nAnalyze this project structure and identify issues at each priority level.", fileList.String()),
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting analysis: %w", err)
	}

	var analysis Analysis
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
		return nil, fmt.Errorf("error parsing analysis: %w", err)
	}

	return &analysis, nil
}

// AnalyzeChat analyzes chat history and returns structured analysis
func (a *NotesAnalyzer) AnalyzeChat(ctx context.Context, chatHistory string) (*Analysis, error) {
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
		return nil, fmt.Errorf("error getting analysis: %w", err)
	}

	var analysis Analysis
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
		return nil, fmt.Errorf("error parsing analysis: %w", err)
	}

	return &analysis, nil
}

// GetErrorFix analyzes chat history for specific error patterns and returns structured analysis
func (a *NotesAnalyzer) GetErrorFix(ctx context.Context, chatHistory string, errorType string) (*Analysis, error) {
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
		return nil, fmt.Errorf("error getting error fix: %w", err)
	}

	var analysis Analysis
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
		return nil, fmt.Errorf("error parsing analysis: %w", err)
	}

	return &analysis, nil
}
