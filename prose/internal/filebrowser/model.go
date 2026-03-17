package filebrowser

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Messages emitted to the parent app.

// OpenFileMsg tells the app to open a file in the editor.
type OpenFileMsg struct{ Path string }

// SaveAsMsg tells the app to save current editor content to the given path.
type SaveAsMsg struct{ Path string }

// CloseMsg tells the app to close the file browser.
type CloseMsg struct{}

type browserMode int

const (
	modeBrowse browserMode = iota
	modeNew
	modeRename
	modeDelete
)

type entry struct {
	name  string
	isDir bool
}

// Model is the file browser state.
type Model struct {
	rootDir    string // navigation cannot go above this directory
	currentDir string
	entries    []entry
	selected   int
	mode       browserMode
	saveAs     bool   // when true, selecting/naming a file saves editor content there
	input      []rune // for modeNew / modeRename
	inputCursor int
	width      int
	height     int
	errMsg     string
}

// homeDir returns the user's home directory, falling back to cwd.
func homeDir() string {
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	d, _ := os.Getwd()
	return d
}

// New creates a file browser rooted at the user's home directory.
// startDir is the initial directory shown; it is clamped to be inside home.
func New(startDir string, width, height int) *Model {
	root := homeDir()
	if startDir == "" || !strings.HasPrefix(filepath.Clean(startDir), filepath.Clean(root)) {
		startDir = root
	}
	m := &Model{
		rootDir:    root,
		currentDir: startDir,
		width:      width,
		height:     height,
	}
	m.loadDir()
	return m
}

// NewSaveAs creates a file browser in save-as mode.
func NewSaveAs(startDir string, width, height int) *Model {
	m := New(startDir, width, height)
	m.saveAs = true
	return m
}

// SetSize updates dimensions.
func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) loadDir() {
	m.errMsg = ""
	des, err := os.ReadDir(m.currentDir)
	if err != nil {
		m.errMsg = err.Error()
		m.entries = nil
		return
	}
	sort.Slice(des, func(i, j int) bool {
		if des[i].IsDir() != des[j].IsDir() {
			return des[i].IsDir()
		}
		return strings.ToLower(des[i].Name()) < strings.ToLower(des[j].Name())
	})
	// Only show ".." if we haven't reached the root boundary.
	if filepath.Clean(m.currentDir) != filepath.Clean(m.rootDir) {
		m.entries = []entry{{name: "..", isDir: true}}
	}
	for _, de := range des {
		m.entries = append(m.entries, entry{name: de.Name(), isDir: de.IsDir()})
	}
	if m.selected >= len(m.entries) {
		m.selected = 0
	}
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

	switch m.mode {
	case modeNew:
		return m.handleInputKey(key, msg, func(name string) tea.Cmd {
			path := filepath.Join(m.currentDir, name)
			if m.saveAs {
				return func() tea.Msg { return SaveAsMsg{Path: path} }
			}
			return func() tea.Msg { return OpenFileMsg{Path: path} }
		})

	case modeRename:
		if len(m.entries) == 0 || m.selected >= len(m.entries) {
			m.mode = modeBrowse
			return m, nil
		}
		oldName := m.entries[m.selected].name
		return m.handleInputKey(key, msg, func(newName string) tea.Cmd {
			oldPath := filepath.Join(m.currentDir, oldName)
			newPath := filepath.Join(m.currentDir, newName)
			if err := os.Rename(oldPath, newPath); err != nil {
				m.errMsg = err.Error()
				return nil
			}
			m.loadDir()
			return nil
		})

	case modeDelete:
		return m.handleDeleteKey(key)
	}

	// modeBrowse
	return m.handleBrowseKey(key)
}

func (m *Model) handleBrowseKey(key string) (*Model, tea.Cmd) {
	switch key {
	case "esc":
		return m, func() tea.Msg { return CloseMsg{} }

	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}

	case "down", "j":
		if m.selected < len(m.entries)-1 {
			m.selected++
		}

	case "enter", "right", "l":
		if m.selected < len(m.entries) {
			e := m.entries[m.selected]
			if e.isDir {
				m.enterDir(e.name)
			} else {
				path := filepath.Join(m.currentDir, e.name)
				if m.saveAs {
					return m, func() tea.Msg { return SaveAsMsg{Path: path} }
				}
				return m, func() tea.Msg { return OpenFileMsg{Path: path} }
			}
		}

	case "left", "h", "backspace":
		m.enterDir("..")

	case "n":
		m.mode = modeNew
		m.input = nil
		m.inputCursor = 0

	case "r":
		if m.selected < len(m.entries) && m.entries[m.selected].name != ".." {
			m.mode = modeRename
			m.input = []rune(m.entries[m.selected].name)
			m.inputCursor = len(m.input)
		}

	case "d":
		if m.selected < len(m.entries) && m.entries[m.selected].name != ".." {
			m.mode = modeDelete
		}
	}

	return m, nil
}

func (m *Model) handleDeleteKey(key string) (*Model, tea.Cmd) {
	switch key {
	case "y", "Y":
		if m.selected < len(m.entries) {
			e := m.entries[m.selected]
			if e.name != ".." {
				path := filepath.Join(m.currentDir, e.name)
				var err error
				if e.isDir {
					err = os.RemoveAll(path)
				} else {
					err = os.Remove(path)
				}
				if err != nil {
					m.errMsg = err.Error()
				} else {
					if m.selected > 0 {
						m.selected--
					}
					m.loadDir()
				}
			}
		}
		m.mode = modeBrowse
	case "n", "N", "esc":
		m.mode = modeBrowse
	}
	return m, nil
}

func (m *Model) handleInputKey(key string, msg tea.KeyMsg, onConfirm func(string) tea.Cmd) (*Model, tea.Cmd) {
	switch key {
	case "esc":
		m.mode = modeBrowse
		m.input = nil
	case "enter":
		name := strings.TrimSpace(string(m.input))
		if name != "" && name != ".." {
			cmd := onConfirm(name)
			m.mode = modeBrowse
			m.input = nil
			if cmd != nil {
				return m, cmd
			}
		}
	case "backspace":
		if m.inputCursor > 0 {
			v := m.input
			m.input = append(v[:m.inputCursor-1:m.inputCursor-1], v[m.inputCursor:]...)
			m.inputCursor--
		}
	case "delete":
		v := m.input
		c := m.inputCursor
		if c < len(v) {
			m.input = append(v[:c:c], v[c+1:]...)
		}
	case "left":
		if m.inputCursor > 0 {
			m.inputCursor--
		}
	case "right":
		if m.inputCursor < len(m.input) {
			m.inputCursor++
		}
	case "home", "ctrl+a":
		m.inputCursor = 0
	case "end", "ctrl+e":
		m.inputCursor = len(m.input)
	default:
		if len(msg.Runes) == 1 {
			v := m.input
			c := m.inputCursor
			newV := make([]rune, len(v)+1)
			copy(newV, v[:c])
			newV[c] = msg.Runes[0]
			copy(newV[c+1:], v[c:])
			m.input = newV
			m.inputCursor++
		}
	}
	return m, nil
}

func (m *Model) enterDir(name string) {
	if name == ".." {
		parent := filepath.Dir(m.currentDir)
		root := filepath.Clean(m.rootDir)
		if parent != m.currentDir && strings.HasPrefix(filepath.Clean(parent), root) {
			m.currentDir = parent
		}
	} else {
		m.currentDir = filepath.Join(m.currentDir, name)
	}
	m.selected = 0
	m.loadDir()
}

// View renders the file browser as a full-screen modal.
func (m *Model) View() string {
	modalWidth := m.width - 8
	if modalWidth > 84 {
		modalWidth = 84
	}
	if modalWidth < 44 {
		modalWidth = 44
	}
	innerWidth := modalWidth - 6 // padding + border

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dirPathStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	divStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	dirEntryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	fileEntryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	selectedDirStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15")).Bold(true).Width(innerWidth)
	selectedFileStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15")).Width(innerWidth)
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	cursorStyle := lipgloss.NewStyle().Background(lipgloss.Color("4")).Foreground(lipgloss.Color("15"))

	title := "File Browser"
	if m.saveAs {
		title = "Save As"
	}

	var lines []string
	lines = append(lines, titleStyle.Render(title))
	lines = append(lines, dirPathStyle.Render(truncatePath(m.currentDir, innerWidth)))
	lines = append(lines, divStyle.Render(strings.Repeat("─", innerWidth)))

	if m.errMsg != "" {
		lines = append(lines, errStyle.Render(m.errMsg))
	}

	// Reserve lines for header, footer, input prompts
	listHeight := m.height - 14
	if listHeight < 4 {
		listHeight = 4
	}

	// Scroll window
	startIdx := 0
	if m.selected >= listHeight {
		startIdx = m.selected - listHeight + 1
	}

	for i := startIdx; i < len(m.entries) && i < startIdx+listHeight; i++ {
		e := m.entries[i]
		var raw string
		if e.isDir {
			raw = "▶ " + e.name + "/"
		} else {
			raw = "  " + e.name
		}

		var rendered string
		if i == m.selected {
			if e.isDir {
				rendered = selectedDirStyle.Render(raw)
			} else {
				rendered = selectedFileStyle.Render(raw)
			}
		} else {
			if e.isDir {
				rendered = dirEntryStyle.Render(raw)
			} else {
				rendered = fileEntryStyle.Render(raw)
			}
		}
		lines = append(lines, rendered)
	}

	lines = append(lines, "")

	// Mode-specific prompt
	switch m.mode {
	case modeNew:
		prompt := "New file name:"
		if m.saveAs {
			prompt = "Save as:"
		}
		lines = append(lines, labelStyle.Render(prompt))
		lines = append(lines, renderInput(m.input, m.inputCursor, innerWidth, cursorStyle))
		lines = append(lines, hintStyle.Render("Enter: confirm    Esc: cancel"))

	case modeRename:
		if m.selected < len(m.entries) {
			lines = append(lines, labelStyle.Render("Rename to:"))
			lines = append(lines, renderInput(m.input, m.inputCursor, innerWidth, cursorStyle))
			lines = append(lines, hintStyle.Render("Enter: rename    Esc: cancel"))
		}

	case modeDelete:
		if m.selected < len(m.entries) {
			name := m.entries[m.selected].name
			lines = append(lines, warnStyle.Render("Delete \""+name+"\"?"))
			lines = append(lines, hintStyle.Render("y: confirm    n / Esc: cancel"))
		}

	default:
		if m.saveAs {
			lines = append(lines, hintStyle.Render("Enter:save here  n:new name  ↑↓:nav  ←→:dir  Esc:cancel"))
		} else {
			lines = append(lines, hintStyle.Render("Enter:open  n:new  r:rename  d:delete  ↑↓:nav  ←→:dir  Esc:close"))
		}
	}

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(1, 2).
		Width(modalWidth)

	box := boxStyle.Render(strings.Join(lines, "\n"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

func renderInput(v []rune, cursor, maxWidth int, cs lipgloss.Style) string {
	if len(v) == 0 {
		return cs.Render(" ")
	}
	if cursor >= len(v) {
		return string(v) + cs.Render(" ")
	}
	return string(v[:cursor]) + cs.Render(string(v[cursor:cursor+1])) + string(v[cursor+1:])
}

func truncatePath(p string, maxWidth int) string {
	if maxWidth <= 0 || len(p) <= maxWidth {
		return p
	}
	return "…" + p[len(p)-maxWidth+1:]
}
