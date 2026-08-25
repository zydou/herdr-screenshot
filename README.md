# herdr-screenshot

把 [herdr](https://github.com/herdrdev/herdr) pane 读取的 ANSI 文本渲染成 PNG 截图， 用于发送到 Telegram。

[freeze](https://github.com/charmbracelet/freeze) 的极简单用途裁剪版：无配置、无语法高亮、无交互， 只做一件事。

```sh
herdr pane read <pane_id> --source visible --format ansi > screen.ansi
herdr-screenshot screen.ansi -o screen.png
```

## 行为

- 渲染参数固定：字号 16、行高 1.2、无 padding/margin、背景 `#1a1b2c`、默认前景 `#C4C4C4`
- 输出仅 PNG，长边超过 2560px 时等比缩小（Telegram 服务端会缩到 2560，直接在本地缩省上传带宽）
- 内嵌 Maple Mono Normal NL NF CN（中文/powerline/Nerd Font 图标）与 Noto Color Emoji（彩色 emoji），
  二进制自包含，不依赖系统字体
- 静态链接（CGO_ENABLED=0），无任何运行时依赖
- 仅支持 linux 和 macOS

## 已知限制

- ZWJ 组合 emoji 序列（如 👨‍👩‍👧）显示为占位框而非组合字形——与上游 resvg 的
  per-glyph fallback 行为一致；单码点 emoji（✅💭🔧⏵️ 等）正常
- 畸形 SGR 序列可能触发 resvg 内部 panic（沿袭自 freeze 的行为）

## 构建

字体不进仓库，`make fonts` 按 Makefile 里 pin 的版本从上游下载到 `font/` （Maple Mono v7.9 release 资产 + noto-emoji 固定 commit），`build`/`test`/`cross` 都依赖它：

```sh
make fonts    # 下载内嵌字体（首次构建前需要）
make build    # 静态编译 herdr-screenshot
make test     # 单元测试
make cross    # 交叉编译 linux/macOS × amd64/arm64 到 dist/（CI release 用）
```

二进制约 38MB，其中 ~31MB 是两个内嵌字体。

## 许可

`internal/resvg` 内嵌 kanrichan/resvg-go（GPL-3.0）编译的 resvg wasm 及其 Rust 源码（含一个对 usvg 0.45.1 的本地补丁，见 `internal/resvg/internal/vendor/`）， 分发二进制时需遵循 GPL-3.0。
字体为 SIL OFL 1.1（Maple Mono、Noto Color Emoji）。
