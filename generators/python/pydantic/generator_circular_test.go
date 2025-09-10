package pydantic

import (
	"context"
	"strings"
	"testing"

	"github.com/WhatsApp-Platform/typegen/generators"
	"github.com/WhatsApp-Platform/typegen/parser"
	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

func TestGenerateSimpleCircularReference(t *testing.T) {
	// Test case: User -> Profile -> User (direct circular reference)
	input := `
struct User {
	id: int64
	name: string
	profile: Profile
}

struct Profile {
	bio: string
	user: User
}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	// Check for basic structure
	basicExpected := []string{
		"from pydantic import BaseModel",
		"class User(BaseModel):",
		"    id: int",
		"    name: str",
		"class Profile(BaseModel):",
		"    bio: str",
	}

	for _, exp := range basicExpected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}

	// Check that at least one forward reference exists (either direction)
	hasUserForwardRef := strings.Contains(result, "profile: 'Profile'")
	hasProfileForwardRef := strings.Contains(result, "user: 'User'")
	if !hasUserForwardRef && !hasProfileForwardRef {
		t.Errorf("Expected at least one forward reference between User and Profile, but got:\n%s", result)
	}

	// Check for model_rebuild() calls
	if !strings.Contains(result, "User.model_rebuild()") {
		t.Errorf("Expected User.model_rebuild() call, but got:\n%s", result)
	}
	if !strings.Contains(result, "Profile.model_rebuild()") {
		t.Errorf("Expected Profile.model_rebuild() call, but got:\n%s", result)
	}
}

func TestGenerateOptionalCircularReference(t *testing.T) {
	// Test case: Circular reference with optional fields
	input := `
struct Node {
	value: int64
	next: ?Node
	prev: ?Node
}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := []string{
		"from typing import Optional",
		"from pydantic import BaseModel",
		"class Node(BaseModel):",
		"    value: int",
		"    next: Optional['Node']",  // Self-reference with forward ref
		"    prev: Optional['Node']",  // Self-reference with forward ref
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}

	// Check for model_rebuild() call
	if !strings.Contains(result, "Node.model_rebuild()") {
		t.Errorf("Expected Node.model_rebuild() call, but got:\n%s", result)
	}
}

func TestGenerateComplexCircularChain(t *testing.T) {
	// Test case: A -> B -> C -> A (circular chain)
	input := `
struct Company {
	name: string
	departments: []Department
}

struct Department {
	name: string
	employees: []Employee
}

struct Employee {
	name: string
	company: Company
}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	expected := []string{
		"from typing import List",
		"from pydantic import BaseModel",
		"class Company(BaseModel):",
		"    name: str",
		"    departments: List['Department']",  // Forward reference
		"class Department(BaseModel):",
		"    name: str",
		"    employees: List['Employee']",  // Forward reference
		"class Employee(BaseModel):",
		"    name: str",
		"    company: Company",  // Company is already defined
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}

	// Check for model_rebuild() calls for all types in the cycle
	rebuilds := []string{
		"Company.model_rebuild()",
		"Department.model_rebuild()",
		"Employee.model_rebuild()",
	}
	for _, rebuild := range rebuilds {
		if !strings.Contains(result, rebuild) {
			t.Errorf("Expected %s call, but got:\n%s", rebuild, result)
		}
	}
}

func TestGenerateMultipleCycles(t *testing.T) {
	// Test case: Multiple independent cycles in the same file
	// Cycle 1: User <-> Profile
	// Cycle 2: Post <-> Comment
	input := `
struct User {
	id: int64
	profile: Profile
	posts: []Post
}

struct Profile {
	bio: string
	user: User
}

struct Post {
	title: string
	author: User
	comments: []Comment
}

struct Comment {
	text: string
	post: Post
}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	// Check that forward references are used appropriately (flexible about ordering)
	// At least some forward references should exist
	forwardRefs := []string{
		"profile: 'Profile'",
		"user: 'User'", 
		"posts: List['Post']",
		"author: 'User'",
		"comments: List['Comment']",
	}

	foundForwardRefs := 0
	for _, ref := range forwardRefs {
		if strings.Contains(result, ref) {
			foundForwardRefs++
		}
	}

	if foundForwardRefs < 2 {
		t.Errorf("Expected at least 2 forward references among %v, but found %d in:\n%s", forwardRefs, foundForwardRefs, result)
	}

	// Check for model_rebuild() calls for all cyclic types
	rebuilds := []string{
		"User.model_rebuild()",
		"Profile.model_rebuild()",
		"Post.model_rebuild()",
		"Comment.model_rebuild()",
	}
	for _, rebuild := range rebuilds {
		if !strings.Contains(result, rebuild) {
			t.Errorf("Expected %s call, but got:\n%s", rebuild, result)
		}
	}
}

func TestGenerateNonCircularReferences(t *testing.T) {
	// Test case: Ensure non-circular references don't get forward refs
	input := `
struct Address {
	street: string
	city: string
}

struct Person {
	name: string
	address: Address
}

struct Company {
	name: string
	headquarters: Address
	employees: []Person
}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	// Check that NO forward references are used (no quotes around type names)
	unexpectedForwardRefs := []string{
		"address: 'Address'",
		"headquarters: 'Address'",
		"employees: List['Person']",
	}

	for _, unexp := range unexpectedForwardRefs {
		if strings.Contains(result, unexp) {
			t.Errorf("Unexpected forward reference %q found:\n%s", unexp, result)
		}
	}

	// Check that NO model_rebuild() calls are generated
	if strings.Contains(result, "model_rebuild()") {
		t.Errorf("Unexpected model_rebuild() call found for non-circular types:\n%s", result)
	}

	// Check correct non-forward references are used
	expectedRefs := []string{
		"address: Address",
		"headquarters: Address",
		"employees: List[Person]",
	}

	for _, exp := range expectedRefs {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected reference %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateCircularWithEnums(t *testing.T) {
	// Test case: Circular reference involving enums with payloads
	input := `
enum Result {
	success: User
	error: string
}

struct User {
	id: int64
	last_result: ?Result
}`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	// Check for forward reference in the enum variant
	if !strings.Contains(result, "payload: 'User'") {
		t.Errorf("Expected forward reference in enum variant, but got:\n%s", result)
	}

	// Check for model_rebuild() calls
	rebuilds := []string{
		"Result_Success.model_rebuild()",
		"User.model_rebuild()",
	}
	for _, rebuild := range rebuilds {
		if !strings.Contains(result, rebuild) {
			t.Errorf("Expected %s call, but got:\n%s", rebuild, result)
		}
	}

	// The Result union type itself should also be rebuilt
	if !strings.Contains(result, "# Rebuild union type after all variants") {
		t.Log("Note: Union type rebuild comment not found, but this might be implementation-specific")
	}
}

func TestGenerateCircularWithTypeAlias(t *testing.T) {
	// Test case: Circular reference through type aliases
	input := `
type UserID = int64

struct User {
	id: UserID
	friends: []User
	manager: ?Manager
}

type Manager = User`

	program, err := parser.Parse(strings.NewReader(input), "test.tg")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generation error: %v", err)
	}

	result, exists := fs.GetFileString("test.py")
	if !exists {
		t.Fatal("test.py should have been generated")
	}

	// Check for self-reference with forward ref
	if !strings.Contains(result, "friends: List['User']") {
		t.Errorf("Expected forward reference for self-referencing field, but got:\n%s", result)
	}

	// Check for reference to type alias (could be forward or direct depending on ordering)
	hasForwardRef := strings.Contains(result, "manager: Optional['Manager']")
	hasDirectRef := strings.Contains(result, "manager: Optional[Manager]")
	if !hasForwardRef && !hasDirectRef {
		t.Errorf("Expected reference to Manager type alias (either forward or direct), but got:\n%s", result)
	}

	// Check for model_rebuild() call
	if !strings.Contains(result, "User.model_rebuild()") {
		t.Errorf("Expected User.model_rebuild() call, but got:\n%s", result)
	}
}