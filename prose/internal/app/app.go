package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"urai/internal/config"
	"urai/internal/editor"
	"urai/internal/export"
	"urai/internal/filebrowser"
	"urai/internal/fountain"
	"urai/internal/layout"
	"urai/internal/preview"
	"urai/internal/settings"
	"urai/internal/suggestions"
)

// Model is the root application model.
type Model struct {
	editor      *editor.Model
	preview     *preview.Model
	aiPanel     *suggestions.Model
	settingsPanel *settings.Model
	layout      *layout.SplitLayout
	statusBar   *layout.StatusBar

	viewMode     layout.ViewMode
	prevViewMode layout.ViewMode
	config       *config.Config

	width  int
	height int

	message        string
	showHelp       bool
	showSettings   bool
	showFileBrowser bool
	fileBrowser    *filebrowser.Model
	quitting       bool
	isFountain     bool
	fountainPopup  *fountain.Popup
}

// New creates a new application model.
func New(fp string, cfg *config.Config) *Model {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	ed := editor.New(fp)
	prev := preview.New(80, 24)
	split := layout.NewSplitLayout(80, 24)
	status := layout.NewStatusBar(80)
	aiP := suggestions.New(cfg.AI, 40, 24)
	settingsP := settings.New(*cfg, 80, 24)

	// Auto-start in split mode for markdown files
	initialMode := layout.ViewModeEdit
	if strings.ToLower(filepath.Ext(fp)) == ".md" {
		initialMode = layout.ViewModeSplit
	}

	m := &Model{
		editor:        ed,
		preview:       prev,
		aiPanel:       aiP,
		settingsPanel: settingsP,
		layout:        split,
		statusBar:     status,
		viewMode:      initialMode,
		config:        cfg,
		width:         80,
		height:        24,
		showHelp:      true,
		isFountain:    fountain.IsFountainFile(fp),
		fountainPopup: &fountain.Popup{},
	}

	if initialMode == layout.ViewModeSplit {
		m.preview.SetContentImmediate(ed.Content())
	}

	return m
}

// Init initializes the application.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.editor.Init(),
		tea.EnterAltScreen,
	)
}

// Update handles messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.layout.SetSize(m.width, m.height)
		m.statusBar.SetWidth(m.width)
		m.settingsPanel.SetSize(m.width, m.height)
		if m.fileBrowser != nil {
			m.fileBrowser.SetSize(m.width, m.height)
		}

		editorMsg := tea.WindowSizeMsg{
			Width:  m.editorWidth(),
			Height: m.editorHeight(),
		}
		m.editor.Update(editorMsg)
		m.preview.SetSize(m.previewWidth(), m.editorHeight())
		m.aiPanel.SetSize(m.aiPanelWidth(), m.editorHeight())
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case suggestions.SuggestionMsg:
		m.aiPanel.HandleSuggestion(msg)
		if msg.Error != nil {
			m.message = fmt.Sprintf("AI error: %v", msg.Error)
		} else {
			m.message = "AI suggestion ready — Ctrl+L to insert"
		}
		return m, nil

	case suggestions.InsertMsg:
		m.editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(msg.Text)})
		m.message = "Suggestion inserted"
		return m, nil

	case settings.SaveMsg:
		*m.config = msg.Config
		m.aiPanel.UpdateConfig(msg.Config.AI)
		m.showSettings = false
		if err := m.config.Save(); err != nil {
			m.message = fmt.Sprintf("Settings save error: %v", err)
		} else {
			m.message = "Settings saved"
		}
		return m, nil

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case settings.CloseMsg:
		m.showSettings = false
		return m, nil

	case filebrowser.OpenFileMsg:
		m.openFile(msg.Path)
		m.showFileBrowser = false
		return m, nil

	case filebrowser.SaveAsMsg:
		m.editor.SetFilepath(msg.Path)
		if err := m.editor.Save(); err != nil {
			m.message = fmt.Sprintf("Save error: %v", err)
		} else {
			m.message = fmt.Sprintf("Saved: %s", filepath.Base(msg.Path))
			m.isFountain = fountain.IsFountainFile(msg.Path)
		}
		m.showFileBrowser = false
		return m, nil

	case filebrowser.CloseMsg:
		m.showFileBrowser = false
		return m, nil
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Settings modal captures all input
	if m.showSettings {
		updated, cmd := m.settingsPanel.Update(msg)
		m.settingsPanel = updated
		return m, cmd
	}

	// File browser captures all input
	if m.showFileBrowser {
		updated, cmd := m.fileBrowser.Update(msg)
		m.fileBrowser = updated
		return m, cmd
	}

	// Help overlay: any key dismisses
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	// Esc exits focus mode
	if key == "esc" && m.viewMode == layout.ViewModeFocus {
		m.viewMode = m.prevViewMode
		m.syncEditorDimensions()
		m.message = ""
		return m, nil
	}

	// Global keybindings
	switch key {
	case "ctrl+q":
		m.quitting = true
		return m, tea.Quit

	case "ctrl+\\":
		if m.viewMode != layout.ViewModeFocus {
			m.viewMode = layout.CycleViewMode(m.viewMode)
			m.message = fmt.Sprintf("View: %s", layout.ViewModeName(m.viewMode))
			m.syncEditorDimensions()
			m.updatePreviewContent()
			if m.viewMode == layout.ViewModeAI {
				line, _ := m.editor.Cursor().LineCol()
				m.aiPanel.SetEditorContext(m.editor.Content(), line)
			}
		}
		return m, nil

	case "ctrl+f":
		if m.viewMode == layout.ViewModeFocus {
			m.viewMode = m.prevViewMode
		} else {
			m.prevViewMode = m.viewMode
			m.viewMode = layout.ViewModeFocus
			m.message = ""
		}
		m.syncEditorDimensions()
		return m, nil

	case "ctrl+s":
		if m.editor.Filepath() == "" {
			// No filepath — open file browser in save-as mode
			startDir, _ := os.Getwd()
			m.fileBrowser = filebrowser.NewSaveAs(startDir, m.width, m.height)
			m.showFileBrowser = true
			return m, nil
		}
		m.editor.Update(msg)
		if m.editor.Error() != nil {
			m.message = fmt.Sprintf("Error: %v", m.editor.Error())
		} else {
			m.message = "Saved"
		}
		return m, nil

	case "ctrl+e":
		return m.exportHTML()

	case "ctrl+p":
		return m.exportPDF()

	case "ctrl+?", "ctrl+h":
		m.showHelp = true
		return m, nil

	case "f2":
		m.showSettings = true
		m.settingsPanel = settings.New(*m.config, m.width, m.height)
		return m, nil

	case "f3":
		startDir := currentDir(m.editor.Filepath())
		m.fileBrowser = filebrowser.New(startDir, m.width, m.height)
		m.showFileBrowser = true
		return m, nil

	case "ctrl+r":
		// Request AI suggestion — do nothing if the document is empty
		if strings.TrimSpace(m.editor.Content()) == "" {
			m.message = "Nothing to generate from — add some text first"
			return m, nil
		}
		if m.viewMode != layout.ViewModeAI {
			m.viewMode = layout.ViewModeAI
			m.syncEditorDimensions()
		}
		line, _ := m.editor.Cursor().LineCol()
		m.aiPanel.SetEditorContext(m.editor.Content(), line)
		cmd := m.aiPanel.RequestSuggestion()
		if cmd != nil {
			m.message = "Generating suggestion..."
		}
		return m, cmd

	case "ctrl+l":
		// Insert AI suggestion at cursor
		cmd := m.aiPanel.InsertSuggestion()
		if cmd == nil {
			m.message = "No suggestion available — press Ctrl+R first"
		}
		return m, cmd

	case "f4":
		// Toggle AI suggestion mode (line / paragraph)
		m.aiPanel.ToggleMode()
		line, _ := m.editor.Cursor().LineCol()
		m.aiPanel.SetEditorContext(m.editor.Content(), line)
		m.message = "AI mode toggled"
		return m, nil
	}

	// Preview-only mode: scroll only
	if m.viewMode == layout.ViewModePreview {
		switch key {
		case "up", "k":
			m.preview.ScrollUp()
		case "down", "j":
			m.preview.ScrollDown()
		}
		return m, nil
	}

	// AI panel: scroll suggestion text, rest goes to editor
	if m.viewMode == layout.ViewModeAI {
		switch key {
		case "up", "k":
			m.aiPanel.ScrollUp()
			return m, nil
		case "down", "j":
			m.aiPanel.ScrollDown()
			return m, nil
		}
	}

	// Fountain autocomplete intercepts Tab (accept) and Esc (dismiss)
	if m.isFountain && m.fountainPopup.Visible() {
		switch key {
		case "tab":
			suffix := m.fountainPopup.Accept()
			m.fountainPopup.Dismiss()
			if suffix != "" {
				updatedEditor, _ := m.editor.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(suffix)})
				m.editor = updatedEditor.(*editor.Model)
				if m.viewMode == layout.ViewModeSplit {
					m.updatePreviewContent()
				}
			}
			return m, nil
		case "up":
			m.fountainPopup.SelectPrev()
			return m, nil
		case "down":
			m.fountainPopup.SelectNext()
			return m, nil
		case "esc":
			m.fountainPopup.Dismiss()
			return m, nil
		}
	}

	// Forward to editor
	updatedEditor, cmd := m.editor.Update(msg)
	m.editor = updatedEditor.(*editor.Model)

	if m.viewMode == layout.ViewModeSplit {
		m.updatePreviewContent()
	}

	if m.viewMode == layout.ViewModeAI {
		line, _ := m.editor.Cursor().LineCol()
		m.aiPanel.SetEditorContext(m.editor.Content(), line)
	}

	// Update fountain completions after each edit
	if m.isFountain && m.viewMode != layout.ViewModeFocus {
		m.updateFountainPopup()
	}

	m.message = ""
	return m, cmd
}

func (m *Model) updateFountainPopup() {
	lines := strings.Split(m.editor.Content(), "\n")
	curLine, curCol := m.editor.Cursor().LineCol()
	var prefix string
	if curLine < len(lines) {
		runes := []rune(lines[curLine])
		if curCol <= len(runes) {
			prefix = string(runes[:curCol])
		} else {
			prefix = lines[curLine]
		}
	}
	wasVisible := m.fountainPopup.Visible()
	m.fountainPopup.Update(prefix)
	if wasVisible != m.fountainPopup.Visible() {
		m.syncEditorDimensions()
	}
}

func (m *Model) updatePreviewContent() {
	if m.viewMode != layout.ViewModeEdit && m.viewMode != layout.ViewModeFocus && m.viewMode != layout.ViewModeAI {
		m.preview.SetContentImmediate(m.editor.Content())
	}
}

func (m *Model) syncEditorDimensions() {
	editorMsg := tea.WindowSizeMsg{
		Width:  m.editorWidth(),
		Height: m.editorHeight(),
	}
	m.editor.Update(editorMsg)
	m.preview.SetSize(m.previewWidth(), m.editorHeight())
	m.aiPanel.SetSize(m.aiPanelWidth(), m.editorHeight())
}

func (m *Model) exportHTML() (tea.Model, tea.Cmd) {
	fp := m.editor.Filepath()
	if fp == "" {
		m.message = "Save file first"
		return m, nil
	}

	outputPath := strings.TrimSuffix(fp, ".md") + ".html"
	exporter := export.NewHTMLExporter()

	if err := exporter.Export(m.editor.Content(), outputPath); err != nil {
		m.message = fmt.Sprintf("Export failed: %v", err)
	} else {
		m.message = fmt.Sprintf("Exported: %s", fp)
	}

	return m, nil
}

func (m *Model) exportPDF() (tea.Model, tea.Cmd) {
	fp := m.editor.Filepath()
	if fp == "" {
		m.message = "Save file first"
		return m, nil
	}

	outputPath := strings.TrimSuffix(fp, ".md") + ".pdf"
	exporter := export.NewPDFExporter()

	if err := exporter.Export(m.editor.Content(), outputPath); err != nil {
		m.message = fmt.Sprintf("PDF export failed: %v", err)
	} else {
		m.message = fmt.Sprintf("Exported: %s", filepath.Base(outputPath))
	}

	return m, nil
}

func (m *Model) focusWidth() int {
	w := 80
	if m.width < w+20 {
		w = m.width - 10
	}
	if w < 20 {
		w = m.width
	}
	return w
}

func (m *Model) editorWidth() int {
	switch m.viewMode {
	case layout.ViewModeSplit, layout.ViewModeAI:
		return m.layout.LeftWidth()
	case layout.ViewModeFocus:
		return m.focusWidth()
	default:
		return m.width
	}
}

func (m *Model) previewWidth() int {
	if m.viewMode == layout.ViewModeSplit {
		return m.layout.RightWidth()
	}
	return m.width
}

func (m *Model) aiPanelWidth() int {
	if m.viewMode == layout.ViewModeAI {
		return m.layout.RightWidth()
	}
	return m.width
}

func (m *Model) editorHeight() int {
	if m.viewMode == layout.ViewModeFocus {
		return m.height
	}
	h := m.height - 1 // status bar
	if m.isFountain {
		h -= m.fountainPopup.Height()
	}
	return h
}

// View renders the application.
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	if m.showSettings {
		return m.settingsPanel.View()
	}

	if m.showFileBrowser {
		return m.fileBrowser.View()
	}

	if m.showHelp {
		return m.renderHelp()
	}

	var content string

	switch m.viewMode {
	case layout.ViewModeEdit:
		content = m.layout.RenderSingle(m.editor.View())

	case layout.ViewModeSplit:
		content = m.layout.Render(m.editor.View(), m.preview.View())

	case layout.ViewModePreview:
		content = m.layout.RenderSingle(m.preview.View())

	case layout.ViewModeAI:
		content = m.layout.Render(m.editor.View(), m.aiPanel.View())

	case layout.ViewModeFocus:
		content = m.layout.RenderFocus(m.editor.ViewFocus(), m.focusWidth(), m.editorHeight())
		return content
	}

	line, col := m.editor.Cursor().LineCol()
	statusInfo := layout.StatusInfo{
		Filename: m.editor.Filepath(),
		Modified: m.editor.Modified(),
		Line:     line,
		Col:      col,
		ViewMode: m.viewMode,
		Message:  m.message,
	}
	statusBar := m.statusBar.Render(statusInfo)

	parts := []string{content}
	if m.isFountain && m.fountainPopup.Visible() {
		parts = append(parts, m.fountainPopup.View(m.width))
	}
	parts = append(parts, statusBar)
	result := lipgloss.JoinVertical(lipgloss.Left, parts...)

	if m.config.BackgroundColor != "" {
		result = lipgloss.NewStyle().
			Background(lipgloss.Color(m.config.BackgroundColor)).
			Width(m.width).
			Render(result)
	}

	return result
}

// openFile loads a file into the editor and updates related state.
func (m *Model) openFile(path string) {
	m.editor.LoadFile(path)
	m.isFountain = fountain.IsFountainFile(path)
	m.fountainPopup = &fountain.Popup{}
	m.message = fmt.Sprintf("Opened: %s", filepath.Base(path))

	// Auto-switch view mode based on file type
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".md" {
		m.viewMode = layout.ViewModeSplit
	} else {
		m.viewMode = layout.ViewModeEdit
	}
	m.syncEditorDimensions()
	m.updatePreviewContent()
}

// currentDir returns the directory to start the file browser in.
func currentDir(editorPath string) string {
	if editorPath != "" {
		return filepath.Dir(editorPath)
	}
	dir, _ := os.Getwd()
	return dir
}

func (m *Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		switch m.viewMode {
		case layout.ViewModePreview:
			m.preview.ScrollUp()
		case layout.ViewModeSplit:
			m.preview.ScrollUp()
		case layout.ViewModeAI:
			m.aiPanel.ScrollUp()
		}
	case tea.MouseButtonWheelDown:
		switch m.viewMode {
		case layout.ViewModePreview:
			m.preview.ScrollDown()
		case layout.ViewModeSplit:
			m.preview.ScrollDown()
		case layout.ViewModeAI:
			m.aiPanel.ScrollDown()
		}
	}
	return m, nil
}

// renderHelp renders the keyboard shortcuts overlay.
func (m *Model) renderHelp() string {
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("4")).
		Padding(1, 4)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12")).
		MarginBottom(1)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("3")).
		MarginTop(1)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("7")).
		Width(18)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15"))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	type shortcut struct{ key, desc string }
	type section struct {
		name    string
		entries []shortcut
	}

	sections := []section{
		{"File", []shortcut{
			{"Ctrl+S", "Save"},
			{"Ctrl+Q", "Quit"},
			{"Ctrl+E", "Export HTML"},
			{"Ctrl+P", "Export PDF"},
		}},
		{"Edit", []shortcut{
			{"Ctrl+Z", "Undo"},
			{"Ctrl+Y", "Redo"},
			{"Ctrl+X", "Cut"},
			{"Ctrl+C", "Copy"},
			{"Ctrl+V", "Paste"},
			{"Ctrl+A", "Select All"},
		}},
		{"Navigation", []shortcut{
			{"Arrow keys", "Move cursor"},
			{"Alt+← / Alt+→", "Move by word"},
			{"Home / End", "Line start / end"},
			{"PgUp / PgDn", "Page up / down"},
			{"Shift+↑↓←→", "Select text"},
		}},
		{"View", []shortcut{
			{"Ctrl+\\", "Cycle: Edit/Split/Preview/AI"},
			{"Ctrl+F", "Toggle focus mode"},
			{"Esc", "Exit focus mode"},
			{"F2", "Settings"},
			{"F3", "File browser"},
		}},
		{"AI Suggestions", []shortcut{
			{"Ctrl+R", "Request AI suggestion"},
			{"Ctrl+L", "Insert suggestion at cursor"},
			{"F4", "Toggle Line/Paragraph mode"},
		}},
	}

	var lines []string
	lines = append(lines, titleStyle.Render("urai  ·  உரை  ·  prose, speech, commentary"))
	lines = append(lines, "")

	for _, sec := range sections {
		lines = append(lines, sectionStyle.Render(sec.name))
		for _, s := range sec.entries {
			row := keyStyle.Render(s.key) + descStyle.Render(s.desc)
			lines = append(lines, row)
		}
	}

	lines = append(lines, "")
	lines = append(lines, dimStyle.Render("Press any key to start writing..."))

	box := borderStyle.Render(strings.Join(lines, "\n"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
