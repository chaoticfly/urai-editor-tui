package history

import "urai/internal/buffer"

// Command represents an undoable/redoable operation.
type Command interface {
	Execute() error
	Undo() error
	Description() string
}

// InsertCommand represents an insert operation.
type InsertCommand struct {
	buf      *buffer.GapBuffer
	cursor   *buffer.Cursor
	pos      int
	text     string
	oldPos   int
}

// NewInsertCommand creates a new insert command.
func NewInsertCommand(buf *buffer.GapBuffer, cursor *buffer.Cursor, pos int, text string) *InsertCommand {
	return &InsertCommand{
		buf:    buf,
		cursor: cursor,
		pos:    pos,
		text:   text,
		oldPos: cursor.Position(),
	}
}

func (c *InsertCommand) Execute() error {
	c.buf.InsertString(c.pos, c.text)
	c.cursor.SetPosition(c.pos + len([]rune(c.text)))
	return nil
}

func (c *InsertCommand) Undo() error {
	c.buf.Delete(c.pos, len([]rune(c.text)))
	c.cursor.SetPosition(c.oldPos)
	return nil
}

func (c *InsertCommand) Description() string {
	return "Insert"
}

// DeleteCommand represents a delete operation.
type DeleteCommand struct {
	buf         *buffer.GapBuffer
	cursor      *buffer.Cursor
	pos         int
	count       int
	deletedText string
	oldPos      int
}

// NewDeleteCommand creates a new delete command.
func NewDeleteCommand(buf *buffer.GapBuffer, cursor *buffer.Cursor, pos int, count int) *DeleteCommand {
	return &DeleteCommand{
		buf:    buf,
		cursor: cursor,
		pos:    pos,
		count:  count,
		oldPos: cursor.Position(),
	}
}

func (c *DeleteCommand) Execute() error {
	c.deletedText = c.buf.Delete(c.pos, c.count)
	c.cursor.SetPosition(c.pos)
	return nil
}

func (c *DeleteCommand) Undo() error {
	c.buf.InsertString(c.pos, c.deletedText)
	c.cursor.SetPosition(c.oldPos)
	return nil
}

func (c *DeleteCommand) Description() string {
	return "Delete"
}

// DeleteBackwardCommand represents a backward delete (backspace) operation.
type DeleteBackwardCommand struct {
	buf         *buffer.GapBuffer
	cursor      *buffer.Cursor
	pos         int
	count       int
	deletedText string
}

// NewDeleteBackwardCommand creates a new backward delete command.
func NewDeleteBackwardCommand(buf *buffer.GapBuffer, cursor *buffer.Cursor, pos int, count int) *DeleteBackwardCommand {
	return &DeleteBackwardCommand{
		buf:    buf,
		cursor: cursor,
		pos:    pos,
		count:  count,
	}
}

func (c *DeleteBackwardCommand) Execute() error {
	start := c.pos - c.count
	if start < 0 {
		c.count = c.pos
		start = 0
	}
	c.deletedText = c.buf.Delete(start, c.count)
	c.cursor.SetPosition(start)
	return nil
}

func (c *DeleteBackwardCommand) Undo() error {
	start := c.pos - c.count
	if start < 0 {
		start = 0
	}
	c.buf.InsertString(start, c.deletedText)
	c.cursor.SetPosition(c.pos)
	return nil
}

func (c *DeleteBackwardCommand) Description() string {
	return "DeleteBackward"
}

// ReplaceCommand represents a replace operation (for selections).
type ReplaceCommand struct {
	buf         *buffer.GapBuffer
	cursor      *buffer.Cursor
	pos         int
	deleteCount int
	newText     string
	deletedText string
	oldPos      int
}

// NewReplaceCommand creates a new replace command.
func NewReplaceCommand(buf *buffer.GapBuffer, cursor *buffer.Cursor, pos int, deleteCount int, newText string) *ReplaceCommand {
	return &ReplaceCommand{
		buf:         buf,
		cursor:      cursor,
		pos:         pos,
		deleteCount: deleteCount,
		newText:     newText,
		oldPos:      cursor.Position(),
	}
}

func (c *ReplaceCommand) Execute() error {
	c.deletedText = c.buf.Delete(c.pos, c.deleteCount)
	c.buf.InsertString(c.pos, c.newText)
	c.cursor.SetPosition(c.pos + len([]rune(c.newText)))
	return nil
}

func (c *ReplaceCommand) Undo() error {
	c.buf.Delete(c.pos, len([]rune(c.newText)))
	c.buf.InsertString(c.pos, c.deletedText)
	c.cursor.SetPosition(c.oldPos)
	return nil
}

func (c *ReplaceCommand) Description() string {
	return "Replace"
}
