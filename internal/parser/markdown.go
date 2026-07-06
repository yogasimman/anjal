// Copyright (c) 2026 Yogasimman Ravisagar
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"hash/fnv"
	"os"
	"strings"

	"github.com/yogasimman/anjal/internal/httpclient"
	"github.com/yogasimman/anjal/internal/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// ParseFile reads a markdown file from disk, parses the AST, and extracts all
// valid HTTP request blocks and the global Collection-level Auth.
func ParseFile(filepath string) ([]models.APIRequest, *models.Auth, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %w", err)
	}

	return Parse(content)
}

// Parse reads markdown content from a byte slice, parses the AST, and extracts
// all valid HTTP request blocks and the global Collection-level Auth.
func Parse(content []byte) ([]models.APIRequest, *models.Auth, error) {
	md := goldmark.New()
	reader := text.NewReader(content)
	doc := md.Parser().Parse(reader)

	var requests []models.APIRequest
	var currentTitle string
	var globalAuth *models.Auth

	// Walk the AST looking for Headings, FencedCodeBlocks, and Paragraphs
	err := ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.Heading:
			// Capture the text of the heading to use as the title for the next API request
			currentTitle = string(n.Text(content))

		case *ast.Paragraph:
			// Scan standard paragraphs for global directives like @auth
			lines := n.Lines()
			for i := 0; i < lines.Len(); i++ {
				seg := lines.At(i)
				line := strings.TrimSpace(string(seg.Value(content)))
				if strings.HasPrefix(line, "@auth") {
					if auth, err := extractAuth(line); err == nil {
						globalAuth = auth
					} else {
						fmt.Printf("Warning: invalid global auth: %v\n", err)
					}
				}
			}

		case *ast.FencedCodeBlock:
			// Only process code blocks labeled with ```http
			lang := string(n.Language(content))
			if strings.ToLower(lang) == "http" {
				var rawText bytes.Buffer
				lines := n.Lines()
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					rawText.Write(line.Value(content))
				}

				req, err := parseHTTPBlock(currentTitle, rawText.String())
				if err != nil {
					fmt.Printf("Warning: Failed to parse HTTP block '%s': %v\n", currentTitle, err)
				} else {
					// Ensure ID uniqueness
					originalID := req.ID
					counter := 1
					for {
						isDuplicate := false
						for _, existingReq := range requests {
							if existingReq.ID == req.ID {
								isDuplicate = true
								break
							}
						}
						if !isDuplicate {
							break
						}
						req.ID = fmt.Sprintf("%s-%d", originalID, counter)
						counter++
					}
					requests = append(requests, req)
				}
			}
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error traversing markdown AST: %w", err)
	}

	return requests, globalAuth, nil
}

// parseHTTPBlock translates the raw text inside an ```http block into an APIRequest.
func parseHTTPBlock(title, block string) (models.APIRequest, error) {
	req := models.APIRequest{
		Title:       title,
		QueryParams: make(map[string]string),
		Headers:     make(map[string]string),
	}

	if req.Title == "" {
		req.Title = "Untitled Request"
	}

	scanner := bufio.NewScanner(strings.NewReader(block))
	isBody := false
	var bodyBuilder strings.Builder

	// Read the first line (Method and URL)
	if scanner.Scan() {
		firstLine := strings.TrimSpace(scanner.Text())
		parts := strings.SplitN(firstLine, " ", 2)
		if len(parts) != 2 {
			return req, fmt.Errorf("invalid first line, expected 'METHOD URL', got: %s", firstLine)
		}
		req.Method = strings.ToUpper(parts[0])
		req.URL = parts[1]
	}

	var explicitID string

	for scanner.Scan() {
		line := scanner.Text()

		if isBody {
			bodyBuilder.WriteString(line + "\n")
			continue
		}

		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			isBody = true
			continue
		}

		if strings.HasPrefix(trimmedLine, "@") {
			parts := strings.Fields(trimmedLine)
			if len(parts) >= 2 && parts[0] == "@id" {
				explicitID = strings.Join(parts[1:], " ")
			}
			parseDirective(trimmedLine, &req)
			continue
		}

		if colonIdx := strings.Index(trimmedLine, ":"); colonIdx != -1 {
			key := strings.TrimSpace(trimmedLine[:colonIdx])
			value := strings.TrimSpace(trimmedLine[colonIdx+1:])
			req.Headers[key] = value
		}
	}

	req.Body = strings.TrimSuffix(bodyBuilder.String(), "\n")

	if explicitID != "" {
		req.ID = explicitID
	} else {
		req.ID = generateID(req.Title, req.Method, req.URL)
	}

	return req, nil
}

func generateID(title, method, urlStr string) string {
	h := fnv.New64a()
	h.Write([]byte(title + "|" + method + "|" + urlStr))
	return fmt.Sprintf("req-%x", h.Sum64())[:15]
}

// extractAuth is a reusable helper to parse an @auth line into a *models.Auth.
func extractAuth(line string) (*models.Auth, error) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil, fmt.Errorf("missing auth type")
	}

	authType := parts[1]
	params := make(map[string]string)

	switch authType {
	case "bearer":
		if len(parts) >= 3 {
			params["token"] = strings.Join(parts[2:], " ")
		}
	case "basic":
		if len(parts) >= 4 {
			params["username"] = parts[2]
			params["password"] = parts[3]
		}
	case "apikey":
		if len(parts) >= 3 {
			params["key"] = parts[2]
		}
		if len(parts) >= 4 {
			params["header"] = parts[3]
		}
	case "custom":
		if len(parts) >= 3 {
			params["prefix"] = parts[2]
		}
		if len(parts) >= 4 {
			params["token"] = strings.Join(parts[3:], " ")
		}
	case "cookie":
		if len(parts) >= 4 {
			params["name"] = parts[2]
			params["value"] = strings.Join(parts[3:], " ")
		}
	default:
		return nil, fmt.Errorf("unknown auth type '%s'", authType)
	}

	return httpclient.FillAuth(authType, params)
}

func parseDirective(line string, req *models.APIRequest) {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return
	}

	directive := parts[0]

	switch directive {
	case "@query":
		if len(parts) >= 3 {
			key := parts[1]
			value := strings.Join(parts[2:], " ")
			req.QueryParams[key] = value
		}
	case "@header":
		rest := strings.TrimSpace(line[len("@header"):])
		if colonIdx := strings.Index(rest, ":"); colonIdx != -1 {
			key := strings.TrimSpace(rest[:colonIdx])
			value := strings.TrimSpace(rest[colonIdx+1:])
			req.Headers[key] = value
		} else {
			fields := strings.Fields(rest)
			if len(fields) >= 2 {
				req.Headers[fields[0]] = strings.Join(fields[1:], " ")
			}
		}
	case "@id":
		// Already captured in parseHTTPBlock; no-op here
	case "@auth":
		if auth, err := extractAuth(line); err == nil {
			req.Auth = auth
		} else {
			fmt.Printf("Warning: invalid auth: %v\n", err)
		}
	}
}

// ---------------------------------------------------------------------------
// CRUD operations — the markdown file is the database
// ---------------------------------------------------------------------------

// FindByID returns the first request with the given ID, or nil.
func FindByID(requests []models.APIRequest, id string) *models.APIRequest {
	for i := range requests {
		if requests[i].ID == id {
			return &requests[i]
		}
	}
	return nil
}

// RequestToMarkdown serializes a single APIRequest back into a markdown block.
func RequestToMarkdown(req models.APIRequest) string {
	var b strings.Builder

	// Heading
	b.WriteString("# " + req.Title + "\n\n")

	// Fenced code block
	b.WriteString("```http\n")
	b.WriteString(req.Method + " " + req.URL + "\n")

	// Explicit ID (if it doesn't match the auto-generated one)
	autoID := generateID(req.Title, req.Method, req.URL)
	if req.ID != autoID {
		b.WriteString("@id " + req.ID + "\n")
	}

	// Query params
	for k, v := range req.QueryParams {
		b.WriteString("@query " + k + " " + v + "\n")
	}

	// Auth
	if req.Auth != nil && req.Auth.Type != "none" {
		b.WriteString("@auth " + req.Auth.Type)
		switch req.Auth.Type {
		case "bearer":
			if token := req.Auth.Params["token"]; token != "" {
				b.WriteString(" " + token)
			}
		case "basic":
			if u := req.Auth.Params["username"]; u != "" {
				b.WriteString(" " + u)
			}
			if p := req.Auth.Params["password"]; p != "" {
				b.WriteString(" " + p)
			}
		case "apikey":
			if k := req.Auth.Params["key"]; k != "" {
				b.WriteString(" " + k)
			}
			if h := req.Auth.Params["header"]; h != "" {
				b.WriteString(" " + h)
			}
		case "custom":
			if p := req.Auth.Params["prefix"]; p != "" {
				b.WriteString(" " + p)
			}
			if t := req.Auth.Params["token"]; t != "" {
				b.WriteString(" " + t)
			}
		case "cookie":
			if n := req.Auth.Params["name"]; n != "" {
				b.WriteString(" " + n)
			}
			if v := req.Auth.Params["value"]; v != "" {
				b.WriteString(" " + v)
			}
		}
		b.WriteString("\n")
	}

	// Headers
	for k, v := range req.Headers {
		b.WriteString(k + ": " + v + "\n")
	}

	// Body
	if req.Body != "" {
		b.WriteString("\n" + req.Body + "\n")
	}

	b.WriteString("```\n")
	return b.String()
}

// AddRequest appends a request to the markdown file. If the request has no ID,
// one is generated automatically.
func AddRequest(filepath string, req models.APIRequest) error {
	if req.ID == "" {
		req.ID = generateID(req.Title, req.Method, req.URL)
	}

	block := RequestToMarkdown(req)

	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for append: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString("\n" + block); err != nil {
		return fmt.Errorf("failed to write request: %w", err)
	}

	return nil
}

// SaveAll overwrites the markdown file with the full list of requests.
// Use this after modifying the in-memory slice (e.g., after Update/Delete).
func SaveAll(filepath string, requests []models.APIRequest) error {
	var b strings.Builder
	for i, req := range requests {
		b.WriteString(RequestToMarkdown(req))
		if i < len(requests)-1 {
			b.WriteString("\n")
		}
	}

	if err := os.WriteFile(filepath, []byte(b.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// UpdateRequest reads a markdown file, finds the request with the given ID,
// replaces it, and writes the file back.
func UpdateRequest(filepath string, id string, updated models.APIRequest) error {
	requests, _, err := ParseFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	found := false
	for i := range requests {
		if requests[i].ID == id {
			updated.ID = id // preserve the original ID
			requests[i] = updated
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("request with ID '%s' not found", id)
	}

	return SaveAll(filepath, requests)
}

// DeleteRequest reads a markdown file, removes the request with the given ID,
// and writes the file back.
func DeleteRequest(filepath string, id string) error {
	requests, _, err := ParseFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	filtered := requests[:0]
	found := false
	for _, req := range requests {
		if req.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, req)
	}

	if !found {
		return fmt.Errorf("request with ID '%s' not found", id)
	}

	return SaveAll(filepath, filtered)
}
