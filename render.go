package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/beevik/etree"
	"github.com/charmbracelet/x/ansi"

	"github.com/zydou/herdr-screenshot/font"
	resvg "github.com/zydou/herdr-screenshot/internal/resvg"
)

// Fixed render parameters — herdr-screenshot has no configuration on purpose.
const (
	fontSize        = 16.0
	lineHeight      = 1.2
	background      = "#1a1b2c"
	foreground      = "#C4C4C4"                    // freeze's charm style default text colour
	fontFamily      = "Maple Mono Normal NL NF CN" // must match the TTF's own family name
	defaultFontSize = 14.0                         // chroma legacy constant, part of the height formula
	maxSide         = 2560.0                       // Telegram recompresses photos past this anyway
)

// clamp is inlined from freeze's cut.go.
func clamp(n, low, high int) int {
	if n < low {
		return low
	}
	if n > high {
		return high
	}
	return n
}

// lineCount replicates chroma's line semantics: SplitTokensIntoLines drops
// the empty trailing token line, so an input ending in "\n" has exactly as
// many lines as it has newlines.
func lineCount(input string) int {
	return len(strings.Split(strings.TrimSuffix(input, "\n"), "\n"))
}

// contentWidth measures the widest display line of ANSI-stripped input,
// tabs counted as 6 columns (freeze's measurement quirk), using
// ansi.StringWidth — the same primitive lipgloss.Width is built on.
func contentWidth(stripped string) float64 {
	tabWidth := 6
	longest := 0
	for _, l := range strings.Split(strings.ReplaceAll(stripped, "\t", strings.Repeat(" ", tabWidth)), "\n") {
		if w := ansi.StringWidth(l); w > longest {
			longest = w
		}
	}
	return float64(longest+1) * (fontSize / fontHeightToWidthRatio)
}

// buildSVG assembles the skeleton SVG and dispatches the ANSI stream into it,
// mirroring freeze's main.go pipeline for the fixed parameters. It returns
// the final image dimensions for the rasterizer.
func buildSVG(input string) (*etree.Document, float64, float64, error) {
	stripped := ansi.Strip(input)
	n := lineCount(stripped)

	// Height: chroma computed 10 + int(16.8*(lines+1)) at a 14px base; freeze
	// then rescaled it. Keep the exact expression order, truncation included.
	chromaH := 10 + int(16.8*float64(n+1))
	imageHeight := float64(chromaH)
	imageHeight *= 4 // scale: 4x supersampling for PNG output
	imageHeight *= (fontSize / defaultFontSize)
	imageHeight *= (lineHeight / lineHeight)

	// Width: auto, from the widest line.
	imageWidth := contentWidth(stripped) * 4

	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	svgEl := doc.CreateElement("svg")
	svgEl.CreateAttr("width", fmt.Sprintf("%.2f", imageWidth))
	svgEl.CreateAttr("height", fmt.Sprintf("%.2f", imageHeight))
	svgEl.CreateAttr("xmlns", "http://www.w3.org/2000/svg")

	terminal := svgEl.CreateElement("rect")
	terminal.CreateAttr("width", "100%")
	terminal.CreateAttr("height", "100%")
	terminal.CreateAttr("fill", background)

	g := svgEl.CreateElement("g")
	g.CreateAttr("font-family", fontFamily)
	g.CreateAttr("font-size", fmt.Sprintf("%.2fpx", fontSize*4))
	g.CreateAttr("fill", foreground)

	lines := make([]*etree.Element, n)
	for i := range lines {
		text := g.CreateElement("text")
		text.CreateAttr("xml:space", "preserve")
		// freeze positions each line at (padding+margin, (i+1)*fontSize*lineHeight);
		// both are zero here except the line height term.
		x := fmt.Sprintf("%.2fpx", 0.0)
		y := fmt.Sprintf("%.2fpx", float64(i+1)*(fontSize*lineHeight*4))
		text.CreateAttr("x", x)
		text.CreateAttr("y", y)
		lines[i] = text
	}

	cfg := &geom{Scale: 4, FontSize: fontSize, LineHeight: lineHeight * 4}
	d := dispatcher{scale: 4, cfg: cfg, svg: g, lines: lines}

	parser := ansi.NewParser()
	parser.SetHandler(ansi.Handler{
		Print:     d.Print,
		HandleCsi: d.CsiDispatch,
		Execute:   d.Execute,
	})
	for _, line := range strings.Split(input, "\n") {
		parser.Parse([]byte(line))
		d.Execute(ansi.LF) // simulate a newline
	}

	// Final dimensions, mirroring freeze's svg.Move/SetDimensions formatting.
	svgEl.SelectAttr("width").Value = fmt.Sprintf("%.2f", imageWidth)
	svgEl.SelectAttr("height").Value = fmt.Sprintf("%.2f", imageHeight)
	terminal.CreateAttr("x", fmt.Sprintf("%.2fpx", 0.0))
	terminal.CreateAttr("y", fmt.Sprintf("%.2fpx", 0.0))
	terminal.SelectAttr("width").Value = fmt.Sprintf("%.2f", imageWidth)
	terminal.SelectAttr("height").Value = fmt.Sprintf("%.2f", imageHeight)

	if p := os.Getenv("HERDR_DEBUG_SVG"); p != "" {
		if err := doc.WriteToFile(p); err != nil {
			return nil, 0, 0, err
		}
	}
	return doc, imageWidth, imageHeight, nil
}

// renderPNG rasterizes the SVG with the vendored resvg (color-emoji capable),
// capping the long side at 2560px.
func renderPNG(doc *etree.Document, w, h float64, output string) error {
	svg, err := doc.WriteToBytes()
	if err != nil {
		return fmt.Errorf("serialize SVG: %w", err)
	}

	worker, err := resvg.NewDefaultWorker(context.Background())
	if err != nil {
		return fmt.Errorf("start resvg: %w", err)
	}
	defer worker.Close() //nolint: errcheck

	fontdb, err := worker.NewFontDBDefault()
	if err != nil {
		return fmt.Errorf("create fontdb: %w", err)
	}
	// Noto Color Emoji is always loaded on top: resvg falls back to it per
	// character for glyphs the primary font lacks.
	if err := fontdb.LoadFontData(font.MapleMonoTTF); err != nil {
		return fmt.Errorf("load font: %w", err)
	}
	if err := fontdb.LoadFontData(font.NotoColorEmojiTTF); err != nil {
		return fmt.Errorf("load emoji font: %w", err)
	}

	scale := 1.0
	if w > maxSide || h > maxSide {
		scale = maxSide / max(w, h)
	}

	pixmap, err := worker.NewPixmap(uint32(w*scale), uint32(h*scale))
	if err != nil {
		return fmt.Errorf("create pixmap: %w", err)
	}
	defer pixmap.Close() //nolint: errcheck

	// NewTreeFromData consumes the fontdb (text shaping happens during parse).
	tree, err := worker.NewTreeFromData(svg, &resvg.Options{
		Dpi:                192,
		ShapeRenderingMode: resvg.ShapeRenderingModeGeometricPrecision,
		TextRenderingMode:  resvg.TextRenderingModeOptimizeLegibility,
		ImageRenderingMode: resvg.ImageRenderingModeOptimizeQuality,
		DefaultSizeWidth:   float32(w),
		DefaultSizeHeight:  float32(h),
	}, fontdb)
	if err != nil {
		return fmt.Errorf("parse SVG: %w", err)
	}
	defer tree.Close() //nolint: errcheck

	if err := tree.Render(resvg.TransformFromScale(float32(scale), float32(scale)), pixmap); err != nil {
		return fmt.Errorf("render: %w", err)
	}
	png, err := pixmap.EncodePNG()
	if err != nil {
		return fmt.Errorf("encode PNG: %w", err)
	}
	if err := os.WriteFile(output, png, 0o600); err != nil {
		return fmt.Errorf("write output: %w", err)
	}
	return nil
}
