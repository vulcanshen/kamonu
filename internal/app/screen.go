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

// minWidth × minHeight is kamonu's least size (tdp L1): the drawing's 50
// columns with a cell of room and the frame on either side, and a column
// beside the popup on either side (tdp F7); and rows enough for six of its
// rows over the letters, the separator and the buttons, the popup two rows
// short of the screen. Smaller, the lib shows only the name and the size
// it needs.
const minWidth, minHeight = 2*cols + 6, 16

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
	a.MinWidth, a.MinHeight = minWidth, minHeight
	k.a = a

	k.list = &terminu.List{
		// A long name is cut with … where [1] is narrow (layout/panel).
		Columns:   []terminu.Column{{Title: "Name", Shrink: true}},
		EmptyFact: "No kamon yet",
		EmptyHint: "Press [A] to add one",
	}
	k.panel = &terminu.Panel{Label: "Kamon", Body: k.list, Actions: []terminu.Action{
		{Key: "Enter", Label: "Draw", Desc: "its cells", Scope: terminu.ItemScope, Hint: "draw", Do: k.draw},
		{Key: "e", Label: "Edit", Desc: "name, letters, U frame", Scope: terminu.ItemScope, Do: k.edit},
		// The key is a capital P: the lib marks it where it falls in the
		// label, Ex[P]ort, once it matches letters in either case (tdp M5).
		{Key: "P", Label: "Export", Desc: "write its svg", Scope: terminu.ItemScope, Do: k.export},
		// Asked when it runs, about the kamon under the cursor; from the Space
		// menu Esc on the confirm goes back to it, and a kamon deleted clears
		// it (tdp T1).
		{Key: "X", Label: "Delete", Desc: "this kamon's file", Scope: terminu.ItemScope, Clear: true,
			AskFor: func() *terminu.Ask { ask := k.deleteAsk(k.current()); return &ask }, Do: k.delete},
		{Key: "A", Label: "Add", Desc: "a new kamon", Scope: terminu.PanelScope, Hint: "add", Do: k.add},
	}}
	k.panel.Actions = append(k.panel.Actions, k.list.Actions()...)
	k.preview = &preview{k: k}
	// [2] has one item, the kamon it shows (KM8, KM17): Enter shows the mark
	// as its app shows it, and [e]dit opens its cells; [1]'s Edit is [e]dit
	// too, each told by the panel it is on.
	preview := &terminu.Panel{Label: "Preview", Body: k.preview, Actions: []terminu.Action{
		{Key: "Enter", Label: "Show", Desc: "the mark as its app shows it", Scope: terminu.ItemScope, Hint: "show", Do: k.show},
		{Key: "e", Label: "Edit", Desc: "draw its cells", Scope: terminu.ItemScope, Do: k.draw},
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
	a.Start = k.start
	return k
}

// start checks that kamonu can read its own data before it is used (tdp E3,
// when it cannot be read): the directory of kamon, then each kamon. The
// first one it cannot read opens MOVE ASIDE: Enter puts it aside as
// <name>.yaml.bak and checks again, Esc leaves kamonu with the path and the
// reason on stderr. One at a time, so each says which file it is. A kamon is
// one file of many: it goes aside with nothing new in its place.
func (k *kamonu) start() tea.Cmd {
	names, err := k.st.names()
	if err != nil {
		return k.a.MoveAside(k.st.conf(), err, k.restart)
	}
	for _, n := range names {
		if _, err := k.st.load(n); err != nil {
			return k.a.MoveAsideRecord(k.st.path(n), err, k.restart)
		}
	}
	return nil
}

// restart is start after a file was put aside: the list without it.
func (k *kamonu) restart() tea.Cmd {
	k.reload("")
	return k.start()
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
		return k.a.ErrorAbout("Draw failed", name, "Can't read "+name+": "+why(err))
	}
	return k.openEditor(name, km)
}

// show plays the kamon under the cursor as its app's splash plays its mark
// (KM8): the name under it, no version or introduction, which kamonu does
// not know. Any key closes it.
func (k *kamonu) show() tea.Cmd {
	name := k.current()
	km, err := k.st.load(name)
	if err != nil {
		return k.a.ErrorAbout("Show failed", name, "Can't read "+name+": "+why(err))
	}
	s, err := terminu.NewSplash(km.svg(), name, "", "")
	if err != nil {
		return k.a.ErrorAbout("Show failed", name, "Can't show "+name+": "+err.Error())
	}
	return k.a.PlaySplash(s)
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
			k.a.Log(terminu.Info, "created "+name)
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
		return k.a.ErrorAbout("Edit failed", old, "Can't read "+old+": "+why(err))
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
				k.a.Log(terminu.Info, "renamed "+old+" to "+name)
			}
			k.a.Log(terminu.Info, "saved "+name+": letters "+letters+", U frame "+onOff(km.frame))
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
		return k.a.ErrorAbout("Export failed", name, "Can't export "+name+": "+why(err))
	}
	// <name>.svg is there: a menu asks whether to write over it or to write
	// a new file, whose name a text input asks for, the next free number
	// offered (KM10: one value is an input, not a form, components/dialog/
	// form). Esc on the input goes back to the menu; once written, the whole
	// stack closes.
	next := k.st.nextSVG(name)
	return k.a.MenuAbout("Export", name, terminu.MenuSection{Actions: []terminu.Action{
		{Label: "Overwrite " + name + ".svg", Desc: "write over the svg there", Do: func() tea.Cmd {
			if err := k.st.export(name, km); err != nil {
				return k.a.ErrorAbout("Export failed", name, "Can't export "+name+": "+why(err))
			}
			return k.exported(name)
		}},
		{Label: "Export as new…", Desc: "a file of another name; " + name + ".svg stays", Do: func() tea.Cmd {
			return k.a.Text(terminu.TextInput{
				Title: "Export as", Name: "File", Glyph: glyphExport, Verb: "export",
				Prompt:  "Name of the new svg; " + name + ".svg stays as it is",
				Suggest: next + ".svg",
				Check:   func(v string) error { return k.st.checkSVG(strings.TrimSuffix(v, ".svg")) },
				Commit: func(v string) error {
					if err := k.st.export(strings.TrimSuffix(v, ".svg"), km); err != nil {
						return errors.New("Can't export " + name + ": " + why(err))
					}
					return nil
				},
				Done: func(v string) tea.Cmd { return tea.Batch(k.a.CloseAll(), k.exported(strings.TrimSuffix(v, ".svg"))) },
			})
		}},
	}})
}

// exported says in a Success toast where the svg went (components/dialog/
// toast); a toast folds to whatever lines the path needs.
func (k *kamonu) exported(file string) tea.Cmd {
	return k.a.Toast(terminu.Success, "Exported to "+shown(k.st.svgPath(file)))
}

// deleteAsk is the confirm before name is deleted, what goes said first.
func (k *kamonu) deleteAsk(name string) terminu.Ask {
	return terminu.Ask{
		Title: "Delete", Question: "Delete " + name + "?", Verb: "delete",
		Details: []string{name + ".yaml is removed from " + shown(k.st.conf()) + "."}, Warn: []int{0},
	}
}

// delete runs once the confirm is accepted (Action.Ask).
func (k *kamonu) delete() tea.Cmd {
	name := k.current()
	if err := k.st.remove(name); err != nil {
		return k.a.ErrorAbout("Delete failed", name, "Can't delete "+name+": "+why(err))
	}
	k.a.Log(terminu.Info, "deleted "+name)
	k.reload("")
	return nil
}

// onOff is a yes or no for the App Log.
func onOff(on bool) string {
	if on {
		return "on"
	}
	return "off"
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
func (p *preview) HasItem() bool                                  { return p.k.current() != "" }
func (p *preview) Scroll() string                                 { return p.pager.Scroll() }

// MoveKeys are the pager's, which lists them only while the mark does not
// fit (tdp M6).
func (p *preview) MoveKeys() []terminu.KeyRow { return p.pager.MoveKeys() }
