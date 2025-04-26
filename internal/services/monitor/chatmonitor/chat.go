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
	client       *openai.Client
	cfg          *config.Config
	running      bool
	stopChan     chan struct{}
	doneChan     chan struct{}
	notesDir     string
	startTime    time.Time
	pidManager   *pid.PIDManager
	pidFile      string
	projectName  string
	notesManager *notes.NotesManager
}

func NewMonitor(cfg *config.Config, projectName string) (*Monitor, error) {
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

	// Create notes manager
	notesManager, err := notes.NewNotesManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create notes manager: %v", err)
	}

	return &Monitor{
		client:       client,
		cfg:          cfg,
		running:      false,
		stopChan:     make(chan struct{}),
		doneChan:     make(chan struct{}),
		notesDir:     notesDir,
		startTime:    time.Now(),
		pidManager:   pidManager,
		pidFile:      pidFile,
		projectName:  projectName,
		notesManager: notesManager,
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
		// Silently handle cleanup errors
		_ = err
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

	// Ticker for screenshot analysis (every 30 seconds)
	screenshotTicker := time.NewTicker(30 * time.Second)
	defer screenshotTicker.Stop()

	// Ticker for progress notes (every 5 minutes)
	progressTicker := time.NewTicker(5 * time.Minute)
	defer progressTicker.Stop()

	for {
		select {
		case <-m.stopChan:
			return
		case <-screenshotTicker.C:
			// Log screenshot analysis errors
			if err := m.analyzeScreenshot(); err != nil {
				fmt.Printf("Error analyzing screenshot: %v\n", err)
			}
		case <-progressTicker.C:
			// Generate progress note for the last 5 minutes
			progressNote, err := m.notesManager.GenerateProgressFromMonitor(m.projectName, 5*time.Minute)
			if err != nil {
				fmt.Printf("Error generating progress note: %v\n", err)
				continue
			}

			// Save the progress note
			if err := m.notesManager.SaveProjectProgress(progressNote); err != nil {
				fmt.Printf("Error saving progress note: %v\n", err)
			}
		}
	}
}

// formatContextForAI formats recent records into a context string for the AI
func formatContextForAI(records []*notes.Interaction) string {
	if len(records) == 0 {
		return "No recent context available."
	}

	var context strings.Builder
	context.WriteString("Recent context:\n\n")

	for _, r := range records {
		context.WriteString(fmt.Sprintf("Interaction at %s:\n", r.Timestamp.Format("2006-01-02 15:04:05")))
		context.WriteString(fmt.Sprintf("Context: %s\n", r.Context.CurrentState))
		if len(r.Context.FilesChanged) > 0 {
			context.WriteString(fmt.Sprintf("Files Changed: %s\n", strings.Join(r.Context.FilesChanged, ", ")))
		}
		context.WriteString(fmt.Sprintf("Current Approach: %s\n", r.Analysis.CurrentApproach))
		if len(r.Analysis.AlternativeApproaches) > 0 {
			context.WriteString(fmt.Sprintf("Alternative Approaches: %s\n", strings.Join(r.Analysis.AlternativeApproaches, ", ")))
		}
		context.WriteString("\n")
	}

	return context.String()
}

func (m *Monitor) analyzeScreenshot() error {
	// Create screenshots directory if it doesn't exist
	dir := filepath.Join(os.Getenv("HOME"), ".wash-screenshots")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create screenshots directory: %v", err)
	}

	// Generate filename with timestamp
	filename := fmt.Sprintf("screenshot-%s.png", time.Now().Format("2006-01-02-15-04-05"))
	screenshotPath := filepath.Join(dir, filename)

	// Take screenshot of Cursor window
	if err := screenshot.CaptureWindow("Cursor", screenshotPath); err != nil {
		return fmt.Errorf("failed to capture Cursor window: %v", err)
	}

	// Read screenshot file
	data, err := os.ReadFile(screenshotPath)
	if err != nil {
		return fmt.Errorf("failed to read screenshot file: %v", err)
	}

	// Convert screenshot to base64
	screenshotBase64 := base64.StdEncoding.EncodeToString(data)

	// Get recent interactions for context
	recentInteractions, err := m.notesManager.LoadInteractions(m.projectName)
	if err != nil {
		return fmt.Errorf("failed to load recent interactions: %v", err)
	}

	// Filter to last 5 minutes
	cutoff := time.Now().Add(-5 * time.Minute)
	var recentRecords []*notes.Interaction
	for _, interaction := range recentInteractions {
		if interaction.Timestamp.After(cutoff) {
			recentRecords = append(recentRecords, interaction)
		}
	}

	contextStr := formatContextForAI(recentRecords)

	// Create the analysis prompt with context
	prompt := `You are observing a conversation between a user and an AI coding assistant in the Cursor IDE.
Your task is to analyze the screenshot and provide a concise summary of the interaction.

Based on the screenshot, please analyze:
1. The user's request or question. Consider what they're trying to accomplish (this will most likely be in the bottom right corner of the screenshot where the user input for the chat is)
2. The AI assistant's response and actions (the response willusually be above the user input on the right side of the screenshot)
3. Code changes or modifications that seem to occur
4. The overall context of the interaction (e.g., debugging, feature implementation)

IMPORTANT: Keep all descriptions brief and to the point. Each field should be 1 sentence maximum.
Focus on the key points and avoid unnecessary details.

Format your response as a JSON object with the following structure:
{
    "user_request": "brief description of the user goal expressed in the chat in the lower right corner of the screenshot",
    "ai_action": "brief description of the AI's main action - or the user's action if they edit the code directly.",
    "context": "brief context (e.g., debugging, feature implementation)",
    "code_changes": ["which file(s) were edited, if any"]
}` + "\n\n" + contextStr

	// Add retry logic for transient network errors
	maxRetries := 3
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		// Create the chat completion request
		resp, err := m.client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: "gpt-4.1-mini",
				Messages: []openai.ChatCompletionMessage{
					{
						Role: "user",
						MultiContent: []openai.ChatMessagePart{
							{
								Type: "text",
								Text: prompt,
							},
							{
								Type: "image_url",
								ImageURL: &openai.ChatMessageImageURL{
									URL: fmt.Sprintf("data:image/png;base64,%s", screenshotBase64),
								},
							},
						},
					},
				},
				MaxTokens: 1000,
			},
		)
		if err == nil {
			// Parse the response into an analysis struct
			var analysis struct {
				UserRequest string   `json:"user_request"`
				AIAction    string   `json:"ai_action"`
				Context     string   `json:"context"`
				CodeChanges []string `json:"code_changes"`
			}

			if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
				return fmt.Errorf("failed to parse analysis response: %v", err)
			}

			// Create a new monitor note
			note := &notes.MonitorNote{
				Timestamp:   time.Now(),
				ProjectName: m.projectName,
				Interaction: struct {
					UserRequest string   `json:"user_request"`
					AIAction    string   `json:"ai_action"`
					Context     string   `json:"context"`
					CodeChanges []string `json:"code_changes"`
				}{
					UserRequest: analysis.UserRequest,
					AIAction:    analysis.AIAction,
					Context:     analysis.Context,
					CodeChanges: analysis.CodeChanges,
				},
			}

			// Save note using the notes manager
			if err := m.notesManager.SaveMonitorNote(m.projectName, note); err != nil {
				return fmt.Errorf("failed to save monitor note: %v", err)
			}

			return nil
		}

		// Check if this is a retryable error
		if strings.Contains(err.Error(), "tls: bad record MAC") ||
			strings.Contains(err.Error(), "connection reset by peer") ||
			strings.Contains(err.Error(), "i/o timeout") {
			lastErr = err
			// Wait before retrying (exponential backoff)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		// If it's not a retryable error, return immediately
		return fmt.Errorf("failed to analyze screenshot: %v", err)
	}

	// If we've exhausted all retries, return the last error
	return fmt.Errorf("failed to analyze screenshot after %d retries: %v", maxRetries, lastErr)
}

// StartTime returns the time when the monitor was started
func (m *Monitor) StartTime() time.Time {
	return m.startTime
}
