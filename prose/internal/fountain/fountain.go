package fountain

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// IsFountainFile reports whether the given file path is a Fountain screenplay.
func IsFountainFile(fp string) bool {
	ext := strings.ToLower(filepath.Ext(fp))
	return ext == ".fountain" || ext == ".spmd"
}

// Completion is a single autocomplete candidate.
type Completion struct {
	Display string
	Insert  string // full replacement for the typed prefix
}

var sceneHeadings = []string{
	"INT. ",
	"EXT. ",
	"INT./EXT. ",
	"I/E. ",
}

var timeOfDay = []string{
	" - DAY",
	" - NIGHT",
	" - CONTINUOUS",
	" - LATER",
	" - MOMENTS LATER",
	" - DUSK",
	" - DAWN",
	" - MORNING",
	" - EVENING",
	" - AFTERNOON",
	" - FLASHBACK",
	" - DREAM",
}

var transitions = []string{
	"FADE IN:",
	"FADE OUT:",
	"FADE TO BLACK:",
	"CUT TO:",
	"SMASH CUT TO:",
	"MATCH CUT TO:",
	"DISSOLVE TO:",
	"JUMP CUT TO:",
	"WIPE TO:",
	"IRIS IN:",
	"IRIS OUT:",
	"TIME CUT:",
}

var parentheticals = []string{
	"(beat)",
	"(pause)",
	"(sighs)",
	"(quietly)",
	"(loudly)",
	"(to himself)",
	"(to herself)",
	"(into phone)",
	"(reading)",
	"(smiling)",
	"(laughing)",
	"(crying)",
	"(angrily)",
	"(sarcastically)",
	"(whispering)",
	"(shouting)",
}

var characterExtensions = []string{
	"(V.O.)",
	"(O.S.)",
	"(O.C.)",
	"(CONT'D)",
	"(PRE-LAP)",
}

// GetCompletions returns autocomplete candidates for the given typed line prefix
// (text from line start to cursor).
func GetCompletions(typed string) []Completion {
	trimmed := strings.TrimLeft(typed, " \t")
	upper := strings.ToUpper(trimmed)

	if trimmed == "" {
		return nil
	}

	switch {
	// Parentheticals
	case strings.HasPrefix(trimmed, "("):
		return filterCompletions(parentheticals, trimmed)

	// Scene heading + location already typed → suggest time-of-day suffix
	case hasSceneLocation(upper):
		var results []Completion
		for _, tod := range timeOfDay {
			full := trimmed + tod
			results = append(results, Completion{Display: full, Insert: full})
		}
		return results

	// Scene heading starters
	case strings.HasPrefix(upper, "INT") || strings.HasPrefix(upper, "EXT"):
		return filterCompletions(sceneHeadings, upper)

	case strings.HasPrefix(upper, "INT./EXT") || strings.HasPrefix(upper, "I/E"):
		return filterCompletions(sceneHeadings[2:], upper)

	// Transitions
	case matchesAny(upper, "FADE", "CUT", "SMASH", "MATCH", "DISSOLVE", "WIPE", "IRIS", "JUMP", "TIME"):
		return filterCompletions(transitions, upper)

	// Character cue with a trailing space → suggest extensions
	case isCharacterCue(upper) && strings.HasSuffix(trimmed, " "):
		return filterCompletions(characterExtensions, trimmed)
	}

	return nil
}

// Suffix returns the portion of completion that should be inserted after what's already typed.
func Suffix(typed, completion string) string {
	typedRunes := []rune(typed)
	compRunes := []rune(completion)
	n := 0
	for n < len(typedRunes) && n < len(compRunes) {
		if unicode.ToUpper(typedRunes[n]) != unicode.ToUpper(compRunes[n]) {
			break
		}
		n++
	}
	if n >= len(compRunes) {
		return ""
	}
	return string(compRunes[n:])
}

func filterCompletions(candidates []string, upper string) []Completion {
	var results []Completion
	for _, c := range candidates {
		if strings.HasPrefix(strings.ToUpper(c), upper) {
			results = append(results, Completion{Display: c, Insert: c})
		}
	}
	return results
}

func matchesAny(upper string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(upper, p) {
			return true
		}
	}
	return false
}

// hasSceneLocation returns true when the line already contains a scene type +
// location (e.g. "INT. OFFICE") but no time-of-day suffix yet.
func hasSceneLocation(upper string) bool {
	for _, prefix := range []string{"INT. ", "EXT. ", "INT./EXT. ", "I/E. "} {
		if strings.HasPrefix(upper, prefix) && len(upper) > len(prefix) && !strings.Contains(upper, " - ") {
			return true
		}
	}
	return false
}

func isCharacterCue(upper string) bool {
	if len(upper) < 2 {
		return false
	}
	for _, r := range upper {
		if !unicode.IsUpper(r) && r != ' ' && r != '\'' && r != '.' && r != '-' {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// Popup
// ---------------------------------------------------------------------------

// Popup is the autocomplete popup state.
type Popup struct {
	completions []Completion
	selected    int
	visible     bool
	typed       string
	dismissed   bool // manually dismissed; re-opens on next char
}

// Update refreshes the popup based on the current typed line prefix.
// Call this after every keystroke.
func (p *Popup) Update(typed string) {
	// If the user dismissed at this exact prefix, don't re-open until it changes.
	if p.dismissed && typed == p.typed {
		return
	}
	if typed != p.typed {
		p.dismissed = false
	}

	p.typed = typed
	p.completions = GetCompletions(typed)
	if len(p.completions) == 0 {
		p.visible = false
		p.selected = 0
		return
	}
	p.visible = true
	if p.selected >= len(p.completions) {
		p.selected = 0
	}
}

// Dismiss hides the popup until the typed prefix changes.
func (p *Popup) Dismiss() {
	p.visible = false
	p.dismissed = true
}

// Visible reports whether the popup should be shown.
func (p *Popup) Visible() bool {
	return p.visible && len(p.completions) > 0
}

// SelectNext moves to the next completion.
func (p *Popup) SelectNext() {
	if len(p.completions) == 0 {
		return
	}
	p.selected = (p.selected + 1) % len(p.completions)
}

// SelectPrev moves to the previous completion.
func (p *Popup) SelectPrev() {
	if len(p.completions) == 0 {
		return
	}
	p.selected = (p.selected - 1 + len(p.completions)) % len(p.completions)
}

// Accept returns the text suffix to insert for the current selection.
func (p *Popup) Accept() string {
	if !p.visible || len(p.completions) == 0 {
		return ""
	}
	return Suffix(p.typed, p.completions[p.selected].Insert)
}

// Height returns how many terminal lines the popup occupies (0 when hidden).
func (p *Popup) Height() int {
	if !p.Visible() {
		return 0
	}
	return 1
}

// View renders the popup as a single-line strip.
func (p *Popup) View(width int) string {
	if !p.Visible() {
		return ""
	}

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Padding(0, 1)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("7")).
		Padding(0, 1)

	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)

	bgStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("235")).
		Width(width)

	const maxShow = 5
	var parts []string
	for i, c := range p.completions {
		if i >= maxShow {
			break
		}
		if i > 0 {
			parts = append(parts, sepStyle.Render("│"))
		}
		if i == p.selected {
			parts = append(parts, selectedStyle.Render(c.Display))
		} else {
			parts = append(parts, normalStyle.Render(c.Display))
		}
	}
	if len(p.completions) > maxShow {
		parts = append(parts, sepStyle.Render("│"), normalStyle.Render("..."))
	}

	hint := hintStyle.Render("  Tab:accept  ↑↓:cycle  Esc:dismiss")
	content := strings.Join(parts, "") + hint

	return bgStyle.Render(content)
}
