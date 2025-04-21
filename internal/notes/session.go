package notes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BaseRecord is the common structure for all record types
type BaseRecord struct {
	SessionID  string    `json:"session_id"`
	Timestamp  time.Time `json:"timestamp"`
	RecordType string    `json:"record_type"`
}

// Session represents a focused period of work
type Session struct {
	ID          string        `json:"id"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time,omitempty"`
	ProjectName string        `json:"project_name"`
	ProjectGoal string        `json:"project_goal"`
	Records     []interface{} `json:"records"`
}

// SessionManager handles session creation and management
type SessionManager struct {
	mu             sync.RWMutex
	sessions       map[string]*Session
	currentSession *Session
	baseDir        string
}

// NewSessionManager creates a new session manager
func NewSessionManager(baseDir string) (*SessionManager, error) {
	// Create sessions directory if it doesn't exist
	sessionsDir := filepath.Join(baseDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating sessions directory: %w", err)
	}

	return &SessionManager{
		sessions: make(map[string]*Session),
		baseDir:  sessionsDir,
	}, nil
}

// StartSession creates a new session
func (sm *SessionManager) StartSession(projectName, goal string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// End current session if exists
	if sm.currentSession != nil {
		sm.currentSession.EndTime = time.Now()
		if err := sm.saveSession(sm.currentSession); err != nil {
			return nil, fmt.Errorf("error saving current session: %w", err)
		}
	}

	// Create new session
	session := &Session{
		ID:          uuid.New().String(),
		StartTime:   time.Now(),
		ProjectName: projectName,
		ProjectGoal: goal,
		Records:     make([]interface{}, 0),
	}

	sm.sessions[session.ID] = session
	sm.currentSession = session

	// Save session to disk
	if err := sm.saveSession(session); err != nil {
		return nil, fmt.Errorf("error saving new session: %w", err)
	}

	return session, nil
}

// EndSession ends the current session
func (sm *SessionManager) EndSession() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentSession == nil {
		return fmt.Errorf("no active session")
	}

	sm.currentSession.EndTime = time.Now()
	if err := sm.saveSession(sm.currentSession); err != nil {
		return fmt.Errorf("error saving session: %w", err)
	}

	sm.currentSession = nil
	return nil
}

// AddRecord adds a record to the current session
func (sm *SessionManager) AddRecord(record interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.currentSession == nil {
		return fmt.Errorf("no active session")
	}

	// Add base record fields if the record has them
	if baseRec, ok := record.(interface{ SetBaseRecord(BaseRecord) }); ok {
		baseRec.SetBaseRecord(BaseRecord{
			SessionID:  sm.currentSession.ID,
			Timestamp:  time.Now(),
			RecordType: getRecordType(record),
		})
	}

	sm.currentSession.Records = append(sm.currentSession.Records, record)
	return sm.saveSession(sm.currentSession)
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(id string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sessions[id]
}

// GetCurrentSession returns the current active session
func (sm *SessionManager) GetCurrentSession() *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentSession
}

// GetRecentRecords returns records from the last duration
func (sm *SessionManager) GetRecentRecords(sessionID string, duration time.Duration) []interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session := sm.sessions[sessionID]
	if session == nil {
		return nil
	}

	cutoff := time.Now().Add(-duration)
	var recent []interface{}

	for _, record := range session.Records {
		if baseRec, ok := record.(interface{ GetTimestamp() time.Time }); ok {
			if baseRec.GetTimestamp().After(cutoff) {
				recent = append(recent, record)
			}
		}
	}

	return recent
}

// saveSession saves a session to disk
func (sm *SessionManager) saveSession(session *Session) error {
	sessionPath := filepath.Join(sm.baseDir, fmt.Sprintf("%s.json", session.ID))
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling session: %w", err)
	}

	return os.WriteFile(sessionPath, data, 0644)
}

// Helper function to get record type
func getRecordType(record interface{}) string {
	switch record.(type) {
	case *Interaction:
		return "interaction"
	case *Note:
		return "note"
	case *CodeChange:
		return "code_change"
	default:
		return "unknown"
	}
}
