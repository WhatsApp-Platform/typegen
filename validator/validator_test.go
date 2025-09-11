package validator

import (
	"strings"
	"testing"

	"github.com/WhatsApp-Platform/typegen/parser"
	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

func TestValidator_ValidModule(t *testing.T) {
	schema := `
struct User {
	id: int64
	name: string
	email: ?string
	tags: []string
	metadata: [string]string
}

enum Status {
	active
	pending: string
}

type UserID = int64
const MAX_USERS = 1000
`

	program, err := parser.Parse(strings.NewReader(schema), "test.tg")
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}
	
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	validator := NewValidator()
	result := validator.Validate(module)

	if result.HasErrors() {
		t.Errorf("Valid module should not have errors, got: %s", result.String())
	}
}

func TestValidator_UndefinedType(t *testing.T) {
	schema := `
struct User {
	id: int64
	profile: NonExistentType
}
`

	program, err := parser.Parse(strings.NewReader(schema), "test.tg")
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}
	
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	validator := NewValidator()
	result := validator.Validate(module)

	if !result.HasErrors() {
		t.Error("Expected validation errors for undefined types")
	}
	
	foundUndefinedError := false
	for _, err := range result.Errors {
		if err.Type == UndefinedTypeError && strings.Contains(err.Message, "NonExistentType") {
			foundUndefinedError = true
			break
		}
	}
	if !foundUndefinedError {
		t.Error("Expected UndefinedTypeError for NonExistentType")
	}
}

func TestValidator_InvalidStructName(t *testing.T) {
	schema := `
struct user_info {
	id: int64
}
`

	program, err := parser.Parse(strings.NewReader(schema), "test.tg")
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}
	
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	validator := NewValidator()
	result := validator.Validate(module)

	if !result.HasErrors() {
		t.Error("Expected naming convention error")
	}
	
	foundNamingError := false
	for _, err := range result.Errors {
		if err.Type == NamingConventionError && strings.Contains(err.Message, "PascalCase") {
			foundNamingError = true
			break
		}
	}
	if !foundNamingError {
		t.Error("Expected naming convention error for struct name")
	}
}

func TestValidator_InvalidFieldName(t *testing.T) {
	schema := `
struct User {
	userID: int64
}
`

	program, err := parser.Parse(strings.NewReader(schema), "test.tg")
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}
	
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	validator := NewValidator()
	result := validator.Validate(module)

	if !result.HasErrors() {
		t.Error("Expected naming convention error")
	}
	
	foundNamingError := false
	for _, err := range result.Errors {
		if err.Type == NamingConventionError && strings.Contains(err.Message, "snake_case") {
			foundNamingError = true
			break
		}
	}
	if !foundNamingError {
		t.Error("Expected naming convention error for field name")
	}
}

func TestValidator_DuplicateFields(t *testing.T) {
	schema := `
struct User {
	id: int64
	name: string
	id: string
}
`

	program, err := parser.Parse(strings.NewReader(schema), "test.tg")
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}
	
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	validator := NewValidator()
	result := validator.Validate(module)

	if !result.HasErrors() {
		t.Error("Expected duplicate field error")
	}
	
	foundDuplicateError := false
	for _, err := range result.Errors {
		if err.Type == DuplicateFieldError {
			foundDuplicateError = true
			break
		}
	}
	if !foundDuplicateError {
		t.Error("Expected duplicate field error")
	}
}

func TestValidator_InvalidPrimitiveType(t *testing.T) {
	schema := `
struct User {
	id: integer
}
`

	program, err := parser.Parse(strings.NewReader(schema), "test.tg")
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}
	
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	validator := NewValidator()
	result := validator.Validate(module)

	if !result.HasErrors() {
		t.Error("Expected validation errors for 'integer'")
	}
	
	// "integer" should either be an invalid primitive type OR an undefined type
	// (since it's not PascalCase, it could be treated as a NamedType that doesn't exist)
	foundExpectedError := false
	for _, err := range result.Errors {
		if (err.Type == InvalidPrimitiveError && strings.Contains(err.Message, "integer")) ||
		   (err.Type == UndefinedTypeError && strings.Contains(err.Message, "integer")) {
			foundExpectedError = true
			break
		}
	}
	if !foundExpectedError {
		t.Errorf("Expected either invalid primitive type or undefined type error for 'integer', got errors: %v", result.Errors)
	}
}

func TestValidator_InvalidMapKey(t *testing.T) {
	schema := `
struct User {
	metadata: [bool]string
}
`

	program, err := parser.Parse(strings.NewReader(schema), "test.tg")
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}
	
	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"test.tg": program,
	})

	validator := NewValidator()
	result := validator.Validate(module)

	if !result.HasErrors() {
		t.Error("Expected invalid map key error")
	}
	
	foundMapKeyError := false
	for _, err := range result.Errors {
		if err.Type == InvalidMapKeyError {
			foundMapKeyError = true
			break
		}
	}
	if !foundMapKeyError {
		t.Error("Expected invalid map key error")
	}
}

func TestNamingConventionRules(t *testing.T) {
	// Test snake_case validation
	if !IsValidSnakeCase("user_name") {
		t.Error("user_name should be valid snake_case")
	}
	if IsValidSnakeCase("UserName") {
		t.Error("UserName should not be valid snake_case")
	}
	
	// Test PascalCase validation
	if !IsValidPascalCase("UserName") {
		t.Error("UserName should be valid PascalCase")
	}
	if IsValidPascalCase("user_name") {
		t.Error("user_name should not be valid PascalCase")
	}
	
	// Test CONSTANT_CASE validation
	if !IsValidConstantCase("MAX_SIZE") {
		t.Error("MAX_SIZE should be valid CONSTANT_CASE")
	}
	if IsValidConstantCase("max_size") {
		t.Error("max_size should not be valid CONSTANT_CASE")
	}
}

func TestPrimitiveTypeValidation(t *testing.T) {
	validTypes := []string{
		"int8", "int16", "int32", "int64",
		"nat8", "nat16", "nat32", "nat64",
		"float32", "float64",
		"string", "bool", "json",
		"datetime", "date", "time",
	}

	for _, typ := range validTypes {
		if !IsValidPrimitiveType(typ) {
			t.Errorf("Type %s should be valid", typ)
		}
	}

	invalidTypes := []string{"integer", "double", "str", "boolean", "timestamp"}
	for _, typ := range invalidTypes {
		if IsValidPrimitiveType(typ) {
			t.Errorf("Type %s should be invalid", typ)
		}
	}
}

func TestMapKeyValidation(t *testing.T) {
	validKeys := []string{"string", "int8", "int16", "int32", "int64", "nat8", "nat16", "nat32", "nat64"}
	for _, key := range validKeys {
		if !IsValidMapKeyType(key) {
			t.Errorf("Key type %s should be valid", key)
		}
	}

	invalidKeys := []string{"float32", "float64", "bool", "json", "datetime"}
	for _, key := range invalidKeys {
		if IsValidMapKeyType(key) {
			t.Errorf("Key type %s should be invalid", key)
		}
	}
}

func TestValidationResult_String(t *testing.T) {
	result := NewValidationResult()
	result.AddError(UndefinedTypeError, "undefined type 'Foo'", "test.tg", 5, 10, "define the type")
	result.AddError(NamingConventionError, "should use PascalCase", "test.tg", 8, 1, "use 'MyType'")

	str := result.String()

	if !strings.Contains(str, "Validation errors found (2)") {
		t.Error("Expected error count in string output")
	}
	if !strings.Contains(str, "test.tg:") {
		t.Error("Expected filename in string output")
	}
	if !strings.Contains(str, "5:10: undefined type 'Foo'") {
		t.Error("Expected error message with position")
	}
	if !strings.Contains(str, "Suggestion: define the type") {
		t.Error("Expected suggestion in output")
	}
}

func TestNamingConventionSuggestions(t *testing.T) {
	if SuggestPascalCase("user_name") != "UserName" {
		t.Error("Expected user_name -> UserName")
	}
	if SuggestSnakeCase("UserName") != "user_name" {
		t.Error("Expected UserName -> user_name")
	}
	if SuggestConstantCase("userName") != "USER_NAME" {
		t.Error("Expected userName -> USER_NAME")
	}
}

func TestValidator_CircularDependencies_Allowed(t *testing.T) {
	schemaA := `
struct NodeA {
	id: string
	b_node: NodeB
}
`

	schemaB := `
struct NodeB {
	id: string
	a_node: NodeA
}
`

	programA, err := parser.Parse(strings.NewReader(schemaA), "a.tg")
	if err != nil {
		t.Fatalf("Failed to parse schema A: %v", err)
	}
	
	programB, err := parser.Parse(strings.NewReader(schemaB), "b.tg")
	if err != nil {
		t.Fatalf("Failed to parse schema B: %v", err)
	}

	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"a.tg": programA,
		"b.tg": programB,
	})

	validator := NewValidator()
	result := validator.Validate(module)

	if result.HasErrors() {
		t.Errorf("Circular dependencies should be allowed, but got errors: %s", result.String())
	}
}

func TestValidator_SelfReference_Allowed(t *testing.T) {
	schema := `
struct TreeNode {
	value: string
	children: []TreeNode
	parent: ?TreeNode
}
`

	program, err := parser.Parse(strings.NewReader(schema), "tree.tg")
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	module := ast.NewModule("test", map[string]*ast.ProgramNode{
		"tree.tg": program,
	})

	validator := NewValidator()
	result := validator.Validate(module)

	if result.HasErrors() {
		t.Errorf("Self-references should be allowed, but got errors: %s", result.String())
	}
}