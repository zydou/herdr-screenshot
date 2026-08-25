BINARY := herdr-screenshot
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64

# Pinned font sources — bump these to update the embedded fonts.
# Maple variant: Normal (full-width CJK as normal), NL (no ligatures), NF
# (Nerd Font), CN, unhinted — i.e. MapleMonoNormalNL-NF-CN-unhinted.zip.
MAPLE_VERSION := v7.9
NOTO_EMOJI_SHA := 8998f5dd683424a73e2314a8c1f1e359c19e8742

MAPLE_TTF := font/MapleMonoNormalNL-NF-CN-Regular.ttf
NOTO_TTF := font/NotoColorEmoji.ttf

.PHONY: build test cross clean fonts

$(MAPLE_TTF):
	@mkdir -p font
	curl -fL -o font/MapleMonoNormalNL-NF-CN-unhinted.zip \
		https://github.com/subframe7536/maple-font/releases/download/$(MAPLE_VERSION)/MapleMonoNormalNL-NF-CN-unhinted.zip
	unzip -p font/MapleMonoNormalNL-NF-CN-unhinted.zip MapleMonoNormalNL-NF-CN-Regular.ttf > $@
	rm font/MapleMonoNormalNL-NF-CN-unhinted.zip

$(NOTO_TTF):
	@mkdir -p font
	curl -fL -o $@ \
		https://raw.githubusercontent.com/googlefonts/noto-emoji/$(NOTO_EMOJI_SHA)/fonts/NotoColorEmoji.ttf

fonts: $(MAPLE_TTF) $(NOTO_TTF)

build: fonts
	CGO_ENABLED=0 go build -o $(BINARY) .

test: fonts
	go test ./...

cross: fonts
	mkdir -p dist
	@for platform in $(PLATFORMS); do \
		CGO_ENABLED=0 GOOS=$${platform%/*} GOARCH=$${platform#*/} go build -o dist/$(BINARY)-$${platform%/*}-$${platform#*/} . ; \
	done
	@ls -la dist/

clean:
	rm -rf dist $(BINARY)
