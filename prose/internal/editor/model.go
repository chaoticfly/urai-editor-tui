package editor

import (
	"os"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"urai/internal/buffer"
	"urai/internal/history"
)

// Model represents the editor state.
type Model struct {
	buffer    *buffer.GapBuffer
	cursor    *buffer.Cursor
	selection *buffer.SelectionManager
	history   *history.History
	keyMap    *KeyMap

	filepath string
	modified bool

	width  int
	height int

	scrollOffset int // Line offset for scrolling
	pageSize     int // Lines visible in viewport

	err error

	// Word count — recomputed lazily only after a modification.
	wordCount      int
	wordCountDirty bool

	// Search state set by the find bar; used by the view for highlighting.
	searchMatches      []int
	searchMatchLen     int
	searchCurrentMatch int
}

// New creates a new editor model.
func New(filepath string) *Model {
	content := ""
	if filepath != "" {
		if data, err := os.ReadFile(filepath); err == nil {
			content = string(data)
		}
	}

	buf := buffer.NewGapBuffer(content)
	cur := buffer.NewCursor(buf)

	return &Model{
		buffer:             buf,
		cursor:             cur,
		selection:          buffer.NewSelectionManager(cur),
		history:            history.NewHistory(),
		keyMap:             NewKeyMap(DefaultKeyBindings()),
		filepath:           filepath,
		modified:           false,
		width:              80,
		height:             24,
		pageSize:           20,
		wordCountDirty:     true,
		searchCurrentMatch: -1,
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
		m.pageSize = m.height - 2 // Reserve space for status bar
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := m.keyMap.LookupMsg(msg)

	switch action {
	case ActionQuit:
		return m, tea.Quit

	case ActionSave:
		m.save()
		return m, nil

	case ActionUndo:
		m.history.Undo()
		return m, nil

	case ActionRedo:
		m.history.Redo()
		return m, nil

	case ActionCut:
		m.cut()
		return m, nil

	case ActionCopy:
		m.copy()
		return m, nil

	case ActionPaste:
		m.paste()
		return m, nil

	case ActionSelectAll:
		m.selectAll()
		return m, nil

	case ActionMoveUp:
		m.selection.ClearSelection()
		m.cursor.MoveUp()
		m.ensureCursorVisible()
		return m, nil

	case ActionMoveDown:
		m.selection.ClearSelection()
		m.cursor.MoveDown()
		m.ensureCursorVisible()
		return m, nil

	case ActionMoveLeft:
		m.selection.ClearSelection()
		m.cursor.MoveLeft()
		m.ensureCursorVisible()
		return m, nil

	case ActionMoveRight:
		m.selection.ClearSelection()
		m.cursor.MoveRight()
		m.ensureCursorVisible()
		return m, nil

	case ActionMoveWordLeft:
		m.selection.ClearSelection()
		m.cursor.MoveWordLeft()
		m.ensureCursorVisible()
		return m, nil

	case ActionMoveWordRight:
		m.selection.ClearSelection()
		m.cursor.MoveWordRight()
		m.ensureCursorVisible()
		return m, nil

	case ActionMoveLineStart:
		m.selection.ClearSelection()
		m.cursor.MoveToLineStart()
		return m, nil

	case ActionMoveLineEnd:
		m.selection.ClearSelection()
		m.cursor.MoveToLineEnd()
		return m, nil

	case ActionMovePageUp:
		m.selection.ClearSelection()
		for i := 0; i < m.pageSize; i++ {
			m.cursor.MoveUp()
		}
		m.ensureCursorVisible()
		return m, nil

	case ActionMovePageDown:
		m.selection.ClearSelection()
		for i := 0; i < m.pageSize; i++ {
			m.cursor.MoveDown()
		}
		m.ensureCursorVisible()
		return m, nil

	case ActionSelectUp:
		m.selection.StartSelection()
		m.cursor.MoveUp()
		m.ensureCursorVisible()
		return m, nil

	case ActionSelectDown:
		m.selection.StartSelection()
		m.cursor.MoveDown()
		m.ensureCursorVisible()
		return m, nil

	case ActionSelectLeft:
		m.selection.StartSelection()
		m.cursor.MoveLeft()
		m.ensureCursorVisible()
		return m, nil

	case ActionSelectRight:
		m.selection.StartSelection()
		m.cursor.MoveRight()
		m.ensureCursorVisible()
		return m, nil

	case ActionInsertNewline:
		m.insertText("\n")
		return m, nil

	case ActionDeleteBackward:
		m.deleteBackward()
		return m, nil

	case ActionDeleteForward:
		m.deleteForward()
		return m, nil

	case ActionInsertTab:
		m.insertText("\t")
		return m, nil

	case ActionNone:
		// Handle regular character input
		if len(msg.Runes) > 0 {
			m.insertText(string(msg.Runes))
		}
		return m, nil
	}

	return m, nil
}

func (m *Model) insertText(text string) {
	// If there's a selection, delete it first
	if sel := m.selection.GetSelection(); sel != nil {
		cmd := history.NewReplaceCommand(m.buffer, m.cursor, sel.Start, sel.End-sel.Start, text)
		m.history.Execute(cmd)
		m.selection.ClearSelection()
	} else {
		cmd := history.NewInsertCommand(m.buffer, m.cursor, m.cursor.Position(), text)
		m.history.Execute(cmd)
	}
	m.modified = true
	m.wordCountDirty = true
	m.ensureCursorVisible()
}

func (m *Model) deleteBackward() {
	if sel := m.selection.GetSelection(); sel != nil {
		cmd := history.NewDeleteCommand(m.buffer, m.cursor, sel.Start, sel.End-sel.Start)
		m.history.Execute(cmd)
		m.selection.ClearSelection()
	} else if m.cursor.Position() > 0 {
		cmd := history.NewDeleteBackwardCommand(m.buffer, m.cursor, m.cursor.Position(), 1)
		m.history.Execute(cmd)
	}
	m.modified = true
	m.wordCountDirty = true
}

func (m *Model) deleteForward() {
	if sel := m.selection.GetSelection(); sel != nil {
		cmd := history.NewDeleteCommand(m.buffer, m.cursor, sel.Start, sel.End-sel.Start)
		m.history.Execute(cmd)
		m.selection.ClearSelection()
	} else if m.cursor.Position() < m.buffer.Length() {
		cmd := history.NewDeleteCommand(m.buffer, m.cursor, m.cursor.Position(), 1)
		m.history.Execute(cmd)
	}
	m.modified = true
	m.wordCountDirty = true
}

func (m *Model) cut() {
	if sel := m.selection.GetSelection(); sel != nil {
		text := m.selection.SelectedText(m.buffer)
		clipboard.WriteAll(text)
		cmd := history.NewDeleteCommand(m.buffer, m.cursor, sel.Start, sel.End-sel.Start)
		m.history.Execute(cmd)
		m.selection.ClearSelection()
		m.modified = true
	}
}

func (m *Model) copy() {
	if sel := m.selection.GetSelection(); sel != nil {
		text := m.selection.SelectedText(m.buffer)
		clipboard.WriteAll(text)
	}
}

func (m *Model) paste() {
	text, err := clipboard.ReadAll()
	if err != nil || text == "" {
		return
	}
	m.insertText(text)
}

func (m *Model) selectAll() {
	m.cursor.SetPosition(0)
	m.selection.StartSelection()
	m.cursor.SetPosition(m.buffer.Length())
}

func (m *Model) save() {
	if m.filepath == "" {
		return
	}

	content := m.buffer.String()
	err := os.WriteFile(m.filepath, []byte(content), 0644)
	if err != nil {
		m.err = err
		return
	}

	m.modified = false
	m.err = nil
}

func (m *Model) ensureCursorVisible() {
	line := m.cursor.Line()

	if line < m.scrollOffset {
		m.scrollOffset = line
	} else if line >= m.scrollOffset+m.pageSize {
		m.scrollOffset = line - m.pageSize + 1
	}
}

// Accessors for use by view and other components

func (m *Model) Buffer() *buffer.GapBuffer {
	return m.buffer
}

func (m *Model) Cursor() *buffer.Cursor {
	return m.cursor
}

func (m *Model) Selection() *buffer.SelectionManager {
	return m.selection
}

func (m *Model) Filepath() string {
	return m.filepath
}

func (m *Model) SetFilepath(path string) {
	m.filepath = path
}

func (m *Model) Modified() bool {
	return m.modified
}

func (m *Model) ScrollOffset() int {
	return m.scrollOffset
}

func (m *Model) PageSize() int {
	return m.pageSize
}

func (m *Model) Width() int {
	return m.width
}

func (m *Model) Height() int {
	return m.height
}

func (m *Model) Error() error {
	return m.err
}

func (m *Model) Content() string {
	return m.buffer.String()
}

// LoadFile replaces the editor content with the contents of path.
// If path is a new (non-existent) file the editor starts empty.
func (m *Model) LoadFile(path string) {
	content := ""
	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			content = string(data)
		}
	}
	m.filepath = path
	m.buffer = buffer.NewGapBuffer(content)
	m.cursor = buffer.NewCursor(m.buffer)
	m.selection = buffer.NewSelectionManager(m.cursor)
	m.history = history.NewHistory()
	m.modified = false
	m.wordCountDirty = true
	m.err = nil
	m.scrollOffset = 0
	m.ClearSearchState()
}

// WordCount returns the current word count, recomputing only when the buffer
// has changed since the last call.
func (m *Model) WordCount() int {
	if m.wordCountDirty {
		m.wordCount = len(strings.Fields(m.buffer.String()))
		m.wordCountDirty = false
	}
	return m.wordCount
}

// SetSearchState informs the view which buffer positions to highlight as
// search matches. matches is a sorted slice of rune positions; current is
// the index of the active match (-1 for none).
func (m *Model) SetSearchState(matches []int, matchLen, current int) {
	m.searchMatches = matches
	m.searchMatchLen = matchLen
	m.searchCurrentMatch = current
}

// ClearSearchState removes all search highlights.
func (m *Model) ClearSearchState() {
	m.searchMatches = nil
	m.searchMatchLen = 0
	m.searchCurrentMatch = -1
}

// JumpToPosition moves the cursor to the given buffer position and scrolls
// the viewport so it is visible.
func (m *Model) JumpToPosition(pos int) {
	m.cursor.SetPosition(pos)
	m.ensureCursorVisible()
}

// ReplaceAt replaces matchLen runes at pos with text, using the undo history.
func (m *Model) ReplaceAt(pos, matchLen int, text string) {
	cmd := history.NewReplaceCommand(m.buffer, m.cursor, pos, matchLen, text)
	m.history.Execute(cmd)
	m.modified = true
	m.wordCountDirty = true
}

// Save writes the current buffer to disk and returns any error.
func (m *Model) Save() error {
	m.save()
	return m.err
}
