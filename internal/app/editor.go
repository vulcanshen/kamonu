package app

// The editor: a kamon's 25 × 19 cells in a popup, each cell turned by Enter
// or Space to the next letter and back to none (the checkbox's Space, tdp
// K5), made a letter by its digit, emptied by Backspace, and all emptied by
// x after a confirm; a row under them says which digit and colour is which
// letter, and p shows every letter gold as the splash will. Under that the
// action row, Save As and Save, and Tab between the two (as on a form). It is
// kamonu's own popup, outside tdp F1's seven kinds: docs/dev-remarks.md,
// 偏離 tdp. kamonu draws the inside; the frame, the animation and the popups
// over it are the lib's (terminu.Popup).

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	terminu "github.com/vulcanshen/terminu-go"
)

const (
	glyphDraw   = "\U000F02C1" // nf-md-grid: the editor's title
	glyphSquare = "\U0000F0C8" // nf-fa-square: a cell of a mark (tdp S5)
)

// The buttons, left to right.
const (
	saveAs = iota
	save
)

var buttonLabels = [...]string{saveAs: "Save As", save: "Save"}

type editor struct {
	k         *kamonu
	kamon     string // the kamon's name
	km        kamon
	r, c      int // the cursor on the grid
	onButtons bool
	button    int
	dirty     bool // a cell was turned: Esc asks first (components/dialog/form)
	gold      bool // p: every letter gold, as the splash draws it
	top       int  // the first row of the grid shown, when it does not all fit
}

// color is how the editor draws a cell: its letter's colour, or gold for
// every letter while p shows the mark as the splash draws it.
func (e *editor) color(r, c int) terminu.Color {
	if e.gold && e.km.cells[r][c] > 0 && !e.km.onFrame(r, c) {
		return goldColor
	}
	return e.km.color(r, c)
}

// newEditor opens on the first cell the cursor can stop on.
func newEditor(k *kamonu, name string, km kamon) *editor {
	e := &editor{k: k, kamon: name, km: km, button: save}
	for i := range rows * cols {
		if !km.onFrame(i/cols, i%cols) {
			e.r, e.c = i/cols, i%cols
			break
		}
	}
	return e
}

// openEditor opens the editor on the kamon name: "DRAW: <name>".
func (k *kamonu) openEditor(name string, km kamon) tea.Cmd {
	return k.a.Popup(terminu.Popup{Title: "Draw", Glyph: glyphDraw, Subject: name, Body: newEditor(k, name, km)})
}

// move steps the cursor dr rows and dc columns, past the U frame's cells
// (they cannot change: tdp S5) to the next one it can stop on; when there is
// none that way, it stays (tdp K12: what moves in space does not wrap).
func (e *editor) move(dr, dc int) {
	for r, c := e.r+dr, e.c+dc; r >= 0 && r < rows && c >= 0 && c < cols; r, c = r+dr, c+dc {
		if !e.km.onFrame(r, c) {
			e.r, e.c = r, c
			return
		}
	}
}

// cycle turns the cell to the next letter, and after the last back to none.
func (e *editor) cycle() {
	p := &e.km.cells[e.r][e.c]
	*p = (*p + 1) % uint8(len(e.km.letters)+1)
	e.dirty = true
}

// set makes the cell letter n, 0 for none; a letter the kamon does not have
// changes nothing.
func (e *editor) set(n uint8) {
	if int(n) > len(e.km.letters) {
		return
	}
	if p := &e.km.cells[e.r][e.c]; *p != n {
		*p = n
		e.dirty = true
	}
}

// toggle is a digit: it makes the cell letter n, and empties it when it is
// that letter already.
func (e *editor) toggle(n uint8) {
	if e.km.cells[e.r][e.c] == n {
		n = 0
	}
	e.set(n)
}

// empty reports a drawing with no cell of any letter.
func (e *editor) empty() bool { return e.km.cells == [rows][cols]uint8{} }

// clear empties the cell, whatever letter it is.
func (e *editor) clear() { e.set(0) }

// Size is the inside: the grid with a blank cell on either side; the grid, a
// blank row, the letters, the separator, the action row.
func (e *editor) Size(sw, sh int) (int, int) { return 2*cols + 2, rows + 4 }

func (e *editor) Hint() []terminu.KeyHint {
	if !e.onButtons {
		return []terminu.KeyHint{{Key: "Enter|Space", Desc: "cycle"}, {Key: "Tab", Desc: "actions"}, {Key: "Esc", Desc: "cancel"}}
	}
	return []terminu.KeyHint{{Key: "Enter", Desc: strings.ToLower(buttonLabels[e.button])}, {Key: "Tab", Desc: "grid"}, {Key: "Esc", Desc: "cancel"}}
}

func (e *editor) Draw(w, h int, layer terminu.Color) []string {
	square := glyphSquare
	if terminu.IconWidth() == 1 {
		square += " " // a square takes two columns (tdp S5, E5)
	}
	grid := make([]string, 0, rows)
	for r := range rows {
		var b strings.Builder
		for c := range cols {
			st := terminu.Fg(e.color(r, c))
			if !e.onButtons && r == e.r && c == e.c {
				// The cursor is the layer colour behind the square, which
				// keeps its own colour: a cell's colour is its value (tdp
				// components/color, the cursor block).
				st.Bg = layer
			}
			b.WriteString(st.Render(square))
		}
		grid = append(grid, padRow(b.String(), w))
	}
	// The letters, each after the digit that paints it (a hotkey is shown:
	// tdp M5) and its colour, gold with the rest under p; one space between
	// them when two do not fit.
	var legend []string
	for i, l := range e.km.letters {
		swatch := letterColors[i]
		if e.gold {
			swatch = goldColor
		}
		legend = append(legend, terminu.Fg(terminu.Text).Render("["+strconv.Itoa(i+1)+"] ")+
			terminu.Fg(swatch).Render(square)+terminu.Fg(terminu.Text).Render(l))
	}
	letters := strings.Join(legend, "  ")
	if terminu.Width(letters) > w-2 {
		letters = strings.Join(legend, " ")
	}
	// The letters, the separator and the action row stay at the bottom.
	tail := []string{padRow(letters, w), " " + terminu.Fg(terminu.Overlay0).Render(terminu.Repeat("─", w-2)) + " "}
	var buttons []string
	for i, label := range buttonLabels {
		b := " " + label + " "
		if e.onButtons && e.button == i {
			b = terminu.On(layer).Render(b)
		} else {
			b = terminu.Style{Fg: terminu.Text, Bg: terminu.Surface1}.Render(b)
		}
		buttons = append(buttons, b)
	}
	tail = append(tail, terminu.PadLeft(strings.Join(buttons, " "), w-1)+" ")

	var body []string
	switch {
	case h >= rows+4:
		body = append(append(grid, ""), tail...) // a blank row above the letters
	case h >= rows+3:
		body = append(grid, tail...) // one row short: the blank row goes
	default:
		// Shorter still (a terminal under 27 rows): the grid scrolls the
		// least that keeps the cursor's row in view, and the rows under it
		// stay, as a form's fields scroll over its buttons
		// (components/dialog/form, layout/popup).
		// TODO(lib KM9): the bottom frame line cannot say which rows show.
		area := max(h-len(tail), 0)
		if !e.onButtons && area > 0 {
			e.top = min(e.top, e.r)
			e.top = max(e.top, e.r-area+1)
		}
		e.top = min(max(e.top, 0), rows-area)
		body = append(grid[e.top:e.top+area:e.top+area], tail...)
	}
	body = body[:min(len(body), max(h, 0))]
	for i, l := range body {
		body[i] = terminu.Fit(l, w) // a screen too small for the drawing cuts it (tdp E5)
	}
	return body
}

// padRow is a row of the inside, one cell from the frame on either side.
func padRow(s string, w int) string { return " " + terminu.Fit(s, w-2) + " " }

func (e *editor) Key(key string, _ tea.KeyMsg) (bool, tea.Cmd) {
	switch key {
	case "Tab", "Shift-Tab":
		e.onButtons = !e.onButtons
		return true, nil
	case "Esc":
		if !e.dirty {
			return false, nil // the lib closes it
		}
		a := e.k.a
		return true, a.Confirm(terminu.Ask{Title: "Discard", Question: "Discard changes to " + e.kamon + "?", Verb: "discard"},
			func() tea.Cmd { return a.ClosePopup(e) })
	case "q":
		return true, nil // no quit flow from a form (components/dialog/form)
	case "p":
		e.gold = !e.gold
		return true, nil
	case "x":
		if e.empty() {
			return true, nil // nothing to clear: nothing to ask
		}
		return true, e.k.a.Confirm(terminu.Ask{Title: "Clear", Question: "Clear every cell of " + e.kamon + "?", Verb: "clear"},
			func() tea.Cmd {
				e.km.cells = [rows][cols]uint8{}
				e.dirty = true
				return nil
			})
	case "Space":
		if !e.onButtons {
			e.cycle()
		}
		return true, nil
	case "Backspace":
		if !e.onButtons {
			e.clear()
		}
		return true, nil
	}
	if e.onButtons {
		// The buttons are a row of their own: the next and the one before,
		// round, and the first and the last (components/dialog/form, the
		// action area; tdp K12).
		n := len(buttonLabels)
		switch key {
		case "l", "→", "j", "↓":
			e.button = (e.button + 1) % n
		case "h", "←", "k", "↑":
			e.button = (e.button + n - 1) % n
		case "gg":
			e.button = 0
		case "G":
			e.button = n - 1
		case "Enter":
			if e.button == save {
				return true, e.k.save(e)
			}
			return true, e.k.saveAs(e)
		default:
			return false, nil
		}
		return true, nil
	}
	switch key {
	case "h", "←":
		e.move(0, -1)
	case "l", "→":
		e.move(0, 1)
	case "k", "↑":
		e.move(-1, 0)
	case "j", "↓":
		e.move(1, 0)
	case "Enter":
		e.cycle()
	case "1", "2", "3", "4", "5", "6":
		e.toggle(key[0] - '0')
	default:
		return false, nil
	}
	return true, nil
}

// QuitNote is what quitting now loses: the quit flow's confirm says it in
// Peach (tdp K9).
func (e *editor) QuitNote() string {
	if !e.dirty {
		return ""
	}
	return "Changes to " + e.kamon + " are not saved."
}

// digits are the keys of n letters: "1", or a range, "1–3" (tdp M5).
func digits(n int) string {
	if n == 1 {
		return "1"
	}
	return "1–" + strconv.Itoa(n)
}

func (e *editor) Keys() []terminu.KeySection {
	return []terminu.KeySection{
		{Rows: []terminu.KeyRow{
			{Key: "Enter|Space", Desc: "the cell's next letter; after the last, none"},
			{Key: digits(len(e.km.letters)), Desc: "make the cell that letter; again to empty it"},
			{Key: "Backspace", Desc: "empty the cell"},
			{Key: "p", Desc: "every letter gold, as in the splash, or back"},
			{Key: "x", Desc: "empty every cell (asks first)"},
			{Key: "Enter", Desc: "on a button: press it"},
			{Key: "Tab|Shift-Tab", Desc: "between the grid and the buttons, back to the cell left"},
			{Key: "Esc", Desc: "cancel (asks first when something changed)"},
		}},
		{Title: "move", Rows: []terminu.KeyRow{
			{Key: "h|j|k|l", Desc: "left, down, up, right, past the U frame"},
			{Key: "h|j|k|l", Desc: "on the buttons: the one before or after, round"},
			{Key: "gg|G", Desc: "on the buttons: the first or the last"},
		}},
	}
}

// save writes the kamon back to its file and closes the editor; a failed
// write opens an error popup over it (components/layout/popup).
func (k *kamonu) save(e *editor) tea.Cmd {
	if err := k.st.save(e.kamon, e.km); err != nil {
		return k.a.Error("Save failed", "Can't save "+e.kamon+": "+why(err))
	}
	k.reload(e.kamon)
	return k.a.ClosePopup(e)
}

// saveAs asks for a new name and writes the drawing there, the letters and
// the U frame with it; the kamon it came from stays as it was. The title says
// what is done, so an error names the value besides: "Name can't have
// control characters" (tdp v0.6.11 components/input).
func (k *kamonu) saveAs(e *editor) tea.Cmd {
	return k.a.Text(terminu.TextInput{
		Title: "Save as", Name: "Name", Glyph: glyphSave, Prompt: "Name for the new kamon; " + e.kamon + " stays as it was", Verb: "save",
		Check: k.st.checkName,
		Done: func(v string) tea.Cmd {
			if err := k.st.create(v, e.km); err != nil {
				return k.a.Error("Save as failed", "Can't save "+v+": "+why(err))
			}
			k.reload(v)
			return k.a.ClosePopup(e)
		},
	})
}
