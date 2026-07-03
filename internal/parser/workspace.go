package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yogasimman/anjal/internal/models"
)

// LoadWorkspace automatically finds the correct Anjal directory, scans it for
// markdown files, and parses all of them into Collections.
func LoadWorkspace() ([]models.Collection, error) {
	workspacePath, err := findOrCreateWorkspace()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace: %w", err)
	}

	return LoadWorkspaceFrom(workspacePath)
}

// LoadWorkspaceFrom scans a specific directory for .md files and parses them.
// It also wires up collection-level auth: if a request has no Auth of its own
// but the collection has a global Auth, the request inherits it.
func LoadWorkspaceFrom(dir string) ([]models.Collection, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read workspace directory: %w", err)
	}

	var collections []models.Collection

	for _, entry := range entries {
		// Only process markdown files
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".md") {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())

		// ParseFile now returns the global Auth as well
		requests, globalAuth, err := ParseFile(fullPath)
		if err != nil {
			fmt.Printf("Warning: Failed to parse %s: %v\n", entry.Name(), err)
			continue
		}

		// Only add the collection if it actually contains HTTP blocks
		if len(requests) > 0 {
			col := models.Collection{
				Name:     entry.Name(),
				FilePath: fullPath,
				Auth:     globalAuth,
				Requests: requests,
			}

			// Cascading inheritance: if a request has no Auth, inherit from collection
			if globalAuth != nil {
				for i := range col.Requests {
					if col.Requests[i].Auth == nil {
						col.Requests[i].Auth = globalAuth
					}
				}
			}

			collections = append(collections, col)
		}
	}

	return collections, nil
}

// findOrCreateWorkspace determines where the markdown files live based on priority:
// 1. .anjal/ in the current directory
// 2. anjal/ in the current directory
// 3. ~/.anjal/ in the user's home directory (auto-created if missing)
func findOrCreateWorkspace() (string, error) {
	cwd, err := os.Getwd()
	if err == nil {
		// Check local .anjal/
		localHidden := filepath.Join(cwd, ".anjal")
		if info, err := os.Stat(localHidden); err == nil && info.IsDir() {
			return localHidden, nil
		}
		// Check local anjal/

		localVisible := filepath.Join(cwd, "anjal")
		if info, err := os.Stat(localVisible); err == nil && info.IsDir() {
			return localVisible, nil
		}

	}

	// Fallback to Global Workspace (~/.anjal)
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}

	globalPath := filepath.Join(home, ".anjal")

	// Check if global exists
	if info, err := os.Stat(globalPath); err == nil && info.IsDir() {
		return globalPath, nil
	}

	// If we reach here, no workspace exists anywhere. We must create the global one.
	err = os.MkdirAll(globalPath, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create global workspace: %w", err)
	}

	// Create a friendly default file for new users
	err = CreateWelcomeFile(globalPath)
	if err != nil {
		return "", fmt.Errorf("failed to create welcome file: %w", err)
	}

	return globalPath, nil
}

// CreateWelcomeFile drops a default markdown file into a newly created workspace
func CreateWelcomeFile(workspacePath string) error {
	welcomePath := filepath.Join(workspacePath, "welcome.md")

	content := []byte(`# Welcome to Anjal

This is your default scratchpad. You can create as many ` + "`.md`" + ` files as you want in this folder. Anjal will parse them automatically.

## Sample Request: JSON Placeholder

` + "```http" + `
GET https://jsonplaceholder.typicode.com/todos/1
` + "```" + `
`)

	return os.WriteFile(welcomePath, content, 0644)
}
