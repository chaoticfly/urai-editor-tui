# urai — உரை

> *உரை (urai)* · Tamil · noun — **"prose; speech; commentary"**
> The art of giving words their meaning.

A distraction-free terminal writing tool for screenwriters and authors.
Supports Markdown, Fountain screenplays, AI-assisted writing, and full
file management — all from the keyboard.

![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)
![Go](https://img.shields.io/badge/go-1.24%2B-00ADD8.svg)

---

## Features

- **Rich text editing** — gap-buffer engine, undo/redo, selection, system clipboard
- **Markdown** — live split-preview rendered with Glamour
- **Fountain screenplays** — context-aware autocomplete for scene headings,
  transitions, parentheticals, and character cues
- **AI suggestions** — works with [Ollama](https://ollama.com) (local) or any
  OpenAI-compatible API
- **File browser** — navigate, open, create, rename, and delete files (F3)
- **Export** — HTML and PDF via headless Chrome
- **View modes** — Edit · Split · Preview · AI · Focus
- **Mouse support** — scroll wheel in preview and AI panels
- **Configurable** — background colour, AI provider, system prompt, and more

---

## Requirements

| Dependency | Version | Notes |
|---|---|---|
| [Go](https://golang.org/dl/) | 1.24+ | Build tool |
| [Google Chrome](https://www.google.com/chrome/) or Chromium | any | PDF/HTML export only |
| [Ollama](https://ollama.com) | any | Optional — for local AI |

---

## Building from source

```bash
# Clone
git clone https://github.com/your-username/urai.git
cd urai

# Build (outputs urai / urai.exe in prose/)
cd prose
go build -o urai ./cmd/urai/

# Run
./urai
```

On Windows (PowerShell):

```powershell
cd prose
go build -o urai.exe ./cmd/urai/
.\urai.exe
```

To install system-wide:

```bash
go install urai/cmd/urai@latest
# or from source:
go build -o ~/.local/bin/urai ./cmd/urai/
```

---

## Usage

```
urai [OPTIONS] [FILE]

Options:
  -config <path>   Path to a custom config file
  -version         Print version
  -help            Print this help
```

**Examples:**

```bash
urai                          # New document
urai screenplay.fountain      # Open a Fountain screenplay
urai notes.md                 # Open a Markdown file
urai -config ~/my-config.json # Use a custom config
```

---

## Keybindings

### File

| Key | Action |
|---|---|
| `F3` | File browser (open · new · rename · delete) |
| `Ctrl+S` | Save (opens Save As for new files) |
| `Ctrl+E` | Export to HTML |
| `Ctrl+P` | Export to PDF |
| `Ctrl+Q` | Quit |

### Edit

| Key | Action |
|---|---|
| `Ctrl+Z` / `Ctrl+Y` | Undo / Redo |
| `Ctrl+X` / `Ctrl+C` / `Ctrl+V` | Cut / Copy / Paste |
| `Ctrl+A` | Select all |

### Navigation

| Key | Action |
|---|---|
| Arrow keys | Move cursor |
| `Alt+←` / `Alt+→` | Move by word |
| `Home` / `End` | Line start / end |
| `PgUp` / `PgDn` | Page up / down |
| `Shift+↑↓←→` | Select text |

### View

| Key | Action |
|---|---|
| `Ctrl+\` | Cycle: Edit → Split → Preview → AI |
| `Ctrl+F` | Toggle focus mode |
| `Esc` | Exit focus mode |
| `F2` | Settings |

### AI Suggestions

| Key | Action |
|---|---|
| `Ctrl+R` | Request AI suggestion for current line/paragraph |
| `Ctrl+L` | Insert suggestion at cursor |
| `F4` | Toggle Line / Paragraph mode |

### Fountain Autocomplete

Triggers automatically in `.fountain` files.

| Key | Action |
|---|---|
| `Tab` | Accept selected completion |
| `↑` / `↓` | Cycle through suggestions |
| `Esc` | Dismiss popup |

**What triggers completions:**

| Typed | Suggestions |
|---|---|
| `INT`, `EXT` | Scene heading starters |
| `INT. LOCATION` | `- DAY`, `- NIGHT`, `- CONTINUOUS`, … |
| `FADE`, `CUT`, `DISSOLVE`, … | Transition variants |
| `(` | `(beat)`, `(pause)`, `(quietly)`, `(V.O.)`, … |
| `CHARACTER ` (uppercase + space) | `(V.O.)`, `(O.S.)`, `(CONT'D)`, … |

---

## Configuration

Config is stored at:

- **Linux / macOS:** `~/.config/prose/config.json`
- **Windows:** `%APPDATA%\prose\config.json`

```json
{
  "tab_size": 4,
  "word_wrap": true,
  "show_line_numbers": true,
  "background_color": "",
  "ai": {
    "provider": "ollama",
    "base_url": "http://localhost:11434",
    "api_key": "",
    "model": "llama3",
    "system_prompt": "You are a helpful writing assistant. Suggest improvements or completions for the provided text. Be concise."
  }
}
```

`background_color` accepts terminal colour codes (`0`, `235`) or hex (`#1a1a2e`).
Leave empty to use the terminal's default background.

### AI setup

**Ollama (local, recommended):**

```bash
ollama pull llama3
# urai will connect to http://localhost:11434 by default
```

**OpenAI or compatible API:**

Open Settings (`F2`), set Provider to `OpenAI Compatible`, enter your
Base URL and API key.

---

## Project structure

```
prose/
├── cmd/urai/        # main package
├── internal/
│   ├── ai/          # HTTP client for chat completions
│   ├── app/         # root application model
│   ├── buffer/      # gap buffer + cursor + selection
│   ├── config/      # JSON config load/save
│   ├── editor/      # editor model + view + keybindings
│   ├── export/      # HTML and PDF export
│   ├── filebrowser/ # file browser modal
│   ├── fountain/    # Fountain autocomplete engine
│   ├── history/     # undo/redo command stack
│   ├── layout/      # split layout + status bar
│   ├── preview/     # Markdown preview
│   ├── settings/    # settings modal
│   └── suggestions/ # AI suggestions panel
├── go.mod
└── .gitignore
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

---

## License

[MIT](LICENSE) © 2026 urai contributors
