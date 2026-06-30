package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yogasimman/anjal/internal/env"
	"github.com/yogasimman/anjal/internal/httpclient"
	"github.com/yogasimman/anjal/internal/models"
	"github.com/yogasimman/anjal/internal/parser"
)

func main() {
	args := os.Args[1:]

	// If a file is passed on the command line, parse and execute it
	if len(args) >= 1 {
		runFile(args[0])
		return
	}

	// Otherwise, load the workspace
	collections, err := parser.LoadWorkspace()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading workspace: %v\n", err)
		os.Exit(1)
	}

	if len(collections) == 0 {
		fmt.Println("No API collections found in workspace.")
		fmt.Println("Create a .md file with HTTP blocks to get started.")
		fmt.Println("Or run: anjal testdata/google.md")
		return
	}

	for _, col := range collections {
		fmt.Printf("\n📁 %s (%d requests)\n", col.Name, len(col.Requests))
		for _, req := range col.Requests {
			fmt.Printf("   %s %s %s\n", req.Method, req.URL, dim("("+req.ID+")"))
		}
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

		// Apply global auth if request doesn't have its own
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

		// Preview body (first 500 chars)
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
	if req.Auth != nil {
		for k, v := range req.Auth.Params {
			req.Auth.Params[k] = env.Resolve(v, vars)
		}
	}
}

func dim(s string) string {
	return "\033[2m" + s + "\033[0m"
}

func indent(s, prefix string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
