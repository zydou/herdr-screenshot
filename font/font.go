// Package font embeds the two fonts rendered into every screenshot.
//
// Maple Mono Normal NL NF CN (SIL OFL 1.1, Reserved Font Name "Maple Mono") —
// a Nerd Font with full CJK coverage, so terminal snapshots render powerline
// glyphs and Chinese without any user-installed fonts. Fetched at build time
// by `make fonts` (see the pinned version in the Makefile).
// https://github.com/subframe7536/maple-font
//
// Noto Color Emoji (SIL OFL 1.1, Reserved Font Name "Noto") — CBDT color
// bitmap emoji, loaded into resvg's font database as the per-character
// fallback for glyphs the primary font lacks.
// https://github.com/googlefonts/noto-emoji
//
// Both licenses require their text to accompany redistribution; see the
// repositories above for the full OFL text.
package font

import (
	_ "embed"
)

//go:embed MapleMonoNormalNL-NF-CN-Regular.ttf
var MapleMonoTTF []byte

//go:embed NotoColorEmoji.ttf
var NotoColorEmojiTTF []byte
