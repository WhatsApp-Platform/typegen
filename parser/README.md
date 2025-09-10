# TypeGen Parser

The TypeGen parser is a complete implementation of the TypeGen language specification using goyacc (Go's yacc implementation). It provides lexical analysis, parsing, and AST generation for TypeGen schema definition files.

## Architecture

The parser follows a clean, modular architecture:

```
parser/
├── ast/           # Abstract Syntax Tree definitions
├── grammar/       # goyacc-based parser and lexer
└── parser.go      # Public API
```

### AST Package (`ast/`)

Defines the Abstract Syntax Tree nodes that represent parsed TypeGen code:

- **`node.go`**: Base interfaces (`Node`, `Declaration`, `Type`) and common functionality
- **`program.go`**: Root AST node (`ProgramNode`) and import declarations (`ImportNode`)  
- **`declarations.go`**: Type declarations (`StructNode`, `EnumNode`, `TypeAliasNode`, `ConstantNode`, `FieldNode`, `EnumVariantNode`) and constant values (`IntConstant`, `StringConstant`)
- **`types.go`**: Type expressions (`PrimitiveType`, `NamedType`, `ArrayType`, `MapType`, `OptionalType`)

### Grammar Package (`grammar/`)

Contains the goyacc-generated parser and integrated lexer:

- **`grammar.y`**: yacc grammar specification for TypeGen language
- **`parser.go`**: Generated parser (via `go generate`)
- **`lexer.go`**: Custom lexer implementing goyacc's `yyLexer` interface
- **`generate.go`**: Contains `go:generate` directive for parser generation

### Public API (`parser.go`)

Provides simple functions for parsing TypeGen files:

- `ParseFile(filename) (*ast.ProgramNode, error)`: Parse a single `.tg` file
- `Parse(io.Reader, filename) (*ast.ProgramNode, error)`: Parse from any reader
- `ParseModule(directory) (map[string]*ast.ProgramNode, error)`: Parse all `.tg` files in a directory

## Supported Language Features

The parser supports the complete TypeGen language specification:

### Type System
- **Primitive types**: `int8`, `int16`, `int32`, `int64`, `int`, `bigint`, `nat8`, `nat16`, `nat32`, `nat64`, `nat`, `bignat`, `float32`, `float64`, `decimal`, `string`, `bool`, `json`, `time`, `date`, `datetime`, `timetz`, `datetz`, `datetimetz`
- **Array types**: `[]ElementType`
- **Map types**: `[KeyType]ValueType` 
- **Optional types**: `?Type` (in field declarations)
- **Named types**: References to user-defined types
- **Qualified names**: Cross-module references like `Module.Type`

### Declarations
- **Structs**: `struct Name { field: Type, optional_field: ?Type }`
- **Enums**: `enum Name { variant, variant_with_payload: Type }`
- **Type aliases**: `type Alias = ActualType`
- **Constants**: `const CONSTANT_NAME = value` (integer or string literals)

### Modules
- **Imports**: `import Module.Path.Name` with dot-separated module paths
- **Module structure**: Go-like directory-based module system

### Naming Conventions
- **User-Defined Types**: PascalCase (`User`, `ProjectPermissions`)
- **Field Names**: snake_case (`user_id`, `created_at`)
- **Primitive Types**: flatcase (`int64`, `string`)
- **Constants**: CONSTANT_CASE (`MAX_RETRIES`, `API_KEY`)

## Example Usage

### Parsing a Single File

```go
import "github.com/WhatsApp-Platform/typegen/parser"

program, err := parser.ParseFile("schema.tg")
if err != nil {
    log.Fatal(err)
}

// Access the parsed AST
fmt.Printf("Found %d declarations\n", len(program.Declarations))
for _, decl := range program.Declarations {
    switch d := decl.(type) {
    case *ast.StructNode:
        fmt.Printf("Struct: %s\n", d.Name)
    case *ast.EnumNode:
        fmt.Printf("Enum: %s\n", d.Name)
    case *ast.TypeAliasNode:
        fmt.Printf("Type alias: %s\n", d.Name)
    case *ast.ConstantNode:
        fmt.Printf("Constant: %s\n", d.Name)
    }
}
```

### Parsing a Module

```go
programs, err := parser.ParseModule("./schemas")
if err != nil {
    log.Fatal(err)
}

for filename, program := range programs {
    fmt.Printf("File %s: %d declarations\n", filename, len(program.Declarations))
}
```

## Error Handling

The parser provides comprehensive error reporting with precise location information:

```go
program, err := parser.ParseFile("invalid.tg")
if err != nil {
    if parseErr, ok := err.(*parser.ParseError); ok {
        fmt.Printf("Parse failed: %s\n", parseErr.Message)
        for _, errMsg := range parseErr.Errors {
            fmt.Printf("  %s\n", errMsg) // includes file:line:column
        }
    }
}
```

## Code Generation

After parsing, the AST can be used to generate code for different target languages. The AST nodes provide `String()` methods for debugging and simple code generation:

```go
program, _ := parser.ParseFile("user.tg")
fmt.Println(program.String()) // Pretty-printed TypeGen source
```

## Parser Generation

The parser uses goyacc for code generation. To regenerate the parser after modifying the grammar:

```bash
cd parser/grammar
go generate
```

This runs `go tool goyacc -o parser.go grammar.y` to generate the parser from the yacc grammar specification.

## Testing

The parser includes comprehensive tests covering all language features:

```bash
go test ./parser
```

Tests validate:
- All syntax constructs (structs, enums, type aliases, constants)
- Type expressions (arrays, maps, optionals, qualified names)
- Import declarations with module paths
- Constants with integer and string values
- CONSTANT_CASE naming validation
- Error cases and recovery
- Position tracking for error reporting

## Implementation Details

### Lexer Integration
The parser uses a custom lexer integrated with goyacc that:
- Uses Go's `text/scanner` for tokenization
- Automatically skips comments and whitespace
- Provides precise position tracking for errors
- Maps keywords to goyacc-generated token constants

### Grammar Design  
The yacc grammar is designed for:
- **Unambiguity**: No shift/reduce or reduce/reduce conflicts
- **Extensibility**: Easy to add new syntax constructs
- **Error recovery**: Graceful handling of syntax errors
- **Left-recursion**: Efficient parsing of lists and sequences

### AST Design
The AST follows these principles:
- **Immutable**: AST nodes don't change after creation
- **Typed**: Strong Go type system prevents invalid trees
- **Printable**: All nodes implement `String()` for debugging
- **Visitable**: Interface-based design supports visitor patterns