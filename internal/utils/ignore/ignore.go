package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// DefaultIgnorePatterns contains common patterns to ignore
var DefaultIgnorePatterns = []string{
	// Version control
	".git",
	".svn",
	".hg",

	// Dependencies
	"node_modules",
	"vendor",
	"bower_components",
	"jspm_packages",

	// Build outputs
	"dist",
	"build",
	"out",
	"target",
	"bin",

	// IDE and editor files
	".idea",
	".vscode",
	"*.swp",
	"*.swo",

	// OS files
	".DS_Store",
	"Thumbs.db",

	// Logs and databases
	"*.log",
	"*.sqlite",
	"*.db",

	// Cache directories
	".cache",
	"tmp",
	"temp",
}

// ShouldIgnore checks if a path should be ignored based on patterns
func ShouldIgnore(path string, patterns []string) bool {
	// Convert path to forward slashes for consistent matching
	path = filepath.ToSlash(path)

	// Check each pattern
	for _, pattern := range patterns {
		// Handle exact matches
		if path == pattern {
			return true
		}

		// Handle directory matches
		if strings.HasSuffix(pattern, "/") {
			if strings.HasPrefix(path, pattern) {
				return true
			}
		}

		// Handle wildcard matches
		if strings.Contains(pattern, "*") {
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err == nil && matched {
				return true
			}
		}
	}

	return false
}

// LoadGitignorePatterns loads patterns from .gitignore file
func LoadGitignorePatterns(rootPath string) ([]string, error) {
	gitignorePath := filepath.Join(rootPath, ".gitignore")
	patterns := make([]string, 0)

	// Add default patterns
	patterns = append(patterns, DefaultIgnorePatterns...)

	// Try to read .gitignore file
	file, err := os.Open(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return patterns, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		patterns = append(patterns, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return patterns, nil
}
