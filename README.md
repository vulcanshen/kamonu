# kamonu

**Language**: English · [繁體中文](README-zh_TW.md)

[![License: GPL-3.0](https://img.shields.io/badge/license-GPL--3.0-blue)](LICENSE)

**Draw the mark of a terminu family app, letter by letter, in the terminal.**

Every app of the terminu family has its own version of the family mark (家紋, *kamon*): the same blue U frame, with the
app's letters in gold inside it, on a grid of 25 × 19 cells. With kamonu you say which letters go in the frame, draw
each one in a colour of its own with the keyboard, and export the svg the family uses, ready to become an app's
`docs/icon.svg`: every letter gold, and marked one by one in the order of the app's name.

## Features

- A list of your marks beside a preview of the one under the cursor, as the family draws its marks: gold letters in
  the blue U frame.
- A new mark starts with the U frame in place, or blank if you like.
- Each letter in a colour of its own while you draw, so you always see which cell belongs to which letter.
- Draw with `hjkl` and `Space`; the U frame cannot be changed by mistake, and the cursor steps over it.
- Change a mark's name, letters or U frame later, without losing the drawing.
- Export writes the svg in one key press, each letter marked in its order.
- Save over the mark, or save the drawing as a new one and keep the old.
- Nothing is lost by accident: leaving a drawing with changes asks first, and so does deleting a mark.

## Install

> kamonu is **macOS / Linux only** (WSL works). There is no Windows build.

**Homebrew** (macOS / Linux):

```bash
brew install vulcanshen/tap/kamonu
```

**Install script** (the latest release binary into `~/.local/bin`, or `/usr/local/bin` as root):

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/kamonu/main/install.sh | sh
```

Building from source is in [`docs/dev-remarks.md`](docs/dev-remarks.md).

**Requirements**: a terminal with 24-bit colour (truecolor) and a [Nerd Font](https://www.nerdfonts.com/).

Some fonts make an icon two cells wide, and some make the powerline symbols (the ends of the capsules on the frames, the
arrow after the cursor) two cells wide too. kamonu measures both when it starts. If it gets one wrong, set it yourself:
`KAMONU__ICON_WIDTH` and `KAMONU__POWERLINE_WIDTH` (`1` or `2`) for kamonu alone, or `TERMINU__ICON_WIDTH` and
`TERMINU__POWERLINE_WIDTH` for the whole terminu family: set them once and every app of the family reads them. A family
app that runs a program in its own terminal sets the two `TERMINU__` variables for it.

**The first time it starts**, the list is empty: nothing is written until you add your first mark.

### Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/vulcanshen/kamonu/main/uninstall.sh | sh
```

Removes the binary, then asks — never assumes — about `~/.kamonu`, where your marks are. Installed with Homebrew:
`brew uninstall kamonu`, and delete `~/.kamonu` yourself if you do not want your marks any more.

## Quick start

1. Run `kamonu`.
2. Press `A`. On **Name** press `Enter`, type the app's name (`filu`) and press `Enter`; press `j` to go to
   **Letters**, press `Enter`, type the letters that go in the frame (`fil`: the U is the frame) and press `Enter`.
   Press `Tab` to go to **Create** and `Enter`.
3. Move with `hjkl`. On each cell of a letter, press the letter's digit: `1` for the F, `2` for the I, `3` for the L.
   The row under the cells says which digit and colour is which letter.
4. Press `Tab` to go to the buttons and `Enter` on **Save**.
5. Press `P` to export. The svg is now `~/.kamonu/data/svg/filu.svg`: copy it to the app's `docs/icon.svg`.

## Usage

### The list

| Key | What it does |
|---|---|
| `Enter` | Draw the mark under the cursor |
| `E` | Edit its name, its letters and whether it has the U frame |
| `P` | Export its svg; when there is one already, a menu asks: **Overwrite** it, or **Export as new** under a name you give |
| `X` | Delete the mark (asks first) |
| `A` | Add a new mark: its name, its letters, and whether it starts with the U frame |
| `Space` | Everything you can do here |
| `?` | The keys of this panel |
| `Tab\|1–2` | The other panel |
| `q` | Quit |

On **[2] PREVIEW**, `e` and `Enter` draw the mark it shows. When the terminal is too short for the whole mark, `j|k`,
`u|d` and `gg|G` scroll it.

### Drawing

| Key | What it does |
|---|---|
| `h\|j\|k\|l` | Move a cell left, down, up, right; the cursor steps over the U frame |
| `Enter\|Space` | Turn the cell to the next letter; after the last one, back to empty |
| `Backspace` | Empty the cell, whatever letter it is |
| `1`–`6` | Make the cell that letter: `1` the first, `2` the second, …; the same digit again empties it |
| `p` | Show every letter gold, as the svg has them; `p` again for their colours |
| `x` | Empty every cell (asks first) |
| `Tab` | Between the cells and the buttons |
| `h\|j\|k\|l` | On the buttons: between **Save As** and **Save** |
| `Enter` | On a button: **Save** writes over the mark, **Save As** asks for a new name and leaves the old mark as it was |
| `?` | The keys of the drawing |
| `Esc` | Leave; asks first when something changed |

### Changing the letters

Each letter keeps its cells by its place. Change `FIL` to `FAL` and the cells of the second letter are now the A's,
still in the shape of the I; change `KB` to `FIL` and the K's cells are the F's, the B's are the I's, and the L has none
yet. With fewer letters than before, the cells of the letters that are gone are emptied; putting the U frame in empties
the cells under it. Draw them again as you need.

## Settings and where things are kept

kamonu has no settings file. Everything is kept in `~/.kamonu/data`; set `KAMONU__DATA` to keep it somewhere else.

| Where | What |
|---|---|
| `conf/<name>.yaml` | The mark: its letters, whether it has the U frame, and its cells, drawn as a picture you can read |
| `svg/<name>.svg` | What **Export** writes. When it is there, Export asks: **Overwrite** it, or **Export as new**, under a name you give (`<name>_1.svg`, the first number not taken, is suggested) |

## Limitations

- A mark is always the family's 25 × 19 cells, in the family's colours: the blue U frame and gold letters. The
  colours of the letters are only for drawing; the svg has them all gold.
- At most 6 letters, each one of A to Z.
- Renaming or deleting a mark leaves the svg already exported as it is.

## Links

- [CHANGELOG](CHANGELOG.md)
- [Developer remarks](docs/dev-remarks.md)

## The terminu family

kamonu follows [terminu design](https://github.com/vulcanshen/terminu), the design the terminu family of terminal apps
share. The others:

| App | What it is |
|---|---|
| [kbu](https://github.com/vulcanshen/kbu) | A Kubernetes TUI dashboard |
| [filu](https://github.com/vulcanshen/filu) | A terminal file manager you don't have to learn |
| [sshu](https://github.com/vulcanshen/sshu) | A terminal front end for ssh and sftp |
| [webu](https://github.com/vulcanshen/webu) | A terminal browser that reads a web page as a document |
| [locku](https://github.com/vulcanshen/locku) | A screensaver with a PIN lock, for the terminal |
| [ccdrawu](https://github.com/vulcanshen/ccdrawu) | A sidebar and a status line for Claude Code |

## License

[GPL-3.0](LICENSE)
