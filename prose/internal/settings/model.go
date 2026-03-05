package settings

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"urai/internal/config"
)

// SaveMsg is emitted when the user saves settings.
type SaveMsg struct {
	Config config.Config
}

// CloseMsg is emitted when the settings panel is closed.
type CloseMsg struct{}

const (
	fieldProvider = iota
	fieldBaseURL
	fieldAPIKey
	fieldModel
	fieldSystemPrompt
	fieldBgColor
	numFields
)

var fieldLabels = [numFields]string{
	"Provider",
	"Base URL",
	"API Key",
	"Model",
	"System Prompt",
	"Background Color",
}

// Model is the settings panel state.
type Model struct {
	appConfig config.Config
	focused   int
	values    [numFields][]rune
	cursors   [numFields]int
	width     int
	height    int
}

// New creates a new settings model from the current config.
func New(cfg config.Config, width, height int) *Model {
	m := &Model{
		appConfig: cfg,
		width:     width,
		height:    height,
	}

	m.values[fieldProvider] = []rune(cfg.AI.Provider)
	if string(m.values[fieldProvider]) == "" {
		m.values[fieldProvider] = []rune("ollama")
	}
	m.values[fieldBaseURL] = []rune(cfg.AI.BaseURL)
	m.values[fieldAPIKey] = []rune(cfg.AI.APIKey)
	m.values[fieldModel] = []rune(cfg.AI.Model)
	m.values[fieldSystemPrompt] = []rune(cfg.AI.SystemPrompt)
	m.values[fieldBgColor] = []rune(cfg.BackgroundColor)

	for i := range m.cursors {
		m.cursors[i] = len(m.values[i])
	}

	return m
}

// SetSize updates the panel dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Update handles keyboard input.
func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	return m.handleKey(keyMsg)
}

func (m *Model) handleKey(msg tea.KeyMsg) (*Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		return m, func() tea.Msg { return CloseMsg{} }

	case "ctrl+s":
		return m, m.buildSaveCmd()

	case "tab":
		m.focused = (m.focused + 1) % numFields
		return m, nil

	case "shift+tab":
		m.focused = (m.focused - 1 + numFields) % numFields
		return m, nil
	}

	// Provider field: toggle with space or enter
	if m.focused == fieldProvider {
		if key == " " || key == "enter" {
			if string(m.values[fieldProvider]) == "ollama" {
				m.values[fieldProvider] = []rune("openai")
			} else {
				m.values[fieldProvider] = []rune("ollama")
			}
		}
		return m, nil
	}

	// Text field editing
	switch key {
	case "left":
		if m.cursors[m.focused] > 0 {
			m.cursors[m.focused]--
		}
	case "right":
		if m.cursors[m.focused] < len(m.values[m.focused]) {
			m.cursors[m.focused]++
		}
	case "home", "ctrl+a":
		if m.focused == fieldSystemPrompt {
			// Move to start of current line
			v := m.values[m.focused]
			c := m.cursors[m.focused]
			for c > 0 && v[c-1] != '\n' {
				c--
			}
			m.cursors[m.focused] = c
		} else {
			m.cursors[m.focused] = 0
		}
	case "end", "ctrl+e":
		if m.focused == fieldSystemPrompt {
			v := m.values[m.focused]
			c := m.cursors[m.focused]
			for c < len(v) && v[c] != '\n' {
				c++
			}
			m.cursors[m.focused] = c
		} else {
			m.cursors[m.focused] = len(m.values[m.focused])
		}
	case "backspace":
		if m.cursors[m.focused] > 0 {
			v := m.values[m.focused]
			c := m.cursors[m.focused]
			m.values[m.focused] = append(v[:c-1:c-1], v[c:]...)
			m.cursors[m.focused]--
		}
	case "delete":
		v := m.values[m.focused]
		c := m.cursors[m.focused]
		if c < len(v) {
			m.values[m.focused] = append(v[:c:c], v[c+1:]...)
		}
	case "enter":
		if m.focused == fieldSystemPrompt {
			m.insertRune('\n')
		} else {
			m.focused = (m.focused + 1) % numFields
		}
	default:
		if len(msg.Runes) == 1 {
			m.insertRune(msg.Runes[0])
		}
	}

	return m, nil
}

func (m *Model) insertRune(r rune) {
	v := m.values[m.focused]
	c := m.cursors[m.focused]
	newV := make([]rune, len(v)+1)
	copy(newV, v[:c])
	newV[c] = r
	copy(newV[c+1:], v[c:])
	m.values[m.focused] = newV
	m.cursors[m.focused]++
}

func (m *Model) buildSaveCmd() tea.Cmd {
	cfg := m.appConfig
	cfg.AI.Provider = string(m.values[fieldProvider])
	cfg.AI.BaseURL = string(m.values[fieldBaseURL])
	cfg.AI.APIKey = string(m.values[fieldAPIKey])
	cfg.AI.Model = string(m.values[fieldModel])
	cfg.AI.SystemPrompt = string(m.values[fieldSystemPrompt])
	cfg.BackgroundColor = string(m.values[fieldBgColor])
	return func() tea.Msg {
		return SaveMsg{Config: cfg}
	}
}

// View renders the settings modal.
func (m *Model) View() string {
	modalWidth := m.width - 8
	if modalWidth > 76 {
		modalWidth = 76
	}
	if modalWidth < 40 {
		modalWidth = 40
	}
	innerWidth := modalWidth - 4 // account for border + padding

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	focusedBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(0, 1).
		Width(innerWidth - 2)

	normalBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(0, 1).
		Width(innerWidth - 2)

	var lines []string
	lines = append(lines, titleStyle.Render("Settings"))
	lines = append(lines, "")

	for i := 0; i < numFields; i++ {
		lines = append(lines, labelStyle.Render(fieldLabels[i]))

		if i == fieldProvider {
			provider := string(m.values[fieldProvider])
			ollamaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			openaiStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
			if provider == "ollama" {
				ollamaStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
			} else {
				openaiStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
			}
			providerLine := ollamaStyle.Render("● Ollama") + "   " + openaiStyle.Render("● OpenAI Compatible")
			if i == m.focused {
				providerLine = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Render("  ") + providerLine
			} else {
				providerLine = "  " + providerLine
			}
			lines = append(lines, providerLine)
			lines = append(lines, "")
			continue
		}

		// Render text value with cursor
		fieldContent := m.renderFieldContent(i, innerWidth-4)

		if i == fieldSystemPrompt {
			// Multi-line textarea
			var fieldStyle lipgloss.Style
			if i == m.focused {
				fieldStyle = focusedBorderStyle.Height(4)
			} else {
				fieldStyle = normalBorderStyle.Height(4)
			}
			lines = append(lines, fieldStyle.Render(fieldContent))
		} else if i == fieldBgColor {
			// Single-line field with hint
			var fieldStyle lipgloss.Style
			if i == m.focused {
				fieldStyle = focusedBorderStyle
			} else {
				fieldStyle = normalBorderStyle
			}
			lines = append(lines, fieldStyle.Render(fieldContent))
			lines = append(lines, hintStyle.Render("  e.g. 0, 235, #1a1a2e — empty = terminal default"))
		} else {
			// Single-line field
			var fieldStyle lipgloss.Style
			if i == m.focused {
				fieldStyle = focusedBorderStyle
			} else {
				fieldStyle = normalBorderStyle
			}
			lines = append(lines, fieldStyle.Render(fieldContent))
		}
		lines = append(lines, "")
	}

	lines = append(lines, hintStyle.Render("Tab: next field    Ctrl+S: save    Esc: cancel"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(1, 2).
		Width(modalWidth)

	box := boxStyle.Render(strings.Join(lines, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

// renderFieldContent renders the field value with cursor inserted.
func (m *Model) renderFieldContent(field, maxWidth int) string {
	v := m.values[field]
	c := m.cursors[field]

	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15"))

	if field == fieldSystemPrompt {
		// Multi-line: render lines, inserting cursor at correct position
		full := string(v[:c]) + "│" + string(v[c:])
		return full
	}

	// Single-line: truncate to fit, keeping cursor visible
	var display string
	if c >= len(v) {
		// Cursor at end
		display = string(v) + cursorStyle.Render(" ")
	} else {
		before := string(v[:c])
		at := string(v[c : c+1])
		after := string(v[c+1:])
		display = before + cursorStyle.Render(at) + after
	}

	// Truncate from left if too long
	raw := string(v)
	if len(raw) > maxWidth {
		start := c - maxWidth + 1
		if start < 0 {
			start = 0
		}
		if start > c {
			start = c
		}
		runeStart := start
		if runeStart+maxWidth > len(v) {
			runeStart = len(v) - maxWidth
			if runeStart < 0 {
				runeStart = 0
			}
		}
		localC := c - runeStart
		vSlice := v[runeStart:]
		if localC >= len(vSlice) {
			display = string(vSlice) + cursorStyle.Render(" ")
		} else {
			before := string(vSlice[:localC])
			at := string(vSlice[localC : localC+1])
			after := string(vSlice[localC+1:])
			display = before + cursorStyle.Render(at) + after
		}
	}

	return display
}
