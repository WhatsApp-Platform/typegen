package generators

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages registered code generators
type Registry struct {
	mu         sync.RWMutex
	generators map[string]func() Generator
}

// defaultRegistry is the global registry instance
var defaultRegistry = NewRegistry()

// NewRegistry creates a new generator registry
func NewRegistry() *Registry {
	return &Registry{
		generators: make(map[string]func() Generator),
	}
}

// Register registers a generator with the given name
func (r *Registry) Register(name string, constructor func() Generator) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generators[name] = constructor
}

// Get retrieves a generator by name
func (r *Registry) Get(name string) (Generator, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	constructor, exists := r.generators[name]
	if !exists {
		return nil, fmt.Errorf("generator %q not found", name)
	}
	
	return constructor(), nil
}

// List returns all registered generator names
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var names []string
	for name := range r.generators {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Global functions that use the default registry

// Register registers a generator globally
func Register(name string, constructor func() Generator) {
	defaultRegistry.Register(name, constructor)
}

// Get retrieves a generator from the global registry
func Get(name string) (Generator, error) {
	return defaultRegistry.Get(name)
}

// List returns all globally registered generator names
func List() []string {
	return defaultRegistry.List()
}