package main

// The editor: a kamon's 25 × 19 cells in a popup, each cell turned by Enter
// or Space to the next letter and back to none (the checkbox's Space, tdp
// K5), made a letter by its digit, emptied by Backspace, and all emptied by
// x after a confirm; a row under them says which digit and colour is which
// letter, and p shows every letter gold as the splash will. Under that the
// action row, Save As and Save, and Tab between the two (as on a form). It is
// kamonu's own popup, outside tdp F1's seven kinds: docs/dev-remarks.md,
// 偏離 tdp.

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	terminu "github.com/vulcanshen/terminu-go"
)

const glyphDraw = "\U000F02C1" // nf-md-grid: the editor's title

// The buttons, left to right.
const (
	saveAs = iota
	save
)

var buttonLabels = [...]string{saveAs: "Save As", save: "Save"}

type editor struct {
	kamon     string // the kamon's name
	km        kamon
	r, c      int // the cursor on the grid
	onButtons bool
	button    int
	dirty     bool // a cell was turned: Esc asks first (components/dialog/form)
	gold      bool // p: every letter gold, as the splash draws it
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
func newEditor(name string, km kamon) *editor {
	e := &editor{kamon: name, km: km, button: save}
	for i := range rows * cols {
		if !km.onFrame(i/cols, i%cols) {
			e.r, e.c = i/cols, i%cols
			break
		}
	}
	return e
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

func (e *editor) size(sw, sh int) (int, int) {
	// The grid, a blank row, the letters, the separator, the action row.
	return popupWidth(sw, 2*cols), min(rows+4+4, max(sh-2, 3))
}

func (e *editor) hint() []terminu.KeyHint {
	if !e.onButtons {
		return []terminu.KeyHint{{Key: "Enter|Space", Desc: "cycle"}, {Key: "Tab", Desc: "actions"}, {Key: "Esc", Desc: "cancel"}}
	}
	return []terminu.KeyHint{{Key: "Enter", Desc: strings.ToLower(buttonLabels[e.button])}, {Key: "Tab", Desc: "grid"}, {Key: "Esc", Desc: "cancel"}}
}

func (e *editor) draw(w, h int, layer terminu.Color) []string {
	inner := w - 2
	square := glyphSquare
	if terminu.IconWidth() == 1 {
		square += " " // a square takes two columns (tdp S5, E5)
	}
	body := make([]string, 0, rows+4)
	for r := range rows {
		var b strings.Builder
		for c := range cols {
			st := terminu.Fg(e.color(r, c))
			if !e.onButtons && r == e.r && c == e.c {
				// TODO(tdp KM3): the cursor is the layer colour behind the
				// square, which keeps its own colour: the cell's state shows.
				st.Bg = layer
			}
			b.WriteString(st.Render(square))
		}
		body = append(body, padRow(b.String(), inner))
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
	if terminu.Width(letters) > inner-2 {
		letters = strings.Join(legend, " ")
	}
	// A blank row above the letters when there is room: on a screen 28 rows
	// high the popup is one short, and the action row must stay.
	if rows+4 <= h-4 {
		body = append(body, "")
	}
	body = append(body, padRow(letters, inner),
		" "+terminu.Fg(terminu.Overlay0).Render(terminu.Repeat("─", inner-2))+" ")
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
	body = append(body, terminu.PadLeft(strings.Join(buttons, " "), inner-1)+" ")
	return box(layer, glyphDraw, "DRAW: "+e.kamon, e.hint(), "", w, body[:min(len(body), max(h-4, 0))])
}

func (e *editor) key(k *kamonu, key string, _ tea.KeyMsg) (bool, tea.Cmd) {
	switch key {
	case "Tab", "Shift-Tab":
		e.onButtons = !e.onButtons
		return true, nil
	case "Esc":
		if !e.dirty {
			return true, k.closeTop()
		}
		return true, k.open(&confirm{
			ask:    terminu.Ask{Title: "Discard", Question: "Discard changes to " + e.kamon + "?", Verb: "discard"},
			accept: k.closeAll,
		})
	case "q":
		return true, nil // no quit flow from a form (components/dialog/form)
	case "p":
		e.gold = !e.gold
		return true, nil
	case "x":
		if e.empty() {
			return true, nil // nothing to clear: nothing to ask
		}
		return true, k.open(&confirm{
			ask: terminu.Ask{Title: "Clear", Question: "Clear every cell of " + e.kamon + "?", Verb: "clear"},
			accept: func() tea.Cmd {
				e.km.cells = [rows][cols]uint8{}
				e.dirty = true
				return nil
			},
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
		switch key {
		case "h", "←":
			e.button = saveAs
		case "l", "→":
			e.button = save
		case "Enter":
			if e.button == save {
				return true, k.save(e)
			}
			return true, k.saveAs(e)
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

func (e *editor) name() string { return "Draw" }

// digits are the keys of n letters: "1", or a range, "1–3" (tdp M5).
func digits(n int) string {
	if n == 1 {
		return "1"
	}
	return "1–" + strconv.Itoa(n)
}

func (e *editor) keyRef() []refSection {
	return []refSection{
		{rows: []terminu.KeyRow{
			{Key: "Enter|Space", Desc: "the cell's next letter; after the last, none"},
			{Key: digits(len(e.km.letters)), Desc: "make the cell that letter; again to empty it"},
			{Key: "Backspace", Desc: "empty the cell"},
			{Key: "p", Desc: "every letter gold, as in the splash, or back"},
			{Key: "x", Desc: "empty every cell (asks first)"},
			{Key: "Enter", Desc: "on a button: press it"},
			{Key: "Tab|Shift-Tab", Desc: "between the grid and the buttons, back to the cell left"},
			{Key: "Esc", Desc: "cancel (asks first when something changed)"},
		}},
		{title: "move", rows: []terminu.KeyRow{
			{Key: "h|j|k|l", Desc: "left, down, up, right, past the U frame"},
			{Key: "h|l", Desc: "on the buttons: between them"},
		}},
	}
}

func (e *editor) typing() bool { return false }

// save writes the kamon back to its file and closes the editor; a failed
// write opens an error popup over it (components/layout/popup).
func (k *kamonu) save(e *editor) tea.Cmd {
	if err := k.st.save(e.kamon, e.km); err != nil {
		return k.open(errorPopup("Save", "Can't save "+e.kamon+": "+why(err), k.width()))
	}
	k.reload(e.kamon)
	return k.closeAll()
}

// saveAs asks for a new name and writes the drawing there, the letters and
// the U frame with it; the kamon it came from stays as it was.
func (k *kamonu) saveAs(e *editor) tea.Cmd {
	return k.open(&lineInput{
		title: "Save as", glyph: glyphSave, prompt: "Name for the new kamon; " + e.kamon + " stays as it was", verb: "save",
		check: k.st.checkName,
		done: func(v string) tea.Cmd {
			if err := k.st.create(v, e.km); err != nil {
				return k.open(errorPopup("Save as", "Can't save "+v+": "+why(err), k.width()))
			}
			k.reload(v)
			return k.closeAll()
		},
	})
}
