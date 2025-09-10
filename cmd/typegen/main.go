package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	
	"github.com/WhatsApp-Platform/typegen/build"
	"github.com/WhatsApp-Platform/typegen/generators"
	"github.com/WhatsApp-Platform/typegen/parser"
	
	// Import generators to register them
	_ "github.com/WhatsApp-Platform/typegen/generators/python/pydantic"
	_ "github.com/WhatsApp-Platform/typegen/generators/go"
)

// configFlags implements flag.Value for collecting multiple key=value config options
type configFlags map[string]string

func (c configFlags) String() string {
	var parts []string
	for k, v := range c {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, ",")
}

func (c configFlags) Set(value string) error {
	parts := strings.SplitN(value, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("config option must be in format key=value, got: %s", value)
	}
	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	if key == "" {
		return fmt.Errorf("config key cannot be empty")
	}
	c[key] = val
	return nil
}

const usage = `TypeGen - generate types from a common definition language

Usage:
  typegen <command> [flags] [arguments]

Commands:
  parse     Parse and validate a TypeGen file
  module    Parse all TypeGen files in a module directory  
  generate  Generate code for entire module
  build     Build all targets defined in typegen.yaml

Use "typegen <command> -h" for more information about a command.

Examples:
  typegen parse user.tg
  typegen module ./api/auth
  typegen generate -generator python+pydantic -o ./generated/python ./schemas
  typegen build
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	switch command {
	case "parse":
		handleParse(os.Args[2:])
	case "module":
		handleModule(os.Args[2:])
	case "generate":
		handleGenerate(os.Args[2:])
	case "build":
		handleBuild(os.Args[2:])
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		fmt.Print(usage)
		os.Exit(1)
	}
}

func handleParse(args []string) {
	parseCmd := flag.NewFlagSet("parse", flag.ExitOnError)
	parseCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: typegen parse [flags] <file>\n\n")
		fmt.Fprintf(os.Stderr, "Parse and validate a TypeGen file\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		parseCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  <file>  Path to the TypeGen file to parse\n")
	}
	
	parseCmd.Parse(args)
	
	if parseCmd.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: parse command requires a file argument\n\n")
		parseCmd.Usage()
		os.Exit(1)
	}
	
	filename := parseCmd.Arg(0)
	
	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("Error: file '%s' does not exist\n", filename)
		os.Exit(1)
	}
	
	// Parse the file
	program, err := parser.ParseFile(filename)
	if err != nil {
		fmt.Printf("Parse error in %s:\n%v\n", filename, err)
		os.Exit(1)
	}
	
	// Print the parsed AST
	fmt.Printf("Successfully parsed %s:\n\n", filename)
	fmt.Println(program.String())
}

func handleModule(args []string) {
	moduleCmd := flag.NewFlagSet("module", flag.ExitOnError)
	moduleCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: typegen module [flags] <directory>\n\n")
		fmt.Fprintf(os.Stderr, "Parse all TypeGen files in a module directory\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		moduleCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  <directory>  Path to the module directory to parse\n")
	}
	
	moduleCmd.Parse(args)
	
	if moduleCmd.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: module command requires a directory argument\n\n")
		moduleCmd.Usage()
		os.Exit(1)
	}
	
	modulePath := moduleCmd.Arg(0)
	
	// Check if directory exists
	if info, err := os.Stat(modulePath); os.IsNotExist(err) {
		fmt.Printf("Error: directory '%s' does not exist\n", modulePath)
		os.Exit(1)
	} else if !info.IsDir() {
		fmt.Printf("Error: '%s' is not a directory\n", modulePath)
		os.Exit(1)
	}
	
	// Parse the module
	programs, err := parser.ParseModule(modulePath)
	if err != nil {
		fmt.Printf("Module parse error in %s:\n%v\n", modulePath, err)
		os.Exit(1)
	}
	
	// Print results
	fmt.Printf("Successfully parsed module %s:\n\n", modulePath)
	
	if len(programs) == 0 {
		fmt.Println("No .tg files found in the module directory.")
		return
	}
	
	for filename, program := range programs {
		fmt.Printf("=== %s ===\n", filename)
		fmt.Println(program.String())
		fmt.Println()
	}
	
	fmt.Printf("Total files parsed: %d\n", len(programs))
}


func handleGenerate(args []string) {
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	
	// Define flags
	generator := generateCmd.String("generator", "", "Target generator for code generation")
	outputDir := generateCmd.String("o", "", "Output directory for generated code")
	config := make(configFlags)
	generateCmd.Var(config, "c", "Configuration option in format key=value (can be used multiple times)")
	
	generateCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: typegen generate [flags] <module-directory>\n\n")
		fmt.Fprintf(os.Stderr, "Generate code for entire module\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		generateCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nArguments:\n")
		fmt.Fprintf(os.Stderr, "  <module-directory>  Path to the module directory to generate from\n")
		fmt.Fprintf(os.Stderr, "\nAvailable generators: %v\n", generators.List())
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  typegen generate -generator python+pydantic -o ./output -c indent=4 -c package=myapp ./schemas\n")
	}
	
	generateCmd.Parse(args)
	
	if generateCmd.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Error: generate command requires a module directory argument\n\n")
		generateCmd.Usage()
		os.Exit(1)
	}
	
	if *generator == "" {
		fmt.Fprintf(os.Stderr, "Error: -generator flag is required\n\n")
		generateCmd.Usage()
		os.Exit(1)
	}
	
	if *outputDir == "" {
		fmt.Fprintf(os.Stderr, "Error: -o flag is required\n\n")
		generateCmd.Usage()
		os.Exit(1)
	}
	
	modulePath := generateCmd.Arg(0)
	
	// Display config options if any were provided
	if len(config) > 0 {
		fmt.Printf("Using config options: %v\n", map[string]string(config))
	}
	
	// Check if module directory exists
	if info, err := os.Stat(modulePath); os.IsNotExist(err) {
		fmt.Printf("Error: module directory '%s' does not exist\n", modulePath)
		os.Exit(1)
	} else if !info.IsDir() {
		fmt.Printf("Error: '%s' is not a directory\n", modulePath)
		os.Exit(1)
	}
	
	// Parse the module
	module, err := parser.ParseModuleToAST(modulePath)
	if err != nil {
		fmt.Printf("Module parse error in %s:\n%v\n", modulePath, err)
		os.Exit(1)
	}
	
	// Get the generator for the specified name
	gen, err := generators.Get(*generator)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Available generators: %v\n", generators.List())
		os.Exit(1)
	}
	
	// Set config on the generator
	gen.SetConfig(map[string]string(config))
	
	// Create filesystem for output
	fs := generators.NewOSFS(*outputDir)
	
	// Generate code
	ctx := context.Background()
	if err := gen.Generate(ctx, module, fs); err != nil {
		fmt.Printf("Generation error: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Generated %s code for module %s in %s\n", *generator, module.Name, *outputDir)
}

func handleBuild(args []string) {
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	
	// Define flags
	configPath := buildCmd.String("f", "", "Path to typegen.yaml configuration file (default: ./typegen.yaml)")
	
	buildCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: typegen build [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Build all targets defined in typegen.yaml\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		buildCmd.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  typegen build\n")
		fmt.Fprintf(os.Stderr, "  typegen build -f custom-config.yaml\n")
	}
	
	buildCmd.Parse(args)
	
	// Load configuration
	config, err := build.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}
	
	// Create builder
	builder := build.NewBuilder(config)
	
	// Validate generators before starting build
	if err := builder.ValidateGenerators(); err != nil {
		fmt.Printf("Configuration validation error: %v\n", err)
		os.Exit(1)
	}
	
	// Execute build
	ctx := context.Background()
	if err := builder.Build(ctx); err != nil {
		fmt.Printf("Build failed: %v\n", err)
		os.Exit(1)
	}
}