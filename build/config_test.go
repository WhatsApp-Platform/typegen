package build

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		yamlContent    string
		expectError    bool
		expectedTasks  int
		expectedVersion int
	}{
		{
			name: "valid config",
			yamlContent: `version: 1
config:
  module-path: example.com/test
generate:
  - generator: go
    input: ./api
    output: ./generated/go
  - generator: python+pydantic
    output: ./generated/python
    config:
      module-path: test.api
`,
			expectError:     false,
			expectedTasks:   2,
			expectedVersion: 1,
		},
		{
			name: "minimal config",
			yamlContent: `generate:
  - generator: go
    output: ./output
`,
			expectError:     false,
			expectedTasks:   1,
			expectedVersion: 1, // Should default to 1
		},
		{
			name: "invalid version",
			yamlContent: `version: 2
generate:
  - generator: go
    output: ./output
`,
			expectError: true,
		},
		{
			name: "missing generator",
			yamlContent: `generate:
  - output: ./output
`,
			expectError: true,
		},
		{
			name: "missing output",
			yamlContent: `generate:
  - generator: go
`,
			expectError: true,
		},
		{
			name: "no generate tasks",
			yamlContent: `version: 1
config:
  module-path: test
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "test-config.yaml")
			
			err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			// Create input directory for tasks if needed
			if !tt.expectError {
				inputDir := filepath.Join(tmpDir, "api")
				err = os.MkdirAll(inputDir, 0755)
				if err != nil {
					t.Fatalf("Failed to create input directory: %v", err)
				}
			}

			// Change to temp directory for relative path resolution
			oldWd, _ := os.Getwd()
			defer os.Chdir(oldWd)
			os.Chdir(tmpDir)

			config, err := LoadConfig(configPath)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if config.Version != tt.expectedVersion {
				t.Errorf("Expected version %d, got %d", tt.expectedVersion, config.Version)
			}

			if len(config.Generate) != tt.expectedTasks {
				t.Errorf("Expected %d tasks, got %d", tt.expectedTasks, len(config.Generate))
			}
		})
	}
}

func TestConfigMerging(t *testing.T) {
	config := &Config{
		Version: 1,
		Config: map[string]string{
			"global-key": "global-value",
			"override-key": "global-override",
		},
		Generate: []GenerateTask{
			{
				Generator: "go",
				Input:     ".",
				Output:    "./output",
				Config: map[string]string{
					"task-key": "task-value",
					"override-key": "task-override",
				},
			},
		},
	}

	merged := config.MergedConfig(0)

	// Check global config is included
	if merged["global-key"] != "global-value" {
		t.Errorf("Expected global-key to be 'global-value', got '%s'", merged["global-key"])
	}

	// Check task config is included
	if merged["task-key"] != "task-value" {
		t.Errorf("Expected task-key to be 'task-value', got '%s'", merged["task-key"])
	}

	// Check task config overrides global config
	if merged["override-key"] != "task-override" {
		t.Errorf("Expected override-key to be 'task-override' (task should override global), got '%s'", merged["override-key"])
	}

	// Check invalid index returns nil
	if config.MergedConfig(-1) != nil || config.MergedConfig(1) != nil {
		t.Error("Expected nil for invalid task index")
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
}

func TestLoadConfigDefaultPath(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Create typegen.yaml in current directory
	yamlContent := `generate:
  - generator: go
    output: ./output
`
	err := os.WriteFile("typegen.yaml", []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Create input directory
	err = os.MkdirAll(".", 0755)
	if err != nil {
		t.Fatalf("Failed to create input directory: %v", err)
	}

	// Load config with empty path (should find typegen.yaml)
	config, err := LoadConfig("")
	if err != nil {
		t.Errorf("Unexpected error loading default config: %v", err)
	}

	if len(config.Generate) != 1 {
		t.Errorf("Expected 1 task, got %d", len(config.Generate))
	}
}