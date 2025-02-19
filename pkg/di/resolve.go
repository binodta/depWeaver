package di

import (
	"fmt"
	"reflect"
)

// Resolve resolves the instance of the given type from the container
func Resolve[T interface{}]() (T, error) {

	var zero T
	t := reflect.TypeOf(&zero).Elem()
	// Resolve the instance from container
	instance, err := dependencyContainer.Resolve(t)
	if err != nil {
		return zero, fmt.Errorf("failed to resolve type %v: %w", t, err)
	}

	// Try to cast the instance to the desired type
	if instance == nil {
		return zero, fmt.Errorf("resolved instance is nil for type %v", t)
	}

	castedInstance, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("failed to cast resolved instance to type %v", t)
	}

	return castedInstance, nil
}
