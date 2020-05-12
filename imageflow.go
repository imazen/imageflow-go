package imageflow

import "fmt"

// Steps is the builder for creating a operation
type Steps struct {
	vertex     []Step
	last       uint
	innerGraph graph
	ioID       int
}

// Decode is used to import a image
func (steps *Steps) Decode() *Steps {
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
func (steps *Steps) Encode(preset Preset) *Steps {
	steps.input(Encode{
		IoID:   steps.ioID,
		Preset: preset.ToPreset(),
	})
	return steps
}

func (steps *Steps) input(step Step) {
	steps.innerGraph.AddVertex()
	steps.vertex = append(steps.vertex, step)
	steps.ioID++
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
	for i := 0; i < len(steps.innerGraph.edges); i++ {
		fmt.Printf("%+v\n", steps.innerGraph.edges[i])
	}
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
	kind string
	to   uint
	from uint
}
type graph struct {
	edges []edge
}

func (gr *graph) AddVertex() {
	//gr.inner = append(gr.inner, []uint{})
}

func (gr *graph) AddEdge(to uint, from uint, kind string) {
	gr.edges = append(gr.edges, edge{to: to, kind: kind, from: from})
}
