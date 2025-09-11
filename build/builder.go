package build

import (
	"context"
	"fmt"

	"github.com/WhatsApp-Platform/typegen/generators"
	"github.com/WhatsApp-Platform/typegen/parser"
	"github.com/WhatsApp-Platform/typegen/parser/ast"
	"github.com/WhatsApp-Platform/typegen/validator"
)

// Builder orchestrates the build process
type Builder struct {
	config          *Config
	moduleCache     map[string]*ast.Module                 // Cache parsed modules
	validationCache map[string]*validator.ValidationResult // Cache validation results
}

// NewBuilder creates a new builder with the given configuration
func NewBuilder(config *Config) *Builder {
	return &Builder{
		config:          config,
		moduleCache:     make(map[string]*ast.Module),
		validationCache: make(map[string]*validator.ValidationResult),
	}
}

// Build executes all generation tasks defined in the configuration
func (b *Builder) Build(ctx context.Context) error {
	if b.config == nil {
		return fmt.Errorf("no configuration provided")
	}

	fmt.Printf("Starting build with %d generation tasks...\n", len(b.config.Generate))

	// Track errors but continue processing all tasks
	var buildErrors []error
	successCount := 0

	for i, task := range b.config.Generate {
		fmt.Printf("\n[%d/%d] Generating %s code from %s to %s...\n",
			i+1, len(b.config.Generate), task.Generator, task.Input, task.Output)

		if err := b.executeTask(ctx, task, i); err != nil {
			buildErrors = append(buildErrors, fmt.Errorf("task %d (%s): %w", i+1, task.Generator, err))
			fmt.Printf("❌ Failed: %v\n", err)
		} else {
			successCount++
			fmt.Printf("✅ Success\n")
		}
	}

	// Report results
	fmt.Printf("\nBuild completed: %d/%d tasks succeeded\n", successCount, len(b.config.Generate))

	if len(buildErrors) > 0 {
		fmt.Printf("\nErrors encountered:\n")
		for _, err := range buildErrors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("build failed with %d errors", len(buildErrors))
	}

	return nil
}

// executeTask executes a single generation task
func (b *Builder) executeTask(ctx context.Context, task GenerateTask, taskIndex int) error {
	// Get the generator for the specified language
	generator, err := generators.Get(task.Generator)
	if err != nil {
		return fmt.Errorf("generator not found: %w", err)
	}

	// Get merged configuration for this task
	mergedConfig := b.config.MergedConfig(taskIndex)

	// Set configuration on the generator
	generator.SetConfig(mergedConfig)

	// Parse the input module (cached)
	module, err := b.getOrParseModule(task.Input)
	if err != nil {
		return err
	}

	// Validate the module before generation (cached)
	result, err := b.getOrValidateModule(module, task.Input)
	if err != nil {
		return err
	}

	if result != nil && result.HasErrors() {
		return fmt.Errorf("validation failed with %d errors:\n%s", result.ErrorCount(), result.String())
	}

	// Create filesystem for output
	fs := generators.NewOSFS(task.Output)

	// Generate code
	if err := generator.Generate(ctx, module, fs); err != nil {
		return fmt.Errorf("code generation failed: %w", err)
	}

	return nil
}

// ValidateGenerators checks if all generators specified in the config are available
func (b *Builder) ValidateGenerators() error {
	availableGenerators := generators.List()
	generatorSet := make(map[string]bool)
	for _, gen := range availableGenerators {
		generatorSet[gen] = true
	}

	var missingGenerators []string
	for i, task := range b.config.Generate {
		if !generatorSet[task.Generator] {
			missingGenerators = append(missingGenerators,
				fmt.Sprintf("task %d: %s", i+1, task.Generator))
		}
	}

	if len(missingGenerators) > 0 {
		return fmt.Errorf("unknown generators: %v\nAvailable generators: %v",
			missingGenerators, availableGenerators)
	}

	return nil
}

// getOrParseModule gets a module from cache or parses it if not cached
func (b *Builder) getOrParseModule(modulePath string) (*ast.Module, error) {
	// Check cache first
	if module, exists := b.moduleCache[modulePath]; exists {
		return module, nil
	}

	// Parse the module
	module, err := parser.ParseModuleToAST(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse module: %w", err)
	}

	// Cache the result
	b.moduleCache[modulePath] = module
	return module, nil
}

// getOrValidateModule gets validation result from cache or validates if not cached
func (b *Builder) getOrValidateModule(module *ast.Module, modulePath string) (*validator.ValidationResult, error) {
	// Check cache first
	if _, exists := b.validationCache[modulePath]; exists {
		return nil, nil
	}

	// Validate the module
	v := validator.NewValidator()
	result := v.Validate(module)

	// Cache the result
	b.validationCache[modulePath] = result
	return result, nil
}
