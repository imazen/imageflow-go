package imageflow

import (
	"io/ioutil"
)

// Steps is the builder for creating a operation
type Steps struct {
	inputs     []ioOperation
	outputs    []ioOperation
	vertex     []Step
	last       uint
	innerGraph graph
	ioID       int
}

type ioOperation interface {
	toBuffer() []byte
	toOutput(*map[string][]byte) *map[string][]byte
	setIo(id uint, direction string)
}

func (file File) toBuffer() []byte {
	bytes, _ := ioutil.ReadFile(file.filename)
	return bytes
}

func (file File) toOutput(m *map[string][]byte) *map[string][]byte {
	return m
}

func (file *File) setIo(id uint, direction string) {
	file.IOID = id
	file.Direction = direction
	file.IO = "placeholder"
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
	task.setIo(uint(steps.ioID), "in")
	steps.innerGraph.AddVertex()
	steps.vertex = append(steps.vertex, Decode{
		IoID: steps.ioID,
	})
	steps.ioID++
	steps.last = uint(len(steps.vertex) - 1)
	return steps
}

// ConstrainWithin is used to constraint a image
func (steps *Steps) ConstrainWithin(w float64, h float64) *Steps {
	steps.input(Constrain{
		W: w,
		H: h,
	})
	return steps
}

// Encode is used to convert the image
func (steps *Steps) Encode(task ioOperation, preset Preset) *Steps {
	task.setIo(uint(steps.ioID), "out")
	steps.outputs = append(steps.outputs, task)
	steps.input(Encode{
		IoID:   steps.ioID,
		Preset: preset.ToPreset(),
	})
	steps.ioID++
	return steps
}

func (steps *Steps) input(step Step) {
	steps.innerGraph.AddVertex()
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
func (steps *Steps) Execute() {
	jsonMap := make(map[string]interface{})
	jsonMap["io"] = append(steps.inputs, steps.outputs...)
	graphMap := make(map[string]interface{})
	nodes := []interface{}{}
	for i := 0; i < len(steps.vertex); i++ {
		nodes = append(nodes, steps.vertex[i].ToStep())
	}
	frameWise := make(map[string]interface{})
	graphMap["nodes"] = nodes
	graphMap["edges"] = steps.innerGraph.edges
	frameWise["graph"] = graphMap
	jsonMap["framewise"] = frameWise
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

func (gr *graph) AddVertex() {
	//gr.inner = append(gr.inner, []uint{})
}

func (gr *graph) AddEdge(from uint, to uint, kind string) {
	gr.edges = append(gr.edges, edge{To: to, Kind: kind, From: from})
}
