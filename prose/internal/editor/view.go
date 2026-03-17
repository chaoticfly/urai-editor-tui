package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// cursorOn/cursorOff emit raw ANSI reverse-video, bypassing lipgloss colour
// detection. This is necessary when the process stdout is not a TTY (e.g. a
// systemd service) — in that case lipgloss strips every attribute, making the
// cursor invisible. Raw escape codes reach the SSH terminal unconditionally.
const (
	cursorOn  = "\x1b[7m"  // reverse video on
	cursorOff = "\x1b[27m" // reverse video off
)

var (

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

	// contentWidth is the number of visible characters per visual row
	// (total width minus line-number prefix: 4 chars + 1 space).
	contentWidth := m.width - 5
	if contentWidth < 1 {
		contentWidth = 1
	}

	// Render buffer lines, wrapping long lines into multiple visual rows.
	// Stop when we've filled pageSize visual rows.
	visualRow := 0
	for lineNum := startLine; lineNum < len(lines) && visualRow < m.pageSize; lineNum++ {
		line := ""
		if lineNum < len(lines) {
			line = lines[lineNum]
		}

		lineStart := m.buffer.LineStart(lineNum)
		runes := []rune(line)

		// Number of visual chunks this buffer line needs.
		numChunks := (len(runes) + contentWidth - 1) / contentWidth
		if numChunks == 0 {
			numChunks = 1 // empty line still occupies one visual row
		}

		for chunkIdx := 0; chunkIdx < numChunks && visualRow < m.pageSize; chunkIdx++ {
			colStart := chunkIdx * contentWidth
			colEnd := colStart + contentWidth
			if colEnd > len(runes) {
				colEnd = len(runes)
			}

			// Line number prefix: show number on first chunk, indent on continuations.
			if chunkIdx == 0 {
				sb.WriteString(lineNumberStyle.Render(fmt.Sprintf("%d", lineNum+1)))
			} else {
				sb.WriteString(lineNumberStyle.Render("↪"))
			}
			sb.WriteString(" ")

			// Render characters for this chunk.
			for col := colStart; col <= colEnd; col++ {
				pos := lineStart + col
				isCursor := lineNum == cursorLine && col == cursorCol

				if col == colEnd {
					// Past the last char of this chunk — only draw cursor if it's here.
					if isCursor {
						sb.WriteString(cursorOn + " " + cursorOff)
					}
					break
				}

				inSelection := sel != nil && pos >= selStart && pos < selEnd

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

				char := string(runes[col])
				switch {
				case isCursor:
					sb.WriteString(cursorOn + char + cursorOff)
				case inSelection:
					sb.WriteString(selectionStyle.Render(char))
				case isCurrent:
					sb.WriteString(currentMatchStyle.Render(char))
				case inMatch:
					sb.WriteString(matchStyle.Render(char))
				default:
					sb.WriteString(char)
				}
			}

			sb.WriteString("\n")
			visualRow++
		}
	}

	// Pad remaining visual rows with tilde markers.
	for visualRow < m.pageSize {
		sb.WriteString(lineNumberStyle.Render("~"))
		sb.WriteString("\n")
		visualRow++
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
					sb.WriteString(cursorOn + char + cursorOff)
				} else if inSelection {
					sb.WriteString(selectionStyle.Render(char))
				} else {
					sb.WriteString(char)
				}
			} else if isCursor {
				sb.WriteString(cursorOn + " " + cursorOff)
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
