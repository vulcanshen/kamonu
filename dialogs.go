package main

// The popups kamonu opens over its editor: a confirm, a note (a key
// reference, an error popup) and a one-line input. They follow
// components/dialog/confirm, components/dialog/note and components/input,
// and look as the lib's do (see popup.go for why kamonu draws them).

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rivo/uniseg"
	terminu "github.com/vulcanshen/terminu-go"
)

// refSection is one section of a key reference: a title (none for the
// first) and its rows.
type refSection struct {
	title string
	rows  []terminu.KeyRow
}

// --- confirm (components/dialog/confirm) ---------------------------------

// confirm asks one question: what to read in Overlay0 (a warning in Peach),
// a blank line, the question in Text last. Enter accepts, Esc cancels (tdp
// F6).
type confirm struct {
	ask     terminu.Ask
	accept  func() tea.Cmd
	quit    bool // the quit flow's confirm (tdp K9)
	content []string
	quest   []string
	top     int
	visible int
}

func (c *confirm) hint() []terminu.KeyHint {
	return []terminu.KeyHint{{Key: "Enter", Desc: c.ask.Verb}, {Key: "Esc", Desc: "cancel"}}
}

func (c *confirm) wrapTo(w int) {
	c.content, c.quest = nil, nil
	for i, d := range c.ask.Details {
		style := terminu.Fg(terminu.Overlay0)
		if slices.Contains(c.ask.Warn, i) {
			style = terminu.Fg(terminu.Peach)
		}
		if d == "" {
			c.content = append(c.content, "")
			continue
		}
		for _, l := range terminu.Wrap(d, w) {
			c.content = append(c.content, style.Render(l))
		}
	}
	for _, l := range terminu.Wrap(c.ask.Question, w) {
		c.quest = append(c.quest, terminu.Fg(terminu.Text).Render(l))
	}
}

func (c *confirm) bodyRows() int {
	n := len(c.quest)
	if len(c.content) > 0 {
		n += len(c.content) + 1
	}
	return n
}

func (c *confirm) size(sw, sh int) (int, int) {
	content := terminu.Width(c.ask.Question)
	for _, d := range c.ask.Details {
		content = max(content, terminu.Width(d))
	}
	content = max(content, terminu.Width(titleCapsule(glyphConfirm, terminu.Upper(c.ask.Title), terminu.Base, 1<<20)), hintWidth(c.hint())+2)
	w := popupWidth(sw, content)
	c.wrapTo(w - 4)
	h := min(c.bodyRows()+4, max(sh-2, 3))
	// Only the details scroll: room for their position at the widest.
	if t := len(c.content); t > 0 && c.bodyRows()+4 > h {
		w = popupWidth(sw, max(content, hintWidth(c.hint())+terminu.Width(scrollText(t, t, t))+5))
		c.wrapTo(w - 4)
		h = min(c.bodyRows()+4, max(sh-2, 3))
	}
	return w, h
}

func (c *confirm) draw(w, h int, layer terminu.Color) []string {
	inner := w - 2
	c.wrapTo(w - 4)
	n := h - 4
	var body []string
	if len(c.content) > 0 {
		// Only the content scrolls; the question stays last.
		c.visible = max(n-len(c.quest)-1, 0)
		c.top = min(max(c.top, 0), max(len(c.content)-c.visible, 0))
		for i := c.top; i < c.top+c.visible; i++ {
			if i < len(c.content) {
				body = append(body, padRow(c.content[i], inner))
			} else {
				body = append(body, "")
			}
		}
		body = append(body, "")
	}
	for _, q := range c.quest {
		if len(body) < n {
			body = append(body, padRow(q, inner))
		}
	}
	for len(body) < n {
		body = append(body, "")
	}
	scroll := ""
	if len(c.content) > c.visible && c.visible > 0 {
		scroll = scrollText(c.top+1, min(c.top+c.visible, len(c.content)), len(c.content))
	}
	return box(layer, glyphConfirm, terminu.Upper(c.ask.Title), c.hint(), scroll, w, body)
}

func (c *confirm) key(k *kamonu, key string, _ tea.KeyMsg) (bool, tea.Cmd) {
	switch key {
	case "Enter":
		return true, tea.Batch(k.closeTop(), c.accept())
	case "Space":
		return true, nil // tdp K5
	case "j", "↓":
		c.top++
	case "k", "↑":
		c.top--
	default:
		return false, nil
	}
	return true, nil
}

func (c *confirm) name() string { return c.ask.Title }

func (c *confirm) keyRef() []refSection {
	secs := []refSection{{rows: []terminu.KeyRow{{Key: "Enter", Desc: c.ask.Verb}, {Key: "Esc", Desc: "cancel"}}}}
	if len(c.ask.Details) > 0 {
		secs = append(secs, refSection{title: "move", rows: []terminu.KeyRow{{Key: "j|k", Desc: "scroll down, up"}}})
	}
	return secs
}

func (c *confirm) typing() bool { return false }

// --- note (components/dialog/note) ---------------------------------------

// note is read-only and scrolls: a key reference, an error popup.
type note struct {
	glyph, title string
	lines        []string // drawn, already coloured
	red          bool     // an error popup: a Red frame
	enterCloses  bool     // an error popup closes with Enter too
	helpCloses   bool     // a key reference closes with ? too
	hint         []terminu.KeyHint
	surface      string
	top, visible int
}

func (n *note) size(sw, sh int) (int, int) {
	content := 0
	for _, l := range n.lines {
		content = max(content, terminu.Width(l))
	}
	content = max(content, terminu.Width(titleCapsule(n.glyph, n.title, terminu.Base, 1<<20)), hintWidth(n.hint)+2)
	if t := len(n.lines); t+4 > max(sh-2, 3) {
		content = max(content, hintWidth(n.hint)+terminu.Width(scrollText(t, t, t))+5)
	}
	return popupWidth(sw, content), min(len(n.lines)+4, max(sh-2, 3))
}

func (n *note) draw(w, h int, layer terminu.Color) []string {
	if n.red {
		layer = terminu.Red
	}
	inner := w - 2
	n.visible = h - 4
	n.top = min(max(n.top, 0), max(len(n.lines)-n.visible, 0))
	var body []string
	for i := n.top; i < n.top+n.visible; i++ {
		if i < len(n.lines) {
			body = append(body, padRow(n.lines[i], inner))
		} else {
			body = append(body, "")
		}
	}
	scroll := ""
	if len(n.lines) > n.visible {
		scroll = scrollText(n.top+1, min(n.top+n.visible, len(n.lines)), len(n.lines))
	}
	return box(layer, n.glyph, n.title, n.hint, scroll, w, body)
}

func (n *note) key(k *kamonu, key string, _ tea.KeyMsg) (bool, tea.Cmd) {
	half := max(n.visible/2, 1)
	switch key {
	case "j", "↓":
		n.top++
	case "k", "↑":
		n.top--
	case "d", "PgDn":
		n.top += half
	case "u", "PgUp":
		n.top -= half
	case "G":
		n.top = len(n.lines)
	case "gg":
		n.top = 0
	case "Enter":
		if n.enterCloses {
			return true, k.closeTop()
		}
		return true, nil
	case "?":
		if n.helpCloses {
			return true, k.closeTop()
		}
		return false, nil
	case "Space":
		return true, nil
	default:
		return false, nil
	}
	n.top = min(max(n.top, 0), max(len(n.lines)-n.visible, 0))
	return true, nil
}

func (n *note) name() string { return n.surface }

func (n *note) keyRef() []refSection {
	close := terminu.KeyRow{Key: "Esc", Desc: "close"}
	if n.enterCloses {
		close.Key = "Enter|Esc"
	}
	return []refSection{{rows: []terminu.KeyRow{close}}, {title: "move", rows: []terminu.KeyRow{
		{Key: "j|k", Desc: "scroll down, up"},
		{Key: "u|d", Desc: "half a page up, down"},
		{Key: "gg|G", Desc: "to the top, the bottom"},
	}}}
}

func (n *note) typing() bool { return false }

// keyReference is the key reference of a surface (components/dialog/note): a
// section title is a hexagon and upper case, one cell from the frame; every
// row two cells in, its key padded to the widest key, two cells, then what
// it does.
func keyReference(surface string, sections []refSection) *note {
	keyW := 0
	for _, s := range sections {
		for _, r := range s.rows {
			keyW = max(keyW, terminu.Width(r.Key))
		}
	}
	var lines []string
	for _, s := range sections {
		if len(s.rows) == 0 {
			continue
		}
		if s.title != "" {
			lines = append(lines, terminu.Fg(terminu.Overlay0).Render(glyphHexagon+" "+terminu.Upper(s.title)))
		}
		for _, r := range s.rows {
			lines = append(lines, " "+terminu.Pad(renderKey(r.Key, terminu.Fg(terminu.Blue)), keyW)+"  "+terminu.Fg(terminu.Text).Render(r.Desc))
		}
	}
	return &note{
		glyph: glyphKeys, title: terminu.Upper(surface + " keys"), lines: lines, helpCloses: true,
		hint: []terminu.KeyHint{{Key: "?|Esc", Desc: "close"}}, surface: surface + " keys",
	}
}

// errorPopup is an action failing (components/dialog/note, error popup): a
// Red frame and text, the fire glyph, Enter or Esc to close.
func errorPopup(title, message string, sw int) *note {
	w := min(max(terminu.Width(message), 20), max(min(sw-2, 120)-4, 1))
	var lines []string
	for _, l := range terminu.Wrap(message, w) {
		lines = append(lines, terminu.Fg(terminu.Red).Render(l))
	}
	return &note{
		glyph: glyphError, title: terminu.Upper(title), lines: lines, red: true, enterCloses: true,
		hint: []terminu.KeyHint{{Key: "Enter|Esc", Desc: "close"}}, surface: title,
	}
}

// --- one-line input (components/input) -----------------------------------

// lineInput is a one-line text input popup with no old value: a prompt, the
// typing row, an error row.
type lineInput struct {
	title, glyph string
	prompt, verb string
	ed           lineEditor
	err          string
	check        func(string) error
	done         func(string) tea.Cmd
}

func (li *lineInput) hint() []terminu.KeyHint {
	return []terminu.KeyHint{{Key: "Enter", Desc: li.verb}, {Key: "Esc", Desc: "cancel"}}
}

// size: a free value's width is not known when it opens (tdp F7); the
// prompt, a blank, the typing row, a blank, the error row.
func (li *lineInput) size(sw, sh int) (int, int) {
	return max(min(sw-2, 120), 1), min(9, max(sh-2, 3))
}

func (li *lineInput) draw(w, h int, layer terminu.Color) []string {
	inner := w - 2
	cw := inner - 2
	angle := terminu.Fg(layer).Render(glyphTyping) + " "
	body := []string{
		padRow(terminu.Fg(terminu.Overlay0).Render(li.prompt), inner),
		"",
		" " + angle + li.ed.view(cw-terminu.Width(angle)) + " ",
		"",
		padRow(terminu.Fg(terminu.Red).Render(terminu.Truncate(li.err, cw)), inner),
	}
	return box(layer, li.glyph, terminu.Upper(li.title), li.hint(), "", w, body[:min(len(body), max(h-4, 0))])
}

func (li *lineInput) key(k *kamonu, key string, msg tea.KeyMsg) (bool, tea.Cmd) {
	switch key {
	case "Enter":
		return true, li.confirm()
	case "Esc":
		return true, k.closeTop()
	}
	li.ed.key(key, msg)
	return true, nil // the typing row takes every key (tdp K8)
}

// confirm is Enter (tdp K3): a wrong value stays with its error; control
// characters are checked before the spaces around the value go
// (components/input/README).
func (li *lineInput) confirm() tea.Cmd {
	if li.ed.hasControl() {
		li.err = "Name can't have control characters"
		return nil
	}
	v := strings.TrimFunc(li.ed.value(), unicode.IsSpace)
	if err := li.check(v); err != nil {
		li.err = err.Error()
		return nil
	}
	li.err = ""
	return li.done(v)
}

func (li *lineInput) name() string         { return li.title }
func (li *lineInput) keyRef() []refSection { return nil }
func (li *lineInput) typing() bool         { return true }

// lineEditor is a one-line value and its cursor, a character at a time: a
// character is a grapheme cluster, a pasted control character one on its
// own (components/input/README).
type lineEditor struct {
	g    []string
	pos  int
	left int // the first character in sight
}

func isControl(r rune) bool { return r < 0x20 || r == 0x7f || (r >= 0x80 && r < 0xa0) }

// insert puts s at the cursor, keeping every control character.
func (e *lineEditor) insert(s string) {
	var add []string
	for s != "" {
		if r := []rune(s)[0]; isControl(r) {
			add = append(add, string(r))
			s = s[len(string(r)):]
			continue
		}
		i := strings.IndexFunc(s, isControl)
		if i < 0 {
			i = len(s)
		}
		gr := uniseg.NewGraphemes(s[:i])
		for gr.Next() {
			add = append(add, gr.Str())
		}
		s = s[i:]
	}
	e.g = slices.Insert(e.g, e.pos, add...)
	e.pos += len(add)
}

// key is an editing key (components/input/README): characters and Space,
// ←|→, Home|End, Backspace|Delete. Other Ctrl- and Alt- keys type nothing.
func (e *lineEditor) key(key string, msg tea.KeyMsg) {
	switch key {
	case "←":
		e.pos = max(e.pos-1, 0)
	case "→":
		e.pos = min(e.pos+1, len(e.g))
	case "Home":
		e.pos = 0
	case "End":
		e.pos = len(e.g)
	case "Backspace":
		if e.pos > 0 {
			e.g = slices.Delete(e.g, e.pos-1, e.pos)
			e.pos--
		}
	case "Delete":
		if e.pos < len(e.g) {
			e.g = slices.Delete(e.g, e.pos, e.pos+1)
		}
	case "Space":
		e.insert(" ")
	default:
		if msg.Paste || (msg.Type == tea.KeyRunes && !msg.Alt) {
			e.insert(string(msg.Runes))
		}
	}
}

func (e *lineEditor) value() string { return strings.Join(e.g, "") }

func (e *lineEditor) hasControl() bool {
	return slices.ContainsFunc(e.g, func(g string) bool { return isControl([]rune(g)[0]) })
}

// ctrlName is how a control character is drawn: \r, \n and \t by name, the
// rest as \xNN.
func ctrlName(g string) string {
	switch r := []rune(g)[0]; r {
	case '\r':
		return `\r`
	case '\n':
		return `\n`
	case '\t':
		return `\t`
	default:
		return fmt.Sprintf(`\x%02x`, r)
	}
}

// view is the value in w cells: Lavender, a pasted control character by
// name in Red, the cursor a cell of Lavender (Red on a control character);
// when it does not fit, the part around the cursor with … where it is cut.
func (e *lineEditor) view(w int) string {
	n := len(e.g)
	items := n
	if e.pos == n {
		items++ // the cursor after the last character
	}
	text := func(i int) string {
		if i == n {
			return " "
		}
		if isControl([]rune(e.g[i])[0]) {
			return ctrlName(e.g[i])
		}
		return e.g[i]
	}
	// right is where the characters from left stop fitting, one cell kept for
	// each … that is needed.
	right := func(left int) int {
		room := w
		if left > 0 {
			room--
		}
		used, r := 0, left
		for r < items && used+terminu.Width(text(r)) <= room {
			used += terminu.Width(text(r))
			r++
		}
		for r < items && r > left && used+1 > room {
			r--
			used -= terminu.Width(text(r))
		}
		return r
	}
	e.left = min(e.left, e.pos)
	for e.left < e.pos && right(e.left) <= e.pos {
		e.left++
	}
	end := right(e.left)
	lavender := terminu.Fg(terminu.Lavender)
	var b strings.Builder
	if e.left > 0 {
		b.WriteString(lavender.Render("…"))
	}
	for i := e.left; i < end; i++ {
		t := text(i)
		ctrl := i < n && isControl([]rune(e.g[i])[0])
		switch {
		case i == e.pos && ctrl:
			b.WriteString(terminu.Style{Fg: terminu.Base, Bg: terminu.Red}.Render(t))
		case i == e.pos:
			b.WriteString(terminu.Style{Fg: terminu.Base, Bg: terminu.Lavender}.Render(t))
		case ctrl:
			b.WriteString(terminu.Fg(terminu.Red).Render(t))
		default:
			b.WriteString(lavender.Render(t))
		}
	}
	if end < items {
		b.WriteString(lavender.Render("…"))
	}
	return terminu.Pad(b.String(), w)
}
