package validator

import (
	"fmt"
	"sort"
	"strings"
)

// ValidationErrorType represents the type of validation error
type ValidationErrorType string

const (
	// Type-related errors
	UndefinedTypeError   ValidationErrorType = "undefined_type"
	InvalidPrimitiveError ValidationErrorType = "invalid_primitive"
	InvalidMapKeyError    ValidationErrorType = "invalid_map_key"
	
	// Naming convention errors
	NamingConventionError ValidationErrorType = "naming_convention"
	
	// Duplicate errors
	DuplicateTypeError     ValidationErrorType = "duplicate_type"
	DuplicateFieldError    ValidationErrorType = "duplicate_field"
	DuplicateVariantError  ValidationErrorType = "duplicate_variant"
	DuplicateConstantError ValidationErrorType = "duplicate_constant"
	
	// Import errors
	InvalidImportError ValidationErrorType = "invalid_import"
	
	// Structure errors
	InvalidOptionalError ValidationErrorType = "invalid_optional"
	InvalidConstantError ValidationErrorType = "invalid_constant"
)

// ValidationError represents a single validation error with context
type ValidationError struct {
	Type        ValidationErrorType
	Message     string
	File        string
	Line        int
	Column      int
	Suggestion  string // Optional suggestion for fixing
}

// Error implements the error interface
func (e ValidationError) Error() string {
	pos := fmt.Sprintf("%s:%d:%d", e.File, e.Line, e.Column)
	msg := fmt.Sprintf("%s: %s", pos, e.Message)
	if e.Suggestion != "" {
		msg += fmt.Sprintf("\n  Suggestion: %s", e.Suggestion)
	}
	return msg
}

// ValidationResult holds the results of validation
type ValidationResult struct {
	Errors []ValidationError
	Valid  bool
}

// HasErrors returns true if there are validation errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorCount returns the number of validation errors
func (r *ValidationResult) ErrorCount() int {
	return len(r.Errors)
}

// AddError adds a validation error to the result
func (r *ValidationResult) AddError(errorType ValidationErrorType, message, file string, line, column int, suggestion string) {
	r.Errors = append(r.Errors, ValidationError{
		Type:       errorType,
		Message:    message,
		File:       file,
		Line:       line,
		Column:     column,
		Suggestion: suggestion,
	})
	r.Valid = false
}

// SortErrors sorts validation errors by file, then by line, then by column
func (r *ValidationResult) SortErrors() {
	sort.Slice(r.Errors, func(i, j int) bool {
		a, b := r.Errors[i], r.Errors[j]
		
		// Sort by file first
		if a.File != b.File {
			return a.File < b.File
		}
		
		// Then by line
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		
		// Finally by column
		return a.Column < b.Column
	})
}

// GroupedErrors returns errors grouped by file for better readability
func (r *ValidationResult) GroupedErrors() map[string][]ValidationError {
	groups := make(map[string][]ValidationError)
	
	for _, err := range r.Errors {
		groups[err.File] = append(groups[err.File], err)
	}
	
	return groups
}

// String returns a formatted string representation of all validation errors
func (r *ValidationResult) String() string {
	if len(r.Errors) == 0 {
		return "No validation errors"
	}
	
	r.SortErrors()
	
	var parts []string
	parts = append(parts, fmt.Sprintf("Validation errors found (%d):", len(r.Errors)))
	parts = append(parts, "")
	
	// Group errors by file
	groups := r.GroupedErrors()
	
	// Sort file names
	var files []string
	for file := range groups {
		files = append(files, file)
	}
	sort.Strings(files)
	
	// Format errors by file
	for i, file := range files {
		if i > 0 {
			parts = append(parts, "")
		}
		
		parts = append(parts, fmt.Sprintf("%s:", file))
		
		for _, err := range groups[file] {
			line := fmt.Sprintf("  %d:%d: %s", err.Line, err.Column, err.Message)
			parts = append(parts, line)
			
			if err.Suggestion != "" {
				parts = append(parts, fmt.Sprintf("    Suggestion: %s", err.Suggestion))
			}
		}
	}
	
	return strings.Join(parts, "\n")
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors: make([]ValidationError, 0),
		Valid:  true,
	}
}