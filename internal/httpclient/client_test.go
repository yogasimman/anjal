package httpclient

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yogasimman/anjal/internal/models"
)

// ===========================================================================
// SECTION 1 — Basic request/response integrity
// ===========================================================================

func TestExecute_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected method GET, got %s", r.Method)
		}
		if r.Header.Get("X-Test-Header") != "TestValue" {
			t.Errorf("Expected header 'X-Test-Header' to be 'TestValue', got '%s'", r.Header.Get("X-Test-Header"))
		}
		if r.URL.Query().Get("search query") != "golang tui" {
			t.Errorf("Expected query param 'search query' to be 'golang tui', got '%s'",
				r.URL.Query().Get("search query"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
		Headers: map[string]string{
			"X-Test-Header": "TestValue",
		},
		QueryParams: map[string]string{
			"search query": "golang tui",
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if resp.Body != `{"status":"success"}` {
		t.Errorf("Expected body `{\"status\":\"success\"}`, got `%s`", resp.Body)
	}
	if resp.Latency <= 0 {
		t.Errorf("Expected positive latency, got %v", resp.Latency)
	}
}

func TestExecute_MinimalRequest(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if resp.Body != "ok" {
		t.Errorf("Expected body 'ok', got `%s`", resp.Body)
	}
}

// ===========================================================================
// SECTION 2 — BODY: transmission & values
// ===========================================================================

func TestExecute_PostWithJSONBody(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"name":"anjal"}` {
			t.Errorf("Expected body `{\"name\":\"anjal\"}`, got `%s`", string(body))
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":1}`))
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "POST",
		URL:    mockServer.URL,
		Body:   `{"name":"anjal"}`,
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
	if resp.Body != `{"id":1}` {
		t.Errorf("Expected body `{\"id\":1}`, got `%s`", resp.Body)
	}
}

func TestExecute_PostWithXMLBody(t *testing.T) {
	xmlBody := `<user><name>anjal</name></user>`

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != xmlBody {
			t.Errorf("Expected XML body, got `%s`", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "POST",
		URL:    mockServer.URL,
		Body:   xmlBody,
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_PostWithPlainTextBody(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "hello world" {
			t.Errorf("Expected body 'hello world', got `%s`", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "POST",
		URL:    mockServer.URL,
		Body:   "hello world",
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_PutRequest(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected method PUT, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"updated":true}` {
			t.Errorf("Expected body `{\"updated\":true}`, got `%s`", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "PUT",
		URL:    mockServer.URL,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"updated":true}`,
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_DeleteRequest(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected method DELETE, got %s", r.Method)
		}
		if r.URL.Query().Get("id") != "42" {
			t.Errorf("Expected query param id=42, got %s", r.URL.Query().Get("id"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "DELETE",
		URL:    mockServer.URL,
		QueryParams: map[string]string{
			"id": "42",
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_EmptyBody(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}
	if resp.Body != "" {
		t.Errorf("Expected empty body, got `%s`", resp.Body)
	}
}

func TestExecute_LargeResponseBody(t *testing.T) {
	largeBody := strings.Repeat("x", 1024*100) // 100 KB
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(largeBody))
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Body != largeBody {
		t.Errorf("Expected body length %d, got %d", len(largeBody), len(resp.Body))
	}
}

// ===========================================================================
// SECTION 3 — AUTHORIZATION: Bearer (JWT), Basic, API key, Custom, none
// ===========================================================================

func TestExecute_AuthBearerToken(t *testing.T) {
	const jwtToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U"

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedAuth := "Bearer " + jwtToken
		if r.Header.Get("Authorization") != expectedAuth {
			t.Errorf("Expected Authorization '%s', got '%s'", expectedAuth, r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"authenticated":true}`))
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
		Auth: &models.Auth{
			Type: "bearer",
			Params: map[string]string{
				"token": jwtToken,
			},
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if resp.Body != `{"authenticated":true}` {
		t.Errorf("Expected body `{\"authenticated\":true}`, got `%s`", resp.Body)
	}
}

func TestExecute_AuthBearerEmptyToken_NoHeaderInjected(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("Expected no Authorization header when token is empty, got '%s'",
				r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
		Auth: &models.Auth{
			Type: "bearer",
			Params: map[string]string{
				"token": "", // Empty token
			},
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_AuthBasic(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("Expected Basic auth credentials")
		}
		if username != "admin" || password != "secret" {
			t.Errorf("Expected admin:secret, got %s:%s", username, password)
		}

		expectedEncoded := base64.StdEncoding.EncodeToString([]byte("admin:secret"))
		expectedHeader := "Basic " + expectedEncoded
		if r.Header.Get("Authorization") != expectedHeader {
			t.Errorf("Expected Authorization '%s', got '%s'", expectedHeader, r.Header.Get("Authorization"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
		Auth: &models.Auth{
			Type: "basic",
			Params: map[string]string{
				"username": "admin",
				"password": "secret",
			},
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_AuthAPIKeyDefaultHeader(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "sk-abc123" {
			t.Errorf("Expected X-API-Key 'sk-abc123', got '%s'", r.Header.Get("X-API-Key"))
		}
		// Must NOT set Authorization header
		if r.Header.Get("Authorization") != "" {
			t.Errorf("Expected no Authorization header for API key auth, got '%s'",
				r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
		Auth: &models.Auth{
			Type: "apikey",
			Params: map[string]string{
				"key": "sk-abc123",
			},
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_AuthAPIKeyCustomHeader(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-My-Custom-Key") != "custom-789" {
			t.Errorf("Expected X-My-Custom-Key 'custom-789', got '%s'", r.Header.Get("X-My-Custom-Key"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
		Auth: &models.Auth{
			Type: "apikey",
			Params: map[string]string{
				"key":    "custom-789",
				"header": "X-My-Custom-Key",
			},
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_AuthCustomPrefix(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token my-secret-token" {
			t.Errorf("Expected Authorization 'Token my-secret-token', got '%s'",
				r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
		Auth: &models.Auth{
			Type: "custom",
			Params: map[string]string{
				"prefix": "Token",
				"token":  "my-secret-token",
			},
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_NoAuth_NilAuth(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Errorf("Expected no Authorization header when Auth is nil, got '%s'",
				r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
		Auth:   nil, // Explicitly no auth
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_AuthUnknownType_ReturnsError(t *testing.T) {
	req := models.APIRequest{
		Method: "GET",
		URL:    "http://example.com",
		Auth: &models.Auth{
			Type:   "oauth2", // Not supported
			Params: map[string]string{},
		},
	}

	_, err := Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Expected an error for unsupported auth type, but got nil")
	}
	if !strings.Contains(err.Error(), "unsupported auth type") {
		t.Errorf("Expected error to mention 'unsupported auth type', got '%v'", err)
	}
}

// ===========================================================================
// SECTION 4 — RESPONSE Content-Type classification
// ===========================================================================

func TestExecute_ContentType_JSON(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"key":"val"}`))
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "json" {
		t.Errorf("Expected ContentType 'json', got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_JSON_WithSuffix(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.api+json")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "json" {
		t.Errorf("Expected ContentType 'json' for +json suffix, got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_XML(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<root/>`))
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "xml" {
		t.Errorf("Expected ContentType 'xml', got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_XML_TextVariant(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "xml" {
		t.Errorf("Expected ContentType 'xml' for text/xml, got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_HTML(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<html><body>hello</body></html>`))
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "html" {
		t.Errorf("Expected ContentType 'html', got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_PlainText(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("just text"))
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "text" {
		t.Errorf("Expected ContentType 'text', got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_Missing_Raw(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write no body — if we Write() any bytes, Go auto-detects text/plain.
		// A true 204 leaves the Content-Type header unset.
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "raw" {
		t.Errorf("Expected ContentType 'raw' when missing, got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_Unknown_Raw(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "raw" {
		t.Errorf("Expected ContentType 'raw' for unknown type, got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_JavaScript(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("const x = 1;"))
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "javascript" {
		t.Errorf("Expected ContentType 'javascript', got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_JavaScript_TextVariant(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "javascript" {
		t.Errorf("Expected ContentType 'javascript' for text/javascript, got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_CSS(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("body { color: red; }"))
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "css" {
		t.Errorf("Expected ContentType 'css', got '%s'", resp.ContentType)
	}
}

func TestExecute_ContentType_FormURLEncoded(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("key1=value1&key2=value2"))
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.ContentType != "form" {
		t.Errorf("Expected ContentType 'form', got '%s'", resp.ContentType)
	}
}

// ===========================================================================
// SECTION 5 — Error paths & edge cases
// ===========================================================================

func TestExecute_InvalidURL(t *testing.T) {
	req := models.APIRequest{
		Method: "GET",
		URL:    "http://192.168.0.1\x7f/invalid",
	}

	_, err := Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Expected an error for a structurally invalid URL, but got nil")
	}
}

func TestExecute_ServerError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error (server errors are valid HTTP responses), got %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
	if resp.Body != `{"error":"internal"}` {
		t.Errorf("Expected body `{\"error\":\"internal\"}`, got `%s`", resp.Body)
	}
}

func TestExecute_LatencyRecorded(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{Method: "GET", URL: mockServer.URL}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.Latency < 50*time.Millisecond {
		t.Errorf("Expected latency >= 50ms, got %v", resp.Latency)
	}
}

func TestExecute_MultipleQueryParams(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("a") != "1" || q.Get("b") != "2" || q.Get("c") != "3" {
			t.Errorf("Expected a=1, b=2, c=3, got a=%s, b=%s, c=%s",
				q.Get("a"), q.Get("b"), q.Get("c"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
		QueryParams: map[string]string{
			"a": "1",
			"b": "2",
			"c": "3",
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_AppendToExistingQueryParams(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("existing") != "yes" {
			t.Errorf("Expected existing=yes, got %s", q.Get("existing"))
		}
		if q.Get("new") != "added" {
			t.Errorf("Expected new=added, got %s", q.Get("new"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL + "?existing=yes",
		QueryParams: map[string]string{
			"new": "added",
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestExecute_ConnectionRefused(t *testing.T) {
	req := models.APIRequest{
		Method: "GET",
		URL:    "http://127.0.0.1:1",
	}

	_, err := Execute(context.Background(), req)
	if err == nil {
		t.Fatal("Expected an error for a refused connection, but got nil")
	}
}

// ===========================================================================
// SECTION 6 — FillAuth: validation & construction
// ===========================================================================

func TestFillAuth_Bearer_Success(t *testing.T) {
	auth, err := FillAuth("bearer", map[string]string{"token": "jwt-abc"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if auth.Type != "bearer" {
		t.Errorf("Expected type 'bearer', got '%s'", auth.Type)
	}
	if auth.Params["token"] != "jwt-abc" {
		t.Errorf("Expected token 'jwt-abc', got '%s'", auth.Params["token"])
	}
}

func TestFillAuth_Bearer_MissingToken(t *testing.T) {
	_, err := FillAuth("bearer", map[string]string{})
	if err == nil {
		t.Fatal("Expected error for missing token, got nil")
	}
}

func TestFillAuth_Bearer_EmptyToken(t *testing.T) {
	_, err := FillAuth("bearer", map[string]string{"token": ""})
	if err == nil {
		t.Fatal("Expected error for empty token, got nil")
	}
}

func TestFillAuth_Basic_Success(t *testing.T) {
	auth, err := FillAuth("basic", map[string]string{
		"username": "admin",
		"password": "secret",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if auth.Type != "basic" {
		t.Errorf("Expected type 'basic', got '%s'", auth.Type)
	}
	if auth.Params["username"] != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", auth.Params["username"])
	}
	if auth.Params["password"] != "secret" {
		t.Errorf("Expected password 'secret', got '%s'", auth.Params["password"])
	}
}

func TestFillAuth_Basic_MissingUsername(t *testing.T) {
	_, err := FillAuth("basic", map[string]string{"password": "secret"})
	if err == nil {
		t.Fatal("Expected error for missing username, got nil")
	}
}

func TestFillAuth_Basic_MissingPassword(t *testing.T) {
	_, err := FillAuth("basic", map[string]string{"username": "admin"})
	if err == nil {
		t.Fatal("Expected error for missing password, got nil")
	}
}

func TestFillAuth_APIKey_Success(t *testing.T) {
	auth, err := FillAuth("apikey", map[string]string{"key": "sk-abc"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if auth.Type != "apikey" {
		t.Errorf("Expected type 'apikey', got '%s'", auth.Type)
	}
	if auth.Params["header"] != "X-API-Key" {
		t.Errorf("Expected default header 'X-API-Key', got '%s'", auth.Params["header"])
	}
}

func TestFillAuth_APIKey_CustomHeader(t *testing.T) {
	auth, err := FillAuth("apikey", map[string]string{
		"key":    "sk-abc",
		"header": "X-Custom",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if auth.Params["header"] != "X-Custom" {
		t.Errorf("Expected header 'X-Custom', got '%s'", auth.Params["header"])
	}
}

func TestFillAuth_APIKey_MissingKey(t *testing.T) {
	_, err := FillAuth("apikey", map[string]string{})
	if err == nil {
		t.Fatal("Expected error for missing key, got nil")
	}
}

func TestFillAuth_Custom_Success(t *testing.T) {
	auth, err := FillAuth("custom", map[string]string{
		"prefix": "Token",
		"token":  "my-token",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if auth.Type != "custom" {
		t.Errorf("Expected type 'custom', got '%s'", auth.Type)
	}
	if auth.Params["prefix"] != "Token" {
		t.Errorf("Expected prefix 'Token', got '%s'", auth.Params["prefix"])
	}
}

func TestFillAuth_Custom_MissingPrefix(t *testing.T) {
	_, err := FillAuth("custom", map[string]string{"token": "my-token"})
	if err == nil {
		t.Fatal("Expected error for missing prefix, got nil")
	}
}

func TestFillAuth_UnknownType(t *testing.T) {
	_, err := FillAuth("oauth2", map[string]string{})
	if err == nil {
		t.Fatal("Expected error for unknown auth type, got nil")
	}
}

func TestFillAuth_CaseInsensitive(t *testing.T) {
	auth, err := FillAuth("BEARER", map[string]string{"token": "jwt"})
	if err != nil {
		t.Fatalf("Expected no error for uppercase type, got %v", err)
	}
	if auth.Type != "bearer" {
		t.Errorf("Expected type normalized to 'bearer', got '%s'", auth.Type)
	}
}

// ===========================================================================
// SECTION 7 — Cookie auth (explicit @auth cookie)
// ===========================================================================

func TestExecute_AuthCookie(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			t.Errorf("Expected cookie 'session_id', got error: %v", err)
		}
		if cookie.Value != "abc123" {
			t.Errorf("Expected cookie value 'abc123', got '%s'", cookie.Value)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	req := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL,
		Auth: &models.Auth{
			Type: "cookie",
			Params: map[string]string{
				"name":  "session_id",
				"value": "abc123",
			},
		},
	}

	resp, err := Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// ===========================================================================
// SECTION 8 — Cookie jar: automatic session persistence
// ===========================================================================

func TestExecute_CookieJar_Persistence(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			// Step 1: server sets a session cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "session",
				Value: "logged-in",
			})
			w.WriteHeader(http.StatusOK)
		case "/profile":
			// Step 2: server checks the cookie came back automatically
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value != "logged-in" {
				t.Errorf("Expected automatic cookie 'session=logged-in', got err=%v", err)
			}
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer mockServer.Close()

	// Step 1 — login, server sets Set-Cookie, jar saves it
	loginReq := models.APIRequest{
		Method: "POST",
		URL:    mockServer.URL + "/login",
	}
	resp1, err := Execute(context.Background(), loginReq)
	if err != nil {
		t.Fatalf("Login request failed: %v", err)
	}
	if resp1.StatusCode != http.StatusOK {
		t.Errorf("Login: expected 200, got %d", resp1.StatusCode)
	}

	// Step 2 — profile, jar automatically sends the cookie back
	profileReq := models.APIRequest{
		Method: "GET",
		URL:    mockServer.URL + "/profile",
	}
	resp2, err := Execute(context.Background(), profileReq)
	if err != nil {
		t.Fatalf("Profile request failed: %v", err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("Profile: expected 200, got %d", resp2.StatusCode)
	}
}

// ===========================================================================
// SECTION 9 — FillAuth cookie validation
// ===========================================================================

func TestFillAuth_Cookie_Success(t *testing.T) {
	auth, err := FillAuth("cookie", map[string]string{
		"name":  "session_id",
		"value": "abc123",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if auth.Type != "cookie" {
		t.Errorf("Expected type 'cookie', got '%s'", auth.Type)
	}
	if auth.Params["name"] != "session_id" {
		t.Errorf("Expected name 'session_id', got '%s'", auth.Params["name"])
	}
	if auth.Params["value"] != "abc123" {
		t.Errorf("Expected value 'abc123', got '%s'", auth.Params["value"])
	}
}

func TestFillAuth_Cookie_MissingName(t *testing.T) {
	_, err := FillAuth("cookie", map[string]string{"value": "abc"})
	if err == nil {
		t.Fatal("Expected error for missing cookie name, got nil")
	}
}

func TestFillAuth_Cookie_MissingValue(t *testing.T) {
	_, err := FillAuth("cookie", map[string]string{"name": "session_id"})
	if err == nil {
		t.Fatal("Expected error for missing cookie value, got nil")
	}
}
