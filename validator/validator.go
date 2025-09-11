package validator

import (
	"fmt"
	"strings"

	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

// Validator validates TypeGen modules for correctness
type Validator struct {
	registry *TypeRegistry
	result   *ValidationResult
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		result: NewValidationResult(),
	}
}

// Validate validates an entire module and returns validation results
func (v *Validator) Validate(module *ast.Module) *ValidationResult {
	v.result = NewValidationResult()
	v.registry = buildTypeRegistry(module)

	// Validate all files in the module recursively
	v.validateModule(module, "")

	// Check for circular dependencies after all types are registered
	v.validateCircularDependencies()

	return v.result
}

// validateModule validates a module and its submodules recursively
func (v *Validator) validateModule(module *ast.Module, basePath string) {
	// Validate files in this module
	for filename, program := range module.Files {
		fullPath := basePath
		if fullPath != "" {
			fullPath += "/"
		}
		fullPath += filename

		v.validateProgram(program, fullPath)
	}

	// Validate submodules recursively
	for subModuleName, subModule := range module.SubModules {
		// Validate submodule name follows snake_case
		if !IsValidSnakeCase(subModuleName) {
			v.result.AddError(
				NamingConventionError,
				fmt.Sprintf("module name '%s' should follow snake_case convention", subModuleName),
				basePath,
				0, 0,
				fmt.Sprintf("use '%s'", SuggestSnakeCase(subModuleName)),
			)
		}

		subBasePath := basePath
		if subBasePath != "" {
			subBasePath += "/"
		}
		subBasePath += subModuleName
		v.validateModule(subModule, subBasePath)
	}
}

// validateProgram validates a single program (file)
func (v *Validator) validateProgram(program *ast.ProgramNode, filename string) {
	// Track names in this file to detect duplicates
	declNames := make(map[string]ast.Declaration)

	// Validate imports
	for _, imp := range program.Imports {
		v.validateImport(imp, filename)
	}

	// Validate declarations
	for _, decl := range program.Declarations {
		v.validateDeclaration(decl, filename, declNames)
	}
}

// validateImport validates an import statement
func (v *Validator) validateImport(imp *ast.ImportNode, filename string) {
	pos := imp.Pos()
	if !IsValidModuleName(imp.Path) {
		v.result.AddError(
			InvalidImportError,
			fmt.Sprintf("import path '%s' should follow snake_case convention for module names", imp.Path),
			filename,
			pos.Line, pos.Column,
			fmt.Sprintf("use '%s'", SuggestModuleName(imp.Path)),
		)
	}
}

// validateDeclaration validates a single declaration
func (v *Validator) validateDeclaration(decl ast.Declaration, filename string, declNames map[string]ast.Declaration) {
	var declName string
	var declType string

	switch d := decl.(type) {
	case *ast.StructNode:
		declName = d.Name
		declType = "struct"
		v.validateStruct(d, filename)

	case *ast.EnumNode:
		declName = d.Name
		declType = "enum"
		v.validateEnum(d, filename)

	case *ast.TypeAliasNode:
		declName = d.Name
		declType = "type alias"
		v.validateTypeAlias(d, filename)

	case *ast.ConstantNode:
		declName = d.Name
		declType = "constant"
		v.validateConstant(d, filename)
	}

	// Check for duplicate declarations
	if existing, exists := declNames[declName]; exists {
		existingPos := existing.Pos()
		declPos := decl.Pos()
		v.result.AddError(
			DuplicateTypeError,
			fmt.Sprintf("duplicate %s '%s' (first declared at line %d)", declType, declName, existingPos.Line),
			filename,
			declPos.Line, declPos.Column,
			"rename one of the declarations",
		)
	} else {
		declNames[declName] = decl
	}
}

// validateStruct validates a struct declaration
func (v *Validator) validateStruct(s *ast.StructNode, filename string) {
	pos := s.Pos()
	// Validate struct name (PascalCase)
	if !IsValidPascalCase(s.Name) {
		v.result.AddError(
			NamingConventionError,
			fmt.Sprintf("struct name '%s' should follow PascalCase convention", s.Name),
			filename,
			pos.Line, pos.Column,
			fmt.Sprintf("use '%s'", SuggestPascalCase(s.Name)),
		)
	}

	// Validate fields
	fieldNames := make(map[string]*ast.FieldNode)
	for _, field := range s.Fields {
		v.validateField(field, filename, fieldNames)
	}
}

// validateField validates a struct field
func (v *Validator) validateField(field *ast.FieldNode, filename string, fieldNames map[string]*ast.FieldNode) {
	pos := field.Pos()
	// Validate field name (snake_case)
	if !IsValidSnakeCase(field.Name) {
		v.result.AddError(
			NamingConventionError,
			fmt.Sprintf("field name '%s' should follow snake_case convention", field.Name),
			filename,
			pos.Line, pos.Column,
			fmt.Sprintf("use '%s'", SuggestSnakeCase(field.Name)),
		)
	}

	// Check for duplicate field names
	if existing, exists := fieldNames[field.Name]; exists {
		existingPos := existing.Pos()
		v.result.AddError(
			DuplicateFieldError,
			fmt.Sprintf("duplicate field '%s' (first declared at line %d)", field.Name, existingPos.Line),
			filename,
			pos.Line, pos.Column,
			"rename one of the fields",
		)
	} else {
		fieldNames[field.Name] = field
	}

	// Validate field type
	v.validateType(field.Type, filename, pos.Line, pos.Column)
}

// validateEnum validates an enum declaration
func (v *Validator) validateEnum(e *ast.EnumNode, filename string) {
	pos := e.Pos()
	// Validate enum name (PascalCase)
	if !IsValidPascalCase(e.Name) {
		v.result.AddError(
			NamingConventionError,
			fmt.Sprintf("enum name '%s' should follow PascalCase convention", e.Name),
			filename,
			pos.Line, pos.Column,
			fmt.Sprintf("use '%s'", SuggestPascalCase(e.Name)),
		)
	}

	// Validate variants
	variantNames := make(map[string]*ast.EnumVariantNode)
	for _, variant := range e.Variants {
		v.validateEnumVariant(variant, filename, variantNames)
	}
}

// validateEnumVariant validates an enum variant
func (v *Validator) validateEnumVariant(variant *ast.EnumVariantNode, filename string, variantNames map[string]*ast.EnumVariantNode) {
	pos := variant.Pos()
	// Validate variant name (snake_case)
	if !IsValidSnakeCase(variant.Name) {
		v.result.AddError(
			NamingConventionError,
			fmt.Sprintf("enum variant '%s' should follow snake_case convention", variant.Name),
			filename,
			pos.Line, pos.Column,
			fmt.Sprintf("use '%s'", SuggestSnakeCase(variant.Name)),
		)
	}

	// Check for duplicate variant names
	if existing, exists := variantNames[variant.Name]; exists {
		existingPos := existing.Pos()
		v.result.AddError(
			DuplicateVariantError,
			fmt.Sprintf("duplicate variant '%s' (first declared at line %d)", variant.Name, existingPos.Line),
			filename,
			pos.Line, pos.Column,
			"rename one of the variants",
		)
	} else {
		variantNames[variant.Name] = variant
	}

	// Validate payload type if present
	if variant.Payload != nil {
		v.validateType(variant.Payload, filename, pos.Line, pos.Column)
	}
}

// validateTypeAlias validates a type alias declaration
func (v *Validator) validateTypeAlias(alias *ast.TypeAliasNode, filename string) {
	pos := alias.Pos()
	// Validate alias name (PascalCase)
	if !IsValidPascalCase(alias.Name) {
		v.result.AddError(
			NamingConventionError,
			fmt.Sprintf("type alias '%s' should follow PascalCase convention", alias.Name),
			filename,
			pos.Line, pos.Column,
			fmt.Sprintf("use '%s'", SuggestPascalCase(alias.Name)),
		)
	}

	// Validate aliased type
	v.validateType(alias.Type, filename, pos.Line, pos.Column)
}

// validateConstant validates a constant declaration
func (v *Validator) validateConstant(constant *ast.ConstantNode, filename string) {
	pos := constant.Pos()
	// Validate constant name (CONSTANT_CASE)
	if !IsValidConstantCase(constant.Name) {
		v.result.AddError(
			NamingConventionError,
			fmt.Sprintf("constant name '%s' should follow CONSTANT_CASE convention", constant.Name),
			filename,
			pos.Line, pos.Column,
			fmt.Sprintf("use '%s'", SuggestConstantCase(constant.Name)),
		)
	}

	// Validate constant value exists (basic check)
	if constant.Value == nil {
		v.result.AddError(
			InvalidConstantError,
			fmt.Sprintf("constant '%s' must have a value", constant.Name),
			filename,
			pos.Line, pos.Column,
			"provide a value for the constant",
		)
	}
}

// validateType validates a type reference
func (v *Validator) validateType(typeNode ast.Type, filename string, line, column int) {
	switch t := typeNode.(type) {
	case *ast.PrimitiveType:
		v.validatePrimitiveType(t, filename, line, column)

	case *ast.NamedType:
		v.validateNamedType(t, filename, line, column)

	case *ast.ArrayType:
		v.validateType(t.ElementType, filename, line, column)

	case *ast.MapType:
		v.validateMapType(t, filename, line, column)

	case *ast.OptionalType:
		v.validateOptionalType(t, filename, line, column)
	}
}

// validatePrimitiveType validates a primitive type
func (v *Validator) validatePrimitiveType(primitive *ast.PrimitiveType, filename string, line, column int) {
	if !IsValidPrimitiveType(primitive.Name) {
		v.result.AddError(
			InvalidPrimitiveError,
			fmt.Sprintf("'%s' is not a valid primitive type", primitive.Name),
			filename,
			line, column,
			"use one of: int8, int16, int32, int64, nat8, nat16, nat32, nat64, float32, float64, string, bool, json, datetime, date, time",
		)
	}
}

// validateNamedType validates a named type reference
func (v *Validator) validateNamedType(named *ast.NamedType, filename string, line, column int) {
	// Check if type exists
	if !v.registry.TypeExists(named.Name, filename) {
		v.result.AddError(
			UndefinedTypeError,
			fmt.Sprintf("undefined type '%s'", named.Name),
			filename,
			line, column,
			"define the type or check the spelling",
		)
	}
}

// validateMapType validates a map type
func (v *Validator) validateMapType(mapType *ast.MapType, filename string, line, column int) {
	// Validate key type - must be primitive and valid as map key
	if primitive, ok := mapType.KeyType.(*ast.PrimitiveType); ok {
		if !IsValidMapKeyType(primitive.Name) {
			v.result.AddError(
				InvalidMapKeyError,
				fmt.Sprintf("map key type '%s' is not valid", primitive.Name),
				filename,
				line, column,
				"use string or integer types for map keys",
			)
		}
	} else {
		v.result.AddError(
			InvalidMapKeyError,
			"map key must be a primitive type",
			filename,
			line, column,
			"use string or integer types for map keys",
		)
	}

	// Validate key and value types
	v.validateType(mapType.KeyType, filename, line, column)
	v.validateType(mapType.ValueType, filename, line, column)
}

// validateOptionalType validates an optional type
func (v *Validator) validateOptionalType(optional *ast.OptionalType, filename string, line, column int) {
	// Check for double-wrapped optionals (??)
	if _, isOptional := optional.ElementType.(*ast.OptionalType); isOptional {
		v.result.AddError(
			InvalidOptionalError,
			"double-wrapped optional types are not allowed",
			filename,
			line, column,
			"use single optional marker ?Type",
		)
	}

	// Validate the wrapped type
	v.validateType(optional.ElementType, filename, line, column)
}

// validateCircularDependencies checks for circular dependencies
func (v *Validator) validateCircularDependencies() {
	cycles := v.registry.ValidateCircularDependencies()

	for _, cycle := range cycles {
		// Find the first type in the cycle to report the error
		types := strings.Split(cycle, " -> ")
		if len(types) > 0 {
			firstType := types[0]
			parts := strings.Split(firstType, "::")
			if len(parts) == 2 {
				file := parts[0]

				v.result.AddError(
					CircularDependencyError,
					fmt.Sprintf("circular dependency detected: %s", cycle),
					file,
					0, 0, // We don't have exact position for circular deps
					"restructure types to eliminate circular references",
				)
			}
		}
	}
}
