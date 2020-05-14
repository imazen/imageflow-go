package imageflow

import (
	"encoding/json"
	"io/ioutil"
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
	toBuffer() []byte
	toOutput([]byte, *map[string][]byte) *map[string][]byte
	setIo(id uint)
	getIo() uint
}

func (file File) toBuffer() []byte {
	bytes, _ := ioutil.ReadFile(file.filename)
	return bytes
}

func (file File) toOutput(data []byte, m *map[string][]byte) *map[string][]byte {
	ioutil.WriteFile(file.filename, data, 0644)
	return m
}

func (file *File) setIo(id uint) {
	file.IOID = id
}

func (file File) getIo() uint {
	return file.IOID
}

// Buffer is io operation related to []byte
type Buffer struct {
	IOID   uint
	buffer []byte
}

func (file Buffer) toBuffer() []byte {
	return file.buffer
}

func (file Buffer) toOutput(data []byte, m *map[string][]byte) *map[string][]byte {
	return m
}

func (file *Buffer) setIo(id uint) {
	file.IOID = id
}

func (file Buffer) getIo() uint {
	return file.IOID
}

// File is io operation related to file
type File struct {
	IOID      uint `json:"io_id"`
	filename  string
	Direction string `json:"direction"`
	IO        string `json:"io"`
}

// Decode is used to import a image
func (steps *Steps) Decode(task ioOperation) *Steps {
	steps.inputs = append(steps.inputs, task)
	task.setIo(uint(steps.ioID))
	steps.vertex = append(steps.vertex, Decode{
		IoID: steps.ioID,
	}.ToStep())
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
	}.ToStep())
	steps.ioID++
	return steps
}

// Rotate90 is to used to rotate by 90 degrees
func (steps *Steps) Rotate90() *Steps {
	rotate := Rotate90{}
	steps.input(rotate.ToStep())
	return steps
}

// Rotate180 is to used to rotate by 180 degrees
func (steps *Steps) Rotate180() *Steps {
	rotate := Rotate180{}
	steps.input(rotate.ToStep())
	return steps
}

// Rotate270 is to used to rotate by 270 degrees
func (steps *Steps) Rotate270() *Steps {
	rotate := Rotate180{}
	steps.input(rotate.ToStep())
	return steps
}

// FlipH is to used to flip image horizontally
func (steps *Steps) FlipH() *Steps {
	rotate := FlipH{}
	steps.input(rotate.ToStep())
	return steps
}

// FlipV is to used to flip image horizontally
func (steps *Steps) FlipV() *Steps {
	rotate := FlipV{}
	steps.input(rotate.ToStep())
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
	steps.vertex = append(steps.vertex, step)
	steps.innerGraph.AddEdge(last, uint(len(steps.vertex)-1), "input")
	steps.innerGraph.AddEdge(steps.last, uint(len(steps.vertex)-1), "canvas")
	steps.last = uint(len(steps.vertex) - 1)
	return steps
}

// Execute the graph
func (steps *Steps) Execute() map[string][]byte {
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
		data := steps.inputs[i].toBuffer()
		job.AddInput(steps.inputs[i].getIo(), data)
	}
	for i := 0; i < len(steps.outputs); i++ {
		job.AddOutput(steps.outputs[i].getIo())
	}
	job.Message(js)

	bufferMap := make(map[string][]byte)
	for i := 0; i < len(steps.outputs); i++ {
		data := job.GetOutput(steps.outputs[i].getIo())
		steps.outputs[i].toOutput(data, &bufferMap)
	}
	return bufferMap
}

// Branch create a alternate path for the output
func (steps *Steps) Branch(f func(*Steps)) *Steps {
	last := steps.last
	f(steps)
	steps.last = last
	return steps
}

// Step specify different nodes
type Step interface {
	ToStep() interface{}
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
