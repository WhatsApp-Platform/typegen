package validator

import (
	"regexp"
	"strings"
	"unicode"
)

// Naming convention regular expressions
var (
	// snake_case: lowercase letters, numbers, underscores, must start with letter
	snakeCaseRegex = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

	// PascalCase: starts with uppercase, followed by letters/numbers
	pascalCaseRegex = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*$`)

	// CONSTANT_CASE: uppercase letters, numbers, underscores, must start with letter
	constantCaseRegex = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

	// smashcase/flatcase: lowercase letters and numbers only, no underscores
	smashCaseRegex = regexp.MustCompile(`^[a-z0-9]+$`)
)

// ValidPrimitiveTypes lists all valid primitive types in TypeGen
var ValidPrimitiveTypes = map[string]bool{
	// Integer types
	"int8":  true,
	"int16": true,
	"int32": true,
	"int64": true,

	// Natural number types
	"nat8":  true,
	"nat16": true,
	"nat32": true,
	"nat64": true,

	// Float types
	"float32": true,
	"float64": true,

	// String and boolean
	"string": true,
	"bool":   true,

	// JSON type
	"json": true,

	// Time types
	"datetime": true,
	"date":     true,
	"time":     true,
}

// ValidMapKeyTypes lists primitive types that can be used as map keys
var ValidMapKeyTypes = map[string]bool{
	"string": true,
	"int8":   true,
	"int16":  true,
	"int32":  true,
	"int64":  true,
	"nat8":   true,
	"nat16":  true,
	"nat32":  true,
	"nat64":  true,
}

// IsValidSnakeCase checks if a string follows snake_case convention
func IsValidSnakeCase(s string) bool {
	return snakeCaseRegex.MatchString(s)
}

// IsValidPascalCase checks if a string follows PascalCase convention
func IsValidPascalCase(s string) bool {
	return pascalCaseRegex.MatchString(s)
}

// IsValidConstantCase checks if a string follows CONSTANT_CASE convention
func IsValidConstantCase(s string) bool {
	return constantCaseRegex.MatchString(s)
}

// IsValidSmashCase checks if a string follows smashcase/flatcase convention
func IsValidSmashCase(s string) bool {
	return smashCaseRegex.MatchString(s)
}

// IsValidPrimitiveType checks if a type name is a valid primitive type
func IsValidPrimitiveType(typeName string) bool {
	return ValidPrimitiveTypes[typeName]
}

// IsValidMapKeyType checks if a type can be used as a map key
func IsValidMapKeyType(typeName string) bool {
	return ValidMapKeyTypes[typeName]
}

// SuggestSnakeCase converts a string to snake_case
func SuggestSnakeCase(s string) string {
	if IsValidSnakeCase(s) {
		return s
	}

	// Convert camelCase/PascalCase to snake_case
	var result strings.Builder

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}

	return result.String()
}

// SuggestPascalCase converts a string to PascalCase
func SuggestPascalCase(s string) string {
	if IsValidPascalCase(s) {
		return s
	}

	// Convert snake_case to PascalCase
	parts := strings.Split(s, "_")
	var result strings.Builder

	for _, part := range parts {
		if len(part) > 0 {
			result.WriteRune(unicode.ToUpper(rune(part[0])))
			if len(part) > 1 {
				result.WriteString(strings.ToLower(part[1:]))
			}
		}
	}

	return result.String()
}

// SuggestConstantCase converts a string to CONSTANT_CASE
func SuggestConstantCase(s string) string {
	if IsValidConstantCase(s) {
		return s
	}

	// Convert to snake_case first, then uppercase
	snakeCase := SuggestSnakeCase(s)
	return strings.ToUpper(snakeCase)
}

// IsValidModuleName checks if a module name follows the correct convention
// Module names should be snake_case separated by dots
func IsValidModuleName(name string) bool {
	if name == "" {
		return false
	}

	parts := strings.SplitSeq(name, ".")
	for part := range parts {
		if !IsValidSnakeCase(part) {
			return false
		}
	}

	return true
}

// SuggestModuleName converts a module name to the correct convention
func SuggestModuleName(name string) string {
	if IsValidModuleName(name) {
		return name
	}

	parts := strings.Split(name, ".")
	var correctedParts []string

	for _, part := range parts {
		correctedParts = append(correctedParts, SuggestSnakeCase(part))
	}

	return strings.Join(correctedParts, ".")
}
