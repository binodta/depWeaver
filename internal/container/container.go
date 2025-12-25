package container

import (
	"reflect"
	"sync"
)

type DependencyContainer struct {
	mu              sync.RWMutex
	dependencies    map[reflect.Type]interface{}
	constructors    map[reflect.Type]func(container *DependencyContainer) (interface{}, error)
	creating        map[reflect.Type]bool
	resolutionStack []reflect.Type // Track dependency chain for better error reporting
}

// New creates a new dependency container
func New() *DependencyContainer {
	return &DependencyContainer{
		dependencies:    make(map[reflect.Type]interface{}),
		constructors:    make(map[reflect.Type]func(container *DependencyContainer) (interface{}, error)),
		creating:        make(map[reflect.Type]bool),
		resolutionStack: make([]reflect.Type, 0),
	}
}
