package findbar

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mode controls whether the bar shows find-only or find+replace.
type Mode int

const (
	ModeFind    Mode = iota
	ModeReplace      // find + replace fields
)

// CloseMsg is sent when the find bar is dismissed.
type CloseMsg struct{}

// JumpMsg is sent when the editor should scroll to a match position.
type JumpMsg struct {
	Pos int
	Len int
}

// ReplaceMsg asks the editor to replace text at the given buffer position.
type ReplaceMsg struct {
	Pos         int
	Len         int
	Replacement string
}

// ReplaceAllMsg asks the editor to replace all current matches.
type ReplaceAllMsg struct {
	Matches     []int
	MatchLen    int
	Replacement string
}

// Model is the find/replace bar state.
type Model struct {
	mode         Mode
	query        string
	replaceText  string
	focusReplace bool // true = text cursor is in the replace field

	matches      []int // rune positions of match starts in content
	matchLen     int   // rune length of the search query
	currentMatch int   // index into matches (-1 when no matches)

	content string // last content the search ran against

	width int
}

// New returns a find-only bar.
func New(width int) *Model {
	return &Model{width: width, currentMatch: -1}
}

// NewReplace returns a find+replace bar.
func NewReplace(width int) *Model {
	return &Model{width: width, mode: ModeReplace, currentMatch: -1}
}

func (m *Model) SetWidth(w int)    { m.width = w }
func (m *Model) Query() string     { return m.query }
func (m *Model) Matches() []int    { return m.matches }
func (m *Model) MatchLen() int     { return m.matchLen }
func (m *Model) CurrentMatch() int { return m.currentMatch }
func (m *Model) Mode() Mode        { return m.mode }

// SetMode switches between find and find+replace without clearing state.
func (m *Model) SetMode(mode Mode) {
	m.mode = mode
	if mode == ModeFind {
		m.focusReplace = false
	}
}

// SetContent re-runs the search against new document content.
// Call this when the editor buffer changes while the bar is open.
func (m *Model) SetContent(content string) {
	if m.content == content {
		return
	}
	m.content = content
	m.runSearch()
}

func (m *Model) runSearch() {
	m.matches = nil
	m.matchLen = 0
	if m.query == "" {
		m.currentMatch = -1
		return
	}

	lowerQuery := strings.ToLower(m.query)
	lowerContent := strings.ToLower(m.content)
	queryRunes := []rune(lowerQuery)
	contentRunes := []rune(lowerContent)
	m.matchLen = len(queryRunes)

	for i := 0; i <= len(contentRunes)-m.matchLen; i++ {
		match := true
		for j := 0; j < m.matchLen; j++ {
			if contentRunes[i+j] != queryRunes[j] {
				match = false
				break
			}
		}
		if match {
			m.matches = append(m.matches, i)
		}
	}

	// Keep currentMatch in bounds.
	if m.currentMatch >= len(m.matches) {
		m.currentMatch = 0
	}
	if len(m.matches) == 0 {
		m.currentMatch = -1
	}
}

func (m *Model) advance() tea.Cmd {
	if len(m.matches) == 0 {
		return nil
	}
	if m.currentMatch < 0 {
		m.currentMatch = 0
	} else {
		m.currentMatch = (m.currentMatch + 1) % len(m.matches)
	}
	pos := m.matches[m.currentMatch]
	return func() tea.Msg { return JumpMsg{Pos: pos, Len: m.matchLen} }
}

func (m *Model) retreat() tea.Cmd {
	if len(m.matches) == 0 {
		return nil
	}
	if m.currentMatch < 0 {
		m.currentMatch = len(m.matches) - 1
	} else {
		m.currentMatch = (m.currentMatch - 1 + len(m.matches)) % len(m.matches)
	}
	pos := m.matches[m.currentMatch]
	return func() tea.Msg { return JumpMsg{Pos: pos, Len: m.matchLen} }
}

// CurrentJumpCmd returns a JumpMsg for the current match (used on open).
func (m *Model) CurrentJumpCmd() tea.Cmd {
	if m.currentMatch < 0 || len(m.matches) == 0 {
		return nil
	}
	pos := m.matches[m.currentMatch]
	return func() tea.Msg { return JumpMsg{Pos: pos, Len: m.matchLen} }
}

// Height returns the number of terminal lines this bar occupies.
func (m *Model) Height() int {
	if m.mode == ModeReplace {
		return 2
	}
	return 1
}

// Update handles key events while the find bar is active.
func (m *Model) Update(msg tea.KeyMsg) (*Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		return m, func() tea.Msg { return CloseMsg{} }

	case "enter":
		if m.mode == ModeReplace && m.focusReplace {
			if m.currentMatch < 0 || len(m.matches) == 0 {
				return m, nil
			}
			pos := m.matches[m.currentMatch]
			repl := m.replaceText
			return m, func() tea.Msg {
				return ReplaceMsg{Pos: pos, Len: m.matchLen, Replacement: repl}
			}
		}
		// Find field or find-only mode: advance to next match.
		return m, m.advance()

	case "ctrl+g":
		return m, m.advance()

	case "tab":
		if m.mode == ModeReplace {
			m.focusReplace = !m.focusReplace
		}
		return m, nil

	case "shift+tab":
		if m.mode == ModeReplace {
			m.focusReplace = !m.focusReplace
		}
		return m, nil

	case "ctrl+a":
		if m.mode == ModeReplace && len(m.matches) > 0 {
			matches := make([]int, len(m.matches))
			copy(matches, m.matches)
			repl := m.replaceText
			return m, func() tea.Msg {
				return ReplaceAllMsg{Matches: matches, MatchLen: m.matchLen, Replacement: repl}
			}
		}
		return m, nil

	case "backspace":
		if m.focusReplace {
			if r := []rune(m.replaceText); len(r) > 0 {
				m.replaceText = string(r[:len(r)-1])
			}
		} else {
			if r := []rune(m.query); len(r) > 0 {
				m.query = string(r[:len(r)-1])
				m.runSearch()
				return m, m.CurrentJumpCmd()
			}
		}
		return m, nil

	default:
		if len(msg.Runes) > 0 {
			if m.focusReplace {
				m.replaceText += string(msg.Runes)
			} else {
				m.query += string(msg.Runes)
				m.runSearch()
				// Jump to first match automatically as the user types.
				if len(m.matches) > 0 && m.currentMatch < 0 {
					m.currentMatch = 0
				}
				return m, m.CurrentJumpCmd()
			}
		}
		return m, nil
	}
}

// View renders the find bar as a 1- or 2-line string.
func (m *Model) View() string {
	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("15")).
		Width(m.width)

	labelStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		Background(lipgloss.Color("235"))

	inputStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("237")).
		Foreground(lipgloss.Color("15")).
		Padding(0, 1)

	activeInputStyle := inputStyle.Copy().
		Background(lipgloss.Color("240"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Background(lipgloss.Color("235"))

	warnStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("9")).
		Background(lipgloss.Color("235"))

	// Match count indicator.
	var matchInfo string
	if m.query != "" {
		if len(m.matches) == 0 {
			matchInfo = warnStyle.Render(" no matches")
		} else {
			matchInfo = dimStyle.Render(fmt.Sprintf(" %d/%d", m.currentMatch+1, len(m.matches)))
		}
	}

	// Find field.
	findText := m.query
	if !m.focusReplace {
		findText += "█"
	}
	var findInput string
	if !m.focusReplace {
		findInput = activeInputStyle.Render(findText)
	} else {
		findInput = inputStyle.Render(findText)
	}

	line1 := labelStyle.Render("Find: ") + findInput + matchInfo

	if m.mode == ModeFind {
		hint := dimStyle.Render("  Enter·next  Ctrl+G·next  Esc·close")
		return barStyle.Render(line1 + hint)
	}

	// Replace mode — second line.
	replText := m.replaceText
	if m.focusReplace {
		replText += "█"
	}
	var replInput string
	if m.focusReplace {
		replInput = activeInputStyle.Render(replText)
	} else {
		replInput = inputStyle.Render(replText)
	}

	hint := dimStyle.Render("  Tab·switch  Enter·replace  Ctrl+A·all  Esc·close")
	line2 := labelStyle.Render("Repl: ") + replInput + hint

	return barStyle.Render(line1) + "\n" + barStyle.Render(line2)
}
