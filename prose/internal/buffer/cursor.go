package buffer

// Cursor tracks cursor position in the buffer.
type Cursor struct {
	buf       *GapBuffer
	pos       int // Absolute position in buffer
	preferCol int // Preferred column for vertical movement
}

// NewCursor creates a new cursor for the given buffer.
func NewCursor(buf *GapBuffer) *Cursor {
	return &Cursor{
		buf:       buf,
		pos:       0,
		preferCol: 0,
	}
}

// Position returns the current cursor position.
func (c *Cursor) Position() int {
	return c.pos
}

// SetPosition sets the cursor to the given position.
func (c *Cursor) SetPosition(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > c.buf.Length() {
		pos = c.buf.Length()
	}
	c.pos = pos
	_, c.preferCol = c.buf.PositionToLineCol(c.pos)
}

// Line returns the current line (0-indexed).
func (c *Cursor) Line() int {
	line, _ := c.buf.PositionToLineCol(c.pos)
	return line
}

// Col returns the current column (0-indexed).
func (c *Cursor) Col() int {
	_, col := c.buf.PositionToLineCol(c.pos)
	return col
}

// LineCol returns the current line and column.
func (c *Cursor) LineCol() (int, int) {
	return c.buf.PositionToLineCol(c.pos)
}

// MoveLeft moves the cursor left by one character.
func (c *Cursor) MoveLeft() {
	if c.pos > 0 {
		c.pos--
		_, c.preferCol = c.buf.PositionToLineCol(c.pos)
	}
}

// MoveRight moves the cursor right by one character.
func (c *Cursor) MoveRight() {
	if c.pos < c.buf.Length() {
		c.pos++
		_, c.preferCol = c.buf.PositionToLineCol(c.pos)
	}
}

// MoveUp moves the cursor up one line, maintaining column preference.
func (c *Cursor) MoveUp() {
	line, _ := c.buf.PositionToLineCol(c.pos)
	if line > 0 {
		c.pos = c.buf.LineColToPosition(line-1, c.preferCol)
	}
}

// MoveDown moves the cursor down one line, maintaining column preference.
func (c *Cursor) MoveDown() {
	line, _ := c.buf.PositionToLineCol(c.pos)
	if line < c.buf.LineCount()-1 {
		c.pos = c.buf.LineColToPosition(line+1, c.preferCol)
	}
}

// MoveToLineStart moves the cursor to the start of the current line.
func (c *Cursor) MoveToLineStart() {
	line, _ := c.buf.PositionToLineCol(c.pos)
	c.pos = c.buf.LineStart(line)
	c.preferCol = 0
}

// MoveToLineEnd moves the cursor to the end of the current line.
func (c *Cursor) MoveToLineEnd() {
	line, _ := c.buf.PositionToLineCol(c.pos)
	c.pos = c.buf.LineEnd(line)
	_, c.preferCol = c.buf.PositionToLineCol(c.pos)
}

// MoveToStart moves the cursor to the start of the buffer.
func (c *Cursor) MoveToStart() {
	c.pos = 0
	c.preferCol = 0
}

// MoveToEnd moves the cursor to the end of the buffer.
func (c *Cursor) MoveToEnd() {
	c.pos = c.buf.Length()
	_, c.preferCol = c.buf.PositionToLineCol(c.pos)
}

// MoveWordLeft moves the cursor to the start of the previous word.
func (c *Cursor) MoveWordLeft() {
	if c.pos == 0 {
		return
	}

	// Skip any whitespace before cursor
	for c.pos > 0 && isWhitespace(c.buf.At(c.pos-1)) {
		c.pos--
	}

	// Skip word characters
	for c.pos > 0 && !isWhitespace(c.buf.At(c.pos-1)) {
		c.pos--
	}

	_, c.preferCol = c.buf.PositionToLineCol(c.pos)
}

// MoveWordRight moves the cursor to the start of the next word.
func (c *Cursor) MoveWordRight() {
	length := c.buf.Length()
	if c.pos >= length {
		return
	}

	// Skip current word
	for c.pos < length && !isWhitespace(c.buf.At(c.pos)) {
		c.pos++
	}

	// Skip whitespace
	for c.pos < length && isWhitespace(c.buf.At(c.pos)) {
		c.pos++
	}

	_, c.preferCol = c.buf.PositionToLineCol(c.pos)
}

// isWhitespace returns true if r is a whitespace character.
func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// Selection represents a text selection range.
type Selection struct {
	Start int
	End   int
}

// SelectionManager manages text selection state.
type SelectionManager struct {
	cursor    *Cursor
	anchor    int  // Selection anchor position
	selecting bool // Whether selection is active
}

// NewSelectionManager creates a new selection manager.
func NewSelectionManager(cursor *Cursor) *SelectionManager {
	return &SelectionManager{
		cursor:    cursor,
		anchor:    0,
		selecting: false,
	}
}

// StartSelection begins a selection at the current cursor position.
func (sm *SelectionManager) StartSelection() {
	if !sm.selecting {
		sm.anchor = sm.cursor.Position()
		sm.selecting = true
	}
}

// ClearSelection clears the current selection.
func (sm *SelectionManager) ClearSelection() {
	sm.selecting = false
}

// IsSelecting returns whether a selection is active.
func (sm *SelectionManager) IsSelecting() bool {
	return sm.selecting
}

// GetSelection returns the current selection, if any.
func (sm *SelectionManager) GetSelection() *Selection {
	if !sm.selecting {
		return nil
	}

	start := sm.anchor
	end := sm.cursor.Position()

	if start > end {
		start, end = end, start
	}

	return &Selection{Start: start, End: end}
}

// SelectedText returns the currently selected text.
func (sm *SelectionManager) SelectedText(buf *GapBuffer) string {
	sel := sm.GetSelection()
	if sel == nil {
		return ""
	}
	return buf.Substring(sel.Start, sel.End)
}
