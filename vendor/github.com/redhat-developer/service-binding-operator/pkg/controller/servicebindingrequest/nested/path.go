package nested

import (
	"strconv"
	"strings"
)

// Field is the representation of a term in a field path.
//
// Assuming that 'status.dbCredentials' is a path for a value in a nested object, 'status' and
// 'dbCredentials' are represented as Field values.
//
// Field contains two members: 'Name' and 'Index'. 'Name' is a string like 'status' or
// 'dbCredentials', and 'Index' is the optional integer representation of the value, if it can be
// transformed to a valid positive integer.
type Field struct {
	// Name is the field name.
	Name string
	// Index is the integer representation of the field name in the case it can be converted to an
	// integer value.
	Index *int
}

// NewField creates a new field with the given name.
func NewField(name string) Field {
	f := Field{Name: name}
	if i, err := strconv.Atoi(name); err == nil {
		f.Index = &i
	}
	return f
}

// Path represents a field path.
type Path []Field

// Head returns the path head if one exists.
func (p Path) Head() (Field, bool) {
	if len(p) > 0 {
		return p[0], true
	}
	return Field{}, false
}

// Tail returns the path tail if present.
//
// Returns 'b.c' in the path 'a.b.c'.
func (p Path) Tail() Path {
	_, exists := p.Head()
	if !exists {
		return Path{}
	}
	return p[1:]
}

// HasTail asserts whether path has a tail.
func (p Path) HasTail() bool {
	return len(p.Tail()) > 0
}

// AdjustedPath adjusts the current path depending on the head element.
//
// In the case the head of a path ('a' in the 'a.b.c' path) exists and is different than '*', returns
// itself otherwise returns the path tail ('b.c' in the example).
func (p Path) AdjustedPath() Path {
	head, exists := p.Head()
	if !exists {
		return Path{}
	}
	if head.Name == "*" {
		return p.Tail()
	}
	return p
}

// LastField returns the last field from the receiver.
func (p Path) LastField() (Field, bool) {
	if len(p) > 0 {
		return p[len(p)-1], true
	}
	return Field{}, false
}

// BasePath returns the receiver's base path.
//
// For example, returns 'a.b' from the 'a.b.c' path.
func (p Path) BasePath() Path {
	if len(p) > 1 {
		return p[:len(p)-1]
	}
	return Path{}
}

// Decompose returns the receiver's base path and the last field.
func (p Path) Decompose() (Path, Field) {
	f, _ := p.LastField()
	b := p.BasePath()
	return b, f
}

// Clean creates a new Path without '*' or integer values.
//
// For example, returns 'a.b.c' for 'a.b.*.c' or 'a.b.1.c'.
func (p Path) Clean() Path {
	newPath := make(Path, 0)
	for _, f := range p.AdjustedPath() {
		if f.Index != nil {
			continue
		}
		if f.Name == "*" {
			continue
		}
		newPath = append(newPath, f)
	}
	return newPath
}

// NewPath creates a new path with the given string in the format 'a.b.c'.
func NewPath(s string) Path {
	parts := strings.Split(s, ".")
	return NewPathWithParts(parts)
}

// NewPathWithParts constructs a Path from given parts.
func NewPathWithParts(parts []string) Path {
	path := make(Path, len(parts))
	for i, p := range parts {
		path[i] = NewField(p)
	}
	return path
}
