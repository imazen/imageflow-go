package imageflow

// Decode is used to create a decode node in graph
type Decode struct {
	IoID int `json:"io_id"`
}

// ToStep is used to convert a Decode to step
func (decode *Decode) ToStep() map[string]interface{} {
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
func (encode *Encode) ToStep() map[string]interface{} {
	encodeMap := make(map[string]interface{})
	encodeMap["encode"] = encode
	return encodeMap
}

// MozJPG is used to encode using mozjpg library
type MozJPG struct {
	Quality     string `json:"quality"`
	Progressive bool   `json:"progressive"`
}

// ToPreset is used to convert the MozJPG to a preset
func (preset MozJPG) ToPreset() interface{} {
	presetMap := make(map[string]Preset)
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
type WebPLossless string

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
type Black string

// ToColor is used to convert to black color
func (black *Black) ToColor() string {
	return "Black"
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
	SharpenPercent    string      `json:"sharpen_percent"`
	DownFilter        string      `json:"down_filter"`
	UpFilter          string      `json:"up_filter"`
	ScalingColorspace string      `json:"scaling_colorspace"`
	BackgroundColor   interface{} `json:"background_color"`
	ResampleWhen      string      `json:"resample_when"`
	SharpenWhen       string      `json:"sharpen_when"`
}

// ToStep Converts the Constraint to a step
func (step Constrain) ToStep() interface{} {
	step.Hint.BackgroundColor = step.Hint.BackgroundColor.(Color).ToColor()
	step.Gravity = step.Gravity.(ConstraintGravity).ToGravity()
	step.CanvasColor = step.CanvasColor.(Color).ToColor()
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
	stepMap["region_precentage"] = region
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
