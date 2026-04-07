package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// ShellTool executes shell commands with a configurable timeout.
// The allowlist restricts which commands may be called; an empty allowlist
// permits all commands (useful for trusted/admin-only agents).
type ShellTool struct {
	Timeout   time.Duration
	Allowlist []string // e.g. ["ls","cat","grep"]; empty = allow all
}

func NewShellTool(timeout time.Duration, allowlist ...string) *ShellTool {
	return &ShellTool{Timeout: timeout, Allowlist: allowlist}
}

func (t *ShellTool) Name() string        { return "shell" }
func (t *ShellTool) Description() string { return "Execute a shell command and return stdout+stderr." }
func (t *ShellTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{"type": "string", "description": "The shell command to run."},
		},
		"required": []string{"command"},
	}
}

func (t *ShellTool) Execute(ctx context.Context, input map[string]any) (string, error) {
	command := str(input, "command")
	if command == "" {
		return "", fmt.Errorf("command is required")
	}

	parts := strings.Fields(command)
	if len(t.Allowlist) > 0 && !contains(t.Allowlist, parts[0]) {
		return "", fmt.Errorf("command %q is not in the allowlist", parts[0])
	}

	timeout := t.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	//nolint:gosec // command is allowlist-gated above
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	if err := cmd.Run(); err != nil {
		return buf.String(), fmt.Errorf("command failed: %w\noutput: %s", err, buf.String())
	}
	return buf.String(), nil
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
