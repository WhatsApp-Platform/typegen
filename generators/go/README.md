# Go Code Generator

The Go generator produces idiomatic Go code from TypeGen schema definitions, with full JSON marshaling/unmarshaling support compatible with the TypeGen specification.

## Features

- **Struct Generation**: TypeGen structs → Go structs with JSON tags
- **Enum Support**: Simple enums → constants, complex enums → tagged unions with custom JSON methods
- **Type Aliases**: Direct mapping to Go type aliases (`type UserID = int64`)
- **Optional Fields**: Mapped to Go pointers (`?string` → `*string`)
- **Collections**: Arrays (`[]T`) and maps (`map[K]V`) with full type safety
- **Time Types**: All TypeGen time types → `time.Time` with automatic imports
- **JSON Compatibility**: Generated code works seamlessly with Go's `encoding/json` package

## Type Mappings

### Primitive Types
| TypeGen | Go | Notes |
|---------|----|----|
| `bool` | `bool` | |
| `string` | `string` | |
| `int8`-`int64` | `int8`-`int64` | |
| `nat8`-`nat64` | `uint8`-`uint64` | |
| `float32`, `float64` | `float32`, `float64` | |
| `json` | `interface{}` | |
| `time`, `date`, `datetime` | `time.Time` | Auto-imports `time` package |
| `timetz`, `datetz`, `datetimetz` | `time.Time` | Auto-imports `time` package |

### Complex Types
| TypeGen | Go | Example |
|---------|----|----|
| `[]T` | `[]T` | `[]string` |
| `[K]V` | `map[K]V` | `map[string]int64` |
| `?T` | `*T` | `*string` for optional fields |

### Naming Conventions
- **Fields**: `snake_case` → `PascalCase` with JSON tags (`user_name` → `UserName` with `json:"user_name"`)
- **Types**: Already `PascalCase` in TypeGen, preserved in Go
- **Packages**: Module names converted to lowercase

## Generated Code Examples

### Structs
```typegen
struct User {
    id: int64
    name: string
    email: ?string
}
```

Generates:
```go
type User struct {
    Id    int64   `json:"id"`
    Name  string  `json:"name"`
    Email *string `json:"email"`
}
```

### Simple Enums
```typegen
enum Status {
    active
    inactive
    pending
}
```

Generates:
```go
type Status int

const (
    Status_Active Status = iota
    Status_Inactive
    Status_Pending
)

func (e Status) String() string {
    switch e {
    case Status_Active:
        return "active"
    case Status_Inactive:
        return "inactive"
    case Status_Pending:
        return "pending"
    default:
        return "unknown"
    }
}
```

### Tagged Unions (Complex Enums)
```typegen
enum Result {
    success: string
    error: int64
    pending
}
```

Generates:
```go
type Result struct {
    Payload ResultPayload `json:"-"`
}

type ResultPayload interface {
    resultType() string
}

type Result_Success string
func (Result_Success) resultType() string { return "success" }

type Result_Error int64
func (Result_Error) resultType() string { return "error" }

type Result_Pending struct{}
func (Result_Pending) resultType() string { return "pending" }

// Custom JSON marshaling/unmarshaling methods
func (e Result) MarshalJSON() ([]byte, error) {
    switch payload := e.Payload.(type) {
    case Result_Success:
        return json.Marshal(map[string]interface{}{
            "type": "success",
            "payload": payload,
        })
    case Result_Error:
        return json.Marshal(map[string]interface{}{
            "type": "error", 
            "payload": payload,
        })
    case ResultPending:
        return json.Marshal(map[string]interface{}{
            "type": "pending",
        })
    default:
        return nil, fmt.Errorf("unknown payload type: %T", payload)
    }
}

func (e *Result) UnmarshalJSON(data []byte) error {
    // Implementation handles TypeGen JSON format
    // { "type": "success", "payload": "ok" }
    // { "type": "pending" }
}
```

### Type Aliases
```typegen
type UserID = int64
```

Generates:
```go
type UserID = int64
```

## Usage Examples

### Creating and Using Tagged Unions
```go
// Create variants
success := Result{Payload: Result_Success("Operation completed")}
error := Result{Payload: Result_Error(404)}
pending := Result{Payload: ResultPending{}}

// JSON marshaling (automatic)
data, _ := json.Marshal(success)
// Output: {"type": "success", "payload": "Operation completed"}

// Type checking and access
switch payload := result.Payload.(type) {
case Result_Success:
    fmt.Println("Success:", string(payload))
case Result_Error:
    fmt.Println("Error code:", int64(payload))
case ResultPending:
    fmt.Println("Still pending...")
}
```

### Working with Optional Fields
```go
user := User{
    Id:   123,
    Name: "John Doe",
    Email: &[]string{"john@example.com"}[0], // Helper for string literals
}

// Or using a helper function
func StringPtr(s string) *string { return &s }
user.Email = StringPtr("john@example.com")

// Checking optional fields
if user.Email != nil {
    fmt.Println("Email:", *user.Email)
}
```

## JSON Compatibility

The generated Go code produces JSON that exactly matches the TypeGen specification:

- **Structs**: Standard JSON objects with snake_case field names
- **Simple Enums**: Serialized as strings (`"active"`, `"pending"`)
- **Tagged Unions**: Objects with `type` and optional `payload` fields:
  ```json
  {"type": "success", "payload": "Operation completed"}
  {"type": "pending"}
  ```

## Module Support

The generator supports TypeGen's recursive module system:

```
api/
├── user.tg          → user.go (package api)
├── auth/
│   ├── login.tg     → auth/login.go (package auth)
│   └── token.tg     → auth/token.go (package auth)
└── orders/
    └── order.tg     → orders/order.go (package orders)
```

Each directory becomes a Go package with the directory name (converted to lowercase).

## Import Configuration

When your TypeGen schemas use imports, the Go generator requires a `module-name` configuration to generate proper Go import statements.

### Configuration
```bash
# Required when using imports
typegen generate -generator go -c module-name=github.com/user/project -o ./output ./schemas
```

### Import Conversion

TypeGen imports are converted to Go imports using the configured module name:

| TypeGen Import | Config | Generated Go Import |
|----------------|--------|-------------------|
| `import auth` | `module-name=github.com/user/project` | `"github.com/user/project/auth"` |
| `import some.other.module.auth` | `module-name=github.com/user/project` | `"github.com/user/project/some/other/module/auth"` |

### Example

**TypeGen Schema** (`user.tg`):
```typegen
import auth
import services.payment

struct User {
    id: int64
    token: auth.Token
    payment: services.PaymentInfo
}
```

**Generated Go Code**:
```go
package main

import (
    "github.com/user/project/auth"
    "github.com/user/project/services/payment"
)

type User struct {
    Id      int64                `json:"id"`
    Token   auth.Token           `json:"token"`
    Payment services.PaymentInfo `json:"payment"`
}
```

### Error Handling

If imports are used without providing `module-name`, the generator will error:
```bash
$ typegen generate -generator go -o ./output ./schemas
Generation error: module-name configuration is required when using imports (import: auth)
```

This ensures your generated Go code will compile correctly with proper import paths.

## CLI Usage

```bash
# Generate Go code for entire module
typegen generate -generator go -o ./generated/go ./schemas

# With module name for import support
typegen generate -generator go -c module-name=github.com/user/project -o ./generated/go ./schemas

# Examples
typegen generate -generator go -c module-name=github.com/company/api -o ./internal/types ./api
typegen generate -generator go -o ./out ./examples  # No imports needed
```

## Dependencies

The generated Go code only depends on Go's standard library:
- `encoding/json` - For JSON marshaling/unmarshaling
- `fmt` - For error messages
- `time` - For time-related types (when used)

No external dependencies are required.

## Error Handling

The generator provides comprehensive error handling:
- **Parse Errors**: Clear error messages with file locations
- **Type Errors**: Validation of type references and circular dependencies
- **JSON Errors**: Runtime errors for malformed JSON during unmarshaling

## Testing

The generator includes comprehensive unit tests using the InMemoryFS framework:

```bash
# Run Go generator tests
go test ./generators/go

# Run all generator tests  
go test ./generators/...
```

Tests cover:
- All TypeGen language features
- Edge cases (empty structs, complex nested types)
- JSON marshaling/unmarshaling correctness
- Error conditions and recovery
- Module generation and recursive structures

## Architecture

The Go generator follows the pluggable architecture defined in `generators/generator.go`:

- **Interface Compliance**: Implements `generators.Generator`
- **Modular Design**: Separate functions for structs, enums, type aliases
- **Dependency Resolution**: Topological sorting to avoid forward references  
- **Import Management**: Automatic tracking and generation of import statements
- **Testing Framework**: Uses `generators.InMemoryFS` for fast, isolated tests