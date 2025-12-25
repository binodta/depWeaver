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

	// Special case: detect if someone passed a slice of constructors
	if constructorType != nil && (constructorType.Kind() == reflect.Slice || constructorType.Kind() == reflect.Array) {
		return fmt.Errorf("constructor must be a function, got %v. Did you mean to pass individual constructors instead of a slice? Use InitWithScope() or spread the slice elements", constructorType)
	}

	if constructorType.Kind() != reflect.Func {
		return fmt.Errorf("constructor must be a function, got %T", constructor)
	}

	// Validate constructor signature: must return (T) or (T, error)
	if constructorType.NumOut() == 0 || constructorType.NumOut() > 2 {
		return fmt.Errorf("constructor %v must return either (T) or (T, error), but returns %d values", constructorType, constructorType.NumOut())
	}
	if constructorType.NumOut() == 2 {
		if !constructorType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("constructor %v: second return value must be of type error, got %v", constructorType, constructorType.Out(1))
		}
	}

	// Get the primary return type and parameter types
	returnType := constructorType.Out(0)
	numIn := constructorType.NumIn()
	paramTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		paramTypes[i] = constructorType.In(i)
	}

	// Wrap the constructor to work with the container
	wrappedConstructor := func(container *DependencyContainer, scopeID string) (interface{}, error) {
		// Use reflection to call the constructor with dependencies
		constructorValue := reflect.ValueOf(constructor)

		// Prepare arguments for the constructor
		args := make([]reflect.Value, numIn)
		for i := 0; i < numIn; i++ {
			argType := paramTypes[i]
			arg, err := container.resolveWithScope(argType, scopeID)
			if err != nil {
				return nil, fmt.Errorf("error resolving dependency %v (parameter %d of %v): %w", argType, i+1, constructorType, err)
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
		paramTypes:  paramTypes,
	}

	return nil
}

// RegisterRuntimeConstructor allows registration of constructors after initialization
func (dc *DependencyContainer) RegisterRuntimeConstructor(
	constructor interface{},
	scope Scope,
) error {
	return dc.RegisterConstructorWithScope(constructor, scope)
}

// OverrideConstructor replaces an existing constructor and clears any cached instances
func (dc *DependencyContainer) OverrideConstructor(
	constructor interface{},
	scope Scope,
) error {
	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		return fmt.Errorf("constructor must be a function")
	}
	returnType := constructorType.Out(0)

	// Register it
	if err := dc.RegisterConstructorWithScope(constructor, scope); err != nil {
		return err
	}

	// Invalidate caches
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Clear from singleton cache
	delete(dc.dependencies, returnType)

	// Clear from all scope caches
	for _, scopeCache := range dc.scopedInstances {
		delete(scopeCache, returnType)
	}

	return nil
}

// RegisterNamedConstructorWithScope adds a constructor function with a specific name and scope
func (dc *DependencyContainer) RegisterNamedConstructorWithScope(
	name string,
	constructor interface{},
	scope Scope,
) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		return fmt.Errorf("constructor must be a function, got %T", constructor)
	}

	// Validate constructor signature
	if constructorType.NumOut() == 0 || constructorType.NumOut() > 2 {
		return fmt.Errorf("constructor %v must return either (T) or (T, error)", constructorType)
	}
	if constructorType.NumOut() == 2 {
		if !constructorType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("constructor %v: second return value must be of type error", constructorType)
		}
	}

	returnType := constructorType.Out(0)
	numIn := constructorType.NumIn()
	paramTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		paramTypes[i] = constructorType.In(i)
	}

	wrappedConstructor := func(container *DependencyContainer, scopeID string) (interface{}, error) {
		constructorValue := reflect.ValueOf(constructor)
		args := make([]reflect.Value, numIn)
		for i := 0; i < numIn; i++ {
			argType := paramTypes[i]
			arg, err := container.resolveWithScope(argType, scopeID)
			if err != nil {
				return nil, fmt.Errorf("error resolving dependency %v for named %q: %w", argType, name, err)
			}
			args[i] = reflect.ValueOf(arg)
		}

		results := constructorValue.Call(args)
		if constructorType.NumOut() == 1 {
			return results[0].Interface(), nil
		}
		if errVal := results[1]; !errVal.IsNil() {
			return nil, errVal.Interface().(error)
		}
		return results[0].Interface(), nil
	}

	if dc.namedConstructors[name] == nil {
		dc.namedConstructors[name] = make(map[reflect.Type]*Registration)
	}

	dc.namedConstructors[name][returnType] = &Registration{
		constructor: wrappedConstructor,
		scope:       scope,
		paramTypes:  paramTypes,
	}

	// Invalidate caches for this named dependency
	if dc.namedDependencies[name] != nil {
		delete(dc.namedDependencies[name], returnType)
	}
	for _, scopeCache := range dc.namedScopedInstances {
		if namedCache, ok := scopeCache[name]; ok {
			delete(namedCache, returnType)
		}
	}

	return nil
}
