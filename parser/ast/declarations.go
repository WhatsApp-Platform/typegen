package ast

import (
	"fmt"
	"strings"
)

// StructNode represents a struct declaration
type StructNode struct {
	BaseNode
	Name   string
	Fields []*FieldNode
}

func (n *StructNode) DeclNode() {}

func (n *StructNode) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("struct %s {", n.Name))
	
	for _, field := range n.Fields {
		parts = append(parts, fmt.Sprintf("  %s", field.String()))
	}
	
	parts = append(parts, "}")
	return strings.Join(parts, "\n")
}

// FieldNode represents a field in a struct
type FieldNode struct {
	BaseNode
	Name     string
	Type     Type
	Optional bool
}

func (n *FieldNode) String() string {
	if n.Optional {
		return fmt.Sprintf("%s: ?%s", n.Name, n.Type.String())
	}
	return fmt.Sprintf("%s: %s", n.Name, n.Type.String())
}

// EnumNode represents an enum declaration
type EnumNode struct {
	BaseNode
	Name     string
	Variants []*EnumVariantNode
}

func (n *EnumNode) DeclNode() {}

func (n *EnumNode) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("enum %s {", n.Name))
	
	for _, variant := range n.Variants {
		parts = append(parts, fmt.Sprintf("  %s", variant.String()))
	}
	
	parts = append(parts, "}")
	return strings.Join(parts, "\n")
}

// EnumVariantNode represents a variant in an enum
type EnumVariantNode struct {
	BaseNode
	Name    string
	Payload Type
}

func (n *EnumVariantNode) String() string {
	if n.Payload != nil {
		return fmt.Sprintf("%s: %s", n.Name, n.Payload.String())
	}
	return n.Name
}

// TypeAliasNode represents a type alias declaration
type TypeAliasNode struct {
	BaseNode
	Name string
	Type Type
}

func (n *TypeAliasNode) DeclNode() {}

func (n *TypeAliasNode) String() string {
	return fmt.Sprintf("type %s = %s", n.Name, n.Type.String())
}

// ConstantValue represents a constant value (integer or string)
type ConstantValue interface {
	Node
	ConstantValueNode()
}

// IntConstant represents an integer constant value
type IntConstant struct {
	BaseNode
	Value int64
}

func (n *IntConstant) ConstantValueNode() {}

func (n *IntConstant) String() string {
	return fmt.Sprintf("%d", n.Value)
}

// StringConstant represents a string constant value
type StringConstant struct {
	BaseNode
	Value string
}

func (n *StringConstant) ConstantValueNode() {}

func (n *StringConstant) String() string {
	return fmt.Sprintf("\"%s\"", n.Value)
}

// ConstantNode represents a constant declaration
type ConstantNode struct {
	BaseNode
	Name  string
	Value ConstantValue
}

func (n *ConstantNode) DeclNode() {}

func (n *ConstantNode) String() string {
	return fmt.Sprintf("const %s = %s", n.Name, n.Value.String())
}