package tools

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HTTPTool makes HTTP GET or POST requests.
type HTTPTool struct {
	client *http.Client
}

func NewHTTPTool() *HTTPTool {
	return &HTTPTool{client: &http.Client{Timeout: 30 * time.Second}}
}

func (t *HTTPTool) Name() string        { return "http_fetch" }
func (t *HTTPTool) Description() string { return "Make an HTTP GET or POST request and return the response body." }
func (t *HTTPTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url":    map[string]any{"type": "string"},
			"method": map[string]any{"type": "string", "enum": []string{"GET", "POST"}, "default": "GET"},
			"body":   map[string]any{"type": "string", "description": "Request body for POST."},
		},
		"required": []string{"url"},
	}
}

func (t *HTTPTool) Execute(ctx context.Context, input map[string]any) (string, error) {
	url := str(input, "url")
	if url == "" {
		return "", fmt.Errorf("url is required")
	}
	method := strings.ToUpper(str(input, "method"))
	if method == "" {
		method = "GET"
	}
	body := str(input, "body")

	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MB cap
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}
	return fmt.Sprintf("HTTP %d\n%s", resp.StatusCode, string(respBody)), nil
}
