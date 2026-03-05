package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ViewMode represents the current view mode.
type ViewMode int

const (
	ViewModeEdit    ViewMode = iota // editor only
	ViewModeSplit                   // editor + markdown preview
	ViewModePreview                 // preview only
	ViewModeAI                      // editor + AI suggestions
	ViewModeFocus                   // distraction-free editor
)

// CycleViewMode cycles through Edit/Split/Preview/AI; Focus is a separate toggle.
func CycleViewMode(current ViewMode) ViewMode {
	switch current {
	case ViewModeEdit:
		return ViewModeSplit
	case ViewModeSplit:
		return ViewModePreview
	case ViewModePreview:
		return ViewModeAI
	case ViewModeAI:
		return ViewModeEdit
	default:
		return ViewModeEdit
	}
}

// ViewModeName returns the name of the view mode.
func ViewModeName(mode ViewMode) string {
	switch mode {
	case ViewModeEdit:
		return "Edit"
	case ViewModeSplit:
		return "Split"
	case ViewModePreview:
		return "Preview"
	case ViewModeAI:
		return "AI"
	case ViewModeFocus:
		return "Focus"
	default:
		return "Unknown"
	}
}

// SplitLayout manages a horizontal split view.
type SplitLayout struct {
	Width      int
	Height     int
	LeftRatio  float64 // Ratio of width for left pane (0.0-1.0)
	GapWidth   int     // Width of gap between panes
}

// NewSplitLayout creates a new split layout.
func NewSplitLayout(width, height int) *SplitLayout {
	return &SplitLayout{
		Width:     width,
		Height:    height,
		LeftRatio: 0.5,
		GapWidth:  1,
	}
}

// SetSize updates the layout dimensions.
func (s *SplitLayout) SetSize(width, height int) {
	s.Width = width
	s.Height = height
}

// LeftWidth returns the width of the left pane.
func (s *SplitLayout) LeftWidth() int {
	return int(float64(s.Width-s.GapWidth) * s.LeftRatio)
}

// RightWidth returns the width of the right pane.
func (s *SplitLayout) RightWidth() int {
	return s.Width - s.LeftWidth() - s.GapWidth
}

// ContentHeight returns the height available for content (minus status bar).
func (s *SplitLayout) ContentHeight() int {
	return s.Height - 1
}

// Render combines left and right content into a split view.
func (s *SplitLayout) Render(left, right string) string {
	leftWidth := s.LeftWidth()
	rightWidth := s.RightWidth()
	contentHeight := s.ContentHeight()

	// Split content into lines
	leftLines := splitLines(left, contentHeight)
	rightLines := splitLines(right, contentHeight)

	// Styles for panes
	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(contentHeight)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(contentHeight)

	dividerStyle := lipgloss.NewStyle().
		Width(s.GapWidth).
		Foreground(lipgloss.Color("8"))

	// Build divider
	divider := strings.Repeat("│\n", contentHeight)
	divider = strings.TrimSuffix(divider, "\n")

	// Render panes
	leftPane := leftStyle.Render(strings.Join(leftLines, "\n"))
	rightPane := rightStyle.Render(strings.Join(rightLines, "\n"))
	dividerPane := dividerStyle.Render(divider)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, dividerPane, rightPane)
}

// RenderFocus renders content centered for distraction-free writing.
func (s *SplitLayout) RenderFocus(content string, focusWidth, contentHeight int) string {
	if focusWidth > s.Width {
		focusWidth = s.Width
	}

	inner := lipgloss.NewStyle().
		Width(focusWidth).
		Height(contentHeight).
		Render(content)

	return lipgloss.Place(s.Width, contentHeight, lipgloss.Center, lipgloss.Top, inner)
}

// RenderSingle renders content as a full-width pane.
func (s *SplitLayout) RenderSingle(content string) string {
	contentHeight := s.ContentHeight()
	lines := splitLines(content, contentHeight)

	style := lipgloss.NewStyle().
		Width(s.Width).
		Height(contentHeight)

	return style.Render(strings.Join(lines, "\n"))
}

// splitLines splits content into lines and pads/truncates to height.
func splitLines(content string, height int) []string {
	lines := strings.Split(content, "\n")

	// Pad with empty lines if needed
	for len(lines) < height {
		lines = append(lines, "")
	}

	// Truncate if needed
	if len(lines) > height {
		lines = lines[:height]
	}

	return lines
}
