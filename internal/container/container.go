package container

import (
	"reflect"
	"sync"
)

type DependencyContainer struct {
	mu           sync.RWMutex
	dependencies map[reflect.Type]interface{}
	constructors map[reflect.Type]func(container *DependencyContainer) (interface{}, error)
	creating     map[reflect.Type]bool
}

// New creates a new dependency container
func New() *DependencyContainer {
	return &DependencyContainer{
		dependencies: make(map[reflect.Type]interface{}),
		constructors: make(map[reflect.Type]func(container *DependencyContainer) (interface{}, error)),
		creating:     make(map[reflect.Type]bool),
	}
}
