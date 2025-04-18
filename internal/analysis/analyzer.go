package analysis

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/sashabaranov/go-openai"
)

const (
	systemPrompt = `You are an expert software architect and Cursor AI assistant. Before suggesting any changes, carefully analyze the provided code and ask yourself:

1. Is the current implementation already optimal?
   - Does it follow best practices?
   - Is it performant and maintainable?
   - Are there any actual issues that need addressing?

2. Would refactoring provide meaningful benefits?
   - Would the benefits outweigh the risks of change?
   - Is the current solution actually the best approach?
   - Are there simpler alternatives that would work as well?

If the current implementation is already optimal, acknowledge this and explain why. If changes are needed, provide clear, step-by-step instructions for implementing improvements.`
)

type Analyzer struct {
	client *openai.Client
	cfg    *config.Config
}

func NewAnalyzer(cfg *config.Config) *Analyzer {
	client := openai.NewClient(cfg.OpenAIKey)
	return &Analyzer{
		client: client,
		cfg:    cfg,
	}
}

func (a *Analyzer) AnalyzeFile(ctx context.Context, filePath string) (string, error) {
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
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}

	// Format the response as meeting notes
	notes := fmt.Sprintf(`# Wash Meeting Notes - File Analysis
*Generated on %s*

## Key Insights
%s

## Action Items
- [ ] Review and implement suggested improvements
- [ ] Consider alternative approaches discussed
- [ ] Document any successful strategies for future reference

## Next Steps
- [ ] Follow up on identified issues
- [ ] Implement recommended changes
- [ ] Schedule next review if needed
`, time.Now().Format(time.RFC3339), resp.Choices[0].Message.Content)

	return notes, nil
}

func (a *Analyzer) AnalyzeProjectStructure(ctx context.Context, rootPath string) (string, error) {
	var structure string
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip common directories
			if info.Name() == "node_modules" || info.Name() == ".git" {
				return filepath.SkipDir
			}
			structure += fmt.Sprintf("📁 %s\n", path)
		} else {
			structure += fmt.Sprintf("  📄 %s\n", path)
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
					Role: openai.ChatMessageRoleSystem,
					Content: `You are an expert software architect. Analyze the provided project structure and provide insights about:
1. Overall project organization
2. Potential improvements in file/directory structure
3. Missing or redundant components
4. Best practices and recommendations`,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: structure,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("error creating chat completion: %w", err)
	}

	// Format the response as meeting notes
	notes := fmt.Sprintf(`# Wash Meeting Notes - Project Structure Analysis
*Generated on %s*

## Key Insights
%s

## Action Items
- [ ] Review and implement suggested improvements
- [ ] Consider alternative approaches discussed
- [ ] Document any successful strategies for future reference

## Next Steps
- [ ] Follow up on identified issues
- [ ] Implement recommended changes
- [ ] Schedule next review if needed
`, time.Now().Format(time.RFC3339), resp.Choices[0].Message.Content)

	return notes, nil
}
