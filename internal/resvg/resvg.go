package resvg

import (
	"errors"

	"github.com/zydou/herdr-screenshot/internal/resvg/internal"
)

//go:generate go run internal/gen/gen.go

var (
	ErrWorkerIsBeingUsed = errors.New("worker is being used")
	ErrPointerIsNil      = errors.New("pointer is nil")
)

// Render render the SVG as a PNG by default
func (wk *Worker) Render(svg []byte) ([]byte, error) {
	if !wk.used.CompareAndSwap(false, true) {
		return nil, ErrWorkerIsBeingUsed
	}
	defer wk.used.Store(false)
	options, err := internal.UsvgOptionsDefault(wk.ctx, wk.mod)
	if err != nil {
		return nil, err
	}
	defer internal.UsvgOptionsDelete(wk.ctx, wk.mod, options)
	// An empty database: text renders only if the SVG carries no text nodes,
	// same as before — this convenience path never had fonts to offer.
	fontdb, err := internal.FontdbDatabaseDefault(wk.ctx, wk.mod)
	if err != nil {
		return nil, err
	}
	// Ownership of the fontdb moves into the tree.
	tree, err := internal.UsvgTreeFromData(wk.ctx, wk.mod, svg, options, fontdb)
	if err != nil {
		return nil, err
	}
	defer internal.UsvgTreeDelete(wk.ctx, wk.mod, tree)
	width, err := internal.UsvgTreeGetWidth(wk.ctx, wk.mod, tree)
	if err != nil {
		return nil, err
	}
	height, err := internal.UsvgTreeGetHeight(wk.ctx, wk.mod, tree)
	if err != nil {
		return nil, err
	}
	pixmap, err := internal.TinySkiaPixmapNew(wk.ctx, wk.mod, uint32(width), uint32(height))
	if err != nil {
		return nil, err
	}
	defer internal.TinySkiaPixmapDelete(wk.ctx, wk.mod, pixmap)
	transform, err := internal.TinySkiaTransformIdentity(wk.ctx, wk.mod)
	if err != nil {
		return nil, err
	}
	defer internal.TinySkiaTransformDelete(wk.ctx, wk.mod, transform)
	err = internal.ResvgTreeRender(wk.ctx, wk.mod, tree, transform, pixmap)
	if err != nil {
		return nil, err
	}
	return internal.TinySkiaPixmapEncodePng(wk.ctx, wk.mod, pixmap)
}
