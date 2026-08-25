package main

import (
	"fmt"
	"os"
	"strings"
)

const defaultOutput = "screen.png"

func usage() {
	fmt.Fprintf(os.Stderr, "usage: herdr-screenshot <file.ansi> [-o out.png]\n\nRender an ANSI text file (as read by `herdr pane read --format ansi`) to a PNG screenshot with embedded fonts.\n")
}

func main() {
	var output string
	var args []string
	rest := os.Args[1:]
	for i := 0; i < len(rest); i++ {
		arg := rest[i]
		switch {
		case arg == "-o":
			if i+1 >= len(rest) {
				fatal("flag needs an argument", fmt.Errorf("-o"))
			}
			i++
			output = rest[i]
		case strings.HasPrefix(arg, "-o="):
			output = strings.TrimPrefix(arg, "-o=")
		case arg == "-h" || arg == "--help":
			usage()
			return
		default:
			args = append(args, arg)
		}
	}
	if output == "" {
		output = defaultOutput
	}
	if len(args) != 1 {
		usage()
		os.Exit(2)
	}

	input, err := os.ReadFile(args[0])
	if err != nil {
		fatal("cannot read input", err)
	}
	if len(input) == 0 {
		fatal("no input", fmt.Errorf("%s is empty", args[0]))
	}

	doc, w, h, err := buildSVG(string(input))
	if err != nil {
		fatal("cannot build SVG", err)
	}
	if err := renderPNG(doc, w, h, output); err != nil {
		fatal("cannot render", err)
	}
	fmt.Println(output)
}

func fatal(msg string, err error) {
	fmt.Fprintf(os.Stderr, "herdr-screenshot: %s: %v\n", msg, err)
	os.Exit(1)
}
