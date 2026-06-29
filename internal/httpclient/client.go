package httpclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/yogasimman/anjal/internal/models"
)

func init() {
	jar, _ := cookiejar.New(nil)
	pooledClient.Jar = jar
}

var pooledClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	},
}

func Execute(ctx context.Context, request models.APIRequest) (models.APIResponse, error) {

	u, err := url.Parse(request.URL)
	if err != nil {
		return models.APIResponse{}, err
	}

	if len(request.QueryParams) > 0 {
		q := u.Query()
		for key, value := range request.QueryParams {
			q.Set(key, value)
		}
		u.RawQuery = q.Encode()
	}

	var bodyReader io.Reader
	if request.Body != "" {
		bodyReader = bytes.NewBuffer([]byte(request.Body))
	}

	req, err := http.NewRequestWithContext(ctx, request.Method, u.String(), bodyReader)
	if err != nil {
		return models.APIResponse{}, err
	}

	if request.Headers != nil {
		for key, value := range request.Headers {
			req.Header.Set(key, value)
		}
	}

	if err := applyAuth(req, request.Auth); err != nil {
		return models.APIResponse{}, err
	}

	start := time.Now()

	resp, err := pooledClient.Do(req)
	latency := time.Since(start)

	if err != nil {
		return models.APIResponse{}, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return models.APIResponse{}, err
	}

	contentType := detectContentType(resp.Header.Get("Content-Type"))

	return models.APIResponse{
		StatusCode:  resp.StatusCode,
		Status:      resp.Status,
		Body:        string(respBytes),
		Headers:     resp.Header,
		Latency:     latency,
		ContentType: contentType,
	}, nil
}

// applyAuth injects the appropriate Authorization (or custom) header onto the
// request based on the Auth config. A nil auth is a no-op. Returns an error
// only when the auth type is not recognised.
func applyAuth(req *http.Request, auth *models.Auth) error {
	if auth == nil {
		return nil
	}

	switch strings.ToLower(auth.Type) {
	case "bearer":
		token, ok := auth.Params["token"]
		if ok && token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "basic":
		username := auth.Params["username"]
		password := auth.Params["password"]
		encoded := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		req.Header.Set("Authorization", "Basic "+encoded)
	case "apikey":
		key, ok := auth.Params["key"]
		headerName := auth.Params["header"]
		if headerName == "" {
			headerName = "X-API-Key"
		}
		if ok && key != "" {
			req.Header.Set(headerName, key)
		}
	case "custom":
		prefix, _ := auth.Params["prefix"]
		token, _ := auth.Params["token"]
		if token != "" {
			req.Header.Set("Authorization", prefix+" "+token)
		} else {
			req.Header.Set("Authorization", prefix)
		}
	case "cookie":
		req.AddCookie(&http.Cookie{
			Name:  auth.Params["name"],
			Value: auth.Params["value"],
		})
	default:
		return fmt.Errorf("unsupported auth type: %s", auth.Type)
	}

	return nil
}

// FillAuth validates the supplied auth type and params, then returns a
// populated *models.Auth ready to drop into APIRequest.Auth. The frontend
// calls this to build auth config before handing it to Execute.
func FillAuth(authType string, params map[string]string) (*models.Auth, error) {
	auth := &models.Auth{
		Type:   strings.ToLower(authType),
		Params: params,
	}

	switch auth.Type {
	case "bearer":
		token, ok := auth.Params["token"]
		if !ok || token == "" {
			return nil, fmt.Errorf("bearer auth requires a token")
		}
	case "basic":
		if auth.Params["username"] == "" {
			return nil, fmt.Errorf("basic auth requires a username")
		}
		if auth.Params["password"] == "" {
			return nil, fmt.Errorf("basic auth requires a password")
		}
	case "apikey":
		key, ok := auth.Params["key"]
		if !ok || key == "" {
			return nil, fmt.Errorf("apikey auth requires a key")
		}
		if auth.Params["header"] == "" {
			auth.Params["header"] = "X-API-Key"
		}
	case "custom":
		if auth.Params["prefix"] == "" {
			return nil, fmt.Errorf("custom auth requires a prefix")
		}
	case "cookie":
		if auth.Params["name"] == "" {
			return nil, fmt.Errorf("cookie auth requires a name")
		}
		if auth.Params["value"] == "" {
			return nil, fmt.Errorf("cookie auth requires a value")
		}
	default:
		return nil, fmt.Errorf("unsupported auth type: %s", authType)
	}

	return auth, nil
}

// detectContentType classifies a MIME Content-Type header into a display category.
func detectContentType(contentType string) string {
	if contentType == "" {
		return "raw"
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "raw"
	}

	switch {
	case strings.HasPrefix(mediaType, "application/json") ||
		strings.HasSuffix(mediaType, "+json"):
		return "json"
	case strings.HasPrefix(mediaType, "application/xml") ||
		strings.HasPrefix(mediaType, "text/xml") ||
		strings.HasSuffix(mediaType, "+xml"):
		return "xml"
	case strings.HasPrefix(mediaType, "text/html"):
		return "html"
	case strings.HasPrefix(mediaType, "application/javascript") ||
		strings.HasPrefix(mediaType, "text/javascript"):
		return "javascript"
	case strings.HasPrefix(mediaType, "text/css"):
		return "css"
	case strings.HasPrefix(mediaType, "application/x-www-form-urlencoded"):
		return "form"
	case strings.HasPrefix(mediaType, "text/plain"):
		return "text"
	default:
		return "raw"
	}
}
