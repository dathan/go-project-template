// Package tools defines the Tool interface and all built-in OS tool implementations.
package tools

import "context"

// Tool is the interface every agent tool must implement.
type Tool interface {
	Name() string
	Description() string
	// Parameters returns a JSON Schema object describing the tool's input.
	Parameters() map[string]any
	Execute(ctx context.Context, input map[string]any) (string, error)
}

func str(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return v
}
