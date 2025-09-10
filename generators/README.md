# TypeGen Code Generators

This package provides the core infrastructure for generating code from TypeGen modules in various target languages.

## Overview

The generators package defines a pluggable architecture that allows TypeGen to generate code for multiple programming languages from a single schema definition. Each generator implements a common interface and can handle complex module structures with recursive subdirectories.

## Architecture

### Core Interfaces

#### Generator Interface

```go
type Generator interface {
    Generate(ctx context.Context, module *ast.Module, dest FS) error
}
```

All code generators must implement this interface:
- `ctx`: Context for cancellation and timeouts
- `module`: The parsed TypeGen module (may contain submodules)
- `dest`: Filesystem abstraction for writing generated files

#### FS Interface

```go
type FS interface {
    WriteFile(name string, data []byte, perm os.FileMode) error
    MkdirAll(path string, perm os.FileMode) error
    Join(elem ...string) string
}
```

Filesystem abstraction that supports:
- Writing files with automatic directory creation
- Creating directory hierarchies
- Platform-agnostic path joining

### Implementations

#### osFS

Production filesystem implementation using the `os` package:

```go
fs := generators.NewOSFS("/output/directory")
```

#### InMemoryFS

Testing filesystem implementation for unit tests:

```go
fs := generators.NewInMemoryFS()
// Write files in memory
err := generator.Generate(ctx, module, fs)
// Verify results
content, exists := fs.GetFileString("output.py")
```

## Generator Registry

The package includes a global registry system for managing generators:

```go
// Register a generator (typically in init())
generators.Register("python+pydantic", func() generators.Generator {
    return pydantic.NewGenerator()
})

// Retrieve a generator
generator, err := generators.Get("python")

// List available generators
languages := generators.List() // ["python+pydantic"]
```

## Module Structure

Generators work with `ast.Module` objects that represent complete TypeGen modules:

```go
type Module struct {
    Path       string                      // Module directory path
    Name       string                      // Module name
    Files      map[string]*ProgramNode     // .tg files in this module
    SubModules map[string]*Module          // Nested submodules
}
```

### Key Features

- **Recursive Structure**: Modules can contain submodules to any depth
- **Cross-Module References**: Generators can resolve types across module boundaries
- **Flexible File Organization**: Each module can contain multiple `.tg` files

## Usage Examples

### Basic Generator Implementation

```go
type MyGenerator struct {
    // generator state
}

func (g *MyGenerator) Generate(ctx context.Context, module *ast.Module, dest generators.FS) error {
    // Generate code for each file in the module
    for filename, program := range module.Files {
        code, err := g.generateProgram(program)
        if err != nil {
            return err
        }
        
        outputFile := strings.TrimSuffix(filename, ".tg") + ".my"
        if err := dest.WriteFile(outputFile, []byte(code), 0644); err != nil {
            return err
        }
    }
    
    // Recursively process submodules
    for subName, subModule := range module.SubModules {
        subPath := dest.Join(subName)
        if err := g.generateSubmodule(ctx, subModule, dest, subPath); err != nil {
            return err
        }
    }
    
    return nil
}
```

### Using the Generator

```go
import (
    "context"
    "github.com/WhatsApp-Platform/typegen/generators"
    "github.com/WhatsApp-Platform/typegen/parser"
)

// Parse a module
module, err := parser.ParseModuleToAST("./schemas")
if err != nil {
    log.Fatal(err)
}

// Get generator
generator, err := generators.Get("python+pydantic")
if err != nil {
    log.Fatal(err)
}

// Generate code
fs := generators.NewOSFS("./output")
ctx := context.Background()
if err := generator.Generate(ctx, module, fs); err != nil {
    log.Fatal(err)
}
```

## Testing

The package provides comprehensive testing utilities:

### InMemoryFS for Testing

```go
func TestMyGenerator(t *testing.T) {
    // Create test module
    program, _ := parser.Parse(strings.NewReader("struct User { id: int64 }"), "user.tg")
    module := ast.NewModule("test", map[string]*ast.ProgramNode{
        "user.tg": program,
    })
    
    // Test generation
    fs := generators.NewInMemoryFS()
    generator := &MyGenerator{}
    err := generator.Generate(context.Background(), module, fs)
    
    // Verify results
    require.NoError(t, err)
    assert.True(t, fs.FileExists("user.my"))
    
    content, exists := fs.GetFileString("user.my")
    require.True(t, exists)
    assert.Contains(t, content, "struct User")
}
```

### Testing Utilities

InMemoryFS provides several helper methods for testing:

- `FileExists(path)` / `DirExists(path)` - Check existence
- `GetFile(path)` / `GetFileString(path)` - Retrieve content  
- `ListFiles()` / `ListDirs()` - List all files/directories
- `Exists(path)` - Check if file or directory exists

## Directory Structure

```
generators/
├── README.md              # This file
├── generator.go           # Core interfaces and osFS implementation
├── generator_test.go      # InMemoryFS tests
├── testing.go             # InMemoryFS implementation for testing
├── registry.go            # Global generator registry
├── python/                # Python code generators
│   └── pydantic/          # Python + Pydantic generator implementation
│       ├── README.md
│       ├── generator.go
│       ├── generator_test.go
│       └── generator_module_test.go
└── go/                    # Go generator implementation
    ├── README.md
    ├── generator.go
    └── generator_test.go
```

## Adding New Generators

To add a new language generator:

1. Create a new subdirectory: `generators/mylang/`
2. Implement the `Generator` interface
3. Register your generator in an `init()` function
4. Add comprehensive tests using `InMemoryFS`
5. Document your generator with a README.md

Example generator registration:

```go
package mylang

import "github.com/WhatsApp-Platform/typegen/generators"

func init() {
    generators.Register("mylang", func() generators.Generator {
        return NewMyLangGenerator()
    })
}
```

## Best Practices

### Error Handling
- Always provide context in error messages
- Include the file/module name when reporting errors
- Use `fmt.Errorf` with `%w` for error wrapping

### Path Handling
- Use `dest.Join()` for path operations (not `filepath.Join` directly)
- Handle both files and subdirectories consistently
- Normalize path separators using the FS interface

### Testing
- Test with both simple and complex module structures
- Verify both file content and directory structure
- Use InMemoryFS for fast, isolated tests
- Test edge cases (empty modules, deep nesting, etc.)

### Performance
- Reset any generator state between modules
- Process large modules efficiently
- Consider memory usage for very large schemas

## Available Generators

- **python+pydantic**: Generates Python code with Pydantic models (see `python/pydantic/README.md`)
- **go**: Generates idiomatic Go code with JSON marshaling support (see `go/README.md`)

More generators coming soon!