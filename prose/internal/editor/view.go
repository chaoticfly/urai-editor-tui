package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("7")).
			Foreground(lipgloss.Color("0"))

	selectionStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("4")).
			Foreground(lipgloss.Color("15"))

	lineNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Width(4).
			Align(lipgloss.Right)

	matchStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("3")).
			Foreground(lipgloss.Color("0"))

	currentMatchStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("11")).
				Foreground(lipgloss.Color("0"))
)

// View renders the editor.
func (m *Model) View() string {
	lines := m.buffer.Lines()
	var sb strings.Builder

	// Calculate visible range
	startLine := m.scrollOffset
	endLine := startLine + m.pageSize
	if endLine > len(lines) {
		endLine = len(lines)
	}

	// Get cursor position
	cursorLine, cursorCol := m.cursor.LineCol()

	// Get selection if any
	sel := m.selection.GetSelection()
	var selStart, selEnd int
	if sel != nil {
		selStart = sel.Start
		selEnd = sel.End
	}

	// Advance through sorted search matches, skipping those that end before
	// the first visible line so the inner loop starts at the right index.
	matchIdx := 0
	if m.searchMatchLen > 0 && len(m.searchMatches) > 0 {
		visibleStart := m.buffer.LineStart(startLine)
		for matchIdx < len(m.searchMatches) &&
			m.searchMatches[matchIdx]+m.searchMatchLen <= visibleStart {
			matchIdx++
		}
	}

	// Render visible lines
	for lineNum := startLine; lineNum < endLine; lineNum++ {
		line := ""
		if lineNum < len(lines) {
			line = lines[lineNum]
		}

		// Line number
		lineNumStr := lineNumberStyle.Render(fmt.Sprintf("%d", lineNum+1))
		sb.WriteString(lineNumStr)
		sb.WriteString(" ")

		// Line content with cursor, selection, and search highlights.
		lineStart := m.buffer.LineStart(lineNum)
		runes := []rune(line)

		for col := 0; col <= len(runes); col++ {
			pos := lineStart + col
			isCursor := lineNum == cursorLine && col == cursorCol
			inSelection := sel != nil && pos >= selStart && pos < selEnd

			// Advance past matches that end before this position.
			for matchIdx < len(m.searchMatches) &&
				m.searchMatches[matchIdx]+m.searchMatchLen <= pos {
				matchIdx++
			}
			inMatch := false
			isCurrent := false
			if m.searchMatchLen > 0 && matchIdx < len(m.searchMatches) {
				mp := m.searchMatches[matchIdx]
				if pos >= mp && pos < mp+m.searchMatchLen {
					inMatch = true
					isCurrent = matchIdx == m.searchCurrentMatch
				}
			}

			if col < len(runes) {
				char := string(runes[col])
				switch {
				case isCursor:
					sb.WriteString(cursorStyle.Render(char))
				case inSelection:
					sb.WriteString(selectionStyle.Render(char))
				case isCurrent:
					sb.WriteString(currentMatchStyle.Render(char))
				case inMatch:
					sb.WriteString(matchStyle.Render(char))
				default:
					sb.WriteString(char)
				}
			} else if isCursor {
				sb.WriteString(cursorStyle.Render(" "))
			}
		}

		sb.WriteString("\n")
	}

	// Pad remaining lines
	for lineNum := endLine; lineNum < startLine+m.pageSize; lineNum++ {
		lineNumStr := lineNumberStyle.Render("~")
		sb.WriteString(lineNumStr)
		sb.WriteString("\n")
	}

	return sb.String()
}

// ViewFocus renders the editor without line numbers or tilde markers (for focus mode).
func (m *Model) ViewFocus() string {
	lines := m.buffer.Lines()
	var sb strings.Builder

	startLine := m.scrollOffset
	endLine := startLine + m.pageSize
	if endLine > len(lines) {
		endLine = len(lines)
	}

	cursorLine, cursorCol := m.cursor.LineCol()
	sel := m.selection.GetSelection()
	var selStart, selEnd int
	if sel != nil {
		selStart = sel.Start
		selEnd = sel.End
	}

	for lineNum := startLine; lineNum < endLine; lineNum++ {
		line := ""
		if lineNum < len(lines) {
			line = lines[lineNum]
		}

		lineStart := m.buffer.LineStart(lineNum)
		runes := []rune(line)

		for col := 0; col <= len(runes); col++ {
			pos := lineStart + col
			isCursor := lineNum == cursorLine && col == cursorCol
			inSelection := sel != nil && pos >= selStart && pos < selEnd

			if col < len(runes) {
				char := string(runes[col])
				if isCursor {
					sb.WriteString(cursorStyle.Render(char))
				} else if inSelection {
					sb.WriteString(selectionStyle.Render(char))
				} else {
					sb.WriteString(char)
				}
			} else if isCursor {
				sb.WriteString(cursorStyle.Render(" "))
			}
		}

		sb.WriteString("\n")
	}

	// Pad remaining lines with blank lines (no ~ markers)
	remaining := (startLine + m.pageSize) - endLine
	for i := 0; i < remaining; i++ {
		sb.WriteString("\n")
	}

	return sb.String()
}

// StatusBar renders the status bar.
func (m *Model) StatusBar() string {
	// Left side: filename and modified indicator
	filename := m.filepath
	if filename == "" {
		filename = "[No Name]"
	}
	if m.modified {
		filename += " [+]"
	}

	// Right side: cursor position
	line, col := m.cursor.LineCol()
	position := fmt.Sprintf("Ln %d, Col %d", line+1, col+1)

	// Calculate spacing
	leftLen := len(filename)
	rightLen := len(position)
	spaces := m.width - leftLen - rightLen
	if spaces < 1 {
		spaces = 1
	}

	statusStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("7")).
		Foreground(lipgloss.Color("0")).
		Width(m.width)

	status := filename + strings.Repeat(" ", spaces) + position

	return statusStyle.Render(status)
}
