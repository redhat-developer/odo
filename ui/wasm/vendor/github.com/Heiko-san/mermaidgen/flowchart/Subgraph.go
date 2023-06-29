package flowchart

import (
	"fmt"
)

// Subgraph represents a subgraph block on the Flowchart graph where nested
// Subgraphs and Nodes can be added. Create an instance of Subgraph via
// Flowchart's or Subgraph's AddSubgraph method, do not create instances
// directly. Already defined IDs can be looked up via Flowchart's GetSubgraph
// method or iterated over via its ListSubgraphs method.
type Subgraph struct {
	id        string      // virtual ID for lookup
	flowchart *Flowchart  // top lvl pointer
	items     []graphItem // sub-items to render
	Title     string      // The title of this Subgraph.
}

// ID provides access to the Subgraph's readonly field id.
func (sg *Subgraph) ID() (id string) {
	return sg.id
}

// Flowchart provides access to the Subgraph's underlying Flowchart to be able
// to access Adder, Getter and Lister methods or to lookup Styles.
func (sg *Subgraph) Flowchart() (topLevel *Flowchart) {
	return sg.flowchart
}

// Implements graphItem, see String() for further details.
func (sg *Subgraph) renderGraph() string {
	text := fmt.Sprintln("subgraph", sg.Title)
	for _, item := range sg.items {
		text += item.renderGraph()
	}
	text += "end\n"
	return text
}

// String renders this graph element to a subgraph block.
func (sg *Subgraph) String() (renderedElement string) {
	return sg.renderGraph()
}

// AddSubgraph is used to add another nested Subgraph below this Subgraph layer.
// If the provided ID already exists, no new Subgraph is created and nil is
// returned. The ID can later be used to lookup the created Subgraph using
// Flowchart's GetSubgraph method.
func (sg *Subgraph) AddSubgraph(id string) (newSubgraph *Subgraph) {
	_, alreadyExists := sg.flowchart.subgraphs[id]
	if alreadyExists {
		return nil
	}
	s := &Subgraph{id: id, flowchart: sg.flowchart}
	sg.flowchart.subgraphs[id] = s
	sg.items = append(sg.items, s)
	return s
}

// AddNode is used to add a new Node to this Subgraph layer. If the provided ID
// already exists, no new Node is created and nil is returned. The ID can later
// be used to lookup the created Node using Flowchart's GetNode method.
func (sg *Subgraph) AddNode(id string) (newNode *Node) {
	_, alreadyExists := sg.flowchart.nodes[id]
	if alreadyExists {
		return nil
	}
	n := &Node{id: id, Shape: NShapeRect}
	sg.flowchart.nodes[id] = n
	sg.items = append(sg.items, n)
	return n
}
