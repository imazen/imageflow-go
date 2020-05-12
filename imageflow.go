package imageflow

// Steps is the builder for creating a operation
type Steps struct {
	vertex     []Step
	last       int
	innerGraph graph
}

// Decode is used to import a image
func (steps *Steps) Decode() {
	steps.innerGraph.AddVertex()
	steps.vertex = append(steps.vertex)
}

// Step specify different nodes
type Step interface {
	ToStep() interface{}
}

type edge struct {
	kind string
	to   int
}
type graph struct {
	inner [][]edge
}

func (gr *graph) AddVertex() {
	gr.inner = append(gr.inner, []edge{})
}

func (gr *graph) AddEdge(to int, from int, kind string) {
	gr.inner[from] = append(gr.inner[from], edge{to: to, kind: kind})
}
