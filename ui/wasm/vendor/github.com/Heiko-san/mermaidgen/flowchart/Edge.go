package flowchart

import (
	"fmt"
	"strconv"
	"strings"
)

type edgeShape string

// Shape definitions for Edges as described at
// https://mermaidjs.github.io/flowchart.html#links-between-nodes.
// When added to a Flowchart, Edges get the EShapeArrow shape as the default.
const (
	EShapeArrow       edgeShape = `-->`
	EShapeDottedArrow edgeShape = `-.->`
	EShapeThickArrow  edgeShape = `==>`
	EShapeLine        edgeShape = `---`
	EShapeDottedLine  edgeShape = `-.-`
	EShapeThickLine   edgeShape = `===`
)

// Edge represents a connection between 2 Nodes.
// Create an instance of Edge via Flowchart's AddEdge method, do not create
// instances directly. Already defined IDs (indices) can be looked up via
// Flowchart's GetEdge method or iterated over via its ListEdges method.
type Edge struct {
	id    int
	From  *Node      // Pointer to the Node where the Edge starts.
	To    *Node      // Pointer to the Node where the Edge ends.
	Shape edgeShape  // The shape of this Edge.
	Text  []string   // Optional text lines to be added along the Edge.
	Style *EdgeStyle // Optional CSS style.
}

// ID provides access to the Edge's readonly field id.
// Since Edges don't really have IDs, this actually is the index when the Edge
// was added, which is used for linkStyle lines.
func (e *Edge) ID() (id int) {
	return e.id
}

// String renders this graph element to an edge definition line.
// If Style member is set an additional linkStyle line will be created.
func (e *Edge) String() (renderedElement string) {
	line := string(e.Shape)
	if len(e.Text) > 0 {
		line += fmt.Sprintf(`|"%s"|`, strings.Join(e.Text, "<br/>"))
	}
	text := fmt.Sprintf("%s %s %s\n", e.From.id, line, e.To.id)
	if e.Style != nil {
		text += fmt.Sprintf(e.Style.String(), strconv.Itoa(e.id))
	}
	return text
}

// AddLines adds one or more lines of text to the Text member.
// This text gets rendered along the Edge, separated by <br/>'s.
func (e *Edge) AddLines(lines ...string) {
	e.Text = append(e.Text, lines...)
}
