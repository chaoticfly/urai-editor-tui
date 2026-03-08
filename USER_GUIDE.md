# urai — User Guide

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

### Find & Replace

| Key | Action |
|---|---|
| `Ctrl+G` | Find |
| `Ctrl+H` | Find & Replace |

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
ollama pull llama3.2
# urai will connect to http://localhost:11434 by default
```

**OpenAI or compatible API:**

Open Settings (`F2`), set Provider to `OpenAI Compatible`, enter your
Base URL and API key.
