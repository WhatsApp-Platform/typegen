package ast

import "fmt"

// Position represents a position in the source code
type Position struct {
	Filename string
	Line     int
	Column   int
}

func (p Position) String() string {
	if p.Filename != "" {
		return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Node is the base interface for all AST nodes
type Node interface {
	Pos() Position
	String() string
}

// Declaration represents any top-level declaration
type Declaration interface {
	Node
	DeclNode()
}

// Type represents any type expression
type Type interface {
	Node
	TypeNode()
}

// BaseNode provides common functionality for AST nodes
type BaseNode struct {
	Position Position
}

func (n *BaseNode) Pos() Position {
	return n.Position
}