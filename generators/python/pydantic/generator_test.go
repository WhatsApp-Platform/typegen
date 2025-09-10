package pydantic

import (
	"context"
	"strings"
	"testing"

	"github.com/WhatsApp-Platform/typegen/generators"
	"github.com/WhatsApp-Platform/typegen/parser"
	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

func TestGenerateStruct(t *testing.T) {
	input := `struct User {
		id: int64
		name: string
		active: bool
	}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create a simple module for testing single-file generation
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	// Get the generated file content
	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := []string{
		"from pydantic import BaseModel",
		"class User(BaseModel):",
		"    id: int",
		"    name: str",
		"    active: bool",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateOptionalFields(t *testing.T) {
	input := `struct User {
		id: int64
		email: ?string
	}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create a simple module for testing single-file generation
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	// Get the generated file content
	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := []string{
		"from typing import Optional",
		"from pydantic import BaseModel",
		"class User(BaseModel):",
		"    id: int",
		"    email: Optional[str]",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateArrayAndMap(t *testing.T) {
	input := `struct User {
		tags: []string
		metadata: [string]int32
	}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create a simple module for testing single-file generation
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	// Get the generated file content
	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := []string{
		"from typing import List",
		"from typing import Dict",
		"from pydantic import BaseModel",
		"class User(BaseModel):",
		"    tags: List[str]",
		"    metadata: Dict[str, int]",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateSimpleEnum(t *testing.T) {
	input := `enum Status {
		active
		inactive
		pending
	}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create a simple module for testing single-file generation
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	// Get the generated file content
	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := []string{
		"from enum import Enum",
		"from typing import Any",
		"from pydantic_core import CoreSchema, core_schema",
		"from pydantic import GetCoreSchemaHandler",
		"class Status(Enum):",
		"    ACTIVE = \"active\"",
		"    INACTIVE = \"inactive\"",
		"    PENDING = \"pending\"",
		"def __get_pydantic_core_schema__(cls, _source_type: Any, _handler: GetCoreSchemaHandler) -> CoreSchema:",
		"def _validate_from_json(cls, v: Any) -> 'Status':",
		"if type_str == \"active\":",
		"return cls.ACTIVE",
		"if type_str == \"inactive\":",
		"return cls.INACTIVE", 
		"if type_str == \"pending\":",
		"return cls.PENDING",
		"def _serialize_to_json(self) -> dict:",
		"return {\"type\": self.value}",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateEnumWithPayloads(t *testing.T) {
	input := `enum Result {
		success
		error: string
	}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create a simple module for testing single-file generation
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	// Get the generated file content
	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := []string{
		"from typing import Union",
		"from typing import Literal",
		"from pydantic import BaseModel",
		"class Result_Success(BaseModel):",
		"    type: Literal['success'] = 'success'",
		"class Result_Error(BaseModel):",
		"    type: Literal['error'] = 'error'",
		"    payload: str",
		"Result = Union[Result_Success, Result_Error]",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateTypeAlias(t *testing.T) {
	input := `type UserID = int64`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create a simple module for testing single-file generation
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	// Get the generated file content
	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := "UserID = int"
	if !strings.Contains(result, expected) {
		t.Errorf("Expected result to contain %q, but got:\n%s", expected, result)
	}
}

func TestGenerateComplexExample(t *testing.T) {
	input := `struct User {
		id: int64
		email: ?string
		tags: []string
		metadata: [string]string
	}

	enum Status {
		active
		pending: string
	}

	type Project = string`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create a simple module for testing single-file generation
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	// Get the generated file content
	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	// Just verify it doesn't crash and contains key elements
	expected := []string{
		"class User(BaseModel):",
		"Status_Active(BaseModel):",
		"Status_Pending(BaseModel):",
		"Status = Union[",
		"Project = str",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateIntConstant(t *testing.T) {
	input := `const MAX_RETRIES = 5`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create a simple module for testing single-file generation
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	// Get the generated file content
	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := []string{
		"from typing import Final",
		"MAX_RETRIES: Final[int] = 5",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}

	// Check that __init__.py includes the constant
	initContent, exists := fs.GetFileString("__init__.py")
	if !exists {
		t.Fatal("__init__.py should have been generated")
	}

	if !strings.Contains(initContent, "MAX_RETRIES") {
		t.Errorf("Expected __init__.py to include MAX_RETRIES export, but got:\n%s", initContent)
	}
}

func TestGenerateStringConstant(t *testing.T) {
	input := `const API_URL = "https://api.example.com"`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create a simple module for testing single-file generation
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	// Get the generated file content
	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := []string{
		"from typing import Final",
		`API_URL: Final[str] = "https://api.example.com"`,
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}

	// Check that __init__.py includes the constant
	initContent, exists := fs.GetFileString("__init__.py")
	if !exists {
		t.Fatal("__init__.py should have been generated")
	}

	if !strings.Contains(initContent, "API_URL") {
		t.Errorf("Expected __init__.py to include API_URL export, but got:\n%s", initContent)
	}
}

func TestGenerateConstantsWithOtherDeclarations(t *testing.T) {
	input := `const MAX_SIZE = 1024
const API_KEY = "secret"

struct User {
	id: int64
	name: string
}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Create a simple module for testing single-file generation
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	// Get the generated file content
	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := []string{
		"from typing import Final",
		"from pydantic import BaseModel",
		"MAX_SIZE: Final[int] = 1024",
		`API_KEY: Final[str] = "secret"`,
		"class User(BaseModel):",
		"    id: int",
		"    name: str",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}

	// Check that __init__.py includes constants and the struct
	initContent, exists := fs.GetFileString("__init__.py")
	if !exists {
		t.Fatal("__init__.py should have been generated")
	}

	expectedInInit := []string{"MAX_SIZE", "API_KEY", "User"}
	for _, exp := range expectedInInit {
		if !strings.Contains(initContent, exp) {
			t.Errorf("Expected __init__.py to include %s export, but got:\n%s", exp, initContent)
		}
	}
}
