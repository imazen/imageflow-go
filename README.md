# imageflow-go
![Windows](https://github.com/imazen/imageflow-go/workflows/Windows/badge.svg)![Macos](https://github.com/imazen/imageflow-go/workflows/Macos/badge.svg)![Linux](https://github.com/imazen/imageflow-go/workflows/Linux/badge.svg)

# Usage Example

```go
	step := NewStep()
	step.Decode(&File{filename: "image.jpg"}).ConstrainWithinW(400).Branch(func(step *Steps) {
		step.ConstrainWithin(100, 100).Encode(&File{filename: "small.jpg"}, MozJPEG{})
	}).Encode(&File{filename: "medium.jpg"}, MozJPEG{}).Execute()

```