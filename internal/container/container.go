package container

import (
	"reflect"
	"sync"
)

// Scope defines the lifetime of a dependency
type Scope int

const (
	Singleton Scope = iota // Created once and cached (default behavior)
	Transient              // Created every time it's requested
	Scoped                 // Created once per scope context
)

// Registration holds constructor and scope information
type Registration struct {
	constructor func(container *DependencyContainer, scopeID string) (interface{}, error)
	scope       Scope
}

type DependencyContainer struct {
	mu              sync.RWMutex
	dependencies    map[reflect.Type]interface{}            // Singleton cache
	constructors    map[reflect.Type]*Registration          // Constructor registrations with scope
	creating        map[reflect.Type]bool                   // Track types being created (circular dependency detection)
	resolutionStack []reflect.Type                          // Track dependency chain for better error reporting
	scopedInstances map[string]map[reflect.Type]interface{} // Scoped instances by context ID
}

// New creates a new dependency container
func New() *DependencyContainer {
	return &DependencyContainer{
		dependencies:    make(map[reflect.Type]interface{}),
		constructors:    make(map[reflect.Type]*Registration),
		creating:        make(map[reflect.Type]bool),
		resolutionStack: make([]reflect.Type, 0),
		scopedInstances: make(map[string]map[reflect.Type]interface{}),
	}
}
