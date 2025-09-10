# TypeGen Build Package

The build package provides build orchestration functionality for TypeGen, allowing users to define all their generation targets in a single YAML configuration file and execute them with one command.

## Overview

The build system allows you to:
- Configure multiple generation tasks in `typegen.yaml`
- Execute all tasks with `typegen build`
- Share global configuration across tasks
- Override global config with per-task settings
- Get comprehensive error reporting and progress tracking

## Configuration File Structure

### Basic Structure

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

### Complete Example

```yaml
version: 1
config:
  # Global configuration shared by all tasks
  module-name: github.com/myproject/schemas
  some-option: global-value
generate:
  # Generate Go code
  - generator: go
    input: ./api
    output: ./backend/generated
    config:
      module-name: github.com/myproject/backend  # Overrides global
      
  # Generate Python code  
  - generator: python+pydantic
    input: ./api
    output: ./frontend/api
    config:
      module-name: frontend.api
      
  # Generate from different input
  - generator: go
    input: ./user-schemas
    output: ./services/user/generated
```

## Configuration Reference

### Root Level Fields

| Field      | Type     | Required | Default | Description |
|------------|----------|----------|---------|-------------|
| `version`  | int      | No       | 1       | Configuration file version |
| `config`   | object   | No       | {}      | Global configuration options |
| `generate` | array    | Yes      | -       | List of generation tasks |

### Generate Task Fields

| Field       | Type     | Required | Default | Description |
|-------------|----------|----------|---------|-------------|
| `generator` | string   | Yes      | -       | Name of the generator to use |
| `input`     | string   | No       | "."     | Input directory containing .tg files |
| `output`    | string   | Yes      | -       | Output directory for generated code |
| `config`    | object   | No       | {}      | Task-specific configuration options |

### Path Resolution

- **Relative paths** are resolved relative to the config file's directory
- **Absolute paths** are used as-is
- The `input` directory must exist and contain .tg files
- The `output` directory will be created if it doesn't exist

### Configuration Merging

Task-specific configurations are merged with global configurations:
1. Start with global config values
2. Override with task-specific config values
3. Task config takes precedence over global config

Example:
```yaml
config:
  module-name: global.module
  timeout: 30
generate:
  - generator: go
    output: ./out
    config:
      module-name: task.module  # Overrides global
      # timeout: 30 inherited from global
```

## CLI Usage

### Basic Commands

```bash
# Build with default typegen.yaml
typegen build

# Build with custom config file
typegen build -f custom-config.yaml

# Show help
typegen build -h
```

### Command Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-f` | Path to configuration file | `./typegen.yaml` |

## API Usage

### Loading Configuration

```go
import "github.com/WhatsApp-Platform/typegen/build"

// Load default config (./typegen.yaml)
config, err := build.LoadConfig("")

// Load custom config file
config, err := build.LoadConfig("custom-config.yaml")
```

### Building Projects

```go
// Create builder
builder := build.NewBuilder(config)

// Validate generators before building
if err := builder.ValidateGenerators(); err != nil {
    log.Fatal(err)
}

// Execute build
ctx := context.Background()
if err := builder.Build(ctx); err != nil {
    log.Fatal(err)
}
```

### Configuration Manipulation

```go
// Get merged config for a specific task
merged := config.MergedConfig(0) // First task

// Access task details
for i, task := range config.Generate {
    fmt.Printf("Task %d: %s -> %s\n", i, task.Input, task.Output)
}
```

## Error Handling

The build system provides comprehensive error reporting:

### Validation Errors
- Missing required fields (`generator`, `output`)
- Invalid configuration version
- Non-existent input directories
- Unknown generators

### Build Errors
- Individual task failures don't stop the entire build
- All errors are collected and reported at the end
- Exit code indicates build success (0) or failure (non-zero)

### Example Error Output

```
Starting build with 3 generation tasks...

[1/3] Generating go code from ./api to ./backend/generated...
✅ Success

[2/3] Generating python code from ./api to ./frontend/api...
❌ Failed: generator "python": module-name is required

[3/3] Generating typescript code from ./api to ./web/types...
✅ Success

Build completed: 2/3 tasks succeeded

Errors encountered:
  - task 2 (python): generator "python": module-name is required
Build failed with 1 errors
```

## Available Generators

The build system works with any registered generator. Current built-in generators:

- **`go`** - Go code generation with JSON marshaling
- **`python+pydantic`** - Python + Pydantic models

Use `typegen build` with an invalid generator to see the current list of available generators.

## File Structure

```
build/
├── README.md          # This documentation
├── config.go          # Configuration loading and validation
├── config_test.go     # Configuration tests
├── builder.go         # Build orchestration
└── builder_test.go    # Builder tests
```

## Testing

Run the build package tests:

```bash
go test github.com/WhatsApp-Platform/typegen/build
```

The test suite covers:
- Configuration loading and validation
- Default value application
- Path resolution
- Error handling
- Configuration merging
- Generator validation