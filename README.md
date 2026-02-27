# imageflow-go

[![CI](https://github.com/imazen/imageflow-go/actions/workflows/ci.yml/badge.svg)](https://github.com/imazen/imageflow-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/imazen/imageflow-go.svg)](https://pkg.go.dev/github.com/imazen/imageflow-go)

Go bindings for [libimageflow](https://github.com/imazen/imageflow) — fast, safe image processing. Resize, crop, watermark, and encode images using a pipeline API that compiles down to optimized native code.

Licensed under AGPLv3. Commercial licenses available on a sliding scale — if you run more than one image processing server, the savings should cover the cost.

## Platform support

Tested in CI against libimageflow v2.3.0-rc01:

| Platform | Library | Runner |
|---|---|---|
| Linux x86_64 (glibc) | `libimageflow.so` | `ubuntu-24.04` |
| Linux ARM64 (glibc) | `libimageflow.so` | `ubuntu-24.04-arm` |
| Linux x86_64 (musl) | `libimageflow.a` (static) | `ubuntu-24.04` |
| Linux ARM64 (musl) | `libimageflow.a` (static) | `ubuntu-24.04-arm` |
| macOS ARM64 | `libimageflow.dylib` | `macos-latest` |
| macOS x86_64 | `libimageflow.dylib` | `macos-13` |
| Windows x86_64 | `imageflow.dll` | `windows-latest` |

## Installation

1. Download the appropriate libimageflow binary for your platform from the [v2.3.0-rc01 release](https://github.com/imazen/imageflow/releases/tag/v2.3.0-rc01).

2. Place the library where your linker and loader can find it (e.g. `/usr/local/lib`, or the working directory with `LD_LIBRARY_PATH=.`).

3. Install the Go package:

```bash
go get github.com/imazen/imageflow-go
```

## Quick start

Decode a JPEG, constrain it to 400px wide, and encode to WebP:

```go
package main

import (
	"log"
	"os"

	imageflow "github.com/imazen/imageflow-go"
)

func main() {
	step := imageflow.NewStep()
	results, err := step.
		Decode(imageflow.NewFile("input.jpg")).
		ConstrainWithinW(400).
		Encode(imageflow.GetBuffer("output"), imageflow.WebP{Quality: 80}).
		Execute()
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile("output.webp", results["output"], 0644)
}
```

## Examples

### Multiple outputs from a single decode

Generate a thumbnail, a medium image, and a full-size image in one pass:

```go
step := imageflow.NewStep()
results, err := step.
	Decode(imageflow.NewBuffer(inputBytes)).
	Branch(func(s *imageflow.Steps) {
		s.ConstrainWithin(150, 150).
			Encode(imageflow.GetBuffer("thumb"), imageflow.MozJPEG{Quality: 60})
	}).
	Branch(func(s *imageflow.Steps) {
		s.ConstrainWithin(800, 800).
			Encode(imageflow.GetBuffer("medium"), imageflow.MozJPEG{Quality: 80})
	}).
	Encode(imageflow.GetBuffer("full"), imageflow.MozJPEG{}).
	Execute()
```

### Watermark

```go
step := imageflow.NewStep()
results, err := step.
	Decode(imageflow.NewFile("photo.jpg")).
	Watermark(
		imageflow.NewFile("logo.png"),
		imageflow.ConstraintGravity{X: 100, Y: 100}, // bottom-right
		"within",
		imageflow.PercentageFitBox{X1: 70, Y1: 70, X2: 100, Y2: 100},
		0.5,  // opacity
		nil,
	).
	Encode(imageflow.NewFile("watermarked.jpg"), imageflow.MozJPEG{}).
	Execute()
```

### Crop, filter, and encode to multiple formats

```go
step := imageflow.NewStep()
results, err := step.
	Decode(imageflow.NewBuffer(inputBytes)).
	Region(imageflow.Region{
		X1: 100, Y1: 100, X2: 500, Y2: 500,
		BackgroundColor: imageflow.Black{},
	}).
	GrayscaleFlat().
	Contrast(1.2).
	Branch(func(s *imageflow.Steps) {
		s.Encode(imageflow.GetBuffer("png"), imageflow.LosslessPNG{})
	}).
	Encode(imageflow.GetBuffer("jpeg"), imageflow.MozJPEG{Quality: 85}).
	Execute()
```

### Command string API

Use querystring-style commands for simple operations:

```go
step := imageflow.NewStep()
results, err := step.
	Decode(imageflow.NewFile("input.jpg")).
	Command("width=300&height=200&mode=max&format=png").
	Encode(imageflow.GetBuffer("out"), imageflow.LosslessPNG{}).
	Execute()
```

## Encoding presets

| Preset | Format | Key options |
|---|---|---|
| `MozJPEG{Quality: 85, Progressive: true}` | JPEG | Quality 0-100, default 90 |
| `LosslessPNG{MaxDeflate: true}` | PNG | Lossless, optional max compression |
| `LossyPNG{Quality: 80, Speed: 5}` | PNG | pngquant, quality + speed tradeoff |
| `WebP{Quality: 80}` | WebP | Lossy, quality 0-100 |
| `WebPLossless{}` | WebP | Lossless |
| `GIF{}` | GIF | |

Shorthand methods: `.JPEG(out)`, `.PNG(out)`, `.WebP(out)`, `.GIF(out)`.

## Transforms

Rotation: `Rotate90()`, `Rotate180()`, `Rotate270()`

Flip: `FlipH()`, `FlipV()`

Constrain: `ConstrainWithin(w, h)`, `ConstrainWithinW(w)`, `ConstrainWithinH(h)`, `Constrain(opts)`

Crop/pad: `Region(...)`, `RegionPercentage(...)`, `CropWhitespace(threshold, padding)`

Canvas: `ExpandCanvas(...)`, `FillRect(x1, y1, x2, y2, color)`

Color filters: `GrayscaleFlat()`, `GrayscaleNTSC()`, `GrayscaleBT709()`, `GrayscaleRY()`, `Sepia()`, `Invert()`, `Alpha(v)`, `Contrast(v)`, `Brightness(v)`, `Saturation(v)`, `WhiteBalanceSRGB(threshold)`

Compositing: `DrawExact(fn, rect)`, `CopyRectangle(fn, rect)`, `Watermark(...)`

## I/O types

| Constructor | Input | Output |
|---|---|---|
| `NewFile("path")` | Reads from disk | Writes to disk |
| `NewBuffer([]byte)` | Decodes from memory | — |
| `GetBuffer("key")` | — | Returns bytes in result map |
| `NewURL("https://...")` | HTTP GET | — |
