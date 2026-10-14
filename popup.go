package main

// kamonu's own popups: the grid editor and the popups opened over it (the
// discard confirm, SAVE AS, an error popup, a key reference, the quit
// confirm). The lib draws its popups in its own stack, under anything an app
// draws over its frame, so these are kamonu's, drawn after the lib's frame
// and looking as the lib's do (components/layout/popup). docs/dev-remarks.md,
// 偏離 tdp: F1, E8.

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	terminu "github.com/vulcanshen/terminu-go"
)

// popup is one kind of kamonu's popups: its size, look and keys.
type popup interface {
	// size is the popup's outer size on a sw × sh screen, fixed when it
	// opens (tdp F7).
	size(sw, sh int) (w, h int)
	// draw is h lines of w cells, the frame included, in the layer colour.
	draw(w, h int, layer terminu.Color) []string
	// key handles a key on the top popup and reports whether it did; the
	// core keys it leaves are the stack's (tdp K1).
	key(k *kamonu, key string, msg tea.KeyMsg) (bool, tea.Cmd)
	// name is the popup in its key reference's title.
	name() string
	// keyRef is what its key reference lists: only its own keys (tdp K6).
	keyRef() []refSection
	// typing reports the input state: ? and q are characters (tdp K8).
	typing() bool
}

// The open and close animation (tdp F2): a line from the middle (4 frames),
// a hold (1), then the popup opens up and down (4); 32ms a frame.
const (
	frameTime  = 32 * time.Millisecond
	openFrames = 9
)

type animMsg struct{}

// layer is one popup on kamonu's stack.
type layer struct {
	p       popup
	w, h    int
	x, y    int
	frame   int // 1–9 while opening, 9 when open, down to 0 while closing
	closing bool
	// below runs when the layer is gone.
	below func() tea.Cmd
}

func (k *kamonu) open(p popup) tea.Cmd { return k.openThen(p, nil) }

// openThen opens p on the stack; below runs once it has closed.
func (k *kamonu) openThen(p popup, below func() tea.Cmd) tea.Cmd {
	l := &layer{p: p, frame: 1, below: below}
	k.place(l)
	k.pops = append(k.pops, l)
	return k.animate()
}

func (k *kamonu) place(l *layer) {
	sw, sh := k.a.Size()
	l.w, l.h = l.p.size(sw, sh)
	l.x, l.y = (sw-l.w)/2, (sh-l.h)/2
}

// top is the top popup that is not closing.
func (k *kamonu) top() *layer {
	for i := len(k.pops) - 1; i >= 0; i-- {
		if !k.pops[i].closing {
			return k.pops[i]
		}
	}
	return nil
}

// closeTop starts closing the top popup (tdp F3, K4).
func (k *kamonu) closeTop() tea.Cmd {
	l := k.top()
	if l == nil {
		return nil
	}
	if c, ok := l.p.(*confirm); ok && c.quit {
		k.quitting = false
	}
	l.closing = true
	l.frame = min(l.frame, openFrames)
	return k.animate()
}

// closeAll closes the whole stack: the top one animates, the rest go at once
// (tdp T1, when what was done ends the stack).
func (k *kamonu) closeAll() tea.Cmd {
	t := k.top()
	if t == nil {
		return nil
	}
	kept := k.pops[:0]
	for _, l := range k.pops {
		if l == t || l.closing {
			kept = append(kept, l)
		}
	}
	k.pops = kept
	k.quitting = false
	return k.closeTop()
}

// animate keeps one tick going while a layer opens or closes.
func (k *kamonu) animate() tea.Cmd {
	if k.anim {
		return nil
	}
	k.anim = true
	return tea.Tick(frameTime, func(time.Time) tea.Msg { return animMsg{} })
}

// step is one frame of every layer that opens or closes.
func (k *kamonu) step() tea.Cmd {
	k.anim = false
	var cmds []tea.Cmd
	kept := k.pops[:0]
	busy := false
	for _, l := range k.pops {
		switch {
		case l.closing:
			l.frame--
			if l.frame <= 0 {
				if l.below != nil {
					cmds = append(cmds, l.below())
				}
				continue
			}
			busy = true
		case l.frame < openFrames:
			l.frame++
			busy = busy || l.frame < openFrames
		}
		kept = append(kept, l)
	}
	k.pops = kept
	if busy {
		cmds = append(cmds, k.animate())
	}
	return tea.Batch(cmds...)
}

// popKey is a key while one of kamonu's popups is on top: its own, then the
// core keys it leaves (tdp K1).
func (k *kamonu) popKey(l *layer, msg tea.KeyMsg) tea.Cmd {
	key := terminu.KeyName(msg)
	if l.frame < openFrames {
		return nil // still opening: no keys (tdp F2)
	}
	if key == "Ctrl-C" {
		if k.quitting {
			return tea.Quit // the second Ctrl-C of a quit flow (tdp K9)
		}
		return k.startQuit()
	}
	if !l.p.typing() {
		// g waits for the key after it; a lone g does nothing (tdp K12).
		pending := k.pending
		k.pending = ""
		switch {
		case key == "g" && pending == "":
			k.pending = "g"
			return nil
		case key == "g":
			key = "gg"
		}
	}
	if done, cmd := l.p.key(k, key, msg); done {
		return cmd
	}
	if l.p.typing() {
		return nil
	}
	switch key {
	case "Esc":
		return k.closeTop()
	case "?":
		return k.open(keyReference(l.p.name(), l.p.keyRef()))
	case "q":
		return k.startQuit()
	}
	return nil
}

// startQuit is the quit flow over kamonu's popups (tdp K9): the app's quit
// confirm, on top of them.
func (k *kamonu) startQuit() tea.Cmd {
	if k.quitting {
		return nil
	}
	k.quitting = true
	ask := *k.a.Quit
	for _, l := range k.pops {
		if e, ok := l.p.(*editor); ok && e.dirty {
			ask.Details, ask.Warn = []string{"Changes to " + e.kamon + " are not saved."}, []int{0}
		}
	}
	return k.open(&confirm{ask: ask, quit: true, accept: func() tea.Cmd { return tea.Quit }})
}

// compose draws kamonu's popups over the lib's frame: everything under the
// top one dimmed, each in its layer colour (tdp F8).
func (k *kamonu) compose(lines []string) []string {
	if len(k.pops) == 0 {
		return lines
	}
	w, _ := k.a.Size()
	top := k.top()
	if top != nil {
		lines = terminu.DimLines(lines)
	}
	for i, l := range k.pops {
		color := terminu.LayerColor(i + 1)
		full := l.p.draw(l.w, l.h, color)
		if l != top && !l.closing {
			full = terminu.DimLines(full)
		}
		rows, first, dx := l.drawFrame(full, color)
		terminu.Composite(lines, rows, l.x+dx, l.y+first, w)
	}
	return lines
}

// drawFrame is the layer as it looks in its animation frame; rows are
// relative to the layer's top (components/layout/popup, the animation).
func (l *layer) drawFrame(full []string, color terminu.Color) (rows []string, first, x int) {
	f := l.frame
	switch {
	case f <= 0:
		return nil, 0, 0
	case f <= 5:
		// The line, from the middle out: a quarter of the width a frame.
		n := min(f, 4)
		width := max((l.w*n+3)/4, 1)
		if width%2 != l.w%2 && width < l.w {
			width++
		}
		return []string{terminu.Fg(color).Render(terminu.Repeat("━", width))}, l.h / 2, (l.w - width) / 2
	case f < openFrames:
		// The rows around the middle: a quarter of the height a frame.
		n := max((l.h*(f-5)+3)/4, 1)
		start := (l.h - n) / 2
		return full[start : start+n], start, 0
	}
	return full, 0, 0
}

// Glyphs tdp names by code point (tdp E4).
const (
	glyphRoundLeft  = "\U0000E0B6" // nf-ple-left_half_circle_thick: a popup title's left end
	glyphRoundRight = "\U0000E0B4" // nf-ple-right_half_circle_thick: its right end
	glyphHexagon    = "\U000F02D8" // nf-md-hexagon: before a section title
	glyphTyping     = "\U0000F105" // nf-fa-angle_right: before a typing row
	glyphKeys       = "\U0000F059" // nf-fa-question_circle: a key reference
	glyphConfirm    = "\U000F0ECC" // nf-md-shield_alert: a confirm
	glyphError      = "\U000F0238" // nf-md-fire: an error popup
	glyphSquare     = "\U0000F0C8" // nf-fa-square: a cell of a mark (tdp S5)
)

// box draws a popup's frame around body (components/layout/popup): heavy
// lines in color, the title capsule one segment from the corner, the hint
// after one segment, the scroll position before the last; a blank row above
// and below body, whose rows come in as wide as the inside.
func box(color terminu.Color, glyph, title string, hint []terminu.KeyHint, scroll string, w int, body []string) []string {
	inner := w - 2
	line := terminu.Fg(color)
	out := make([]string, 0, len(body)+4)

	capsule := titleCapsule(glyph, title, color, inner-2)
	top := line.Render("━") + capsule
	out = append(out, line.Render("┏")+top+line.Render(terminu.Repeat("━", inner-terminu.Width(top))+"┓"))

	blank := terminu.Repeat(" ", inner)
	for _, r := range append(append([]string{blank}, body...), blank) {
		out = append(out, line.Render("┃")+terminu.Pad(r, inner)+line.Render("┃"))
	}

	if scroll != "" {
		scroll = " " + scroll + " "
	}
	tail := ""
	if scroll != "" {
		tail = "━"
	}
	hintText := ""
	if h := renderHints(hint, inner-4-terminu.Width(scroll)-terminu.Width(tail)); h != "" {
		hintText = " " + h + " "
	}
	fill := inner - 1 - terminu.Width(hintText) - terminu.Width(scroll) - terminu.Width(tail)
	out = append(out, line.Render("┗━")+hintText+line.Render(terminu.Repeat("━", fill)+scroll+tail+"┛"))
	return out
}

// titleCapsule is a popup's title: glyph and text in a capsule rounded at
// both ends, filled with color, the text in Base bold.
func titleCapsule(glyph, title string, color terminu.Color, maxW int) string {
	text := terminu.Truncate(glyph+" "+title, maxW-terminu.Width(glyphRoundLeft)-terminu.Width(glyphRoundRight))
	if text == "" {
		return ""
	}
	return terminu.Fg(color).Render(glyphRoundLeft) + terminu.On(color).Render(text) + terminu.Fg(color).Render(glyphRoundRight)
}

// renderHints writes pairs as "Key:desc Key:desc" (tdp M5), keys in Blue and
// the rest in Overlay0, dropping whole pairs from the end until they fit in w.
func renderHints(pairs []terminu.KeyHint, w int) string {
	for n := len(pairs); n > 0; n-- {
		out, width := "", n-1
		for i, p := range pairs[:n] {
			if i > 0 {
				out += " "
			}
			out += renderKey(p.Key, terminu.Fg(terminu.Blue)) + terminu.Fg(terminu.Overlay0).Render(":"+p.Desc)
			width += terminu.Width(p.Key) + 1 + terminu.Width(p.Desc)
		}
		if width <= w {
			return out
		}
	}
	return ""
}

// renderKey draws a key; keys doing one thing are joined with "|" in the
// words' colour (tdp M5: "j|k").
func renderKey(key string, style terminu.Style) string {
	out := ""
	for i, part := range splitBar(key) {
		if i > 0 {
			out += terminu.Fg(terminu.Overlay0).Render("|")
		}
		out += style.Render(part)
	}
	return out
}

func splitBar(key string) []string {
	if key == "|" {
		return []string{key}
	}
	var parts []string
	start := 0
	for i := 0; i < len(key); i++ {
		if key[i] == '|' {
			parts = append(parts, key[start:i])
			start = i + 1
		}
	}
	return append(parts, key[start:])
}

// hintWidth is how wide pairs are written in full.
func hintWidth(pairs []terminu.KeyHint) int {
	w := max(len(pairs)-1, 0)
	for _, p := range pairs {
		w += terminu.Width(p.Key) + 1 + terminu.Width(p.Desc)
	}
	return w
}

// padRow is a row of a popup's body, one cell from the frame on either side.
func padRow(s string, inner int) string { return " " + terminu.Fit(s, inner-2) + " " }

// popupWidth is tdp F7's width: content + 4 when it is known when the popup
// opens, at most min(screen − 2, 120).
func popupWidth(sw, content int) int { return min(content+4, max(min(sw-2, 120), 1)) }

// scrollText is "N-M of T": the lines in sight of a popup without a cursor.
func scrollText(first, last, total int) string {
	return strconv.Itoa(first) + "-" + strconv.Itoa(last) + " of " + strconv.Itoa(total)
}
