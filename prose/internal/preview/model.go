package preview

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// debounceInterval is the time to wait before re-rendering.
const debounceInterval = 100 * time.Millisecond

// Model represents the preview pane state.
type Model struct {
	renderer      *Renderer
	content       string
	rendered      string
	width         int
	height        int
	scrollOffset  int
	needsRender   bool
	lastContent   string
	debounceTimer *time.Timer
}

// New creates a new preview model.
func New(width, height int) *Model {
	renderer, _ := NewRenderer(width)
	return &Model{
		renderer:     renderer,
		width:        width,
		height:       height,
		scrollOffset: 0,
		needsRender:  false,
	}
}

// Init initializes the model.
func (m *Model) Init() tea.Cmd {
	return nil
}

// renderMsg signals that content should be re-rendered.
type renderMsg struct{}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.renderer != nil {
			m.renderer.SetWidth(m.width)
		}
		m.needsRender = true
		return m, nil

	case renderMsg:
		m.doRender()
		return m, nil
	}

	return m, nil
}

// SetContent sets the markdown content to preview.
func (m *Model) SetContent(content string) tea.Cmd {
	if content == m.lastContent {
		return nil
	}

	m.content = content
	m.lastContent = content
	m.needsRender = true

	// Return a command that will trigger rendering after debounce
	return func() tea.Msg {
		time.Sleep(debounceInterval)
		return renderMsg{}
	}
}

// SetContentImmediate sets content and renders immediately.
func (m *Model) SetContentImmediate(content string) {
	m.content = content
	m.lastContent = content
	m.doRender()
}

// doRender performs the actual rendering.
func (m *Model) doRender() {
	if m.renderer == nil {
		m.rendered = m.content
		return
	}

	rendered, err := m.renderer.Render(m.content)
	if err != nil {
		m.rendered = m.content
		return
	}

	m.rendered = rendered
	m.needsRender = false
}

// View renders the preview.
func (m *Model) View() string {
	if m.needsRender {
		m.doRender()
	}

	lines := strings.Split(m.rendered, "\n")

	// Apply scroll offset
	if m.scrollOffset > 0 && m.scrollOffset < len(lines) {
		lines = lines[m.scrollOffset:]
	}

	// Limit to height
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

// SetSize updates the preview dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	if m.renderer != nil {
		m.renderer.SetWidth(width)
	}
	m.needsRender = true
}

// Rendered returns the current rendered content.
func (m *Model) Rendered() string {
	return m.rendered
}
