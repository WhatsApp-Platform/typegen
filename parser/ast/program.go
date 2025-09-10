package ast

import (
	"fmt"
	"strings"
)

// ProgramNode represents the root of an AST
type ProgramNode struct {
	BaseNode
	Imports      []*ImportNode
	Declarations []Declaration
}

func (n *ProgramNode) String() string {
	var parts []string
	
	if len(n.Imports) > 0 {
		for _, imp := range n.Imports {
			parts = append(parts, imp.String())
		}
		parts = append(parts, "")
	}
	
	for _, decl := range n.Declarations {
		parts = append(parts, decl.String())
	}
	
	return strings.Join(parts, "\n")
}

// ImportNode represents an import statement
type ImportNode struct {
	BaseNode
	Path string
}

func (n *ImportNode) String() string {
	return fmt.Sprintf("import %s", n.Path)
}