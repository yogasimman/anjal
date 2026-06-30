package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// ===========================================================================
// loadWorkspaceFrom tests
// ===========================================================================

func TestLoadWorkspaceFrom_ValidMDFiles(t *testing.T) {
	dir := t.TempDir()

	// Create two valid .md files with HTTP blocks
	writeFile(t, dir, "api.md", `# Get Users

`+"```http"+`
GET https://api.example.com/users
`+"```"+`
`)

	writeFile(t, dir, "auth.md", `# Login

`+"```http"+`
POST https://api.example.com/login
Content-Type: application/json

{"user":"admin"}
`+"```"+`
`)

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collections) != 2 {
		t.Fatalf("Expected 2 collections, got %d", len(collections))
	}

	// Check first collection
	if collections[0].Name != "api.md" {
		t.Errorf("Expected 'api.md', got '%s'", collections[0].Name)
	}
	if len(collections[0].Requests) != 1 {
		t.Errorf("Expected 1 request in api.md, got %d", len(collections[0].Requests))
	}
	if collections[0].Requests[0].Title != "Get Users" {
		t.Errorf("Expected title 'Get Users', got '%s'", collections[0].Requests[0].Title)
	}

	// Check second collection
	if collections[1].Name != "auth.md" {
		t.Errorf("Expected 'auth.md', got '%s'", collections[1].Name)
	}
	if len(collections[1].Requests) != 1 {
		t.Errorf("Expected 1 request in auth.md, got %d", len(collections[1].Requests))
	}
	if collections[1].Requests[0].Method != "POST" {
		t.Errorf("Expected POST, got '%s'", collections[1].Requests[0].Method)
	}
	if collections[1].Requests[0].Body != `{"user":"admin"}` {
		t.Errorf("Expected body, got '%s'", collections[1].Requests[0].Body)
	}
}

func TestLoadWorkspaceFrom_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collections) != 0 {
		t.Fatalf("Expected 0 collections for empty dir, got %d", len(collections))
	}
}

func TestLoadWorkspaceFrom_SkipsNonMDFiles(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "notes.txt", "just some text")
	writeFile(t, dir, "script.sh", "#!/bin/bash")
	writeFile(t, dir, "api.md", `# API

`+"```http"+`
GET https://api.example.com/ping
`+"```"+`
`)

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collections) != 1 {
		t.Fatalf("Expected 1 collection, got %d", len(collections))
	}
	if collections[0].Name != "api.md" {
		t.Errorf("Expected 'api.md', got '%s'", collections[0].Name)
	}
}

func TestLoadWorkspaceFrom_SkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()

	subDir := filepath.Join(dir, "subfolder")
	os.MkdirAll(subDir, 0755)
	writeFile(t, subDir, "nested.md", `# Nested

`+"```http"+`
GET https://api.example.com/nested
`+"```"+`
`)

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collections) != 0 {
		t.Fatalf("Expected 0 collections (subdirectory skipped), got %d", len(collections))
	}
}

func TestLoadWorkspaceFrom_MDWithNoHTTPBlocks(t *testing.T) {
	dir := t.TempDir()

	// A .md file that has no ```http blocks
	writeFile(t, dir, "notes.md", `# Just Notes

This file has no HTTP blocks.

`+"```json"+`
{"key": "value"}
`+"```"+`
`)

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// File exists but has no HTTP blocks → should not appear in collections
	if len(collections) != 0 {
		t.Fatalf("Expected 0 collections, got %d", len(collections))
	}
}

func TestLoadWorkspaceFrom_MixedFiles(t *testing.T) {
	dir := t.TempDir()

	// Mix: valid .md, .md with no HTTP blocks, non-.md file
	writeFile(t, dir, "api.md", `# API

`+"```http"+`
GET https://api.example.com/ping
`+"```"+`
`)
	writeFile(t, dir, "readme.md", "# Readme\n\nNo HTTP here.\n")
	writeFile(t, dir, "config.json", `{"key":"val"}`)

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collections) != 1 {
		t.Fatalf("Expected 1 collection, got %d", len(collections))
	}
	if collections[0].Name != "api.md" {
		t.Errorf("Expected 'api.md', got '%s'", collections[0].Name)
	}
}

func TestLoadWorkspaceFrom_NonExistentDirectory(t *testing.T) {
	_, err := loadWorkspaceFrom("/tmp/does-not-exist-anjal-test-12345")
	if err == nil {
		t.Fatal("Expected error for non-existent directory, got nil")
	}
}

func TestLoadWorkspaceFrom_MultipleRequestsPerFile(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "users.md", `# Get All

`+"```http"+`
GET https://api.example.com/users
`+"```"+`

# Create User

`+"```http"+`
POST https://api.example.com/users

{"name":"anjal"}
`+"```"+`

# Delete User

`+"```http"+`
DELETE https://api.example.com/users/1
`+"```"+`
`)

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collections) != 1 {
		t.Fatalf("Expected 1 collection, got %d", len(collections))
	}

	if len(collections[0].Requests) != 3 {
		t.Fatalf("Expected 3 requests in users.md, got %d", len(collections[0].Requests))
	}

	// Verify titles
	expectedTitles := []string{"Get All", "Create User", "Delete User"}
	for i, expected := range expectedTitles {
		if collections[0].Requests[i].Title != expected {
			t.Errorf("Request %d: expected title '%s', got '%s'", i, expected, collections[0].Requests[i].Title)
		}
	}
}

// ===========================================================================
// createWelcomeFile tests
// ===========================================================================

func TestCreateWelcomeFile(t *testing.T) {
	dir := t.TempDir()

	err := createWelcomeFile(dir)
	if err != nil {
		t.Fatalf("createWelcomeFile failed: %v", err)
	}

	// Verify file exists
	welcomePath := filepath.Join(dir, "welcome.md")
	if _, err := os.Stat(welcomePath); os.IsNotExist(err) {
		t.Fatal("Expected welcome.md to exist")
	}

	// Verify it's parseable and contains at least one request
	requests, _, err := ParseFile(welcomePath)
	if err != nil {
		t.Fatalf("Failed to parse welcome.md: %v", err)
	}

	if len(requests) == 0 {
		t.Fatal("Expected welcome.md to contain at least 1 HTTP request")
	}

	if requests[0].Title != "Sample Request: JSON Placeholder" {
		t.Errorf("Expected title, got '%s'", requests[0].Title)
	}
	if requests[0].Method != "GET" {
		t.Errorf("Expected GET, got '%s'", requests[0].Method)
	}
}

// ===========================================================================
// findOrCreateWorkspace tests
// ===========================================================================

func TestFindOrCreateWorkspace_LocalHidden(t *testing.T) {
	// Create a temp dir, chdir into it, and create .anjal/
	baseDir := t.TempDir()
	anjalDir := filepath.Join(baseDir, ".anjal")
	os.MkdirAll(anjalDir, 0755)

	t.Chdir(baseDir)

	path, err := findOrCreateWorkspace()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path != anjalDir {
		t.Errorf("Expected '%s', got '%s'", anjalDir, path)
	}
}

func TestFindOrCreateWorkspace_LocalVisible(t *testing.T) {
	// Create a temp dir, chdir into it, and create anjal/ (visible, no dot)
	baseDir := t.TempDir()
	anjalDir := filepath.Join(baseDir, "anjal")
	os.MkdirAll(anjalDir, 0755)

	t.Chdir(baseDir)

	path, err := findOrCreateWorkspace()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path != anjalDir {
		t.Errorf("Expected '%s', got '%s'", anjalDir, path)
	}
}

func TestFindOrCreateWorkspace_LocalHiddenPriority(t *testing.T) {
	// Both .anjal/ and anjal/ exist → .anjal/ should win
	baseDir := t.TempDir()
	hiddenDir := filepath.Join(baseDir, ".anjal")
	visibleDir := filepath.Join(baseDir, "anjal")
	os.MkdirAll(hiddenDir, 0755)
	os.MkdirAll(visibleDir, 0755)

	t.Chdir(baseDir)

	path, err := findOrCreateWorkspace()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path != hiddenDir {
		t.Errorf("Expected hidden '%s' to win, got '%s'", hiddenDir, path)
	}
}

func TestFindOrCreateWorkspace_LocalVisible_FallbackWhenNoHidden(t *testing.T) {
	// Only anjal/ exists, .anjal/ does not → anjal/ should be chosen
	baseDir := t.TempDir()
	visibleDir := filepath.Join(baseDir, "anjal")
	os.MkdirAll(visibleDir, 0755)

	t.Chdir(baseDir)

	path, err := findOrCreateWorkspace()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if path != visibleDir {
		t.Errorf("Expected '%s', got '%s'", visibleDir, path)
	}
}

// ===========================================================================
// Collection struct test
// ===========================================================================

func TestCollection_Fields(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")
	writeFile(t, dir, "test.md", `# Test

`+"```http"+`
GET https://api.example.com/health
`+"```"+`
`)

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collections) != 1 {
		t.Fatal("Expected 1 collection")
	}

	col := collections[0]
	if col.Name != "test.md" {
		t.Errorf("Expected Name 'test.md', got '%s'", col.Name)
	}
	if col.FilePath != filePath {
		t.Errorf("Expected FilePath '%s', got '%s'", filePath, col.FilePath)
	}
	if len(col.Requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(col.Requests))
	}
	if col.Requests[0].ID == "" {
		t.Error("Expected request to have an ID")
	}
}

// ===========================================================================
// Helper
// ===========================================================================

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file %s: %v", name, err)
	}
}

// ===========================================================================
// Cascading inheritance: collection-level global auth
// ===========================================================================

func TestLoadWorkspaceFrom_GlobalAuthInheritance(t *testing.T) {
	dir := t.TempDir()

	// File with global auth + requests that have no auth of their own
	writeFile(t, dir, "users.md", `@auth bearer global-jwt-token

# Get User

`+"```http"+`
GET https://api.example.com/users/1
`+"```"+`

# Create User

`+"```http"+`
POST https://api.example.com/users

{"name":"anjal"}
`+"```"+`
`)

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collections) != 1 {
		t.Fatalf("Expected 1 collection, got %d", len(collections))
	}

	col := collections[0]

	// Collection-level auth should be set
	if col.Auth == nil {
		t.Fatal("Expected collection-level auth to be set")
	}
	if col.Auth.Type != "bearer" {
		t.Errorf("Expected bearer, got '%s'", col.Auth.Type)
	}
	if col.Auth.Params["token"] != "global-jwt-token" {
		t.Errorf("Expected token 'global-jwt-token', got '%s'", col.Auth.Params["token"])
	}

	// Both requests should inherit the global auth
	if len(col.Requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(col.Requests))
	}

	for i, req := range col.Requests {
		if req.Auth == nil {
			t.Errorf("Request %d: expected Auth to be inherited from collection", i)
		} else if req.Auth.Type != "bearer" {
			t.Errorf("Request %d: expected bearer, got '%s'", i, req.Auth.Type)
		}
	}
}

func TestLoadWorkspaceFrom_RequestAuthOverridesGlobal(t *testing.T) {
	dir := t.TempDir()

	// Global basic auth, but one request has its own bearer auth
	writeFile(t, dir, "api.md", `@auth basic admin secret

# Public Endpoint

`+"```http"+`
GET https://api.example.com/public
`+"```"+`

# Secure Endpoint (overrides global)

`+"```http"+`
GET https://api.example.com/secure
@auth bearer override-token
`+"```"+`
`)

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collections) != 1 {
		t.Fatalf("Expected 1 collection, got %d", len(collections))
	}

	col := collections[0]

	if len(col.Requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(col.Requests))
	}

	// Request 0: should inherit global basic auth
	r0 := col.Requests[0]
	if r0.Auth == nil {
		t.Fatal("Request 0: expected Auth")
	}
	if r0.Auth.Type != "basic" {
		t.Errorf("Request 0: expected basic (inherited), got '%s'", r0.Auth.Type)
	}

	// Request 1: should have its own bearer auth (overrides global)
	r1 := col.Requests[1]
	if r1.Auth == nil {
		t.Fatal("Request 1: expected Auth")
	}
	if r1.Auth.Type != "bearer" {
		t.Errorf("Request 1: expected bearer (own auth), got '%s'", r1.Auth.Type)
	}
	if r1.Auth.Params["token"] != "override-token" {
		t.Errorf("Request 1: expected token 'override-token', got '%s'", r1.Auth.Params["token"])
	}
}

func TestLoadWorkspaceFrom_NoGlobalAuth(t *testing.T) {
	dir := t.TempDir()

	// File with no global auth
	writeFile(t, dir, "simple.md", `# Simple Request

`+"```http"+`
GET https://api.example.com/ping
`+"```"+`
`)

	collections, err := loadWorkspaceFrom(dir)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(collections) != 1 {
		t.Fatalf("Expected 1 collection, got %d", len(collections))
	}

	col := collections[0]

	// No global auth
	if col.Auth != nil {
		t.Errorf("Expected no collection-level auth, got %+v", col.Auth)
	}

	// Request should have no auth
	if len(col.Requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(col.Requests))
	}
	if col.Requests[0].Auth != nil {
		t.Errorf("Expected request to have no auth, got %+v", col.Requests[0].Auth)
	}
}
