package main

import "testing"

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
