# Changelog

Each release names the version of terminu design (tdp) it follows.

## Unreleased

- `Ex[P]ort` asks first when the svg is there: **Overwrite** it, or **Export** a new file, at first `<name>_1.svg` (the
  first number not taken), or a name you give
- Draw: on the buttons `h|j|k|l` move to the one before or after, round, and `gg|G` to the first or the last
- Built on terminu-go v0.1.8

## v0.1.0 — tdp v0.6.3

- The repository: module `github.com/vulcanshen/kamonu`, GPL-3.0, READMEs in English and Traditional Chinese, built on
  terminu-go
- `[1] KAMON`, the marks in `~/.kamonu/data/conf` (`KAMONU__DATA`), and `[2] PREVIEW`, the one under the cursor drawn as
  the splash draws it
- `[A]dd` opens NEW KAMON: a name, the letters in the frame (at most 6, A to Z), and whether the U frame goes in;
  **Create** makes the file at once and opens it
- `[Enter] Draw` opens the cells: `hjkl` to move past the U frame, `Enter|Space` to turn a cell to the next letter,
  `1`–`6` to make it that letter or empty it again, `Backspace` to empty it, `x` to empty every cell (asks first), `p` to
  see it all gold as the splash will, each letter in a colour of its own with a row saying which digit and colour is
  which, **Save** and **Save As**; `Esc` with changes asks first
- `[E]dit` changes a mark's name, letters and U frame; the cells keep their places
- `Ex[P]ort` writes the svg to `~/.kamonu/data/svg`: tdp S5's, a 250 × 250 viewBox, the U frame and then each letter after
  its comment, all gold
- `[X] Delete` (asks first)
- A mark is kept as yaml, its cells a picture of the mark, one row a line
- `kamonu version` and `kamonu help`
