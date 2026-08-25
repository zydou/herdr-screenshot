# resvg-go (vendored)

A SVG renderer written in Go & WASM depended on resvg without CGO.

Vendored from [kanrichan/resvg-go](https://github.com/kanrichan/resvg-go) at
`v0.0.2-0.20231001163256-63db194ca9f5` (resvg 0.35.0) and upgraded to
**resvg 0.45.1**, which adds color font rendering — required for the embedded
Noto Color Emoji to render in PNG output. Local changes on top of upstream:

- `internal/resvg.rs` / `internal/Cargo.toml`: resvg 0.45 API (text shaping
  happens during `Tree::from_data`, fontdb moves into `Options`), fontdb 0.23.
- `NewTreeFromData(svg, options, fontdb)` consumes the fontdb;
  `Tree.ConvertText` and the `resvg::Tree` round-trip are gone.
- wasm target renamed `wasm32-wasi` → `wasm32-wasip1`.
- `internal/vendor/usvg`: local patch of usvg 0.45.1 guarding an out-of-bounds
  `glyphs.remove()` in the per-glyph font fallback — a ZWJ emoji sequence
  (e.g. 👨‍👩‍👧) next to regular text panicked the wasm. Wired in via
  `[patch.crates-io]` in `internal/Cargo.toml`.

## Rebuilding the wasm

Requires a Rust toolchain with the `wasm32-wasip1` target:

```sh
go generate ./internal/resvg/...
```

which runs `cargo build --release --target wasm32-wasip1` inside
`internal/` and gzips the result to `internal/resvg.wasm.gz`.

## Usage

```go
worker, _ := NewDefaultWorker(context.Background())
defer worker.Close()

fontdb, _ := worker.NewFontDBDefault()
fontdb.LoadFontData(mapleMonoTTF)
fontdb.LoadFontData(notoColorEmojiTTF)

pixmap, _ := worker.NewPixmap(512, 512)
defer pixmap.Close()

// The fontdb is consumed here (text shaping happens during parse).
tree, _ := worker.NewTreeFromData(svg, &Options{}, fontdb)
defer tree.Close()
tree.Render(TransformFromScale(1, 1), pixmap)

png, _ := pixmap.EncodePNG()
```

## Thanks

- [resvg](https://github.com/linebender/resvg) - an SVG rendering library written in Rust
- [wazero](https://github.com/tetratelabs/wazero) - the zero dependency WebAssembly runtime for Go developers
