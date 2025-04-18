package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// Analyzer represents a code analyzer
type Analyzer struct {
	client *openai.Client
}

// NewAnalyzer creates a new code analyzer
func NewAnalyzer(client *openai.Client) *Analyzer {
	return &Analyzer{
		client: client,
	}
}

// AnalyzeFile analyzes a single file for potential optimizations and improvements
func (a *Analyzer) AnalyzeFile(ctx context.Context, filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	systemPrompt := `You are an expert software architect and Cursor AI assistant. Before suggesting any changes, carefully analyze the provided code and ask yourself:

1. Is the current implementation already optimal?
   - Does it follow best practices?
   - Is it performant and maintainable?
   - Are there any actual issues that need addressing?

2. Would refactoring provide meaningful benefits?
   - Would the benefits outweigh the risks of change?
   - Is the current solution actually the best approach?
   - Are there simpler alternatives that would work as well?

If the current implementation is already optimal, acknowledge this and explain why. If changes are needed, provide clear, step-by-step instructions for Cursor's AI to implement improvements.`

	resp, err := a.client.CreateChatCompletion(
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

	return resp.Choices[0].Message.Content, nil
}

// AnalyzeProjectStructure analyzes the project structure and suggests improvements
func (a *Analyzer) AnalyzeProjectStructure(ctx context.Context, dirPath string) (string, error) {
	var fileList strings.Builder
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(dirPath, path)
			fileList.WriteString(fmt.Sprintf("- %s\n", relPath))
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error walking directory: %w", err)
	}

	systemPrompt := `You are an expert software architect. Analyze the provided project structure and provide insights about:
1. Overall project organization
2. Potential improvements in file/directory structure
3. Missing or redundant components
4. Best practices and recommendations

Format your response in a clear, structured way with sections for each aspect.`

	resp, err := a.client.CreateChatCompletion(
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
					Content: fileList.String(),
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error getting analysis: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

// AnalyzeChat analyzes chat history and provides insights
func (a *Analyzer) AnalyzeChat(ctx context.Context, chatHistory string) (string, error) {
	systemPrompt := `You are an expert AI assistant analyzing ongoing chat history. Your task is to:
1. Identify key discussion points and decisions made
2. Extract actionable suggestions for code improvements
3. Generate clear, copy-pasteable prompts that the user can use in Cursor chat
4. Note any patterns in communication that could be improved
5. Track progress on major tasks and decisions

Format your response as follows:

KEY POINTS:
- List main discussion topics and decisions
- Highlight any blockers or challenges identified

ACTIONABLE SUGGESTIONS:
- Provide specific prompts that the user can copy into Cursor chat
- Focus on architectural improvements, best practices, and optimization opportunities
- Each suggestion should be a complete, self-contained prompt

COMMUNICATION PATTERNS:
- Note any recurring issues in how requests are framed
- Suggest better ways to phrase questions or requests
- Highlight successful communication strategies

PROGRESS TRACKING:
- Summarize what has been accomplished
- List pending items or next steps
- Note any dependencies or prerequisites

DO NOT generate code directly. Instead, provide prompts that the user can copy into Cursor chat to get the desired code.`

	resp, err := a.client.CreateChatCompletion(
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
					Content: chatHistory,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error getting analysis: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}

// AnalyzeChatSummary analyzes chat history summaries and provides insights
func (a *Analyzer) AnalyzeChatSummary(ctx context.Context, summary string) (string, error) {
	systemPrompt := `You are an expert AI assistant analyzing chat history summaries. Your task is to:
1. Identify the main patterns and themes in the conversation
2. Highlight recurring issues or misunderstandings
3. Note successful communication strategies
4. Provide actionable recommendations for improvement
5. Track the overall progress of the interaction

Format your response in a clear, structured way with these sections:
- Key Patterns and Themes
- Communication Strengths
- Areas for Improvement
- Actionable Recommendations
- Overall Progress`

	resp, err := a.client.CreateChatCompletion(
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
					Content: summary,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error getting analysis: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}
