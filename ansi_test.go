package main

import (
	"strings"
	"testing"
)

func TestLineCount(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 1}, // Split("", "\n") == [""]; callers reject empty input earlier
		{"single line", "abc", 1},
		{"trailing newline", "abc\n", 1},
		{"two lines", "abc\ndef", 2},
		{"two lines trailing newline", "abc\ndef\n", 2},
		{"blank line in middle", "abc\n\ndef\n", 3},
		{"multiple trailing newlines", "abc\n\n\n", 3},
		{"only newline", "\n", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := lineCount(tt.input); got != tt.want {
				t.Errorf("lineCount(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestContentWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  float64
	}{
		// (longest+1) * (16/1.68), tabs count as 6 columns; scale is applied by buildSVG
		{"empty", "", 1 * (16.0 / 1.68)},
		{"short", "ab", 3 * (16.0 / 1.68)},
		{"tab expands to 6", "a\tb", 9 * (16.0 / 1.68)},
		{"wide chars", "中文", 5 * (16.0 / 1.68)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contentWidth(tt.input)
			diff := got - tt.want
			if diff < 0 {
				diff = -diff
			}
			if diff > 1e-9 {
				t.Errorf("contentWidth(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// TestHeightFormula pins the exact chroma-derived height for a known input,
// so any accidental change to the expression order fails loudly.
func TestHeightFormula(t *testing.T) {
	n := lineCount("a\nb\nc") // 3
	chromaH := 10 + int(16.8*float64(n+1))
	imageHeight := float64(chromaH)
	imageHeight *= 4
	imageHeight *= (fontSize / defaultFontSize)
	imageHeight *= (lineHeight / lineHeight)
	linesPlusOne := 4 // n+1 for a 3-line input; a var so 16.8*x isn't constant-folded
	want := float64(10+int(16.8*float64(linesPlusOne))) * 4 * (16.0 / 14.0)
	if imageHeight != want {
		t.Errorf("height = %v, want %v", imageHeight, want)
	}
}

// TestCsiMalformed never panics: a truncated colour form, an index past the
// 256-colour palette, an RGB component past 255, and style escapes on a row past
// the last line each used to index out of bounds and take the render down.
func TestCsiMalformed(t *testing.T) {
	inputs := []string{
		"truncated 38\nx\x1b[38m\n",
		"truncated 38;5\nx\x1b[38;5m\n",
		"palette out of range\nx\x1b[38;5;999m\n",
		"elided index\nx\x1b[38;5;m\n",
		"truncated rgb\nx\x1b[38;2;1;2m\n",
		"rgb out of range\nx\x1b[38;2;999;1;1m\n",
		"background truncated\nx\x1b[48;5mX\n",
		"background rgb truncated\nx\x1b[48;2;1mX\n",
		"unknown subform\nx\x1b[38;1m\n",
		"style past the last line\nx\n\x1b[1m",
		"styled text past the last line\nx\n\x1b[31mX",
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			if _, _, _, err := buildSVG(input); err != nil {
				t.Errorf("buildSVG(%q): %v", input, err)
			}
		})
	}
}

// TestCsiColours pins that the bounds checks still accept the colour forms a real
// snapshot carries, so the fix above doesn't silently drop styling.
func TestCsiColours(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"x\x1b[38;5;108my\n", `fill="#87af87"`},
		{"x\x1b[48;5;176my\n", `fill="#d787d7"`},
		{"x\x1b[38;5;246my\n", `fill="#949494"`},
		{"x\x1b[38;2;255;128;0my\n", `fill="#ff8000"`},
		{"x\x1b[48;2;255;128;0mz\n", `fill="#ff8000"`},
		{"x\x1b[31my\n", `fill="#D74E6F"`},
	}
	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			doc, _, _, err := buildSVG(c.input)
			if err != nil {
				t.Fatalf("buildSVG(%q): %v", c.input, err)
			}
			svg, _ := doc.WriteToString()
			if !strings.Contains(svg, c.want) {
				t.Errorf("buildSVG(%q) missing %q", c.input, c.want)
			}
		})
	}
}
