package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	e.onButtons = true
	e.Key("1", keyMsg("1"))
	if e.km.cells[0][0] != 0 {
		t.Fatal("a digit on the buttons painted the cell")
	}

	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("Enter", "?")
	hs.contains("1–3")
	hs.contains("make the cell that letter; again to empty it")
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
	frameSquare := terminu.Fg(frameColor).Render(glyphSquare + " ")
	before := strings.Count(strings.Join(hs.k.a.Frame(), "\n"), frameSquare)
	for range rows {
		hs.keys("j")
	}
	after := strings.Count(strings.Join(hs.k.a.Frame(), "\n"), frameSquare)
	if after <= before {
		t.Fatalf("the grid did not scroll down to the U frame's bottom: %d frame cells, then %d", before, after)
	}
	hs.contains("[3] " + glyphSquare + " L")
	hs.contains("Save As")
	hs.fullWidth()
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
// round, gg and G the first and the last (tdp v0.6.7, components/dialog/form).
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
	hs.contains("SAVE FAILED") // what failed (components/dialog/note, error popup)
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
	hs.contains("Press [1], then [A] to add one") // a key to press (layout/screen, empty states)
	hs.keys("A")
	hs.contains("NEW KAMON")
	hs.keys("Tab", "Enter")
	hs.contains("Name can't be empty")
	// The failed submit put the focus on Name (components/dialog/form).
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
	hs.keys("E")
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
	hs.keys("E", "j", "j", "Space", "Tab", "Enter")
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
	hs.contains("Ex[P]ort")
	hs.contains("[X] Delete")
	hs.keys("Esc", "P")
	hs.lacks("EXPORT: filu") // none there yet: nothing to ask
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

func TestExportWhenThere(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"filu": []byte(filYAML)})
	km := fil(t)
	hs.keys("P", "P")
	hs.contains("EXPORT: filu")
	hs.contains("filu_1.svg")
	hs.contains("Overwrite filu.svg")
	hs.contains("Enter:edit Tab:actions Esc:cancel")
	// Tab goes to Export, the main submit: a new file, the next number.
	hs.keys("Tab")
	hs.contains("Enter:export Tab:fields Esc:cancel")
	hs.keys("Enter")
	hs.lacks("EXPORT: filu")
	svgIs(t, hs, "filu_1", km)
	// A name of its own; one there is refused, with ".svg" or without.
	hs.keys("P")
	hs.contains("filu_2.svg")
	hs.keys("Enter")
	hs.contains("filu_2.svg") // the old value as the form shows it
	hs.keys("f", "i", "l", "u", ".", "s", "v", "g", "Enter")
	hs.contains("File is taken: filu.svg is already there") // names the value
	hs.keys("Backspace", "Backspace", "Backspace", "Backspace", "Backspace", "Backspace", "Backspace", "Backspace",
		"m", "a", "r", "k", "Enter", "Tab", "Enter")
	svgIs(t, hs, "mark", km)
	if hs.k.st.svgExists("filu_2") {
		t.Fatal("Export wrote the number it offered, not the name given")
	}
	// Overwrite, left of Export.
	if err := os.WriteFile(hs.k.st.svgPath("filu"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	hs.keys("P", "Tab", "h")
	hs.contains("Enter:overwrite filu.svg")
	hs.keys("Enter")
	hs.lacks("EXPORT: filu")
	svgIs(t, hs, "filu", km)
	// Esc writes nothing (the first Esc puts the toast away: tdp K4).
	hs.keys("Esc", "P", "Esc")
	hs.lacks("EXPORT: filu")
	if hs.k.st.svgExists("filu_2") {
		t.Fatal("Esc wrote an svg")
	}
}

// The name on Overwrite is the file's, in the hint too (tdp v0.6.9, KM6).
func TestExportOverwriteKeepsTheCase(t *testing.T) {
	hs := newHarness(t, 80, 40, map[string][]byte{"KBU": blank()})
	hs.keys("P", "Esc", "P") // Esc: the toast first (tdp K4)
	hs.contains("EXPORT: KBU")
	hs.contains("Overwrite KBU.svg")
	hs.keys("Tab", "h")
	hs.contains("Enter:overwrite KBU.svg")
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
	hs.keys("P", "Tab", "Enter")
	hs.contains("Can't export filu: permission denied")
	hs.keys("Esc")
	hs.contains("EXPORT: filu") // the form stays: Enter tries again
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
	hs.contains("Enter:draw")
	hs.keys("Space")
	hs.contains("[Enter] Draw")
	hs.keys("Esc", "Enter")
	hs.contains("DRAW: filu")

	hs = newHarness(t, 80, 40, nil)
	hs.keys("Tab")
	hs.lacks("Enter:draw")
	hs.keys("Enter")
	hs.lacks("DRAW")
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

// TestLettersPrompt: the prompt says whose letters they are (components/
// input, the one-line input).
func TestLettersPrompt(t *testing.T) {
	hs := newHarness(t, 100, 40, map[string][]byte{"filu": []byte(filYAML)})
	hs.keys("E", "j", "Enter")
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
		t.Fatal("the preview's failure is not in Red (layout/panel, errors)")
	}
	hs.keys("Enter") // on [2]: as it says, why
	hs.contains("DRAW FAILED")
	hs.contains("Can't read broken")
	hs.keys("Esc", "Tab", "Enter") // on [1] too
	hs.contains("DRAW FAILED")
	hs.keys("Esc", "E")
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
		{"Esc", "Esc", "E"},          // EDIT: terminu
		{"Esc", "P"},                 // exported: a toast
		{"Esc", "P"},                 // EXPORT: terminu, the svg there
		{"Esc", "X"},                 // delete
		{"Esc", "q"},                 // quit
	}
	sizes := [][2]int{{80, 40}, {120, 50}, {60, 30}, {56, 28}, {40, 20}, {120, 16}, {50, 50}}
	km := standIn(t)
	for _, widths := range [][2]int{{1, 1}, {2, 1}, {2, 2}} {
		t.Run(fmt.Sprintf("icon %d, powerline %d", widths[0], widths[1]), func(t *testing.T) {
			// The widths are the lib's, one for the whole program: set once
			// the harnesses are made (newHarness sets them back), and only
			// read while the sizes run side by side.
			var hss []*harness
			for _, size := range sizes {
				hss = append(hss, newHarness(t, size[0], size[1], map[string][]byte{"terminu": km.yaml(), "broken": []byte("cells: [")}))
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

// TestEditorBodyWidth is tdp E5 for the inside kamonu draws itself: every
// row exactly as wide as the lib asks for, never more rows than it has room
// for, however small the screen leaves it.
func TestEditorBodyWidth(t *testing.T) {
	km := standIn(t)
	for _, icon := range []int{1, 2} {
		terminu.SetIconWidth(icon, 1)
		for _, onButtons := range []bool{false, true} {
			e := newEditor(nil, "terminu", km)
			e.onButtons = onButtons
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
