# Changelog

Each release names the version of terminu design (tdp) it follows.

## Unreleased

- Export, when the svg is there already: a menu asks whether to **Overwrite** it or **Export as new**; Export as new asks
  for the file's name (the next free `<name>_<n>.svg` is given as an example) and refuses one that is there with
  `File is taken: …`. It was a form with a File field and two buttons
- The export action is called Export again (it was ExPort)
- `[2] PREVIEW`: `e` draws the mark it shows, and so does `Enter`; on a terminal too short for the whole mark it scrolls
  with `j|k`, `u|d` and `gg|G` and says which rows it shows. The mark is drawn from the panel's top left, as other
  panels show their content; a mark it cannot read says so in red, and `Enter` shows why
- Draw on a terminal under 27 rows (80 × 24, say): the cells scroll with the cursor, and the letters and the buttons
  stay in view; they were cut off before
- Draw: a cell between each letter's square and the letter, and two between letters; on a screen too narrow for
  them on one row, the rest go on the next
- Export: a failure opens an error popup with all of it, instead of a toast that could cut it
- An error popup's title says what failed: `SAVE FAILED`, `DRAW FAILED`
- Edit: the Letters prompt says whose letters they are
- Delete from the Space menu: `Esc` on the confirm goes back to the menu
- With no mark in the list, its key reference lists no moving keys
- Text in a confirm, a toast and an error popup wraps at spaces, and between wide characters (terminu-go v0.1.13)
- `go install github.com/vulcanshen/kamonu/cmd/kamonu@latest`: the program is in `cmd/kamonu` now, and the release
  files are the terminu family's

## v0.1.2 — tdp v0.6.11

- Draw opens on the lib's popup stack (terminu-go `App.Popup`): the confirms, Save As, error popups and key reference
  over it are the lib's, as everywhere else in the family; it looks and works as before
- Save As: `Enter` with nothing typed closes it, as an input with no old value does
- Built on terminu-go v0.1.10 (tdp v0.6.11)

## v0.1.1 — tdp v0.6.9

- `Ex[P]ort` asks first when the svg is there: **Overwrite** it, or **Export** a new file, at first `<name>_1.svg` (the
  first number not taken), or a name you give
- Draw: on the buttons `h|j|k|l` move to the one before or after, round, and `gg|G` to the first or the last
- Built on terminu-go v0.1.8

## v0.1.0 — tdp v0.6.3

- The repository: module `github.com/vulcanshen/kamonu`, GPL-3.0, READMEs in English and Traditional Chinese, built on
  terminu-go
- `[1] KAMON`, the marks in `~/.kamonu/data/conf` (`KAMONU__DATA`), and `[2] PREVIEW`, the one under the cursor drawn as
  the family draws its marks, all gold
- `[A]dd` opens NEW KAMON: a name, the letters in the frame (at most 6, A to Z), and whether the U frame goes in;
  **Create** makes the file at once and opens it
- `[Enter] Draw` opens the cells: `hjkl` to move past the U frame, `Enter|Space` to turn a cell to the next letter,
  `1`–`6` to make it that letter or empty it again, `Backspace` to empty it, `x` to empty every cell (asks first), `p` to
  see it all gold as the svg has it, each letter in a colour of its own with a row saying which digit and colour is
  which, **Save** and **Save As**; `Esc` with changes asks first
- `[E]dit` changes a mark's name, letters and U frame; the cells keep their places
- `Ex[P]ort` writes the svg to `~/.kamonu/data/svg`: tdp S5's, a 250 × 250 viewBox, the U frame and then each letter after
  its comment, all gold
- `[X] Delete` (asks first)
- A mark is kept as yaml, its cells a picture of the mark, one row a line
- `kamonu version` and `kamonu help`
