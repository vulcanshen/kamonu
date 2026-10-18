package app

// The screen: [1] KAMON, the kamon in the data directory, and [2] PREVIEW,
// the one under the cursor (docs/dev-remarks.md, the screen).

import (
	"errors"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	terminu "github.com/vulcanshen/terminu-go"
)

// Title glyphs of kamonu's inputs and forms (the app's choice:
// components/layout/popup).
const (
	glyphAdd    = "\U000F0415" // nf-md-plus_box: NEW KAMON
	glyphCog    = "\U000F0493" // nf-md-cog: EDIT
	glyphExport = "\U000F0207" // nf-md-export: EXPORT
	glyphPencil = "\U000F03EB" // nf-md-pencil: a name, the letters, a file
	glyphSave   = "\U000F0193" // nf-md-content_save: SAVE AS
)

// previewWidth is [2]'s width: the mark's 50 columns, a cell of room and the
// frame on either side.
const previewWidth = 2*cols + 4

type kamonu struct {
	a       *terminu.App
	st      store
	list    *terminu.List
	panel   *terminu.Panel // [1] KAMON
	preview *preview
}

func newKamonu(st store) *kamonu {
	k := &kamonu{st: st}
	a := terminu.New("kamonu")
	a.Quit = &terminu.Ask{Title: "Quit", Question: "Quit kamonu?", Verb: "quit"}
	a.QuitDesc = "leave kamonu"
	a.Splash = splash()
	k.a = a

	k.list = &terminu.List{
		Columns:   []terminu.Column{{Title: "Name", Flex: true}},
		EmptyFact: "No kamon yet",
		// Short enough for [1] at 80 columns, 24 cells inside: a longer hint
		// is cut to the frame (layout/screen, empty states).
		EmptyHint: "Press [A] to add one",
	}
	k.panel = &terminu.Panel{Label: "Kamon", Body: k.list, Actions: []terminu.Action{
		{Key: "Enter", Label: "Draw", Desc: "its cells", Scope: terminu.ItemScope, Hint: "draw", Do: k.draw},
		{Key: "E", Label: "Edit", Desc: "name, letters, U frame", Scope: terminu.ItemScope, Do: k.edit},
		// "ExPort": the key is a capital P, written in its place (tdp M5).
		{Key: "P", Label: "ExPort", Desc: "write its svg", Scope: terminu.ItemScope, Do: k.export},
		{Key: "X", Label: "Delete", Desc: "this kamon's file", Scope: terminu.ItemScope, Do: k.delete},
		{Key: "A", Label: "Add", Desc: "a new kamon", Scope: terminu.PanelScope, Hint: "add", Do: k.add},
	}}
	k.panel.Actions = append(k.panel.Actions, k.list.Actions()...)
	k.preview = &preview{k: k}
	// Enter on the preview draws the kamon it shows: the most direct thing to
	// do with a picture of a mark (tdp K3; provisional, docs/dev-remarks.md).
	preview := &terminu.Panel{Label: "Preview", Body: k.preview, Actions: []terminu.Action{
		{Key: "Enter", Label: "Draw", Desc: "the kamon shown here", Scope: terminu.PanelScope, Hint: "draw",
			Hidden: func() bool { return k.current() == "" }, Do: k.draw},
	}}
	k.preview.panel = preview
	a.AddScreen(&terminu.Screen{
		Name:   "Kamon",
		Panels: []*terminu.Panel{k.panel, preview},
		Layout: func(w, h int) []terminu.Rect {
			left := max(w-previewWidth, 0)
			return []terminu.Rect{{X: 0, Y: 0, W: left, H: h}, {X: left, Y: 0, W: w - left, H: h}}
		},
		NarrowBelow: 80,
	})
	k.reload("")
	return k
}

// current is the kamon under the cursor; "" when there is none.
func (k *kamonu) current() string {
	r, ok := k.list.Current()
	if !ok {
		return ""
	}
	return r.ID
}

// reload reads the data directory again, the cursor on sel when it is given.
func (k *kamonu) reload(sel string) {
	names, err := k.st.names()
	if err != nil {
		k.panel.Fail = &terminu.Failure{Fact: "Can't read " + shown(k.st.conf()), Hint: why(err)}
		k.list.SetRows(nil)
		return
	}
	k.panel.Fail = nil
	rows := make([]terminu.Row, len(names))
	for i, n := range names {
		rows[i] = terminu.Row{ID: n, Cells: []string{n}}
	}
	k.list.SetRows(rows)
	if sel != "" {
		k.list.Select(sel)
	}
	k.preview.forget()
}

// draw opens the editor on the kamon under the cursor.
func (k *kamonu) draw() tea.Cmd {
	name := k.current()
	km, err := k.st.load(name)
	if err != nil {
		return k.a.Error("Draw failed", "Can't read "+name+": "+why(err))
	}
	return k.openEditor(name, km)
}

// The fields NEW KAMON and EDIT share: the name, the letters, and
// whether the U frame goes in (Base).

func nameField(name *string, prompt string, check func() string) *terminu.FormField {
	return &terminu.FormField{Label: "Name", Value: func() string { return *name }, Empty: "not set",
		Edit: func() terminu.Input {
			return terminu.TextInput{Title: "Name", Glyph: glyphPencil, Prompt: prompt, Verb: "set", Old: *name,
				Check: checkFormat, Done: func(v string) tea.Cmd { *name = v; return nil }}
		},
		Check: check}
}

func lettersField(letters *string, prompt string) *terminu.FormField {
	return &terminu.FormField{Label: "Letters", Value: func() string { return *letters }, Empty: "not set",
		Edit: func() terminu.Input {
			return terminu.TextInput{Title: "Letters", Glyph: glyphPencil, Prompt: prompt,
				Verb: "set", Old: *letters,
				Check: checkLetters, Done: func(v string) tea.Cmd { *letters = strings.ToUpper(v); return nil }}
		},
		Check: func() string {
			if *letters == "" {
				return errNoLetters.Error()
			}
			return ""
		}}
}

func baseField(on bool) *terminu.FormField {
	return &terminu.FormField{Label: "Base", Options: []string{"U frame"}, Multi: true, Checked: []bool{on}}
}

// add is NEW KAMON: a name, the letters and whether the U frame goes in.
// Create makes the file at once, then opens the editor on it.
func (k *kamonu) add() tea.Cmd {
	name, letters := "", ""
	base := baseField(true)
	return k.a.Form(&terminu.Form{
		Title: "New kamon", Glyph: glyphAdd, Name: "New kamon", Button: "Create",
		Fields: []*terminu.FormField{
			nameField(&name, "Name of the new kamon", func() string {
				if err := k.st.checkName(name); err != nil {
					return err.Error()
				}
				return ""
			}),
			lettersField(&letters, "The letters in the U frame of the new kamon, in the order of the app's name"),
			base,
		},
		Submit: func() (tea.Cmd, error) {
			km := kamon{frame: base.Checked[0], letters: splitLetters(letters)}
			if err := k.st.create(name, km); err != nil {
				return nil, errors.New("Can't create " + name + ": " + why(err))
			}
			k.reload(name)
			return k.openEditor(name, km), nil
		},
	})
}

// edit is EDIT: <name>, the name, the letters and the U frame of the kamon
// under the cursor. The cells stay where they are (kamon.setLetters,
// kamon.setFrame): what that makes of them is the user's to draw again.
func (k *kamonu) edit() tea.Cmd {
	old := k.current()
	km, err := k.st.load(old)
	if err != nil {
		return k.a.Error("Edit failed", "Can't read "+old+": "+why(err))
	}
	name, letters := old, strings.Join(km.letters, "")
	base := baseField(km.frame)
	return k.a.Form(&terminu.Form{
		// The name after ": " in its own case (tdp v0.6.1, layout/popup).
		Title: "Edit", Subject: old, Glyph: glyphCog, Name: old, Button: "Save",
		Fields: []*terminu.FormField{
			nameField(&name, "New name for "+old, func() string {
				if name == old {
					return ""
				}
				if err := k.st.checkName(name); err != nil {
					return err.Error()
				}
				return ""
			}),
			lettersField(&letters, "The letters in the U frame of "+old+", in the order of the app's name"),
			base,
		},
		Submit: func() (tea.Cmd, error) {
			km.setLetters(splitLetters(letters))
			km.setFrame(base.Checked[0])
			if err := k.st.save(old, km); err != nil {
				return nil, errors.New("Can't save " + old + ": " + why(err))
			}
			if name != old {
				if err := k.st.rename(old, name); err != nil {
					k.reload(old)
					return nil, errors.New("Can't rename " + old + ": " + why(err))
				}
			}
			k.reload(name)
			return nil, nil
		},
	})
}

// export writes the svg of the kamon under the cursor to svg/<name>.svg and
// says where in a toast. When <name>.svg is there, EXPORT: <name> asks
// first: Overwrite writes over it; Export, the main submit, writes File, a
// new one, at first the next <name>_<n>.svg.
func (k *kamonu) export() tea.Cmd {
	name := k.current()
	km, err := k.st.load(name)
	if err == nil && !k.st.svgExists(name) {
		err = k.st.export(name, km)
		if err == nil {
			return k.exported(name)
		}
	}
	if err != nil {
		return k.a.Error("Export failed", "Can't export "+name+": "+why(err))
	}
	// <name>.svg is there.
	file := k.st.nextSVG(name)
	write := func(f string) (tea.Cmd, error) {
		if err := k.st.export(f, km); err != nil {
			return nil, errors.New("Can't export " + name + ": " + why(err))
		}
		return k.exported(f), nil
	}
	return k.a.Form(&terminu.Form{
		Title: "Export", Subject: name, Glyph: glyphExport, Name: "Export", Button: "Export",
		Fields: []*terminu.FormField{{Label: "File", Value: func() string { return file + ".svg" },
			Edit: func() terminu.Input {
				return terminu.TextInput{Title: "File", Glyph: glyphPencil, Prompt: "Name of the new svg; " + name + ".svg stays as it is",
					Verb: "set", Old: file + ".svg",
					Check: func(v string) error { return k.st.checkSVG(strings.TrimSuffix(v, ".svg")) },
					Done:  func(v string) tea.Cmd { file = strings.TrimSuffix(v, ".svg"); return nil }}
			}}},
		// The file it writes over is on the button (the user's ask), as it
		// is in the hint too: "Enter:overwrite KBU.svg" (tdp v0.6.9).
		Buttons: [][]*terminu.FormButton{{{Label: "Overwrite", Subject: name + ".svg", Submit: func() (tea.Cmd, error) { return write(name) }}}},
		Submit: func() (tea.Cmd, error) {
			if err := k.st.checkSVG(file); err != nil {
				return nil, errors.New("Can't export " + name + ": " + err.Error())
			}
			return write(file)
		},
	})
}

// exported says in a toast where the svg went.
func (k *kamonu) exported(file string) tea.Cmd {
	msg := "Exported to " + shown(k.st.svgPath(file))
	// A toast shows three rows at most and cuts the rest (components/dialog/
	// toast); a path too long for that is said from the data directory,
	// where README says svg/ is.
	if w, _ := k.a.Size(); terminu.Width(msg) > 3*(min(w-2, 120)-4) {
		msg = "Exported to svg/" + file + ".svg"
	}
	return k.a.Toast(terminu.Info, msg)
}

func (k *kamonu) delete() tea.Cmd {
	name := k.current()
	return k.a.Confirm(terminu.Ask{
		Title: "Delete", Question: "Delete " + name + "?", Verb: "delete",
		Details: []string{name + ".yaml is removed from " + shown(k.st.conf()) + "."}, Warn: []int{0},
	}, func() tea.Cmd {
		if err := k.st.remove(name); err != nil {
			return k.a.Error("Delete failed", "Can't delete "+name+": "+why(err))
		}
		k.reload("")
		return nil
	})
}

// preview is [2]: the kamon under the cursor as the splash draws its mark,
// every letter gold, with a row of background above and below (tdp S5),
// from the top left of the panel, a cell from the frame
// (components/layout/panel). The lib's Pager shows it: when the panel is
// shorter than the mark, it scrolls and says which rows it shows. A kamon
// that cannot be read is the panel's Failure, its fact in Red.
type preview struct {
	k     *kamonu
	panel *terminu.Panel // [2], whose Fail says what could not be read
	name  string         // what m and err are of
	m     mark
	err   error
	read  bool
	pager terminu.Pager
}

// forget makes the next draw read the file again.
func (p *preview) forget() { p.read = false }

// Empty is a preview with nothing to show: no kamon, or one it cannot read.
func (p *preview) Empty() bool { return p.k.current() == "" || p.err != nil }

func (p *preview) Draw(w, h int, focused bool) []string {
	p.panel.Fail = nil
	name := p.k.current()
	if name == "" {
		p.pager.SetLines(nil)
		return terminu.EmptyState(w, h, "Nothing to preview", "Press [1], then [A] to add one")
	}
	if !p.read || name != p.name {
		p.name, p.read = name, true
		var km kamon
		km, p.err = p.k.st.load(name)
		p.m = km.picture()
	}
	if p.err != nil {
		p.pager.SetLines(nil)
		// Enter draws it, and drawing it opens an error popup that says why.
		p.panel.Fail = &terminu.Failure{Fact: "Can't draw " + name, Hint: "Press [Enter] to see why"}
		return terminu.EmptyState(w, h, "", "")
	}
	square := glyphSquare
	if terminu.IconWidth() == 1 {
		square += " " // a square takes two columns (tdp S5, E5)
	}
	backRow := terminu.Fg(backColor).Render(strings.Repeat(square, cols))
	mark := []string{backRow}
	for r := range rows {
		var b strings.Builder
		for c := range cols {
			b.WriteString(terminu.Fg(p.m[r][c].color()).Render(square))
		}
		mark = append(mark, b.String())
	}
	mark = append(mark, backRow)
	p.pager.SetLines(mark)
	return p.pager.Draw(w, h, focused)
}

func (p *preview) Key(key string, msg tea.KeyMsg) (bool, tea.Cmd) { return p.pager.Key(key, msg) }
func (p *preview) HasItem() bool                                  { return false }
func (p *preview) Scroll() string                                 { return p.pager.Scroll() }

// MoveKeys are the pager's, only while the mark does not fit (tdp M6: a key
// with nothing to act on is not listed).
func (p *preview) MoveKeys() []terminu.KeyRow {
	if p.pager.Scroll() == "" {
		return nil
	}
	return p.pager.MoveKeys()
}
