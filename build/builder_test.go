package build

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/WhatsApp-Platform/typegen/generators"
	"github.com/WhatsApp-Platform/typegen/parser/ast"
)

// MockGenerator for testing
type MockGenerator struct {
	config    map[string]string
	generated bool
	shouldErr bool
}

func NewMockGenerator() generators.Generator {
	return &MockGenerator{
		config: make(map[string]string),
	}
}

func (g *MockGenerator) SetConfig(config map[string]string) {
	g.config = config
}

func (g *MockGenerator) Generate(ctx context.Context, module *ast.Module, dest generators.FS) error {
	g.generated = true
	if g.shouldErr {
		return fmt.Errorf("mock generation error")
	}
	return nil
}

func TestBuilder(t *testing.T) {
	// Register mock generator
	generators.Register("mock", NewMockGenerator)
	defer func() {
		// Note: There's no unregister function, so this test affects global state
		// In a real scenario, you might want to use a separate registry for tests
	}()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "successful build",
			config: &Config{
				Version: 1,
				Config: map[string]string{
					"global": "value",
				},
				Generate: []GenerateTask{
					{
						Generator: "mock",
						Input:     ".",
						Output:    "./output",
						Config: map[string]string{
							"task": "value",
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "unknown generator",
			config: &Config{
				Version: 1,
				Generate: []GenerateTask{
					{
						Generator: "nonexistent",
						Input:     ".",
						Output:    "./output",
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder(tt.config)

			err := builder.ValidateGenerators()
			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected validation error: %v", err)
				return
			}

			// Note: We can't easily test the full Build() method without creating
			// actual .tg files and setting up the parser. In practice, this would
			// require integration tests.
		})
	}
}

func TestValidateGenerators(t *testing.T) {
	// Get list of available generators
	available := generators.List()
	
	tests := []struct {
		name          string
		generators    []string
		expectError   bool
		errorContains string
	}{
		{
			name:        "all valid generators",
			generators:  available, // Use all actually registered generators
			expectError: false,
		},
		{
			name:          "invalid generator",
			generators:    []string{"nonexistent"},
			expectError:   true,
			errorContains: "unknown generators",
		},
		{
			name:          "mixed valid and invalid",
			generators:    append(available, "nonexistent"),
			expectError:   true,
			errorContains: "nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Version:  1,
				Generate: make([]GenerateTask, len(tt.generators)),
			}

			for i, gen := range tt.generators {
				config.Generate[i] = GenerateTask{
					Generator: gen,
					Input:     ".",
					Output:    "./output",
				}
			}

			builder := NewBuilder(config)
			err := builder.ValidateGenerators()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %s", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}