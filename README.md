# Go Binding for [Imageflow](https://github.com/imazen/imageflow)

![Windows](https://github.com/imazen/imageflow-go/workflows/Windows/badge.svg)![Macos](https://github.com/imazen/imageflow-go/workflows/Macos/badge.svg)![Linux](https://github.com/imazen/imageflow-go/workflows/Linux/badge.svg)

Quickly scale or modify images and optimize them for the web.

If the AGPLv3 does not work for you, you can get a commercial license on a sliding scale. If you have more than 1 server doing image processing your savings should cover the cost.

Docs are [here](https://pkg.go.dev/github.com/imazen/imageflow-go)

# Installation

Imageflow dependents on `libimageflow` for image processing capabilities. `libimageflow` is available as the static and dynamic shared library on Linux and macOS. Currently `libimageflow` is available as a dynamic library for Windows. Prebuilt shared libraries are available [here](https://github.com/imazen/imageflow/releases). Add `libimageflow`to OS path. Then it can be downloaded using`go get`.

```bash
$ go get github.com/imazen/imageflow-go
```

# Usage

A simple go program to create two image of different size.

```go
package main;

import (
	"io/ioutil"

	imageflow "github.com/imazen/imageflow-go"
)

func main(){
	step:=imageflow.NewStep()
	data,_:=step
	.Decode(imageflow.NewURL("https://jpeg.org/images/jpeg2000-home.jpg"))
	.Branch(func(step *imageflow.Steps){
		step
		.ConstrainWithin(200,200)
		.Encode(imageflow.NewFile("test_1.jpg"),imageflow.MozJPEG{})
	}).ConstrainWithin(400,400)
	.Encode(imageflow.GetBuffer("test"),imageflow.MozJPEG{})
	.Execute()
	ioutil.WriteFile("test_2.jpeg",data["test"],0644)
}

```
