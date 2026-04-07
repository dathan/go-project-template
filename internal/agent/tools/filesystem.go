package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileReadTool reads file contents or lists a directory.
type FileReadTool struct {
	Root string // sandbox root; empty = no restriction
}

func NewFileReadTool(root string) *FileReadTool { return &FileReadTool{Root: root} }

func (t *FileReadTool) Name() string        { return "file_read" }
func (t *FileReadTool) Description() string { return "Read a file or list directory contents." }
func (t *FileReadTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{"type": "string", "description": "Absolute or relative file/directory path."},
		},
		"required": []string{"path"},
	}
}

func (t *FileReadTool) Execute(_ context.Context, input map[string]any) (string, error) {
	path, err := t.safePath(str(input, "path"))
	if err != nil {
		return "", err
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("stat %s: %w", path, err)
	}

	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", err
		}
		var sb strings.Builder
		for _, e := range entries {
			sb.WriteString(e.Name())
			if e.IsDir() {
				sb.WriteString("/")
			}
			sb.WriteString("\n")
		}
		return sb.String(), nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	return string(b), nil
}

// FileWriteTool writes content to a file, creating parent directories as needed.
type FileWriteTool struct {
	Root string
}

func NewFileWriteTool(root string) *FileWriteTool { return &FileWriteTool{Root: root} }

func (t *FileWriteTool) Name() string        { return "file_write" }
func (t *FileWriteTool) Description() string { return "Write content to a file." }
func (t *FileWriteTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path":    map[string]any{"type": "string"},
			"content": map[string]any{"type": "string"},
		},
		"required": []string{"path", "content"},
	}
}

func (t *FileWriteTool) Execute(_ context.Context, input map[string]any) (string, error) {
	path, err := t.safePath(str(input, "path"))
	if err != nil {
		return "", err
	}
	content := str(input, "content")

	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return "", fmt.Errorf("mkdir: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return "", fmt.Errorf("writing %s: %w", path, err)
	}
	return fmt.Sprintf("wrote %d bytes to %s", len(content), path), nil
}

// safePath resolves a path and ensures it's within Root (if set).
func (t *FileReadTool) safePath(p string) (string, error) {
	return safePath(t.Root, p)
}

func (t *FileWriteTool) safePath(p string) (string, error) {
	return safePath(t.Root, p)
}

func safePath(root, p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("path is required")
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if root != "" {
		rootAbs, err := filepath.Abs(root)
		if err != nil {
			return "", err
		}
		if !strings.HasPrefix(abs, rootAbs) {
			return "", fmt.Errorf("path %q escapes sandbox root %q", abs, rootAbs)
		}
	}
	return abs, nil
}
