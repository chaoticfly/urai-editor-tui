package editor

import "github.com/charmbracelet/bubbletea"

// Key constants for common key sequences
const (
	KeyCtrlA = "ctrl+a"
	KeyCtrlC = "ctrl+c"
	KeyCtrlE = "ctrl+e"
	KeyCtrlK = "ctrl+k"
	KeyCtrlQ = "ctrl+q"
	KeyCtrlS = "ctrl+s"
	KeyCtrlV = "ctrl+v"
	KeyCtrlX = "ctrl+x"
	KeyCtrlY = "ctrl+y"
	KeyCtrlZ = "ctrl+z"

	KeyCtrlBackslash = "ctrl+\\"

	KeyAltLeft  = "alt+left"
	KeyAltRight = "alt+right"

	KeyShiftUp    = "shift+up"
	KeyShiftDown  = "shift+down"
	KeyShiftLeft  = "shift+left"
	KeyShiftRight = "shift+right"

	KeyUp     = "up"
	KeyDown   = "down"
	KeyLeft   = "left"
	KeyRight  = "right"
	KeyHome   = "home"
	KeyEnd    = "end"
	KeyPgUp   = "pgup"
	KeyPgDown = "pgdown"

	KeyEnter     = "enter"
	KeyBackspace = "backspace"
	KeyDelete    = "delete"
	KeyTab       = "tab"
	KeyEsc       = "esc"
)

// KeyBinding represents a single key binding.
type KeyBinding struct {
	Key         string
	Description string
	Action      Action
}

// Action represents the action to perform for a key binding.
type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionSave
	ActionUndo
	ActionRedo
	ActionCut
	ActionCopy
	ActionPaste
	ActionSelectAll
	ActionCycleView
	ActionCommandPalette
	ActionMoveUp
	ActionMoveDown
	ActionMoveLeft
	ActionMoveRight
	ActionMoveWordLeft
	ActionMoveWordRight
	ActionMoveLineStart
	ActionMoveLineEnd
	ActionMovePageUp
	ActionMovePageDown
	ActionSelectUp
	ActionSelectDown
	ActionSelectLeft
	ActionSelectRight
	ActionInsertNewline
	ActionDeleteBackward
	ActionDeleteForward
	ActionInsertTab
)

// DefaultKeyBindings returns the default key bindings.
func DefaultKeyBindings() []KeyBinding {
	return []KeyBinding{
		{KeyCtrlQ, "Quit", ActionQuit},
		{KeyCtrlS, "Save", ActionSave},
		{KeyCtrlZ, "Undo", ActionUndo},
		{KeyCtrlY, "Redo", ActionRedo},
		{KeyCtrlX, "Cut", ActionCut},
		{KeyCtrlC, "Copy", ActionCopy},
		{KeyCtrlV, "Paste", ActionPaste},
		{KeyCtrlA, "Select All", ActionSelectAll},
		{KeyCtrlBackslash, "Cycle View", ActionCycleView},
		{KeyCtrlK, "Command Palette", ActionCommandPalette},
		{KeyUp, "Move Up", ActionMoveUp},
		{KeyDown, "Move Down", ActionMoveDown},
		{KeyLeft, "Move Left", ActionMoveLeft},
		{KeyRight, "Move Right", ActionMoveRight},
		{KeyAltLeft, "Move Word Left", ActionMoveWordLeft},
		{KeyAltRight, "Move Word Right", ActionMoveWordRight},
		{KeyHome, "Line Start", ActionMoveLineStart},
		{KeyEnd, "Line End", ActionMoveLineEnd},
		{KeyPgUp, "Page Up", ActionMovePageUp},
		{KeyPgDown, "Page Down", ActionMovePageDown},
		{KeyShiftUp, "Select Up", ActionSelectUp},
		{KeyShiftDown, "Select Down", ActionSelectDown},
		{KeyShiftLeft, "Select Left", ActionSelectLeft},
		{KeyShiftRight, "Select Right", ActionSelectRight},
		{KeyEnter, "New Line", ActionInsertNewline},
		{KeyBackspace, "Delete Backward", ActionDeleteBackward},
		{KeyDelete, "Delete Forward", ActionDeleteForward},
		{KeyTab, "Insert Tab", ActionInsertTab},
	}
}

// KeyMap maps key strings to actions.
type KeyMap struct {
	bindings map[string]Action
}

// NewKeyMap creates a new key map from bindings.
func NewKeyMap(bindings []KeyBinding) *KeyMap {
	km := &KeyMap{
		bindings: make(map[string]Action),
	}
	for _, b := range bindings {
		km.bindings[b.Key] = b.Action
	}
	return km
}

// Lookup returns the action for a given key.
func (km *KeyMap) Lookup(key string) Action {
	if action, ok := km.bindings[key]; ok {
		return action
	}
	return ActionNone
}

// LookupMsg returns the action for a tea.KeyMsg.
func (km *KeyMap) LookupMsg(msg tea.KeyMsg) Action {
	return km.Lookup(msg.String())
}
