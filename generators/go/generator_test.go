package golang

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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"type User struct {",
		"Id int64 `json:\"id\"`",
		"Name string `json:\"name\"`",
		"Active bool `json:\"active\"`",
		"}",
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"type User struct {",
		"Id int64 `json:\"id\"`",
		"Email *string `json:\"email,omitempty\"`",
		"}",
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
		metadata: [string]int64
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"type User struct {",
		"Tags []string `json:\"tags\"`",
		"Metadata map[string]int64 `json:\"metadata\"`",
		"}",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateTimeTypes(t *testing.T) {
	input := `struct Event {
		created_at: time
		date_only: date
		full_datetime: datetime
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"import \"time\"",
		"type Event struct {",
		"CreatedAt time.Time `json:\"created_at\"`",
		"DateOnly time.Time `json:\"date_only\"`",
		"FullDatetime time.Time `json:\"full_datetime\"`",
		"}",
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"import (",
		"\"encoding/json\"",
		"\"fmt\"",
		")",
		"type Status int",
		"const (",
		"Status_Active Status = iota",
		"Status_Inactive",
		"Status_Pending",
		")",
		"func (e Status) String() string {",
		"switch e {",
		"case Status_Active:",
		"return \"active\"",
		"case Status_Inactive:",
		"return \"inactive\"",
		"case Status_Pending:",
		"return \"pending\"",
		"func (e Status) MarshalJSON() ([]byte, error) {",
		"return json.Marshal(map[string]string{\"type\": e.String()})",
		"func (e *Status) UnmarshalJSON(data []byte) error {",
		"case \"active\":",
		"*e = Status_Active",
		"case \"inactive\":",
		"*e = Status_Inactive",
		"case \"pending\":",
		"*e = Status_Pending",
		"default:",
		"return \"unknown\"",
		"}",
		"}",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateSimpleEnumJSONSerialization(t *testing.T) {
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	// Verify JSON marshaling methods are present
	jsonMethods := []string{
		"func (e Status) MarshalJSON() ([]byte, error) {",
		"return json.Marshal(map[string]string{\"type\": e.String()})",
		"func (e *Status) UnmarshalJSON(data []byte) error {",
		"var obj map[string]string",
		"missing 'type' field",
		"case \"active\":",
		"*e = Status_Active",
		"case \"inactive\":",
		"*e = Status_Inactive",
		"case \"pending\":",
		"*e = Status_Pending",
		"unknown enum value:",
	}

	for _, method := range jsonMethods {
		if !strings.Contains(result, method) {
			t.Errorf("Expected result to contain %q, but got:\n%s", method, result)
		}
	}
}

func TestGenerateTaggedUnionEnum(t *testing.T) {
	input := `enum Result {
		success: string
		error: int64
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"import (",
		"\"encoding/json\"",
		"\"fmt\"",
		"type Result struct {",
		"Payload ResultPayload `json:\"-\"`",
		"}",
		"type ResultPayload interface {",
		"resultType() string",
		"}",
		"type Result_Success string",
		"func (Result_Success) resultType() string {",
		"return \"success\"",
		"}",
		"type Result_Error int64",
		"func (Result_Error) resultType() string {",
		"return \"error\"",
		"}",
		"type Result_Pending struct{}",
		"func (Result_Pending) resultType() string {",
		"return \"pending\"",
		"}",
		"func (e Result) MarshalJSON() ([]byte, error) {",
		"func (e *Result) UnmarshalJSON(data []byte) error {",
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"type UserID = int64",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGeneratePrimitiveTypes(t *testing.T) {
	input := `struct AllTypes {
		int8_field: int8
		int16_field: int16
		int32_field: int32
		int64_field: int64
		nat8_field: nat8
		nat16_field: nat16
		nat32_field: nat32
		nat64_field: nat64
		float32_field: float32
		float64_field: float64
		bool_field: bool
		string_field: string
		json_field: json
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"Int8Field int8 `json:\"int8_field\"`",
		"Int16Field int16 `json:\"int16_field\"`",
		"Int32Field int32 `json:\"int32_field\"`",
		"Int64Field int64 `json:\"int64_field\"`",
		"Nat8Field uint8 `json:\"nat8_field\"`",
		"Nat16Field uint16 `json:\"nat16_field\"`",
		"Nat32Field uint32 `json:\"nat32_field\"`",
		"Nat64Field uint64 `json:\"nat64_field\"`",
		"Float32Field float32 `json:\"float32_field\"`",
		"Float64Field float64 `json:\"float64_field\"`",
		"BoolField bool `json:\"bool_field\"`",
		"StringField string `json:\"string_field\"`",
		"JsonField interface{} `json:\"json_field\"`",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateEmptyStruct(t *testing.T) {
	input := `struct Empty {}`

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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"type Empty struct {",
		"}",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateMultipleDeclarations(t *testing.T) {
	input := `struct User {
		id: UserID
		name: string
	}

	type UserID = int64

	enum Status {
		active
		inactive
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	// Check that all types are present
	expected := []string{
		"package test",
		"type UserID = int64",
		"type Status int",
		"type User struct {",
		"Id UserID `json:\"id\"`",
		"Name string `json:\"name\"`",
		"}",
		"Status_Active Status = iota",
		"Status_Inactive",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestGenerateTaggedUnionJSONMethods(t *testing.T) {
	input := `enum Color {
		red
		green: string
		rgba: RGBA
	}

	struct RGBA {
		r: nat8
		g: nat8
		b: nat8
		a: nat8
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	// Check that marshaling method contains proper JSON structure
	marshalExpected := []string{
		"func (e Color) MarshalJSON() ([]byte, error) {",
		"switch payload := e.Payload.(type) {",
		"case Color_Red:",
		"\"type\": \"red\"",
		"case Color_Green:",
		"\"type\": \"green\"",
		"\"payload\": payload",
		"case Color_Rgba:",
		"\"type\": \"rgba\"",
		"\"payload\": payload",
	}

	// Check that unmarshaling method handles all cases
	unmarshalExpected := []string{
		"func (e *Color) UnmarshalJSON(data []byte) error {",
		"var raw map[string]json.RawMessage",
		"case \"red\":",
		"e.Payload = Color_Red{}",
		"case \"green\":",
		"payloadBytes, exists := raw[\"payload\"]",
		"var payload Color_Green",
		"json.Unmarshal(payloadBytes, &payload)",
		"e.Payload = payload",
		"case \"rgba\":",
	}

	allExpected := append(marshalExpected, unmarshalExpected...)

	for _, exp := range allExpected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}

func TestToPascalCase(t *testing.T) {
	g := NewGenerator()

	tests := []struct {
		input    string
		expected string
	}{
		{"user_id", "UserId"},
		{"first_name", "FirstName"},
		{"api_key", "ApiKey"},
		{"simple", "Simple"},
		{"", ""},
		{"a", "A"},
		{"a_b_c_d", "ABCD"},
	}

	for _, tt := range tests {
		result := g.toPascalCase(tt.input)
		if result != tt.expected {
			t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"const MAX_RETRIES = 5",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		`const API_URL = "https://api.example.com"`,
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
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
	result, exists := fs.GetFileString("test.go")
	if !exists {
		t.Fatal("test.go should have been generated")
	}

	expected := []string{
		"package test",
		"const MAX_SIZE = 1024",
		`const API_KEY = "secret"`,
		"type User struct {",
		"Id int64 `json:\"id\"`",
		"Name string `json:\"name\"`",
		"}",
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Expected result to contain %q, but got:\n%s", exp, result)
		}
	}
}
