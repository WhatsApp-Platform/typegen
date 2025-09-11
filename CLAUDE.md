# Claude Code Agent Instructions

This document contains instructions and context for Claude Code agents working on the TypeGen project.

## Project Overview

TypeGen is a schema definition language and code generator that allows defining types in `.tg` files and generating equivalent types for multiple target languages (Go, Python + Pydantic, etc.).

## Project Structure

```
typegen/
├── README.md              # Main project documentation and language spec
├── CLAUDE.md              # This file - agent instructions
├── go.mod                 # Go module definition
├── cmd/
│   └── typegen/
│       └── main.go        # CLI entry point
├── parser/                # Complete goyacc-based parser implementation
│   ├── README.md          # Parser documentation and API guide
│   ├── parser.go          # Public parsing API with recursive module support
│   ├── ast/               # Abstract Syntax Tree definitions
│   │   ├── node.go        # Base interfaces and common functionality
│   │   ├── program.go     # Root AST nodes and imports
│   │   ├── module.go      # Module structure with recursive submodules
│   │   ├── declarations.go # Type declarations (struct, enum, alias)
│   │   └── types.go       # Type expressions and primitives
│   └── grammar/           # goyacc parser and lexer
│       ├── grammar.y      # yacc grammar specification
│       ├── parser.go      # Generated parser (via go generate)
│       ├── lexer.go       # Custom lexer implementation
│       └── generate.go    # go:generate directive
├── generators/            # Code generation framework
│   ├── README.md          # Generator architecture documentation
│   ├── generator.go       # Core interfaces (Generator, FS)
│   ├── generator_test.go  # InMemoryFS tests
│   ├── testing.go         # InMemoryFS implementation for testing
│   ├── registry.go        # Global generator registry
│   ├── python/            # Python code generators
│   │   └── pydantic/      # Python + Pydantic code generator
│   │       ├── README.md      # Python+Pydantic generator documentation
│   │       ├── generator.go   # Python+Pydantic generator implementation
│   │       ├── generator_test.go        # Legacy single-file tests
│   │       └── generator_module_test.go # Recursive module tests
│   └── go/                # Go code generator
│       ├── README.md      # Go generator documentation
│       ├── generator.go   # Go generator implementation
│       └── generator_test.go        # Comprehensive Go generator tests
├── build/                 # Build system for multi-target generation
│   ├── README.md          # Build system documentation
│   ├── config.go          # YAML configuration loading and validation
│   ├── config_test.go     # Configuration tests
│   ├── builder.go         # Build orchestration and execution
│   └── builder_test.go    # Builder tests
├── validator/             # Schema validation system
│   ├── errors.go          # Validation error types and formatting
│   ├── rules.go           # Naming conventions and primitive type validation
│   ├── resolver.go        # Type resolution and circular dependency detection
│   ├── validator.go       # Core validation framework
│   └── validator_test.go  # Comprehensive validation tests
└── examples/              # Example .tg files for testing
    ├── simple.tg          # Basic struct example
    ├── user_clean.tg      # Complex example with imports
    └── user.tg           # Example with comments (currently unsupported)
```

## Key Tools and Technologies

### Go Tools Integration
- Go now has custom tools support: `go get -tool <path>/<tool>`
- Execute with `go tool <tool>`
- For goyacc: `//go:generate go tool goyacc -o parser.go grammar.y`

### Parser Implementation
- Built with **goyacc** (Go's yacc implementation) for robust parsing
- Custom lexer integrated with generated parser using `text/scanner`
- Zero shift/reduce conflicts in grammar
- Comprehensive error reporting with file:line:column positions

### CLI Commands
- `go run ./cmd/typegen parse <file>` - Parse and validate single .tg file
- `go run ./cmd/typegen module <dir>` - Parse all .tg files in directory (non-recursive)
- `go run ./cmd/typegen generate -generator <generator> <module-dir> -o <output-dir>` - Generate code for entire module (recursive)
- `go run ./cmd/typegen build [-f config.yaml]` - Build all targets defined in typegen.yaml
- `go build ./cmd/typegen` - Build standalone CLI binary

**Examples:**
- `go run ./cmd/typegen generate -generator python+pydantic -o ./generated/python ./schemas`
- `go run ./cmd/typegen generate -generator go -o ./generated/go ./api`
- `go run ./cmd/typegen build` - Build all targets from typegen.yaml
- `go run ./cmd/typegen build -f custom-config.yaml` - Build from custom config

### Testing and Development
- `go test ./parser` - Run parser tests
- `go test ./generators` - Run generator framework tests
- `go test ./generators/python/pydantic` - Run Python+Pydantic generator tests
- `go test ./generators/go` - Run Go generator tests
- `go test ./build` - Run build system tests
- `go test ./validator` - Run validation system tests
- `go test ./...` - Run all tests in the project
- `go generate ./parser` - Regenerate parser from grammar
- All tests currently pass

## Current Implementation Status

✅ **Completed:**
- Complete parser implementation with goyacc
- Full TypeGen language support (structs, enums, type aliases, constants, imports)
- AST generation and manipulation with recursive module support
- CLI tool for parsing, validation, and code generation
- **Schema validation system** with comprehensive error checking:
  - Type resolution and undefined type detection
  - Naming convention enforcement (snake_case, PascalCase, CONSTANT_CASE)
  - Duplicate detection (types, fields, variants, constants)
  - Type safety validation (map keys, optional types, primitive types)
  - Circular dependency detection with detailed error reporting
  - Integrated into CLI with `--skip-validation` bypass option
- **Code generation framework** with pluggable architecture
- **Python + Pydantic code generator** with full feature support
- **Go code generator** with JSON marshaling/unmarshaling support
- **Build system** with YAML configuration for multi-target generation
- **Recursive module parsing** and generation
- **InMemoryFS testing framework** for generator testing
- **Global generator registry** for extensibility
- Comprehensive test suite (parser + generators + build + validator)
- Complete documentation (README.md, parser/README.md, generators/README.md, generators/go/README.md, generators/python/pydantic/README.md, build/README.md)

🚧 **Next Steps:**
- Code generation for TypeScript
- Additional target languages (Rust, Java, C#, etc.)
- Advanced serialization options and custom JSON formats
- Enhanced cross-module reference support

✨ **Recently Added:**
- **Constants support**: Integer and string constants with CONSTANT_CASE validation
- **AST nodes for constants**: `ConstantNode`, `IntConstant`, `StringConstant` types
- **Parser grammar updates**: Support for `const NAME = value` declarations
- **Comprehensive constants testing**: Full test coverage including validation
- **Go generator**: Complete Go code generation with idiomatic patterns
- **Tagged union support**: Wrapper struct approach with custom JSON methods
- **Simplified payload types**: Direct type aliases instead of wrapper structs
- **Enhanced testing**: Comprehensive test coverage for Go generator

## TypeGen Language Reference

### Syntax Examples
```typegen
import some.module.path
import auth

const MAX_RETRIES = 5           // Integer constant
const API_URL = "https://api.example.com"  // String constant

struct User {
  id: int64
  email: ?string            // Optional field
  tags: []string            // Array type
  metadata: [string]string  // Map type
  auth: auth.Token          // Qualified type reference
}

enum Status {
  active                   // Simple variant
  pending: string          // Variant with payload
  archived: ArchivedInfo   // Variant with struct payload
}

type UserID = int64         // Type alias
```

### Key Language Features
- **Imports**: Dot-separated module paths (`some.module.auth`)
- **Constants**: Integer and string constants (`const MAX_SIZE = 1024`, `const API_KEY = "secret"`)
- **Optional fields**: `?Type` syntax for nullable fields
- **Array types**: `[]ElementType`
- **Map types**: `[KeyType]ValueType`
- **Qualified names**: Cross-module references (`module.Type`)
- **All primitive types**: int8-64, nat8-64, float32/64, string, bool, json, time/date variants
- **Strict naming conventions**:
  - *snake_case* for module names
  - *snake_case* for field names
  - *smashcase* for builtin types
  - *PascalCase* for user-defined types
  - *CONSTANT_CASE* for constants
  - ALL OTHERS ARE ERRORS!

## Common Tasks for Agents

### Adding New Language Features
1. Update `parser/grammar/grammar.y` with new syntax rules
2. Add corresponding AST nodes in `parser/ast/`
3. Regenerate parser: `cd parser/grammar && go generate`
4. Add tests to `parser/parser_test.go`
5. Update documentation

### Debugging Parser Issues
1. Use debug output: Check `parser/grammar/y.output` for conflicts
2. Test with minimal examples in `examples/`
3. Run `go test ./parser -v` for detailed test output
4. Check lexer tokenization with custom debug tools

### Code Generation Development

**Modern Approach (Recommended):**
1. Parse modules using `parser.ParseModuleToAST(modulePath)` for recursive directory parsing
2. Implement the `generators.Generator` interface:
   ```go
   type Generator interface {
       Generate(ctx context.Context, module *ast.Module, dest FS) error
   }
   ```
3. Register your generator: `generators.Register("mylang", NewMyLangGenerator)`
4. Use `generators.InMemoryFS` for testing
5. Follow target language naming conventions (AST stores original names)

**Legacy Approach (Single Files):**
1. Parse files using `parser.ParseFile()` for single .tg files
2. Use `parser.ParseModule()` for flat directory parsing (non-recursive)
3. Traverse AST using type switches on `ast.Declaration` and `ast.Type`

### Working with AST

**Modern Module-Based Approach:**
```go
// Parse entire module recursively
module, err := parser.ParseModuleToAST("./schemas")

// Access module structure
for filename, program := range module.Files {
    // Process each .tg file in the module
}

// Access submodules recursively
for subModuleName, subModule := range module.SubModules {
    // Process nested submodules
}

// Get all files across all submodules
allFiles := module.AllFiles() // map[string]*ast.ProgramNode with relative paths

// Find declarations across the entire module
if decl, filename, found := module.FindDeclaration("User"); found {
    // Found User declaration in filename
}
```

**Single File Approach:**
```go
// Parse single file
program, err := parser.ParseFile("schema.tg")

// Traverse declarations
for _, decl := range program.Declarations {
    switch d := decl.(type) {
    case *ast.StructNode:
        // Handle struct: d.Name, d.Fields
    case *ast.EnumNode:
        // Handle enum: d.Name, d.Variants
    case *ast.TypeAliasNode:
        // Handle alias: d.Name, d.Type
    case *ast.ConstantNode:
        // Handle constant: d.Name, d.Value (IntConstant or StringConstant)
    }
}
```

## Error Handling Notes

- Parser provides detailed error reporting with `ParseError` type
- Lexical errors include precise position information
- Grammar conflicts are resolved at build time (currently zero conflicts)
- Module parsing continues on individual file failures

## Development Workflow

1. Make changes to grammar or AST
2. Regenerate parser if needed: `go generate ./...`
3. Run tests: `go test ./parser`
4. Test with examples: `go run ./cmd/typegen parse examples/user_clean.tg`

## Important Implementation Details

### Parser Details
- **Lexer skips comments** automatically (scanner.SkipComments)
- **Token constants** are generated by goyacc, not manually defined
- **AST nodes** are immutable after creation
- **Module system** follows Go-like directory structure with recursive parsing
- **Qualified names** support dot notation for cross-module references
- **Error recovery** gracefully handles syntax errors without crashing

### Generator Framework Details
- **FS Interface**: Abstract filesystem for testing and production
- **InMemoryFS**: Fast in-memory filesystem for unit tests
- **Global Registry**: Extensible system for plugging in new generators
- **Recursive Processing**: Handles nested directory structures automatically
- **Path Handling**: Platform-agnostic path operations via FS.Join()

### Module System Features
- **Recursive Structure**: `ast.Module` contains `SubModules map[string]*Module`
- **Directory Filtering**: Automatically skips `.git`, `node_modules`, hidden dirs
- **Cross-Module Search**: `FindDeclaration()` searches across all submodules
- **Relative Paths**: File paths in submodules include directory prefixes

## Generator Development Guidelines

### Creating New Generators
1. **Create package**: `generators/mylang/`
2. **Implement interface**: `generators.Generator`
3. **Register in init()**: `generators.Register("mylang", NewGenerator)`
4. **Add comprehensive tests** using `generators.InMemoryFS`
5. **Document in README.md** with examples and type mappings

### Best Practices
- **Use InMemoryFS for testing**: Fast, isolated, deterministic tests
- **Handle recursive modules**: Process `module.SubModules` recursively
- **Provide detailed errors**: Include file names and context in error messages
- **Follow target conventions**: Convert naming styles appropriately
- **Reset generator state**: Clear imports/state between files
- **Test edge cases**: Empty modules, deep nesting, cross-references

### Testing Patterns
```go
func TestMyGenerator(t *testing.T) {
    // Create test data
    program, _ := parser.Parse(strings.NewReader("struct User { id: int64 }"), "user.tg")
    module := ast.NewModule("test", map[string]*ast.ProgramNode{"user.tg": program})
    
    // Test generation
    fs := generators.NewInMemoryFS()
    generator := NewMyLangGenerator()
    err := generator.Generate(context.Background(), module, fs)
    
    // Verify results
    require.NoError(t, err)
    assert.True(t, fs.FileExists("user.mylang"))
    content, _ := fs.GetFileString("user.mylang")
    assert.Contains(t, content, "struct User")
}
```

## Go Generator Specifics

### Tagged Union Implementation
The Go generator uses a sophisticated wrapper struct approach for complex enums:

**TypeGen:**
```typegen
enum Result {
    success: string
    error: int64
    pending
}
```

**Generated Go:**
```go
type Result struct {
    Payload ResultPayload `json:"-"`
}

type ResultPayload interface {
    resultType() string
}

type ResultSuccess string     // Direct type alias
type ResultError int64        // Direct type alias
type ResultPending struct{}   // Empty struct for simple variants

// Custom JSON marshaling/unmarshaling methods automatically generated
```

### Key Features
- **Direct Type Aliases**: Payload types are direct aliases (`type FooBaz int64`) not wrapper structs
- **JSON Compatibility**: Custom `MarshalJSON`/`UnmarshalJSON` methods for TypeGen JSON format
- **Automatic Imports**: Time package imported when time types are used
- **Topological Sorting**: Dependencies resolved automatically to avoid forward references
- **Field Name Conversion**: `snake_case` → `PascalCase` with proper JSON tags

### Usage Patterns
```go
// Creating tagged union values
result := Result{Payload: ResultSuccess("Operation completed")}
error := Result{Payload: ResultError(404)}

// Type checking
switch payload := result.Payload.(type) {
case ResultSuccess:
    fmt.Println("Success:", string(payload))  // Direct access
case ResultError:
    fmt.Println("Error code:", int64(payload))  // Direct access
}
```

## Build System Details

### Configuration Structure
The build system uses YAML configuration files (typically `typegen.yaml`) to define multiple generation targets:

```yaml
version: 1                    # Configuration version (optional, defaults to 1)
config:                       # Global configuration (optional)
  module-name: example.com/project
generate:                     # List of generation tasks (required)
  - generator: go             # Generator name (required)
    input: api-schemas        # Input directory (optional, defaults to ".")
    output: backend/generated # Output directory (required)
    config:                   # Task-specific config (optional)
      module-name: example.com/backend/api
```

### Build Process
1. **Configuration Loading**: Parse and validate YAML configuration
2. **Generator Validation**: Ensure all specified generators are available
3. **Task Execution**: Execute generation tasks in sequence
4. **Error Collection**: Continue processing all tasks, collect all errors
5. **Progress Reporting**: Display success/failure for each task

### Key Features
- **Multi-target builds**: Generate for multiple languages in one command
- **Config inheritance**: Global config merged with per-task config (task takes precedence)
- **Path resolution**: Relative paths resolved to absolute paths automatically
- **Comprehensive error handling**: Validation errors and generation errors reported separately
- **Progress tracking**: Clear visual indicators (✅/❌) for each task

### CLI Integration
- `typegen build` - Build with default `./typegen.yaml`
- `typegen build -f config.yaml` - Build with custom config file
- Full error reporting and exit codes for CI/CD integration

### Build System Architecture
The build package contains:
- **config.go**: YAML parsing, validation, default application, path resolution
- **builder.go**: Build orchestration, task execution, error collection
- **Comprehensive tests**: Configuration loading, validation, merging, error handling
