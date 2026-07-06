// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/yogasimman/anjal/internal/env"
	"github.com/yogasimman/anjal/internal/httpclient"
	"github.com/yogasimman/anjal/internal/models"
	"github.com/yogasimman/anjal/internal/parser"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// shutdown is called when the app closes. It ensures background processes terminate.
func (a *App) shutdown(ctx context.Context) {
	fmt.Println("Shutting down Anjal GUI gracefully...")
}

// GetCollections parses the local workspace and returns all collections.
func (a *App) GetCollections() ([]models.Collection, error) {
	return parser.LoadWorkspace()
}

// GetCollectionsFrom parses a specific directory for collections.
func (a *App) GetCollectionsFrom(dir string) ([]models.Collection, error) {
	return parser.LoadWorkspaceFrom(dir)
}

// PromptOpenWorkspace prompts the user to select a folder, ensures .anjal exists, and returns the path.
func (a *App) PromptOpenWorkspace() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Anjal Workspace Directory",
	})
	if err != nil {
		return "", err
	}
	if dir == "" {
		return "", nil // User cancelled
	}

	anjalDir := filepath.Join(dir, ".anjal")
	if info, err := os.Stat(anjalDir); err != nil || !info.IsDir() {
		// Initialize the workspace
		err = os.MkdirAll(anjalDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to initialize .anjal in workspace: %v", err)
		}
		parser.CreateWelcomeFile(anjalDir)
	}

	return anjalDir, nil
}

// ExecuteRequest executes a given API request and returns the response.
func (a *App) ExecuteRequest(req models.APIRequest) (models.APIResponse, error) {
	return httpclient.Execute(a.ctx, req)
}

// PromptOpenFile opens a file dialog and attempts to parse the selected markdown file.
func (a *App) PromptOpenFile() (*models.Collection, error) {
	filePath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open Anjal Collection",
		Filters: []runtime.FileFilter{
			{DisplayName: "Markdown Files (*.md)", Pattern: "*.md"},
		},
	})

	if err != nil {
		return nil, err
	}

	if filePath == "" {
		return nil, nil // User cancelled
	}

	requests, globalAuth, err := parser.ParseFile(filePath)
	if err != nil {
		return nil, err
	}

	if len(requests) == 0 {
		return nil, fmt.Errorf("no valid HTTP requests found in this file")
	}

	col := models.Collection{
		Name:     filepath.Base(filePath),
		FilePath: filePath,
		Auth:     globalAuth,
		Requests: requests,
	}

	if globalAuth != nil {
		for i := range col.Requests {
			if col.Requests[i].Auth == nil {
				col.Requests[i].Auth = globalAuth
			}
		}
	}

	return &col, nil
}

// CreateRequest adds a new request to the specified markdown collection file.
func (a *App) CreateRequest(collectionPath string, req models.APIRequest) error {
	return parser.AddRequest(collectionPath, req)
}

// UpdateRequest updates an existing request in a collection file.
func (a *App) UpdateRequest(collectionPath string, reqID string, req models.APIRequest) error {
	return parser.UpdateRequest(collectionPath, reqID, req)
}

// DeleteRequest removes a request from a collection file.
func (a *App) DeleteRequest(collectionPath string, reqID string) error {
	return parser.DeleteRequest(collectionPath, reqID)
}

// DeleteCollection deletes a markdown collection file.
func (a *App) DeleteCollection(collectionPath string) error {
	return os.Remove(collectionPath)
}

// CreateCollection creates a new markdown collection file in the specified workspace.
func (a *App) CreateCollection(workspaceDir, name string) error {
	if name == "" {
		return fmt.Errorf("collection name cannot be empty")
	}
	if !strings.HasSuffix(name, ".md") {
		name += ".md"
	}
	
	// If workspaceDir is empty or ".", try to find the active workspace
	if workspaceDir == "" || workspaceDir == "." {
		ws, err := env.ResolveWriteDir()
		if err == nil && ws != "" {
			workspaceDir = ws
		} else {
			// Fallback to the default workspace
			home, _ := os.UserHomeDir()
			workspaceDir = filepath.Join(home, ".anjal")
		}
	}

	filePath := filepath.Join(workspaceDir, name)
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("collection file already exists")
	}

	// Add a dummy request so the parser picks it up as a valid collection
	content := fmt.Sprintf("# %s\n\n```http\nGET https://api.example.com\n```\n", strings.TrimSuffix(name, ".md"))
	return os.WriteFile(filePath, []byte(content), 0644)
}

// LoadEnvForCollection loads environment variables scoped to a specific collection.
func (a *App) LoadEnvForCollection(collectionName string) (map[string]string, error) {
	return env.LoadForCollection(collectionName)
}

// SaveEnvForCollection saves a specific environment variable for a collection.
func (a *App) SaveEnvForCollection(collectionName, key, value string) error {
	return env.Save(collectionName, key, value)
}

// DeleteEnvForCollection deletes a specific environment variable for a collection.
func (a *App) DeleteEnvForCollection(collectionName, key string) error {
	return env.Delete(collectionName, key)
}

// PromptSaveFilePath prompts the user for a save location and returns the chosen filepath.
func (a *App) PromptSaveFilePath(title, defaultName string) (string, error) {
	filePath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           title,
		DefaultFilename: defaultName,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files (*.json)", Pattern: "*.json"},
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "Markdown Files (*.md)", Pattern: "*.md"},
		},
	})

	if err != nil {
		return "", err
	}

	return filePath, nil
}

// WriteFile writes the given content to the specified filepath.
func (a *App) WriteFile(filePath, content string) error {
	return os.WriteFile(filePath, []byte(content), 0644)
}
