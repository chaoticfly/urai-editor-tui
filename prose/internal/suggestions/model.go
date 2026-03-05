package suggestions

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"urai/internal/ai"
	"urai/internal/config"
)

// Mode controls whether suggestions are based on the current line or paragraph.
type Mode int

const (
	ModeLine Mode = iota
	ModeParagraph
)

// SuggestionMsg carries an AI response or error back to the app.
type SuggestionMsg struct {
	Text  string
	Error error
}

// InsertMsg asks the app to insert text at the editor cursor.
type InsertMsg struct {
	Text string
}

// Model is the AI suggestions panel state.
type Model struct {
	aiConfig   config.AIConfig
	mode       Mode
	context    string
	suggestion string
	loading    bool
	errMsg     string
	scroll     int
	width      int
	height     int
}

// New creates a new suggestions model.
func New(cfg config.AIConfig, width, height int) *Model {
	return &Model{
		aiConfig: cfg,
		mode:     ModeLine,
		width:    width,
		height:   height,
	}
}

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// UpdateConfig replaces the AI config (called when settings are saved).
func (m *Model) UpdateConfig(cfg config.AIConfig) {
	m.aiConfig = cfg
}

// SetEditorContext extracts and stores the context (line or paragraph) based
// on the current cursor position in the editor content.
func (m *Model) SetEditorContext(content string, cursorLine int) {
	lines := strings.Split(content, "\n")
	if m.mode == ModeLine {
		if cursorLine >= 0 && cursorLine < len(lines) {
			m.context = lines[cursorLine]
		}
	} else {
		m.context = extractParagraph(lines, cursorLine)
	}
}

func extractParagraph(lines []string, cursorLine int) string {
	if len(lines) == 0 {
		return ""
	}
	if cursorLine >= len(lines) {
		cursorLine = len(lines) - 1
	}

	start := cursorLine
	for start > 0 && strings.TrimSpace(lines[start-1]) != "" {
		start--
	}
	end := cursorLine
	for end < len(lines)-1 && strings.TrimSpace(lines[end+1]) != "" {
		end++
	}

	return strings.Join(lines[start:end+1], "\n")
}

// ToggleMode switches between line and paragraph mode.
func (m *Model) ToggleMode() {
	if m.mode == ModeLine {
		m.mode = ModeParagraph
	} else {
		m.mode = ModeLine
	}
	m.suggestion = ""
	m.errMsg = ""
}

// RequestSuggestion triggers an async AI request. Returns nil if already loading.
func (m *Model) RequestSuggestion() tea.Cmd {
	if m.loading {
		return nil
	}
	if m.aiConfig.BaseURL == "" && m.aiConfig.Provider != "openai" {
		m.errMsg = "No API URL configured. Open Settings (F2) to configure."
		return nil
	}
	if m.context == "" {
		m.errMsg = "No context — move cursor to a line with text."
		return nil
	}

	m.loading = true
	m.errMsg = ""

	ctx := m.context
	cfg := m.aiConfig
	return func() tea.Msg {
		client := ai.New(cfg.BaseURL, cfg.APIKey, cfg.Model)
		result, err := client.Complete(cfg.SystemPrompt, ctx)
		return SuggestionMsg{Text: result, Error: err}
	}
}

// HandleSuggestion processes the AI response.
func (m *Model) HandleSuggestion(msg SuggestionMsg) {
	m.loading = false
	if msg.Error != nil {
		m.errMsg = msg.Error.Error()
		m.suggestion = ""
	} else {
		m.suggestion = msg.Text
		m.errMsg = ""
		m.scroll = 0
	}
}

// InsertSuggestion returns a command to insert the current suggestion.
func (m *Model) InsertSuggestion() tea.Cmd {
	if m.suggestion == "" {
		return nil
	}
	text := m.suggestion
	return func() tea.Msg {
		return InsertMsg{Text: text}
	}
}

// ScrollUp scrolls the suggestion text up.
func (m *Model) ScrollUp() {
	if m.scroll > 0 {
		m.scroll--
	}
}

// ScrollDown scrolls the suggestion text down.
func (m *Model) ScrollDown() {
	lines := strings.Split(m.suggestion, "\n")
	max := len(lines) - 3
	if max < 0 {
		max = 0
	}
	if m.scroll < max {
		m.scroll++
	}
}

// View renders the suggestions panel.
func (m *Model) View() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")).
		Bold(true)

	activeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Padding(0, 1)

	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)

	divider := strings.Repeat("─", m.width-2)
	if m.width <= 2 {
		divider = ""
	}

	var modeStr string
	if m.mode == ModeLine {
		modeStr = activeStyle.Render("Line") + " " + inactiveStyle.Render("Paragraph")
	} else {
		modeStr = inactiveStyle.Render("Line") + " " + activeStyle.Render("Paragraph")
	}

	var lines []string
	lines = append(lines, headerStyle.Render("AI Suggestions"))
	lines = append(lines, modeStr)
	lines = append(lines, divider)

	// Context section
	lines = append(lines, sectionStyle.Render("Context"))
	if m.context == "" {
		lines = append(lines, dimStyle.Render("(no context)"))
	} else {
		ctxLines := strings.Split(m.context, "\n")
		maxCtx := 4
		for i, l := range ctxLines {
			if i >= maxCtx {
				lines = append(lines, dimStyle.Render("..."))
				break
			}
			lines = append(lines, truncate(l, m.width-2))
		}
	}
	lines = append(lines, divider)

	// Suggestion section
	lines = append(lines, sectionStyle.Render("Suggestion"))

	switch {
	case m.loading:
		lines = append(lines, dimStyle.Render("Generating..."))
	case m.errMsg != "":
		// Wrap error message
		for _, l := range wrapText(m.errMsg, m.width-2) {
			lines = append(lines, errStyle.Render(l))
		}
	case m.suggestion == "":
		lines = append(lines, dimStyle.Render("Press Ctrl+R to generate"))
	default:
		suggLines := strings.Split(m.suggestion, "\n")
		// Apply scroll
		if m.scroll > 0 && m.scroll < len(suggLines) {
			suggLines = suggLines[m.scroll:]
		}
		availHeight := m.height - len(lines) - 3
		if availHeight < 1 {
			availHeight = 1
		}
		for i, l := range suggLines {
			if i >= availHeight {
				lines = append(lines, dimStyle.Render("... (scroll down)"))
				break
			}
			lines = append(lines, truncate(l, m.width-2))
		}
	}

	// Footer hints
	for len(lines) < m.height-1 {
		lines = append(lines, "")
	}
	if len(lines) >= m.height {
		lines = lines[:m.height-1]
	}
	lines = append(lines, hintStyle.Render("Ctrl+R: suggest  Ctrl+L: insert  F4: mode"))

	return strings.Join(lines, "\n")
}

func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) > maxWidth {
		return string(runes[:maxWidth-1]) + "…"
	}
	return s
}

func wrapText(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}
	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) <= width {
			line += " " + w
		} else {
			lines = append(lines, line)
			line = w
		}
	}
	lines = append(lines, line)
	return lines
}
