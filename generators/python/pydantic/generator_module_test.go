package pydantic

import (
	"context"
	"strings"
	"testing"

	"github.com/WhatsApp-Platform/typegen/generators"
	"github.com/WhatsApp-Platform/typegen/parser"
	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

func TestGenerate_SimpleModule(t *testing.T) {
	// Create a simple module with two files
	userProgram, err := parser.Parse(strings.NewReader(`
		struct User {
			id: int64
			name: string
		}
	`), "user.tg")
	if err != nil {
		t.Fatalf("Failed to parse user.tg: %v", err)
	}

	authProgram, err := parser.Parse(strings.NewReader(`
		enum AuthMethod {
			password
			oauth: string
		}
	`), "auth.tg")
	if err != nil {
		t.Fatalf("Failed to parse auth.tg: %v", err)
	}

	// Create module
	files := map[string]*ast.ProgramNode{
		"user.tg": userProgram,
		"auth.tg": authProgram,
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

	// Verify files were created
	expectedFiles := []string{"__init__.py", "user.py", "auth.py"}
	actualFiles := fs.ListFiles()

	if len(actualFiles) != len(expectedFiles) {
		t.Fatalf("Expected %d files, got %d: %v", len(expectedFiles), len(actualFiles), actualFiles)
	}

	for _, expectedFile := range expectedFiles {
		if !fs.FileExists(expectedFile) {
			t.Errorf("Expected file %s not found", expectedFile)
		}
	}

	// Verify __init__.py content
	if !fs.FileExists("__init__.py") {
		t.Error("__init__.py should exist")
	}

	// Verify user.py content
	userContent, exists := fs.GetFileString("user.py")
	if !exists {
		t.Error("user.py should exist")
	}
	if !strings.Contains(userContent, "class User(BaseModel):") {
		t.Error("user.py should contain User class")
	}

	// Verify auth.py content
	authContent, exists := fs.GetFileString("auth.py")
	if !exists {
		t.Error("auth.py should exist")
	}
	if !strings.Contains(authContent, "AuthMethod = Union[") {
		t.Error("auth.py should contain AuthMethod union")
	}
}

func TestGenerate_ModuleWithSubmodules(t *testing.T) {
	// Create main module files
	mainFile, err := parser.Parse(strings.NewReader(`
		struct Config {
			database: Database
		}
	`), "config.tg")
	if err != nil {
		t.Fatalf("Failed to parse config.tg: %v", err)
	}

	// Create submodule files
	dbFile, err := parser.Parse(strings.NewReader(`
		struct Database {
			host: string
			port: int32
		}
	`), "database.tg")
	if err != nil {
		t.Fatalf("Failed to parse database.tg: %v", err)
	}

	authFile, err := parser.Parse(strings.NewReader(`
		struct AuthConfig {
			secret_key: string
			expiry: int64
		}
	`), "auth.tg")
	if err != nil {
		t.Fatalf("Failed to parse auth.tg: %v", err)
	}

	// Create module structure
	mainModule := ast.NewModule("/test/module", map[string]*ast.ProgramNode{
		"config.tg": mainFile,
	})

	dbSubModule := ast.NewModule("/test/module/db", map[string]*ast.ProgramNode{
		"database.tg": dbFile,
	})

	authSubModule := ast.NewModule("/test/module/auth", map[string]*ast.ProgramNode{
		"auth.tg": authFile,
	})

	mainModule.SubModules["db"] = dbSubModule
	mainModule.SubModules["auth"] = authSubModule

	// Generate with InMemoryFS
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, mainModule, fs)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify main module files
	expectedFiles := []string{
		"__init__.py",
		"config.py",
		"db/__init__.py",
		"db/database.py",
		"auth/__init__.py",
		"auth/auth.py",
	}

	for _, expectedFile := range expectedFiles {
		if !fs.FileExists(expectedFile) {
			t.Errorf("Expected file %s not found. Available files: %v", expectedFile, fs.ListFiles())
		}
	}

	// Verify directories were created
	if !fs.DirExists("db") {
		t.Error("db subdirectory should exist")
	}
	if !fs.DirExists("auth") {
		t.Error("auth subdirectory should exist")
	}

	// Verify submodule __init__.py files
	if !fs.FileExists("db/__init__.py") {
		t.Error("db/__init__.py should exist")
	}

	if !fs.FileExists("auth/__init__.py") {
		t.Error("auth/__init__.py should exist")
	}

	// Verify content of generated files
	configContent, exists := fs.GetFileString("config.py")
	if !exists {
		t.Error("config.py should exist")
	}
	if !strings.Contains(configContent, "class Config(BaseModel):") {
		t.Error("config.py should contain Config class")
	}

	dbContent, exists := fs.GetFileString("db/database.py")
	if !exists {
		t.Error("db/database.py should exist")
	}
	if !strings.Contains(dbContent, "class Database(BaseModel):") {
		t.Error("db/database.py should contain Database class")
	}

	authContent, exists := fs.GetFileString("auth/auth.py")
	if !exists {
		t.Error("auth/auth.py should exist")
	}
	if !strings.Contains(authContent, "class AuthConfig(BaseModel):") {
		t.Error("auth/auth.py should contain AuthConfig class")
	}
}

func TestGenerate_DeepNestedSubmodules(t *testing.T) {
	// Create nested structure: main/sub1/sub2/file.tg
	deepFile, err := parser.Parse(strings.NewReader(`
		struct DeepStruct {
			value: string
		}
	`), "deep.tg")
	if err != nil {
		t.Fatalf("Failed to parse deep.tg: %v", err)
	}

	// Create the nested structure
	mainModule := ast.NewModule("/test/module", map[string]*ast.ProgramNode{})
	sub1Module := ast.NewModule("/test/module/sub1", map[string]*ast.ProgramNode{})
	sub2Module := ast.NewModule("/test/module/sub1/sub2", map[string]*ast.ProgramNode{
		"deep.tg": deepFile,
	})

	sub1Module.SubModules["sub2"] = sub2Module
	mainModule.SubModules["sub1"] = sub1Module

	// Generate
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err = generator.Generate(ctx, mainModule, fs)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify nested structure
	expectedFiles := []string{
		"__init__.py",
		"sub1/__init__.py",
		"sub1/sub2/__init__.py",
		"sub1/sub2/deep.py",
	}

	for _, expectedFile := range expectedFiles {
		if !fs.FileExists(expectedFile) {
			t.Errorf("Expected file %s not found. Available files: %v", expectedFile, fs.ListFiles())
		}
	}

	// Verify content
	deepContent, exists := fs.GetFileString("sub1/sub2/deep.py")
	if !exists {
		t.Error("sub1/sub2/deep.py should exist")
	}
	if !strings.Contains(deepContent, "class DeepStruct(BaseModel):") {
		t.Error("sub1/sub2/deep.py should contain DeepStruct class")
	}
}

func TestGenerate_EmptyModule(t *testing.T) {
	// Create empty module
	module := ast.NewModule("/test/empty", map[string]*ast.ProgramNode{})

	// Generate
	fs := generators.NewInMemoryFS()
	generator := NewGenerator()
	ctx := context.Background()

	err := generator.Generate(ctx, module, fs)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should still create __init__.py
	if !fs.FileExists("__init__.py") {
		t.Error("Empty module should still create __init__.py")
	}

	// Should only have __init__.py
	files := fs.ListFiles()
	if len(files) != 1 || files[0] != "__init__.py" {
		t.Errorf("Expected only __init__.py, got: %v", files)
	}
}