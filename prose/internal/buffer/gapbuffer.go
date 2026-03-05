package buffer

// GapBuffer provides efficient text editing with a gap buffer data structure.
// The gap is a section of the buffer where insertions/deletions are O(1).
// Moving the gap requires copying text, but cursor locality makes this efficient.
type GapBuffer struct {
	data     []rune
	gapStart int
	gapEnd   int
}

const initialGapSize = 64

// NewGapBuffer creates a new gap buffer with optional initial content.
func NewGapBuffer(content string) *GapBuffer {
	runes := []rune(content)
	size := len(runes) + initialGapSize

	gb := &GapBuffer{
		data:     make([]rune, size),
		gapStart: len(runes),
		gapEnd:   size,
	}

	copy(gb.data[:len(runes)], runes)
	return gb
}

// Length returns the number of characters (excluding the gap).
func (gb *GapBuffer) Length() int {
	return len(gb.data) - gb.gapSize()
}

// gapSize returns the current gap size.
func (gb *GapBuffer) gapSize() int {
	return gb.gapEnd - gb.gapStart
}

// moveGap moves the gap to the specified position.
func (gb *GapBuffer) moveGap(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > gb.Length() {
		pos = gb.Length()
	}

	if pos == gb.gapStart {
		return
	}

	if pos < gb.gapStart {
		// Move gap left: shift text right into the gap
		amount := gb.gapStart - pos
		copy(gb.data[gb.gapEnd-amount:gb.gapEnd], gb.data[pos:gb.gapStart])
		gb.gapStart = pos
		gb.gapEnd -= amount
	} else {
		// Move gap right: shift text left into the gap
		amount := pos - gb.gapStart
		copy(gb.data[gb.gapStart:gb.gapStart+amount], gb.data[gb.gapEnd:gb.gapEnd+amount])
		gb.gapStart += amount
		gb.gapEnd += amount
	}
}

// expandGap ensures the gap is at least minSize.
func (gb *GapBuffer) expandGap(minSize int) {
	if gb.gapSize() >= minSize {
		return
	}

	newGapSize := max(minSize, initialGapSize)
	newData := make([]rune, len(gb.data)+newGapSize-gb.gapSize())

	// Copy text before gap
	copy(newData[:gb.gapStart], gb.data[:gb.gapStart])

	// Copy text after gap
	afterGapLen := len(gb.data) - gb.gapEnd
	newGapEnd := len(newData) - afterGapLen
	copy(newData[newGapEnd:], gb.data[gb.gapEnd:])

	gb.data = newData
	gb.gapEnd = newGapEnd
}

// Insert inserts a rune at the specified position.
func (gb *GapBuffer) Insert(pos int, r rune) {
	gb.InsertString(pos, string(r))
}

// InsertString inserts a string at the specified position.
func (gb *GapBuffer) InsertString(pos int, s string) {
	runes := []rune(s)
	if len(runes) == 0 {
		return
	}

	gb.moveGap(pos)
	gb.expandGap(len(runes))

	copy(gb.data[gb.gapStart:], runes)
	gb.gapStart += len(runes)
}

// Delete removes count characters starting at pos.
func (gb *GapBuffer) Delete(pos, count int) string {
	if count <= 0 || pos < 0 || pos >= gb.Length() {
		return ""
	}

	if pos+count > gb.Length() {
		count = gb.Length() - pos
	}

	gb.moveGap(pos)

	// Characters to delete are now right after the gap
	deleted := string(gb.data[gb.gapEnd : gb.gapEnd+count])
	gb.gapEnd += count

	return deleted
}

// DeleteBackward removes count characters before pos.
func (gb *GapBuffer) DeleteBackward(pos, count int) string {
	if count <= 0 || pos <= 0 {
		return ""
	}

	start := pos - count
	if start < 0 {
		count = pos
		start = 0
	}

	return gb.Delete(start, count)
}

// At returns the character at the specified position.
func (gb *GapBuffer) At(pos int) rune {
	if pos < 0 || pos >= gb.Length() {
		return 0
	}

	if pos < gb.gapStart {
		return gb.data[pos]
	}
	return gb.data[pos+gb.gapSize()]
}

// String returns the entire buffer content as a string.
func (gb *GapBuffer) String() string {
	result := make([]rune, gb.Length())
	copy(result[:gb.gapStart], gb.data[:gb.gapStart])
	copy(result[gb.gapStart:], gb.data[gb.gapEnd:])
	return string(result)
}

// Substring returns a substring from start to end (exclusive).
func (gb *GapBuffer) Substring(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > gb.Length() {
		end = gb.Length()
	}
	if start >= end {
		return ""
	}

	result := make([]rune, end-start)
	for i := start; i < end; i++ {
		result[i-start] = gb.At(i)
	}
	return string(result)
}

// Lines returns the content split into lines.
func (gb *GapBuffer) Lines() []string {
	content := gb.String()
	if len(content) == 0 {
		return []string{""}
	}

	var lines []string
	start := 0
	runes := []rune(content)

	for i, r := range runes {
		if r == '\n' {
			lines = append(lines, string(runes[start:i]))
			start = i + 1
		}
	}

	// Add final line (may be empty)
	lines = append(lines, string(runes[start:]))
	return lines
}

// LineCount returns the number of lines.
func (gb *GapBuffer) LineCount() int {
	return len(gb.Lines())
}

// LineStart returns the position of the start of the given line (0-indexed).
func (gb *GapBuffer) LineStart(line int) int {
	if line <= 0 {
		return 0
	}

	currentLine := 0

	for i := 0; i < gb.Length(); i++ {
		if gb.At(i) == '\n' {
			currentLine++
			if currentLine == line {
				return i + 1
			}
		}
	}

	return gb.Length()
}

// LineEnd returns the position of the end of the given line (before newline).
func (gb *GapBuffer) LineEnd(line int) int {
	start := gb.LineStart(line)

	for i := start; i < gb.Length(); i++ {
		if gb.At(i) == '\n' {
			return i
		}
	}

	return gb.Length()
}

// PositionToLineCol converts a buffer position to line and column.
func (gb *GapBuffer) PositionToLineCol(pos int) (line, col int) {
	if pos < 0 {
		pos = 0
	}
	if pos > gb.Length() {
		pos = gb.Length()
	}

	line = 0
	col = 0

	for i := 0; i < pos; i++ {
		if gb.At(i) == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}

	return line, col
}

// LineColToPosition converts line and column to a buffer position.
func (gb *GapBuffer) LineColToPosition(line, col int) int {
	if line < 0 {
		line = 0
	}

	pos := gb.LineStart(line)
	lineEnd := gb.LineEnd(line)

	targetPos := pos + col
	if targetPos > lineEnd {
		return lineEnd
	}

	return targetPos
}
