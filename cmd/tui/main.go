// TUI is a standalone Bubble Tea client that talks to the backend API.
// It supports OAuth login (opens browser), viewing the dashboard, and chatting
// with the agent.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")).MarginBottom(1)
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	normalStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	userBubble   = lipgloss.NewStyle().Background(lipgloss.Color("99")).Foreground(lipgloss.Color("255")).Padding(0, 1)
	botBubble    = lipgloss.NewStyle().Background(lipgloss.Color("237")).Foreground(lipgloss.Color("252")).Padding(0, 1)
)

// ── View enum ─────────────────────────────────────────────────────────────────

type view int

const (
	viewMenu view = iota
	viewDashboard
	viewAgentChat
)

// ── Messages ──────────────────────────────────────────────────────────────────

type meLoaded struct{ name, email, role string }
type agentResponse struct{ text string }
type errMsg struct{ err error }

// ── Model ─────────────────────────────────────────────────────────────────────

type chatMsg struct {
	role    string // "you" | "agent"
	content string
}

type model struct {
	apiBase  string
	token    string
	view     view
	cursor   int
	input    string
	status   string
	userName string
	userRole string
	messages []chatMsg
}

const menuItemAgentChat = "Agent Chat"

var menuItems = []string{"Dashboard", menuItemAgentChat, "Quit"}

func initialModel(apiBase, token string) model {
	return model{apiBase: apiBase, token: token, view: viewMenu}
}

// ── Init ──────────────────────────────────────────────────────────────────────

func (m model) Init() tea.Cmd { return nil }

// ── Update ────────────────────────────────────────────────────────────────────

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		return m.handleKey(msg)

	case meLoaded:
		m.userName = msg.name
		m.userRole = msg.role
		m.status = ""
		return m, nil

	case agentResponse:
		m.messages = append(m.messages, chatMsg{role: "agent", content: msg.text})
		m.status = ""
		return m, nil

	case errMsg:
		m.status = errorStyle.Render("Error: " + msg.err.Error())
		return m, nil
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.view {
	case viewMenu:
		return m.handleMenuKey(msg)
	case viewDashboard:
		if msg.String() == "q" || msg.String() == "esc" {
			m.view = viewMenu
		}
	case viewAgentChat:
		return m.handleChatKey(msg)
	}
	return m, nil
}

func (m model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(menuItems)-1 {
			m.cursor++
		}
	case "enter", " ":
		switch menuItems[m.cursor] {
		case "Dashboard":
			m.view = viewDashboard
			m.status = "Loading…"
			return m, m.fetchMe()
		case menuItemAgentChat:
			m.view = viewAgentChat
		case "Quit":
			return m, tea.Quit
		}
	case "ctrl+c", "q":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) handleChatKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		m.view = viewMenu
	case "enter":
		if m.input == "" {
			break
		}
		prompt := m.input
		m.input = ""
		m.messages = append(m.messages, chatMsg{role: "you", content: prompt})
		m.status = "Thinking…"
		return m, m.sendPrompt(prompt)
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	default:
		m.input += msg.String()
	}
	return m, nil
}

// ── View ──────────────────────────────────────────────────────────────────────

func (m model) View() string {
	switch m.view {
	case viewDashboard:
		return m.viewDashboard()
	case viewAgentChat:
		return m.viewChat()
	default:
		return m.viewMenu()
	}
}

func (m model) viewMenu() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("AppTemplate TUI"))
	b.WriteString("\n")
	if m.token == "" {
		b.WriteString(errorStyle.Render("No token — run with --token <jwt>") + "\n\n")
	}
	for i, item := range menuItems {
		if i == m.cursor {
			b.WriteString(activeStyle.Render("▶ "+item) + "\n")
		} else {
			b.WriteString(normalStyle.Render("  "+item) + "\n")
		}
	}
	b.WriteString("\n" + dimStyle.Render("↑/↓ navigate · enter select · q quit"))
	return b.String()
}

func (m model) viewDashboard() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Dashboard"))
	if m.status != "" {
		b.WriteString(m.status + "\n")
	} else {
		b.WriteString(successStyle.Render("Name: "+m.userName) + "\n")
		b.WriteString(normalStyle.Render("Role: "+m.userRole) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("esc/q → menu"))
	return b.String()
}

func (m model) viewChat() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(menuItemAgentChat))
	b.WriteString("\n")
	for _, msg := range m.messages {
		if msg.role == "you" {
			b.WriteString(userBubble.Render("you") + " " + msg.content + "\n")
		} else {
			b.WriteString(botBubble.Render("agent") + " " + msg.content + "\n")
		}
	}
	if m.status != "" {
		b.WriteString(dimStyle.Render(m.status) + "\n")
	}
	b.WriteString("\n> " + m.input + "█")
	b.WriteString("\n" + dimStyle.Render("enter send · esc menu"))
	return b.String()
}

// ── Commands ──────────────────────────────────────────────────────────────────

func (m model) fetchMe() tea.Cmd {
	return func() tea.Msg {
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, m.apiBase+"/api/v1/me", nil)
		req.Header.Set("Authorization", "Bearer "+m.token)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()
		var data struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return errMsg{err}
		}
		name := data.Name
		if name == "" {
			name = data.Email
		}
		return meLoaded{name: name, email: data.Email, role: data.Role}
	}
}

func (m model) sendPrompt(prompt string) tea.Cmd {
	return func() tea.Msg {
		body := strings.NewReader(`{"prompt":` + jsonStr(prompt) + `}`)
		req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, m.apiBase+"/api/v1/agent/prompt", body)
		req.Header.Set("Authorization", "Bearer "+m.token)
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errMsg{err}
		}
		defer resp.Body.Close()
		var data struct {
			Response string `json:"response"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return errMsg{err}
		}
		return agentResponse{text: data.Response}
	}
}

func jsonStr(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	}
	_ = cmd.Start()
}

// ── Main ──────────────────────────────────────────────────────────────────────

func main() {
	apiBase := flag.String("api", "http://127.0.0.1:8080", "Backend API base URL")
	token := flag.String("token", os.Getenv("APP_TOKEN"), "JWT bearer token (or set APP_TOKEN env var)")
	flag.Parse()

	if *token == "" {
		fmt.Println("No token provided. Opening browser for login…")
		openBrowser(*apiBase + "/auth/google")
		fmt.Println("After login, copy the token from the URL fragment and re-run with --token <jwt>")
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(*apiBase, *token))
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
