package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	terminu "github.com/vulcanshen/terminu-go"
)

// harness runs kamonu without a terminal: keys in, the frame out as text.
type harness struct {
	t    *testing.T
	m    model
	k    *kamonu
	quit bool // the program asked to quit
}

// newHarness is kamonu on a w × h screen over a data directory holding the
// given kamon (name → yaml).
func newHarness(t *testing.T, w, h int, kamon map[string][]byte) *harness {
	t.Helper()
	terminu.SetIconWidth(1, 1)
	st := store{dir: filepath.Join(t.TempDir(), "data")}
	if len(kamon) > 0 {
		if err := os.MkdirAll(st.conf(), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for name, b := range kamon {
		if err := os.WriteFile(st.path(name), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	k := newKamonu(st)
	hs := &harness{t: t, m: model{k}, k: k}
	hs.send(tea.WindowSizeMsg{Width: w, Height: h})
	return hs
}

// send gives msg to the program and runs what it sets off, animations
// included, until nothing is left to run.
func (hs *harness) send(msg tea.Msg) {
	_, cmd := hs.m.Update(msg)
	hs.run(cmd)
}

// longTimer is how long a command may take before the harness stops waiting
// for it: an animation frame is 32ms, a toast's time on the screen seconds.
// What is not waited for does not happen, so a toast stays to be read.
const longTimer = 200 * time.Millisecond

func (hs *harness) run(cmd tea.Cmd) {
	deadline := time.Now().Add(3 * time.Second)
	queue := []tea.Cmd{cmd}
	for len(queue) > 0 && time.Now().Before(deadline) {
		c := queue[0]
		queue = queue[1:]
		if c == nil {
			continue
		}
		done := make(chan tea.Msg, 1)
		go func() { done <- c() }()
		var msg tea.Msg
		select {
		case msg = <-done:
		case <-time.After(longTimer):
			continue
		}
		switch msg := msg.(type) {
		case nil:
		case tea.BatchMsg:
			queue = append(queue, msg...)
		case tea.QuitMsg:
			hs.quit = true
			return
		default:
			_, next := hs.m.Update(msg)
			queue = append(queue, next)
		}
	}
}

// keys presses each key: a name as terminu.KeyName writes it ("Enter",
// "Esc", "Space", "Tab", "←"), or text to type.
func (hs *harness) keys(keys ...string) {
	for _, k := range keys {
		hs.send(keyMsg(k))
	}
}

func keyMsg(k string) tea.KeyMsg {
	named := map[string]tea.KeyType{
		"Enter": tea.KeyEnter, "Esc": tea.KeyEsc, "Tab": tea.KeyTab, "Shift-Tab": tea.KeyShiftTab,
		"Space": tea.KeySpace, "Backspace": tea.KeyBackspace, "Delete": tea.KeyDelete,
		"←": tea.KeyLeft, "→": tea.KeyRight, "↑": tea.KeyUp, "↓": tea.KeyDown,
		"Home": tea.KeyHome, "End": tea.KeyEnd, "Ctrl-C": tea.KeyCtrlC,
	}
	if t, ok := named[k]; ok {
		if t == tea.KeySpace {
			return tea.KeyMsg{Type: t, Runes: []rune{' '}}
		}
		return tea.KeyMsg{Type: t}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

// paste pastes text, control characters and all.
func (hs *harness) paste(text string) {
	hs.send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text), Paste: true})
}

// frame is the screen as plain text.
func (hs *harness) frame() []string {
	var out []string
	for _, l := range hs.k.frame() {
		out = append(out, terminu.StripStyles(l))
	}
	return out
}

func (hs *harness) text() string { return strings.Join(hs.frame(), "\n") }

// contains fails unless the screen shows s.
func (hs *harness) contains(s string) {
	hs.t.Helper()
	if !strings.Contains(hs.text(), s) {
		hs.t.Fatalf("the screen does not show %q:\n%s", s, hs.text())
	}
}

func (hs *harness) lacks(s string) {
	hs.t.Helper()
	if strings.Contains(hs.text(), s) {
		hs.t.Fatalf("the screen shows %q:\n%s", s, hs.text())
	}
}

// fullWidth checks tdp L4: every line exactly as wide as the terminal.
func (hs *harness) fullWidth() {
	hs.t.Helper()
	w, h := hs.k.a.Size()
	lines := hs.k.frame()
	if len(lines) != h {
		hs.t.Fatalf("%d lines on a screen %d high", len(lines), h)
	}
	for i, l := range lines {
		if got := terminu.Width(l); got != w {
			hs.t.Fatalf("line %d is %d cells on a screen %d wide: %q", i, got, w, terminu.StripStyles(l))
		}
	}
}
