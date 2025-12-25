package container

import (
	"fmt"
	"reflect"
)

// RegisterConstructor adds a constructor function for a specific type with Singleton scope (default)
// @Param constructor interface{} - constructor function
func (dc *DependencyContainer) RegisterConstructor(
	constructor interface{},
) error {
	return dc.RegisterConstructorWithScope(constructor, Singleton)
}

// RegisterConstructorWithScope adds a constructor function with a specific scope
// @Param constructor interface{} - constructor function
// @Param scope Scope - lifetime scope for the dependency
func (dc *DependencyContainer) RegisterConstructorWithScope(
	constructor interface{},
	scope Scope,
) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		return fmt.Errorf("constructor must be a function")
	}

	// Validate constructor signature: must return (T) or (T, error)
	if constructorType.NumOut() == 0 || constructorType.NumOut() > 2 {
		return fmt.Errorf("constructor must return either (T) or (T, error)")
	}
	if constructorType.NumOut() == 2 {
		if !constructorType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("second return value must be of type error")
		}
	}

	// Get the primary return type of the constructor (the type it creates)
	returnType := constructorType.Out(0)

	// Wrap the constructor to work with the container
	wrappedConstructor := func(container *DependencyContainer, scopeID string) (interface{}, error) {
		// Use reflection to call the constructor with dependencies
		constructorValue := reflect.ValueOf(constructor)
		constructorType := constructorValue.Type()

		// Prepare arguments for the constructor
		args := make([]reflect.Value, constructorType.NumIn())
		for i := 0; i < constructorType.NumIn(); i++ {
			argType := constructorType.In(i)
			arg, err := container.resolveWithScope(argType, scopeID)
			if err != nil {
				return nil, fmt.Errorf("error resolving dependency for %v: %v", argType, err)
			}
			args[i] = reflect.ValueOf(arg)
		}

		// Call the constructor
		results := constructorValue.Call(args)
		// Handle (T) signature
		if constructorType.NumOut() == 1 {
			return results[0].Interface(), nil
		}
		// Handle (T, error) signature
		if errVal := results[1]; !errVal.IsNil() {
			return nil, errVal.Interface().(error)
		}
		return results[0].Interface(), nil
	}

	dc.constructors[returnType] = &Registration{
		constructor: wrappedConstructor,
		scope:       scope,
	}

	return nil
}

// RegisterRuntimeConstructor allows registration of constructors after initialization
// @Param constructor interface{} - constructor function
// @Param scope Scope - lifetime scope for the dependency
func (dc *DependencyContainer) RegisterRuntimeConstructor(
	constructor interface{},
	scope Scope,
) error {
	// Runtime registration is the same as regular registration
	// The container is already thread-safe with locks
	return dc.RegisterConstructorWithScope(constructor, scope)
}
