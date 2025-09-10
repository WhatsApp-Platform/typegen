package ast

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Module represents a collection of TypeGen files that form a module
type Module struct {
	// Path is the module directory path
	Path string
	
	// Name is the module name (derived from path)
	Name string
	
	// Files contains the parsed program for each .tg file in the module
	// Key is the filename (without path), value is the parsed program
	Files map[string]*ProgramNode
	
	// SubModules contains nested submodules
	// Key is the subdirectory name, value is the submodule
	SubModules map[string]*Module
}

// NewModule creates a new module from a map of files
func NewModule(modulePath string, files map[string]*ProgramNode) *Module {
	name := filepath.Base(modulePath)
	return &Module{
		Path:       modulePath,
		Name:       name,
		Files:      files,
		SubModules: make(map[string]*Module),
	}
}

// GetFile returns the program for a specific file
func (m *Module) GetFile(filename string) (*ProgramNode, bool) {
	prog, exists := m.Files[filename]
	return prog, exists
}

// FileNames returns a sorted list of all file names in the module
func (m *Module) FileNames() []string {
	var names []string
	for name := range m.Files {
		names = append(names, name)
	}
	return names
}

// AllDeclarations returns all declarations from all files in the module and submodules
func (m *Module) AllDeclarations() []Declaration {
	var decls []Declaration
	
	// Add declarations from files in this module
	for _, program := range m.Files {
		decls = append(decls, program.Declarations...)
	}
	
	// Add declarations from submodules recursively
	for _, subModule := range m.SubModules {
		decls = append(decls, subModule.AllDeclarations()...)
	}
	
	return decls
}

// AllImports returns all unique import paths from all files in the module and submodules
func (m *Module) AllImports() []string {
	importSet := make(map[string]bool)
	
	// Add imports from files in this module
	for _, program := range m.Files {
		for _, imp := range program.Imports {
			importSet[imp.Path] = true
		}
	}
	
	// Add imports from submodules recursively
	for _, subModule := range m.SubModules {
		for _, imp := range subModule.AllImports() {
			importSet[imp] = true
		}
	}
	
	var imports []string
	for imp := range importSet {
		imports = append(imports, imp)
	}
	return imports
}

// FindDeclaration finds a declaration by name across all files in the module and submodules
func (m *Module) FindDeclaration(name string) (Declaration, string, bool) {
	// Search in files of this module
	for filename, program := range m.Files {
		for _, decl := range program.Declarations {
			switch d := decl.(type) {
			case *StructNode:
				if d.Name == name {
					return d, filename, true
				}
			case *EnumNode:
				if d.Name == name {
					return d, filename, true
				}
			case *TypeAliasNode:
				if d.Name == name {
					return d, filename, true
				}
			}
		}
	}
	
	// Search in submodules recursively
	for subModuleName, subModule := range m.SubModules {
		if decl, filename, found := subModule.FindDeclaration(name); found {
			// Return path relative to the submodule
			return decl, filepath.Join(subModuleName, filename), true
		}
	}
	
	return nil, "", false
}

// String returns a string representation of the module
func (m *Module) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Module: %s (%s)", m.Name, m.Path))
	parts = append(parts, "")
	
	for filename, program := range m.Files {
		parts = append(parts, fmt.Sprintf("=== %s ===", filename))
		parts = append(parts, program.String())
		parts = append(parts, "")
	}
	
	// Add submodule information
	for subModuleName, subModule := range m.SubModules {
		parts = append(parts, fmt.Sprintf("=== SubModule: %s ===", subModuleName))
		parts = append(parts, subModule.String())
		parts = append(parts, "")
	}
	
	return strings.Join(parts, "\n")
}

// GetSubModule returns a submodule by name
func (m *Module) GetSubModule(name string) (*Module, bool) {
	subModule, exists := m.SubModules[name]
	return subModule, exists
}

// AllFiles returns all file paths including submodules (with relative paths)
func (m *Module) AllFiles() map[string]*ProgramNode {
	allFiles := make(map[string]*ProgramNode)
	
	// Add files from this module
	for filename, program := range m.Files {
		allFiles[filename] = program
	}
	
	// Add files from submodules recursively
	for subModuleName, subModule := range m.SubModules {
		for filename, program := range subModule.AllFiles() {
			// Prefix with submodule name
			relativePath := filepath.Join(subModuleName, filename)
			allFiles[relativePath] = program
		}
	}
	
	return allFiles
}

// SubModuleNames returns a sorted list of submodule names
func (m *Module) SubModuleNames() []string {
	var names []string
	for name := range m.SubModules {
		names = append(names, name)
	}
	return names
}