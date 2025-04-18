package chat

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/brinleekidd/wash-cli/internal/screenshot"
	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/sashabaranov/go-openai"
)

type ChatMonitor struct {
	client    *openai.Client
	cfg       *config.Config
	isRunning bool
	stopChan  chan struct{}
	notesDir  string
	startTime time.Time
}

func NewChatMonitor(cfg *config.Config) (*ChatMonitor, error) {
	client := openai.NewClient(cfg.OpenAIKey)

	// Create notes directory if it doesn't exist
	notesDir := filepath.Join(os.Getenv("HOME"), ".wash", "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create notes directory: %v", err)
	}

	return &ChatMonitor{
		client:    client,
		cfg:       cfg,
		isRunning: false,
		stopChan:  make(chan struct{}),
		notesDir:  notesDir,
		startTime: time.Now(),
	}, nil
}

func (m *ChatMonitor) Start() error {
	if m.isRunning {
		return fmt.Errorf("chat monitor is already running")
	}

	// Create initial note file with header
	headerPath := filepath.Join(m.notesDir, "chat_analysis.txt")
	header := fmt.Sprintf("# Continuous Chat Analysis\n*Started on %s*\n\n## Conversation Patterns and Insights\n\n",
		m.startTime.Format("1/2/2006, 3:04:05 PM"))

	if err := os.WriteFile(headerPath, []byte(header), 0644); err != nil {
		return fmt.Errorf("failed to create header file: %v", err)
	}

	m.isRunning = true
	go m.monitorLoop()
	return nil
}

func (m *ChatMonitor) Stop() error {
	if !m.isRunning {
		return fmt.Errorf("chat monitor is not running")
	}

	m.stopChan <- struct{}{}
	m.isRunning = false
	return nil
}

func (m *ChatMonitor) monitorLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.analyzeScreenshot(); err != nil {
				fmt.Printf("Error analyzing screenshot: %v\n", err)
			}
		case <-m.stopChan:
			return
		}
	}
}

func (m *ChatMonitor) analyzeScreenshot() error {
	// Take screenshot of primary display
	screenshot, err := screenshot.Capture(0)
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %v", err)
	}

	// Read the screenshot file
	imageBytes, err := os.ReadFile(screenshot.Path)
	if err != nil {
		return fmt.Errorf("failed to read screenshot: %v", err)
	}

	// Convert image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)

	// Analyze the screenshot using OpenAI Vision API
	resp, err := m.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4-vision-preview",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are an expert code reviewer and development assistant. Analyze the Cursor chat console screenshot and provide detailed insights about the development process, potential issues, and suggestions for improvement. Format your response in markdown with the following sections: KEY POINTS, ACTIONABLE SUGGESTIONS, COMMUNICATION PATTERNS, and PROGRESS TRACKING.",
				},
				{
					Role: "user",
					MultiContent: []openai.ChatMessagePart{
						{
							Type: "text",
							Text: "Please analyze this screenshot of a Cursor chat console. Look for any user input and potential mistakes or misguidance in the conversation. Focus on identifying where the user might have gone wrong or been misled. Format your response in markdown with the following sections: KEY POINTS, ACTIONABLE SUGGESTIONS, COMMUNICATION PATTERNS, and PROGRESS TRACKING.",
						},
						{
							Type: "image_url",
							ImageURL: &openai.ChatMessageImageURL{
								URL: fmt.Sprintf("data:image/png;base64,%s", base64Image),
							},
						},
					},
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to analyze screenshot: %v", err)
	}

	// Append the analysis to the main note file
	notePath := filepath.Join(m.notesDir, "chat_analysis.txt")
	analysis := fmt.Sprintf("\n### Analysis at %s\n\n%s\n---\n",
		time.Now().Format("1/2/2006, 3:04:05 PM"),
		resp.Choices[0].Message.Content)

	// Open the file in append mode
	f, err := os.OpenFile(notePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open note file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(analysis); err != nil {
		return fmt.Errorf("failed to write analysis: %v", err)
	}

	return nil
}
