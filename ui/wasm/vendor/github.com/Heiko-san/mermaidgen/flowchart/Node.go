package flowchart

import (
	"fmt"
	"strings"
)

type nodeShape string

// Shape definitions for Nodes as described at
// https://mermaidjs.github.io/flowchart.html#nodes--shapes.
// When added to a Flowchart or Subgraph, Nodes get the NShapeRect shape as the
// default.
const (
	NShapeRect      nodeShape = `["%s"]`
	NShapeRoundRect nodeShape = `("%s")`
	NShapeCircle    nodeShape = `(("%s"))`
	NShapeRhombus   nodeShape = `{"%s"}`
	NShapeFlagLeft  nodeShape = `>"%s"]`
)

// Node represents a single, unique node of the Flowchart graph.
// Create an instance of Node via Flowchart's or Subgraph's AddNode method, do
// not create instances directly. Already defined IDs can be looked up via
// Flowchart's GetNode method or iterated over via its ListNodes method.
type Node struct {
	id       string
	Shape    nodeShape  // The shape of this Node.
	Text     []string   // The body text, ID if no text is added.
	Link     string     // Optional URL for a click-hook.
	LinkText string     // Optional tooltip for the link.
	Style    *NodeStyle // Optional CSS style.
}

// ID provides access to the Node's readonly field id.
func (n *Node) ID() (id string) {
	return n.id
}

// Implements graphItem, see String() for further details.
func (n *Node) renderGraph() string {
	textbox := n.id
	if len(n.Text) > 0 {
		textbox = strings.Join(n.Text, "<br/>")
	}
	text := n.id + fmt.Sprintf(string(n.Shape), textbox) + "\n"
	if n.Style != nil {
		text += fmt.Sprintf("class %s %s\n", n.id, n.Style.id)
	}
	if n.Link != "" {
		linktxt := n.Link
		if n.LinkText != "" {
			linktxt = n.LinkText
		}
		text += fmt.Sprintf("click %s \"%s\" \"%s\"\n",
			n.id, n.Link, linktxt)
	}
	return text
}

// String renders this graph element to a node definition line.
// If Style member is set an additional class line will be created.
// If Link member is set an additional click line will be created.
func (n *Node) String() (renderedElement string) {
	return n.renderGraph()
}

// AddLines adds one or more lines of text to the Text member.
// This text gets rendered to the Node's body, separated by <br/>'s.
// If no text is added, the Node's ID is rendered to its body.
func (n *Node) AddLines(lines ...string) {
	n.Text = append(n.Text, lines...)
}
