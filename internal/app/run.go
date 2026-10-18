// Package app is kamonu: the list of kamon, the drawing, the store and the
// svg, on terminu-go.
package app

import (
	_ "embed"

	terminu "github.com/vulcanshen/terminu-go"

	"github.com/vulcanshen/kamonu/internal/version"
)

// Run opens kamonu on the kamon in its data directory.
func Run() error {
	dir, err := dataDir()
	if err != nil {
		return err
	}
	return newKamonu(store{dir: dir}).a.Run()
}

// icon is the mark the splash draws (tdp S5, E6): kamonu's own, K, A, M, O
// and N in the U frame, drawn and exported with kamonu. It is a copy of
// docs/icon.svg, which go:embed cannot reach; TestIconIsTheSplash keeps the
// two the same.
//
//go:embed icon.svg
var icon []byte

// splash is the family's easter egg, V on a panel (tdp S).
func splash() terminu.Splash {
	s, err := terminu.NewSplash(icon, "kamonu", version.Display(), "Draw the mark of a terminu family app")
	if err != nil {
		panic(err) // the embedded icon is broken: a mistake in the program
	}
	return s
}
