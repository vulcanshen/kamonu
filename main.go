// Command kamonu draws the terminu family's mark (家紋, kamon): a 25 × 19
// grid in a U frame, letter by letter, exported as the svg tdp S5 describes.
package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	terminu "github.com/vulcanshen/terminu-go"
)

// version is set when it is built for a release (tdp E6).
var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Println("kamonu " + version)
			return
		case "help", "--help", "-h":
			fmt.Print(usage)
			return
		default:
			fmt.Fprintf(os.Stderr, "kamonu: unknown command %q\n\n%s", os.Args[1], usage)
			os.Exit(2)
		}
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "kamonu:", err)
		os.Exit(1)
	}
}

const usage = `kamonu — draw the terminu family's mark

Usage:
  kamonu            run it
  kamonu version    print the version (also --version, -v)
  kamonu help       print this (also --help, -h)

Environment:
  KAMONU__DATA                where the kamon are kept, in conf/, and their svg
                              Export writes, in svg/ (default ~/.kamonu/data)
  KAMONU__ICON_WIDTH          1 or 2: how many cells a Nerd Font icon takes,
                              instead of asking the terminal
  KAMONU__POWERLINE_WIDTH     1 or 2: the same for a powerline symbol
  TERMINU__ICON_WIDTH         the same as KAMONU__ICON_WIDTH, for every terminu app
  TERMINU__POWERLINE_WIDTH    the same as KAMONU__POWERLINE_WIDTH, for every terminu app
`

func run() error {
	dir, err := dataDir()
	if err != nil {
		return err
	}
	k := newKamonu(store{dir: dir})
	terminu.Probe(k.a.Name)
	_, err = tea.NewProgram(model{k}, tea.WithAltScreen(), tea.WithOutput(screenOutput{os.Stdout})).Run()
	return err
}

// model is the lib's App with kamonu's own popups over it: kamonu runs the
// program itself, so that it can draw them after the lib's frame (see
// popup.go).
type model struct{ k *kamonu }

func (m model) Init() tea.Cmd { return m.k.a.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k := m.k
	switch msg := msg.(type) {
	case animMsg:
		return m, k.step()
	case tea.KeyMsg:
		if l := k.top(); l != nil {
			return m, k.popKey(l, msg)
		}
	case tea.WindowSizeMsg:
		_, cmd := k.a.Update(msg)
		for _, l := range k.pops {
			k.place(l)
		}
		return m, cmd
	}
	_, cmd := k.a.Update(msg)
	return m, cmd
}

func (m model) View() string { return strings.Join(m.k.frame(), "\n") }

// screenOutput is the terminal minus every erase-to-the-end-of-the-line, as
// the lib's own App.Run writes it: every line is the full width (tdp L4),
// and when an icon takes two cells the erase would wipe the right frame
// (tdp E5). It embeds the file so bubbletea still finds the terminal.
type screenOutput struct{ *os.File }

var eraseRight = []byte("\x1b[K")

func (o screenOutput) Write(p []byte) (int, error) {
	if !bytes.Contains(p, eraseRight) {
		return o.File.Write(p)
	}
	_, err := o.File.Write(bytes.ReplaceAll(p, eraseRight, nil))
	return len(p), err
}

// icon is the mark the splash draws (tdp S5, E6): kamonu's own, K, A, M, O
// and N in the U frame, drawn and exported with kamonu.
//
//go:embed docs/icon.svg
var icon []byte

// splash is the family's easter egg, V on a panel (tdp S).
func splash() terminu.Splash {
	s, err := terminu.NewSplash(icon, "kamonu", version, "Draw the mark of a terminu family app")
	if err != nil {
		panic(err) // the embedded icon is broken: a mistake in the program
	}
	return s
}
