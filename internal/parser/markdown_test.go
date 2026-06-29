package parser

import (
	"os"
	"testing"

	"github.com/yogasimman/anjal/internal/models"
)

// ===========================================================================
// parseHTTPBlock tests
// ===========================================================================

func TestParseHTTPBlock_BasicGET(t *testing.T) {
	block := `GET https://api.example.com/users`
	req, err := parseHTTPBlock("Get Users", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Title != "Get Users" {
		t.Errorf("Expected title 'Get Users', got '%s'", req.Title)
	}
	if req.Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", req.Method)
	}
	if req.URL != "https://api.example.com/users" {
		t.Errorf("Expected URL 'https://api.example.com/users', got '%s'", req.URL)
	}
}

func TestParseHTTPBlock_Untitled(t *testing.T) {
	block := `GET https://api.example.com/users`
	req, err := parseHTTPBlock("", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Title != "Untitled Request" {
		t.Errorf("Expected title 'Untitled Request', got '%s'", req.Title)
	}
}

func TestParseHTTPBlock_InvalidFirstLine(t *testing.T) {
	block := `INVALID_LINE`
	_, err := parseHTTPBlock("Test", block)
	if err == nil {
		t.Fatal("Expected error for invalid first line, got nil")
	}
}

func TestParseHTTPBlock_WithHeaders(t *testing.T) {
	block := `POST https://api.example.com/data
Content-Type: application/json
X-Custom-Header: custom-value
Accept: application/json`
	req, err := parseHTTPBlock("Create Data", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", req.Headers["Content-Type"])
	}
	if req.Headers["X-Custom-Header"] != "custom-value" {
		t.Errorf("Expected X-Custom-Header 'custom-value', got '%s'", req.Headers["X-Custom-Header"])
	}
	if req.Headers["Accept"] != "application/json" {
		t.Errorf("Expected Accept 'application/json', got '%s'", req.Headers["Accept"])
	}
}

func TestParseHTTPBlock_WithBody(t *testing.T) {
	block := `POST https://api.example.com/data
Content-Type: application/json

{"name": "anjal", "version": 1}`
	req, err := parseHTTPBlock("Create", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Body != `{"name": "anjal", "version": 1}` {
		t.Errorf("Expected body, got '%s'", req.Body)
	}
}

func TestParseHTTPBlock_MultiLineBody(t *testing.T) {
	block := `POST https://api.example.com/data
Content-Type: application/json

{
  "name": "anjal",
  "version": 1
}`
	req, err := parseHTTPBlock("Create", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	expected := `{
  "name": "anjal",
  "version": 1
}`
	if req.Body != expected {
		t.Errorf("Expected multi-line body, got '%s'", req.Body)
	}
}

func TestParseHTTPBlock_NoBlankLine_NoBody(t *testing.T) {
	// No blank line means no body — everything after first line is headers
	block := `GET https://api.example.com/health
Accept: text/plain`
	req, err := parseHTTPBlock("Health", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Body != "" {
		t.Errorf("Expected no body, got '%s'", req.Body)
	}
	if req.Headers["Accept"] != "text/plain" {
		t.Errorf("Expected Accept 'text/plain', got '%s'", req.Headers["Accept"])
	}
}

// ===========================================================================
// @directive tests
// ===========================================================================

func TestParseHTTPBlock_QueryDirective(t *testing.T) {
	block := `GET https://api.example.com/search
@query q golang
@query limit 10`
	req, err := parseHTTPBlock("Search", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.QueryParams["q"] != "golang" {
		t.Errorf("Expected q='golang', got '%s'", req.QueryParams["q"])
	}
	if req.QueryParams["limit"] != "10" {
		t.Errorf("Expected limit='10', got '%s'", req.QueryParams["limit"])
	}
}

func TestParseHTTPBlock_HeaderDirective(t *testing.T) {
	block := `GET https://api.example.com/data
@header Authorization: Bearer secret123
@header X-Trace-Id trace-456`
	req, err := parseHTTPBlock("Data", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Headers["Authorization"] != "Bearer secret123" {
		t.Errorf("Expected Authorization 'Bearer secret123', got '%s'", req.Headers["Authorization"])
	}
	if req.Headers["X-Trace-Id"] != "trace-456" {
		t.Errorf("Expected X-Trace-Id 'trace-456', got '%s'", req.Headers["X-Trace-Id"])
	}
}

func TestParseHTTPBlock_HeaderDirective_SpaceSyntax(t *testing.T) {
	// @header Key Value  (without colon)
	block := `GET https://api.example.com/data
@header X-Api-Key sk-abc123`
	req, err := parseHTTPBlock("Data", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Headers["X-Api-Key"] != "sk-abc123" {
		t.Errorf("Expected X-Api-Key 'sk-abc123', got '%s'", req.Headers["X-Api-Key"])
	}
}

// ===========================================================================
// @auth directive tests
// ===========================================================================

func TestParseHTTPBlock_AuthBearer(t *testing.T) {
	block := `GET https://api.example.com/profile
@auth bearer eyJhbGciOiJIUzI1NiJ9.abc.def`
	req, err := parseHTTPBlock("Profile", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Auth == nil {
		t.Fatal("Expected Auth to be set")
	}
	if req.Auth.Type != "bearer" {
		t.Errorf("Expected auth type 'bearer', got '%s'", req.Auth.Type)
	}
	if req.Auth.Params["token"] != "eyJhbGciOiJIUzI1NiJ9.abc.def" {
		t.Errorf("Expected token, got '%s'", req.Auth.Params["token"])
	}
}

func TestParseHTTPBlock_AuthBasic(t *testing.T) {
	block := `GET https://api.example.com/admin
@auth basic admin secret`
	req, err := parseHTTPBlock("Admin", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Auth == nil {
		t.Fatal("Expected Auth to be set")
	}
	if req.Auth.Type != "basic" {
		t.Errorf("Expected auth type 'basic', got '%s'", req.Auth.Type)
	}
	if req.Auth.Params["username"] != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", req.Auth.Params["username"])
	}
	if req.Auth.Params["password"] != "secret" {
		t.Errorf("Expected password 'secret', got '%s'", req.Auth.Params["password"])
	}
}

func TestParseHTTPBlock_AuthAPIKey(t *testing.T) {
	block := `GET https://api.example.com/data
@auth apikey sk-abc123`
	req, err := parseHTTPBlock("Data", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Auth == nil {
		t.Fatal("Expected Auth to be set")
	}
	if req.Auth.Type != "apikey" {
		t.Errorf("Expected auth type 'apikey', got '%s'", req.Auth.Type)
	}
	if req.Auth.Params["key"] != "sk-abc123" {
		t.Errorf("Expected key 'sk-abc123', got '%s'", req.Auth.Params["key"])
	}
	if req.Auth.Params["header"] != "X-API-Key" {
		t.Errorf("Expected default header 'X-API-Key', got '%s'", req.Auth.Params["header"])
	}
}

func TestParseHTTPBlock_AuthAPIKey_WithCustomHeader(t *testing.T) {
	block := `GET https://api.example.com/data
@auth apikey sk-abc123 X-Custom-Key`
	req, err := parseHTTPBlock("Data", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Auth.Params["header"] != "X-Custom-Key" {
		t.Errorf("Expected header 'X-Custom-Key', got '%s'", req.Auth.Params["header"])
	}
}

func TestParseHTTPBlock_AuthCustom(t *testing.T) {
	block := `GET https://api.example.com/data
@auth custom Token my-secret-token`
	req, err := parseHTTPBlock("Data", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Auth == nil {
		t.Fatal("Expected Auth to be set")
	}
	if req.Auth.Type != "custom" {
		t.Errorf("Expected auth type 'custom', got '%s'", req.Auth.Type)
	}
	if req.Auth.Params["prefix"] != "Token" {
		t.Errorf("Expected prefix 'Token', got '%s'", req.Auth.Params["prefix"])
	}
	if req.Auth.Params["token"] != "my-secret-token" {
		t.Errorf("Expected token 'my-secret-token', got '%s'", req.Auth.Params["token"])
	}
}

func TestParseHTTPBlock_AuthCookie(t *testing.T) {
	block := `GET https://api.example.com/profile
@auth cookie session_id abc123`
	req, err := parseHTTPBlock("Profile", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.Auth == nil {
		t.Fatal("Expected Auth to be set")
	}
	if req.Auth.Type != "cookie" {
		t.Errorf("Expected auth type 'cookie', got '%s'", req.Auth.Type)
	}
	if req.Auth.Params["name"] != "session_id" {
		t.Errorf("Expected name 'session_id', got '%s'", req.Auth.Params["name"])
	}
	if req.Auth.Params["value"] != "abc123" {
		t.Errorf("Expected value 'abc123', got '%s'", req.Auth.Params["value"])
	}
}

func TestParseHTTPBlock_AuthInvalid_Logged(t *testing.T) {
	// Missing token for bearer should log a warning but not crash
	block := `GET https://api.example.com/profile
@auth bearer`
	req, err := parseHTTPBlock("Profile", block)
	if err != nil {
		t.Fatalf("Expected no error (invalid auth is logged, not fatal), got %v", err)
	}
	if req.Auth != nil {
		t.Errorf("Expected Auth to be nil when FillAuth fails, got %+v", req.Auth)
	}
}

// ===========================================================================
// Full document integration test
// ===========================================================================

func TestParse_MultipleRequests(t *testing.T) {
	content := []byte(`# Get Users

` + "```http" + `
GET https://api.example.com/users
Accept: application/json
` + "```" + `

# Create User

` + "```http" + `
POST https://api.example.com/users
Content-Type: application/json

{"name": "anjal"}
` + "```" + `

# Delete User

` + "```http" + `
DELETE https://api.example.com/users/42
@auth bearer my-jwt-token
` + "```" + `
`)

	requests, err := Parse(content)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(requests) != 3 {
		t.Fatalf("Expected 3 requests, got %d", len(requests))
	}

	// Request 1: GET
	if requests[0].Title != "Get Users" {
		t.Errorf("Request 0 title: expected 'Get Users', got '%s'", requests[0].Title)
	}
	if requests[0].Method != "GET" {
		t.Errorf("Request 0 method: expected 'GET', got '%s'", requests[0].Method)
	}
	if requests[0].URL != "https://api.example.com/users" {
		t.Errorf("Request 0 URL: expected '...', got '%s'", requests[0].URL)
	}

	// Request 2: POST
	if requests[1].Title != "Create User" {
		t.Errorf("Request 1 title: expected 'Create User', got '%s'", requests[1].Title)
	}
	if requests[1].Method != "POST" {
		t.Errorf("Request 1 method: expected 'POST', got '%s'", requests[1].Method)
	}
	if requests[1].Body != `{"name": "anjal"}` {
		t.Errorf("Request 1 body: expected '...', got '%s'", requests[1].Body)
	}

	// Request 3: DELETE with auth
	if requests[2].Title != "Delete User" {
		t.Errorf("Request 2 title: expected 'Delete User', got '%s'", requests[2].Title)
	}
	if requests[2].Method != "DELETE" {
		t.Errorf("Request 2 method: expected 'DELETE', got '%s'", requests[2].Method)
	}
	if requests[2].Auth == nil || requests[2].Auth.Type != "bearer" {
		t.Error("Request 2: expected bearer auth")
	}
}

func TestParse_NonHTTPBlock_Skipped(t *testing.T) {
	content := []byte(`# Some JSON

` + "```json" + `
{"key": "value"}
` + "```" + `

# Real Request

` + "```http" + `
GET https://api.example.com/health
` + "```" + `
`)

	requests, err := Parse(content)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("Expected 1 request (json block skipped), got %d", len(requests))
	}
	if requests[0].Title != "Real Request" {
		t.Errorf("Expected title 'Real Request', got '%s'", requests[0].Title)
	}
}

func TestParse_NoHTTPBlocks(t *testing.T) {
	content := []byte("# Just a heading\n\nSome text without any HTTP blocks.\n")
	requests, err := Parse(content)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(requests) != 0 {
		t.Fatalf("Expected 0 requests, got %d", len(requests))
	}
}

func TestParse_DirectivesWithQueryParams(t *testing.T) {
	content := []byte(`# Search Users

` + "```http" + `
GET https://api.example.com/search
@query name anjal
@query status active
` + "```" + `
`)

	requests, err := Parse(content)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	if requests[0].QueryParams["name"] != "anjal" {
		t.Errorf("Expected name='anjal', got '%s'", requests[0].QueryParams["name"])
	}
	if requests[0].QueryParams["status"] != "active" {
		t.Errorf("Expected status='active', got '%s'", requests[0].QueryParams["status"])
	}
}

// ===========================================================================
// File-based ParseFile integration test
// ===========================================================================

func TestParseFile_ValidFile(t *testing.T) {
	requests, err := ParseFile("../../testdata/sample.md")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(requests))
	}

	if requests[0].Title != "Get All Users" {
		t.Errorf("Expected title 'Get All Users', got '%s'", requests[0].Title)
	}
	if requests[0].Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", requests[0].Method)
	}

	if requests[1].Title != "Login" {
		t.Errorf("Expected title 'Login', got '%s'", requests[1].Title)
	}
	if requests[1].Method != "POST" {
		t.Errorf("Expected method 'POST', got '%s'", requests[1].Method)
	}
	if requests[1].Auth == nil || requests[1].Auth.Type != "bearer" {
		t.Error("Expected bearer auth on login request")
	}
}

func TestParseFile_NonExistent(t *testing.T) {
	_, err := ParseFile("../../testdata/does-not-exist.md")
	if err == nil {
		t.Fatal("Expected error for non-existent file, got nil")
	}
}

// ===========================================================================
// Helper: verify ApplyAuth integration with parsed requests
// ===========================================================================

func TestParsedRequest_ApplyAuth_EndToEnd(t *testing.T) {
	content := []byte(`# Authenticated Request

` + "```http" + `
GET https://api.example.com/secure
@auth bearer my-jwt-token
` + "```" + `
`)

	requests, err := Parse(content)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}

	req := requests[0]
	if req.Auth == nil {
		t.Fatal("Expected Auth to be populated by parser")
	}
	if req.Auth.Type != "bearer" {
		t.Errorf("Expected bearer, got '%s'", req.Auth.Type)
	}
	if req.Auth.Params["token"] != "my-jwt-token" {
		t.Errorf("Expected token 'my-jwt-token', got '%s'", req.Auth.Params["token"])
	}

	// This auth should pass FillAuth validation (already validated during parsing)
	// The httpclient.Execute will call applyAuth internally to inject the header
}

// ===========================================================================
// SECTION — ID generation
// ===========================================================================

func TestParseHTTPBlock_AutoGeneratedID(t *testing.T) {
	block := `GET https://api.example.com/users`
	req, err := parseHTTPBlock("Get Users", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.ID == "" {
		t.Fatal("Expected auto-generated ID, got empty string")
	}
	// Same inputs should produce the same ID
	req2, _ := parseHTTPBlock("Get Users", block)
	if req.ID != req2.ID {
		t.Errorf("Expected stable ID, got '%s' vs '%s'", req.ID, req2.ID)
	}
}

func TestParseHTTPBlock_ExplicitID(t *testing.T) {
	block := `GET https://api.example.com/users
@id my-custom-id`
	req, err := parseHTTPBlock("Get Users", block)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if req.ID != "my-custom-id" {
		t.Errorf("Expected explicit ID 'my-custom-id', got '%s'", req.ID)
	}
}

func TestParseHTTPBlock_DifferentRequestsDifferentIDs(t *testing.T) {
	block1 := `GET https://api.example.com/users`
	block2 := `POST https://api.example.com/users`

	req1, _ := parseHTTPBlock("Get Users", block1)
	req2, _ := parseHTTPBlock("Create User", block2)

	if req1.ID == req2.ID {
		t.Errorf("Expected different IDs, got same '%s'", req1.ID)
	}
}

// ===========================================================================
// SECTION — RequestToMarkdown round-trip
// ===========================================================================

func TestRequestToMarkdown_RoundTrip(t *testing.T) {
	original := `GET https://api.example.com/users
Accept: application/json
@query status active
@auth bearer my-jwt`

	req, err := parseHTTPBlock("Get Users", original)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	serialized := RequestToMarkdown(req)

	// Parse it back
	requests, err := Parse([]byte(serialized))
	if err != nil {
		t.Fatalf("Failed to re-parse serialized output: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("Expected 1 request after round-trip, got %d", len(requests))
	}

	reparsed := requests[0]
	if reparsed.Title != req.Title {
		t.Errorf("Title: expected '%s', got '%s'", req.Title, reparsed.Title)
	}
	if reparsed.Method != req.Method {
		t.Errorf("Method: expected '%s', got '%s'", req.Method, reparsed.Method)
	}
	if reparsed.URL != req.URL {
		t.Errorf("URL: expected '%s', got '%s'", req.URL, reparsed.URL)
	}
	if reparsed.ID != req.ID {
		t.Errorf("ID: expected '%s', got '%s'", req.ID, reparsed.ID)
	}
	if len(reparsed.Headers) != len(req.Headers) {
		t.Errorf("Headers count: expected %d, got %d", len(req.Headers), len(reparsed.Headers))
	}
	if len(reparsed.QueryParams) != len(req.QueryParams) {
		t.Errorf("QueryParams count: expected %d, got %d", len(req.QueryParams), len(reparsed.QueryParams))
	}
	if reparsed.Auth == nil || reparsed.Auth.Type != req.Auth.Type {
		t.Error("Auth lost during round-trip")
	}
}

func TestRequestToMarkdown_ExplicitID(t *testing.T) {
	block := `GET https://api.example.com/data
@id my-explicit-id`

	req, _ := parseHTTPBlock("Data", block)
	serialized := RequestToMarkdown(req)

	// Re-parse and verify the explicit ID survived
	requests, _ := Parse([]byte(serialized))
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if requests[0].ID != "my-explicit-id" {
		t.Errorf("Expected ID 'my-explicit-id', got '%s'", requests[0].ID)
	}
}

// ===========================================================================
// SECTION — FindByID
// ===========================================================================

func TestFindByID_Found(t *testing.T) {
	req1 := models.APIRequest{ID: "req-aaa", Title: "First"}
	req2 := models.APIRequest{ID: "req-bbb", Title: "Second"}
	requests := []models.APIRequest{req1, req2}

	found := FindByID(requests, "req-bbb")
	if found == nil {
		t.Fatal("Expected to find request")
	}
	if found.Title != "Second" {
		t.Errorf("Expected 'Second', got '%s'", found.Title)
	}
}

func TestFindByID_NotFound(t *testing.T) {
	requests := []models.APIRequest{{ID: "req-aaa", Title: "First"}}

	found := FindByID(requests, "req-nonexistent")
	if found != nil {
		t.Errorf("Expected nil, got %+v", found)
	}
}

// ===========================================================================
// SECTION — CRUD: Add / Update / Delete (file-based)
// ===========================================================================

func TestAddRequest(t *testing.T) {
	filepath := "../../testdata/crud_test.md"
	os.Remove(filepath) // clean slate

	req := models.APIRequest{
		Title:  "New Request",
		Method: "GET",
		URL:    "https://api.example.com/health",
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}

	err := AddRequest(filepath, req)
	if err != nil {
		t.Fatalf("AddRequest failed: %v", err)
	}

	// Read back
	requests, err := ParseFile(filepath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if requests[0].Title != "New Request" {
		t.Errorf("Expected title 'New Request', got '%s'", requests[0].Title)
	}
	if requests[0].Method != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", requests[0].Method)
	}
	if requests[0].ID == "" {
		t.Error("Expected auto-generated ID")
	}

	os.Remove(filepath)
}

func TestAddRequest_Multiple(t *testing.T) {
	filepath := "../../testdata/crud_multi.md"
	os.Remove(filepath)

	req1 := models.APIRequest{Title: "First", Method: "GET", URL: "https://api.example.com/one"}
	req2 := models.APIRequest{Title: "Second", Method: "POST", URL: "https://api.example.com/two"}

	AddRequest(filepath, req1)
	AddRequest(filepath, req2)

	requests, _ := ParseFile(filepath)
	if len(requests) != 2 {
		t.Fatalf("Expected 2 requests, got %d", len(requests))
	}
	if requests[0].Title != "First" {
		t.Errorf("Expected 'First', got '%s'", requests[0].Title)
	}
	if requests[1].Title != "Second" {
		t.Errorf("Expected 'Second', got '%s'", requests[1].Title)
	}

	os.Remove(filepath)
}

func TestUpdateRequest(t *testing.T) {
	filepath := "../../testdata/crud_update.md"
	os.Remove(filepath)

	// Add a request
	original := models.APIRequest{
		Title:  "Original",
		Method: "GET",
		URL:    "https://api.example.com/old",
	}
	AddRequest(filepath, original)

	// Parse to get the auto-generated ID
	requests, _ := ParseFile(filepath)
	originalID := requests[0].ID

	// Update it
	updated := models.APIRequest{
		Title:  "Updated",
		Method: "POST",
		URL:    "https://api.example.com/new",
		Body:   `{"key": "value"}`,
	}
	err := UpdateRequest(filepath, originalID, updated)
	if err != nil {
		t.Fatalf("UpdateRequest failed: %v", err)
	}

	// Verify
	requests, _ = ParseFile(filepath)
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request after update, got %d", len(requests))
	}
	if requests[0].Title != "Updated" {
		t.Errorf("Expected title 'Updated', got '%s'", requests[0].Title)
	}
	if requests[0].Method != "POST" {
		t.Errorf("Expected method 'POST', got '%s'", requests[0].Method)
	}
	if requests[0].ID != originalID {
		t.Errorf("ID changed: expected '%s', got '%s'", originalID, requests[0].ID)
	}
	if requests[0].Body != `{"key": "value"}` {
		t.Errorf("Expected body, got '%s'", requests[0].Body)
	}

	os.Remove(filepath)
}

func TestUpdateRequest_NotFound(t *testing.T) {
	filepath := "../../testdata/crud_update_nf.md"
	os.Remove(filepath)

	AddRequest(filepath, models.APIRequest{Title: "Only", Method: "GET", URL: "https://api.example.com/x"})

	err := UpdateRequest(filepath, "req-nonexistent", models.APIRequest{Title: "X"})
	if err == nil {
		t.Fatal("Expected error for non-existent ID, got nil")
	}

	os.Remove(filepath)
}

func TestDeleteRequest(t *testing.T) {
	filepath := "../../testdata/crud_delete.md"
	os.Remove(filepath)

	req1 := models.APIRequest{Title: "Keep", Method: "GET", URL: "https://api.example.com/keep"}
	req2 := models.APIRequest{Title: "Remove", Method: "DELETE", URL: "https://api.example.com/remove"}

	AddRequest(filepath, req1)
	AddRequest(filepath, req2)

	// Parse to find the ID of the one to delete
	requests, _ := ParseFile(filepath)
	var removeID string
	for _, r := range requests {
		if r.Title == "Remove" {
			removeID = r.ID
			break
		}
	}

	err := DeleteRequest(filepath, removeID)
	if err != nil {
		t.Fatalf("DeleteRequest failed: %v", err)
	}

	requests, _ = ParseFile(filepath)
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request after delete, got %d", len(requests))
	}
	if requests[0].Title != "Keep" {
		t.Errorf("Expected 'Keep', got '%s'", requests[0].Title)
	}

	os.Remove(filepath)
}

func TestDeleteRequest_NotFound(t *testing.T) {
	filepath := "../../testdata/crud_delete_nf.md"
	os.Remove(filepath)
	AddRequest(filepath, models.APIRequest{Title: "Only", Method: "GET", URL: "https://api.example.com/x"})

	err := DeleteRequest(filepath, "req-nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent ID, got nil")
	}

	os.Remove(filepath)
}
