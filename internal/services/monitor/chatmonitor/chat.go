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

	ticker := time.NewTicker(30 * time.Second)
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
		context.WriteString(fmt.Sprintf("Analysis: %s\n", r.Analysis.CurrentApproach))
		if len(r.Analysis.Issues) > 0 {
			context.WriteString(fmt.Sprintf("Issues: %s\n", strings.Join(r.Analysis.Issues, ", ")))
		}
		if len(r.Analysis.Solutions) > 0 {
			context.WriteString(fmt.Sprintf("Solutions: %s\n", strings.Join(r.Analysis.Solutions, ", ")))
		}
		context.WriteString("\n")
	}

	return context.String()
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
	prompt := fmt.Sprintf(`You are an expert software architect and intermediary between a human developer and their AI coding agent. 
Your role is to analyze the chat interactions in the provided window screenshots and do two things:
1. Identify potential issues and improvements, and record better solutions. Especially issues that have been caused by human error/bias misguiding the AI via poor prompts/communication.
2. Document best practices they use and the solutions to how they fix bugs.

Recent context:
%s

Based on this context and the current screenshot, please analyze the interaction and provide:
1. Current approach being taken
2. Any potential issues or improvements
3. Better solutions or approaches
4. Best practices observed

Format your response as a JSON object with the following structure:
{
    "current_approach": "string",
    "issues": ["string"],
    "solutions": ["string"],
    "best_practices": ["string"]
}`, contextStr)

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
	if err != nil {
		return fmt.Errorf("failed to analyze screenshot: %v", err)
	}

	// Parse the response into an analysis struct
	var analysis struct {
		CurrentApproach string   `json:"current_approach"`
		Issues          []string `json:"issues"`
		Solutions       []string `json:"solutions"`
		BestPractices   []string `json:"best_practices"`
	}

	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &analysis); err != nil {
		return fmt.Errorf("failed to parse analysis response: %v", err)
	}

	// Create monitor note
	interaction := &notes.MonitorNote{
		Timestamp:   time.Now(),
		ProjectName: m.projectName,
		ProjectGoal: m.cfg.ProjectGoal,
		Context: struct {
			CurrentState string   `json:"current_state"`
			FilesChanged []string `json:"files_changed,omitempty"`
		}{
			CurrentState: "Analyzing chat interaction",
		},
		Analysis: struct {
			CurrentApproach string   `json:"current_approach"`
			Issues          []string `json:"issues,omitempty"`
			Solutions       []string `json:"solutions,omitempty"`
			BestPractices   []string `json:"best_practices,omitempty"`
		}{
			CurrentApproach: analysis.CurrentApproach,
			Issues:          analysis.Issues,
			Solutions:       analysis.Solutions,
			BestPractices:   analysis.BestPractices,
		},
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
