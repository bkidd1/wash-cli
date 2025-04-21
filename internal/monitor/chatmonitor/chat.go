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

	"github.com/bkidd1/wash-cli/internal/pid"
	"github.com/bkidd1/wash-cli/internal/screenshot"
	"github.com/bkidd1/wash-cli/pkg/config"
	"github.com/sashabaranov/go-openai"
)

type Monitor struct {
	client     *openai.Client
	cfg        *config.Config
	running    bool
	stopChan   chan struct{}
	doneChan   chan struct{}
	notesDir   string
	startTime  time.Time
	pidManager *pid.PIDManager
	pidFile    string
}

func NewMonitor(cfg *config.Config) (*Monitor, error) {
	fmt.Println("Creating new monitor...")
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

	// Create .gitignore in notes directory to prevent accidental commits
	gitignorePath := filepath.Join(notesDir, ".gitignore")
	if err := os.WriteFile(gitignorePath, []byte("*\n"), 0644); err != nil {
		return nil, fmt.Errorf("failed to create .gitignore: %v", err)
	}

	// Set up PID file path (keep this in .wash for system files)
	pidFile := filepath.Join(os.Getenv("HOME"), ".wash", "chat_monitor.pid")
	pidManager := pid.NewPIDManager(pidFile)

	return &Monitor{
		client:     client,
		cfg:        cfg,
		running:    false,
		stopChan:   make(chan struct{}),
		doneChan:   make(chan struct{}),
		notesDir:   notesDir,
		startTime:  time.Now(),
		pidManager: pidManager,
		pidFile:    pidFile,
	}, nil
}

func (m *Monitor) Start() error {
	if m.running {
		return fmt.Errorf("monitor is already running")
	}

	// Check if another instance is running using PID manager
	if pid, err := m.pidManager.CheckRunning(); err == nil && pid > 0 {
		return fmt.Errorf("monitor is already running (PID: %d)", pid)
	}

	// Create initial note file with header
	headerPath := filepath.Join(m.notesDir, "chat_analysis.txt")
	header := `You are an expert software architect and intermediary between a human developer and their AI coding agent. 
Your role is to analyze the chat interactions in the provided window screenshots and do two things:
1. Identify potential issues and improvements, and record better solutions. Especially issues that have been caused by human error/bias misguiding the AI via poor prompts/communication.
2. Document best practices they use and the solutions to how they fix bugs.

Your observations will be recorded in the wash notes directory. 
Your observations will serve as context for AI coding agents to solve future issues/make better decisions in the current project.
Here is the format you should use to record your observations:

## Current Approach
[Initializing chat monitor and starting analysis]

## Issues
- [Waiting for first analysis]
- [System will analyze chat interactions every 30 seconds]

## Solutions
- [Waiting for first analysis]
- [Will provide better approaches and implementation steps]

## Best Practices
- [Waiting for first analysis]
- [Will document effective patterns and successful fixes]`

	if err := os.WriteFile(headerPath, []byte(header), 0644); err != nil {
		return fmt.Errorf("failed to create header file: %v", err)
	}

	// Write PID file using PID manager
	if err := m.pidManager.WritePID(); err != nil {
		return fmt.Errorf("failed to write PID file: %v", err)
	}

	// Set up signal handling for cleanup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		m.cleanup()
		os.Exit(0)
	}()

	m.running = true
	go m.monitorLoop()
	fmt.Println("Monitor started successfully!")

	// Wait for the monitor loop to complete
	<-m.doneChan
	return nil
}

func (m *Monitor) cleanup() {
	if m.running {
		m.running = false
		close(m.stopChan)
		close(m.doneChan)
	}

	// Clean up PID file using PID manager
	if err := m.pidManager.Cleanup(); err != nil {
		fmt.Printf("Warning: failed to cleanup PID file: %v\n", err)
	}
}

func (m *Monitor) Stop() error {
	if !m.running {
		return fmt.Errorf("monitor is not running")
	}

	// Check if process is running using PID manager
	pid, err := m.pidManager.CheckRunning()
	if err != nil || pid == 0 {
		return fmt.Errorf("monitor is not running")
	}

	fmt.Printf("Stopping monitor (PID: %d)...\n", pid)

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

	fmt.Println("Monitor stopped successfully")
	return nil
}

func (m *Monitor) monitorLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Monitoring...")
			if err := m.analyzeScreenshot(); err != nil {
				fmt.Printf("Error analyzing screenshot: %v\n", err)
			}
		case <-m.stopChan:
			m.running = false
			close(m.doneChan)
			return
		}
	}
}

func (m *Monitor) analyzeScreenshot() error {
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
				Content: `You are an expert software architect and intermediary between a human developer and their AI coding agent. 
Your role is to analyze the chat interactions in the provided window screenshots and do two things:
1. Identify potential issues and improvements, and record better solutions. Especially issues that have been caused by human error/bias misguiding the AI via poor prompts/communication.
2. Document best practices they use and the solutions to how they fix bugs.

Your observations will be recorded in the wash notes directory. 
Your observations will serve as context for AI coding agents to solve future issues/make better decisions in the current project.
Here is the format you should use to record your observations:

## Current Approach
[1-2 sentences describing what the user is trying to achieve and their current method]

## Issues
- [Potential issue or problem identified]
- [Human error/bias in communication]
- [Misunderstanding or miscommunication]

## Solutions
- [Better approach or solution]
- [Implementation step]
- [Improvement to communication or prompt]

## Best Practices
- [Effective pattern or practice observed]
- [Successful bug fix or problem-solving approach]
- [Particularly effective communication strategy]

For each section, provide multiple bullet points where relevant. Each bullet point should be a complete, standalone statement that can be parsed into the JSON structure.`,
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
	}

	// Get the analysis
	resp, err := m.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return fmt.Errorf("failed to get analysis: %v", err)
	}

	// Delete the screenshot after we have the analysis
	if err := os.Remove(screenshotPath); err != nil {
		return fmt.Errorf("failed to delete screenshot: %v", err)
	}

	// Append the analysis to the notes file with a timestamp
	analysisPath := filepath.Join(m.notesDir, "chat_analysis.txt")
	analysis := resp.Choices[0].Message.Content

	// Open file in append mode
	f, err := os.OpenFile(analysisPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
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
