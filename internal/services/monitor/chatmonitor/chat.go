package chatmonitor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bkidd1/wash-cli/internal/pid"
	"github.com/bkidd1/wash-cli/internal/services/notes"
	"github.com/bkidd1/wash-cli/internal/services/screenshot"
	"github.com/bkidd1/wash-cli/internal/utils/config"
	"github.com/sashabaranov/go-openai"
)

type Monitor struct {
	client      *openai.Client
	cfg         *config.Config
	running     bool
	stopChan    chan struct{}
	doneChan    chan struct{}
	notesDir    string
	startTime   time.Time
	pidManager  *pid.PIDManager
	pidFile     string
	projectName string
}

func NewMonitor(cfg *config.Config, projectName string) (*Monitor, error) {
	fmt.Println("Creating new monitor...")
	client := openai.NewClient(cfg.OpenAIKey)

	// If project name not provided, use current directory name
	if projectName == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %v", err)
		}
		projectName = filepath.Base(cwd)
	}

	// Create project-specific notes directory in ~/.wash/projects/
	notesDir := filepath.Join(os.Getenv("HOME"), ".wash", "projects", projectName, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create notes directory: %v", err)
	}

	// Create PID manager
	pidFile := filepath.Join(os.Getenv("HOME"), ".wash", "chat_monitor.pid")
	pidManager := pid.NewPIDManager(pidFile)

	return &Monitor{
		client:      client,
		cfg:         cfg,
		running:     false,
		stopChan:    make(chan struct{}),
		doneChan:    make(chan struct{}),
		notesDir:    notesDir,
		startTime:   time.Now(),
		pidManager:  pidManager,
		pidFile:     pidFile,
		projectName: projectName,
	}, nil
}

func (m *Monitor) Start() error {
	if m.running {
		return fmt.Errorf("monitor is already running")
	}

	// Write PID file
	if err := m.pidManager.WritePID(); err != nil {
		return fmt.Errorf("failed to write PID file: %v", err)
	}

	m.running = true
	go m.monitorLoop()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		m.Stop()
	}()

	return nil
}

func (m *Monitor) cleanup() {
	if err := m.pidManager.Cleanup(); err != nil {
		fmt.Printf("Warning: Failed to cleanup PID file: %v\n", err)
	}
}

func (m *Monitor) Stop() error {
	if !m.running {
		return fmt.Errorf("monitor is not running")
	}

	close(m.stopChan)
	<-m.doneChan
	m.running = false

	m.cleanup()
	return nil
}

func (m *Monitor) monitorLoop() {
	defer close(m.doneChan)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-ticker.C:
			if err := m.analyzeScreenshot(); err != nil {
				fmt.Printf("Warning: Failed to analyze screenshot: %v\n", err)
			}
		}
	}
}

func (m *Monitor) analyzeScreenshot() error {
	// Take screenshot
	screenshotData, err := screenshot.Capture(0) // Capture primary display
	if err != nil {
		return fmt.Errorf("failed to capture screenshot: %v", err)
	}

	// Read screenshot file
	data, err := os.ReadFile(screenshotData.Path)
	if err != nil {
		return fmt.Errorf("failed to read screenshot file: %v", err)
	}

	// Convert screenshot to base64
	screenshotBase64 := base64.StdEncoding.EncodeToString(data)

	// Analyze screenshot with OpenAI
	resp, err := m.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4VisionPreview,
			Messages: []openai.ChatCompletionMessage{
				{
					Role: openai.ChatMessageRoleSystem,
					Content: "You are a development assistant analyzing a screenshot of a developer's screen. " +
						"Focus on identifying the current development activity, potential issues, and best practices.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("Analyze this screenshot of a developer's screen and provide insights about the current development activity.\n\n![Screenshot](data:image/png;base64,%s)", screenshotBase64),
				},
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to analyze screenshot: %v", err)
	}

	// Create monitor note
	interaction := &notes.MonitorNote{
		Timestamp:   time.Now(),
		ProjectName: m.projectName,
		ProjectGoal: m.cfg.ProjectGoal,
	}

	// Parse response and update note
	response := resp.Choices[0].Message.Content
	lines := strings.Split(response, "\n")
	for i, line := range lines {
		switch {
		case strings.HasPrefix(line, "Current State:"):
			interaction.Context.CurrentState = strings.TrimSpace(strings.TrimPrefix(line, "Current State:"))
		case strings.HasPrefix(line, "Current Approach:"):
			interaction.Analysis.CurrentApproach = strings.TrimSpace(strings.TrimPrefix(line, "Current Approach:"))
		case strings.HasPrefix(line, "Issues:"):
			interaction.Analysis.Issues = append(interaction.Analysis.Issues, strings.TrimSpace(strings.TrimPrefix(line, "Issues:")))
		case strings.HasPrefix(line, "Solutions:"):
			interaction.Analysis.Solutions = append(interaction.Analysis.Solutions, strings.TrimSpace(strings.TrimPrefix(line, "Solutions:")))
		case strings.HasPrefix(line, "Best Practices:"):
			interaction.Analysis.BestPractices = append(interaction.Analysis.BestPractices, strings.TrimSpace(strings.TrimPrefix(line, "Best Practices:")))
		case i > 0 && strings.HasPrefix(lines[i-1], "Issues:"):
			interaction.Analysis.Issues = append(interaction.Analysis.Issues, strings.TrimSpace(line))
		case i > 0 && strings.HasPrefix(lines[i-1], "Solutions:"):
			interaction.Analysis.Solutions = append(interaction.Analysis.Solutions, strings.TrimSpace(line))
		case i > 0 && strings.HasPrefix(lines[i-1], "Best Practices:"):
			interaction.Analysis.BestPractices = append(interaction.Analysis.BestPractices, strings.TrimSpace(line))
		}
	}

	// Save note to file
	noteFile := filepath.Join(m.notesDir, fmt.Sprintf("monitor_%s.json", time.Now().Format("2006-01-02-15-04-05")))
	noteData, err := json.MarshalIndent(interaction, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal monitor note: %v", err)
	}

	if err := os.WriteFile(noteFile, noteData, 0644); err != nil {
		return fmt.Errorf("failed to save monitor note: %v", err)
	}

	return nil
}
