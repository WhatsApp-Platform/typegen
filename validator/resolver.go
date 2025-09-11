package validator

import (
	"fmt"
	"strings"

	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

// TypeRegistry keeps track of all type declarations in a module
type TypeRegistry struct {
	types       map[string]*TypeInfo     // Fully qualified name -> TypeInfo
	moduleTypes map[string]*TypeInfo     // Module path qualified name -> TypeInfo
	currentFile string                   // Current file being processed
}

// TypeInfo contains information about a declared type
type TypeInfo struct {
	Name     string
	DeclType string // "struct", "enum", "alias", "constant"  
	File     string
	Line     int
	Column   int
}

// NewTypeRegistry creates a new type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types:       make(map[string]*TypeInfo),
		moduleTypes: make(map[string]*TypeInfo),
	}
}

// RegisterType registers a type declaration in the registry
func (r *TypeRegistry) RegisterType(name, declType, file string, line, column int) {
	qualifiedName := r.qualifyName(name, file)
	typeInfo := &TypeInfo{
		Name:     name,
		DeclType: declType,
		File:     file,
		Line:     line,
		Column:   column,
	}
	
	r.types[qualifiedName] = typeInfo
	
	// Also register by module path for cross-module lookups
	modulePath := r.fileToModulePath(file)
	moduleQualifiedName := fmt.Sprintf("%s.%s", modulePath, name)
	r.moduleTypes[moduleQualifiedName] = typeInfo
}

// qualifyName creates a fully qualified type name based on file location
func (r *TypeRegistry) qualifyName(name, file string) string {
	// For now, we'll use file path as the qualifier
	// In a full implementation, this would use module paths
	return fmt.Sprintf("%s::%s", file, name)
}

// fileToModulePath converts a file path to a module path
func (r *TypeRegistry) fileToModulePath(file string) string {
	// Convert file path to module path by removing .tg extension and replacing / with .
	// e.g., "auth/user.tg" -> "auth.user", "user.tg" -> "user"
	path := strings.TrimSuffix(file, ".tg")
	return strings.ReplaceAll(path, "/", ".")
}

// TypeExists checks if a type exists in the registry
func (r *TypeRegistry) TypeExists(name, currentFile string) bool {
	// Check primitive types first
	if IsValidPrimitiveType(name) {
		return true
	}
	
	// Check qualified name in current file
	qualifiedName := r.qualifyName(name, currentFile)
	if _, exists := r.types[qualifiedName]; exists {
		return true
	}
	
	// Check if it's a cross-file reference
	// For now, we'll check all files in the module
	for qualName := range r.types {
		if strings.HasSuffix(qualName, "::"+name) {
			return true
		}
	}
	
	return false
}

// QualifiedTypeExists checks if a qualified type exists for a given module path
func (r *TypeRegistry) QualifiedTypeExists(qualifiedName, modulePath string) bool {
	// qualifiedName is like "auth.Token", modulePath is like "some.auth.module"
	// We need to match this against our registered module types
	
	// Try the exact qualified name first
	if _, exists := r.moduleTypes[qualifiedName]; exists {
		return true
	}
	
	// Try matching the module path
	parts := strings.SplitN(qualifiedName, ".", 2)
	if len(parts) != 2 {
		return false
	}
	
	moduleAlias := parts[0]  // "auth"
	typeName := parts[1]     // "Token"
	
	// Look for types that match the module path ending
	for moduleTypeKey := range r.moduleTypes {
		// Check if this is the right type and the module path matches
		if strings.HasSuffix(moduleTypeKey, "."+typeName) {
			// Extract the module part of the key
			moduleKeyParts := strings.SplitN(moduleTypeKey, "."+typeName, 2)
			if len(moduleKeyParts) > 0 {
				keyModulePath := moduleKeyParts[0]
				// Check if the import path ends with the module alias
				if strings.HasSuffix(modulePath, "."+moduleAlias) || modulePath == moduleAlias {
					return true
				}
				// Also check direct module path match
				if keyModulePath == modulePath {
					return true
				}
			}
		}
	}
	
	return false
}

// FindType finds type information by name
func (r *TypeRegistry) FindType(name, currentFile string) (*TypeInfo, bool) {
	// Check qualified name in current file first
	qualifiedName := r.qualifyName(name, currentFile)
	if info, exists := r.types[qualifiedName]; exists {
		return info, true
	}
	
	// Check cross-file references
	for qualName, info := range r.types {
		if strings.HasSuffix(qualName, "::"+name) {
			return info, true
		}
	}
	
	return nil, false
}

// GetAllTypes returns all registered types
func (r *TypeRegistry) GetAllTypes() map[string]*TypeInfo {
	return r.types
}


// buildTypeRegistry builds a type registry for the entire module
func buildTypeRegistry(module *ast.Module) *TypeRegistry {
	registry := NewTypeRegistry()
	
	// Process all files in the module recursively
	processModuleForRegistry(module, "", registry)
	
	return registry
}

// processModuleForRegistry processes a module and its submodules for type registration
func processModuleForRegistry(module *ast.Module, basePath string, registry *TypeRegistry) {
	// Process files in this module
	for filename, program := range module.Files {
		fullPath := basePath
		if fullPath != "" {
			fullPath += "/"
		}
		fullPath += filename
		
		registry.currentFile = fullPath
		
		// Register all type declarations
		for _, decl := range program.Declarations {
			pos := decl.Pos()
			switch d := decl.(type) {
			case *ast.StructNode:
				registry.RegisterType(d.Name, "struct", fullPath, pos.Line, pos.Column)
				
			case *ast.EnumNode:
				registry.RegisterType(d.Name, "enum", fullPath, pos.Line, pos.Column)
				
			case *ast.TypeAliasNode:
				registry.RegisterType(d.Name, "alias", fullPath, pos.Line, pos.Column)
				
			case *ast.ConstantNode:
				registry.RegisterType(d.Name, "constant", fullPath, pos.Line, pos.Column)
			}
		}
	}
	
	// Process submodules recursively
	for subModuleName, subModule := range module.SubModules {
		subBasePath := basePath
		if subBasePath != "" {
			subBasePath += "/"
		}
		subBasePath += subModuleName
		processModuleForRegistry(subModule, subBasePath, registry)
	}
}

