package imageflow

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Steps is the builder for creating a operation
type Steps struct {
	inputs     []ioOperation
	outputs    []ioOperation
	vertex     []interface{}
	last       uint
	innerGraph graph
	ioID       int
}

type ioOperation interface {
	toBuffer() ([]byte, error)
	toOutput([]byte, *map[string][]byte) *map[string][]byte
	setIo(id uint)
	getIo() uint
}

func (file File) toBuffer() ([]byte, error) {
	bytes, errorInRead := ioutil.ReadFile(file.Filename)
	if errorInRead != nil {
		return nil, errorInRead
	}
	return bytes, nil
}

func (file File) toOutput(data []byte, m *map[string][]byte) *map[string][]byte {
	ioutil.WriteFile(file.Filename, data, 0644)
	return m
}

func (file *File) setIo(id uint) {
	file.iOID = id
}

func (file File) getIo() uint {
	return file.iOID
}

func (file URL) toBuffer() ([]byte, error) {
	bytes, errorInURL := http.Get(file.URL)
	if errorInURL != nil {
		return nil, errorInURL
	}
	data, errorInRead := ioutil.ReadAll(bytes.Body)
	if errorInRead != nil {
		return nil, errorInRead
	}
	return data, nil
}

func (file URL) toOutput(data []byte, m *map[string][]byte) *map[string][]byte {
	return m
}

func (file *URL) setIo(id uint) {
	file.iOID = id
}

func (file URL) getIo() uint {
	return file.iOID
}

// URL is used to make a http request to get file and use it
type URL struct {
	URL  string
	iOID uint
}

// Buffer is io operation related to []byte
type Buffer struct {
	iOID   uint
	Buffer []byte
}

func (file Buffer) toBuffer() ([]byte, error) {
	return file.Buffer, nil
}

func (file Buffer) toOutput(data []byte, m *map[string][]byte) *map[string][]byte {
	return m
}

func (file *Buffer) setIo(id uint) {
	file.iOID = id
}

func (file Buffer) getIo() uint {
	return file.iOID
}

// File is io operation related to file
type File struct {
	iOID     uint
	Filename string
}

// Decode is used to import a image
func (steps *Steps) Decode(task ioOperation) *Steps {
	steps.inputs = append(steps.inputs, task)
	task.setIo(uint(steps.ioID))
	steps.vertex = append(steps.vertex, Decode{
		IoID: steps.ioID,
	}.toStep())
	steps.ioID++
	steps.last = uint(len(steps.vertex) - 1)
	return steps
}

// ConstrainWithin is used to constraint a image
func (steps *Steps) ConstrainWithin(w float64, h float64) *Steps {
	steps.input(constrainWithinMap(w, h))
	return steps
}

// ConstrainWithinH is used to constraint a image
func (steps *Steps) ConstrainWithinH(h float64) *Steps {
	steps.input(constrainWithinMap(nil, h))
	return steps
}

// ConstrainWithinW is used to constraint a image
func (steps *Steps) ConstrainWithinW(w float64) *Steps {
	steps.input(constrainWithinMap(w, nil))
	return steps
}

func constrainWithinMap(w interface{}, h interface{}) map[string]interface{} {
	constrainMap := make(map[string]interface{})
	dataMap := make(map[string]interface{})
	dataMap["mode"] = "within"
	if w != nil {
		dataMap["w"] = w
	}
	if h != nil {
		dataMap["h"] = h
	}
	constrainMap["constrain"] = dataMap

	return constrainMap
}

// Encode is used to convert the image
func (steps *Steps) Encode(task ioOperation, preset Preset) *Steps {
	task.setIo(uint(steps.ioID))
	steps.outputs = append(steps.outputs, task)
	steps.input(Encode{
		IoID:   steps.ioID,
		Preset: preset.ToPreset(),
	}.toStep())
	steps.ioID++
	return steps
}

// Rotate90 is to used to rotate by 90 degrees
func (steps *Steps) Rotate90() *Steps {
	rotate := Rotate90{}
	steps.input(rotate.toStep())
	return steps
}

// Rotate180 is to used to rotate by 180 degrees
func (steps *Steps) Rotate180() *Steps {
	rotate := Rotate180{}
	steps.input(rotate.toStep())
	return steps
}

// Rotate270 is to used to rotate by 270 degrees
func (steps *Steps) Rotate270() *Steps {
	rotate := Rotate180{}
	steps.input(rotate.toStep())
	return steps
}

// FlipH is to used to flip image horizontally
func (steps *Steps) FlipH() *Steps {
	rotate := FlipH{}
	steps.input(rotate.toStep())
	return steps
}

// FlipV is to used to flip image horizontally
func (steps *Steps) FlipV() *Steps {
	rotate := FlipV{}
	steps.input(rotate.toStep())
	return steps
}

func (steps *Steps) input(step interface{}) {
	steps.vertex = append(steps.vertex, step)
	steps.innerGraph.AddEdge(steps.last, uint(len(steps.vertex)-1), "input")
	steps.last = uint(len(steps.vertex) - 1)
}

func (steps *Steps) canvas(f func(*Steps), step Step) *Steps {
	last := steps.last
	f(steps)
	steps.vertex = append(steps.vertex, step.toStep())
	steps.innerGraph.AddEdge(last, uint(len(steps.vertex)-1), "input")
	steps.innerGraph.AddEdge(steps.last, uint(len(steps.vertex)-1), "canvas")
	steps.last = uint(len(steps.vertex) - 1)
	return steps
}

// CopyRectangle copy a image
func (steps *Steps) CopyRectangle(f func(steps *Steps), rect RectangleToCanvas) *Steps {
	return steps.canvas(f, rect)
}

// DrawExact copy a image
func (steps *Steps) DrawExact(f func(steps *Steps), rect DrawExact) *Steps {
	return steps.canvas(f, rect)
}

// Execute the graph
func (steps *Steps) Execute() (map[string][]byte, error) {
	jsonMap := make(map[string]interface{})
	graphMap := make(map[string]interface{})
	nodeMap := make(map[int]interface{})
	for i := 0; i < len(steps.vertex); i++ {
		nodeMap[i] = steps.vertex[i]
	}
	frameWise := make(map[string]interface{})
	graphMap["nodes"] = nodeMap
	graphMap["edges"] = steps.innerGraph.edges
	frameWise["graph"] = graphMap
	jsonMap["framewise"] = frameWise
	js, _ := json.Marshal(jsonMap)
	job := New()
	for i := 0; i < len(steps.inputs); i++ {
		data, errorInBuffer := steps.inputs[i].toBuffer()
		if errorInBuffer != nil {
			return nil, errorInBuffer
		}
		errorInInput := job.AddInput(steps.inputs[i].getIo(), data)
		if errorInInput != nil {
			return nil, errorInInput
		}
	}
	for i := 0; i < len(steps.outputs); i++ {
		errorInOutput := job.AddOutput(steps.outputs[i].getIo())
		if errorInOutput != nil {
			return nil, errorInOutput
		}
	}
	errorInMessage := job.Message(js)

	if errorInMessage != nil {
		return nil, errorInMessage
	}

	bufferMap := make(map[string][]byte)
	for i := 0; i < len(steps.outputs); i++ {
		data, errorInOutput := job.GetOutput(steps.outputs[i].getIo())
		if errorInOutput != nil {
			return nil, errorInOutput
		}
		steps.outputs[i].toOutput(data, &bufferMap)
	}
	return bufferMap, nil
}

// Branch create a alternate path for the output
func (steps *Steps) Branch(f func(*Steps)) *Steps {
	last := steps.last
	f(steps)
	steps.last = last
	return steps
}

// Region is used to crop or add padding to image
func (steps *Steps) Region(region Region) *Steps {
	steps.input(region.toStep())
	return steps
}

// RegionPercentage is used to crop or add padding to image using percentage
func (steps *Steps) RegionPercentage(region RegionPercentage) *Steps {
	steps.input(region.toStep())
	return steps
}

// CropWhitespace is used to remove whitespace around the image
func (steps *Steps) CropWhitespace(threshold int, padding float64) *Steps {
	steps.input(CropWhitespace{Threshold: threshold, PercentagePadding: padding}.toStep())
	return steps
}

// FillRect is used create a rectangle on the image
func (steps *Steps) FillRect(x1 float64, y1 float64, x2 float64, y2 float64, color Color) *Steps {
	steps.input(FillRect{X1: x1, Y1: y1, X2: x2, Y2: y2, Color: color}.toStep())
	return steps
}

// ExpandCanvas is used create a rectangle on the image
func (steps *Steps) ExpandCanvas(canvas ExpandCanvas) *Steps {
	steps.input(canvas.toStep())
	return steps
}

// Watermark is used to watermark a image
func (steps *Steps) Watermark(data ioOperation, gravity interface{}, fitMode string, fitBox FitBox, opacity float32, hint interface{}) *Steps {
	data.setIo(uint(steps.ioID))
	steps.inputs = append(steps.inputs, data)
	steps.input(Watermark{
		IoID:    uint(steps.ioID),
		Gravity: gravity,
		FitMode: fitMode,
		FitBox:  fitBox,
	}.toStep())
	steps.ioID++
	return steps
}

// Command is used to execute query like strings
func (steps *Steps) Command(cmd string) *Steps {
	cmdMap := make(map[string]map[string]string)
	dataMap := make(map[string]string)
	dataMap["kind"] = "ir4"
	dataMap["value"] = cmd
	cmdMap["command_string"] = dataMap
	steps.input(cmdMap)
	return steps
}

// WhiteBalanceSRGB histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) WhiteBalanceSRGB(threshold float32) *Steps {
	steps.input(doubleMap("white_balance_histogram_area_threshold_srgb", "threshold", threshold))
	return steps
}

// GrayscaleNTSC histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) GrayscaleNTSC() *Steps {
	return steps.colorFilterSRGB("grayscale_ntsc")
}

// GrayscaleFlat histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) GrayscaleFlat() *Steps {
	return steps.colorFilterSRGB("grayscale_flat")
}

// GrayscaleBT709 histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) GrayscaleBT709() *Steps {
	return steps.colorFilterSRGB("grayscale_bt709")
}

// GrayscaleRY histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) GrayscaleRY() *Steps {
	return steps.colorFilterSRGB("grayscale_ry")
}

// Sepia histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) Sepia() *Steps {
	return steps.colorFilterSRGB("sepia")
}

// Invert histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) Invert() *Steps {
	return steps.colorFilterSRGB("invert")
}

func (steps *Steps) colorFilterSRGB(value string) *Steps {
	steps.input(singleMap("color_filter_srgb", value))
	return steps
}

// Alpha histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) Alpha(value float32) *Steps {
	return steps.colorFilterSRGBValue("alpha", value)
}

// Contrast histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) Contrast(value float32) *Steps {
	return steps.colorFilterSRGBValue("contrast", value)
}

// Brightness histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) Brightness(value float32) *Steps {
	return steps.colorFilterSRGBValue("brightness", value)
}

// Saturation histogram area
// This command is not recommended as it operates in the sRGB space and does not produce perfect results.
func (steps *Steps) Saturation(value float32) *Steps {
	return steps.colorFilterSRGBValue("saturation", value)
}

func (steps *Steps) colorFilterSRGBValue(name string, value float32) *Steps {
	steps.input(doubleMap("color_filter_srgb", name, value))
	return steps
}

// Step specify different nodes
type Step interface {
	toStep() interface{}
}

type edge struct {
	Kind string `json:"kind"`
	To   uint   `json:"to"`
	From uint   `json:"from"`
}
type graph struct {
	edges []edge
}

func (gr *graph) AddEdge(from uint, to uint, kind string) {
	gr.edges = append(gr.edges, edge{To: to, Kind: kind, From: from})
}

// NewStep creates a step that can be used to specify how graph should be processed
func NewStep() Steps {
	return Steps{
		vertex: []interface{}{},
		last:   0,
		ioID:   0,
		innerGraph: graph{
			edges: []edge{},
		},
	}
}
