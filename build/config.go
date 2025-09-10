package build

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the structure of typegen.yaml
type Config struct {
	Version  int                    `yaml:"version"`
	Config   map[string]string      `yaml:"config"`
	Generate []GenerateTask         `yaml:"generate"`
}

// GenerateTask represents a single generation task
type GenerateTask struct {
	Generator string            `yaml:"generator"`
	Input     string            `yaml:"input"`
	Output    string            `yaml:"output"`
	Config    map[string]string `yaml:"config"`
}

// LoadConfig loads and validates the typegen.yaml configuration
func LoadConfig(configPath string) (*Config, error) {
	// If no config path provided, look for typegen.yaml in current directory
	if configPath == "" {
		configPath = "typegen.yaml"
	}
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}
	
	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Apply defaults and validate
	if err := config.applyDefaults(); err != nil {
		return nil, err
	}
	
	if err := config.validate(); err != nil {
		return nil, err
	}
	
	return &config, nil
}

// applyDefaults applies default values to the configuration
func (c *Config) applyDefaults() error {
	// Default version to 1
	if c.Version == 0 {
		c.Version = 1
	}
	
	// Initialize global config if nil
	if c.Config == nil {
		c.Config = make(map[string]string)
	}
	
	// Apply defaults to generate tasks
	for i := range c.Generate {
		task := &c.Generate[i]
		
		// Default input to current directory
		if task.Input == "" {
			task.Input = "."
		}
		
		// Initialize task config if nil
		if task.Config == nil {
			task.Config = make(map[string]string)
		}
		
		// Convert relative paths to absolute paths
		if !filepath.IsAbs(task.Input) {
			absInput, err := filepath.Abs(task.Input)
			if err != nil {
				return fmt.Errorf("failed to resolve input path %s: %w", task.Input, err)
			}
			task.Input = absInput
		}
		
		if task.Output != "" && !filepath.IsAbs(task.Output) {
			absOutput, err := filepath.Abs(task.Output)
			if err != nil {
				return fmt.Errorf("failed to resolve output path %s: %w", task.Output, err)
			}
			task.Output = absOutput
		}
	}
	
	return nil
}

// validate validates the configuration
func (c *Config) validate() error {
	// Validate version
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version: %d (supported: 1)", c.Version)
	}
	
	// Validate generate tasks
	if len(c.Generate) == 0 {
		return fmt.Errorf("no generate tasks defined")
	}
	
	for i, task := range c.Generate {
		if task.Generator == "" {
			return fmt.Errorf("generate task %d: generator is required", i)
		}
		
		if task.Output == "" {
			return fmt.Errorf("generate task %d: output is required", i)
		}
		
		// Validate input directory exists
		if info, err := os.Stat(task.Input); os.IsNotExist(err) {
			return fmt.Errorf("generate task %d: input directory does not exist: %s", i, task.Input)
		} else if !info.IsDir() {
			return fmt.Errorf("generate task %d: input path is not a directory: %s", i, task.Input)
		}
	}
	
	return nil
}

// MergedConfig returns the merged configuration for a specific task
// Task configs take precedence over global configs
func (c *Config) MergedConfig(taskIndex int) map[string]string {
	if taskIndex < 0 || taskIndex >= len(c.Generate) {
		return nil
	}
	
	merged := make(map[string]string)
	
	// Start with global config
	for k, v := range c.Config {
		merged[k] = v
	}
	
	// Override with task-specific config
	for k, v := range c.Generate[taskIndex].Config {
		merged[k] = v
	}
	
	return merged
}