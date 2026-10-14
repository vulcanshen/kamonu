package main

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// store is the data directory (tdp E3): conf/<name>.yaml a kamon, and
// svg/<name>.svg what Export writes of it.
type store struct{ dir string }

// dataDir is KAMONU__DATA, or ~/.kamonu/data (tdp E2, E3).
func dataDir() (string, error) {
	if d := os.Getenv("KAMONU__DATA"); d != "" {
		return d, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".kamonu", "data"), nil
}

func (s store) conf() string               { return filepath.Join(s.dir, "conf") }
func (s store) path(name string) string    { return filepath.Join(s.conf(), name+".yaml") }
func (s store) svgPath(name string) string { return filepath.Join(s.dir, "svg", name+".svg") }

// shown is a path as the user reads it: the home directory as ~.
func shown(p string) string {
	if home, err := os.UserHomeDir(); err == nil {
		if rel, err := filepath.Rel(home, p); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.Join("~", rel)
		}
	}
	return p
}

// names are the kamon in the directory, in order; none when it does not
// exist yet.
func (s store) names() ([]string, error) {
	entries, err := os.ReadDir(s.conf())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if n, ok := strings.CutSuffix(e.Name(), ".yaml"); ok && !e.IsDir() && n != "" {
			out = append(out, n)
		}
	}
	slices.SortFunc(out, func(a, b string) int { return strings.Compare(strings.ToLower(a), strings.ToLower(b)) })
	return out, nil
}

func (s store) exists(name string) bool {
	_, err := os.Stat(s.path(name))
	return err == nil
}

func (s store) load(name string) (kamon, error) {
	b, err := os.ReadFile(s.path(name))
	if err != nil {
		return kamon{}, err
	}
	return parseKamon(b)
}

// create writes a new kamon; it fails when one by that name is there.
func (s store) create(name string, km kamon) error {
	if err := os.MkdirAll(s.conf(), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(s.path(name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	_, err = f.Write(km.yaml())
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}

// save writes km over the kamon.
func (s store) save(name string, km kamon) error { return replace(s.path(name), km.yaml()) }

// export writes the kamon's svg (tdp S5) to svg/<name>.svg, over the one
// there: it is made from the yaml, which is what is kept.
func (s store) export(name string, km kamon) error {
	p := s.svgPath(name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return replace(p, km.svg())
}

// replace writes b to a file beside p first, then in its place, so a failed
// write leaves the old one whole.
func replace(p string, b []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(p), "."+filepath.Base(p)+".*")
	if err != nil {
		return err
	}
	_, err = tmp.Write(b)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(tmp.Name(), 0o644)
	}
	if err == nil {
		err = os.Rename(tmp.Name(), p)
	}
	if err != nil {
		os.Remove(tmp.Name())
	}
	return err
}

func (s store) rename(old, name string) error {
	if s.exists(name) {
		return errTaken
	}
	return os.Rename(s.path(old), s.path(name))
}

func (s store) remove(name string) error { return os.Remove(s.path(name)) }

var errTaken = errors.New("Name is taken")

// why is what went wrong with a file, without its path: the user knows
// which kamon it is, and a save's path is a temporary file beside it.
func why(err error) string {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	var le *os.LinkError
	if errors.As(err, &le) {
		return le.Err.Error()
	}
	return err.Error()
}

// checkFormat is a name's own rules, checked in its input popup: it becomes
// a file name in the data directory.
func checkFormat(v string) error {
	switch {
	case strings.Contains(v, "/"):
		return errors.New("Name can't have /")
	case strings.HasPrefix(v, "."):
		return errors.New("Name can't start with .")
	}
	return nil
}

// checkName is every rule of a new name: given, its format, not taken.
func (s store) checkName(v string) error {
	if v == "" {
		return errors.New("Name can't be empty")
	}
	if err := checkFormat(v); err != nil {
		return err
	}
	if s.exists(v) {
		return errTaken
	}
	return nil
}
