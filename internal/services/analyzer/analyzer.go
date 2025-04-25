package analyzer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bkidd1/wash-cli/internal/services/notes"
	"github.com/sashabaranov/go-openai"
)

const (
	terminalSystemPrompt = "You are an expert software architect and intermediary between a human developer and their AI coding agent. Your role is to analyze their code and interactions to identify potential issues and improvements. Especially issues that may have been caused by human error/bias misguiding the AI via poor prompts/communication.\n\n" +
		"CRITICAL: The reminders are the highest priority context. They usually indicate how an issue was succesfully solved in the past - or how the user prefers to solve issues. AS LONG AS THEY ARE RELEVENT TO THE ISSUE AT HAND, you should consider them first.\n\n" +
		"The wash notes also provide additional context about recent work and decisions. Use these to inform your analysis as well if relevant to the issue at hand.\n\n" +
		"Focus your analysis on three priority levels:\n\n" +
		"1. Critical! Must Fix\n" +
		"   Security vulnerabilities\n" +
		"   Data corruption risks\n" +
		"   Performance bottlenecks\n" +
		"   Major architectural flaws\n" +
		"   Breaking changes\n" +
		"   Issues related to user reminders\n\n" +
		"2. Should Fix\n" +
		"   Code maintainability issues\n" +
		"   Common best practice violations\n" +
		"   Performance issues\n" +
		"   Potential future problems\n" +
		"   Suboptimal patterns\n\n" +
		"3. Could Fix\n" +
		"   Alternative tool/language recommendations\n" +
		"   Code style suggestions\n" +
		"   Documentation improvements\n" +
		"   Minor refactoring opportunities\n\n" +
		"Limit yourself to one \"Could Fix\" per response.\n\n" +
		"For each issue identified, provide a concise and clear description of the problem.\n\n" +
		"It may also be the case that the code is currently optimal and changing things would be unneeded. If no issues are found at a particular priority level, say \"No issues found\". Don't print any response for subcriteria if you find no issue.\n\n" +
		"DO NOT include any introductory text, summaries, conclusions or direct references to the provided context. Start directly with the priority levels and their issues."
)

// TerminalAnalyzer represents a code analyzer that returns formatted terminal output
type TerminalAnalyzer struct {
	client        *openai.Client
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
		client:        client,
		projectGoal:   projectGoal,
		rememberNotes: rememberNotes,
		notesManager:  notesManager,
	}
}

// UpdateProjectContext updates the project goal
func (a *TerminalAnalyzer) UpdateProjectContext(projectGoal string) {
	a.projectGoal = projectGoal
}

// getContextualPrompt returns the system prompt with project context
func (a *TerminalAnalyzer) getContextualPrompt() string {
	var context strings.Builder

	// Add the system prompt
	context.WriteString(terminalSystemPrompt)
	context.WriteString("\n\n")

	// Add recent monitor notes if available
	if a.notesManager != nil {
		// Get the current working directory name as project name
		cwd, err := os.Getwd()
		if err == nil {
			projectName := filepath.Base(cwd)
			// Get recent monitor notes
			monitorDir := a.notesManager.GetMonitorNotesDir(projectName)

			// Create monitor directory if it doesn't exist
			if err := os.MkdirAll(monitorDir, 0755); err != nil {
				fmt.Printf("Warning: Could not create monitor directory: %v\n", err)
			} else {
				files, err := os.ReadDir(monitorDir)
				if err == nil {
					context.WriteString("RECENT WASH NOTES (USE THESE TO INFORM YOUR ANALYSIS):\n")
					cutoff := time.Now().Add(-5 * time.Minute)

					// Read files in reverse chronological order
					for i := len(files) - 1; i >= 0; i-- {
						file := files[i]
						if filepath.Ext(file.Name()) != ".json" {
							continue
						}

						data, err := os.ReadFile(filepath.Join(monitorDir, file.Name()))
						if err != nil {
							continue
						}

						var note notes.MonitorNote
						if err := json.Unmarshal(data, &note); err != nil {
							continue
						}

						if note.Timestamp.After(cutoff) {
							context.WriteString(fmt.Sprintf("- %s: User asked '%s'\n", note.Timestamp.Format("2006-01-02 15:04:05"), note.Interaction.UserRequest))
							context.WriteString(fmt.Sprintf("  AI responded: %s\n", note.Interaction.AIAction))
						}
					}
					context.WriteString("\n")
				}
			}
		}
	}

	// Add project goal
	context.WriteString(fmt.Sprintf("PROJECT GOAL:\n%s\n\n", a.projectGoal))

	return context.String()
}

// AnalyzeFile analyzes a single file and returns formatted terminal output
func (a *TerminalAnalyzer) AnalyzeFile(ctx context.Context, filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	resp, err := a.client.CreateChatCompletion(
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

	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: a.getContextualPrompt() + "\n\nFocus on project structure, organization, and architecture. Format your response EXACTLY as follows:\n\n1. Critical! Must Fix\n[list critical issues here]\n\n2. Should Fix\n[list should fix issues here]\n\n3. Could Fix\n[list could fix issues here]\n\nDo not include any other sections or text.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Project Structure:\n%s\n\nAnalyze this project structure and identify issues at each priority level.", fileList.String()),
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
	resp, err := a.client.CreateChatCompletion(
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
	resp, err := a.client.CreateChatCompletion(
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

// BugAnalysis represents the analysis of a bug
type BugAnalysis struct {
	Analysis           string
	PotentialCauses    string
	SuggestedSolutions string
	RelatedContext     string
}

// AnalyzeBug analyzes a bug description and provides potential solutions
func (a *TerminalAnalyzer) AnalyzeBug(ctx context.Context, description string) (*BugAnalysis, error) {
	// Get project context from remember notes
	contextPrompt := a.getContextualPrompt()

	// Add remember notes to the context if they exist
	if len(a.rememberNotes) > 0 {
		contextPrompt += "\n\nREMEMBER NOTES (USE THESE TO INFORM YOUR ANALYSIS):\n"
		for _, note := range a.rememberNotes {
			contextPrompt += fmt.Sprintf("- %s\n", note)
		}
	}

	// Create chat completion request
	resp, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: "gpt-4",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: contextPrompt + "\n\nFor bug analysis, you MUST format your response EXACTLY as follows:\n\n# Potential Causes\n[list potential causes here]\n\n# Suggested Solutions\n[list suggested solutions here]\n\nDo not include any other sections or text.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Bug description: %s", description),
				},
			},
			MaxTokens: 1000,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze bug: %w", err)
	}

	// Parse the response into sections
	content := resp.Choices[0].Message.Content
	sections := parseSections(content)

	return &BugAnalysis{
		Analysis:           "",
		PotentialCauses:    sections["Potential Causes"],
		SuggestedSolutions: sections["Suggested Solutions"],
		RelatedContext:     "",
	}, nil
}

// parseSections splits the AI response into sections
func parseSections(content string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(content, "\n")

	currentSection := ""
	var currentContent strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			// If we were building a previous section, save it
			if currentSection != "" {
				sections[strings.TrimSpace(currentSection)] = strings.TrimSpace(currentContent.String())
				currentContent.Reset()
			}
			// Extract new section name (remove # and any leading/trailing spaces)
			currentSection = strings.TrimSpace(strings.TrimPrefix(line, "#"))
		} else if currentSection != "" {
			// Add the line to the current section
			currentContent.WriteString(line + "\n")
		}
	}

	// Save the last section
	if currentSection != "" {
		sections[strings.TrimSpace(currentSection)] = strings.TrimSpace(currentContent.String())
	}

	// Ensure required sections exist with meaningful defaults
	requiredSections := []string{"Potential Causes", "Suggested Solutions"}
	for _, section := range requiredSections {
		if content, exists := sections[section]; !exists || strings.TrimSpace(content) == "" {
			sections[section] = fmt.Sprintf("No %s information provided by the analysis", strings.ToLower(section))
		}
	}

	return sections
}

// GetProjectGoal returns the project goal
func (a *TerminalAnalyzer) GetProjectGoal() string {
	return a.projectGoal
}

// GetRememberNotes returns the remember notes
func (a *TerminalAnalyzer) GetRememberNotes() []string {
	return a.rememberNotes
}
