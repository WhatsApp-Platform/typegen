package parser

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/WhatsApp-Platform/typegen/parser/ast"
	"github.com/WhatsApp-Platform/typegen/parser/grammar"
)

// ParseError represents a parsing error
type ParseError struct {
	Message string
	Errors  []string
}

func (e *ParseError) Error() string {
	if len(e.Errors) == 0 {
		return e.Message
	}
	return fmt.Sprintf("%s:\n%s", e.Message, strings.Join(e.Errors, "\n"))
}

// ParseFile parses a TypeGen file and returns the AST
func ParseFile(filename string) (*ast.ProgramNode, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()
	
	return Parse(file, filename)
}

// Parse parses TypeGen source code from a reader and returns the AST
func Parse(input io.Reader, filename string) (*ast.ProgramNode, error) {
	lexer, result := grammar.Parse(input, filename)
	
	// Check for errors
	if errors := lexer.Errors(); len(errors) > 0 {
		return nil, &ParseError{
			Message: "parse errors occurred",
			Errors:  errors,
		}
	}
	
	if result != 0 {
		return nil, &ParseError{
			Message: "parsing failed",
		}
	}
	
	// Get the result from the lexer
	node := lexer.Result()
	if node == nil {
		return nil, &ParseError{
			Message: "no AST produced",
		}
	}
	
	program, ok := node.(*ast.ProgramNode)
	if !ok {
		return nil, &ParseError{
			Message: "invalid AST root node",
		}
	}
	
	return program, nil
}

// ParseModule parses all .tg files in a directory (non-recursive, for backwards compatibility)
func ParseModule(modulePath string) (map[string]*ast.ProgramNode, error) {
	entries, err := os.ReadDir(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module directory %s: %w", modulePath, err)
	}
	
	results := make(map[string]*ast.ProgramNode)
	
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		if !strings.HasSuffix(entry.Name(), ".tg") {
			continue
		}
		
		filePath := filepath.Join(modulePath, entry.Name())
		program, err := ParseFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
		}
		
		results[entry.Name()] = program
	}
	
	return results, nil
}

// ParseModuleToAST parses all .tg files in a directory recursively and returns an ast.Module
func ParseModuleToAST(modulePath string) (*ast.Module, error) {
	return parseModuleRecursive(modulePath)
}

// shouldSkipDirectory returns true if the directory should be skipped during parsing
func shouldSkipDirectory(name string) bool {
	skipDirs := []string{
		".git", ".svn", ".hg",           // Version control
		"node_modules", "vendor",        // Dependencies
		".vscode", ".idea",              // IDEs
		"target", "build", "dist",       // Build outputs
		"__pycache__", ".pytest_cache",  // Python cache
	}
	
	for _, skipDir := range skipDirs {
		if name == skipDir {
			return true
		}
	}
	
	// Skip hidden directories (starting with .)
	return strings.HasPrefix(name, ".")
}

// parseModuleRecursive recursively parses a module directory
func parseModuleRecursive(modulePath string) (*ast.Module, error) {
	entries, err := os.ReadDir(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read module directory %s: %w", modulePath, err)
	}
	
	files := make(map[string]*ast.ProgramNode)
	subModules := make(map[string]*ast.Module)
	
	for _, entry := range entries {
		if entry.IsDir() {
			// Skip certain directories
			if shouldSkipDirectory(entry.Name()) {
				continue
			}
			
			// Parse subdirectory as submodule
			subModulePath := filepath.Join(modulePath, entry.Name())
			subModule, err := parseModuleRecursive(subModulePath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse submodule %s: %w", subModulePath, err)
			}
			
			// Only include submodules that have content
			if len(subModule.Files) > 0 || len(subModule.SubModules) > 0 {
				subModules[entry.Name()] = subModule
			}
		} else if strings.HasSuffix(entry.Name(), ".tg") {
			// Parse .tg file
			filePath := filepath.Join(modulePath, entry.Name())
			program, err := ParseFile(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
			}
			
			files[entry.Name()] = program
		}
	}
	
	// Create the module
	module := ast.NewModule(modulePath, files)
	module.SubModules = subModules
	
	return module, nil
}