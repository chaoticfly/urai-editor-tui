# Changelog

All notable changes to urai will be documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
urai uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

## [0.1.0] — 2026-03-05

### Added

- Gap-buffer text editor with full undo/redo history
- Text selection, cut, copy, paste (system clipboard)
- Markdown split-preview rendered via Glamour
- Fountain screenplay support
  - Context-aware autocomplete: scene headings, time-of-day suffixes,
    transitions, parentheticals, character cue extensions
  - Tab to accept, ↑↓ to cycle, Esc to dismiss
- AI suggestions panel (Ollama and OpenAI-compatible APIs)
  - Line mode and paragraph mode
  - Ctrl+R to request, Ctrl+L to insert
- File browser (F3): navigate, open, create, rename, delete files
- Save As dialog for unnamed documents (Ctrl+S)
- Export to HTML and PDF via headless Chrome (chromedp)
- View modes: Edit, Split, Preview, AI, Focus
- Focus mode for distraction-free writing
- Settings panel (F2): AI provider, model, API key, background colour
- Mouse scroll support
- Configurable background colour
- MIT licence
