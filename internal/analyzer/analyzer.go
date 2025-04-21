package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/sashabaranov/go-openai"
)

const (
	systemPrompt = `You are an expert software architect and intermediary between a human developer and their AI coding agent. Your role is to analyze their code and interactions to identify potential issues and improvements. Especially issues that may have been caused by human error/bias misguiding the AI via poor prompts/communication. Focus on three priority levels:

1. Critical! Must Fix
   Security vulnerabilities
   Data corruption risks
   Performance bottlenecks
   Major architectural flaws
   Breaking changes

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

It may also be the case that the code is currently optimal and changing things would be unneeded. If no issues are found at a particular priority level, say "No issues found".

DO NOT include any introductory text, summaries, or conclusions. Start directly with the priority levels and their issues.`
)

// Analyzer represents a code analyzer
type Analyzer struct {
	Client *openai.Client
	cfg    *config.Config
}

// NewAnalyzer creates a new code analyzer
func NewAnalyzer(apiKey string) *Analyzer {
	client := openai.NewClient(apiKey)
	return &Analyzer{
		Client: client,
		cfg: &config.Config{
			OpenAIKey: apiKey,
		},
	}
}

// AnalyzeFile analyzes a single file for potential optimizations and improvements
func (a *Analyzer) AnalyzeFile(ctx context.Context, filePath string) (string, error) {
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
					Content: systemPrompt,
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

// AnalyzeProjectStructure analyzes the project structure and suggests improvements
func (a *Analyzer) AnalyzeProjectStructure(ctx context.Context, dirPath string) (string, error) {
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
					Content: systemPrompt + "\n\nFocus on project structure, organization, and architecture. DO NOT include any introductory text or summaries.",
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

// AnalyzeChat analyzes chat history and provides insights
func (a *Analyzer) AnalyzeChat(ctx context.Context, chatHistory string) (string, error) {
	resp, err := a.Client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt + "\n\nFocus on the interaction patterns and communication effectiveness.",
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

// AnalyzeChatSummary analyzes chat history summaries and provides insights
func (a *Analyzer) AnalyzeChatSummary(ctx context.Context, summary string) (string, error) {
	resp, err := a.Client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt + "\n\nFocus on the overall interaction patterns and long-term improvements.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: summary,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error getting analysis: %w", err)
	}

	// Format the response with priority levels
	analysis := fmt.Sprintf(`# Summary Analysis
*Generated on %s*

%s`, time.Now().Format(time.RFC3339), resp.Choices[0].Message.Content)

	return analysis, nil
}

// GetErrorFix analyzes chat history for specific error patterns and provides solutions
func (a *Analyzer) GetErrorFix(ctx context.Context, chatHistory string, errorType string) (string, error) {
	resp, err := a.Client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: openai.GPT4,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt + fmt.Sprintf("\n\nFocus on fixing the specific error type: %s", errorType),
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
