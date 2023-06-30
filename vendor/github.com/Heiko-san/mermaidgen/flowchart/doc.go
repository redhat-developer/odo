/*
Package flowchart is an object oriented approach to define mermaid flowcharts as
defined at https://mermaidjs.github.io/flowchart.html and render them to mermaid
code.

You use the constructor NewFlowchart to create a new Flowchart object.

	chart := flowchart.NewFlowchart()

This object is used to construct a graph using various class methods, such as:

	node1 := chart.AddNode("myNodeId1")
	node2 := chart.AddNode("myNodeId2")
	edge1 := chart.AddEdge(node1, node2)

Once the graph is completely defined, it can be "rendered" to mermaid code by
stringifying the Flowchart object.

	fmt.Print(chart)

Which then creates:

	graph TB
	myNodeId1["myNodeId1"]
	myNodeId2["myNodeId2"]
	myNodeId1 --> myNodeId2

The package supports defining CSS styles, that can be assigned to Nodes and
Edges.

	ns1 := chart.NodeStyle("myStyleId1")
	ns1.StrokeWidth = 3
	ns1.StrokeDash = 5
	es1 := chart.EdgeStyle("myStyleId1")
	es1.Stroke = "#f00"
	es1.StrokeWidth = 2
	node1.Style = ns1
	node2.Style = chart.NodeStyle("myStyleId1")
	edge1.Style = es1

There are various useful constants and further styling options, too.

	es1.Stroke = flowchart.ColorRed
	node1.Shape = flowchart.NShapeRoundRect
	node2.Shape = flowchart.NShapeCircle
	edge1.Shape = flowchart.EShapeThickLine

Let's add some text and maybe a link.

	node1.AddLines("my body text")
	node1.Link = "http://www.example.com"
	node1.LinkText = "go to example"

What we have by now is:

	graph TB
	classDef myStyleId1 stroke-width:3px,stroke-dasharray:5px
	myNodeId1("my body text")
	class myNodeId1 myStyleId1
	click myNodeId1 "http://www.example.com" "go to example"
	myNodeId2(("myNodeId2"))
	class myNodeId2 myStyleId1
	myNodeId1 === myNodeId2
	linkStyle 0 stroke-width:2px,stroke:#f00

And there is more. Just explore the package. Start at Flowchart and proceed to
Subgraph.
*/
package flowchart
