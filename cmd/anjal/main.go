package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yogasimman/anjal/internal/env"
	"github.com/yogasimman/anjal/internal/httpclient"
	"github.com/yogasimman/anjal/internal/models"
	"github.com/yogasimman/anjal/internal/parser"
	"github.com/yogasimman/anjal/internal/tui"
)

var Version = "v1.0.0-dev"

func main() {
	args := os.Args[1:]
	
	noUI := false
	var filteredArgs []string
	for _, arg := range args {
		if arg == "-v" || arg == "--version" {
			fmt.Printf("Anjal %s\n", Version)
			return
		} else if arg == "--noui" {
			noUI = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	args = filteredArgs

	var collections []models.Collection
	var err error

	if len(args) >= 1 {
		// Specific file mode
		path := args[0]
		if noUI {
			runFile(path)
			return
		}
		
		if _, statErr := os.Stat(path); statErr == nil {
			requests, globalAuth, parseErr := parser.ParseFile(path)
			if parseErr != nil {
				fmt.Fprintf(os.Stderr, "Parse error: %v\n", parseErr)
				os.Exit(1)
			}
			if len(requests) > 0 {
				col := models.Collection{
					Name:     filepath.Base(path),
					FilePath: path,
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
				collections = append(collections, col)
			}
		} else {
			fmt.Fprintf(os.Stderr, "File not found: %s\n", path)
			os.Exit(1)
		}
	} else {
		// Workspace mode
		collections, err = parser.LoadWorkspace()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
			os.Exit(1)
		}
		
		if noUI {
			if len(collections) == 0 {
				fmt.Println("No API collections found in workspace.")
				return
			}
			fmt.Println("Available Collections:")
			for i, col := range collections {
				fmt.Printf("[%d] %s\n", i+1, col.Name)
			}
			fmt.Print("\nSelect a collection to run (1-", len(collections), "): ")
			
			reader := bufio.NewReader(os.Stdin)
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)
			idx, err := strconv.Atoi(input)
			if err == nil && idx >= 1 && idx <= len(collections) {
				fmt.Println()
				runFile(collections[idx-1].FilePath)
			} else {
				fmt.Println("Invalid selection.")
				os.Exit(1)
			}
			return
		}
	}

	var envVars map[string]string
	if len(collections) > 0 {
		envVars, _ = env.LoadForCollection(filepath.Base(collections[0].FilePath))
	}

	isWorkspace := len(args) == 0
	m := tui.InitialModel(collections, envVars, isWorkspace)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runFile(path string) {
	fmt.Printf("📄 Parsing %s...\n\n", path)

	requests, globalAuth, err := parser.ParseFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	if globalAuth != nil {
		fmt.Printf("🔐 Global auth: %s\n", globalAuth.Type)
	}

	fmt.Printf("Found %d request(s)\n\n", len(requests))

	envVars, _ := env.LoadForCollection(filepath.Base(path))
	ctx := context.Background()

	for i, req := range requests {
		resolveRequest(&req, envVars)
		fmt.Printf("─────────────────────────────────────────────\n")
		fmt.Printf("Request %d: %s %s\n", i+1, req.Method, req.URL)

		if req.Auth == nil && globalAuth != nil {
			req.Auth = globalAuth
		}

		if req.Auth != nil {
			fmt.Printf("   Auth: %s\n", req.Auth.Type)
		}

		resp, err := httpclient.Execute(ctx, req)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("   ✅ Status: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Printf("   ⏱  Latency: %v\n", resp.Latency)
		fmt.Printf("   📦 Content-Type: %s (%s)\n", resp.Headers["Content-Type"], resp.ContentType)

		preview := resp.Body
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		fmt.Printf("   📝 Body (%d bytes):\n", len(resp.Body))
		fmt.Printf("   %s\n\n", indent(preview, "   "))
	}
}

func resolveRequest(req *models.APIRequest, vars map[string]string) {
	req.URL = env.Resolve(req.URL, vars)
	req.Body = env.Resolve(req.Body, vars)
	for k, v := range req.Headers {
		req.Headers[k] = env.Resolve(v, vars)
	}
	for k, v := range req.QueryParams {
		req.QueryParams[k] = env.Resolve(v, vars)
	}
	// Force override Auth if workspace configured it
	authType := vars["WORKSPACE_AUTH_TYPE"]
	if authType != "" && authType != "none" {
		params := make(map[string]string)
		for k, v := range vars {
			if strings.HasPrefix(k, "WORKSPACE_AUTH_") && k != "WORKSPACE_AUTH_TYPE" {
				key := strings.ToLower(strings.TrimPrefix(k, "WORKSPACE_AUTH_"))
				params[key] = v
			}
		}
		req.Auth = &models.Auth{Type: authType, Params: params}
	} else if req.Auth != nil {
		for k, v := range req.Auth.Params {
			req.Auth.Params[k] = env.Resolve(v, vars)
		}
	}
	if req.Auth == nil {
		authType := vars["WORKSPACE_AUTH_TYPE"]
		if authType != "" {
			params := make(map[string]string)
			for k, v := range vars {
				if strings.HasPrefix(k, "WORKSPACE_AUTH_") && k != "WORKSPACE_AUTH_TYPE" {
					paramKey := strings.ToLower(strings.TrimPrefix(k, "WORKSPACE_AUTH_"))
					params[paramKey] = v
				}
			}
			req.Auth = &models.Auth{Type: authType, Params: params}
		}
	}
}

func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
