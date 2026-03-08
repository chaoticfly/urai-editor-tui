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

## Installation

### Linux / macOS (ARM)

```bash
curl -fsSL https://raw.githubusercontent.com/OWNER/urai/master/install.sh | bash
```

or with wget:

```bash
wget -qO- https://raw.githubusercontent.com/OWNER/urai/master/install.sh | bash
```

Installs to `/usr/local/bin` if writable, otherwise `~/.local/bin`.

### Windows

```powershell
irm https://raw.githubusercontent.com/OWNER/urai/master/install.ps1 | iex
```

Installs to `%LOCALAPPDATA%\Programs\urai` and adds it to your user PATH automatically.

### Specific version

```bash
curl -fsSL .../install.sh | VERSION=v0.2.0 bash   # Linux / macOS
$env:VERSION = "v0.2.0"; irm .../install.ps1 | iex # Windows
```

---

## Building from source

```bash
git clone https://github.com/OWNER/urai.git
cd urai/prose
go build -o urai ./cmd/urai/
./urai
```

On Windows (PowerShell):

```powershell
cd urai\prose
go build -o urai.exe ./cmd/urai/
.\urai.exe
```

---

## Documentation

See [USER_GUIDE.md](USER_GUIDE.md) for usage, keybindings, and configuration.

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
