package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	terminu "github.com/vulcanshen/terminu-go"
)

// filYAML is filu's mark: an F, an I and an L, 3 × 5 each, in the U frame.
const filYAML = `frame: true
letters: [F, I, L]
cells: |
  .........................
  .........................
  .........................
  .........................
  .........................
  .........................
  .......111.222.3.........
  .......1....2..3.........
  .......11...2..3.........
  .......1....2..3.........
  .......1...222.333.......
  .........................
  .........................
  .........................
  .........................
  .........................
  .........................
  .........................
  .........................
`

func fil(t *testing.T) kamon {
	t.Helper()
	km, err := parseKamon([]byte(filYAML))
	if err != nil {
		t.Fatal(err)
	}
	return km
}

// blank is a kamon with the U frame, one letter and no cells drawn.
func blank() []byte {
	km := kamon{frame: true, letters: []string{"K"}}
	return km.yaml()
}

// tdpMark is terminu's own mark, a U frame and a T: a copy of tdp's
// principle/icon.svg (tdp S5).
func tdpMark(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/tdp-icon.svg")
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// standIn is tdp's mark as a kamon.
func standIn(t *testing.T) kamon {
	t.Helper()
	m, err := parseMark(tdpMark(t))
	if err != nil {
		t.Fatal(err)
	}
	km := kamon{frame: true, letters: []string{"T"}}
	for r := range rows {
		for c := range cols {
			if m[r][c] == gold {
				km.cells[r][c] = 1
			}
		}
	}
	return km
}

// count is how many cells are letter n (0: none).
func count(km kamon, n uint8) int {
	got := 0
	for r := range rows {
		for c := range cols {
			if km.cells[r][c] == n {
				got++
			}
		}
	}
	return got
}

// --- the yaml ---------------------------------------------------------------

func TestYamlRoundTrip(t *testing.T) {
	km := fil(t)
	if string(km.yaml()) != filYAML {
		t.Fatalf("written back differently:\n%s", km.yaml())
	}
	if count(km, 1) != 8 || count(km, 2) != 9 || count(km, 3) != 7 {
		t.Fatalf("the cells of F, I, L: %d, %d, %d", count(km, 1), count(km, 2), count(km, 3))
	}
}

func TestYamlLettersInCapitals(t *testing.T) {
	km, err := parseKamon([]byte(strings.Replace(filYAML, "[F, I, L]", "[f, Y, n]", 1)))
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(km.letters, ""); got != "FYN" {
		t.Fatalf("letters %q", got)
	}
}

// TestParseKamonRejects: a hand edit that goes wrong is an error, never a
// different drawing.
func TestParseKamonRejects(t *testing.T) {
	row7 := "  .......1....2..3.........\n"
	for name, c := range map[string]struct{ from, to, want string }{
		"tab":               {row7, "\t.......1....2..3.........\n", "tab"},
		"less indent":       {row7, " .......1....2..3.........\n", "yaml"},
		"more indent":       {row7, "   .......1....2..3.........\n", "Row 8 has 26 cells, not 25"},
		"a cell missing":    {row7, "  .......1....2..3........\n", "Row 8 has 24 cells, not 25"},
		"trailing space":    {row7, "  .......1....2..3......... \n", "Row 8 has 26 cells, not 25"},
		"a fourth letter":   {row7, "  .......1....2..4.........\n", `Row 8, column 16 is '4': a cell is . or 1 to 3`},
		"on the U frame":    {row7, "  ....1..1....2..3.........\n", "Row 8, column 5 is on the U frame"},
		"a row missing":     {row7, "", "cells has 18 rows, not 19"},
		"no letters":        {"[F, I, L]", "[]", "Letters can't be empty"},
		"seven letters":     {"[F, I, L]", "[A, B, C, D, E, F, G]", "Letters can't be more than 6"},
		"not a letter":      {"[F, I, L]", "[F, '1', L]", "Letters can only be A to Z"},
		"two in one":        {"[F, I, L]", "[F, IL]", `Letter "IL" is not one letter`},
		"an unknown field":  {"frame:", "frmae:", "frmae"},
		"letters not lists": {"[F, I, L]", "FIL", "cannot unmarshal"},
	} {
		src := strings.Replace(filYAML, c.from, c.to, 1)
		if src == filYAML {
			t.Fatalf("%s: the edit did not apply", name)
		}
		_, err := parseKamon([]byte(src))
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", name, err, c.want)
		}
	}
	if _, err := parseKamon(nil); err == nil || err.Error() != "the file is empty" {
		t.Errorf("empty: %v", err)
	}
}

func TestCheckLetters(t *testing.T) {
	for v, want := range map[string]string{
		"fil": "", "KAMON": "", "abcdef": "", "abcdefg": "Letters can't be more than 6",
		"a1": "Letters can only be A to Z", "a b": "Letters can only be A to Z", "é": "Letters can only be A to Z",
	} {
		got := ""
		if err := checkLetters(v); err != nil {
			got = err.Error()
		}
		if got != want {
			t.Errorf("%q: %q, want %q", v, got, want)
		}
	}
}

// TestSetLetters: a cell keeps its place in the order; a place no longer
// there loses its cells.
func TestSetLetters(t *testing.T) {
	km := fil(t)
	km.setLetters([]string{"F", "A", "L"})
	if count(km, 2) != 9 {
		t.Fatal("FIL to FAL lost the middle letter's cells")
	}
	km.setLetters([]string{"K", "B"})
	if count(km, 1) != 8 || count(km, 2) != 9 || count(km, 3) != 0 {
		t.Fatal("FAL to KB: the L's cells should be gone, the others kept")
	}
	km.setLetters([]string{"F", "I", "L", "U"})
	if count(km, 3) != 0 || count(km, 4) != 0 {
		t.Fatal("new letters have cells")
	}
}

func TestSetFrame(t *testing.T) {
	km := kamon{letters: []string{"K"}}
	km.cells[0][2], km.cells[0][5] = 1, 1 // (0, 2) is the frame's
	km.setFrame(true)
	if km.cells[0][2] != 0 || km.cells[0][5] != 1 {
		t.Fatal("putting the U frame in should clear the cells under it, and only those")
	}
	km.setFrame(false)
	if km.frame || km.cells[0][5] != 1 {
		t.Fatal("taking the U frame out changed the cells")
	}
}

// --- the svg (tdp S5) -------------------------------------------------------

func TestUFrameIsTdps(t *testing.T) {
	tdp, err := parseMark(tdpMark(t))
	if err != nil {
		t.Fatal(err)
	}
	for r := range rows {
		for c := range cols {
			if (tdp[r][c] == frame) != (uFrame[r][c] == frame) {
				t.Fatalf("row %d, column %d: the U frame differs from tdp's mark", r, c)
			}
		}
	}
	for r := range rows {
		for c := range cols {
			if uFrame[r][c] == gold {
				t.Fatal("the U frame has a gold cell")
			}
		}
	}
}

// TestIconIsTheSplash is tdp E6: the splash draws docs/icon.svg, kamonu's
// own mark as tdp S5 has it, the family's U frame and then K, A, M, O, N; the
// copy in the program is the same, byte for byte (tdp project, 目錄).
func TestIconIsTheSplash(t *testing.T) {
	b, err := os.ReadFile("../../docs/icon.svg")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, icon) {
		t.Fatal("the splash does not draw docs/icon.svg")
	}
	if _, err := terminu.NewSplash(icon, "kamonu", "dev", "intro"); err != nil {
		t.Fatalf("the lib's splash does not take docs/icon.svg: %v", err)
	}
	m, err := parseMark(icon)
	if err != nil {
		t.Fatal(err)
	}
	for r := range rows {
		for c := range cols {
			if (m[r][c] == frame) != (uFrame[r][c] == frame) {
				t.Fatalf("row %d, column %d: the U frame differs from the family's", r, c)
			}
		}
	}
	// The letters split as tdp S5 has it: a comment each, in the order shown.
	var comments []string
	for _, l := range strings.Split(string(icon), "\n") {
		if c, ok := strings.CutPrefix(strings.TrimSpace(l), "<!-- "); ok {
			comments = append(comments, strings.TrimSuffix(c, " -->"))
		}
	}
	if got := strings.Join(comments, " "); got != "U K A M O N" {
		t.Fatalf("comments %q, want U K A M O N", got)
	}
}

func TestSvg(t *testing.T) {
	km := fil(t)
	km.setLetters([]string{"F", "I", "L", "U"}) // U has no cells
	svg := string(km.svg())
	// The comments: the U frame first, then the letters in their order; a
	// letter with no cells is left out.
	var comments []string
	for _, l := range strings.Split(svg, "\n") {
		if c, ok := strings.CutPrefix(strings.TrimSpace(l), "<!-- "); ok {
			comments = append(comments, strings.TrimSuffix(c, " -->"))
		}
	}
	if got := strings.Join(comments, " "); got != "U F I L" {
		t.Fatalf("comments %q", got)
	}
	for _, want := range []string{`viewBox="0 0 250 250"`, `<rect x="0" y="0" width="250" height="250" fill="#000000"/>`,
		"<!-- F -->\n" + `    <rect x="70" y="90" width="10" height="10" fill="#f2b753"/>`} {
		if !strings.Contains(svg, want) {
			t.Fatalf("the svg has no %s:\n%s", want, svg)
		}
	}
	// Every letter gold, the U frame as tdp's.
	m, err := parseMark(km.svg())
	if err != nil {
		t.Fatal(err)
	}
	if m != km.picture() {
		t.Fatal("the svg is not the picture")
	}
	if _, err := terminu.NewSplash(km.svg(), "kamonu", "dev", "intro"); err != nil {
		t.Fatalf("the lib's splash does not take what kamonu writes: %v", err)
	}
}

func TestSvgWithoutTheFrame(t *testing.T) {
	km := fil(t)
	km.setFrame(false)
	svg := string(km.svg())
	if strings.Contains(svg, "<!-- U -->") || strings.Contains(svg, "#205090") {
		t.Fatalf("the U frame went in:\n%s", svg)
	}
}

func TestParseRejects(t *testing.T) {
	for _, svg := range []string{
		`<svg><rect x="5" y="30" width="10" height="10" fill="#f2b753"/></svg>`,
		`<svg><rect x="0" y="20" width="10" height="10" fill="#f2b753"/></svg>`,
		`<svg><rect x="250" y="30" width="10" height="10" fill="#f2b753"/></svg>`,
		`<svg><rect x="0" y="30" width="10" height="10" fill="#ff0000"/></svg>`,
		`<svg><rect x="0" y="30" width="20" height="10" fill="#f2b753"/></svg>`,
		`<svg><rect`,
	} {
		if _, err := parseMark([]byte(svg)); err == nil {
			t.Errorf("%s: no error", svg)
		}
	}
}

// --- names ------------------------------------------------------------------

func TestCheckName(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"webu": blank()})
	st := hs.k.st
	for name, want := range map[string]string{
		"": "Name can't be empty", "a/b": "Name can't have /", ".x": "Name can't start with .",
		"webu": "Name is taken", "kamonu": "",
	} {
		got := ""
		if err := st.checkName(name); err != nil {
			got = err.Error()
		}
		if got != want {
			t.Errorf("%q: %q, want %q", name, got, want)
		}
	}
}

func TestNamesSorted(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"webu": nil, "Kbu": nil, "filu": nil})
	if err := os.WriteFile(hs.k.st.conf()+"/notes.txt", nil, 0o644); err != nil {
		t.Fatal(err)
	}
	names, err := hs.k.st.names()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(names, " "); got != "filu Kbu webu" {
		t.Fatalf("names %q", got)
	}
}

// --- the editor -------------------------------------------------------------

func TestEditorSkipsTheFrame(t *testing.T) {
	e := newEditor(nil, "x", kamon{frame: true, letters: []string{"K"}})
	if e.r != 0 || e.c != 0 {
		t.Fatalf("opens on %d,%d, want the first cell", e.r, e.c)
	}
	// Row 1: columns 3 and 4 are the frame's.
	e.r, e.c = 1, 2
	e.move(0, 1)
	if e.c != 5 {
		t.Fatalf("l from column 2 lands on column %d, want 5 (past the frame)", e.c)
	}
	// Column 1: free down to row 17, the frame's in row 18.
	e.r, e.c = 16, 1
	e.move(1, 0)
	if e.r != 17 || e.c != 1 {
		t.Fatalf("j lands on %d,%d", e.r, e.c)
	}
	e.move(1, 0) // nothing below it but the frame: it stays
	if e.r != 17 {
		t.Fatalf("j onto the frame moved to row %d", e.r)
	}
	// Column 2 is the frame's in rows 17 and 18: j from row 16 stays.
	e.r, e.c = 16, 2
	e.move(1, 0)
	if e.r != 16 {
		t.Fatalf("j with only the frame below moved to row %d", e.r)
	}
	// The edges stop the cursor (tdp K12).
	e.r, e.c = 0, 0
	e.move(0, -1)
	e.move(-1, 0)
	if e.r != 0 || e.c != 0 {
		t.Fatalf("moved off the grid to %d,%d", e.r, e.c)
	}
	// Without the U frame, every cell is the user's.
	e = newEditor(nil, "x", kamon{letters: []string{"K"}})
	e.r, e.c = 1, 2
	e.move(0, 1)
	if e.c != 3 {
		t.Fatalf("without the frame, l from column 2 lands on column %d", e.c)
	}
}

func TestEditorCycle(t *testing.T) {
	e := newEditor(nil, "x", kamon{letters: []string{"F", "I", "L"}})
	for _, want := range []uint8{1, 2, 3, 0, 1} {
		e.cycle()
		if got := e.km.cells[0][0]; got != want {
			t.Fatalf("the cell is %d, want %d", got, want)
		}
	}
	if !e.dirty {
		t.Fatal("cycling did not mark the drawing changed")
	}
}

func TestEditorClear(t *testing.T) {
	e := newEditor(nil, "x", kamon{letters: []string{"F", "I", "L"}})
	e.clear()
	if e.dirty {
		t.Fatal("emptying an empty cell marked the drawing changed")
	}
	e.cycle()
	e.cycle()
	e.clear()
	if e.km.cells[0][0] != 0 || !e.dirty {
		t.Fatal("Backspace did not empty the cell")
	}

	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank()})
	hs.keys("Enter", "Space", "Backspace", "?")
	hs.contains("empty the cell")
	hs.keys("Esc", "Tab", "Backspace", "Enter") // on the buttons Backspace does nothing
	if km, _ := hs.k.st.load("blank"); count(km, 1) != 0 {
		t.Fatal("the cell Backspace emptied was saved")
	}
}

func TestEditorShowsTheLetters(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("Enter")
	hs.contains("DRAW: filu")
	hs.contains("[1] " + glyphSquare + " F  [2] " + glyphSquare + " I  [3] " + glyphSquare + " L")
	hs.contains("Enter|Space:cycle Tab:actions Esc:cancel")
	// Each letter in its own colour.
	for _, l := range hs.k.a.Frame() {
		if strings.Contains(terminu.StripStyles(l), "[1] "+glyphSquare) {
			for i, c := range letterColors[:3] {
				if !strings.Contains(l, terminu.Fg(c).Render(glyphSquare)) {
					t.Fatalf("letter %d is not drawn in its colour", i+1)
				}
			}
			return
		}
	}
	t.Fatal("no row of letters")
}

func TestEditorDigits(t *testing.T) {
	e := newEditor(nil, "x", kamon{letters: []string{"F", "I", "L"}})
	for _, c := range []struct {
		key  string
		want uint8
	}{{"2", 2}, {"3", 3}, {"4", 3}, {"6", 3}, {"3", 0}, {"3", 3}, {"1", 1}, {"Backspace", 0}} {
		e.Key(c.key, keyMsg(c.key))
		if got := e.km.cells[0][0]; got != c.want {
			t.Fatalf("after %s the cell is %d, want %d", c.key, got, c.want)
		}
	}
	e = newEditor(nil, "x", kamon{letters: []string{"F", "I", "L"}})
	e.Key("4", keyMsg("4")) // FIL has no fourth letter
	if e.dirty {
		t.Fatal("a digit past the letters changed the drawing")
	}

	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML), "blank": blank()})
	hs.keys("j", "Enter", "?")
	hs.contains("1–3")
	hs.contains("make the cell that letter; again to empty it")
	hs.keys("Esc", "Esc", "k", "Enter", "Tab", "1", "Enter") // on the buttons a digit paints nothing
	if km, _ := hs.k.st.load("blank"); count(km, 1) != 0 {
		t.Fatal("a digit on the buttons painted the cell")
	}
}

// TestEditorGold: p shows the letters gold, as the splash will, and again
// in their colours; the drawing does not change.
func TestEditorGold(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("Enter")
	frame := func() string { return strings.Join(hs.k.a.Frame(), "\n") }
	gold, f := terminu.Fg(goldColor).Render(glyphSquare+" "), terminu.Fg(letterColors[0]).Render(glyphSquare+" ")
	if strings.Contains(frame(), gold) || !strings.Contains(frame(), f) {
		t.Fatal("the editor does not open in the letters' colours")
	}
	hs.keys("p")
	if !strings.Contains(frame(), gold) || strings.Contains(frame(), f) {
		t.Fatal("p did not turn every letter gold")
	}
	hs.keys("2") // drawing goes on in gold
	hs.keys("p")
	if strings.Contains(frame(), gold) || !strings.Contains(frame(), f) {
		t.Fatal("p again did not bring the colours back")
	}
	hs.keys("?")
	hs.contains("every letter gold, as the svg has them, or back")
	hs.lacks("splash") // tdp S2: not in a key reference
	hs.keys("Esc", "Tab", "Enter")
	if km, _ := hs.k.st.load("filu"); km.cells[0][0] != 2 {
		t.Fatal("a digit did not paint in gold view")
	}
}

// TestShortScreen: on a screen 28 rows high the editor is one row short;
// the blank row goes, the letters and the action row stay.
func TestShortScreen(t *testing.T) {
	hs := newHarness(t, 80, 28, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("Enter")
	hs.contains("[3] " + glyphSquare + " L")
	hs.contains("Save As")
	hs.fullWidth()
}

// TestDrawScrolls: on an 80 × 24 terminal the grid does not all fit; it
// scrolls with the cursor, and the letters and the buttons stay in view.
func TestDrawScrolls(t *testing.T) {
	hs := newHarness(t, 80, 24, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("Enter")
	hs.contains("[3] " + glyphSquare + " L")
	hs.contains("Save As")
	hs.contains("1 of 19") // the cursor's row, on the bottom frame line (KM9)
	frameSquare := terminu.Fg(frameColor).Render(glyphSquare + " ")
	before := strings.Count(strings.Join(hs.k.a.Frame(), "\n"), frameSquare)
	for range rows {
		hs.keys("j")
	}
	after := strings.Count(strings.Join(hs.k.a.Frame(), "\n"), frameSquare)
	if after <= before {
		t.Fatalf("the grid did not scroll down to the U frame's bottom: %d frame cells, then %d", before, after)
	}
	hs.lacks("1 of 19")
	hs.contains("[3] " + glyphSquare + " L")
	hs.contains("Save As")
	hs.fullWidth()
	hs.keys("Tab") // on the buttons there is no row to count
	hs.lacks(" of 19")
}

// TestEditorClearAll: x empties every cell, after a confirm; with nothing to
// empty it asks nothing.
func TestEditorClearAll(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML), "blank": blank()})
	hs.keys("Enter", "x")
	hs.lacks("Clear every cell") // nothing drawn: nothing to ask
	hs.keys("Esc")
	hs.lacks("DRAW: blank")

	hs.keys("j", "Enter", "x")
	hs.contains("Clear every cell of filu?")
	hs.contains("Enter:clear Esc:cancel")
	hs.keys("Esc")
	hs.lacks("Clear every cell")
	hs.keys("Esc") // Esc on the confirm changed nothing: Esc closes the drawing
	hs.lacks("Discard changes")
	hs.lacks("DRAW: filu")

	hs.keys("Enter", "x", "Enter")
	hs.lacks("Clear every cell")
	hs.contains("DRAW: filu") // back to the drawing, to draw again or save
	hs.keys("Esc")            // changed: Esc asks first
	hs.contains("Discard changes to filu?")
	hs.keys("Esc", "?")
	hs.contains("empty every cell (asks first)")
	hs.keys("Esc", "Tab", "Enter")
	if km, _ := hs.k.st.load("filu"); count(km, 0) != rows*cols {
		t.Fatal("the cleared drawing was not saved")
	}
}

// TestLettersFit: six letters fit under the cells with icons one cell wide
// and two, on the narrowest screen the tests use.
func TestLettersFit(t *testing.T) {
	km := kamon{frame: true, letters: []string{"K", "A", "M", "O", "N", "U"}}
	for _, icon := range []int{1, 2} {
		hs := newHarness(t, 56, 28, map[string][]byte{"six": km.yaml()})
		terminu.SetIconWidth(icon, 1)
		hs.keys("Enter")
		hs.fullWidth()
		_, last, ok := strings.Cut(hs.text(), "[6] ")
		if !ok || !strings.HasPrefix(strings.TrimLeft(strings.TrimPrefix(last, glyphSquare), " "), "U") {
			t.Fatalf("icon width %d: the sixth letter is cut:\n%s", icon, hs.text())
		}
	}
	terminu.SetIconWidth(1, 1)
}

func TestSave(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank(), "other": blank()})
	hs.keys("Enter")
	hs.contains("DRAW: blank")
	hs.keys("Space", "Tab", "Enter")
	hs.lacks("DRAW: blank")
	km, err := hs.k.st.load("blank")
	if err != nil {
		t.Fatal(err)
	}
	if km.cells[0][0] != 1 || count(km, 1) != 1 {
		t.Fatal("the saved kamon does not have the cell")
	}
	if hs.k.current() != "blank" {
		t.Fatalf("the cursor is on %q", hs.k.current())
	}
}

func TestEscAsksWhenChanged(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank()})
	hs.keys("Enter", "Esc")
	hs.lacks("DRAW: blank") // not changed: Esc closes it

	hs.keys("Enter", "Space", "Esc")
	hs.contains("Discard changes to blank?")
	hs.contains("Enter:discard Esc:cancel")
	hs.keys("Esc") // back to the editor, the change still there
	hs.lacks("Discard changes")
	hs.contains("DRAW: blank")
	hs.keys("Esc", "Enter")
	hs.lacks("DRAW: blank")
	if km, _ := hs.k.st.load("blank"); count(km, 1) != 0 {
		t.Fatal("a discarded change was written")
	}
}

// The buttons move as a form's action area: hjkl the one before or after,
// round, gg and G the first and the last (tdp form.9).
func TestEditorButtons(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank()})
	hs.keys("Enter", "Tab")
	hs.contains("Enter:save Tab:grid") // Save first
	for _, step := range []struct{ key, hint string }{
		{"j", "Enter:save as Tab:grid"}, // after the last, round to the first
		{"j", "Enter:save Tab:grid"},
		{"k", "Enter:save as Tab:grid"},
		{"k", "Enter:save Tab:grid"}, // before the first, round to the last
		{"h", "Enter:save as Tab:grid"},
		{"l", "Enter:save Tab:grid"},
		{"gg", "Enter:save as Tab:grid"},
		{"G", "Enter:save Tab:grid"},
	} {
		if step.key == "gg" {
			hs.keys("g", "g")
		} else {
			hs.keys(step.key)
		}
		hs.contains(step.hint)
	}
	// Tab goes back to the button left.
	hs.keys("h", "Tab")
	hs.contains("Enter|Space:cycle")
	hs.keys("Tab")
	hs.contains("Enter:save as Tab:grid")
}

func TestSaveAs(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank(), "webu": blank()})
	hs.keys("Enter", "Space", "Tab", "h")
	hs.contains("Enter:save as Tab:grid Esc:cancel")
	hs.keys("Enter")
	hs.contains("SAVE AS")
	hs.keys("w", "e", "b", "u", "Enter")
	hs.contains("Name is taken")
	hs.keys("Backspace", "Backspace", "Backspace", "Backspace")
	hs.paste("k\x1bu")
	hs.keys("Enter")
	hs.contains("Name can't have control characters")
	hs.keys("Backspace", "Backspace", "Backspace", "k", "b", "u", "Enter")
	hs.lacks("SAVE AS")
	hs.lacks("DRAW: blank")
	km, err := hs.k.st.load("kbu")
	if err != nil {
		t.Fatal(err)
	}
	if count(km, 1) != 1 || !km.frame || strings.Join(km.letters, "") != "K" {
		t.Fatal("the new kamon does not have the drawing, the letters and the U frame")
	}
	if old, _ := hs.k.st.load("blank"); count(old, 1) != 0 {
		t.Fatal("Save As changed the kamon it came from")
	}
	if hs.k.current() != "kbu" {
		t.Fatalf("the cursor is on %q, want the new kamon", hs.k.current())
	}
}

func TestSaveFailureOpensAnError(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank()})
	hs.keys("Enter", "Space", "Tab")
	if err := os.Chmod(hs.k.st.conf(), 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(hs.k.st.conf(), 0o755)
	hs.keys("Enter")
	hs.contains("SAVE FAILED: blank") // what failed, on which kamon (tdp note.6)
	hs.contains("Can't save blank")
	hs.contains("Enter|Esc:close")
	hs.keys("Enter") // closes the error popup, back on the button
	hs.contains("DRAW: blank")
	hs.contains("Enter:save Tab:grid")
}

func TestEditorKeyReference(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank()})
	hs.keys("Enter", "?")
	hs.contains("DRAW KEYS")
	hs.contains("the cell's next letter; after the last, none")
	hs.keys("?")
	hs.lacks("DRAW KEYS")
	hs.keys("q", "g") // q is no quit on a form; a lone g does nothing
	hs.lacks("QUIT")
	hs.contains("DRAW: blank")
}

func TestQuitFromTheEditor(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank()})
	hs.keys("Enter", "Space", "Ctrl-C")
	hs.contains("Quit kamonu?")
	hs.contains("Changes to blank are not saved.")
	hs.keys("Esc")
	hs.lacks("Quit kamonu?")
	hs.keys("Ctrl-C", "Ctrl-C")
	if !hs.quit {
		t.Fatal("the second Ctrl-C did not quit")
	}
}

func TestNoKeysWhileOpening(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank()})
	opening := hs.k.draw() // its animation has not run yet
	hs.k.a.Update(keyMsg("Space"))
	hs.run(opening)
	hs.keys("Esc") // the Space changed nothing: Esc closes it without asking
	if strings.Contains(hs.text(), "Discard changes") {
		t.Fatal("a key reached the editor while it was opening (tdp F2)")
	}
	hs.lacks("DRAW: blank")
}

// --- the screen -------------------------------------------------------------

func TestNewKamon(t *testing.T) {
	hs := newHarness(t, 80, 40, nil)
	hs.contains("No kamon yet")
	hs.contains("Press [A] to add one")
	hs.contains("Nothing to preview")
	hs.contains("Press [1], then [A] to add one") // a key to press (tdp screen.8)
	hs.keys("A")
	hs.contains("NEW KAMON")
	hs.keys("Tab", "Enter")
	hs.contains("Name can't be empty")
	// The failed submit put the focus on Name (tdp form.16).
	hs.keys("Enter", "k", "a", "m", "o", "n", "u", "Enter", "Tab", "Enter")
	hs.contains("Letters can't be empty")
	// ... and now on Letters.
	hs.keys("Enter", "a", "b", "c", "d", "e", "f", "g", "Enter")
	hs.contains("Letters can't be more than 6")
	hs.keys("Backspace", "Backspace", "Backspace", "Backspace", "Backspace", "Backspace", "Backspace", "1", "Enter")
	hs.contains("Letters can only be A to Z")
	hs.keys("Backspace", "k", "a", "m", "o", "n", "Enter", "Tab", "Enter")
	hs.lacks("NEW KAMON")
	hs.contains("DRAW: kamonu")
	km, err := hs.k.st.load("kamonu")
	if err != nil {
		t.Fatalf("Create did not make the file: %v", err)
	}
	if !km.frame || strings.Join(km.letters, "") != "KAMON" || count(km, 0) != rows*cols {
		t.Fatalf("the new kamon: frame %t, letters %v", km.frame, km.letters)
	}
}

func TestNewKamonWithoutTheFrame(t *testing.T) {
	hs := newHarness(t, 80, 40, nil)
	hs.keys("A", "Enter", "x", "Enter", "j", "Enter", "k", "Enter", "j", "Space", "Tab", "Enter")
	km, err := hs.k.st.load("x")
	if err != nil {
		t.Fatal(err)
	}
	if km.frame {
		t.Fatal("the U frame went in though Base was unticked")
	}
}

func TestEditKamon(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML), "webu": blank()})
	hs.keys("e")
	hs.contains("EDIT: filu")
	hs.contains("FIL")
	// A name that is taken.
	hs.keys("Enter", "w", "e", "b", "u", "Enter", "Tab", "Enter")
	hs.contains("Name is taken")
	hs.keys("Enter", "Backspace", "Backspace", "Backspace", "Backspace", "k", "b", "u", "Enter")
	hs.keys("j", "Enter", "k", "b", "Enter", "Tab", "Enter")
	hs.lacks("EDIT: ")
	if hs.k.st.exists("filu") {
		t.Fatal("the old name is still there")
	}
	km, err := hs.k.st.load("kbu")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(km.letters, "") != "KB" || count(km, 1) != 8 || count(km, 2) != 9 || count(km, 3) != 0 {
		t.Fatalf("FIL to KB: letters %v, cells %d %d %d", km.letters, count(km, 1), count(km, 2), count(km, 3))
	}
	if hs.k.current() != "kbu" {
		t.Fatalf("the cursor is on %q", hs.k.current())
	}
}

func TestEditPutsTheFrameIn(t *testing.T) {
	km := kamon{letters: []string{"K"}}
	km.cells[0][2], km.cells[0][5] = 1, 1
	hs := newHarness(t, 80, 40, map[string][]byte{"k": km.yaml()})
	hs.keys("e", "j", "j", "Space", "Tab", "Enter")
	got, err := hs.k.st.load("k")
	if err != nil {
		t.Fatal(err)
	}
	if !got.frame || got.cells[0][2] != 0 || got.cells[0][5] != 1 {
		t.Fatal("the U frame did not go in, or did not clear the cell under it")
	}
}

func TestExport(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("Space")
	// The label is Export, its key a capital P (tdp M5; KM16): "[P] Export"
	// on a lib that matches the key's case only, Ex[P]ort on one that does not.
	if !strings.Contains(hs.text(), "[P] Export") && !strings.Contains(hs.text(), "Ex[P]ort") {
		t.Fatalf("no Export row with its key:\n%s", hs.text())
	}
	hs.lacks("ExPort")
	hs.contains("[X] Delete")
	hs.keys("Esc", "P")
	hs.lacks("Overwrite filu.svg") // none there yet: nothing to ask
	hs.contains("Exported to")
	svgIs(t, hs, "filu", fil(t))
}

// svgIs fails unless svg/<file>.svg is km's.
func svgIs(t *testing.T, hs *harness, file string, km kamon) {
	t.Helper()
	b, err := os.ReadFile(hs.k.st.svgPath(file))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, km.svg()) {
		t.Fatalf("%s.svg is not the kamon's", file)
	}
}

// TestExportWhenThere: with <name>.svg there, a menu asks: Overwrite it,
// or Export as new, which asks for the name (KM10, decided 2026-10-19).
func TestExportWhenThere(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	km := fil(t)
	hs.keys("P", "Esc", "P") // Esc: the toast of the first export (tdp K4)
	hs.contains("EXPORT")
	hs.contains("Overwrite filu.svg")
	hs.contains("Export as new…")
	hs.contains("EXPORT: filu") // the name as it is (MenuAbout)
	// Export as new: the next number suggested, taken with Enter alone
	// (tdp input.6); Esc goes back to the menu (tdp T1).
	hs.keys("j", "Enter")
	hs.contains("EXPORT AS")
	hs.contains("filu_1.svg")
	hs.keys("Esc")
	hs.contains("Overwrite filu.svg")
	hs.keys("Enter", "Enter") // the input again, the suggestion taken as it is
	hs.lacks("EXPORT AS")
	hs.lacks("Overwrite filu.svg") // written: the whole stack closes
	hs.contains("SUCCESS")         // a Success toast (KM15)
	svgIs(t, hs, "filu_1", km)
	// A name of its own; one there is refused, with ".svg" or without.
	hs.keys("Esc", "P", "j", "Enter")
	hs.contains("filu_2.svg")
	hs.keys("f", "i", "l", "u", ".", "s", "v", "g", "Enter")
	hs.contains("File is taken: filu.svg is already there") // names the value
	hs.keys("Backspace", "Backspace", "Backspace", "Backspace", "Backspace", "Backspace", "Backspace", "Backspace",
		"m", "a", "r", "k", "Enter")
	hs.lacks("EXPORT AS")
	svgIs(t, hs, "mark", km)
	// Overwrite, the first row.
	if err := os.WriteFile(hs.k.st.svgPath("filu"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	hs.keys("Esc", "P", "Enter") // Esc: the toast first (tdp K4)
	hs.lacks("Overwrite filu.svg")
	svgIs(t, hs, "filu", km)
	// Esc writes nothing.
	hs.keys("Esc", "P", "Esc")
	hs.lacks("Overwrite filu.svg")
	if hs.k.st.svgExists("filu_2") {
		t.Fatal("Esc wrote an svg")
	}
}

// The name on Overwrite is the file's, in its own case (KM6).
func TestExportOverwriteKeepsTheCase(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"KBU": blank()})
	hs.keys("P", "Esc", "P") // Esc: the toast first (tdp K4)
	hs.contains("Overwrite KBU.svg")
}

func TestExportFailure(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	if err := os.Chmod(hs.k.st.dir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(hs.k.st.dir, 0o755)
	hs.keys("P")
	hs.contains("EXPORT FAILED") // all of it, in an error popup, never cut as a toast would
	hs.contains("Can't export filu: permission denied")
}

func TestExportWhenThereFails(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("P")
	dir := filepath.Dir(hs.k.st.svgPath("filu"))
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0o755)
	hs.keys("P", "Enter") // Overwrite
	hs.contains("EXPORT FAILED")
	hs.contains("Can't export filu: permission denied")
	// Esc twice: the toast of the first export, then the error popup (tdp K4).
	// The error popup closes back to the menu (tdp popup.5).
	hs.keys("Esc", "Esc")
	hs.contains("Overwrite filu.svg")
	hs.keys("j", "Enter", "m", "a", "r", "k", "Enter") // Export as new
	hs.contains("FAILED")
	hs.contains("mark") // the name typed stays in the input under it (Commit)
	hs.contains("Can't export filu: permission denied")
}

func TestDelete(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank()})
	hs.keys("X")
	hs.contains("Delete blank?")
	hs.contains("blank.yaml is removed from")
	hs.keys("Enter")
	if hs.k.st.exists("blank") {
		t.Fatal("the file was not deleted")
	}
	hs.contains("No kamon yet")
	// Nothing to move among: no moving keys in [1]'s key reference (tdp M6).
	hs.keys("?")
	hs.lacks("half a page up, down")
}

// TestDeleteFromTheMenu: the confirm opened from the Space menu keeps the
// menu under it; Esc goes back to it (tdp T1, F4), and the question names
// the kamon under the cursor when the action runs.
func TestDeleteFromTheMenu(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank(), "filu": []byte(filYAML)})
	hs.keys("j", "Space", "X")
	hs.contains("Delete filu?")
	hs.keys("Esc")
	hs.lacks("Delete filu?")
	hs.contains("ITEM OPERATION") // the menu is still there
	hs.keys("X", "Enter")         // from the menu again, accepted: deleted, and the menu goes with it
	if !hs.k.st.exists("blank") || hs.k.st.exists("filu") {
		t.Fatal("not the kamon under the cursor was deleted")
	}
	hs.lacks("ITEM OPERATION")
	hs.keys("X") // on the panel, the kamon now under the cursor
	hs.contains("Delete blank?")
}

func TestPreviewIsGold(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	hs.frame()
	km := fil(t)
	if hs.k.preview.m != km.picture() || hs.k.preview.m[6][7] != gold {
		t.Fatal("the preview is not the mark as the splash draws it")
	}
}

// TestPreviewEnter: Enter on [2] PREVIEW draws the kamon it shows (tdp K3);
// with nothing shown, it is not there.
func TestPreviewEnter(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("Tab")
	hs.contains("Enter:show")
	hs.keys("Space")
	hs.contains("ITEM OPERATION") // one item, the kamon shown (KM8)
	hs.contains("[Enter] Show")
	hs.contains("[e]dit") // KM17
	// Enter plays the mark as its app's splash does, its name under it.
	hs.keys("Esc", "Enter")
	hs.contains("Press any key to close")
	hs.contains("filu")
	hs.lacks("[1] KAMON")
	hs.keys("j") // any key closes it
	hs.contains("[1] KAMON")
	hs.keys("!") // it only looks, and the App Log has it (tdp note.7)
	hs.contains("Showed filu")
	hs.keys("Esc", "e") // on [2], Edit draws
	hs.contains("DRAW: filu")

	hs = newHarness(t, 80, 40, nil)
	hs.keys("Tab")
	hs.lacks("Enter:show")
	hs.keys("Enter")
	hs.lacks("DRAW")
}

// TestSideBySide: at 60 columns both panels are side by side (tdp L1,
// screen.5), [1] 20 wide with whole names and the mark cut on the right
// with …; at 74 the whole mark fits, and under 60 only the focused panel is
// drawn.
func TestSideBySide(t *testing.T) {
	marks := map[string][]byte{"ccdrawu": []byte(filYAML), "filu": []byte(filYAML)}
	hs := newHarness(t, 60, 32, marks)
	hs.contains("[1] KAMON")
	hs.contains("[2] PREVIEW")
	hs.contains("ccdrawu ")
	hs.contains("…")
	row := hs.frame()[1] // [1]'s frame heavy, focused; [2]'s light
	if i := strings.Index(row, "│"); i < 0 || terminu.Width(row[:i]) != listWidth {
		t.Fatalf("[2] does not start after [1]'s %d columns:\n%s", listWidth, hs.text())
	}
	hs.fullWidth()

	hs = newHarness(t, 74, 32, marks)
	hs.lacks("…")

	hs = newHarness(t, 59, 32, marks)
	hs.contains("[1] KAMON")
	hs.lacks("[2] PREVIEW")
}

// TestPreviewScrolls: on a panel shorter than the mark, [2] PREVIEW
// scrolls and says which rows it shows; when the mark fits, it says
// nothing and lists no scrolling keys.
func TestPreviewScrolls(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("Tab", "?")
	hs.lacks("scroll down, up")
	hs.keys("Esc")
	hs.lacks(" of 21")

	hs = newHarness(t, 80, 16, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("Tab")
	hs.contains("1-13 of 21")
	hs.keys("j")
	hs.contains("2-14 of 21")
	hs.keys("G")
	hs.contains("9-21 of 21")
	hs.keys("j") // stops at the bottom
	hs.contains("9-21 of 21")
	hs.keys("?")
	hs.contains("scroll down, up")
}

// TestLettersPrompt: the prompt says whose letters they are, as a
// one-line input's does (tdp input.2).
func TestLettersPrompt(t *testing.T) {
	hs := newHarness(t, 100, 40, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("e", "j", "Enter")
	hs.contains("The letters in the U frame of filu, in the order of the app's name")
	hs.keys("Esc", "Esc", "A", "j", "Enter")
	hs.contains("The letters in the U frame of the new kamon, in the order of the app's name")
}

func TestBrokenFile(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"broken": []byte(strings.Replace(filYAML, "..3......", "..3.....", 1))})
	hs.contains("Can't draw broken")
	hs.contains("Press [Enter] to see why")
	hs.keys("Tab") // [2] focused: its colours undimmed
	if !strings.Contains(strings.Join(hs.k.a.Frame(), "\n"), terminu.Fg(terminu.Red).Render("Can't draw broken")) {
		t.Fatal("the preview's failure is not in Red (tdp panel.15)")
	}
	hs.keys("Enter") // on [2]: as it says, why
	hs.contains("SHOW FAILED: broken")
	hs.contains("Can't read broken")
	hs.keys("Esc", "Tab", "Enter") // on [1] too
	hs.contains("DRAW FAILED")
	hs.keys("Esc", "e")
	hs.contains("Can't read broken")
}

// TestErrorsWithoutPaths: a failed save names the kamon, not the temporary
// file beside it.
func TestErrorsWithoutPaths(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"blank": blank()})
	hs.keys("Enter", "Space", "Tab")
	if err := os.Chmod(hs.k.st.conf(), 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(hs.k.st.conf(), 0o755)
	hs.keys("Enter")
	hs.contains("Can't save blank: permission denied")
	hs.lacks(hs.k.st.dir)
}

// TestFullWidth is tdp L4 and E5: every line exactly as wide as the
// terminal, on the screen and every kind of popup kamonu opens, with icons
// and powerline symbols one cell wide, icons two, and both two; on screens
// the drawing fits, and ones it is wider than, taller than, or both.
func TestFullWidth(t *testing.T) {
	steps := [][]string{
		nil,                          // [1] KAMON, [2] PREVIEW of a kamon it cannot read
		{"Enter"},                    // an error popup: Can't read broken
		{"Esc", "j", "Enter"},        // DRAW: terminu
		{"?"},                        // its key reference
		{"Esc", "Space", "Esc"},      // a cell turned: Esc asks to discard
		{"Esc", "Tab", "h", "Enter"}, // SAVE AS
		{"Esc", "x"},                 // clear every cell
		{"Esc", "Ctrl-C"},            // the quit confirm, saying what is not saved
		{"Esc", "Esc", "Enter"},      // discarded: the screen again
		{"Space"},                    // the Space menu
		{"G", "Enter"},               // the global operation popup over it
		{"Esc", "Esc", "?"},          // the panel's key reference
		{"Esc", "A"},                 // NEW KAMON
		{"Enter"},                    // its Name
		{"Esc", "Esc", "e"},          // EDIT: terminu
		{"Esc", "P"},                 // exported: a toast
		{"Esc", "P"},                 // the svg there: the menu
		{"j", "Enter"},               // EXPORT AS
		{"Esc", "Esc", "X"},          // back to the menu, out of it: delete
		{"Esc", "q"},                 // quit
		{"Esc", "j", "Enter"},        // DRAW: zz, six letters wider than the grid
	}
	// The least size (56 × 16) and above it: the drawing wider and taller
	// than the screen (56 × 16), taller (120 × 16), wider (56 × 50, zz's
	// letters with icons two cells wide).
	sizes := [][2]int{{80, 40}, {120, 50}, {60, 30}, {56, 28}, {56, 16}, {120, 16}, {56, 50}}
	km := standIn(t)
	six := kamon{frame: true, letters: []string{"K", "A", "M", "O", "N", "U"}}
	for _, widths := range [][2]int{{1, 1}, {2, 1}, {2, 2}} {
		t.Run(fmt.Sprintf("icon %d, powerline %d", widths[0], widths[1]), func(t *testing.T) {
			// The widths are the lib's, one for the whole program: set once
			// the harnesses are made (newHarness sets them back), and only
			// read while the sizes run side by side.
			var hss []*harness
			for _, size := range sizes {
				hss = append(hss, newHarness(t, size[0], size[1], map[string][]byte{"terminu": km.yaml(), "broken": []byte("cells: ["), "zz": six.yaml()}))
			}
			terminu.SetIconWidth(widths[0], widths[1])
			for i, size := range sizes {
				hs := hss[i]
				t.Run(fmt.Sprintf("%d×%d", size[0], size[1]), func(t *testing.T) {
					t.Parallel()
					hs.t = t
					for j, s := range steps {
						hs.keys(s...)
						t.Logf("step %d", j)
						hs.fullWidth()
					}
				})
			}
		})
	}
	terminu.SetIconWidth(1, 1)
}

// TestLongName: a name longer than [1] is cut with … at its end, not at the
// frame (tdp panel.5).
func TestLongName(t *testing.T) {
	long := "a-very-long-name-for-a-mark-of-the-family"
	hs := newHarness(t, 80, 40, map[string][]byte{long: blank()})
	hs.contains("a-very-long-name")
	hs.contains("…")
	hs.lacks(long)
	hs.fullWidth()
}

// TestTooSmall: under kamonu's least size the lib draws only its name and
// the size it needs, every line as wide as the screen, down to 1 × 1, and
// the screen comes back as the terminal grows (tdp L1, E5).
func TestTooSmall(t *testing.T) {
	for _, size := range [][2]int{{55, 40}, {80, 15}, {80, 5}, {20, 3}, {1, 1}} {
		hs := newHarness(t, size[0], size[1], map[string][]byte{"filu": []byte(filYAML)})
		hs.fullWidth()
		hs.contains("k") // kamonu, or as much of it as fits
		if size[0] >= 40 && size[1] >= 2 {
			hs.contains("needs a terminal of at least 56×16")
		}
		hs.keys("Enter", "A", "q")
		hs.fullWidth()
		hs.send(tea.WindowSizeMsg{Width: 80, Height: 40})
		hs.contains("[1] KAMON")
		hs.lacks("DRAW")
		hs.lacks("NEW KAMON")
	}
}

// TestEditorBodyWidth is tdp E5 for the inside kamonu draws itself: every
// row exactly as wide as the lib asks for, never more rows than it has room
// for, however small the screen leaves it.
func TestEditorBodyWidth(t *testing.T) {
	km := standIn(t)
	for _, icon := range []int{1, 2} {
		terminu.SetIconWidth(icon, 1)
		for _, onGrid := range []bool{true, false} {
			e := newEditor(nil, "terminu", km)
			e.Focus(onGrid)
			for w := 1; w <= 2*cols+2; w++ {
				for _, h := range []int{0, 1, 5, rows + 3, rows + 4, rows + 10} {
					body := e.Draw(w, h, terminu.Lavender)
					if len(body) > h {
						t.Fatalf("icon %d, %d×%d: %d rows", icon, w, h, len(body))
					}
					for i, l := range body {
						if got := terminu.Width(l); got != w {
							t.Fatalf("icon %d, %d×%d: row %d is %d cells: %q", icon, w, h, i, got, terminu.StripStyles(l))
						}
					}
				}
			}
		}
	}
	terminu.SetIconWidth(1, 1)
}
