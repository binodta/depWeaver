package container

import (
	"fmt"
	"reflect"
)

// Resolve public method to resolve dependencies
// @Param t reflect.Type - type of the dependency
func (dc *DependencyContainer) Resolve(t reflect.Type) (interface{}, error) {
	return dc.resolve(t)
}

// resolve pkg method to resolve dependencies
// @Param t reflect.Type - type of the dependency
// @Return interface{} - instance of the dependency
func (dc *DependencyContainer) resolve(t reflect.Type) (interface{}, error) {
	dc.mu.Lock()

	// Check for circular dependencies
	if dc.creating[t] {
		dc.mu.Unlock()
		return nil, fmt.Errorf("circular dependency detected for type %v", t)
	}

	// Check if we already have an instance
	if dep, exists := dc.dependencies[t]; exists {
		dc.mu.Unlock()
		return dep, nil
	}

	// Find the constructor for this type
	constructor, exists := dc.constructors[t]
	if !exists {
		dc.mu.Unlock()
		return nil, fmt.Errorf("no constructor registered for type %v", t)
	}

	// Mark as currently being created to prevent recursion
	dc.creating[t] = true
	dc.mu.Unlock()

	// Create the instance
	instance, err := constructor(dc)
	if err != nil {
		dc.mu.Lock()
		delete(dc.creating, t)
		dc.mu.Unlock()
		return nil, err
	}

	// Store the created instance
	dc.mu.Lock()
	dc.dependencies[t] = instance
	delete(dc.creating, t)
	dc.mu.Unlock()

	return instance, nil
}
