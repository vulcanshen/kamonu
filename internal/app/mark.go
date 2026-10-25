package app

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	terminu "github.com/vulcanshen/terminu-go"
	"gopkg.in/yaml.v3"
)

// A mark is the family's 25 × 19 cells (tdp S5): x 0–240, y 30–210 in a
// 250 × 250 viewBox, a cell a 10 × 10 rect.
const (
	cols, rows = 25, 19
	cellSize   = 10
	originY    = 30 // the y of the mark's first row
)

// cell is what one cell of a mark is, as the splash draws it.
type cell byte

const (
	back  cell = iota // the background: no rect
	frame             // the U frame, #205090
	gold              // a letter, #f2b753
)

// The mark's colours (tdp S5).
var (
	frameColor = terminu.Hex("#205090")
	goldColor  = terminu.Hex("#f2b753")
	backColor  = terminu.Hex("#313244") // Surface0: the background squares
)

func (c cell) color() terminu.Color {
	switch c {
	case frame:
		return frameColor
	case gold:
		return goldColor
	}
	return backColor
}

// mark is a mark's cells, row by row, as the splash draws them.
type mark [rows][cols]cell

// uFrame is the family's U frame (tdp S5: the same in every app). A kamon
// does not hold its cells, only whether it goes in, so it cannot be drawn
// over or deleted.
//
//go:embed u-frame.svg
var uFrameSVG []byte

var uFrame = func() mark {
	m, err := parseMark(uFrameSVG)
	if err != nil {
		panic(err) // the embedded U frame is broken: a mistake in the program
	}
	return m
}()

// parseMark reads an svg as tdp S5 writes a mark. The background rect and
// XML comments are skipped; any other rect must be a #205090 or #f2b753 cell
// on the grid.
func parseMark(svg []byte) (mark, error) {
	var m mark
	d := xml.NewDecoder(bytes.NewReader(svg))
	for {
		tok, err := d.Token()
		if err == io.EOF {
			return m, nil
		}
		if err != nil {
			return m, err
		}
		t, ok := tok.(xml.StartElement)
		if !ok || t.Name.Local != "rect" {
			continue
		}
		attr := map[string]string{}
		for _, a := range t.Attr {
			attr[a.Name.Local] = a.Value
		}
		fill := strings.ToLower(attr["fill"])
		if fill == "#000000" {
			continue
		}
		x, errX := strconv.Atoi(attr["x"])
		y, errY := strconv.Atoi(attr["y"])
		if errX != nil || errY != nil || attr["width"] != "10" || attr["height"] != "10" ||
			x%cellSize != 0 || y%cellSize != 0 || x < 0 || x >= cols*cellSize || y < originY || y >= originY+rows*cellSize {
			return m, fmt.Errorf("a cell off the 25 × 19 grid (x %q, y %q)", attr["x"], attr["y"])
		}
		c := &m[(y-originY)/cellSize][x/cellSize]
		switch fill {
		case "#205090":
			*c = frame
		case "#f2b753":
			*c = gold
		default:
			return m, fmt.Errorf("a cell in %s: a mark has only #205090 and #f2b753", fill)
		}
	}
}

// maxLetters is the most letters a kamon has: six 3 × 5 letters fit in the
// U frame with room around them, and six colours tell them apart
// (docs/dev-remarks.md, 分字).
const maxLetters = 6

// letterColors tell a kamon's letters apart in kamonu, the first letter the
// first colour; the svg draws them all gold (tdp S5). Warm and cool take
// turns, so letters next to each other differ the most.
var letterColors = [maxLetters]terminu.Color{
	terminu.Peach,
	terminu.Hex("#94e2d5"), // Teal
	terminu.Hex("#cba6f7"), // Mauve
	terminu.Green,
	terminu.Red,
	terminu.Yellow,
}

// kamon is what conf/<name>.yaml holds (docs/dev-remarks.md, 存檔格式):
// whether the U frame goes in, the letters in the order the splash shows
// them, and which letter each cell is, 1 for the first, 0 for none.
type kamon struct {
	frame   bool
	letters []string
	cells   [rows][cols]uint8
}

// onFrame reports a cell of the U frame, when it goes in.
func (km *kamon) onFrame(r, c int) bool { return km.frame && uFrame[r][c] == frame }

// color is how kamonu draws a cell: the U frame, its letter's colour, or
// the background.
func (km *kamon) color(r, c int) terminu.Color {
	if km.onFrame(r, c) {
		return frameColor
	}
	if n := km.cells[r][c]; n > 0 {
		return letterColors[n-1]
	}
	return backColor
}

// picture is the kamon as the splash draws it: the U frame, every letter
// gold.
func (km *kamon) picture() mark {
	var m mark
	for r := range rows {
		for c := range cols {
			switch {
			case km.onFrame(r, c):
				m[r][c] = frame
			case km.cells[r][c] > 0:
				m[r][c] = gold
			}
		}
	}
	return m
}

// setLetters gives the kamon new letters. A cell keeps its place in the
// order: from FIL to FAL the middle letter's cells are the A's; the cells of
// a place no longer there turn back to the background.
func (km *kamon) setLetters(letters []string) {
	km.letters = letters
	for r := range rows {
		for c := range cols {
			if int(km.cells[r][c]) > len(letters) {
				km.cells[r][c] = 0
			}
		}
	}
}

// setFrame puts the U frame in or takes it out; the cells under it, when it
// goes in, turn back to the background.
func (km *kamon) setFrame(on bool) {
	km.frame = on
	for r := range rows {
		for c := range cols {
			if km.onFrame(r, c) {
				km.cells[r][c] = 0
			}
		}
	}
}

// checkLetters is the letters' own rules, as typed: A to Z in either case,
// at most six.
func checkLetters(v string) error {
	for _, ch := range v {
		if (ch < 'A' || ch > 'Z') && (ch < 'a' || ch > 'z') {
			return errors.New("Letters can only be A to Z")
		}
	}
	if len(v) > maxLetters {
		return fmt.Errorf("Letters can't be more than %d", maxLetters)
	}
	return nil
}

var errNoLetters = errors.New("Letters can't be empty")

// splitLetters is typed letters one by one, in capitals.
func splitLetters(v string) []string { return strings.Split(strings.ToUpper(v), "") }

// kamonFile is the yaml as it is written.
type kamonFile struct {
	Frame   bool     `yaml:"frame"`
	Letters []string `yaml:"letters"`
	Cells   string   `yaml:"cells"`
}

// linesError is yaml's unmarshal errors, a line each: it reads as the first,
// the line that says why, for the screen; the whole goes to the App Log
// (components, control characters; kamonu.read).
type linesError struct{ yaml *yaml.TypeError }

func (e *linesError) Error() string { return e.yaml.Errors[0] }
func (e *linesError) Unwrap() error { return e.yaml }

// parseKamon reads a kamon's yaml. Whatever is wrong is an error: a cell
// that is not 25 to a row, an unknown letter, a letter on the U frame.
func parseKamon(b []byte) (kamon, error) {
	var f kamonFile
	d := yaml.NewDecoder(bytes.NewReader(b))
	d.KnownFields(true)
	if err := d.Decode(&f); err != nil {
		if err == io.EOF {
			return kamon{}, errors.New("the file is empty")
		}
		var te *yaml.TypeError
		if errors.As(err, &te) && len(te.Errors) > 0 {
			return kamon{}, &linesError{te}
		}
		return kamon{}, err
	}
	km := kamon{frame: f.Frame}
	if len(f.Letters) == 0 {
		return kamon{}, errNoLetters
	}
	for _, l := range f.Letters {
		if utf8.RuneCountInString(l) != 1 {
			return kamon{}, fmt.Errorf("Letter %q is not one letter", l)
		}
	}
	if err := checkLetters(strings.Join(f.Letters, "")); err != nil {
		return kamon{}, err
	}
	km.letters = splitLetters(strings.Join(f.Letters, ""))
	lines := strings.Split(strings.TrimSuffix(f.Cells, "\n"), "\n")
	if f.Cells == "" {
		lines = nil
	}
	if len(lines) != rows {
		return kamon{}, fmt.Errorf("cells has %d rows, not %d", len(lines), rows)
	}
	for r, line := range lines {
		if n := utf8.RuneCountInString(line); n != cols {
			return kamon{}, fmt.Errorf("Row %d has %d cells, not %d", r+1, n, cols)
		}
		for c, ch := range []rune(line) {
			switch {
			case ch == '.':
				continue
			case ch >= '1' && ch <= rune('0'+len(km.letters)):
				km.cells[r][c] = uint8(ch - '0')
			default:
				return kamon{}, fmt.Errorf("Row %d, column %d is %q: a cell is . or 1 to %d", r+1, c+1, ch, len(km.letters))
			}
			if km.onFrame(r, c) {
				return kamon{}, fmt.Errorf("Row %d, column %d is on the U frame", r+1, c+1)
			}
		}
	}
	return km, nil
}

// yaml writes the kamon: the cells a picture, a row a line, . for none and
// the letter's place for the rest.
func (km *kamon) yaml() []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "frame: %t\n", km.frame)
	fmt.Fprintf(&b, "letters: [%s]\n", strings.Join(km.letters, ", "))
	b.WriteString("cells: |\n")
	for r := range rows {
		b.WriteString("  ")
		for c := range cols {
			if n := km.cells[r][c]; n > 0 {
				b.WriteByte('0' + n)
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteString("\n")
	}
	return b.Bytes()
}

// svg writes the kamon as tdp S5 does: the background, the U frame after
// its comment, then each letter after its own, in the letters' order, all
// gold. A letter with no cells is left out: the splash has nothing of it to
// show.
func (km *kamon) svg() []byte {
	var b bytes.Buffer
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 250 250">` + "\n")
	b.WriteString(`    <rect x="0" y="0" width="250" height="250" fill="#000000"/>` + "\n")
	part := func(comment, fill string, in func(r, c int) bool) {
		first := true
		for r := range rows {
			for c := range cols {
				if !in(r, c) {
					continue
				}
				if first {
					b.WriteString("\n    <!-- " + comment + " -->\n")
					first = false
				}
				fmt.Fprintf(&b, `    <rect x="%d" y="%d" width="10" height="10" fill="%s"/>`+"\n", c*cellSize, originY+r*cellSize, fill)
			}
		}
	}
	part("U", "#205090", km.onFrame)
	for i, l := range km.letters {
		part(l, "#f2b753", func(r, c int) bool { return int(km.cells[r][c]) == i+1 })
	}
	b.WriteString("</svg>\n")
	return b.Bytes()
}
