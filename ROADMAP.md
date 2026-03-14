# urai — Roadmap & Planned Features

This document tracks features that are planned, in design, or actively being built.
Completed items move to [CHANGELOG.md](CHANGELOG.md).

---

## Recently shipped (post-0.1.0)

### AI suggestions — bug fixes & accessibility
- Fixed multi-line suggestion insertion (was broken via KeyRunes hack; now uses `DirectInsert`)
- Added **Enter** as an alternative insert key when the AI panel is visible
  (`Enter` or `Ctrl+L` both insert the suggestion)
- Updated panel footer hint to reflect both keys
- Fixed silent failure when OpenAI provider was selected but no API key was set —
  now shows a clear error in the panel instead of a silent HTTP 401

### Crash recovery
- New recovery package writes `.filename.urai-recover` alongside each file
  every 30 seconds whenever the buffer has unsaved changes
- On launch, if a recovery file is newer than the saved file, it is loaded
  automatically and the status bar notifies the user
- Panic handler in `main` writes a last-resort recovery snapshot before crashing
- Recovery file is deleted on every successful save

### Session persistence
- Last opened file is saved to `~/.local/share/prose/session.json` on every
  save and on quit
- Launching `urai` with no arguments reopens the last file automatically
  (only if it still exists on disk)

### Auto-save
- The existing `auto_save` / `auto_save_delay_ms` config fields are now
  implemented: a background tick saves the file on the configured interval
  when the buffer is modified

### Robustness
- Config parse error now falls back to defaults with a warning instead of
  exiting — a corrupted config file no longer prevents the editor from opening
- New `Filepath()` and `Content()` accessors on `app.Model` for external use
  (panic handler, tests)

---

## Planned

### Plugin system

**Status:** Design complete, implementation not started.

#### Overview

Plugins are executables (any language) that the editor communicates with over
**stdin/stdout JSON**. No shared libraries, no embedding — just subprocesses.

A plugin lives in its own directory under `~/.config/prose/plugins/<name>/`
alongside a `plugin.json` manifest.

#### Manifest (`plugin.json`)

```json
{
  "id": "word-goal",
  "name": "Word Goal",
  "description": "Notify when hitting word-count milestones",
  "version": "1.0.0",
  "command": "./word-goal.sh",
  "timeout_ms": 3000,
  "triggers": {
    "events": ["on_save"],
    "shortcut": "ctrl+shift+g",
    "shortcut_alt": "alt+g"
  }
}
```

`triggers` is optional. A plugin can subscribe to events only, a shortcut
only, or both.

#### Events (auto-triggered)

| Event | When |
|---|---|
| `on_open` | File opened (including session restore) |
| `on_save` | After every successful save |
| `on_quit` | Before a clean exit (best-effort) |
| `on_change` | Content changed — debounced 800 ms after last keystroke |

#### Shortcut namespace

- Primary: `ctrl+shift+<letter>` (`a`–`z`)
  Requires kitty keyboard protocol (Kitty, WezTerm, foot, Ghostty).
  On older terminals the key simply does not fire; the plugin still works
  via events.
- Optional fallback: `shortcut_alt` field accepts `alt+<letter>` for
  broader terminal compatibility.

Shortcut keys are checked **before** editor key handling so plugins always
take priority.

#### Context sent to plugin (stdin, one JSON object)

```json
{
  "event": "on_save",
  "trigger": "event",
  "filepath": "/home/user/notes.md",
  "content": "...",
  "cursor_line": 10,
  "cursor_col": 3,
  "word_count": 1500,
  "view_mode": "split"
}
```

`trigger` is `"event"` or `"shortcut"`.

#### Response from plugin (stdout, one JSON object)

```json
{
  "message": "1,500 words — milestone!",
  "actions": [
    { "type": "set_message",     "text": "Linted: 0 issues" },
    { "type": "insert_at_cursor","text": "<!-- reviewed -->" },
    { "type": "replace_content", "content": "..." }
  ]
}
```

Empty stdout is valid and means the plugin ran silently.
All fields are optional.

#### Allowed actions (bounded — plugins cannot instruct the editor to do
arbitrary things)

| Action type | Effect |
|---|---|
| `set_message` | Display text in the status bar |
| `insert_at_cursor` | Insert text at the current cursor position |
| `replace_content` | Replace the entire editor buffer (e.g. formatter/linter) |

Plugins that need to write files, open a browser, etc. do so on their own —
they just cannot instruct the editor to perform side effects beyond this list.

#### Error handling

- Timeout (default 3 s, configurable per plugin) → error in status bar,
  editor unaffected
- Non-zero exit / crash → error in status bar
- Plugin stderr → appended to `~/.local/share/prose/plugin-errors.log`
- Bad manifest → skipped at load, error logged
- `replace_content` with empty string → treated as no-op (safeguard)

#### New packages

```
prose/internal/plugin/
  manifest.go   — Plugin struct, JSON loading, validation
  runner.go     — Subprocess execution, stdin/stdout serialisation, timeout
  manager.go    — Discover plugins, index by event and shortcut, dispatch
```

#### Integration points in app.go

1. **Init** — `manager.Load()` scans `~/.config/prose/plugins/`, skips
   invalid manifests, logs errors
2. **handleKey** — checks `ctrl+shift+*` / `alt+*` before forwarding to
   editor; dispatches matching plugin async
3. **On events** — `on_save` after save, `on_open` after `openFile`,
   `on_quit` before `tea.Quit`, `on_change` after debounce tick
4. **`pluginResultMsg` handler** — applies actions, surfaces errors in
   status bar

#### Out of scope (keeping it simple)

- No in-app plugin enable/disable UI — to disable a plugin, move or delete
  its directory
- No inter-plugin communication
- No hot-reload of manifests (restart editor to pick up new plugins)
- No sandboxing — plugins run with the user's full permissions

#### Open questions (resolve before implementing)

- [ ] Confirm shortcut namespace: `ctrl+shift+*` only, or also `alt+*` fallback?
- [ ] `on_change` debounce: 800 ms good? Adjust?
- [ ] `replace_content`: silent replacement or confirmation prompt first?
- [ ] Plugin error log: file only, or also a toggleable in-app log panel?

---

### Git versioning

**Status:** Design complete, implementation not started.

#### Overview

urai can use git as a first-class versioning backend for the folder you are
writing in. Every save becomes a commit. The user can tag meaningful points
("Chapter 1 done", "Draft 2") and browse or restore any past version from
inside the editor. No git knowledge required — the editor handles all
plumbing.

The feature is **opt-in per folder**: running the init action (or using the
version panel) on a directory that is not already a git repo asks for
confirmation before `git init`-ing it.

#### Behaviour

| Moment | What happens |
|---|---|
| Ctrl+S on a tracked file | `git add <file> && git commit -m "Save — <timestamp>"` |
| User tags a version | `git tag -a <label> -m "<label>"` on the current HEAD |
| User opens version panel | Shows a scrollable list of commits and tags for the current file |
| User selects a past commit | Loads that version's content into the editor buffer (does **not** `git checkout` — buffer is marked modified so the user decides whether to save) |
| User saves after restoring | Creates a new commit with message `"Restore — <tag/hash>"` |

Auto-commits only fire for files that are inside the git repo root. Files
outside (or in repos the user has not opted into) are silently skipped.

#### Version panel UI

Toggled with **F5** (placeholder — subject to change). Replaces the right
pane in split view (same pattern as the AI panel):

```
┌─ Versions ──────────────────────┐
│ ★ Draft 2 (v2)        3 days ago│
│   Save — 2026-03-10 14:22       │
│   Save — 2026-03-10 11:05       │
│ ★ Draft 1 (v1)        1 week ago│
│   Save — 2026-03-03 09:47       │
│                                  │
│ Enter: restore  T: tag  Esc: back│
└──────────────────────────────────┘
```

- `★` marks tagged versions
- Commits shown are scoped to the **current file** (`git log -- <file>`)
- Untracked files show a message: "Not tracked — press T to initialise"

#### Tagging

Press **T** inside the version panel. A small inline input appears:

```
Tag name: [Draft 2        ]   Enter to confirm · Esc to cancel
```

Creates `git tag -a "<label>" -m "<label>"` on HEAD. Tag names are
validated: no spaces (replaced with `-`), no special characters.

#### Key design decisions

- **No branch management** — linear history only (no merge, no rebase UI).
  Power users can use git directly in the terminal alongside the editor.
- **No remote push** — local repo only. Users who want remote backup point
  their repo at a remote themselves.
- **Scoped to file, not repo** — history panel shows only commits that
  touched the current file, not the entire repo log.
- **Never destructive** — restoring a past version loads it into the buffer
  but does not rewrite history. `git reset --hard` is never called.
- **Silent when git is absent** — if `git` is not in `PATH`, the feature
  is disabled entirely and no UI is shown.

#### New package

```
prose/internal/gitvcs/
  repo.go     — Detect repo root, init repo, stage + commit, tag
  log.go      — git log scoped to a file, parse output into structs
  restore.go  — Read file content at a given ref (git show <ref>:<path>)
```

All git operations are synchronous shell invocations (`exec.Command("git", ...)`)
wrapped in the same async `tea.Cmd` pattern used elsewhere. No libgit2 binding.

#### Integration points in app.go

1. **On save** — after successful save, fire `gitvcs.Commit(filepath)` as
   an async cmd; failure is logged but never blocks the editor
2. **F5** — toggle version panel (new `ViewModeVersion` layout mode)
3. **Version panel key handling** — Up/Down scroll, Enter restore, T tag,
   Esc back
4. **`gitCommitMsg` / `gitRestoreMsg`** — new message types handled in
   `Update`

#### Open questions (resolve before implementing)

- [ ] Auto-commit every save, or only on explicit user action?
  (Auto gives full history; explicit gives cleaner log)
- [ ] What happens to the auto-commit message format? Timestamp only, or
  include word count / file name?
- [ ] Should `git init` be triggered automatically on first save in an
  untracked folder, or only when the user opens the version panel?
- [ ] F5 for the panel — does this conflict with anything planned?

---

### Quick file picker

**Status:** Design complete. Can ship as a native feature now; can also be
reimplemented as a plugin once the plugin system exists.

#### Overview

A fuzzy-search popup that lets the user jump to any file in the working tree
without leaving the keyboard. Think Ctrl+P in VS Code or fzf — but embedded,
no external dependency.

This is distinct from the existing **file browser** (F3), which is a
full-screen navigator for creating, renaming, and deleting files. The quick
picker is purely for opening files fast.

#### Trigger

**Ctrl+P** — opens the picker as a floating overlay at the top of the screen.
Does not change view mode; closes and restores state on Esc or after opening.

#### UI

```
┌─ Open file ─────────────────────────────┐
│ > notes                                 │
├─────────────────────────────────────────┤
│   docs/notes.md                         │
│ ▶ notes/chapter-01.md                   │
│   notes/chapter-02.md                   │
│   scratch/notes-old.txt                 │
└─────────────────────────────────────────┘
```

- Input line at the top with live fuzzy filtering
- Results list below (capped at ~8 visible rows, scrollable)
- `▶` marks the selected entry
- Arrow keys or `ctrl+n` / `ctrl+p` move selection
- **Enter** opens the highlighted file
- **Esc** dismisses without opening

#### File indexing

On open, walk the working directory (or the git repo root if inside one) up
to a configurable depth (default 4). Respect `.gitignore` when a repo is
present. Hidden directories (`.git`, `.cache`, etc.) are always skipped.

The walk is synchronous and capped at 5,000 files — fast enough for prose
projects without needing a background indexer.

#### Fuzzy matching

Score each path against the query using a simple character-subsequence match
weighted by:
1. Consecutive character runs (bonus)
2. Match at start of filename vs. path component vs. middle of path

Results are sorted by score descending. No external library needed — the
algorithm is ~50 lines.

#### Recently opened files

When the query is empty, show the last 10 opened files (drawn from session
history) at the top of the list, above the directory walk results. This makes
switching between two documents instant.

#### New package

```
prose/internal/picker/
  model.go   — BubbleTea model, input handling, result list, rendering
  fuzzy.go   — Scoring and ranking
  walk.go    — Directory traversal, .gitignore filtering
```

#### Plugin path (future)

Once the plugin system is live, an alternative implementation could replace
or augment this with an external tool (e.g. `fzf`, `fd`) via a plugin that
uses `replace_content`-style actions — useful for users with very large
trees. The native picker ships first so the feature works out of the box.

#### Open questions (resolve before implementing)

- [ ] Walk from cwd or git repo root? (Repo root feels more natural when
  writing a book in a single repo)
- [ ] Should Ctrl+P also show recent files when inside the existing file
  browser (F3), or is it always a fresh overlay?
- [ ] Cap depth at 4 levels, or make it configurable in settings?

---

## Under consideration (no design yet)

- **Word count goals** — configurable target with progress indicator in
  status bar
- **Custom themes** — user-defined colour palette loaded from config
- **Multiple cursors** — limited multi-cursor editing
- **Git gutter** — show diff markers (modified/added/removed lines) against
  HEAD in the line-number gutter
- **Spell check** — underline unknown words, cycle through suggestions
- **Command palette** (Ctrl+K) — fuzzy-search all actions and plugin shortcuts
