package app

import (
	"os"
	"path/filepath"
	"testing"

	terminu "github.com/vulcanshen/terminu-go"
	"github.com/vulcanshen/terminu-go/tdptest"
)

// newApp is kamonu as Run makes it, all but Run: its data where KAMONU__DATA
// or HOME says.
func newApp() *terminu.App {
	dir, err := dataDir()
	if err != nil {
		panic(err)
	}
	return newKamonu(store{dir: dir}).a
}

// TestTDP launches kamonu as for the first time on the sizes and icon widths
// tdptest walks, every screen's Space menu, global operations, App Log and
// key reference opened and closed (tdp L4, K4, L1).
func TestTDP(t *testing.T) {
	tdptest.Check(t, tdptest.App{Name: "kamonu", New: newApp})
}

// TestMoveAside: a kamon kamonu cannot read stops it at the start (tdp E3,
// when it cannot be read): Enter puts it aside as <name>.yaml.bak, makes
// nothing new and goes on to the next one; Esc leaves.
func TestMoveAside(t *testing.T) {
	st := store{dir: filepath.Join(t.TempDir(), "data")}
	if err := os.MkdirAll(st.conf(), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, b := range map[string][]byte{"aa": []byte("cells: ["), "bb": []byte("frame: ["), "filu": []byte(filYAML)} {
		if err := os.WriteFile(st.path(name), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	d := tdptest.Start(t, newKamonu(st).a, 80, 40)
	if !d.Has("MOVE ASIDE") || !d.Has("aa.yaml can't be read") || !d.Has("Move it to aa.yaml.bak and go on without it?") {
		t.Fatalf("no MOVE ASIDE for aa:\n%s", d.Screen())
	}
	d.Press("Enter")
	if _, err := os.Stat(st.path("aa") + ".bak"); err != nil {
		t.Fatal("aa.yaml was not put aside")
	}
	if _, err := os.Stat(st.path("aa")); err == nil {
		t.Fatal("a new aa.yaml was made in its place")
	}
	if !d.Has("bb.yaml can't be read") {
		t.Fatalf("no MOVE ASIDE for bb, the next one:\n%s", d.Screen())
	}
	d.Press("Enter")
	if d.Has("MOVE ASIDE") || !d.Has("filu") || d.Has("aa") {
		t.Fatalf("not back to the list without the two:\n%s", d.Screen())
	}

	if err := os.WriteFile(st.path("cc"), []byte("cells: ["), 0o644); err != nil {
		t.Fatal(err)
	}
	d = tdptest.Start(t, newKamonu(st).a, 80, 40)
	d.Press("Esc")
	if !d.Quit() {
		t.Fatal("Esc on MOVE ASIDE did not leave")
	}
	if _, err := os.Stat(st.path("cc")); err != nil {
		t.Fatal("Esc moved cc.yaml")
	}
}

// TestTDPWithMarks is the same walk with kamon in the data directory, the
// list and the preview full.
func TestTDPWithMarks(t *testing.T) {
	tdptest.Check(t, tdptest.App{Name: "kamonu", New: newApp, Seed: func(home string) {
		conf := filepath.Join(home, ".kamonu", "data", "conf")
		if err := os.MkdirAll(conf, 0o755); err != nil {
			panic(err)
		}
		for name, b := range map[string][]byte{"filu": []byte(filYAML), "blank": blank()} {
			if err := os.WriteFile(filepath.Join(conf, name+".yaml"), b, 0o644); err != nil {
				panic(err)
			}
		}
	}})
}
