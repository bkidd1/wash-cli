package chatmonitor

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/brinleekidd/wash-cli/internal/pid"
	"github.com/brinleekidd/wash-cli/internal/screenshot"
	"github.com/brinleekidd/wash-cli/pkg/config"
	"github.com/sashabaranov/go-openai"
)

type ChatMonitor struct {
	client     *openai.Client
	cfg        *config.Config
	isRunning  bool
	stopChan   chan struct{}
	doneChan   chan struct{}
	notesDir   string
	startTime  time.Time
	pidManager *pid.PIDManager
}

func NewChatMonitor(cfg *config.Config) (*ChatMonitor, error) {
	fmt.Println("Creating new chat monitor...")
	client := openai.NewClient(cfg.OpenAIKey)

	// Get the current working directory to create project-specific path
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %v", err)
	}

	// Create project-specific notes directory in ~/.wash/projects/
	projectPath := filepath.Base(cwd)
	notesDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", projectPath, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create notes directory: %v", err)
	}
	fmt.Printf("Notes directory: %s\n", notesDir)

	// Create .gitignore in notes directory to prevent accidental commits
	gitignorePath := filepath.Join(notesDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("*\n"), 0644); err != nil {
		return nil, fmt.Errorf("failed to create .gitignore: %v", err)
	}

	// Set up PID file path (keep this in .wash for system files)
	pidFile := filepath.Join(os.Getenv("HOME"), ".wash", "chat_monitor.pid")
	pidManager := pid.NewPIDManager(pidFile)

	return &ChatMonitor{
		client:     client,
		cfg:        cfg,
		isRunning:  false,
		stopChan:   make(chan struct{}),
		doneChan:   make(chan struct{}),
		notesDir:   notesDir,
		startTime:  time.Now(),
		pidManager: pidManager,
	}, nil
}

func (m *ChatMonitor) Start() error {
	if m.isRunning {
		return fmt.Errorf("chat monitor is already running")
	}

	// Check if another instance is running using PID manager
	if pid, err := m.pidManager.CheckRunning(); err == nil && pid > 0 {
		return fmt.Errorf("chat monitor is already running (PID: %d)", pid)
	}

	fmt.Println("Starting chat monitor...")
	// Create initial note file with header
	headerPath := filepath.Join(m.notesDir, "chat_analysis.txt")
	header := `# Chat Analysis

## Current Approach
[Initializing chat monitor and starting analysis]

## Better Solutions
1. [Waiting for first analysis]
   - Key benefits: [To be determined]
   - Implementation steps: [To be determined]

2. [Alternative solution pending]
   - Key benefits: [To be determined]
   - Implementation steps: [To be determined]

## Technical Considerations
- [Waiting for first analysis]
- [System will analyze chat interactions every 30 seconds]

## Best Practices
- [Waiting for first analysis]
- [Will provide recommendations based on observed patterns]
`

	if err := os.WriteFile(headerPath, []byte(header), 0644); err != nil {
		return fmt.Errorf("failed to create header file: %v", err)
	}
	fmt.Printf("Created analysis file: %s\n", headerPath)

	// Write PID file using PID manager
	if err := m.pidManager.WritePID(); err != nil {
		return fmt.Errorf("failed to write PID file: %v", err)
	}
	fmt.Printf("Wrote PID file (PID: %d)\n", os.Getpid())

	// Set up signal handling for cleanup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, cleaning up...")
		m.cleanup()
		os.Exit(0)
	}()

	m.isRunning = true
	go m.monitorLoop()
	fmt.Println("Chat monitor started successfully")

	// Wait for the monitor loop to complete
	<-m.doneChan
	return nil
}

func (m *ChatMonitor) cleanup() {
	if m.isRunning {
		m.isRunning = false
		close(m.stopChan)
		close(m.doneChan)
	}

	// Clean up PID file using PID manager
	if err := m.pidManager.Cleanup(); err != nil {
		fmt.Printf("Warning: failed to cleanup PID file: %v\n", err)
	}
}

func (m *ChatMonitor) Stop() error {
	// Check if process is running using PID manager
	pid, err := m.pidManager.CheckRunning()
	if err != nil || pid == 0 {
		return fmt.Errorf("chat monitor is not running")
	}

	fmt.Printf("Stopping chat monitor (PID: %d)...\n", pid)

	// Get the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %v", err)
	}

	// Send interrupt signal
	if err := process.Signal(os.Interrupt); err != nil {
		// If interrupt fails, try terminating
		if err := process.Signal(syscall.SIGTERM); err != nil {
			return fmt.Errorf("failed to stop process: %v", err)
		}
	}

	// Wait a moment for the process to clean up
	time.Sleep(100 * time.Millisecond)

	// Clean up PID file using PID manager
	if err := m.pidManager.Cleanup(); err != nil {
		fmt.Printf("Warning: failed to cleanup PID file: %v\n", err)
	}

	fmt.Println("Chat monitor stopped successfully")
	return nil
}

func (m *ChatMonitor) monitorLoop() {
	fmt.Println("Starting monitor loop...")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Taking screenshot and analyzing...")
			if err := m.analyzeScreenshot(); err != nil {
				fmt.Printf("Error analyzing screenshot: %v\n", err)
			}
		case <-m.stopChan:
			fmt.Println("Received stop signal, exiting monitor loop")
			m.isRunning = false
			close(m.doneChan)
			return
		}
	}
}

func (m *ChatMonitor) analyzeScreenshot() error {
	// Take a screenshot
	screenshotPath := filepath.Join(m.notesDir, "latest_screenshot.png")
	if err := screenshot.CaptureWindow("Cursor", screenshotPath); err != nil {
		return fmt.Errorf("failed to capture screenshot: %v", err)
	}

	// Read the screenshot file
	imageData, err := os.ReadFile(screenshotPath)
	if err != nil {
		return fmt.Errorf("failed to read screenshot: %v", err)
	}

	// Convert to base64
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Create the request with the correct model and message format
	req := openai.ChatCompletionRequest{
		Model: "gpt-4.1-mini",
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: `You are an expert AI assistant analyzing chat interactions. Your task is to analyze the chat screenshot and provide insights in the following format:

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

## Error Tracking
- [Error Type]: [Brief description of the error]
  - Root Cause: [Why the error occurs]
  - Solution: [How to fix it]
  - Prevention: [How to avoid it in the future]
  - Command: [The command to fix it, if applicable]

## Technical Considerations
- [Important technical detail 1]
- [Important technical detail 2]

## Best Practices
- [Relevant best practice 1]
- [Relevant best practice 2]

Focus on suggesting better approaches than what was attempted, with clear benefits and implementation steps.`,
			},
			{
				Role: openai.ChatMessageRoleUser,
				MultiContent: []openai.ChatMessagePart{
					{
						Type: "text",
						Text: "Please analyze this chat interaction and provide insights.",
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
		MaxTokens: 1000,
	}

	// Get the analysis from OpenAI
	resp, err := m.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to get analysis: %v", err)
	}

	// Append the analysis to the notes file with a timestamp
	analysisPath := filepath.Join(m.notesDir, "chat_analysis.txt")
	analysis := resp.Choices[0].Message.Content

	// Open file in append mode
	f, err := os.OpenFile(analysisPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open analysis file: %v", err)
	}
	defer f.Close()

	// Write the analysis with a timestamp
	timestamp := time.Now().Format("1/2/2006, 3:04:05 PM")
	if _, err := f.WriteString(fmt.Sprintf("\n\n### Analysis at %s\n\n%s\n", timestamp, analysis)); err != nil {
		return fmt.Errorf("failed to write analysis: %v", err)
	}

	return nil
}
