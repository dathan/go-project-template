package tools

import (
	"context"
	"fmt"

	"github.com/playwright-community/playwright-go"
)

// PlaywrightTool drives a headless browser via playwright-go.
// Run `make playwright-install` (or `go run github.com/playwright-community/playwright-go/cmd/playwright install`)
// before using this tool.
type PlaywrightTool struct{}

func NewPlaywrightTool() *PlaywrightTool { return &PlaywrightTool{} }

func (t *PlaywrightTool) Name() string { return "browser" }
func (t *PlaywrightTool) Description() string {
	return "Open a URL in a headless browser and return the page text content."
}
func (t *PlaywrightTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"url":    map[string]any{"type": "string", "description": "URL to navigate to."},
			"action": map[string]any{"type": "string", "enum": []string{"get_text", "screenshot"}, "default": "get_text"},
		},
		"required": []string{"url"},
	}
}

func (t *PlaywrightTool) Execute(_ context.Context, input map[string]any) (string, error) {
	url := str(input, "url")
	if url == "" {
		return "", fmt.Errorf("url is required")
	}
	action := str(input, "action")
	if action == "" {
		action = "get_text"
	}

	pw, err := playwright.Run()
	if err != nil {
		return "", fmt.Errorf("starting playwright: %w", err)
	}
	defer pw.Stop() //nolint:errcheck

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("launching browser: %w", err)
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return "", fmt.Errorf("new page: %w", err)
	}

	if _, err := page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return "", fmt.Errorf("navigating to %s: %w", url, err)
	}

	switch action {
	case "screenshot":
		path := "/tmp/screenshot.png"
		if _, err := page.Screenshot(playwright.PageScreenshotOptions{Path: playwright.String(path)}); err != nil {
			return "", fmt.Errorf("screenshot: %w", err)
		}
		return fmt.Sprintf("screenshot saved to %s", path), nil

	default: // get_text
		text, err := page.Locator("body").InnerText()
		if err != nil {
			return "", fmt.Errorf("getting text: %w", err)
		}
		return text, nil
	}
}
