package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dathan/go-project-template/pkg"
)

// AgentHandler proxies prompts to the configured LLM agent.
type AgentHandler struct {
	agent *pkg.Agent
}

func NewAgentHandler(agent *pkg.Agent) *AgentHandler {
	return &AgentHandler{agent: agent}
}

// Prompt handles a single-turn prompt request.
// POST /api/v1/agent/prompt
func (h *AgentHandler) Prompt(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Prompt string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	response := h.agent.Prompt(req.Prompt)
	writeJSON(w, http.StatusOK, map[string]string{"response": response})
}

// Stream streams the agent response as Server-Sent Events.
// GET /api/v1/agent/stream?prompt=...
func (h *AgentHandler) Stream(w http.ResponseWriter, r *http.Request) {
	prompt := r.URL.Query().Get("prompt")
	if prompt == "" {
		writeError(w, http.StatusBadRequest, "prompt query parameter is required")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	// Run prompt in blocking mode and send as a single SSE event.
	// TODO: replace with true streaming when langchaingo supports it.
	response := h.agent.Prompt(prompt)
	fmt.Fprintf(w, "data: %s\n\n", jsonEscape(response))
	flusher.Flush()

	fmt.Fprintf(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
}

func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
