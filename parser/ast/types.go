package ast

import "fmt"

// PrimitiveType represents a primitive type
type PrimitiveType struct {
	BaseNode
	Name string
}

func (n *PrimitiveType) TypeNode() {}

func (n *PrimitiveType) String() string {
	return n.Name
}

// NamedType represents a reference to a user-defined type
type NamedType struct {
	BaseNode
	Name string
}

func (n *NamedType) TypeNode() {}

func (n *NamedType) String() string {
	return n.Name
}

// ArrayType represents an array/slice type
type ArrayType struct {
	BaseNode
	ElementType Type
}

func (n *ArrayType) TypeNode() {}

func (n *ArrayType) String() string {
	return fmt.Sprintf("[]%s", n.ElementType.String())
}

// MapType represents a mapping type [KeyType]ValueType
type MapType struct {
	BaseNode
	KeyType   Type
	ValueType Type
}

func (n *MapType) TypeNode() {}

func (n *MapType) String() string {
	return fmt.Sprintf("[%s]%s", n.KeyType.String(), n.ValueType.String())
}

// OptionalType represents an optional type ?Type
type OptionalType struct {
	BaseNode
	ElementType Type
}

func (n *OptionalType) TypeNode() {}

func (n *OptionalType) String() string {
	return fmt.Sprintf("?%s", n.ElementType.String())
}