package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// Analyzer represents a code analyzer
type Analyzer struct {
	client *openai.Client
	apiKey string
}

// NewAnalyzer creates a new code analyzer
func NewAnalyzer(apiKey string) *Analyzer {
	client := openai.NewClient(apiKey)
	return &Analyzer{
		client: client,
		apiKey: apiKey,
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
3. Note any patterns in communication that could be improved
4. Track progress on major tasks and decisions

Format your response as follows:

# ISSUE: [Brief description of the main issue/topic]

## Problem
- List specific problems or challenges identified
- Include relevant error messages or symptoms

## Debug Steps Taken
- List steps already attempted
- Note any successful or failed approaches

## Root Cause
- Identify the underlying cause of the issue
- Explain why the problem occurs

## Action Items
1. [Specific, actionable task]
2. [Specific, actionable task]
3. [Specific, actionable task]

## Technical Details
- Expected behavior: [description]
- Actual behavior: [description]
- Error codes: [if applicable]
- File paths: [if relevant]

## Next Steps
1. [Immediate next action]
2. [Follow-up action]
3. [Long-term consideration]`

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

func (a *Analyzer) analyzeChat(chatContent string) string {
	// Create a more focused prompt for the AI
	prompt := fmt.Sprintf(`Analyze the following chat interaction and provide a concise, actionable analysis focusing on better solutions than what was attempted. Format the response as follows:

# Chat Analysis

## Current Approach
[1-2 sentences describing what the user is trying to do]

## Better Solutions
1. [Primary solution] - [1-2 sentence explanation]
   - Key benefits: [bullet points]
   - Implementation steps: [brief steps]

2. [Alternative solution] - [1-2 sentence explanation]
   - Key benefits: [bullet points]
   - Implementation steps: [brief steps]

## Technical Considerations
- [Important technical detail 1]
- [Important technical detail 2]

## Best Practices
- [Relevant best practice 1]
- [Relevant best practice 2]

Chat content:
%s`, chatContent)

	// Get analysis from AI
	analysis, err := a.getAIResponse(prompt)
	if err != nil {
		return fmt.Sprintf("Error analyzing chat: %v", err)
	}

	return analysis
}

func (a *Analyzer) getAIResponse(prompt string) (string, error) {
	// Create the request body
	requestBody := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are a helpful AI assistant that analyzes chat interactions and suggests better solutions.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	}

	// Convert request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// Parse response
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("error parsing response: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from AI")
	}

	return response.Choices[0].Message.Content, nil
}

// GetErrorFix analyzes chat history for specific error types and provides fix suggestions
func (a *Analyzer) GetErrorFix(ctx context.Context, chatHistory string, errorType string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are an expert AI assistant analyzing chat history for error patterns and solutions. Your task is to:
1. Look for instances of the error type: "%s"
2. Extract the root cause, solution, and prevention steps
3. Provide the specific command to fix the error if available
4. Format the response in a clear, actionable way

Format your response as follows:

# Error Fix: %s

## Root Cause
[Explain why this error occurs]

## Solution
[Step-by-step solution]

## Prevention
[How to avoid this error in the future]

## Fix Command
[The specific command to fix this error, if available]

If no specific instances of this error are found in the chat history, provide general best practices for handling similar errors.`, errorType, errorType)

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
		return "", fmt.Errorf("error getting error fix: %w", err)
	}

	return resp.Choices[0].Message.Content, nil
}
