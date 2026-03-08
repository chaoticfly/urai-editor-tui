package preview

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
)

// RenderedMsg is sent when async rendering completes.
type RenderedMsg struct {
	content  string
	rendered string
	width    int
}

// Model represents the preview pane state.
type Model struct {
	style         string // glamour style name, e.g. "dark" or "light"
	content       string
	rendered      string
	renderedFor   string // content the current rendered output corresponds to
	renderedWidth int    // width at which it was rendered
	width         int
	height        int
	scrollOffset  int
}

// New creates a new preview model. style should be a glamour style name
// ("dark", "light", "dracula", etc.) and must be determined before the
// BubbleTea program starts so goroutines never query the terminal themselves.
func New(width, height int, style string) *Model {
	if style == "" {
		style = "dark"
	}
	return &Model{
		style:  style,
		width:  width,
		height: height,
	}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case RenderedMsg:
		// Discard stale results if content or width changed while rendering.
		if msg.content == m.content && msg.width == m.width {
			m.rendered = msg.rendered
			m.renderedFor = msg.content
			m.renderedWidth = msg.width
		}
		return m, nil
	}

	return m, nil
}

// SetContent schedules an async render if content or width changed.
// Returns nil if the cached output is already up to date.
func (m *Model) SetContent(content string) tea.Cmd {
	m.content = content
	if content == m.renderedFor && m.width == m.renderedWidth {
		return nil
	}
	return m.renderAsync(content, m.width)
}

// SetContentImmediate sets content and renders synchronously.
// Skips rendering if the cache is already current.
func (m *Model) SetContentImmediate(content string) {
	m.content = content
	if content == m.renderedFor && m.width == m.renderedWidth {
		return
	}
	rendered := m.renderSync(content, m.width)
	m.rendered = rendered
	m.renderedFor = content
	m.renderedWidth = m.width
}

// renderAsync returns a tea.Cmd that renders in a goroutine.
// A fresh Glamour renderer is created inside the goroutine. The style is a
// plain string (safe to capture) so no mutable state is shared with the main loop.
func (m *Model) renderAsync(content string, width int) tea.Cmd {
	style := m.style
	return func() tea.Msg {
		return RenderedMsg{
			content:  content,
			rendered: renderWithStyle(content, width, style),
			width:    width,
		}
	}
}

func (m *Model) renderSync(content string, width int) string {
	return renderWithStyle(content, width, m.style)
}

// renderWithStyle creates a one-shot Glamour renderer using a static style name.
// Using WithStandardStyle (not WithAutoStyle) avoids terminal OSC queries,
// which would corrupt output when called from goroutines during the program run.
func renderWithStyle(content string, width int, style string) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return content
	}
	rendered, err := r.Render(content)
	if err != nil {
		return content
	}
	return rendered
}

// View renders the preview.
func (m *Model) View() string {
	lines := strings.Split(m.rendered, "\n")

	if m.scrollOffset > 0 && m.scrollOffset < len(lines) {
		lines = lines[m.scrollOffset:]
	}
	if len(lines) > m.height {
		lines = lines[:m.height]
	}

	return strings.Join(lines, "\n")
}

// ScrollUp scrolls the preview up.
func (m *Model) ScrollUp() {
	if m.scrollOffset > 0 {
		m.scrollOffset--
	}
}

// ScrollDown scrolls the preview down.
func (m *Model) ScrollDown() {
	lines := strings.Split(m.rendered, "\n")
	maxScroll := len(lines) - m.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollOffset < maxScroll {
		m.scrollOffset++
	}
}

// SetStyle changes the Glamour style and schedules a re-render.
func (m *Model) SetStyle(style string) tea.Cmd {
	if style == "" {
		style = "dark"
	}
	if style == m.style {
		return nil
	}
	m.style = style
	// Invalidate cache so next render uses the new style.
	m.renderedFor = ""
	m.renderedWidth = 0
	return m.SetContent(m.content)
}

// SetSize updates the preview dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Rendered returns the current rendered content.
func (m *Model) Rendered() string {
	return m.rendered
}
