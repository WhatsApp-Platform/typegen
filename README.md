# TypeGen

**A schema definition language and multi-target code generator for type-safe APIs**

TypeGen allows you to define your data schemas once in `.tg` files and generate equivalent types for multiple programming languages, ensuring consistency and type safety across your entire stack.

[![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://golang.org/)
[![Python](https://img.shields.io/badge/python-3670A0?style=for-the-badge&logo=python&logoColor=ffdd54)](https://python.org/)

## 🎯 Goals and Features

### Core Goals
- **Single Source of Truth**: Define types once, use everywhere
- **Type Safety**: Generate idiomatic, type-safe code for each target language
- **JSON Interoperability**: Guaranteed lossless JSON encoding/decoding between languages
- **Developer Experience**: Simple syntax with powerful features like imports and constants

### Key Features
- ✅ **Multiple Target Languages**: Go, Python + Pydantic (TypeScript coming soon)
- ✅ **Rich Type System**: Structs, enums, type aliases, constants, and primitive types
- ✅ **Module System**: Organize schemas with imports and nested modules
- ✅ **Build System**: Multi-target generation with YAML configuration
- ✅ **JSON Wire Format**: Standardized JSON encoding for cross-language compatibility
- ✅ **CLI Tools**: Parse, validate, and generate code from the command line

## 📦 Installation

### Via Go Install (Recommended)
```bash
go install github.com/WhatsApp-Platform/typegen/cmd/typegen@latest
```

### For Go Projects (Recommended)
Add TypeGen as a tool dependency in your Go project:

```bash
# Add as a tool dependency (Go 1.24+)
go get -tool github.com/WhatsApp-Platform/typegen/cmd/typegen@latest

# Use in your project
go tool typegen --help
```

This approach ensures all team members use the same TypeGen version and integrates cleanly with your Go module.

### From Source
```bash
git clone https://github.com/WhatsApp-Platform/typegen.git
cd typegen
go build ./cmd/typegen
```

## 🚀 Quick Start

### 1. Create a TypeGen Schema

Create `user.tg`:

```typegen
import auth

const MAX_USERNAME_LENGTH = 50
const API_VERSION = "v1"

struct User {
    id: int64
    name: string
    email: ?string                    // Optional field
    tags: []string                   // Array
    metadata: [string]string         // Map
    auth: auth.Token                 // Cross-module reference
    created_at: datetime
}

enum UserStatus {
    active                           // Simple variant
    pending: string                  // Variant with payload
    suspended: SuspensionInfo        // Variant with struct payload
}

struct SuspensionInfo {
    reason: string
    until: ?date
}

type UserID = int64                  // Type alias
```

### 2. Generate Code

```bash
# Generate Go code
typegen generate -generator go -o ./generated/go ./schemas

# Generate Python + Pydantic code
typegen generate -generator python+pydantic -o ./generated/python ./schemas
```

### 3. Use Build Configuration (Recommended)

Create `typegen.yaml`:

```yaml
version: 1
generate:
  - generator: go
    input: ./schemas
    output: ./backend/generated
    config:
      module-name: github.com/myproject/api
  - generator: python+pydantic
    input: ./schemas
    output: ./frontend/generated
    config:
      module-name: myproject
```

Build all targets:

```bash
typegen build
```

## 🔌 JSON Wire Format

TypeGen generates JSON that's compatible across all target languages. The format follows these rules:

### Structs
```typegen
struct User {
    name: string
    age: int64
}
```

**JSON:**
```json
{
    "name": "Alice",
    "age": 30
}
```

### Enums (Tagged Unions)
```typegen
enum Result {
    success: string
    error: int64
    pending
}
```

**JSON:**
```json
{"type": "success", "payload": "Operation completed"}
{"type": "error", "payload": 404}
{"type": "pending"}
```

### Optional Fields
```typegen
struct Profile {
    name: string
    bio: ?string
}
```

**JSON:**
```json
{"name": "Alice", "bio": "Software Engineer"}
{"name": "Bob"}  // bio omitted (null)
```

### Collections
```typegen
struct Data {
    tags: []string
    counts: [string]int64
}
```

**JSON:**
```json
{
    "tags": ["api", "backend"],
    "counts": {"requests": 1500, "errors": 3}
}
```

## 📖 Command Line Reference

### Core Commands

#### `typegen parse <file>`
Parse and validate a single `.tg` file.

```bash
typegen parse user.tg
```

#### `typegen module <directory>`
Parse and validate all `.tg` files in a directory (non-recursive).

```bash
typegen module ./api-schemas
```

#### `typegen generate`
Generate code for an entire module (recursive).

**Syntax:**
```bash
typegen generate -generator <generator> [options] <input-dir> -o <output-dir>
```

**Options:**
- `-generator <name>`: Target generator (`go`, `python+pydantic`)
- `-o <dir>`: Output directory (required)
- `-c <key=value>`: Configuration override (repeatable)

**Examples:**
```bash
# Generate Go code
typegen generate -generator go -o ./generated/go ./schemas

# Generate Python with custom module name
typegen generate -generator python+pydantic -o ./api \
  -c module-name=myapp.api ./schemas

# Generate with multiple config overrides
typegen generate -generator go -o ./backend \
  -c module-name=github.com/myapp/backend \
  -c package-name=api ./schemas
```

#### `typegen build`
Build all targets defined in a YAML configuration file.

**Syntax:**
```bash
typegen build [-f <config-file>]
```

**Options:**
- `-f <file>`: Configuration file (default: `./typegen.yaml`)

**Examples:**
```bash
# Use default typegen.yaml
typegen build

# Use custom config file
typegen build -f production.yaml
```

### Available Generators

| Generator | Description |
|-----------|-------------|
| `go` | Go structs with JSON marshaling/unmarshaling |
| `python+pydantic` | Python classes with Pydantic validation |

## 💼 Code Examples

### Generated Go Code

**TypeGen Input:**
```typegen
struct User {
    id: int64
    name: string
    status: UserStatus
}

enum UserStatus {
    active
    pending: string
}
```

**Generated Go:**
```go
// Code generated by TypeGen. DO NOT EDIT.

package test_schema

import (
	"encoding/json"
	"fmt"
)

type UserStatus struct {
	Payload UserStatusPayload `json:"-"`
}

type UserStatusPayload interface {
	userstatusType() string
}

type UserStatus_Active struct{}
func (UserStatus_Active) userstatusType() string {
	return "active"
}

type UserStatus_Pending string
func (UserStatus_Pending) userstatusType() string {
	return "pending"
}

func (e UserStatus) MarshalJSON() ([]byte, error) {
	switch payload := e.Payload.(type) {
	case UserStatus_Active:
		return json.Marshal(map[string]interface{}{
			"type": "active",
		})
	case UserStatus_Pending:
		return json.Marshal(map[string]interface{}{
			"type": "pending",
			"payload": payload,
		})
	default:
		return nil, fmt.Errorf("unknown payload type: %T", payload)
	}
}

func (e *UserStatus) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	typeBytes, exists := raw["type"]
	if !exists {
		return fmt.Errorf("missing 'type' field")
	}

	var typeStr string
	if err := json.Unmarshal(typeBytes, &typeStr); err != nil {
		return err
	}

	switch typeStr {
	case "active":
		e.Payload = UserStatus_Active{}
	case "pending":
		payloadBytes, exists := raw["payload"]
		if !exists {
			return fmt.Errorf("missing 'payload' field for type 'pending'")
		}
		var payload UserStatus_Pending
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return err
		}
		e.Payload = payload
	default:
		return fmt.Errorf("unknown type: %s", typeStr)
	}

	return nil
}

type User struct {
	Id int64 `json:"id"`
	Name string `json:"name"`
	Status UserStatus `json:"status"`
}
```

### Generated Python + Pydantic Code

**Generated Python:**
```python
from enum import Enum
from pydantic import BaseModel
from typing import Literal
from typing import Union

# Code generated by TypeGen. DO NOT EDIT.

class UserStatus_Active(BaseModel):
    type: Literal['active'] = 'active'

class UserStatus_Pending(BaseModel):
    type: Literal['pending'] = 'pending'
    payload: str

UserStatus = Union[UserStatus_Active, UserStatus_Pending]

class User(BaseModel):
    id: int
    name: str
    status: UserStatus
```

## 🏗️ TypeGen Language Syntax

### Type System

#### Primitive Types
```typegen
// Integer types
nat8, nat16, nat32, nat64        // Unsigned integers
int8, int16, int32, int64        // Signed integers
bignat, bigint                   // Arbitrary precision

// Floating point
float32, float64                 // IEEE 754 floating point
decimal                          // Arbitrary precision decimal

// Other primitives
string                           // UTF-8 string
bool                             // Boolean
json                             // Raw JSON (validated)

// Date/time types
time, date, datetime             // UTC normalized
timetz, datetz, datetimetz       // With timezone offset
```

#### Collection Types
```typegen
struct Product {
    tags: []string                   // Array
    prices: [string]decimal          // Map
    variants: ?[]ProductVariant      // Optional array
}
```

#### User-Defined Types
```typegen
// Struct definition
struct Address {
    street: string
    city: string
    postal_code: ?string
}

// Enum with variants
enum PaymentMethod {
    cash                             // Simple variant
    credit_card: CreditCardInfo      // Variant with payload
    bank_transfer: BankInfo
}

// Type alias
type UserID = string
type Coordinates = [float64]float64
```

#### Constants
```typegen
const MAX_RETRY_COUNT = 5
const API_BASE_URL = "https://api.example.com"
const TIMEOUT_SECONDS = 30
```

### Module System

#### Directory Structure
```
api/
├── auth/
│   ├── login.tg
│   └── session.tg
├── orders/
│   ├── cart.tg
│   ├── checkout.tg
│   └── payment/
│       └── methods.tg
└── users/
    └── profile.tg
```

#### Imports and References
```typegen
// In api/orders/checkout.tg
import auth
import orders.payment.methods

struct CheckoutRequest {
    user_session: auth.Session
    payment: methods.PaymentMethod
    items: []CheckoutItem
}
```

### Naming Conventions

TypeGen enforces strict naming conventions that are converted to target language conventions during generation:

- **User-Defined Types**: `PascalCase` → Go: `PascalCase`, Python: `PascalCase`
- **Field Names**: `snake_case` → Go: `PascalCase`, Python: `snake_case`
- **Constants**: `CONSTANT_CASE` → Go: `CONSTANT_CASE`, Python: `CONSTANT_CASE`
- **Primitive Types**: `flatcase` (e.g., `int64`, `string`)
- **Module Names**: `snake_case` separated by dots (e.g., `auth.session_management`)

## 🔧 Build Configuration

Create powerful build pipelines with `typegen.yaml`:

```yaml
version: 1

# Global configuration inherited by all tasks
config:
  module-name: github.com/mycompany/api

# Generation targets
generate:
  # Backend Go API
  - generator: go
    input: ./api-schemas
    output: ./backend/pkg/api
    config:
      module-name: github.com/mycompany/backend/pkg/api

  # Frontend Python client
  - generator: python+pydantic
    input: ./api-schemas
    output: ./clients/python/src
    config:
      module-name: mycompany.api.client

  # Mobile app schemas
  - generator: go
    input: ./mobile-schemas
    output: ./mobile/shared/types
```

### Build Features

- **Multi-target Generation**: Build for multiple languages in one command
- **Configuration Inheritance**: Share global config, override per-task
- **Automatic Path Resolution**: Handles relative and absolute paths
- **Comprehensive Error Reporting**: Continue processing all tasks, collect all errors
- **Progress Tracking**: Clear visual indicators (✅/❌) for each task

Run the build:
```bash
typegen build              # Uses ./typegen.yaml
typegen build -f prod.yaml # Uses custom config
```

## 🧪 Development Status

**Current Version**: Beta

✅ **Completed Features:**
- Complete TypeGen parser with goyacc
- Full language support (structs, enums, constants, imports, modules)
- Go code generator with JSON marshaling
- Python + Pydantic code generator
- YAML-based build system
- Recursive module processing
- CLI tools and validation
- Comprehensive test suite

🚧 **Coming Soon:**
- TypeScript/JavaScript generator
- Rust code generator
- Advanced validation rules
- Custom JSON field naming
- Schema versioning and migration tools

---

## ⚠️ Disclaimer

This project was heavily "vibe-coded" during development. While it includes comprehensive tests and works reliably for its intended use cases, the codebase may contain unconventional patterns, experimental approaches, or architectural decisions that prioritized rapid development over traditional software engineering practices.

The library is functional and battle-tested through its test suite, but contributors should expect to encounter creative solutions and may want to refactor certain areas for production-critical applications.

---

## 📄 License

This project is available under an open source license. See the repository for details.

## 🤝 Contributing

Contributions are welcome! Please see the repository for contribution guidelines and development setup instructions.

---

*Built with ❤️ and a lot of vibe-coding*
