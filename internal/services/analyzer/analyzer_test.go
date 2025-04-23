package analyzer

import (
	"testing"
)

func TestNewTerminalAnalyzer(t *testing.T) {
	apiKey := "test-key"
	projectGoal := "test project"
	rememberNotes := []string{"note1", "note2"}

	analyzer := NewTerminalAnalyzer(apiKey, projectGoal, rememberNotes)

	if analyzer == nil {
		t.Error("Expected analyzer to be created, got nil")
	}

	if analyzer.projectGoal != projectGoal {
		t.Errorf("Expected projectGoal to be %s, got %s", projectGoal, analyzer.projectGoal)
	}

	if len(analyzer.rememberNotes) != len(rememberNotes) {
		t.Errorf("Expected %d remember notes, got %d", len(rememberNotes), len(analyzer.rememberNotes))
	}
}

func TestUpdateProjectContext(t *testing.T) {
	analyzer := NewTerminalAnalyzer("test-key", "initial goal", []string{"note1"})

	newGoal := "updated goal"

	analyzer.UpdateProjectContext(newGoal)

	if analyzer.projectGoal != newGoal {
		t.Errorf("Expected projectGoal to be %s, got %s", newGoal, analyzer.projectGoal)
	}

	// Verify remember notes remain unchanged
	if len(analyzer.rememberNotes) != 1 || analyzer.rememberNotes[0] != "note1" {
		t.Errorf("Expected remember notes to remain unchanged, got %v", analyzer.rememberNotes)
	}
}
