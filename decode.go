package imageflow

// Decode is used to create a decode node in graph
type Decode struct {
	IoID int `json:"io_id"`
}

// ToStep is used to convert a Decode to step
func (decode Decode) ToStep() map[string]interface{} {
	decodeMap := make(map[string]interface{})
	decodeMap["decode"] = decode
	return decodeMap
}

// Preset is a interface for encoder used to convert to image
type Preset interface {
	ToPreset() interface{}
}

// Encode is used to convert to a image
type Encode struct {
	IoID   int         `json:"io_id"`
	Preset interface{} `json:"preset"`
}

// ToStep is used to convert a Encode to step
func (encode Encode) ToStep() interface{} {
	encodeMap := make(map[string]interface{})
	encodeMap["encode"] = encode
	return encodeMap
}

// MozJPEG is used to encode using mozjpeg library
type MozJPEG struct {
	Quality     uint `json:"quality"`
	Progressive bool `json:"progressive"`
}

// ToPreset is used to convert the MozJPG to a preset
func (preset MozJPEG) ToPreset() interface{} {
	presetMap := make(map[string]Preset)
	if preset.Quality == 0 {
		preset.Quality = 100
	}
	presetMap["mozjpeg"] = preset
	return presetMap
}

// GIF is used to encode to gif
type GIF string

// ToPreset is used to convert the GIF to preset
func (gif GIF) ToPreset() string {
	return "gif"
}

// LosslessPNG is a encoder for lodepng
type LosslessPNG struct {
	MaxDeflate bool `json:"max_deflate"`
}

// ToPreset is used to LosslessPNG to Preset
func (preset LosslessPNG) ToPreset() interface{} {
	presetMap := make(map[string]Preset)
	presetMap["lodepng"] = preset
	return presetMap
}

// LossyPNG is used for encoding pngquant
type LossyPNG struct {
	Quality        int  `json:"quality"`
	MinimumQuality int  `json:"minimum_quality"`
	Speed          int  `json:"speed"`
	MaximumDeflate bool `json:"maximum_deflate"`
}

// ToPreset is used to convert LossPNG to preset
func (preset LossyPNG) ToPreset() interface{} {
	presetMap := make(map[string]Preset)
	presetMap["pngquant"] = preset
	return presetMap
}

// WebP is used to encode image using webp encoder
type WebP struct {
	Quality int `json:"quality"`
}

// ToPreset is used to convert WebP to preset
func (preset WebP) ToPreset() interface{} {
	presetMap := make(map[string]Preset)
	presetMap["webplossy"] = preset
	return presetMap
}

// WebPLossless is used to encode using webplossless encoder
type WebPLossless struct{}

// ToPreset is used to convert WebPLossless to preset
func (preset WebPLossless) ToPreset() string {
	return "webplossless"
}

// Constrain is used to specify constraints for the image
// W The width constraint in pixels
// H The height constraint in pixels
// Mode A constraint mode
// Gravity determines how the image is anchored when cropped or padded. {x: 0, y: 0} represents top-left, {x: 50, y: 50} represents center, {x:100, y:100} represents bottom-right. Default: center
// Hints See resampling hints
// Canvas_color See Color. The color of padding added to the image.
type Constrain struct {
	Mode        string         `json:"mode"`
	W           float64        `json:"w"`
	H           float64        `json:"h"`
	Hint        ConstraintHint `json:"hints"`
	Gravity     interface{}    `json:"gravity"`
	CanvasColor interface{}    `json:"canvas_color"`
}

// ConstraintGravity anchors the image
type ConstraintGravity struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// ToGravity is used to convert to gravity
func (gravity ConstraintGravity) ToGravity() interface{} {
	gravityMap := make(map[string]ConstraintGravity)
	gravityMap["percentage"] = gravity
	return gravityMap
}

// Color must be implemented by all the Colors that can be used by imageflow
type Color interface {
	ToColor() interface{}
}

// Black is the Implementation of interface Color and used as color black
type Black struct{}

// ToColor is used to convert to black color
func (black Black) ToColor() interface{} {
	return "black"
}

// Transparent is the Implementation of interface Color and used as color transparent
type Transparent string

// ToColor is used to convert Transparent to Color
func (color Transparent) ToColor() string {
	return "transparent"
}

// ConstraintHint is used to provided changes when resampling
// SharpenPercent (0..100) The amount of sharpening to apply during resampling
// UpFilter The resampling filter to use if upscaling in one or more directions
// DownFilter The resampling filter to use if downscaling in both directions.
// ScalingColorspace Use linear for the best results, or srgb to mimick poorly-written software. srgb can destroy image highlights.
// BackgroundColor The background color to apply.
// ResampleWhen One of size_differs, size_differs_or_sharpening_requested, or always.
// SharpenWhen One of downscaling, upscaling, size_differs, or always
// Supported Filters
// robidoux_sharp - A sharper version of the above
// robidoux_fast - A faster, less accurate version of robidoux
// ginseng - The default and suggested upsampling filter
// ginseng_sharp
// lanczos
// lanczos_sharp
// lanczos_2
// lanczos_2_sharp
// cubic
// cubic_sharp
// catmull_rom
// mitchell
// cubic_b_spline
// hermite
// jinc
// triangle
// linear
// box
// fastest
// n_cubic
// n_cubic_sharp
type ConstraintHint struct {
	SharpenPercent    interface{} `json:"sharpen_percent"`
	DownFilter        interface{} `json:"down_filter"`
	UpFilter          interface{} `json:"up_filter"`
	ScalingColorspace interface{} `json:"scaling_colorspace"`
	BackgroundColor   interface{} `json:"background_color"`
	ResampleWhen      interface{} `json:"resample_when"`
	SharpenWhen       interface{} `json:"sharpen_when"`
}

// ToStep Converts the Constraint to a step
func (step Constrain) ToStep() interface{} {
	if step.Hint.BackgroundColor != nil {
		step.Hint.BackgroundColor = step.Hint.BackgroundColor.(Color).ToColor()
	}
	if step.Gravity != nil {
		step.Gravity = step.Gravity.(ConstraintGravity).ToGravity()
	}
	if step.CanvasColor != nil {
		step.CanvasColor = step.CanvasColor.(Color).ToColor()
	}
	stepMap := make(map[string]Step)
	stepMap["constrain"] = step
	return stepMap
}

// Region is like a crop command, but you can specify coordinates outside of the image and thereby add padding.
// It's like a window.
type Region struct {
	X1              float64     `json:"x1"`
	Y1              float64     `json:"y1"`
	X2              float64     `json:"x2"`
	Y2              float64     `json:"y2"`
	BackgroundColor interface{} `json:"background_color"`
}

// ToStep create a step from Region
func (region Region) ToStep() interface{} {
	region.BackgroundColor = region.BackgroundColor.(Color).ToColor()
	stepMap := make(map[string]Step)
	stepMap["region"] = region
	return stepMap
}

// RegionPercentage is like a crop command, but you can specify coordinates outside of the image and thereby add padding.
// It's like a window.
type RegionPercentage struct {
	X1              float64     `json:"x1"`
	Y1              float64     `json:"y1"`
	X2              float64     `json:"x2"`
	Y2              float64     `json:"y2"`
	BackgroundColor interface{} `json:"background_color"`
}

// ToStep create a step from Region
func (region RegionPercentage) ToStep() interface{} {
	region.BackgroundColor = region.BackgroundColor.(Color).ToColor()
	stepMap := make(map[string]Step)
	stepMap["region_percent"] = region
	return stepMap
}

// CropWhitespace remove whitespace at the edges
// Threshold: 1..255 determines how much noise/edges to tolerate before cropping is finalized. 80 is a good starting point.
// PercentPadding determines how much of the image to restore after cropping to provide some padding. 0.5 (half a percent) is a good starting point.
type CropWhitespace struct {
	Threshold         int     `json:"threshold"`
	PercentagePadding float64 `json:"percentage_padding"`
}

// ToStep create a step from Region
func (region CropWhitespace) ToStep() interface{} {
	stepMap := make(map[string]Step)
	stepMap["crop_whitespace"] = region
	return stepMap
}

// Rotate90 rotate the image by 90 degree
type Rotate90 struct{}

// ToStep is used to convert the rotate to step
func (rotate Rotate90) ToStep() string {
	return "rotate_90"
}

// Rotate180 rotate the image by 90 degree
type Rotate180 struct{}

// ToStep is used to convert the rotate to step
func (rotate Rotate180) ToStep() string {
	return "rotate_180"
}

// Rotate270 rotate the image by 90 degree
type Rotate270 struct{}

// ToStep is used to convert the rotate to step
func (rotate Rotate270) ToStep() string {
	return "rotate_270"
}

// FlipH is used to flip the image horizontally
type FlipH struct{}

// FlipV is used to flip the image vertical
type FlipV struct{}

// ToStep is used to convert the rotate to step
func (rotate FlipH) ToStep() string {
	return "flip_h"
}

// ToStep is used to convert the rotate to step
func (rotate FlipV) ToStep() string {
	return "flip_v"
}

// FillRect is  used to fill a rectangle
type FillRect struct {
	X1    float64     `json:"x1"`
	Y1    float64     `json:"y1"`
	X2    float64     `json:"x2"`
	Y2    float64     `json:"y2"`
	Color interface{} `json:"color"`
}

// ToStep create a step from FillRect
func (region FillRect) ToStep() interface{} {
	stepMap := make(map[string]Step)
	region.Color = region.Color.(Color).ToColor()
	stepMap["fill_rect"] = region
	return stepMap
}

// ExpandCanvas is used to expand the image
type ExpandCanvas struct {
	Left   float64     `json:"left"`
	Right  float64     `json:"right"`
	Top    float64     `json:"top"`
	Bottom float64     `json:"bottom"`
	Color  interface{} `json:"color"`
}

// ToStep create a step from FillRect
func (region ExpandCanvas) ToStep() interface{} {
	stepMap := make(map[string]Step)
	region.Color = region.Color.(Color).ToColor()
	stepMap["expand_canvas"] = region
	return stepMap
}

// Watermark is used to create a watermark
type Watermark struct {
	IoID    uint        `json:"io_id"`
	Gravity interface{} `json:"gravity"`
	FitMode string      `json:"fit_mode"`
	FitBox  interface{} `json:"fit_box"`
	Opacity float32     `json:"opacity"`
	Hints   interface{} `json:"hints"`
}

// FitBox is used to specify image position
type FitBox interface {
	toFitBox() interface{}
}

// MarginFitBox is used to specify image position
type MarginFitBox struct {
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
}

// PercentageFitBox is used to specify image position
type PercentageFitBox struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

func (percent PercentageFitBox) toFitBox() interface{} {
	fitMap := make(map[string]FitBox)
	fitMap["image_percentage"] = percent
	return fitMap
}

func (percent MarginFitBox) toFitBox() interface{} {
	fitMap := make(map[string]FitBox)
	fitMap["image_margins"] = percent
	return fitMap
}

// ToStep is used to convert watermark
func (watermark Watermark) ToStep() interface{} {
	if watermark.FitMode == "" {
		watermark.FitMode = "within"
	}
	if watermark.Opacity == 0 {
		watermark.Opacity = 1
	}

	if watermark.FitBox != nil {
		watermark.FitBox = watermark.FitBox.(FitBox).toFitBox()
	}
	stepMap := make(map[string]Step)
	if watermark.Gravity != nil {
		watermark.Gravity = watermark.Gravity.(ConstraintGravity).ToGravity()
	}
	if watermark.Hints != nil {
		watermark.Hints = watermark.Hints.(ConstraintHint)
	}
	stepMap["watermark"] = watermark
	return stepMap
}

func singleMap(name string, value interface{}) map[string]interface{} {
	returnMap := make(map[string]interface{})
	returnMap[name] = value
	return returnMap
}

func doubleMap(first string, second string, value interface{}) map[string]interface{} {
	return singleMap(first, singleMap(second, value))
}
