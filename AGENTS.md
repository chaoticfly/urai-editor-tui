# urai (உரை) — Agent Context

Distraction-free terminal writing tool for screenwriters and authors.
Built in Go using the Bubbletea TUI framework (Elm-like MVU architecture).

---

## Quick start

```bash
cd prose
go build -o urai ./cmd/urai/   # build
./urai                          # run (new document)
./urai notes.md                 # open file
./urai screenplay.fountain      # open Fountain screenplay
```

Go module path: `urai` (not the repo directory name).

---

## Repository layout

```
prose/                    # Go module root (all source lives here)
  cmd/urai/main.go        # entry point — flag parsing, config load, tea.NewProgram
  internal/
    ai/         client.go          OpenAI-compatible HTTP client (Ollama / OpenAI)
    app/        app.go             Root model: owns all sub-models, routes all msgs
    buffer/     gapbuffer.go       Gap-buffer text engine (O(1) insert/delete at cursor)
                cursor.go          Logical cursor over a GapBuffer
    config/     config.go          JSON config load/save (~/.config/prose/config.json)
    editor/     model.go           Editor state: buffer, cursor, selection, history
                view.go            Editor rendering (line numbers, highlights, scrolling)
                keybindings.go     KeyMap — action enum + default bindings
    export/     html.go            Markdown → HTML via goldmark
                pdf.go             HTML → PDF via chromedp (headless Chrome)
    filebrowser/ model.go          Modal file browser (open / save-as / rename / delete)
    findbar/    model.go           Find / find+replace bar (case-insensitive, rune-aware)
    fountain/   fountain.go        Fountain screenplay autocomplete + Popup widget
    history/    history.go         Undo/redo stack
                command.go         Command objects: Insert, Delete, DeleteBackward, Replace
    layout/     split.go           SplitLayout + ViewMode enum + CycleViewMode
                statusbar.go       StatusBar renderer
    preview/    model.go           Async Markdown preview (glamour, goroutine-safe)
    settings/   model.go           Settings modal
    suggestions/ model.go          AI suggestions panel
```

---

## Architecture: Bubbletea MVU

Every package that renders UI exposes the same three methods:
- `Init() tea.Cmd` — initial command (usually nil)
- `Update(tea.Msg) (tea.Model, tea.Cmd)` — state transition
- `View() string` — render to string via lipgloss

**Message flow:**
1. `app.Model.Update` receives all messages first.
2. Overlay priority (highest wins input): ExitConfirm → Settings → FileBrowser → FindBar → Help → global keys → editor.
3. After each editor keystroke, `app` optionally triggers preview re-render (`updatePreviewContent`) and Fountain popup update (`updateFountainPopup`).
4. Sub-models communicate back to app via typed message structs (e.g. `filebrowser.OpenFileMsg`, `settings.SaveMsg`, `findbar.JumpMsg`).

**Never call `View()` to drive logic.** State lives in model fields; rendering is side-effect-free.

---

## Key data structures

### GapBuffer (`buffer/gapbuffer.go`)
Rune-indexed gap buffer. All positions throughout the codebase are **rune offsets**, not byte offsets.
- `Insert(pos, rune)` / `InsertString(pos, string)` — O(1) at gap, O(n) otherwise
- `Delete(pos, count) string` — returns deleted text
- `At(pos) rune`, `String() string`, `Lines() []string`
- `PositionToLineCol(pos) (line, col)` / `LineColToPosition(line, col) int`

### History / Command pattern (`history/`)
All buffer mutations go through `history.History.Execute(Command)`.
Commands: `InsertCommand`, `DeleteCommand`, `DeleteBackwardCommand`, `ReplaceCommand`.
Each command holds enough state to `Undo()` and `Redo()` itself.

### ViewMode (`layout/split.go`)
```
ViewModeEdit    — editor only
ViewModeSplit   — editor + markdown preview (auto-set for .md files)
ViewModePreview — preview only
ViewModeAI      — editor + AI suggestions panel
ViewModeFocus   — distraction-free, centred, no status bar
```
Cycled with `Ctrl+\`. Focus toggled separately with `Ctrl+F`.

---

## Config

File: `~/.config/prose/config.json` (Linux/macOS), `%APPDATA%\prose\config.json` (Windows).

```json
{
  "tab_size": 4,
  "word_wrap": true,
  "show_line_numbers": true,
  "background_color": "",
  "glamour_style": "dark",
  "ai": {
    "provider": "ollama",
    "base_url": "http://localhost:11434",
    "api_key": "",
    "model": "llama3",
    "system_prompt": "..."
  }
}
```

`background_color` accepts terminal color indices (`"235"`) or hex (`"#1a1a2e"`). Empty = terminal default.
`glamour_style` overridden by `GLAMOUR_STYLE` env var (set before program start to avoid terminal corruption).

---

## Fountain autocomplete

Triggered automatically for `.fountain` and `.spmd` files.
`fountain.GetCompletions(typedPrefix)` returns candidates; `fountain.Popup` manages selection state.
The popup occupies **1 terminal line** and is rendered between editor and status bar.
Editor height is reduced by `fountainPopup.Height()` when the popup is visible.

---

## Find/Replace bar (`findbar/`)

- `Ctrl+G` → find-only, `Ctrl+H` → find+replace.
- Search is case-insensitive, rune-aware, linear scan over `[]rune`.
- `JumpMsg` scrolls the editor; `ReplaceAllMsg` applies replacements back-to-front to keep earlier positions valid.
- Tab switches focus between find/replace fields.

---

## AI client (`ai/client.go`)

OpenAI-compatible `/v1/chat/completions` endpoint.
- Works with Ollama (default `http://localhost:11434`) and any OpenAI-compatible API.
- 30 s timeout, 500 max tokens.
- Called from `suggestions.Model` via a goroutine wrapped in a `tea.Cmd`.

---

## Export

| Format | Package | Mechanism |
|--------|---------|-----------|
| HTML | `export/html.go` | goldmark → `.html` file beside the source |
| PDF | `export/pdf.go` | goldmark → HTML → chromedp headless Chrome → `.pdf` |

PDF export requires Google Chrome or Chromium installed on the host.

---

## Keybindings (complete)

| Key | Action |
|-----|--------|
| `Ctrl+S` | Save (Save As if new file) |
| `Ctrl+Q` | Quit (confirm if unsaved) |
| `Ctrl+E` | Export HTML |
| `Ctrl+P` | Export PDF |
| `F1` / `Ctrl+?` | Help overlay |
| `F2` | Settings modal |
| `F3` | File browser |
| `F4` | Toggle AI Line/Paragraph mode |
| `Ctrl+\` | Cycle view mode |
| `Ctrl+F` | Toggle focus mode |
| `Ctrl+G` | Find |
| `Ctrl+H` | Find & Replace |
| `Ctrl+R` | Request AI suggestion |
| `Ctrl+L` | Insert AI suggestion at cursor |
| `Ctrl+Z` / `Ctrl+Y` | Undo / Redo |
| `Ctrl+X/C/V` | Cut / Copy / Paste |
| `Ctrl+A` | Select all |
| Arrow keys | Move cursor |
| `Alt+← / Alt+→` | Word left / right |
| `Home / End` | Line start / end |
| `PgUp / PgDn` | Page up / down |
| `Shift+↑↓←→` | Extend selection |
| `Tab` (Fountain) | Accept autocomplete |
| `↑ / ↓` (Fountain popup) | Cycle completions |
| `Esc` (Fountain popup) | Dismiss popup |
| `Esc` | Exit focus mode |

---

## Coding conventions

- **No global state.** All state is in model structs passed by pointer.
- **All buffer positions are rune offsets**, never byte offsets.
- **Every mutation goes through `history.Execute`** so undo/redo works.
- UI strings are built with `lipgloss` styles; never use raw ANSI escape codes.
- `tea.Cmd` (a `func() tea.Msg`) is the only way to trigger async work or send messages back to the runtime.
- Sub-model updates return `(SubModelType, tea.Cmd)`; the app casts back with type assertion (e.g. `updatedEditor.(*editor.Model)`).
- Height accounting: `editorHeight()` subtracts status bar (1), fountain popup (`fountainPopup.Height()`), and find bar (`findBar.Height()`).
- `glamourStyleFromConfig` must be called **before** `tea.NewProgram` starts; calling `WithAutoStyle()` inside a goroutine corrupts terminal output.

---

## What does NOT work in WASM

This is a native terminal app. The following are hard blockers for WebAssembly:
- **bubbletea** requires a real TTY for raw input/ANSI output.
- **chromedp** (PDF export) controls a host OS process.
- **Filesystem I/O** (`os.ReadFile`, `os.WriteFile`) is sandboxed in WASM.
- **`golang.org/x/sys`** uses syscalls unavailable in `GOARCH=wasm`.

Alternatives for browser access: expose via SSH (e.g. charmbracelet/wish) or wrap with ttyd/wetty.
