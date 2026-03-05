package history

const maxHistorySize = 1000

// History manages undo/redo stacks.
type History struct {
	undoStack []Command
	redoStack []Command
}

// NewHistory creates a new history manager.
func NewHistory() *History {
	return &History{
		undoStack: make([]Command, 0),
		redoStack: make([]Command, 0),
	}
}

// Execute runs a command and adds it to the undo stack.
func (h *History) Execute(cmd Command) error {
	if err := cmd.Execute(); err != nil {
		return err
	}

	h.undoStack = append(h.undoStack, cmd)

	// Limit undo stack size
	if len(h.undoStack) > maxHistorySize {
		h.undoStack = h.undoStack[1:]
	}

	// Clear redo stack when new command is executed
	h.redoStack = h.redoStack[:0]

	return nil
}

// Undo undoes the last command.
func (h *History) Undo() error {
	if len(h.undoStack) == 0 {
		return nil
	}

	cmd := h.undoStack[len(h.undoStack)-1]
	h.undoStack = h.undoStack[:len(h.undoStack)-1]

	if err := cmd.Undo(); err != nil {
		return err
	}

	h.redoStack = append(h.redoStack, cmd)
	return nil
}

// Redo redoes the last undone command.
func (h *History) Redo() error {
	if len(h.redoStack) == 0 {
		return nil
	}

	cmd := h.redoStack[len(h.redoStack)-1]
	h.redoStack = h.redoStack[:len(h.redoStack)-1]

	if err := cmd.Execute(); err != nil {
		return err
	}

	h.undoStack = append(h.undoStack, cmd)
	return nil
}

// CanUndo returns whether there are commands to undo.
func (h *History) CanUndo() bool {
	return len(h.undoStack) > 0
}

// CanRedo returns whether there are commands to redo.
func (h *History) CanRedo() bool {
	return len(h.redoStack) > 0
}

// Clear clears both undo and redo stacks.
func (h *History) Clear() {
	h.undoStack = h.undoStack[:0]
	h.redoStack = h.redoStack[:0]
}

// UndoCount returns the number of undoable commands.
func (h *History) UndoCount() int {
	return len(h.undoStack)
}

// RedoCount returns the number of redoable commands.
func (h *History) RedoCount() int {
	return len(h.redoStack)
}
