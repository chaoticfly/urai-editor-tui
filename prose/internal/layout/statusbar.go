package layout

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// formatCount formats an integer with thousands separators (e.g. 1247 → "1,247").
func formatCount(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	for i, ch := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(ch)
	}
	return b.String()
}

// StatusBar renders a minimal status bar.
type StatusBar struct {
	Width int
	Style lipgloss.Style
}

// NewStatusBar creates a new status bar.
func NewStatusBar(width int) *StatusBar {
	return &StatusBar{
		Width: width,
		Style: lipgloss.NewStyle().
			Background(lipgloss.Color("7")).
			Foreground(lipgloss.Color("0")),
	}
}

// SetWidth updates the status bar width.
func (s *StatusBar) SetWidth(width int) {
	s.Width = width
}

// StatusInfo contains information to display in the status bar.
type StatusInfo struct {
	Filename  string
	Modified  bool
	Line      int
	Col       int
	WordCount int
	ViewMode  ViewMode
	Message   string // Optional message (e.g., "Saved", error messages)
}

// Render renders the status bar.
func (s *StatusBar) Render(info StatusInfo) string {
	// Left section: filename and modified indicator
	filename := info.Filename
	if filename == "" {
		filename = "[No Name]"
	}
	if info.Modified {
		filename += " [+]"
	}

	// Center section: view mode
	viewMode := ViewModeName(info.ViewMode)

	// Right section: word count + cursor position
	position := fmt.Sprintf("%sw  Ln %d, Col %d", formatCount(info.WordCount), info.Line+1, info.Col+1)

	// Message takes precedence over center section
	center := viewMode
	if info.Message != "" {
		center = info.Message
	}

	// Calculate spacing
	leftLen := len(filename)
	centerLen := len(center)
	rightLen := len(position)

	totalContent := leftLen + centerLen + rightLen
	totalSpaces := s.Width - totalContent

	if totalSpaces < 2 {
		// Not enough space, just show left and right
		spaces := s.Width - leftLen - rightLen
		if spaces < 1 {
			spaces = 1
		}
		content := filename + strings.Repeat(" ", spaces) + position
		return s.Style.Width(s.Width).Render(content)
	}

	// Distribute spaces evenly
	leftSpaces := totalSpaces / 2
	rightSpaces := totalSpaces - leftSpaces

	content := filename + strings.Repeat(" ", leftSpaces) + center + strings.Repeat(" ", rightSpaces) + position

	return s.Style.Width(s.Width).Render(content)
}

// HelpBar renders a help bar with key hints.
type HelpBar struct {
	Width int
	Style lipgloss.Style
}

// NewHelpBar creates a new help bar.
func NewHelpBar(width int) *HelpBar {
	return &HelpBar{
		Width: width,
		Style: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
	}
}

// KeyHint represents a key and its action.
type KeyHint struct {
	Key    string
	Action string
}

// Render renders the help bar.
func (h *HelpBar) Render(hints []KeyHint) string {
	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7"))

	var parts []string
	for _, hint := range hints {
		part := keyStyle.Render(hint.Key) + " " + hint.Action
		parts = append(parts, part)
	}

	content := strings.Join(parts, "  │  ")

	return h.Style.Width(h.Width).Render(content)
}
