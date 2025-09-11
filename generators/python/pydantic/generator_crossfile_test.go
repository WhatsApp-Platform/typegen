package pydantic

import (
	"context"
	"strings"
	"testing"

	"github.com/WhatsApp-Platform/typegen/generators"
	"github.com/WhatsApp-Platform/typegen/parser"
	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

func TestGenerateCrossFileImports_SimpleReference(t *testing.T) {
	// Create user.tg that references Status from status.tg
	userProgram, err := parser.Parse(strings.NewReader(`
		struct User {
			id: int64
			name: string
			status: Status
		}
	`), "user.tg")
	if err != nil {
		t.Fatalf("Failed to parse user.tg: %v", err)
	}

	// Create status.tg that defines Status
	statusProgram, err := parser.Parse(strings.NewReader(`
		enum Status {
			active
			inactive
		}
	`), "status.tg")
	if err != nil {
		t.Fatalf("Failed to parse status.tg: %v", err)
	}

	// Create module
	files := map[string]*ast.ProgramNode{
		"user.tg":   userProgram,
		"status.tg": statusProgram,
	}
	module := ast.NewModule("/test/module", files)

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify user.py contains the cross-file import
	userContent, exists := fs.GetFileString("user.py")
	if !exists {
		t.Fatal("user.py should exist")
	}

	if !strings.Contains(userContent, "from .status import Status") {
		t.Error("user.py should import Status from status module")
	}
	if !strings.Contains(userContent, "status: Status") {
		t.Error("user.py should use Status type")
	}

	// Verify status.py doesn't have unnecessary imports
	statusContent, exists := fs.GetFileString("status.py")
	if !exists {
		t.Fatal("status.py should exist")
	}

	if strings.Contains(statusContent, "from .user import") {
		t.Error("status.py should not import from user")
	}
	if !strings.Contains(statusContent, "class Status(Enum)") {
		t.Error("status.py should define Status enum")
	}
}

func TestGenerateCrossFileImports_MultipleReferences(t *testing.T) {
	// Create user.tg that references types from multiple files
	userProgram, err := parser.Parse(strings.NewReader(`
		struct User {
			id: int64
			profile: Profile
			orders: []Order
		}
	`), "user.tg")
	if err != nil {
		t.Fatalf("Failed to parse user.tg: %v", err)
	}

	// Create profile.tg
	profileProgram, err := parser.Parse(strings.NewReader(`
		struct Profile {
			bio: string
			settings: UserSettings
		}
	`), "profile.tg")
	if err != nil {
		t.Fatalf("Failed to parse profile.tg: %v", err)
	}

	// Create order.tg
	orderProgram, err := parser.Parse(strings.NewReader(`
		struct Order {
			id: int64
			status: string
		}
	`), "order.tg")
	if err != nil {
		t.Fatalf("Failed to parse order.tg: %v", err)
	}

	// Create settings.tg
	settingsProgram, err := parser.Parse(strings.NewReader(`
		struct UserSettings {
			theme: string
			notifications: bool
		}
	`), "settings.tg")
	if err != nil {
		t.Fatalf("Failed to parse settings.tg: %v", err)
	}

	// Create module
	files := map[string]*ast.ProgramNode{
		"user.tg":     userProgram,
		"profile.tg":  profileProgram,
		"order.tg":    orderProgram,
		"settings.tg": settingsProgram,
	}
	module := ast.NewModule("/test/module", files)

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify user.py imports from multiple files, sorted alphabetically
	userContent, exists := fs.GetFileString("user.py")
	if !exists {
		t.Fatal("user.py should exist")
	}

	if !strings.Contains(userContent, "from .order import Order") {
		t.Error("user.py should import Order")
	}
	if !strings.Contains(userContent, "from .profile import Profile") {
		t.Error("user.py should import Profile")
	}
	
	// Check that imports are in alphabetical order
	orderImportPos := strings.Index(userContent, "from .order import Order")
	profileImportPos := strings.Index(userContent, "from .profile import Profile")
	if orderImportPos >= profileImportPos {
		t.Error("imports should be sorted alphabetically")
	}

	// Verify profile.py imports UserSettings
	profileContent, exists := fs.GetFileString("profile.py")
	if !exists {
		t.Fatal("profile.py should exist")
	}

	if !strings.Contains(profileContent, "from .settings import UserSettings") {
		t.Error("profile.py should import UserSettings")
	}
}

func TestGenerateCrossFileImports_EnumWithPayload(t *testing.T) {
	// Create result.tg that defines enum with payload from another file
	resultProgram, err := parser.Parse(strings.NewReader(`
		enum Result {
			success: Data
			error: string
		}
	`), "result.tg")
	if err != nil {
		t.Fatalf("Failed to parse result.tg: %v", err)
	}

	// Create data.tg that defines Data
	dataProgram, err := parser.Parse(strings.NewReader(`
		struct Data {
			value: string
			timestamp: int64
		}
	`), "data.tg")
	if err != nil {
		t.Fatalf("Failed to parse data.tg: %v", err)
	}

	// Create module
	files := map[string]*ast.ProgramNode{
		"result.tg": resultProgram,
		"data.tg":   dataProgram,
	}
	module := ast.NewModule("/test/module", files)

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify result.py imports Data for the enum payload
	resultContent, exists := fs.GetFileString("result.py")
	if !exists {
		t.Fatal("result.py should exist")
	}

	if !strings.Contains(resultContent, "from .data import Data") {
		t.Error("result.py should import Data")
	}
	if !strings.Contains(resultContent, "payload: Data") {
		t.Error("result.py should use Data as payload type")
	}
}

func TestGenerateCrossFileImports_NoSelfImport(t *testing.T) {
	// Create user.tg that defines multiple types referencing each other in the same file
	userProgram, err := parser.Parse(strings.NewReader(`
		struct User {
			id: int64
			profile: UserProfile
		}

		struct UserProfile {
			bio: string
			user_id: int64
		}
	`), "user.tg")
	if err != nil {
		t.Fatalf("Failed to parse user.tg: %v", err)
	}

	// Create module with only one file
	files := map[string]*ast.ProgramNode{
		"user.tg": userProgram,
	}
	module := ast.NewModule("/test/module", files)

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify user.py doesn't try to import from itself
	userContent, exists := fs.GetFileString("user.py")
	if !exists {
		t.Fatal("user.py should exist")
	}

	if strings.Contains(userContent, "from .user import") {
		t.Error("user.py should not import from itself")
	}
	if !strings.Contains(userContent, "profile: UserProfile") {
		t.Error("user.py should reference UserProfile directly")
	}
}