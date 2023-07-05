package graph

import (
	"fmt"
	"sort"
	"strings"
)

type Graph struct {
	EntryNodeID string
	nodes       map[string]*Node
	edges       []*Edge
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
	}
}

func (o *Graph) AddNode(id string, text ...string) *Node {
	node := Node{
		ID:   id,
		Text: text,
	}
	o.nodes[id] = &node
	return &node
}

func (o *Graph) AddEdge(from *Node, to *Node, text ...string) *Edge {
	edge := Edge{
		From: from,
		To:   to,
		Text: text,
	}
	o.edges = append(o.edges, &edge)
	return &edge
}

func (o *Graph) ToFlowchart() string {
	var str strings.Builder
	str.WriteString("graph TB\n")
	texts := o.nodes[o.EntryNodeID].Text
	if len(texts) == 0 {
		texts = []string{o.EntryNodeID}
	}
	str.WriteString(fmt.Sprintf("%s[\"%s\"]\n", o.EntryNodeID, strings.Join(texts, "<br/>")))

	keys := make([]string, 0, len(o.nodes))
	for k := range o.nodes {
		keys = append(keys, k)

	}
	sort.Strings(keys)
	for _, key := range keys {
		node := o.nodes[key]
		if node.ID == o.EntryNodeID {
			continue
		}
		if len(node.Text) == 0 {
			node.Text = []string{node.ID}
		}
		str.WriteString(fmt.Sprintf("%s[\"%s\"]\n", node.ID, strings.Join(node.Text, "<br/>")))
	}

	for _, edge := range o.edges {
		str.WriteString(fmt.Sprintf("%s -->|\"%s\"| %s\n", edge.From.ID, strings.Join(edge.Text, "<br/>"), edge.To.ID))
	}

	return str.String()
}

type Node struct {
	ID   string
	Text []string
}

type Edge struct {
	From *Node
	To   *Node
	Text []string
}
