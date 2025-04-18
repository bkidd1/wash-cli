package chat

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
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
	doneChan  chan struct{}
	notesDir  string
	startTime time.Time
	pidFile   string
}

func NewChatMonitor(cfg *config.Config) (*ChatMonitor, error) {
	fmt.Println("Creating new chat monitor...")
	client := openai.NewClient(cfg.OpenAIKey)

	// Create notes directory in the project directory
	notesDir := "wash-notes"
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create notes directory: %v", err)
	}
	fmt.Printf("Notes directory: %s\n", notesDir)

	// Set up PID file path (keep this in .wash for system files)
	pidFile := filepath.Join(os.Getenv("HOME"), ".wash", "chat_monitor.pid")
	// Ensure .wash directory exists for PID file
	if err := os.MkdirAll(filepath.Dir(pidFile), 0755); err != nil {
		return nil, fmt.Errorf("failed to create .wash directory: %v", err)
	}

	return &ChatMonitor{
		client:    client,
		cfg:       cfg,
		isRunning: false,
		stopChan:  make(chan struct{}),
		doneChan:  make(chan struct{}),
		notesDir:  notesDir,
		startTime: time.Now(),
		pidFile:   pidFile,
	}, nil
}

func (m *ChatMonitor) Start() error {
	if m.isRunning {
		return fmt.Errorf("chat monitor is already running")
	}

	// Check if another instance is running by checking PID file
	if pid, err := m.checkRunningInstance(); err == nil && pid > 0 {
		return fmt.Errorf("chat monitor is already running (PID: %d)", pid)
	}

	fmt.Println("Starting chat monitor...")
	// Create initial note file with header
	headerPath := filepath.Join(m.notesDir, "chat_analysis.txt")
	header := fmt.Sprintf("# Continuous Chat Analysis\n*Started on %s*\n\n## Conversation Patterns and Insights\n\n",
		m.startTime.Format("1/2/2006, 3:04:05 PM"))

	if err := os.WriteFile(headerPath, []byte(header), 0644); err != nil {
		return fmt.Errorf("failed to create header file: %v", err)
	}
	fmt.Printf("Created analysis file: %s\n", headerPath)

	// Write PID file with just the PID number
	pid := os.Getpid()
	if err := os.WriteFile(m.pidFile, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
		return fmt.Errorf("failed to write PID file: %v", err)
	}
	fmt.Printf("Wrote PID file: %s (PID: %d)\n", m.pidFile, pid)

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
	return nil
}

func (m *ChatMonitor) checkRunningInstance() (int, error) {
	// Check if PID file exists
	if _, err := os.Stat(m.pidFile); os.IsNotExist(err) {
		fmt.Println("Debug: PID file does not exist")
		return 0, nil
	}

	// Read PID from file
	pidBytes, err := os.ReadFile(m.pidFile)
	if err != nil {
		fmt.Printf("Debug: Failed to read PID file: %v\n", err)
		// Can't read PID file, assume no running instance
		os.Remove(m.pidFile)
		return 0, nil
	}

	// Clean up the PID string and convert to integer
	pidStr := string(pidBytes)
	pidStr = strings.TrimSpace(pidStr)
	pidStr = strings.TrimSuffix(pidStr, "%")
	fmt.Printf("Debug: Read PID string: '%s'\n", pidStr)

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		fmt.Printf("Debug: Invalid PID format: %v\n", err)
		// Invalid PID in file, clean up
		os.Remove(m.pidFile)
		return 0, nil
	}
	fmt.Printf("Debug: Converted PID: %d\n", pid)

	// Check if process exists and is running
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Debug: Process not found: %v\n", err)
		// Process not found, clean up
		os.Remove(m.pidFile)
		return 0, nil
	}

	// On Unix systems, FindProcess always succeeds, so we need to check if the process is actually running
	if err := process.Signal(syscall.Signal(0)); err != nil {
		fmt.Printf("Debug: Process not running: %v\n", err)
		// Process not running, clean up
		os.Remove(m.pidFile)
		return 0, nil
	}

	fmt.Printf("Debug: Found running process with PID %d\n", pid)
	return pid, nil
}

func (m *ChatMonitor) cleanup() {
	if m.isRunning {
		m.isRunning = false
		close(m.stopChan)
		close(m.doneChan)
	}

	// Remove PID file if it belongs to us
	if pid, err := m.checkRunningInstance(); err == nil && pid == os.Getpid() {
		os.Remove(m.pidFile)
	}
}

func (m *ChatMonitor) Stop() error {
	// Check if process is running
	pid, err := m.checkRunningInstance()
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

	// Clean up PID file if it still exists
	if _, err := os.Stat(m.pidFile); err == nil {
		os.Remove(m.pidFile)
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
	fmt.Println("Capturing screenshot...")
	// Take screenshot of primary display
	screenshot, err := screenshot.Capture(0)
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %v", err)
	}
	fmt.Printf("Screenshot captured: %s\n", screenshot.Path)

	// Read the screenshot file
	imageBytes, err := os.ReadFile(screenshot.Path)
	if err != nil {
		return fmt.Errorf("failed to read screenshot: %v", err)
	}
	fmt.Printf("Read screenshot file, size: %d bytes\n", len(imageBytes))

	// Convert image to base64
	base64Image := base64.StdEncoding.EncodeToString(imageBytes)
	fmt.Println("Converted image to base64")

	// Analyze the screenshot using OpenAI Vision API
	fmt.Println("Sending request to OpenAI Vision API...")
	req := openai.ChatCompletionRequest{
		Model: "gpt-4.1-mini",
		Messages: []openai.ChatCompletionMessage{
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
							URL:    fmt.Sprintf("data:image/png;base64,%s", base64Image),
							Detail: "high",
						},
					},
				},
			},
		},
		MaxTokens: 300,
	}

	fmt.Printf("Request details:\nModel: %s\nMaxTokens: %d\n", req.Model, req.MaxTokens)
	fmt.Printf("Message count: %d\n", len(req.Messages))
	fmt.Printf("Content parts: %d\n", len(req.Messages[0].MultiContent))

	resp, err := m.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		fmt.Printf("OpenAI API error details:\n%+v\n", err)
		return fmt.Errorf("failed to analyze screenshot: %v", err)
	}
	fmt.Println("Received response from OpenAI Vision API")

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
	fmt.Printf("Analysis written to: %s\n", notePath)

	return nil
}
