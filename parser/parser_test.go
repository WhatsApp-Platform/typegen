package parser

import (
	"strings"
	"testing"
	
	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

func TestParseSimpleStruct(t *testing.T) {
	input := `
struct User {
  id: int64
  email: string
  age: ?int32
}
`
	
	program, err := Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if program == nil {
		t.Fatal("Program is nil")
	}
	
	if len(program.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(program.Declarations))
	}
	
	structDecl, ok := program.Declarations[0].(*ast.StructNode)
	if !ok {
		t.Fatalf("Expected StructNode, got %T", program.Declarations[0])
	}
	
	if structDecl.Name != "User" {
		t.Errorf("Expected struct name 'User', got '%s'", structDecl.Name)
	}
	
	if len(structDecl.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(structDecl.Fields))
	}
	
	// Test first field
	if structDecl.Fields[0].Name != "id" {
		t.Errorf("Expected field name 'id', got '%s'", structDecl.Fields[0].Name)
	}
	if structDecl.Fields[0].Optional {
		t.Error("Expected 'id' field to be required")
	}
	
	// Test optional field
	if structDecl.Fields[2].Name != "age" {
		t.Errorf("Expected field name 'age', got '%s'", structDecl.Fields[2].Name)
	}
	if !structDecl.Fields[2].Optional {
		t.Error("Expected 'age' field to be optional")
	}
}

func TestParseEnum(t *testing.T) {
	input := `
enum Color {
  red
  green
  blue
  custom: string
}
`
	
	program, err := Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if len(program.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(program.Declarations))
	}
	
	enumDecl, ok := program.Declarations[0].(*ast.EnumNode)
	if !ok {
		t.Fatalf("Expected EnumNode, got %T", program.Declarations[0])
	}
	
	if enumDecl.Name != "Color" {
		t.Errorf("Expected enum name 'Color', got '%s'", enumDecl.Name)
	}
	
	if len(enumDecl.Variants) != 4 {
		t.Fatalf("Expected 4 variants, got %d", len(enumDecl.Variants))
	}
	
	// Test simple variant
	if enumDecl.Variants[0].Name != "red" {
		t.Errorf("Expected variant name 'red', got '%s'", enumDecl.Variants[0].Name)
	}
	if enumDecl.Variants[0].Payload != nil {
		t.Error("Expected 'red' variant to have no payload")
	}
	
	// Test variant with payload
	if enumDecl.Variants[3].Name != "custom" {
		t.Errorf("Expected variant name 'custom', got '%s'", enumDecl.Variants[3].Name)
	}
	if enumDecl.Variants[3].Payload == nil {
		t.Error("Expected 'custom' variant to have payload")
	}
}

func TestParseTypeAlias(t *testing.T) {
	input := `type UserID = int64`
	
	program, err := Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if len(program.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(program.Declarations))
	}
	
	typeAlias, ok := program.Declarations[0].(*ast.TypeAliasNode)
	if !ok {
		t.Fatalf("Expected TypeAliasNode, got %T", program.Declarations[0])
	}
	
	if typeAlias.Name != "UserID" {
		t.Errorf("Expected type alias name 'UserID', got '%s'", typeAlias.Name)
	}
}

func TestParseWithImports(t *testing.T) {
	input := `
import auth.types
import common.utils

struct User {
  id: int64
  auth: auth.types.Auth
}
`
	
	program, err := Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if len(program.Imports) != 2 {
		t.Fatalf("Expected 2 imports, got %d", len(program.Imports))
	}
	
	if program.Imports[0].Path != "auth.types" {
		t.Errorf("Expected import path 'auth.types', got '%s'", program.Imports[0].Path)
	}
	
	if program.Imports[1].Path != "common.utils" {
		t.Errorf("Expected import path 'common.utils', got '%s'", program.Imports[1].Path)
	}
}

func TestParseArrayTypes(t *testing.T) {
	input := `
struct Container {
  items: []string
  metadata: [string]int64
}
`
	
	program, err := Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	structDecl := program.Declarations[0].(*ast.StructNode)
	
	// Test array type
	arrayType, ok := structDecl.Fields[0].Type.(*ast.ArrayType)
	if !ok {
		t.Fatalf("Expected ArrayType, got %T", structDecl.Fields[0].Type)
	}
	
	primitiveType, ok := arrayType.ElementType.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("Expected PrimitiveType, got %T", arrayType.ElementType)
	}
	
	if primitiveType.Name != "string" {
		t.Errorf("Expected element type 'string', got '%s'", primitiveType.Name)
	}
	
	// Test map type
	mapType, ok := structDecl.Fields[1].Type.(*ast.MapType)
	if !ok {
		t.Fatalf("Expected MapType, got %T", structDecl.Fields[1].Type)
	}
	
	keyType, ok := mapType.KeyType.(*ast.PrimitiveType)
	if !ok {
		t.Fatalf("Expected PrimitiveType for key, got %T", mapType.KeyType)
	}
	
	if keyType.Name != "string" {
		t.Errorf("Expected key type 'string', got '%s'", keyType.Name)
	}
}

func TestParseWithComments(t *testing.T) {
	input := `
// This is a file-level comment
import some.module

// This is a struct comment
struct User {
  id: int64         // A required field with a type
  email: ?string    // Optional fields
  // Field comment above
  phone: string     // Field comment inline
}

// Comment before enum
enum Status {
  active    // Simple enum variant
  pending   // Another variant  
}

// Type alias comment
type UserID = int64  // Alias comment
`
	
	program, err := Parse(strings.NewReader(input), "test_with_comments.tg")
	if err != nil {
		t.Fatalf("Parse with comments failed: %v", err)
	}
	
	if program == nil {
		t.Fatal("Program is nil")
	}
	
	// Should have 1 import
	if len(program.Imports) != 1 {
		t.Fatalf("Expected 1 import, got %d", len(program.Imports))
	}
	
	if program.Imports[0].Path != "some.module" {
		t.Errorf("Expected import path 'some.module', got '%s'", program.Imports[0].Path)
	}
	
	// Should have 3 declarations: struct, enum, type alias
	if len(program.Declarations) != 3 {
		t.Fatalf("Expected 3 declarations, got %d", len(program.Declarations))
	}
	
	// Test struct declaration
	structDecl, ok := program.Declarations[0].(*ast.StructNode)
	if !ok {
		t.Fatalf("Expected StructNode, got %T", program.Declarations[0])
	}
	
	if structDecl.Name != "User" {
		t.Errorf("Expected struct name 'User', got '%s'", structDecl.Name)
	}
	
	if len(structDecl.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(structDecl.Fields))
	}
	
	// Test enum declaration
	enumDecl, ok := program.Declarations[1].(*ast.EnumNode)
	if !ok {
		t.Fatalf("Expected EnumNode, got %T", program.Declarations[1])
	}
	
	if enumDecl.Name != "Status" {
		t.Errorf("Expected enum name 'Status', got '%s'", enumDecl.Name)
	}
	
	if len(enumDecl.Variants) != 2 {
		t.Fatalf("Expected 2 variants, got %d", len(enumDecl.Variants))
	}
	
	// Test type alias declaration
	typeAlias, ok := program.Declarations[2].(*ast.TypeAliasNode)
	if !ok {
		t.Fatalf("Expected TypeAliasNode, got %T", program.Declarations[2])
	}
	
	if typeAlias.Name != "UserID" {
		t.Errorf("Expected type alias name 'UserID', got '%s'", typeAlias.Name)
	}
}

func TestParseIntConstant(t *testing.T) {
	input := `const MAX_RETRIES = 5`
	
	program, err := Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if len(program.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(program.Declarations))
	}
	
	constDecl, ok := program.Declarations[0].(*ast.ConstantNode)
	if !ok {
		t.Fatalf("Expected ConstantNode, got %T", program.Declarations[0])
	}
	
	if constDecl.Name != "MAX_RETRIES" {
		t.Errorf("Expected constant name 'MAX_RETRIES', got '%s'", constDecl.Name)
	}
	
	intConst, ok := constDecl.Value.(*ast.IntConstant)
	if !ok {
		t.Fatalf("Expected IntConstant, got %T", constDecl.Value)
	}
	
	if intConst.Value != 5 {
		t.Errorf("Expected constant value 5, got %d", intConst.Value)
	}
}

func TestParseStringConstant(t *testing.T) {
	input := `const API_KEY = "secret123"`
	
	program, err := Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if len(program.Declarations) != 1 {
		t.Fatalf("Expected 1 declaration, got %d", len(program.Declarations))
	}
	
	constDecl, ok := program.Declarations[0].(*ast.ConstantNode)
	if !ok {
		t.Fatalf("Expected ConstantNode, got %T", program.Declarations[0])
	}
	
	if constDecl.Name != "API_KEY" {
		t.Errorf("Expected constant name 'API_KEY', got '%s'", constDecl.Name)
	}
	
	strConst, ok := constDecl.Value.(*ast.StringConstant)
	if !ok {
		t.Fatalf("Expected StringConstant, got %T", constDecl.Value)
	}
	
	if strConst.Value != "secret123" {
		t.Errorf("Expected constant value 'secret123', got '%s'", strConst.Value)
	}
}

func TestParseMultipleConstants(t *testing.T) {
	input := `
const MAX_CONNECTIONS = 100
const DEFAULT_TIMEOUT = 30
const DATABASE_URL = "postgres://localhost/mydb"
`
	
	program, err := Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if len(program.Declarations) != 3 {
		t.Fatalf("Expected 3 declarations, got %d", len(program.Declarations))
	}
	
	// Test first constant
	constDecl1, ok := program.Declarations[0].(*ast.ConstantNode)
	if !ok {
		t.Fatalf("Expected ConstantNode at index 0, got %T", program.Declarations[0])
	}
	if constDecl1.Name != "MAX_CONNECTIONS" {
		t.Errorf("Expected constant name 'MAX_CONNECTIONS', got '%s'", constDecl1.Name)
	}
	
	// Test second constant  
	constDecl2, ok := program.Declarations[1].(*ast.ConstantNode)
	if !ok {
		t.Fatalf("Expected ConstantNode at index 1, got %T", program.Declarations[1])
	}
	if constDecl2.Name != "DEFAULT_TIMEOUT" {
		t.Errorf("Expected constant name 'DEFAULT_TIMEOUT', got '%s'", constDecl2.Name)
	}
	
	// Test third constant
	constDecl3, ok := program.Declarations[2].(*ast.ConstantNode)
	if !ok {
		t.Fatalf("Expected ConstantNode at index 2, got %T", program.Declarations[2])
	}
	if constDecl3.Name != "DATABASE_URL" {
		t.Errorf("Expected constant name 'DATABASE_URL', got '%s'", constDecl3.Name)
	}
}

func TestConstantsMixedWithOtherDeclarations(t *testing.T) {
	input := `
const MAX_SIZE = 1024

struct User {
  id: int64
  name: string
}

const DEFAULT_USER = "guest"
`
	
	program, err := Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	
	if len(program.Declarations) != 3 {
		t.Fatalf("Expected 3 declarations, got %d", len(program.Declarations))
	}
	
	// Test constant declaration
	constDecl, ok := program.Declarations[0].(*ast.ConstantNode)
	if !ok {
		t.Fatalf("Expected ConstantNode at index 0, got %T", program.Declarations[0])
	}
	if constDecl.Name != "MAX_SIZE" {
		t.Errorf("Expected constant name 'MAX_SIZE', got '%s'", constDecl.Name)
	}
	
	// Test struct declaration
	structDecl, ok := program.Declarations[1].(*ast.StructNode)
	if !ok {
		t.Fatalf("Expected StructNode at index 1, got %T", program.Declarations[1])
	}
	if structDecl.Name != "User" {
		t.Errorf("Expected struct name 'User', got '%s'", structDecl.Name)
	}
	
	// Test second constant declaration
	constDecl2, ok := program.Declarations[2].(*ast.ConstantNode)
	if !ok {
		t.Fatalf("Expected ConstantNode at index 2, got %T", program.Declarations[2])
	}
	if constDecl2.Name != "DEFAULT_USER" {
		t.Errorf("Expected constant name 'DEFAULT_USER', got '%s'", constDecl2.Name)
	}
}

func TestConstantNameValidation(t *testing.T) {
	// Test invalid constant name (not CONSTANT_CASE)
	invalidInputs := []string{
		`const myConst = 5`,       // camelCase
		`const MY_const = 5`,      // mixed case
		`const max_retries = 5`,   // snake_case
		`const MaxRetries = 5`,    // PascalCase
		`const MY__CONST = 5`,     // double underscore
		`const _MY_CONST = 5`,     // leading underscore
		`const MY_CONST_ = 5`,     // trailing underscore
	}
	
	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			_, err := Parse(strings.NewReader(input), "test.tg")
			if err == nil {
				t.Errorf("Expected parse error for input: %s", input)
			}
		})
	}
	
	// Test valid constant names
	validInputs := []string{
		`const MAX_RETRIES = 5`,
		`const API_KEY = "secret"`,
		`const DATABASE_TIMEOUT = 30`,
		`const A = 1`,
		`const MAX = 100`,
		`const DEFAULT_CONFIG_2 = "config"`,
	}
	
	for _, input := range validInputs {
		t.Run(input, func(t *testing.T) {
			_, err := Parse(strings.NewReader(input), "test.tg")
			if err != nil {
				t.Errorf("Expected no parse error for input: %s, got: %v", input, err)
			}
		})
	}
}