package container

import (
	"fmt"
	"reflect"
)

// RegisterConstructor adds a constructor function for a specific type
// @Param constructor interface{} - constructor function
func (dc *DependencyContainer) RegisterConstructor(
	constructor interface{},
) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	constructorType := reflect.TypeOf(constructor)
	if constructorType.Kind() != reflect.Func {
		return fmt.Errorf("constructor must be a function")
	}

	// Get the return type of the constructor (the type it creates)
	returnType := constructorType.Out(0)

	// Wrap the constructor to work with the container
	dc.constructors[returnType] = func(container *DependencyContainer) (interface{}, error) {
		// Use reflection to call the constructor with dependencies
		constructorValue := reflect.ValueOf(constructor)
		constructorType := constructorValue.Type()

		// Prepare arguments for the constructor
		args := make([]reflect.Value, constructorType.NumIn())
		for i := 0; i < constructorType.NumIn(); i++ {
			argType := constructorType.In(i)
			arg, err := container.resolve(argType)
			if err != nil {
				return nil, fmt.Errorf("error resolving dependency for %v: %v", argType, err)
			}
			args[i] = reflect.ValueOf(arg)
		}

		// Call the constructor
		results := constructorValue.Call(args)
		return results[0].Interface(), nil
	}

	return nil
}
