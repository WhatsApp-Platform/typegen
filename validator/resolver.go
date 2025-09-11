package validator

import (
	"fmt"
	"strings"

	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

// TypeRegistry keeps track of all type declarations in a module
type TypeRegistry struct {
	types       map[string]*TypeInfo // Fully qualified name -> TypeInfo
	currentFile string               // Current file being processed
}

// TypeInfo contains information about a declared type
type TypeInfo struct {
	Name         string
	DeclType     string // "struct", "enum", "alias", "constant"  
	File         string
	Line         int
	Column       int
	Dependencies []string // Types this type depends on
}

// NewTypeRegistry creates a new type registry
func NewTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		types: make(map[string]*TypeInfo),
	}
}

// RegisterType registers a type declaration in the registry
func (r *TypeRegistry) RegisterType(name, declType, file string, line, column int) {
	qualifiedName := r.qualifyName(name, file)
	r.types[qualifiedName] = &TypeInfo{
		Name:         name,
		DeclType:     declType,
		File:         file,
		Line:         line,
		Column:       column,
		Dependencies: make([]string, 0),
	}
}

// AddDependency adds a type dependency
func (r *TypeRegistry) AddDependency(fromType, toType, file string) {
	qualifiedFrom := r.qualifyName(fromType, file)
	if info, exists := r.types[qualifiedFrom]; exists {
		info.Dependencies = append(info.Dependencies, toType)
	}
}

// qualifyName creates a fully qualified type name based on file location
func (r *TypeRegistry) qualifyName(name, file string) string {
	// For now, we'll use file path as the qualifier
	// In a full implementation, this would use module paths
	return fmt.Sprintf("%s::%s", file, name)
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

// ValidateCircularDependencies checks for circular dependencies in type declarations
func (r *TypeRegistry) ValidateCircularDependencies() []string {
	var cycles []string
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	
	for typeName := range r.types {
		if !visited[typeName] {
			if cycle := r.detectCycle(typeName, visited, recStack, []string{}); cycle != nil {
				cycles = append(cycles, strings.Join(cycle, " -> "))
			}
		}
	}
	
	return cycles
}

// detectCycle performs DFS to detect circular dependencies
func (r *TypeRegistry) detectCycle(typeName string, visited, recStack map[string]bool, path []string) []string {
	visited[typeName] = true
	recStack[typeName] = true
	path = append(path, typeName)
	
	if typeInfo, exists := r.types[typeName]; exists {
		for _, dep := range typeInfo.Dependencies {
			qualifiedDep := r.qualifyName(dep, typeInfo.File)
			
			if !visited[qualifiedDep] {
				if cycle := r.detectCycle(qualifiedDep, visited, recStack, path); cycle != nil {
					return cycle
				}
			} else if recStack[qualifiedDep] {
				// Found cycle - return the cycle path
				cycleStart := -1
				for i, p := range path {
					if p == qualifiedDep {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					return append(path[cycleStart:], qualifiedDep)
				}
			}
		}
	}
	
	recStack[typeName] = false
	return nil
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
				
				// Add dependencies from field types
				for _, field := range d.Fields {
					addTypeDependencies(registry, d.Name, field.Type, fullPath)
				}
				
			case *ast.EnumNode:
				registry.RegisterType(d.Name, "enum", fullPath, pos.Line, pos.Column)
				
				// Add dependencies from variant types
				for _, variant := range d.Variants {
					if variant.Payload != nil {
						addTypeDependencies(registry, d.Name, variant.Payload, fullPath)
					}
				}
				
			case *ast.TypeAliasNode:
				registry.RegisterType(d.Name, "alias", fullPath, pos.Line, pos.Column)
				addTypeDependencies(registry, d.Name, d.Type, fullPath)
				
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

// addTypeDependencies extracts type dependencies from an AST type node
func addTypeDependencies(registry *TypeRegistry, fromType string, typeNode ast.Type, file string) {
	switch t := typeNode.(type) {
	case *ast.NamedType:
		registry.AddDependency(fromType, t.Name, file)
		
	case *ast.ArrayType:
		addTypeDependencies(registry, fromType, t.ElementType, file)
		
	case *ast.MapType:
		addTypeDependencies(registry, fromType, t.KeyType, file)
		addTypeDependencies(registry, fromType, t.ValueType, file)
		
	case *ast.OptionalType:
		addTypeDependencies(registry, fromType, t.ElementType, file)
		
	// PrimitiveType doesn't need dependencies
	}
}