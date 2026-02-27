# Go Binding for [Imageflow](https://github.com/imazen/imageflow)

![Windows](https://github.com/imazen/imageflow-go/workflows/Windows/badge.svg)![Macos](https://github.com/imazen/imageflow-go/workflows/Macos/badge.svg)![Linux](https://github.com/imazen/imageflow-go/workflows/Linux/badge.svg)

Quickly scale or modify images and optimize them for the web.

If the AGPLv3 does not work for you, you can get a commercial license on a sliding scale. If you have more than 1 server doing image processing your savings should cover the cost.

Docs are [here](https://pkg.go.dev/github.com/imazen/imageflow-go)

# Installation

Imageflow depends on `libimageflow` for image processing. Prebuilt shared libraries for Linux, macOS, and Windows are available from the [imageflow releases page](https://github.com/imazen/imageflow/releases/tag/v2.3.0-rc01). Add `libimageflow` to your OS library path. Then install with `go get`:

```bash
$ go get github.com/imazen/imageflow-go
```

# Usage

A simple go program to create two images of different size.

```go
package main

import (
	"os"

	imageflow "github.com/imazen/imageflow-go"
)

func main() {
	step := imageflow.NewStep()
	data, _ := step.
		Decode(imageflow.NewURL("https://jpeg.org/images/jpeg2000-home.jpg")).
		Branch(func(step *imageflow.Steps) {
			step.
				ConstrainWithin(200, 200).
				Encode(imageflow.NewFile("test_1.jpg"), imageflow.MozJPEG{})
		}).ConstrainWithin(400, 400).
		Encode(imageflow.GetBuffer("test"), imageflow.MozJPEG{}).
		Execute()
	os.WriteFile("test_2.jpeg", data["test"], 0644)
}
```
