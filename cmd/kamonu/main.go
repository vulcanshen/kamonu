// Command kamonu draws the terminu family's mark (家紋, kamon): a 25 × 19
// grid in a U frame, letter by letter, exported as the svg tdp S5 describes.
package main

import (
	"fmt"
	"os"

	"github.com/vulcanshen/kamonu/internal/app"
	"github.com/vulcanshen/kamonu/internal/version"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version", "--version", "-v":
			fmt.Println("kamonu " + version.Display())
			return
		case "help", "--help", "-h":
			fmt.Print(usage)
			return
		default:
			fmt.Fprintf(os.Stderr, "kamonu: unknown command %q\n\n%s", os.Args[1], usage)
			os.Exit(2)
		}
	}
	if err := app.Run(); err != nil {
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
